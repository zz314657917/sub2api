package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type handlerStoreEncryptor struct{}

func (handlerStoreEncryptor) Encrypt(value string) (string, error) { return value, nil }
func (handlerStoreEncryptor) Decrypt(value string) (string, error) { return value, nil }

func TestDigitalStorePostgresHandlerOwnerBoundDeliveryDownloadAndResume(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DIGITAL_STORE_TEST_DSN"))
	if dsn == "" {
		t.Fatal("DIGITAL_STORE_TEST_DSN is required")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("digital_store_handler_%d", time.Now().UnixNano())
	defer func() { _, _ = db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `CREATE TABLE payment_orders (id BIGSERIAL PRIMARY KEY, status TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "246_digital_store.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	clientDB, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDB.Close()
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, clientDB)))
	defer client.Close()
	fileBody := []byte("owner-only-file")
	hash := sha256.Sum256(fileBody)
	if _, err = db.ExecContext(ctx, `INSERT INTO store_files(id,filename,encrypted_content,content_size,content_sha256) VALUES(1,'owner.bin',$1,$2,$3)`, base64.StdEncoding.EncodeToString(fileBody), len(fileBody), fmt.Sprintf("%x", hash)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO store_products(id,name,kind,price_cents,enabled,file_id) VALUES(1,'file','file',100,true,1)`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO payment_orders(id,status) VALUES(11,'COMPLETED')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO store_orders(payment_order_id,user_id,product_id,idempotency_key,product_name,kind,price_cents,file_id,delivery_status) VALUES(11,2,1,'owner-only','file','file',100,1,'delivered')`); err != nil {
		t.Fatal(err)
	}

	h := &DigitalStoreHandler{store: service.NewDigitalStoreService(client, nil, handlerStoreEncryptor{}, nil)}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		c.Next()
	})
	router.GET("/store/orders/:id/delivery", h.GetDelivery)
	router.GET("/store/orders/:id/download", h.Download)
	router.POST("/store/orders/:id/resume", h.ResumeOrder)
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/store/orders/11/delivery", nil),
		httptest.NewRequest(http.MethodGet, "/store/orders/11/download", nil),
		httptest.NewRequest(http.MethodPost, "/store/orders/11/resume", nil),
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s returned %d, want 404: %s", req.URL.Path, recorder.Code, recorder.Body.String())
		}
	}

	ownerRouter := gin.New()
	ownerRouter.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 2})
		c.Next()
	})
	ownerRouter.GET("/store/orders/:id/delivery", h.GetDelivery)
	ownerRouter.GET("/store/orders/:id/download", h.Download)
	for _, req := range []*http.Request{httptest.NewRequest(http.MethodGet, "/store/orders/11/delivery", nil), httptest.NewRequest(http.MethodGet, "/store/orders/11/download", nil)} {
		recorder := httptest.NewRecorder()
		ownerRouter.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("owner %s returned %d: %s", req.URL.Path, recorder.Code, recorder.Body.String())
		}
		if req.URL.Path == "/store/orders/11/download" && recorder.Body.String() != string(fileBody) {
			t.Fatalf("download body = %q", recorder.Body.String())
		}
	}
}
