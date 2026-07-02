package service

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPaymentInvoiceSummaryUsesCNYBalancePaidAmountMinusRefundsAndReserved(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentService{entClient: client}
	user := createInvoiceTestUser(t, ctx, client, "summary")

	createInvoiceTestOrder(t, ctx, client, invoiceTestOrderInput{
		userID: user.ID, outTradeNo: "invoice-summary-completed", amount: 100, payAmount: 103,
		status: OrderStatusCompleted, orderType: payment.OrderTypeBalance,
	})
	createInvoiceTestOrder(t, ctx, client, invoiceTestOrderInput{
		userID: user.ID, outTradeNo: "invoice-summary-partial", amount: 100, payAmount: 103,
		refundAmount: 50, status: OrderStatusPartiallyRefunded, orderType: payment.OrderTypeBalance,
	})
	createInvoiceTestOrder(t, ctx, client, invoiceTestOrderInput{
		userID: user.ID, outTradeNo: "invoice-summary-subscription", amount: 99, payAmount: 99,
		status: OrderStatusCompleted, orderType: payment.OrderTypeSubscription,
	})
	createInvoiceTestOrder(t, ctx, client, invoiceTestOrderInput{
		userID: user.ID, outTradeNo: "invoice-summary-usd", amount: 10, payAmount: 10,
		status: OrderStatusCompleted, orderType: payment.OrderTypeBalance, currency: "USD",
	})
	createInvoiceTestInvoice(t, ctx, client, user.ID, 20, InvoiceStatusPending)
	createInvoiceTestInvoice(t, ctx, client, user.ID, 5, InvoiceStatusRejected)
	createInvoiceTestInvoice(t, ctx, client, user.ID, 10, InvoiceStatusIssued)

	got, err := svc.GetInvoiceSummary(ctx, user.ID)

	require.NoError(t, err)
	require.Equal(t, defaultInvoiceCurrency, got.Currency)
	require.Equal(t, 154.5, got.EligibleAmount)
	require.Equal(t, 30.0, got.RequestedAmount)
	require.Equal(t, 124.5, got.AvailableAmount)
}

func TestPaymentInvoiceCreateRejectsInvalidAndExcessAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentService{entClient: client}
	user := createInvoiceTestUser(t, ctx, client, "create")
	createInvoiceTestOrder(t, ctx, client, invoiceTestOrderInput{
		userID: user.ID, outTradeNo: "invoice-create-completed", amount: 100, payAmount: 100,
		status: OrderStatusCompleted, orderType: payment.OrderTypeBalance,
	})

	_, err := svc.CreateInvoiceRequest(ctx, InvoiceRequestCreateInput{
		UserID: user.ID, Amount: 0, InvoiceType: InvoiceTypeVATGeneral, Title: "ACME", TaxNumber: "TAX",
	})
	require.Equal(t, "INVALID_INVOICE_AMOUNT", infraerrors.Reason(err))

	_, err = svc.CreateInvoiceRequest(ctx, InvoiceRequestCreateInput{
		UserID: user.ID, Amount: 10, InvoiceType: "bad", Title: "ACME", TaxNumber: "TAX",
	})
	require.Equal(t, "INVALID_INVOICE_TYPE", infraerrors.Reason(err))

	_, err = svc.CreateInvoiceRequest(ctx, InvoiceRequestCreateInput{
		UserID: user.ID, Amount: 101, InvoiceType: InvoiceTypeVATGeneral, Title: "ACME", TaxNumber: "TAX",
	})
	require.Equal(t, "INVOICE_AMOUNT_EXCEEDED", infraerrors.Reason(err))

	item, err := svc.CreateInvoiceRequest(ctx, InvoiceRequestCreateInput{
		UserID: user.ID, Amount: 100, InvoiceType: InvoiceTypeVATSpecial, Title: "ACME", TaxNumber: "TAX", Remark: "remark",
	})
	require.NoError(t, err)
	require.Equal(t, 100.0, item.Amount)
	require.Equal(t, InvoiceStatusPending, item.Status)
	require.Equal(t, InvoiceTypeVATSpecial, item.InvoiceType)
}

