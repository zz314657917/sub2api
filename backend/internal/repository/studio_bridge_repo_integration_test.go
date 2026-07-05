//go:build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
		AppID:              service.StudioBridgeAppLuoyeAI,
		UserID:             user.ID,
		ChargeKey:          "task:" + uuid.NewString() + ":precharge",
		Amount:             0.8,
		TaskID:             "task-" + uuid.NewString(),
		Mode:               "generate",
		Model:              "gpt-image-2",
		ActorUserID:        "sub2api:actor",
		TeamID:             "team-1",
		ImageCount:         4,
		ImageSize:          service.ImageBillingSize1K,
		ImageSizeSource:    service.ImageSizeSourceInput,
		ImageSizeBreakdown: map[string]int{service.ImageBillingSize1K: 4},
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

	commitCmd := cmd
	commitCmd.ImageCount = 0
	commitCmd.ImageSize = ""
	commitCmd.ImageSizeSource = ""
	commitCmd.ImageSizeBreakdown = nil
	committed, err := repo.CommitStudioBridgeCharge(ctx, commitCmd)
	require.NoError(t, err)
	require.True(t, committed.Applied)

	duplicateCommit, err := repo.CommitStudioBridgeCharge(ctx, commitCmd)
	require.NoError(t, err)
	require.False(t, duplicateCommit.Applied)

	var actualCost float64
	var usageCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(actual_cost), 0)::double precision FROM usage_logs WHERE request_id = $1", "studio:"+cmd.TaskID).Scan(&usageCount, &actualCost))
	require.Equal(t, 1, usageCount)
	require.InDelta(t, 0.5, actualCost, 0.000001)
	var usageAPIKeyID int64
	var durationMs sql.NullInt64
	var imageSize sql.NullString
	var imageSizeSource sql.NullString
	var imageSizeBreakdown sql.NullString
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT api_key_id, duration_ms, image_size, image_size_source, image_size_breakdown::text FROM usage_logs WHERE request_id = $1", "studio:"+cmd.TaskID).Scan(&usageAPIKeyID, &durationMs, &imageSize, &imageSizeSource, &imageSizeBreakdown))
	require.Equal(t, defaultKey.ID, usageAPIKeyID)
	require.True(t, durationMs.Valid)
	require.GreaterOrEqual(t, durationMs.Int64, int64(0))
	require.True(t, imageSize.Valid)
	require.Equal(t, service.ImageBillingSize1K, imageSize.String)
	require.True(t, imageSizeSource.Valid)
	require.Equal(t, service.ImageSizeSourceInput, imageSizeSource.String)
	require.True(t, imageSizeBreakdown.Valid)
	var breakdown map[string]int
	require.NoError(t, json.Unmarshal([]byte(imageSizeBreakdown.String), &breakdown))
	require.Equal(t, map[string]int{service.ImageBillingSize1K: 4}, breakdown)

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

func TestStudioBridgeResolveChargeUsageRefsPrefersOfficialImageAccount(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	repo := &studioBridgeRepository{}

	var userID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, status, balance)
		VALUES ($1, 'x', 'active', 10)
		RETURNING id
	`, "studio-official-image-"+uuid.NewString()+"@example.test").Scan(&userID))

	var ordinaryGroupID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO groups (name, platform, status, subscription_type, rate_multiplier, is_exclusive)
		VALUES ($1, 'openai', 'active', 'standard', 1, FALSE)
		RETURNING id
	`, "studio-ordinary-image-"+uuid.NewString()).Scan(&ordinaryGroupID))

	var officialGroupID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO groups (name, platform, status, subscription_type, rate_multiplier, is_exclusive)
		VALUES ($1, 'openai', 'active', 'standard', 1, FALSE)
		RETURNING id
	`, "studio-official-image-"+uuid.NewString()).Scan(&officialGroupID))

	var ordinaryAccountID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO accounts (name, platform, type, credentials, priority, status)
		VALUES ($1, 'openai', 'apikey', '{"base_url":"https://fluapi.example/v1"}'::jsonb, 1, 'active')
		RETURNING id
	`, "fluapi-image-"+uuid.NewString()).Scan(&ordinaryAccountID))

	var officialAccountID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO accounts (name, platform, type, credentials, priority, status)
		VALUES ($1, 'openai', 'apikey', '{"base_url":"https://api.apimart.ai/v1","model_mapping":{"gpt-image-2":"gpt-image-2"}}'::jsonb, 50, 'active')
		RETURNING id
	`, "apimart-official-image-"+uuid.NewString()).Scan(&officialAccountID))

	_, err := tx.ExecContext(ctx, `
		INSERT INTO account_groups (account_id, group_id, priority)
		VALUES ($1, $2, 1), ($3, $4, 10)
	`, ordinaryAccountID, ordinaryGroupID, officialAccountID, officialGroupID)
	require.NoError(t, err)

	routes, err := json.Marshal([]map[string]any{
		{
			"group_id":   ordinaryGroupID,
			"priority":   1,
			"enabled":    true,
			"image_only": true,
		},
		{
			"group_id":       officialGroupID,
			"priority":       10,
			"enabled":        true,
			"image_only":     true,
			"model_patterns": []string{"gpt-image-2-official"},
		},
	})
	require.NoError(t, err)

	var apiKeyID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO api_keys (user_id, key, name, group_id, status, account_pool_strategy, multi_group_routes)
		VALUES ($1, $2, $3, $4, 'active', 'shared_only', $5::jsonb)
		RETURNING id
	`, userID, "sk-studio-official-image-"+uuid.NewString(), service.DefaultAPIKeyName, ordinaryGroupID, string(routes)).Scan(&apiKeyID))

	refs, err := repo.resolveChargeUsageRefs(ctx, tx, service.StudioBridgeChargeCommand{
		UserID: userID,
		Mode:   "generate",
		Model:  "gpt-image-2-official",
	})
	require.NoError(t, err)
	require.Equal(t, apiKeyID, refs.apiKeyID)
	require.Equal(t, officialAccountID, refs.accountID)
	require.True(t, refs.groupID.Valid)
	require.Equal(t, officialGroupID, refs.groupID.Int64)
}
