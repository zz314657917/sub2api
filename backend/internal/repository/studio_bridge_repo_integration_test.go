//go:build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestStudioBridgeRepositoryReserveRefundLedgerIsIdempotent(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewStudioBridgeRepository(integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("studio-bridge-ledger-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Balance:      10,
	})
	cmd := service.StudioBridgeChargeCommand{
		AppID:     service.StudioBridgeAppLuoyeAI,
		UserID:    user.ID,
		ChargeKey: "task:" + uuid.NewString() + ":precharge",
		Amount:    0.5,
		TaskID:    "task-" + uuid.NewString(),
		Mode:      "generate",
		Model:     "gpt-image-2",
	}

	reserved, err := repo.ReserveStudioBridgeCharge(ctx, cmd)
	require.NoError(t, err)
	require.True(t, reserved.Applied)
	require.InDelta(t, 9.5, reserved.BalanceAfter, 0.000001)

	duplicate, err := repo.ReserveStudioBridgeCharge(ctx, cmd)
	require.NoError(t, err)
	require.False(t, duplicate.Applied)

	conflictCmd := cmd
	conflictCmd.Amount = 0.6
	_, err = repo.ReserveStudioBridgeCharge(ctx, conflictCmd)
	require.ErrorIs(t, err, service.ErrStudioBridgeConflict)

	refundCmd := service.StudioBridgeChargeCommand{
		AppID:              service.StudioBridgeAppLuoyeAI,
		UserID:             user.ID,
		ChargeKey:          "task:" + uuid.NewString() + ":refund",
		RefundForChargeKey: cmd.ChargeKey,
		Amount:             0.5,
		TaskID:             cmd.TaskID,
		Mode:               cmd.Mode,
		Model:              cmd.Model,
	}
	refunded, err := repo.RefundStudioBridgeCharge(ctx, refundCmd)
	require.NoError(t, err)
	require.True(t, refunded.Applied)
	require.InDelta(t, 10, refunded.BalanceAfter, 0.000001)

	duplicateRefund, err := repo.RefundStudioBridgeCharge(ctx, refundCmd)
	require.NoError(t, err)
	require.False(t, duplicateRefund.Applied)

	var balance float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", user.ID).Scan(&balance))
	require.InDelta(t, 10, balance, 0.000001)

	var chargeCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM studio_bridge_charges WHERE app_id = $1 AND charge_key IN ($2, $3)", service.StudioBridgeAppLuoyeAI, cmd.ChargeKey, refundCmd.ChargeKey).Scan(&chargeCount))
	require.Equal(t, 2, chargeCount)
}

func TestStudioBridgeRepositoryCommitLogsNetUsageOnceAfterPartialRefund(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewStudioBridgeRepository(integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("studio-bridge-commit-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Balance:      10,
	})
	mustCreateAccount(t, client, &service.Account{
		Name:     "studio-bridge-account-" + uuid.NewString(),
		Type:     service.AccountTypeAPIKey,
		Platform: service.PlatformOpenAI,
	})
	defaultKey := mustCreateApiKey(t, client, &service.APIKey{
		UserID:              user.ID,
		Key:                 "sk-default-" + uuid.NewString(),
		Name:                service.DefaultAPIKeyName,
		Status:              service.StatusActive,
		AccountPoolStrategy: service.AccountPoolStrategySharedOnly,
	})
	cmd := service.StudioBridgeChargeCommand{
		AppID:       service.StudioBridgeAppLuoyeAI,
		UserID:      user.ID,
		ChargeKey:   "task:" + uuid.NewString() + ":precharge",
		Amount:      0.8,
		TaskID:      "task-" + uuid.NewString(),
		Mode:        "generate",
		Model:       "gpt-image-2",
		ActorUserID: "sub2api:actor",
		TeamID:      "team-1",
	}
	_, err := repo.ReserveStudioBridgeCharge(ctx, cmd)
	require.NoError(t, err)
	_, err = repo.RefundStudioBridgeCharge(ctx, service.StudioBridgeChargeCommand{
		AppID:              service.StudioBridgeAppLuoyeAI,
		UserID:             user.ID,
		ChargeKey:          "task:" + uuid.NewString() + ":refund",
		RefundForChargeKey: cmd.ChargeKey,
		Amount:             0.3,
		TaskID:             cmd.TaskID,
		Mode:               cmd.Mode,
		Model:              cmd.Model,
	})
	require.NoError(t, err)

	committed, err := repo.CommitStudioBridgeCharge(ctx, cmd)
	require.NoError(t, err)
	require.True(t, committed.Applied)

	duplicateCommit, err := repo.CommitStudioBridgeCharge(ctx, cmd)
	require.NoError(t, err)
	require.False(t, duplicateCommit.Applied)

	var actualCost float64
	var usageCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(actual_cost), 0)::double precision FROM usage_logs WHERE request_id = $1", "studio:"+cmd.TaskID).Scan(&usageCount, &actualCost))
	require.Equal(t, 1, usageCount)
	require.InDelta(t, 0.5, actualCost, 0.000001)
	var usageAPIKeyID int64
	var durationMs sql.NullInt64
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT api_key_id, duration_ms FROM usage_logs WHERE request_id = $1", "studio:"+cmd.TaskID).Scan(&usageAPIKeyID, &durationMs))
	require.Equal(t, defaultKey.ID, usageAPIKeyID)
	require.True(t, durationMs.Valid)
	require.GreaterOrEqual(t, durationMs.Int64, int64(0))

	var status string
	var refundedAmount float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT status, refunded_amount::double precision
		FROM studio_bridge_charges
		WHERE app_id = $1 AND charge_key = $2
	`, service.StudioBridgeAppLuoyeAI, cmd.ChargeKey).Scan(&status, &refundedAmount))
	require.Equal(t, "committed", status)
	require.InDelta(t, 0.3, refundedAmount, 0.000001)
}

func TestStudioBridgeRepositoryConcurrentReserveOnlyDeductsOnce(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewStudioBridgeRepository(integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("studio-bridge-concurrent-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Balance:      10,
	})
	cmd := service.StudioBridgeChargeCommand{
		AppID:     service.StudioBridgeAppLuoyeAI,
		UserID:    user.ID,
		ChargeKey: "task:" + uuid.NewString() + ":precharge",
		Amount:    0.5,
	}

	var wg sync.WaitGroup
	results := make(chan *service.StudioBridgeChargeResult, 16)
	errors := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := repo.ReserveStudioBridgeCharge(ctx, cmd)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}
	wg.Wait()
	close(results)
	close(errors)

	require.Empty(t, errors)
	applied := 0
	for result := range results {
		if result.Applied {
			applied++
		}
	}
	require.Equal(t, 1, applied)

	var balance float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", user.ID).Scan(&balance))
	require.InDelta(t, 9.5, balance, 0.000001)
}