func TestPaymentInvoiceDownloadRequiresOwnerIssuedAndFile(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)
	svc := &PaymentService{entClient: client}
	user := createInvoiceTestUser(t, ctx, client, "download-owner")
	other := createInvoiceTestUser(t, ctx, client, "download-other")
	req := createInvoiceTestInvoice(t, ctx, client, user.ID, 10, InvoiceStatusApproved)

	_, err := svc.GetInvoiceDownload(ctx, user.ID, req.ID)
	require.Equal(t, "INVOICE_NOT_ISSUED", infraerrors.Reason(err))

	issued, err := svc.AdminIssueInvoiceRequest(ctx, req.ID, InvoiceAdminIssueInput{
		AdminID: 1, FileName: "invoice.pdf", ContentType: "application/pdf", Reader: bytes.NewReader([]byte("%PDF-test")),
	})
	require.NoError(t, err)
	require.True(t, issued.Downloadable)
	require.True(t, issued.Claimable)
	require.Nil(t, issued.DownloadedAt)
	require.Equal(t, 0, issued.DownloadCount)

	claimSummary, err := svc.GetInvoiceClaimSummary(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1, claimSummary.ClaimableCount)

	_, err = svc.GetInvoiceDownload(ctx, other.ID, req.ID)
	require.Equal(t, "FORBIDDEN", infraerrors.Reason(err))

	file, err := svc.GetInvoiceDownload(ctx, user.ID, req.ID)
	require.NoError(t, err)
	require.Equal(t, "invoice.pdf", file.FileName)
	require.Equal(t, "application/pdf", file.ContentType)
	require.Equal(t, int64(9), file.Size)
	require.FileExists(t, file.Path)
	require.Contains(t, filepath.ToSlash(file.Path), "/invoices/")

	downloaded, err := client.InvoiceRequest.Get(ctx, req.ID)
	require.NoError(t, err)
	require.NotNil(t, downloaded.DownloadedAt)
	require.Equal(t, 1, downloaded.DownloadCount)
	firstDownloadedAt := *downloaded.DownloadedAt

	claimSummary, err = svc.GetInvoiceClaimSummary(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 0, claimSummary.ClaimableCount)

	_, err = svc.GetInvoiceDownload(ctx, user.ID, req.ID)
	require.NoError(t, err)
	downloadedAgain, err := client.InvoiceRequest.Get(ctx, req.ID)
	require.NoError(t, err)
	require.NotNil(t, downloadedAgain.DownloadedAt)
	require.Equal(t, firstDownloadedAt, *downloadedAgain.DownloadedAt)
	require.Equal(t, 2, downloadedAgain.DownloadCount)
}

func TestPaymentInvoiceIssueCreatesSystemTicketNotification(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	t.Setenv("DATA_DIR", t.TempDir())
	systemRepo := newFakeTicketRepo()
	svc := &PaymentService{entClient: client}
	svc.SetSystemTicketService(NewSystemTicketService(systemRepo))
	user := createInvoiceTestUser(t, ctx, client, "notify")
	req := createInvoiceTestInvoice(t, ctx, client, user.ID, 1999, InvoiceStatusApproved)

	issued, err := svc.AdminIssueInvoiceRequest(ctx, req.ID, InvoiceAdminIssueInput{
		AdminID: 1, InvoiceNo: "INV-20260701", FileName: "invoice.pdf", ContentType: "application/pdf", Reader: bytes.NewReader([]byte("%PDF-test")),
	})

	require.NoError(t, err)
	require.True(t, issued.Downloadable)
	got := requireSystemTicketNotification(t, systemRepo, user.ID, SystemTicketEventInvoiceIssued, "invoice_issued:"+strconv.FormatInt(req.ID, 10))
	require.Contains(t, got.Message.Content, "你的发票已开具")
	require.Equal(t, float64(req.ID), got.Metadata["invoice_request_id"])
	require.Equal(t, "INV-20260701", got.Metadata["invoice_no"])
	require.Equal(t, "invoice.pdf", got.Metadata["file_name"])
}

