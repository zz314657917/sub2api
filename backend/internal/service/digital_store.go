package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const digitalStoreFileMaxBytes int64 = 20 << 20

// DigitalStoreService owns catalogue and encrypted data. Payment creation and
// fulfillment use this service later, but are deliberately not coupled here.
type DigitalStoreService struct {
	entClient  *dbent.Client
	settingSvc *SettingService
	encryptor  SecretEncryptor
	paymentSvc *PaymentService
}

func NewDigitalStoreService(client *dbent.Client, settings *SettingService, encryptor SecretEncryptor, paymentSvc *PaymentService) *DigitalStoreService {
	return &DigitalStoreService{entClient: client, settingSvc: settings, encryptor: encryptor, paymentSvc: paymentSvc}
}

func (s *DigitalStoreService) queryOne(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := s.entClient.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	return rows, nil
}

type DigitalStoreProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	PriceCents  int64  `json:"price_cents"`
	Enabled     bool   `json:"enabled"`
	FileID      *int64 `json:"file_id,omitempty"`
}
type DigitalStoreProduct struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Kind           string `json:"kind"`
	PriceCents     int64  `json:"price_cents"`
	StockAvailable int64  `json:"stock_available"`
	Enabled        bool   `json:"enabled"`
	// FileID is populated only for administrator catalogue reads. Public
	// handlers must map their response explicitly and never expose it.
	FileID    *int64    `json:"file_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
type DigitalStoreFile struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}
type DigitalStoreStockInput struct {
	Content  string `json:"content"`
	Username string `json:"username"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}
type DigitalStoreOrder struct {
	OrderID        int64     `json:"order_id"`
	ProductID      int64     `json:"product_id"`
	ProductName    string    `json:"product_name"`
	Kind           string    `json:"kind"`
	PriceCents     int64     `json:"price_cents"`
	PaymentStatus  string    `json:"payment_status"`
	DeliveryStatus string    `json:"delivery_status"`
	CreatedAt      time.Time `json:"created_at"`
}
type DigitalStoreAdminOrder struct {
	DigitalStoreOrder
	UserID int64 `json:"user_id"`
}
type DigitalStoreStock struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}
type DigitalStoreDelivery struct {
	Kind        string `json:"kind"`
	Content     string `json:"content,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Notes       string `json:"notes,omitempty"`
	Filename    string `json:"filename,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

func (s *DigitalStoreService) requireEnabled(ctx context.Context) error {
	if s == nil || s.entClient == nil || s.settingSvc == nil {
		return infraerrors.ServiceUnavailable("STORE_UNAVAILABLE", "store is not configured")
	}
	settings, err := s.settingSvc.GetAllSettings(ctx)
	if err != nil {
		return infraerrors.ServiceUnavailable("STORE_UNAVAILABLE", "store setting is unavailable")
	}
	if !settings.ServiceStoreEnabled {
		return infraerrors.Forbidden("STORE_DISABLED", "store is disabled")
	}
	return nil
}
func validateDigitalStoreProduct(in DigitalStoreProductInput) error {
	n, d, k := strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), strings.TrimSpace(in.Kind)
	if n == "" || len([]rune(n)) > 120 || len([]rune(d)) > 4000 {
		return infraerrors.BadRequest("INVALID_STORE_PRODUCT", "invalid product name or description")
	}
	if k != "card" && k != "account" && k != "file" {
		return infraerrors.BadRequest("INVALID_STORE_PRODUCT", "invalid product kind")
	}
	if in.PriceCents <= 0 || in.PriceCents > 100000000 {
		return infraerrors.BadRequest("INVALID_STORE_PRODUCT", "invalid price_cents")
	}
	if (k == "file") != (in.FileID != nil && *in.FileID > 0) {
		return infraerrors.BadRequest("INVALID_STORE_PRODUCT", "file products require file_id only")
	}
	return nil
}
func (s *DigitalStoreService) AdminCreateProduct(ctx context.Context, in DigitalStoreProductInput) (*DigitalStoreProduct, error) {
	if err := validateDigitalStoreProduct(in); err != nil {
		return nil, err
	}
	if in.Kind == "file" {
		if err := s.requireStoreFile(ctx, *in.FileID); err != nil {
			return nil, infraerrors.NotFound("STORE_FILE_NOT_FOUND", "store file not found")
		}
	}
	var out DigitalStoreProduct
	rows, err := s.queryOne(ctx, `INSERT INTO store_products (name,description,kind,price_cents,enabled,file_id) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,name,description,kind,price_cents,enabled,file_id,created_at`, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), in.Kind, in.PriceCents, in.Enabled, in.FileID)
	if err == nil {
		err = rows.Scan(&out.ID, &out.Name, &out.Description, &out.Kind, &out.PriceCents, &out.Enabled, &out.FileID, &out.CreatedAt)
		rows.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("create store product: %w", err)
	}
	return &out, nil
}
func (s *DigitalStoreService) ListProducts(ctx context.Context, includeDisabled bool) ([]DigitalStoreProduct, error) {
	items, _, err := s.ListProductsPage(ctx, includeDisabled, "", "", 1, 100)
	return items, err
}

