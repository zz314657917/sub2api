package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/lib/pq"
)

// TestDigitalStorePostgresCRUDLists uses only a freshly-created schema in the
// controller-provided QA database. The test intentionally fails without the
// dedicated DSN so a skipped database check cannot be reported as a pass.
func TestDigitalStorePostgresCRUDLists(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DIGITAL_STORE_TEST_DSN"))
	if dsn == "" {
		t.Fatal("DIGITAL_STORE_TEST_DSN is required for DigitalStore PostgreSQL tests")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("digital_store_crud_%d", time.Now().UnixNano())
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
		t.Fatalf("apply migration: %v", err)
	}

	clientDB, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDB.Close()
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, clientDB)))
	defer client.Close()
	svc := NewDigitalStoreService(client, nil, passthroughStoreEncryptor{}, nil)

	fileBody := []byte("blob")
	fileHash := sha256.Sum256(fileBody)
	if _, err = db.ExecContext(ctx, `INSERT INTO store_files(id,filename,encrypted_content,content_size,content_sha256) VALUES
		(1,'old.zip','cipher',0,repeat('a',64)),(2,'new.zip','cipher',0,repeat('b',64)),(3,'snapshot-name.zip',$1,$2,$3)`, base64.StdEncoding.EncodeToString(fileBody), len(fileBody), fmt.Sprintf("%x", fileHash)); err != nil {
		t.Fatal(err)
	}
	fileProduct, err := svc.AdminCreateProduct(ctx, DigitalStoreProductInput{Name: "File plan", Kind: "file", PriceCents: 100, Enabled: true, FileID: storeInt64Ptr(1)})
	if err != nil || fileProduct.FileID == nil || *fileProduct.FileID != 1 {
		t.Fatalf("create file product=%+v err=%v", fileProduct, err)
	}
	cardProduct, err := svc.AdminCreateProduct(ctx, DigitalStoreProductInput{Name: "Gift Card", Description: "searchable", Kind: "card", PriceCents: 200, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AdminImportStock(ctx, cardProduct.ID, []DigitalStoreStockInput{{Content: "TOP-SECRET-CARD"}, {Content: "SECOND-SECRET"}}); err != nil {
		t.Fatal(err)
	}

	products, total, err := svc.ListProductsPage(ctx, true, "gift", "card", 1, 100)
	if err != nil || total != 1 || len(products) != 1 || products[0].ID != cardProduct.ID || products[0].StockAvailable != 2 {
		t.Fatalf("filtered products=%+v total=%d err=%v", products, total, err)
	}
	if _, _, err = svc.ListProductsPage(ctx, true, "", "card", 1, 101); err == nil {
		t.Fatal("page_size > 100 accepted")
	}
	updated, err := svc.AdminUpdateProduct(ctx, fileProduct.ID, DigitalStoreProductInput{Name: "File plan", Kind: "file", PriceCents: 101, Enabled: true, FileID: storeInt64Ptr(2)})
	if err != nil || updated.FileID == nil || *updated.FileID != 2 {
		t.Fatalf("replace product file=%+v err=%v", updated, err)
	}
	if _, err = svc.AdminUpdateProduct(ctx, cardProduct.ID, DigitalStoreProductInput{Name: "Gift Card", Kind: "card", PriceCents: 200, Enabled: true, FileID: storeInt64Ptr(2)}); err == nil {
		t.Fatal("non-file product accepted file_id")
	}

	if _, err = db.ExecContext(ctx, `INSERT INTO payment_orders(id,status) VALUES (11,'COMPLETED'),(12,'PENDING'),(13,'COMPLETED')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `INSERT INTO store_orders(payment_order_id,user_id,product_id,idempotency_key,product_name,kind,price_cents,file_id,inventory_id,delivery_status) VALUES
		(11,7,$1,'one','Gift Card','card',200,NULL,1,'delivered'),
		(12,8,$1,'two','Gift Card','card',200,NULL,2,'reserved'),
		(13,7,$2,'file-one','File plan','file',101,3,NULL,'delivered')`, cardProduct.ID, fileProduct.ID); err != nil {
		t.Fatal(err)
	}
	orders, orderTotal, err := svc.ListOrdersPage(ctx, 7, 1, 20)
	foundCardOrder := false
	for _, order := range orders {
		foundCardOrder = foundCardOrder || (order.OrderID == 11 && order.ProductID == cardProduct.ID)
	}
	if err != nil || orderTotal != 2 || len(orders) != 2 || !foundCardOrder {
		t.Fatalf("user orders=%+v total=%d err=%v", orders, orderTotal, err)
	}
	adminOrders, adminTotal, err := svc.AdminListOrders(ctx, "COMPLETED", 7, cardProduct.ID, 1, 20)
	if err != nil || adminTotal != 1 || len(adminOrders) != 1 || adminOrders[0].UserID != 7 || adminOrders[0].ProductID != cardProduct.ID {
		t.Fatalf("admin orders=%+v total=%d err=%v", adminOrders, adminTotal, err)
	}
	fileOrder, err := svc.GetOrder(ctx, 7, 13)
	if err != nil || fileOrder.ProductID != fileProduct.ID {
		t.Fatalf("file order product id=%+v err=%v", fileOrder, err)
	}
	// Delivery/download access deliberately does not call requireEnabled: a
	// disabled Store blocks new browsing and checkout, not prior paid access.
	delivery, err := svc.GetDelivery(ctx, 7, 13)
	if err != nil || delivery.Kind != "file" || delivery.Filename != "snapshot-name.zip" || delivery.DownloadURL == "" {
		t.Fatalf("file delivery=%+v err=%v", delivery, err)
	}
	download, body, err := svc.GetDownload(ctx, 7, 13)
	if err != nil || download.Filename != "snapshot-name.zip" || string(body) != string(fileBody) {
		t.Fatalf("file download=%+v body=%q err=%v", download, body, err)
	}
	if _, err = svc.GetDelivery(ctx, 8, 13); err == nil {
		t.Fatal("cross-user file delivery accepted")
	}
	if _, _, err = svc.GetDownload(ctx, 8, 13); err == nil {
		t.Fatal("cross-user file download accepted")
	}
	stock, stockTotal, err := svc.AdminListStock(ctx, cardProduct.ID, 1, 20)
	if err != nil || stockTotal != 2 || len(stock) != 2 {
		t.Fatalf("stock=%+v total=%d err=%v", stock, stockTotal, err)
	}
	for _, item := range stock {
		if item.Kind != "card" || item.State != "available" || item.ID <= 0 || item.CreatedAt.IsZero() {
			t.Fatalf("unmasked stock metadata invalid: %+v", item)
		}
	}
	masked, err := json.Marshal(stock)
	if err != nil || strings.Contains(string(masked), "SECRET") || strings.Contains(string(masked), "encrypted_content") {
		t.Fatalf("stock list leaked sensitive content: %s (err=%v)", masked, err)
	}
}

func storeInt64Ptr(v int64) *int64 { return &v }

type passthroughStoreEncryptor struct{}

func (passthroughStoreEncryptor) Encrypt(value string) (string, error) { return value, nil }
func (passthroughStoreEncryptor) Decrypt(value string) (string, error) { return value, nil }
