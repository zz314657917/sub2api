package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type digitalStoreHTTPStub struct {
	gotUserID int64
	file      *service.DigitalStoreFile
	body      []byte
}

func (s *digitalStoreHTTPStub) ListProductsPage(context.Context, bool, string, string, int, int) ([]service.DigitalStoreProduct, int64, error) {
	return nil, 0, nil
}
func (s *digitalStoreHTTPStub) CreateStoreOrder(context.Context, service.DigitalStoreCreateOrderInput) (*service.CreateOrderResponse, error) {
	return nil, errors.New("unexpected checkout")
}
func (s *digitalStoreHTTPStub) ListOrdersPage(context.Context, int64, int, int) ([]service.DigitalStoreOrder, int64, error) {
	return nil, 0, nil
}
func (s *digitalStoreHTTPStub) GetOrder(context.Context, int64, int64) (*service.DigitalStoreOrder, error) {
	return nil, errors.New("not found")
}
func (s *digitalStoreHTTPStub) GetDelivery(_ context.Context, userID, _ int64) (*service.DigitalStoreDelivery, error) {
	s.gotUserID = userID
	return nil, errors.New("not found")
}
func (s *digitalStoreHTTPStub) GetDownload(_ context.Context, userID, _ int64) (*service.DigitalStoreFile, []byte, error) {
	s.gotUserID = userID
	return s.file, s.body, nil
}
func (s *digitalStoreHTTPStub) ResumeStoreOrder(_ context.Context, userID, _ int64) (*service.CreateOrderResponse, error) {
	s.gotUserID = userID
	return nil, errors.New("not found")
}
func (s *digitalStoreHTTPStub) AdminCreateProduct(context.Context, service.DigitalStoreProductInput) (*service.DigitalStoreProduct, error) {
	return nil, nil
}
func (s *digitalStoreHTTPStub) AdminUpdateProduct(context.Context, int64, service.DigitalStoreProductInput) (*service.DigitalStoreProduct, error) {
	return nil, nil
}
func (s *digitalStoreHTTPStub) UploadFile(context.Context, string, []byte) (*service.DigitalStoreFile, error) {
	return nil, nil
}
func (s *digitalStoreHTTPStub) AdminListStock(context.Context, int64, int, int) ([]service.DigitalStoreStock, int64, error) {
	return nil, 0, nil
}
func (s *digitalStoreHTTPStub) AdminImportStock(context.Context, int64, []service.DigitalStoreStockInput) (int, error) {
	return 0, nil
}
func (s *digitalStoreHTTPStub) AdminListOrders(context.Context, string, int64, int64, int, int) ([]service.DigitalStoreAdminOrder, int64, error) {
	return nil, 0, nil
}
func (s *digitalStoreHTTPStub) RetryStoreFulfillment(context.Context, int64) error { return nil }

func TestStoreSafeFilenameRemovesHeaderControlCharacters(t *testing.T) {
	if got := storeSafeFilename("payload\"\r\nX-Injected: yes"); got != "payloadX-Injected: yes" {
		t.Fatalf("storeSafeFilename() = %q", got)
	}
	if got := storeSafeFilename(" "); got != "download" {
		t.Fatalf("empty filename = %q", got)
	}
}

func TestDigitalStoreHandlerRejectsUnauthenticatedDelivery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/store/orders/:id/delivery", (&DigitalStoreHandler{}).GetDelivery)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/store/orders/9/delivery", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestDigitalStoreHandlerRejectsMismatchedWeChatStoreClaimsBeforeCheckout(t *testing.T) {
	t.Setenv("PAYMENT_RESUME_SIGNING_KEY", "digital-store-handler-test-signing-key")
	paymentSvc := service.NewPaymentService(nil, payment.NewRegistry(), nil, nil, nil, nil, nil, nil, nil)
	h := NewDigitalStoreHandler(nil, paymentSvc)
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ claimUser, claimProduct int64 }{{2, 9}, {1, 8}} {
		token, err := service.NewPaymentResumeService([]byte("digital-store-handler-test-signing-key")).CreateWeChatPaymentResumeToken(service.WeChatPaymentResumeClaims{
			OpenID: "openid", PaymentType: payment.TypeWxpay, OrderType: payment.OrderTypeStore,
			StoreUserID: tc.claimUser, StoreProductID: tc.claimProduct, StoreIdempotencyKey: "store-key",
		})
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
			c.Next()
		})
		router.POST("/store/orders", h.CreateOrder)
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/store/orders", strings.NewReader(`{"product_id":9,"wechat_resume_token":"`+token+`"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", recorder.Code, recorder.Body.String())
		}
	}
}

func TestDigitalStoreHandlerUsesAuthenticatedOwnerAndSafeDownloadHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &digitalStoreHTTPStub{file: &service.DigitalStoreFile{Filename: "report\"\r\nX-Injected: no.bin"}, body: []byte("private")}
	h := &DigitalStoreHandler{store: stub}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.GET("/store/orders/:id/delivery", h.GetDelivery)
	router.GET("/store/orders/:id/download", h.Download)
	router.POST("/store/orders/:id/resume", h.ResumeOrder)

	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/store/orders/9/delivery", nil),
		httptest.NewRequest(http.MethodPost, "/store/orders/9/resume", nil),
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if recorder.Code == http.StatusOK {
			t.Fatalf("unexpected owner-A success for %s", req.URL.Path)
		}
		if stub.gotUserID != 7 {
			t.Fatalf("handler used owner %d, want authenticated owner 7", stub.gotUserID)
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/store/orders/9/download", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("download status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("content type = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("cache control = %q", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("nosniff = %q", got)
	}
	if got := recorder.Header().Get("Content-Disposition"); strings.ContainsAny(got, "\r\n") {
		t.Fatalf("unsafe content disposition = %q", got)
	}
	if stub.gotUserID != 7 {
		t.Fatalf("download used owner %d, want 7", stub.gotUserID)
	}
}
