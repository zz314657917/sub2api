package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestDigitalStorePostgresMigration246IsIdempotent intentionally requires an
// isolated DSN.  It is not an optional local test: a missing DSN is a setup
// failure, not evidence that the migration is safe.
func TestDigitalStorePostgresMigration246IsIdempotent(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DIGITAL_STORE_TEST_DSN"))
	if dsn == "" {
		t.Fatal("DIGITAL_STORE_TEST_DSN is required for DigitalStore PostgreSQL tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open isolated postgres: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("ping isolated postgres: %v", err)
	}
	db.SetMaxOpenConns(1)
	schema := fmt.Sprintf("digital_store_migration_%d", time.Now().UnixNano())
	if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec("DROP SCHEMA " + schema + " CASCADE") }()
	if _, err := db.Exec("SET search_path TO " + schema); err != nil {
		t.Fatal(err)
	}
	// Migration 246 deliberately references the pre-existing payment schema.
	// The disposable QA database may start empty, so provide only that prior
	// migration's primary-key contract; production migration order supplies it.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS payment_orders (id BIGSERIAL PRIMARY KEY)`); err != nil {
		t.Fatalf("create payment_orders prerequisite: %v", err)
	}
	body, err := os.ReadFile(filepath.Join("246_digital_store.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := db.Exec(string(body)); err != nil {
			t.Fatalf("apply migration 246 attempt %d: %v", attempt+1, err)
		}
	}
	for _, table := range []string{"store_files", "store_products", "store_inventory", "store_orders"} {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)`, table).Scan(&exists); err != nil || !exists {
			t.Fatalf("migration table %s missing (exists=%v, err=%v)", table, exists, err)
		}
	}
}
