package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invoicerequest"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	InvoiceTypeVATGeneral = "vat_general"
	InvoiceTypeVATSpecial = "vat_special"

	InvoiceStatusPending  = "pending"
	InvoiceStatusApproved = "approved"
	InvoiceStatusRejected = "rejected"
	InvoiceStatusIssued   = "issued"

	defaultInvoiceCurrency       = payment.DefaultPaymentCurrency
	defaultInvoiceStorageSubdir  = "invoices"
	defaultInvoiceMaxUploadBytes = int64(10 << 20)
)

var allowedInvoiceFileTypes = map[string]string{
	".pdf":  "application/pdf",
	".ofd":  "application/ofd",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
}

type InvoiceSummary struct {
	Currency        string  `json:"currency"`
	EligibleAmount  float64 `json:"eligible_amount"`
	RequestedAmount float64 `json:"requested_amount"`
	AvailableAmount float64 `json:"available_amount"`
}

type InvoiceClaimSummary struct {
	ClaimableCount int `json:"claimable_count"`
}

type InvoiceRequestCreateInput struct {
	UserID      int64
	Amount      float64
	InvoiceType string
	Title       string
	TaxNumber   string
	Remark      string
}

type InvoiceRequestListParams struct {
	Page   int
	Limit  int
	Status string
	UserID int64
}

type InvoiceAdminReviewInput struct {
	AdminID   int64
	AdminNote string
}

type InvoiceAdminIssueInput struct {
	AdminID     int64
	AdminNote   string
	InvoiceNo   string
	FileName    string
	ContentType string
	FileSize    int64
	Reader      io.Reader
}

type InvoiceDownloadFile struct {
	Path        string
	FileName    string
	ContentType string
	Size        int64
	ModTime     time.Time
}

type InvoiceRequestView struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	UserEmail       string     `json:"user_email,omitempty"`
	UserName        string     `json:"user_name,omitempty"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	InvoiceType     string     `json:"invoice_type"`
	Title           string     `json:"title"`
	TaxNumber       string     `json:"tax_number"`
	Remark          string     `json:"remark,omitempty"`
	Status          string     `json:"status"`
	AdminNote       string     `json:"admin_note,omitempty"`
	InvoiceNo       string     `json:"invoice_no,omitempty"`
	FileName        string     `json:"file_name,omitempty"`
	FileSize        int64      `json:"file_size,omitempty"`
	FileContentType string     `json:"file_content_type,omitempty"`
	ReviewedBy      *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	DownloadedAt    *time.Time `json:"downloaded_at,omitempty"`
	DownloadCount   int        `json:"download_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Downloadable    bool       `json:"downloadable"`
	Claimable       bool       `json:"claimable"`
}

func (s *PaymentService) GetInvoiceSummary(ctx context.Context, userID int64) (*InvoiceSummary, error) {
	eligible, err := s.calculateInvoiceEligibleAmount(ctx, userID)
	if err != nil {
		return nil, err
	}
	requested, err := s.calculateInvoiceReservedAmount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &InvoiceSummary{
		Currency:        defaultInvoiceCurrency,
		EligibleAmount:  invoiceRoundAmount(eligible),
		RequestedAmount: invoiceRoundAmount(requested),
		AvailableAmount: invoiceRoundAmount(math.Max(0, eligible-requested)),
	}, nil
}

