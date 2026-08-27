package migrations

import (
	"strings"
	"testing"
)

func TestPixelCafeRoomOwnedPlanMigrationSplitsSharedPlansIdempotently(t *testing.T) {
	content, err := FS.ReadFile("236_pixel_cafe_room_owned_plans.sql")
	if err != nil {
		t.Fatalf("read S266 Pixel Cafe migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{
		"ROW_NUMBER() OVER (PARTITION BY plan_id ORDER BY id)",
		"WHERE ranked.owner_rank > 1",
		"INSERT INTO group_buy_plans",
		"UPDATE cafe_rooms",
		"idx_cafe_rooms_plan_active_unique",
		"WHERE deleted_at IS NULL",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing required clause %q", required)
		}
	}
	upper := strings.ToUpper(sql)
	for _, forbidden := range []string{"DELETE FROM ", "TRUNCATE ", "UPDATE GROUP_BUY_ROUNDS", "UPDATE GROUP_BUY_SEATS", "UPDATE PAYMENT_ORDERS"} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("migration touches historical lifecycle data with %q", forbidden)
		}
	}
}