func TestPaymentInvoiceIssueRejectsInvalidFileTypeAndLargeFile(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	t.Setenv("DATA_DIR", t.TempDir())
	svc := &PaymentService{entClient: client}
	user := createInvoiceTestUser(t, ctx, client, "issue")
	req := createInvoiceTestInvoice(t, ctx, client, user.ID, 10, InvoiceStatusApproved)

	_, err := svc.AdminIssueInvoiceRequest(ctx, req.ID, InvoiceAdminIssueInput{
		AdminID: 1, FileName: "invoice.exe", ContentType: "application/octet-stream", Reader: bytes.NewReader([]byte("x")),
	})
	require.Equal(t, "INVALID_INVOICE_FILE_TYPE", infraerrors.Reason(err))

	large := bytes.NewReader(bytes.Repeat([]byte("a"), int(defaultInvoiceMaxUploadBytes+1)))
	_, err = svc.AdminIssueInvoiceRequest(ctx, req.ID, InvoiceAdminIssueInput{
		AdminID: 1, FileName: "invoice.pdf", ContentType: "application/pdf", Reader: large,
	})
	require.Equal(t, "INVOICE_FILE_TOO_LARGE", infraerrors.Reason(err))

	entries, readErr := os.ReadDir(filepath.Join(os.Getenv("DATA_DIR"), defaultInvoiceStorageSubdir))
	if readErr == nil {
		require.Empty(t, entries)
	}
}

func TestPaymentInvoiceSanitizesDownloadFileName(t *testing.T) {
	require.Equal(t, "invoice.pdf", sanitizeInvoiceOriginalFileName("../invoice.pdf"))
	require.Equal(t, "invoice.pdf", sanitizeInvoiceOriginalFileName("..\\invoice.pdf"))
	require.Equal(t, "badinvoice.pdf", sanitizeInvoiceOriginalFileName("bad\r\ninvoice.pdf"))
	require.Equal(t, "invoice.pdf", sanitizeInvoiceOriginalFileName("\r\n"))
	require.NotContains(t, sanitizeInvoiceOriginalFileName("bad\r\ninvoice.pdf"), "\r")
	require.NotContains(t, sanitizeInvoiceOriginalFileName("bad\r\ninvoice.pdf"), "\n")
}

type invoiceTestOrderInput struct {
	userID       int64
	outTradeNo   string
	amount       float64
	payAmount    float64
	refundAmount float64
	status       string
	orderType    string
	currency     string
}

func createInvoiceTestUser(t *testing.T, ctx context.Context, client *dbent.Client, suffix string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail("invoice-" + suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername("invoice-" + suffix).
		Save(ctx)
	require.NoError(t, err)
	return user
}

func createInvoiceTestOrder(t *testing.T, ctx context.Context, client *dbent.Client, input invoiceTestOrderInput) *dbent.PaymentOrder {
	t.Helper()
	currency := input.currency
	if currency == "" {
		currency = payment.DefaultPaymentCurrency
	}
	orderType := input.orderType
	if orderType == "" {
		orderType = payment.OrderTypeBalance
	}
	status := input.status
	if status == "" {
		status = OrderStatusCompleted
	}
	order, err := client.PaymentOrder.Create().
		SetUserID(input.userID).
		SetUserEmail("invoice-order@example.com").
		SetUserName("invoice-order").
		SetAmount(input.amount).
		SetPayAmount(input.payAmount).
		SetFeeRate(0).
		SetRechargeCode("INV-" + input.outTradeNo).
		SetOutTradeNo(input.outTradeNo).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-" + input.outTradeNo).
		SetOrderType(orderType).
		SetStatus(status).
		SetRefundAmount(input.refundAmount).
		SetProviderSnapshot(map[string]any{
			"schema_version": 2,
			"currency":       currency,
		}).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	return order
}

func createInvoiceTestInvoice(t *testing.T, ctx context.Context, client *dbent.Client, userID int64, amount float64, status string) *dbent.InvoiceRequest {
	t.Helper()
	item, err := client.InvoiceRequest.Create().
		SetUserID(userID).
		SetAmount(amount).
		SetCurrency(payment.DefaultPaymentCurrency).
		SetInvoiceType(InvoiceTypeVATGeneral).
		SetTitle("ACME").
		SetTaxNumber("TAX").
		SetStatus(status).
		Save(ctx)
	require.NoError(t, err)
	return item
}