func (s *PaymentService) ListUserInvoices(ctx context.Context, userID int64, params InvoiceRequestListParams) ([]InvoiceRequestView, int, error) {
	params.UserID = userID
	items, total, err := s.listInvoiceRequests(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return invoiceRequestViews(items), total, nil
}

func (s *PaymentService) GetInvoiceClaimSummary(ctx context.Context, userID int64) (*InvoiceClaimSummary, error) {
	if userID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	count, err := s.entClient.InvoiceRequest.Query().
		Where(
			invoicerequest.UserIDEQ(userID),
			invoicerequest.StatusEQ(InvoiceStatusIssued),
			invoicerequest.FilePathNotNil(),
			invoicerequest.FilePathNEQ(""),
			invoicerequest.DownloadedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count invoice claimable requests: %w", err)
	}
	return &InvoiceClaimSummary{ClaimableCount: count}, nil
}

func (s *PaymentService) CreateInvoiceRequest(ctx context.Context, input InvoiceRequestCreateInput) (*InvoiceRequestView, error) {
	if input.UserID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	amount := invoiceRoundAmount(input.Amount)
	if amount <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_AMOUNT", "invoice amount must be greater than zero")
	}
	invoiceType := strings.TrimSpace(input.InvoiceType)
	if !isValidInvoiceType(invoiceType) {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_TYPE", "invalid invoice type")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_TITLE", "invoice title is required")
	}
	taxNumber := strings.TrimSpace(input.TaxNumber)
	if taxNumber == "" {
		return nil, infraerrors.BadRequest("INVALID_TAX_NUMBER", "tax number is required")
	}
	remark := strings.TrimSpace(input.Remark)

	summary, err := s.GetInvoiceSummary(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if amount-summary.AvailableAmount > amountToleranceCNY {
		return nil, infraerrors.BadRequest("INVOICE_AMOUNT_EXCEEDED", "invoice amount exceeds available amount").
			WithMetadata(map[string]string{
				"available_amount": fmt.Sprintf("%.2f", summary.AvailableAmount),
			})
	}

	now := s.nowTime()
	req, err := s.entClient.InvoiceRequest.Create().
		SetUserID(input.UserID).
		SetAmount(amount).
		SetCurrency(defaultInvoiceCurrency).
		SetInvoiceType(invoiceType).
		SetTitle(title).
		SetTaxNumber(taxNumber).
		SetNillableRemark(psNilIfEmpty(remark)).
		SetStatus(InvoiceStatusPending).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create invoice request: %w", err)
	}
	view := invoiceRequestView(req)
	return &view, nil
}

func (s *PaymentService) AdminListInvoiceRequests(ctx context.Context, params InvoiceRequestListParams) ([]InvoiceRequestView, int, error) {
	items, total, err := s.listInvoiceRequests(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return invoiceRequestViews(items), total, nil
}

func (s *PaymentService) AdminGetInvoiceRequest(ctx context.Context, requestID int64) (*InvoiceRequestView, error) {
	req, err := s.getInvoiceRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	view := invoiceRequestView(req)
	return &view, nil
}

func (s *PaymentService) AdminApproveInvoiceRequest(ctx context.Context, requestID int64, input InvoiceAdminReviewInput) (*InvoiceRequestView, error) {
	req, err := s.getInvoiceRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != InvoiceStatusPending {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_STATUS", "only pending invoice requests can be approved")
	}
	now := s.nowTime()
	updated, err := s.entClient.InvoiceRequest.UpdateOneID(requestID).
		SetStatus(InvoiceStatusApproved).
		SetNillableAdminNote(psNilIfEmpty(strings.TrimSpace(input.AdminNote))).
		SetNillableReviewedBy(invoiceAdminIDPtr(input.AdminID)).
		SetReviewedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("approve invoice request: %w", err)
	}
	view := invoiceRequestView(updated)
	return &view, nil
}

func (s *PaymentService) AdminRejectInvoiceRequest(ctx context.Context, requestID int64, input InvoiceAdminReviewInput) (*InvoiceRequestView, error) {
	req, err := s.getInvoiceRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != InvoiceStatusPending && req.Status != InvoiceStatusApproved {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_STATUS", "only pending or approved invoice requests can be rejected")
	}
	note := strings.TrimSpace(input.AdminNote)
	if note == "" {
		return nil, infraerrors.BadRequest("INVALID_ADMIN_NOTE", "reject reason is required")
	}
	now := s.nowTime()
	updated, err := s.entClient.InvoiceRequest.UpdateOneID(requestID).
		SetStatus(InvoiceStatusRejected).
		SetAdminNote(note).
		SetNillableReviewedBy(invoiceAdminIDPtr(input.AdminID)).
		SetReviewedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("reject invoice request: %w", err)
	}
	view := invoiceRequestView(updated)
	return &view, nil
}

