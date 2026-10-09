package service

import (
	"context"
	"database/sql"
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

type failingStoreEncryptor struct{}

func (failingStoreEncryptor) Encrypt(string) (string, error) { return "", fmt.Errorf("encrypt failed") }
func (failingStoreEncryptor) Decrypt(string) (string, error) { return "", fmt.Errorf("decrypt failed") }

func TestDigitalStoreLaunchResponseDecryptFailsClosed(t *testing.T) {
	svc := &DigitalStoreService{encryptor: failingStoreEncryptor{}}
	if _, err := svc.decryptStoreLaunchResponse("ciphertext"); err == nil {
		t.Fatal("invalid encrypted launch response was accepted")
	}
}

func TestDigitalStorePostgresFulfillmentReleaseAndLatePayment(t *testing.T) {
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
	schema := fmt.Sprintf("digital_store_payment_%d", time.Now().UnixNano())
	defer func() { _, _ = db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE payment_orders (id BIGSERIAL PRIMARY KEY, order_type TEXT NOT NULL, status TEXT NOT NULL, paid_at TIMESTAMPTZ NULL, completed_at TIMESTAMPTZ NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "246_digital_store.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	clientDB, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDB.Close()
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, clientDB)))
	defer client.Close()
	svc := NewDigitalStoreService(client, nil, passthroughStoreEncryptor{}, nil)
	if _, err := db.ExecContext(ctx, `INSERT INTO store_products(id,name,kind,price_cents,enabled) VALUES(1,'card','card',100,true)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO payment_orders(id,order_type,status,paid_at) VALUES(11,'store','PAID',NOW()),(12,'store','CANCELLED',NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO store_inventory(id,product_id,kind,encrypted_content) VALUES(1,1,'card','secret'),(2,1,'card','late')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO store_orders(id,payment_order_id,user_id,product_id,idempotency_key,product_name,kind,price_cents,inventory_id) VALUES(1,11,7,1,'fulfilled','Gift card','card',100,1),(2,12,7,1,'late','Gift card','card',100,2)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE store_inventory SET state='reserved',reservation_order_id=id,reserved_at=NOW()`); err != nil {
		t.Fatal(err)
	}
	if err := svc.FulfillStoreOrder(ctx, 11); err != nil {
		t.Fatalf("first fulfillment: %v", err)
	}
	if err := svc.FulfillStoreOrder(ctx, 11); err != nil {
		t.Fatalf("duplicate fulfillment: %v", err)
	}
	var status, delivery, invState, encrypted string
	if err := db.QueryRowContext(ctx, `SELECT po.status,so.delivery_status,si.state,so.encrypted_delivery FROM payment_orders po JOIN store_orders so ON so.payment_order_id=po.id JOIN store_inventory si ON si.id=so.inventory_id WHERE po.id=11`).Scan(&status, &delivery, &invState, &encrypted); err != nil {
		t.Fatal(err)
	}
	if status != "COMPLETED" || delivery != "delivered" || invState != "consumed" || encrypted != "secret" {
		t.Fatalf("unexpected fulfillment state %q %q %q %q", status, delivery, invState, encrypted)
	}
	if err := svc.ReleaseStoreReservation(ctx, 12, "cancel"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE payment_orders SET status='PAID',paid_at=NOW() WHERE id=12`); err != nil {
		t.Fatal(err)
	}
	if err := svc.FulfillStoreOrder(ctx, 12); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT delivery_status FROM store_orders WHERE payment_order_id=12`).Scan(&delivery); err != nil {
		t.Fatal(err)
	}
	if delivery != "needs_attention" {
		t.Fatalf("late payment delivery=%q, want needs_attention", delivery)
	}
}

func TestDigitalStorePostgresRetryNeedsAttentionClaimsExactlyOneReplacement(t *testing.T) {
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
	schema := fmt.Sprintf("digital_store_retry_%d", time.Now().UnixNano())
	defer func() { _, _ = db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, `CREATE TABLE payment_orders (id BIGSERIAL PRIMARY KEY, order_type TEXT NOT NULL, status TEXT NOT NULL, paid_at TIMESTAMPTZ NULL, completed_at TIMESTAMPTZ NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
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
	svc := NewDigitalStoreService(client, nil, passthroughStoreEncryptor{}, nil)
	_, err = db.ExecContext(ctx, `INSERT INTO store_products(id,name,kind,price_cents,enabled) VALUES(1,'card','card',100,true); INSERT INTO payment_orders(id,order_type,status,paid_at) VALUES(21,'store','PAID',NOW()),(22,'store','PENDING',NULL); INSERT INTO store_inventory(id,product_id,kind,encrypted_content) VALUES(10,1,'card','replacement'),(11,1,'card','unpaid'); INSERT INTO store_orders(id,payment_order_id,user_id,product_id,idempotency_key,product_name,kind,price_cents,inventory_id,delivery_status) VALUES(10,21,7,1,'paid','Gift card','card',100,10,'needs_attention'),(11,22,7,1,'unpaid','Gift card','card',100,11,'needs_attention')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RetryStoreFulfillment(ctx, 22); err == nil {
		t.Fatal("unpaid retry was accepted")
	}
	if err := svc.RetryStoreFulfillment(ctx, 21); err != nil {
		t.Fatalf("paid retry: %v", err)
	}
	if err := svc.RetryStoreFulfillment(ctx, 21); err != nil {
		t.Fatalf("duplicate retry: %v", err)
	}
	var delivery, paymentStatus, inventoryState string
	var deliveryInventory int64
	if err := db.QueryRowContext(ctx, `SELECT so.delivery_status,po.status,so.inventory_id,si.state FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id JOIN store_inventory si ON si.id=so.inventory_id WHERE so.payment_order_id=21`).Scan(&delivery, &paymentStatus, &deliveryInventory, &inventoryState); err != nil {
		t.Fatal(err)
	}
	if delivery != "delivered" || paymentStatus != "COMPLETED" || deliveryInventory != 10 || inventoryState != "consumed" {
		t.Fatalf("retry state %q %q %d %q", delivery, paymentStatus, deliveryInventory, inventoryState)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM store_inventory WHERE state='consumed'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("consumed inventory=%d, want 1", count)
	}
}
