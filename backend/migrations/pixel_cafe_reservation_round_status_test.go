package migrations

import (
	"context"
	"database/sql"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const cafeReservationStatusMigration = "242_pixel_cafe_reservation_round_status.sql"

func cafeMigrationStates(t *testing.T, filename string) []string {
	t.Helper()
	content, err := FS.ReadFile(filename)
	require.NoError(t, err)
	clause := regexp.MustCompile(`(?s)ADD CONSTRAINT group_buy_rounds_status_check CHECK \(status IN \((.*?)\)\)`).FindStringSubmatch(string(content))
	require.Len(t, clause, 2)
	var states []string
	for _, match := range regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(clause[1], -1) {
		states = append(states, match[1])
	}
	return states
}

func TestCafeReservationRoundStatusMigration(t *testing.T) {
	states := cafeMigrationStates(t, cafeReservationStatusMigration)
	want := append(cafeMigrationStates(t, "235_pixel_cafe_share_fulfillment.sql"), "reserving", "awaiting_payment")
	require.ElementsMatch(t, want, states, "preserve every legacy state and add both reservation states")
	content, err := FS.ReadFile(cafeReservationStatusMigration)
	require.NoError(t, err)
	for _, forbidden := range []string{"UPDATE ", "DELETE FROM ", "TRUNCATE ", "BEGIN;", "COMMIT;"} {
		require.NotContains(t, strings.ToUpper(string(content)), forbidden)
	}
}

func TestCafeReservationRoundStatusPostgres(t *testing.T) {
	if os.Getenv("SUB2API_RUN_POSTGRES_MIGRATION_TESTS") != "1" {
		t.Skip("set SUB2API_RUN_POSTGRES_MIGRATION_TESTS=1 for isolated PostgreSQL regression")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	container, err := tcpostgres.Run(ctx, "postgres:18-alpine",
		tcpostgres.WithDatabase("cafe_reservation_status_test"),
		tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 30*time.Second)
		defer stop()
		require.NoError(t, container.Terminate(cleanupCtx))
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	legacy := cafeMigrationStates(t, "235_pixel_cafe_share_fulfillment.sql")
	_, err = db.ExecContext(ctx, `CREATE TABLE group_buy_rounds (
id BIGSERIAL PRIMARY KEY, status VARCHAR(20) NOT NULL,
CONSTRAINT group_buy_rounds_status_check CHECK (status IN ('`+strings.Join(legacy, "', '")+`')))`)
	require.NoError(t, err)
	for _, state := range legacy {
		_, err = db.ExecContext(ctx, "INSERT INTO group_buy_rounds(status) VALUES ($1)", state)
		require.NoError(t, err)
	}
	for _, state := range []string{"reserving", "awaiting_payment"} {
		_, err = db.ExecContext(ctx, "INSERT INTO group_buy_rounds(status) VALUES ($1)", state)
		var pgErr *pq.Error
		require.ErrorAs(t, err, &pgErr)
		require.Equal(t, "23514", string(pgErr.Code))
		require.Equal(t, "group_buy_rounds_status_check", pgErr.Constraint)
	}
	migration, err := FS.ReadFile(cafeReservationStatusMigration)
	require.NoError(t, err)
	for pass := 0; pass < 2; pass++ {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, string(migration))
		if err != nil {
			_ = tx.Rollback()
		}
		require.NoError(t, err)
		require.NoError(t, tx.Commit())
		if pass == 0 {
			for _, state := range []string{"reserving", "awaiting_payment"} {
				_, err = db.ExecContext(ctx, "INSERT INTO group_buy_rounds(status) VALUES ($1)", state)
				require.NoError(t, err)
			}
		}
	}
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM group_buy_rounds").Scan(&count))
	require.Equal(t, len(legacy)+2, count, "migration rerun must preserve new and historical rows")
	_, err = db.ExecContext(ctx, "UPDATE group_buy_rounds SET status = 'awaiting_payment' WHERE status = 'reserving'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO group_buy_rounds(status) VALUES ('unknown')")
	var pgErr *pq.Error
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", string(pgErr.Code), "unknown states must still be rejected")
}