func (s *PaymentService) AdminIssueInvoiceRequest(ctx context.Context, requestID int64, input InvoiceAdminIssueInput) (*InvoiceRequestView, error) {
	req, err := s.getInvoiceRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != InvoiceStatusApproved && req.Status != InvoiceStatusPending {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_STATUS", "only pending or approved invoice requests can be issued")
	}
	if input.Reader == nil {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_FILE", "invoice file is required")
	}
	storedPath, contentType, size, err := s.storeInvoiceFile(input)
	if err != nil {
		return nil, err
	}
	if req.FilePath != nil && *req.FilePath != "" && *req.FilePath != storedPath {
		_ = os.Remove(*req.FilePath)
	}
	now := s.nowTime()
	updated, err := s.entClient.InvoiceRequest.UpdateOneID(requestID).
		SetStatus(InvoiceStatusIssued).
		SetNillableAdminNote(psNilIfEmpty(strings.TrimSpace(input.AdminNote))).
		SetNillableInvoiceNo(psNilIfEmpty(strings.TrimSpace(input.InvoiceNo))).
		SetFileName(sanitizeInvoiceOriginalFileName(input.FileName)).
		SetFilePath(storedPath).
		SetFileSize(size).
		SetFileContentType(contentType).
		SetNillableReviewedBy(invoiceAdminIDPtr(input.AdminID)).
		SetReviewedAt(now).
		SetIssuedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		_ = os.Remove(storedPath)
		return nil, fmt.Errorf("issue invoice request: %w", err)
	}
	view := invoiceRequestView(updated)
	s.notifyInvoiceIssuedBestEffort(ctx, view)
	return &view, nil
}

func (s *PaymentService) notifyInvoiceIssuedBestEffort(ctx context.Context, invoice InvoiceRequestView) {
	if s == nil || s.systemTicketSvc == nil || invoice.UserID <= 0 || invoice.ID <= 0 {
		return
	}
	event := NewInvoiceIssuedSystemTicketNotification(invoice)
	s.systemTicketSvc.NotifyEventBestEffort(ctx, "service.payment", invoice.UserID, event)
}

func (s *PaymentService) GetInvoiceDownload(ctx context.Context, userID, requestID int64) (*InvoiceDownloadFile, error) {
	req, err := s.getInvoiceRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.UserID != userID {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this invoice")
	}
	if req.Status != InvoiceStatusIssued {
		return nil, infraerrors.BadRequest("INVOICE_NOT_ISSUED", "invoice is not issued")
	}
	if req.FilePath == nil || strings.TrimSpace(*req.FilePath) == "" {
		return nil, infraerrors.NotFound("INVOICE_FILE_NOT_FOUND", "invoice file not found")
	}
	path := strings.TrimSpace(*req.FilePath)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, infraerrors.NotFound("INVOICE_FILE_NOT_FOUND", "invoice file not found")
	}
	now := s.nowTime()
	update := s.entClient.InvoiceRequest.UpdateOneID(req.ID).
		AddDownloadCount(1).
		SetUpdatedAt(now)
	if req.DownloadedAt == nil {
		update.SetDownloadedAt(now)
	}
	if _, err := update.Save(ctx); err != nil {
		return nil, fmt.Errorf("mark invoice downloaded: %w", err)
	}
	fileName := fmt.Sprintf("invoice-%d", req.ID)
	if req.FileName != nil && strings.TrimSpace(*req.FileName) != "" {
		fileName = sanitizeInvoiceOriginalFileName(*req.FileName)
	}
	contentType := "application/octet-stream"
	if req.FileContentType != nil && strings.TrimSpace(*req.FileContentType) != "" {
		contentType = strings.TrimSpace(*req.FileContentType)
	}
	return &InvoiceDownloadFile{
		Path:        path,
		FileName:    fileName,
		ContentType: contentType,
		Size:        info.Size(),
		ModTime:     info.ModTime(),
	}, nil
}

func (s *PaymentService) listInvoiceRequests(ctx context.Context, params InvoiceRequestListParams) ([]*dbent.InvoiceRequest, int, error) {
	q := s.entClient.InvoiceRequest.Query()
	if params.UserID > 0 {
		q = q.Where(invoicerequest.UserIDEQ(params.UserID))
	}
	if status := strings.TrimSpace(params.Status); status != "" {
		if !isValidInvoiceStatus(status) {
			return nil, 0, infraerrors.BadRequest("INVALID_INVOICE_STATUS", "invalid invoice status")
		}
		q = q.Where(invoicerequest.StatusEQ(status))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count invoice requests: %w", err)
	}
	size, page := applyPagination(params.Limit, params.Page)
	items, err := q.WithUser().
		Order(dbent.Desc(invoicerequest.FieldCreatedAt)).
		Limit(size).
		Offset((page - 1) * size).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query invoice requests: %w", err)
	}
	return items, total, nil
}

