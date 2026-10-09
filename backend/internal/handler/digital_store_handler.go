package handler

import (
	"context"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const digitalStoreMultipartLimit int64 = (20 << 20) + (1 << 20)

type DigitalStoreHandler struct {
	store          digitalStoreHTTPService
	paymentService *service.PaymentService
}

func NewDigitalStoreHandler(store *service.DigitalStoreService, paymentService *service.PaymentService) *DigitalStoreHandler {
	return &DigitalStoreHandler{store: store, paymentService: paymentService}
}

type digitalStoreHTTPService interface {
	ListProductsPage(ctx context.Context, includeDisabled bool, search, kind string, page, pageSize int) ([]service.DigitalStoreProduct, int64, error)
	CreateStoreOrder(ctx context.Context, input service.DigitalStoreCreateOrderInput) (*service.CreateOrderResponse, error)
	ListOrdersPage(ctx context.Context, userID int64, page, pageSize int) ([]service.DigitalStoreOrder, int64, error)
	GetOrder(ctx context.Context, userID, orderID int64) (*service.DigitalStoreOrder, error)
	GetDelivery(ctx context.Context, userID, orderID int64) (*service.DigitalStoreDelivery, error)
	GetDownload(ctx context.Context, userID, orderID int64) (*service.DigitalStoreFile, []byte, error)
	ResumeStoreOrder(ctx context.Context, userID, paymentOrderID int64) (*service.CreateOrderResponse, error)
	AdminCreateProduct(ctx context.Context, in service.DigitalStoreProductInput) (*service.DigitalStoreProduct, error)
	AdminUpdateProduct(ctx context.Context, id int64, in service.DigitalStoreProductInput) (*service.DigitalStoreProduct, error)
	UploadFile(ctx context.Context, filename string, body []byte) (*service.DigitalStoreFile, error)
	AdminListStock(ctx context.Context, productID int64, page, pageSize int) ([]service.DigitalStoreStock, int64, error)
	AdminImportStock(ctx context.Context, productID int64, items []service.DigitalStoreStockInput) (int, error)
	AdminListOrders(ctx context.Context, status string, userID, productID int64, page, pageSize int) ([]service.DigitalStoreAdminOrder, int64, error)
	RetryStoreFulfillment(ctx context.Context, paymentOrderID int64) error
}

type createDigitalStoreOrderRequest struct {
	ProductID         int64  `json:"product_id" binding:"required"`
	IdempotencyKey    string `json:"idempotency_key"`
	PaymentType       string `json:"payment_type"`
	OpenID            string `json:"openid"`
	ReturnURL         string `json:"return_url"`
	PaymentSource     string `json:"payment_source"`
	IsMobile          *bool  `json:"is_mobile,omitempty"`
	WechatResumeToken string `json:"wechat_resume_token"`
}

func (h *DigitalStoreHandler) ListProducts(c *gin.Context) {
	page, pageSize := storePagination(c)
	items, total, err := h.store.ListProductsPage(c.Request.Context(), false, c.Query("search"), c.Query("kind"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	public := make([]gin.H, 0, len(items))
	for _, item := range items {
		public = append(public, gin.H{"id": item.ID, "name": item.Name, "description": item.Description, "kind": item.Kind, "price_cents": item.PriceCents, "stock_available": item.StockAvailable, "enabled": item.Enabled, "created_at": item.CreatedAt})
	}
	response.Paginated(c, public, total, page, pageSize)
}

func (h *DigitalStoreHandler) CreateOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req createDigitalStoreOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	productID := req.ProductID
	if productID <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_STORE_ORDER", "invalid product_id"))
		return
	}
	if strings.TrimSpace(req.WechatResumeToken) != "" {
		claims, err := h.paymentService.ParseWeChatPaymentResumeToken(req.WechatResumeToken)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if claims.OrderType != payment.OrderTypeStore || claims.StoreUserID != subject.UserID || claims.StoreProductID != productID {
			response.ErrorFrom(c, infraerrors.Forbidden("INVALID_WECHAT_PAYMENT_RESUME_TOKEN", "wechat store checkout context does not match the authenticated user or product"))
			return
		}
		req.IdempotencyKey = claims.StoreIdempotencyKey
		req.OpenID = claims.OpenID
		req.PaymentType = claims.PaymentType
	}
	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}
	result, err := h.store.CreateStoreOrder(c.Request.Context(), service.DigitalStoreCreateOrderInput{UserID: subject.UserID, ProductID: productID, IdempotencyKey: req.IdempotencyKey, PaymentType: req.PaymentType, OpenID: req.OpenID, ClientIP: c.ClientIP(), IsMobile: mobile, IsWeChatBrowser: isWeChatBrowser(c), SrcHost: c.Request.Host, SrcURL: c.Request.Referer(), ReturnURL: req.ReturnURL, PaymentSource: req.PaymentSource, Locale: c.GetHeader("Accept-Language")})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *DigitalStoreHandler) ListOrders(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	page, pageSize := storePagination(c)
	items, total, err := h.store.ListOrdersPage(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
func (h *DigitalStoreHandler) GetOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := storeID(c)
	if !ok {
		return
	}
	item, err := h.store.GetOrder(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *DigitalStoreHandler) GetDelivery(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := storeID(c)
	if !ok {
		return
	}
	item, err := h.store.GetDelivery(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, item)
}
func (h *DigitalStoreHandler) Download(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := storeID(c)
	if !ok {
		return
	}
	file, body, err := h.store.GetDownload(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": storeSafeFilename(file.Filename)}))
	c.Data(http.StatusOK, "application/octet-stream", body)
}
func (h *DigitalStoreHandler) ResumeOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := storeID(c)
	if !ok {
		return
	}
	result, err := h.store.ResumeStoreOrder(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *DigitalStoreHandler) AdminListProducts(c *gin.Context) {
	page, pageSize := storePagination(c)
	items, total, err := h.store.ListProductsPage(c.Request.Context(), true, c.Query("search"), c.Query("kind"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
func (h *DigitalStoreHandler) AdminCreateProduct(c *gin.Context) {
	var req service.DigitalStoreProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	item, err := h.store.AdminCreateProduct(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}
func (h *DigitalStoreHandler) AdminUpdateProduct(c *gin.Context) {
	id, ok := storeID(c)
	if !ok {
		return
	}
	var req service.DigitalStoreProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	item, err := h.store.AdminUpdateProduct(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *DigitalStoreHandler) AdminUploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, digitalStoreMultipartLimit)
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	file, err := header.Open()
	if err != nil {
		response.BadRequest(c, "file is unavailable")
		return
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, (20<<20)+1))
	if err != nil {
		response.BadRequest(c, "file is unavailable")
		return
	}
	if int64(len(body)) > 20<<20 {
		response.ErrorFrom(c, infraerrors.BadRequest("STORE_FILE_TOO_LARGE", "file exceeds 20 MiB"))
		return
	}
	item, err := h.store.UploadFile(c.Request.Context(), header.Filename, body)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}
func (h *DigitalStoreHandler) AdminListStock(c *gin.Context) {
	id, ok := storeID(c)
	if !ok {
		return
	}
	page, pageSize := storePagination(c)
	items, total, err := h.store.AdminListStock(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
func (h *DigitalStoreHandler) AdminImportStock(c *gin.Context) {
	id, ok := storeID(c)
	if !ok {
		return
	}
	var req struct {
		Items []service.DigitalStoreStockInput `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	n, err := h.store.AdminImportStock(c.Request.Context(), id, req.Items)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"imported": n})
}
func (h *DigitalStoreHandler) AdminListOrders(c *gin.Context) {
	page, pageSize := storePagination(c)
	userID, err := storeOptionalID(c, "user_id")
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}
	productID, err := storeOptionalID(c, "product_id")
	if err != nil {
		response.BadRequest(c, "invalid product_id")
		return
	}
	items, total, err := h.store.AdminListOrders(c.Request.Context(), c.Query("status"), userID, productID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
func (h *DigitalStoreHandler) AdminRetryOrder(c *gin.Context) {
	id, ok := storeID(c)
	if !ok {
		return
	}
	if err := h.store.RetryStoreFulfillment(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "fulfillment retried"})
}

func storeID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return 0, false
	}
	return id, true
}
func storeOptionalID(c *gin.Context, key string) (int64, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", "invalid "+key)
	}
	return id, nil
}
func storePagination(c *gin.Context) (int, int) {
	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
func storeSafeFilename(filename string) string {
	filename = filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	filename = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == '"' || r == '\\' {
			return -1
		}
		return r
	}, filename)
	if filename == "" || filename == "." {
		return "download"
	}
	return filename
}
