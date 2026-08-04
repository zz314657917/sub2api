package migrations

import (
	"strings"
	"testing"
)

func TestPixelCafeSchemaMigrationIsAdditiveAndIdempotent(t *testing.T) {
	content, err := FS.ReadFile("201_pixel_cafe_schema.sql")
	if err != nil {
		t.Fatalf("read Pixel Cafe migration: %v", err)
	}

	sql := string(content)
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS access_mode",
		"ADD COLUMN IF NOT EXISTS fulfillment_mode",
		"ADD COLUMN IF NOT EXISTS cafe_room_id",
		"ADD COLUMN IF NOT EXISTS seat_no",
		"ADD COLUMN IF NOT EXISTS managed_source_type",
		"CREATE TABLE IF NOT EXISTS cafe_rooms",
		"CREATE TABLE IF NOT EXISTS api_key_account_bindings",
		"idx_cafe_rooms_code_active",
		"idx_group_buy_rounds_cafe_room_live",
		"idx_group_buy_seats_round_seat_no_live",
		"idx_api_keys_managed_source_active",
		"idx_api_key_account_bindings_active_key",
		"idx_api_key_account_bindings_active_seat",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing required clause %q", required)
		}
	}

	upper := strings.ToUpper(sql)
	for _, forbidden := range []string{"DROP ", "TRUNCATE ", "DELETE FROM ", "BEGIN;", "COMMIT;"} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("additive migration contains forbidden statement %q", forbidden)
		}
	}

	for _, required := range []string{"DEFAULT 'aggregate_tier'", "DEFAULT 'normal'", "DEFAULT ''", "status IN ('draft', 'enabled', 'maintenance', 'disabled')"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing compatibility default or enum %q", required)
		}
	}
}