func (s *PaymentService) getInvoiceRequest(ctx context.Context, requestID int64) (*dbent.InvoiceRequest, error) {
	if requestID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INVOICE_ID", "invalid invoice request id")
	}
	req, err := s.entClient.InvoiceRequest.Query().
		Where(invoicerequest.IDEQ(requestID)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "invoice request not found")
	}
	return req, nil
}

func (s *PaymentService) calculateInvoiceEligibleAmount(ctx context.Context, userID int64) (float64, error) {
	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(userID),
			paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
			paymentorder.StatusIn(OrderStatusCompleted, OrderStatusPartiallyRefunded),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query invoice eligible orders: %w", err)
	}
	total := decimal.Zero
	for _, order := range orders {
		if PaymentOrderCurrency(order) != defaultInvoiceCurrency {
			continue
		}
		eligible := order.PayAmount
		if order.RefundAmount > 0 {
			eligible -= calculateGatewayRefundAmount(order.Amount, order.PayAmount, order.RefundAmount, PaymentOrderCurrency(order))
		}
		if eligible > 0 {
			total = total.Add(decimal.NewFromFloat(eligible))
		}
	}
	return invoiceRoundDecimal(total), nil
}

func (s *PaymentService) calculateInvoiceReservedAmount(ctx context.Context, userID int64) (float64, error) {
	items, err := s.entClient.InvoiceRequest.Query().
		Where(
			invoicerequest.UserIDEQ(userID),
			invoicerequest.CurrencyEQ(defaultInvoiceCurrency),
			invoicerequest.StatusNEQ(InvoiceStatusRejected),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query invoice reserved amount: %w", err)
	}
	total := decimal.Zero
	for _, item := range items {
		total = total.Add(decimal.NewFromFloat(item.Amount))
	}
	return invoiceRoundDecimal(total), nil
}

func (s *PaymentService) storeInvoiceFile(input InvoiceAdminIssueInput) (string, string, int64, error) {
	original := sanitizeInvoiceOriginalFileName(input.FileName)
	ext := strings.ToLower(filepath.Ext(original))
	contentType, ok := allowedInvoiceFileTypes[ext]
	if !ok {
		return "", "", 0, infraerrors.BadRequest("INVALID_INVOICE_FILE_TYPE", "unsupported invoice file type")
	}
	if detected := strings.TrimSpace(input.ContentType); detected != "" && detected != "application/octet-stream" {
		if !invoiceContentTypeAllowedForExt(ext, detected) {
			return "", "", 0, infraerrors.BadRequest("INVALID_INVOICE_FILE_TYPE", "invoice file content type does not match extension")
		}
		contentType = detected
	}
	dir := invoiceStorageDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", 0, fmt.Errorf("create invoice storage dir: %w", err)
	}
	name, err := randomInvoiceStoredName(ext)
	if err != nil {
		return "", "", 0, err
	}
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", "", 0, fmt.Errorf("create invoice file: %w", err)
	}
	closed := false
	closeFile := func() {
		if !closed {
			_ = f.Close()
			closed = true
		}
	}
	defer closeFile()

	limited := io.LimitReader(input.Reader, defaultInvoiceMaxUploadBytes+1)
	size, err := io.Copy(f, limited)
	if err != nil {
		closeFile()
		_ = os.Remove(path)
		return "", "", 0, fmt.Errorf("write invoice file: %w", err)
	}
	if size <= 0 {
		closeFile()
		_ = os.Remove(path)
		return "", "", 0, infraerrors.BadRequest("INVALID_INVOICE_FILE", "invoice file is empty")
	}
	if size > defaultInvoiceMaxUploadBytes {
		closeFile()
		_ = os.Remove(path)
		return "", "", 0, infraerrors.BadRequest("INVOICE_FILE_TOO_LARGE", "invoice file exceeds 10MB")
	}
	closeFile()
	return path, contentType, size, nil
}