// ListProductsPage returns catalogue records without any inventory plaintext.
// includeDisabled is reserved for administrator calls; public calls must pass
// false, which also enforces the Store feature flag.
func (s *DigitalStoreService) ListProductsPage(ctx context.Context, includeDisabled bool, search, kind string, page, pageSize int) ([]DigitalStoreProduct, int64, error) {
	if !includeDisabled {
		if err := s.requireEnabled(ctx); err != nil {
			return nil, 0, err
		}
	}
	page, pageSize, err := normalizeStorePage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	search = strings.TrimSpace(search)
	kind = strings.TrimSpace(kind)
	if kind != "" && kind != "card" && kind != "account" && kind != "file" {
		return nil, 0, infraerrors.BadRequest("INVALID_STORE_KIND", "invalid store product kind")
	}
	where, args := storeProductListWhere(includeDisabled, search, kind)
	var total int64
	if err := storeQueryRow(ctx, s.entClient, "SELECT COUNT(*) FROM store_products p"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	q := `SELECT p.id,p.name,p.description,p.kind,p.price_cents,p.enabled,p.file_id,p.created_at,CASE WHEN p.kind='file' THEN 1 ELSE (SELECT COUNT(*) FROM store_inventory i WHERE i.product_id=p.id AND i.state='available') END FROM store_products p` + where + " ORDER BY p.id DESC LIMIT $" + fmt.Sprint(len(args)-1) + " OFFSET $" + fmt.Sprint(len(args))
	rows, err := s.entClient.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []DigitalStoreProduct{}
	for rows.Next() {
		var x DigitalStoreProduct
		if err := rows.Scan(&x.ID, &x.Name, &x.Description, &x.Kind, &x.PriceCents, &x.Enabled, &x.FileID, &x.CreatedAt, &x.StockAvailable); err != nil {
			return nil, 0, err
		}
		out = append(out, x)
	}
	return out, total, rows.Err()
}
func (s *DigitalStoreService) AdminUpdateProduct(ctx context.Context, id int64, in DigitalStoreProductInput) (*DigitalStoreProduct, error) {
	if id <= 0 {
		return nil, infraerrors.BadRequest("INVALID_ID", "invalid product id")
	}
	if err := validateDigitalStoreProduct(in); err != nil {
		return nil, err
	}
	var kind string
	rows, err := s.queryOne(ctx, "SELECT kind FROM store_products WHERE id=$1", id)
	if err == nil {
		err = rows.Scan(&kind)
		rows.Close()
	}
	if err != nil {
		return nil, infraerrors.NotFound("STORE_PRODUCT_NOT_FOUND", "store product not found")
	}
	if kind != in.Kind {
		return nil, infraerrors.BadRequest("STORE_PRODUCT_KIND_IMMUTABLE", "product kind cannot change")
	}
	if in.Kind == "file" {
		if err := s.requireStoreFile(ctx, *in.FileID); err != nil {
			return nil, infraerrors.NotFound("STORE_FILE_NOT_FOUND", "store file not found")
		}
	}
	var out DigitalStoreProduct
	// Store orders snapshot their file_id at checkout, so changing a file
	// product only affects future orders and keeps historical deliveries intact.
	rows, err = s.queryOne(ctx, `UPDATE store_products SET name=$2,description=$3,price_cents=$4,enabled=$5,file_id=$6,updated_at=NOW() WHERE id=$1 RETURNING id,name,description,kind,price_cents,enabled,file_id,created_at`, id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), in.PriceCents, in.Enabled, in.FileID)
	if err == nil {
		err = rows.Scan(&out.ID, &out.Name, &out.Description, &out.Kind, &out.PriceCents, &out.Enabled, &out.FileID, &out.CreatedAt)
		rows.Close()
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *DigitalStoreService) AdminImportStock(ctx context.Context, productID int64, items []DigitalStoreStockInput) (int, error) {
	if productID <= 0 || len(items) == 0 || len(items) > 500 {
		return 0, infraerrors.BadRequest("INVALID_STORE_STOCK", "stock items must contain 1..500 entries")
	}
	var kind string
	rows, err := s.queryOne(ctx, "SELECT kind FROM store_products WHERE id=$1", productID)
	if err == nil {
		err = rows.Scan(&kind)
		rows.Close()
	}
	if err != nil {
		return 0, infraerrors.NotFound("STORE_PRODUCT_NOT_FOUND", "store product not found")
	}
	if kind == "file" {
		return 0, infraerrors.BadRequest("INVALID_STORE_STOCK", "file products have no inventory")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for _, item := range items {
		var raw string
		if kind == "card" {
			raw = strings.TrimSpace(item.Content)
			if raw == "" || len(raw) > 8192 {
				return 0, infraerrors.BadRequest("INVALID_STORE_STOCK", "card content must contain 1..8192 bytes")
			}
		} else {
			if strings.TrimSpace(item.Username) == "" || strings.TrimSpace(item.Password) == "" {
				return 0, infraerrors.BadRequest("INVALID_STORE_STOCK", "account username and password are required")
			}
			if strings.ContainsAny(item.Username+item.Password, "\r\n") || len(item.Username) > 512 || len(item.Password) > 2048 || len(item.Notes) > 4096 {
				return 0, infraerrors.BadRequest("INVALID_STORE_STOCK", "invalid account fields")
			}
			encoded, marshalErr := json.Marshal(struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Notes    string `json:"notes"`
			}{item.Username, item.Password, item.Notes})
			if marshalErr != nil {
				return 0, marshalErr
			}
			raw = string(encoded)
		}
		encrypted, e := s.encryptor.Encrypt(raw)
		if e != nil {
			return 0, fmt.Errorf("encrypt store inventory: %w", e)
		}
		if _, e = tx.ExecContext(ctx, "INSERT INTO store_inventory(product_id,kind,encrypted_content) VALUES($1,$2,$3)", productID, kind, encrypted); e != nil {
			return 0, e
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return len(items), nil
}
func (s *DigitalStoreService) UploadFile(ctx context.Context, filename string, body []byte) (*DigitalStoreFile, error) {
	if int64(len(body)) > digitalStoreFileMaxBytes {
		return nil, infraerrors.BadRequest("STORE_FILE_TOO_LARGE", "file exceeds 20 MiB")
	}
	filename = filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if strings.ContainsAny(filename, "\r\n\x00") {
		return nil, infraerrors.BadRequest("INVALID_STORE_FILE", "invalid filename")
	}
	if filename == "." || filename == "" {
		return nil, infraerrors.BadRequest("INVALID_STORE_FILE", "filename is required")
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	encrypted, err := s.encryptor.Encrypt(encoded)
	if err != nil {
		return nil, fmt.Errorf("encrypt store file: %w", err)
	}
	sum := sha256.Sum256(body)
	out := &DigitalStoreFile{Filename: filename, Size: int64(len(body))}
	rows, err := s.queryOne(ctx, "INSERT INTO store_files(filename,encrypted_content,content_size,content_sha256) VALUES($1,$2,$3,$4) RETURNING id", filename, encrypted, out.Size, hex.EncodeToString(sum[:]))
	if err == nil {
		err = rows.Scan(&out.ID)
		rows.Close()
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *DigitalStoreService) ListOrders(ctx context.Context, userID int64) ([]DigitalStoreOrder, error) {
	items, _, err := s.ListOrdersPage(ctx, userID, 1, 100)
	return items, err
}

func (s *DigitalStoreService) ListOrdersPage(ctx context.Context, userID int64, page, pageSize int) ([]DigitalStoreOrder, int64, error) {
	if userID <= 0 {
		return nil, 0, infraerrors.Unauthorized("UNAUTHORIZED", "user is required")
	}
	page, pageSize, err := normalizeStorePage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := storeQueryRow(ctx, s.entClient, `SELECT COUNT(*) FROM store_orders WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.entClient.QueryContext(ctx, `SELECT so.payment_order_id,so.product_id,so.product_name,so.kind,so.price_cents,po.status,so.delivery_status,so.created_at FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id WHERE so.user_id=$1 ORDER BY so.created_at DESC LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []DigitalStoreOrder{}
	for rows.Next() {
		var x DigitalStoreOrder
		if err := rows.Scan(&x.OrderID, &x.ProductID, &x.ProductName, &x.Kind, &x.PriceCents, &x.PaymentStatus, &x.DeliveryStatus, &x.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, x)
	}
	return out, total, rows.Err()
}

// AdminListOrders supplies management metadata only. It deliberately omits
// delivery and inventory columns, including all encrypted values.
func (s *DigitalStoreService) AdminListOrders(ctx context.Context, status string, userID, productID int64, page, pageSize int) ([]DigitalStoreAdminOrder, int64, error) {
	page, pageSize, err := normalizeStorePage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	status = strings.TrimSpace(status)
	where := " WHERE 1=1"
	args := make([]any, 0, 3)
	if status != "" {
		args = append(args, status)
		where += " AND po.status=$" + fmt.Sprint(len(args))
	}
	if userID > 0 {
		args = append(args, userID)
		where += " AND so.user_id=$" + fmt.Sprint(len(args))
	}
	if productID > 0 {
		args = append(args, productID)
		where += " AND so.product_id=$" + fmt.Sprint(len(args))
	}
	var total int64
	if err := storeQueryRow(ctx, s.entClient, "SELECT COUNT(*) FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	q := `SELECT so.payment_order_id,so.product_id,so.product_name,so.kind,so.price_cents,po.status,so.delivery_status,so.created_at,so.user_id FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id` + where + " ORDER BY so.created_at DESC LIMIT $" + fmt.Sprint(len(args)-1) + " OFFSET $" + fmt.Sprint(len(args))
	rows, err := s.entClient.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []DigitalStoreAdminOrder{}
	for rows.Next() {
		var x DigitalStoreAdminOrder
		if err := rows.Scan(&x.OrderID, &x.ProductID, &x.ProductName, &x.Kind, &x.PriceCents, &x.PaymentStatus, &x.DeliveryStatus, &x.CreatedAt, &x.UserID); err != nil {
			return nil, 0, err
		}
		out = append(out, x)
	}
	return out, total, rows.Err()
}

// AdminListStock is intentionally a masked index: no encrypted or plaintext
// inventory content is selected, even for administrators.
func (s *DigitalStoreService) AdminListStock(ctx context.Context, productID int64, page, pageSize int) ([]DigitalStoreStock, int64, error) {
	if productID <= 0 {
		return nil, 0, infraerrors.BadRequest("INVALID_ID", "invalid product id")
	}
	page, pageSize, err := normalizeStorePage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := storeQueryRow(ctx, s.entClient, `SELECT COUNT(*) FROM store_inventory WHERE product_id=$1`, productID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.entClient.QueryContext(ctx, `SELECT id,kind,state,created_at FROM store_inventory WHERE product_id=$1 ORDER BY id DESC LIMIT $2 OFFSET $3`, productID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []DigitalStoreStock{}
	for rows.Next() {
		var x DigitalStoreStock
		if err := rows.Scan(&x.ID, &x.Kind, &x.State, &x.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, x)
	}
	return out, total, rows.Err()
}

func (s *DigitalStoreService) requireStoreFile(ctx context.Context, id int64) error {
	if id <= 0 {
		return sql.ErrNoRows
	}
	var found int
	return storeQueryRow(ctx, s.entClient, "SELECT 1 FROM store_files WHERE id=$1", id).Scan(&found)
}

func normalizeStorePage(page, pageSize int) (int, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		return 0, 0, infraerrors.BadRequest("INVALID_PAGE_SIZE", "page_size must be between 1 and 100")
	}
	return page, pageSize, nil
}

func storeProductListWhere(includeDisabled bool, search, kind string) (string, []any) {
	where, args := " WHERE 1=1", make([]any, 0, 3)
	if !includeDisabled {
		where += " AND p.enabled=true"
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += " AND (p.name ILIKE $" + fmt.Sprint(len(args)) + " OR p.description ILIKE $" + fmt.Sprint(len(args)) + ")"
	}
	if kind != "" {
		args = append(args, kind)
		where += " AND p.kind=$" + fmt.Sprint(len(args))
	}
	return where, args
}
func (s *DigitalStoreService) GetOrder(ctx context.Context, userID, orderID int64) (*DigitalStoreOrder, error) {
	if userID <= 0 || orderID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_ID", "invalid order id")
	}
	rows, err := s.queryOne(ctx, `SELECT so.payment_order_id,so.product_id,so.product_name,so.kind,so.price_cents,po.status,so.delivery_status,so.created_at FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id WHERE so.user_id=$1 AND so.payment_order_id=$2`, userID, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
	}
	defer rows.Close()
	var out DigitalStoreOrder
	if err = rows.Scan(&out.OrderID, &out.ProductID, &out.ProductName, &out.Kind, &out.PriceCents, &out.PaymentStatus, &out.DeliveryStatus, &out.CreatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *DigitalStoreService) GetDelivery(ctx context.Context, userID, orderID int64) (*DigitalStoreDelivery, error) {
	rows, err := s.queryOne(ctx, `SELECT so.kind,so.encrypted_delivery,so.delivery_status,po.status,f.filename FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id LEFT JOIN store_files f ON f.id=so.file_id WHERE so.user_id=$1 AND so.payment_order_id=$2`, userID, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
	}
	defer rows.Close()
	var kind, status, paymentStatus string
	var encrypted, filename *string
	if err = rows.Scan(&kind, &encrypted, &status, &paymentStatus, &filename); err != nil {
		return nil, err
	}
	if status != "delivered" || paymentStatus != "COMPLETED" {
		return nil, infraerrors.Forbidden("STORE_DELIVERY_UNAVAILABLE", "delivery is not available")
	}
	if kind == "file" {
		if filename == nil || strings.TrimSpace(*filename) == "" {
			return nil, infraerrors.InternalServer("STORE_FILE_INVALID", "file is unavailable")
		}
		return &DigitalStoreDelivery{Kind: kind, Filename: *filename, DownloadURL: fmt.Sprintf("/api/v1/store/orders/%d/download", orderID)}, nil
	}
	if encrypted == nil {
		return nil, infraerrors.Forbidden("STORE_DELIVERY_UNAVAILABLE", "delivery is not available")
	}
	raw, err := s.encryptor.Decrypt(*encrypted)
	if err != nil {
		return nil, infraerrors.InternalServer("STORE_DELIVERY_DECRYPT_FAILED", "delivery is unavailable")
	}
	out := &DigitalStoreDelivery{Kind: kind}
	if kind == "card" {
		out.Content = raw
	} else if kind == "account" {
		var account struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Notes    string `json:"notes"`
		}
		if err := json.Unmarshal([]byte(raw), &account); err != nil {
			return nil, infraerrors.InternalServer("STORE_DELIVERY_INVALID", "delivery is unavailable")
		}
		out.Username, out.Password, out.Notes = account.Username, account.Password, account.Notes
	} else {
		return nil, infraerrors.BadRequest("STORE_DELIVERY_DOWNLOAD_REQUIRED", "file delivery requires download")
	}
	return out, nil
}
func (s *DigitalStoreService) GetDownload(ctx context.Context, userID, orderID int64) (*DigitalStoreFile, []byte, error) {
	rows, err := s.queryOne(ctx, `SELECT f.id,f.filename,f.encrypted_content,f.content_size,f.content_sha256 FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id JOIN store_files f ON f.id=so.file_id WHERE so.user_id=$1 AND so.payment_order_id=$2 AND so.kind='file' AND so.delivery_status='delivered' AND po.status='COMPLETED'`, userID, orderID)
	if err != nil {
		return nil, nil, infraerrors.NotFound("STORE_DOWNLOAD_NOT_FOUND", "file delivery not found")
	}
	defer rows.Close()
	var out DigitalStoreFile
	var encrypted, hash string
	if err = rows.Scan(&out.ID, &out.Filename, &encrypted, &out.Size, &hash); err != nil {
		return nil, nil, err
	}
	encoded, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, nil, infraerrors.InternalServer("STORE_FILE_DECRYPT_FAILED", "file is unavailable")
	}
	body, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || int64(len(body)) != out.Size {
		return nil, nil, infraerrors.InternalServer("STORE_FILE_INVALID", "file is unavailable")
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != hash {
		return nil, nil, infraerrors.InternalServer("STORE_FILE_INVALID", "file is unavailable")
	}
	return &out, body, nil
}
