package migrations

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPixelCafeShareFulfillmentMigrationIsAdditiveAndIdempotent(t *testing.T) {
	content, err := FS.ReadFile("235_pixel_cafe_share_fulfillment.sql")
	if err != nil {
		t.Fatalf("read S252 Pixel Cafe migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS subscription_tier",
		"ADD COLUMN IF NOT EXISTS cafe_fulfillment_version",
		"CREATE TABLE IF NOT EXISTS cafe_round_memberships",
		"ADD COLUMN IF NOT EXISTS membership_id",
		"cafe_round_memberships_share_counts_check",
		"fulfillment_mode <> 'room_subscription'",
		"idx_group_buy_rounds_assigned_account_live",
		"idx_api_key_account_bindings_active_membership",
		"status IN ('open', 'awaiting_account', 'activating', 'active', 'completed', 'refunding', 'refunded', 'failed', 'cancelled')",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing required clause %q", required)
		}
	}
	upper := strings.ToUpper(sql)
	for _, forbidden := range []string{"TRUNCATE ", "DELETE FROM ", "BEGIN;", "COMMIT;"} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("additive migration contains forbidden statement %q", forbidden)
		}
	}
}

func TestPixelCafeShareFulfillmentPostgresBackfillsLegacySeats(t *testing.T) {
	if os.Getenv("SUB2API_RUN_POSTGRES_MIGRATION_TESTS") != "1" {
		t.Skip("set SUB2API_RUN_POSTGRES_MIGRATION_TESTS=1 to run the isolated PostgreSQL migration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	container, err := tcpostgres.Run(
		ctx,
		"postgres:18.1-alpine3.23",
		tcpostgres.WithDatabase("sub2api_cafe_share_migration_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.PingContext(ctx))

	_, err = db.ExecContext(ctx, `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE api_keys (id BIGINT PRIMARY KEY);
CREATE TABLE group_buy_plans (
    id BIGINT PRIMARY KEY,
    fulfillment_mode VARCHAR(32) NOT NULL,
    total_shares INTEGER NOT NULL,
    max_shares_per_user INTEGER NOT NULL
);
CREATE TABLE group_buy_rounds (
    id BIGINT PRIMARY KEY,
    status VARCHAR(24) NOT NULL,
    cafe_room_id BIGINT,
    assigned_account_id BIGINT
);
CREATE TABLE group_buy_seats (
    id BIGINT PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES group_buy_rounds(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    status VARCHAR(24) NOT NULL,
    share_count INTEGER NOT NULL,
    bound_api_key_id BIGINT REFERENCES api_keys(id),
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);
CREATE TABLE api_key_account_bindings (
    id BIGINT PRIMARY KEY,
    seat_id BIGINT NOT NULL REFERENCES group_buy_seats(id),
    status VARCHAR(24) NOT NULL
);
INSERT INTO users (id) VALUES (1), (2);
INSERT INTO api_keys (id) VALUES (10);
INSERT INTO group_buy_plans (id, fulfillment_mode, total_shares, max_shares_per_user)
VALUES (20, 'room_subscription', 10, 10);
INSERT INTO group_buy_rounds (id, status, cafe_room_id, assigned_account_id)
VALUES (30, 'active', 40, 50);
INSERT INTO group_buy_seats (id, round_id, user_id, status, share_count, bound_api_key_id, activated_at, expires_at) VALUES
    (101, 30, 1, 'active', 2, 10, '2026-08-01T00:00:00Z', '2026-09-01T00:00:00Z'),
    (102, 30, 1, 'paid', 1, NULL, NULL, NULL),
    (103, 30, 2, 'locked', 2, NULL, NULL, NULL);
INSERT INTO api_key_account_bindings (id, seat_id, status) VALUES (201, 101, 'active');
`)
	require.NoError(t, err)

	migrationSQL, err := FS.ReadFile("235_pixel_cafe_share_fulfillment.sql")
	require.NoError(t, err)
	for pass := 0; pass < 2; pass++ {
		_, err = db.ExecContext(ctx, string(migrationSQL))
		require.NoErrorf(t, err, "migration pass %d", pass+1)
	}

	type membershipRow struct {
		userID         int64
		status         string
		paidShares     int
		reservedShares int
		boundKeyID     sql.NullInt64
	}
	rows, err := db.QueryContext(ctx, `
SELECT user_id, status, paid_shares, reserved_shares, bound_api_key_id
FROM cafe_round_memberships
WHERE round_id = 30
ORDER BY user_id`)
	require.NoError(t, err)
	defer rows.Close()
	var memberships []membershipRow
	for rows.Next() {
		var item membershipRow
		require.NoError(t, rows.Scan(&item.userID, &item.status, &item.paidShares, &item.reservedShares, &item.boundKeyID))
		memberships = append(memberships, item)
	}
	require.NoError(t, rows.Err())
	require.Len(t, memberships, 2)
	require.Equal(t, membershipRow{userID: 1, status: "active", paidShares: 3, reservedShares: 0, boundKeyID: sql.NullInt64{Int64: 10, Valid: true}}, memberships[0])
	require.Equal(t, membershipRow{userID: 2, status: "locked", paidShares: 0, reservedShares: 2}, memberships[1])

	var linkedSeats int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM group_buy_seats WHERE round_id = 30 AND membership_id IS NOT NULL`).Scan(&linkedSeats))
	require.Equal(t, 3, linkedSeats)
	var fulfillmentVersion string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT cafe_fulfillment_version FROM group_buy_rounds WHERE id = 30`).Scan(&fulfillmentVersion))
	require.Equal(t, "legacy_seat", fulfillmentVersion)
	var seatID, membershipID sql.NullInt64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT seat_id, membership_id FROM api_key_account_bindings WHERE id = 201`).Scan(&seatID, &membershipID))
	require.Equal(t, sql.NullInt64{Int64: 101, Valid: true}, seatID)
	require.False(t, membershipID.Valid)
}