func invoiceRequestViews(items []*dbent.InvoiceRequest) []InvoiceRequestView {
	out := make([]InvoiceRequestView, 0, len(items))
	for _, item := range items {
		out = append(out, invoiceRequestView(item))
	}
	return out
}

func invoiceRequestView(item *dbent.InvoiceRequest) InvoiceRequestView {
	if item == nil {
		return InvoiceRequestView{}
	}
	view := InvoiceRequestView{
		ID:            item.ID,
		UserID:        item.UserID,
		Amount:        invoiceRoundAmount(item.Amount),
		Currency:      item.Currency,
		InvoiceType:   item.InvoiceType,
		Title:         item.Title,
		TaxNumber:     item.TaxNumber,
		Status:        item.Status,
		ReviewedBy:    item.ReviewedBy,
		ReviewedAt:    item.ReviewedAt,
		IssuedAt:      item.IssuedAt,
		DownloadedAt:  item.DownloadedAt,
		DownloadCount: item.DownloadCount,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		Downloadable:  item.Status == InvoiceStatusIssued && item.FilePath != nil && strings.TrimSpace(*item.FilePath) != "",
	}
	view.Claimable = view.Downloadable && view.DownloadedAt == nil
	if item.Remark != nil {
		view.Remark = *item.Remark
	}
	if item.AdminNote != nil {
		view.AdminNote = *item.AdminNote
	}
	if item.InvoiceNo != nil {
		view.InvoiceNo = *item.InvoiceNo
	}
	if item.FileName != nil {
		view.FileName = *item.FileName
	}
	if item.FileSize != nil {
		view.FileSize = *item.FileSize
	}
	if item.FileContentType != nil {
		view.FileContentType = *item.FileContentType
	}
	if item.Edges.User != nil {
		view.UserEmail = item.Edges.User.Email
		view.UserName = item.Edges.User.Username
	}
	return view
}

func isValidInvoiceType(value string) bool {
	switch value {
	case InvoiceTypeVATGeneral, InvoiceTypeVATSpecial:
		return true
	default:
		return false
	}
}

func isValidInvoiceStatus(value string) bool {
	switch value {
	case InvoiceStatusPending, InvoiceStatusApproved, InvoiceStatusRejected, InvoiceStatusIssued:
		return true
	default:
		return false
	}
}

func invoiceRoundAmount(value float64) float64 {
	return decimal.NewFromFloat(value).Round(2).InexactFloat64()
}

func invoiceRoundDecimal(value decimal.Decimal) float64 {
	return value.Round(2).InexactFloat64()
}

func invoiceAdminIDPtr(adminID int64) *int64 {
	if adminID <= 0 {
		return nil
	}
	return &adminID
}

func invoiceStorageDir() string {
	if dataDir := strings.TrimSpace(os.Getenv("DATA_DIR")); dataDir != "" {
		return filepath.Join(dataDir, defaultInvoiceStorageSubdir)
	}
	return filepath.Join("data", defaultInvoiceStorageSubdir)
}

func randomInvoiceStoredName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate invoice filename: %w", err)
	}
	return hex.EncodeToString(buf) + ext, nil
}

func sanitizeInvoiceOriginalFileName(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "invoice.pdf"
	}
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	if name == "" {
		return "invoice.pdf"
	}
	if len(name) > 255 {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		maxBase := 255 - len(ext)
		if maxBase < 1 {
			return name[:255]
		}
		if len(base) > maxBase {
			base = base[:maxBase]
		}
		name = base + ext
	}
	return strings.ReplaceAll(name, `"`, "")
}

func invoiceContentTypeAllowedForExt(ext, contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ext {
	case ".pdf":
		return contentType == "application/pdf"
	case ".ofd":
		return contentType == "application/ofd" || contentType == "application/octet-stream" || contentType == "application/vnd.ofd"
	case ".jpg", ".jpeg":
		return contentType == "image/jpeg" || contentType == "image/pjpeg"
	case ".png":
		return contentType == "image/png"
	default:
		return false
	}
}
