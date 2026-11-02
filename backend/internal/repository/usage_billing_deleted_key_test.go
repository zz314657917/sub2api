//go:build unit

package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

const (
	deletedKeyQuotaSQL     = `(?s)UPDATE api_keys\s+SET quota_used = quota_used \+ \$1,.*WHERE id = \$2 AND deleted_at IS NULL\s+RETURNING`
	deletedKeyRateLimitSQL = `(?s)UPDATE api_keys SET\s+usage_5h = .*WHERE id = \$2 AND deleted_at IS NULL`
)

func TestUsageBillingRepositoryApply_DeletedAPIKeyStillSettles(t *testing.T) {
	for _, billingType := range []string{"balance", "subscription"} {
		for _, counter := range []string{"quota", "rate_limit", "both"} {
			t.Run(billingType+"/"+counter, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer func() { _ = db.Close() }()

				cmd := &service.UsageBillingCommand{
					RequestID:        "deleted-key-" + billingType + "-" + counter,
					APIKeyID:         7,
					UserID:           42,
					AccountID:        99,
					AccountType:      service.AccountTypeAPIKey,
					AccountQuotaCost: 10,
				}
				if billingType == "subscription" {
					subscriptionID := int64(23)
					cmd.SubscriptionID = &subscriptionID
					cmd.SubscriptionCost = 10
					cmd.AccountType = service.AccountTypeBedrock
				} else {
					cmd.BalanceCost = 10
				}
				if counter != "rate_limit" {
					cmd.APIKeyQuotaCost = 10
				}
				if counter != "quota" {
					cmd.APIKeyRateLimitCost = 10
				}

				expectDeletedKeyBillingClaim(mock, cmd)
				if billingType == "subscription" {
					mock.ExpectExec(`(?s)UPDATE user_subscriptions us.*WHERE us.id = \$2.*AND us.deleted_at IS NULL.*AND g.deleted_at IS NULL`).
						WithArgs(10.0, int64(23)).
						WillReturnResult(sqlmock.NewResult(0, 1))
				} else {
					expectDeletedKeyBalanceDeduction(mock)
				}
				if cmd.APIKeyQuotaCost > 0 {
					mock.ExpectQuery(deletedKeyQuotaSQL).
						WithArgs(10.0, int64(7), service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted).
						WillReturnError(sql.ErrNoRows)
				}
				if cmd.APIKeyRateLimitCost > 0 {
					mock.ExpectExec(deletedKeyRateLimitSQL).
						WithArgs(10.0, int64(7)).
						WillReturnResult(sqlmock.NewResult(0, 0))
				}
				mock.ExpectQuery(`(?s)UPDATE accounts SET extra = .*WHERE id = \$2 AND deleted_at IS NULL.*RETURNING`).
					WithArgs(10.0, int64(99)).
					WillReturnRows(sqlmock.NewRows([]string{
						"quota_used", "quota_limit", "quota_daily_used", "quota_daily_limit",
						"quota_weekly_used", "quota_weekly_limit", "quota_monthly_used", "quota_monthly_limit",
					}).AddRow(10.0, 100.0, 10.0, 100.0, 10.0, 100.0, 10.0, 100.0))
				mock.ExpectCommit()

				repo := &usageBillingRepository{db: db}
				result, err := repo.Apply(context.Background(), cmd)
				require.NoError(t, err)
				require.True(t, result.Applied)
				require.False(t, result.APIKeyQuotaExhausted)
				require.NotNil(t, result.QuotaState)
				require.InDelta(t, 10, result.QuotaState.TotalUsed, 0.000001)
				if billingType == "balance" {
					require.NotNil(t, result.NewBalance)
					require.InDelta(t, 90, *result.NewBalance, 0.000001)
					require.InDelta(t, 10, result.BalanceCost, 0.000001)
				} else {
					require.Nil(t, result.NewBalance)
				}

				// A retry must commit without charging any user or account counter again.
				mock.ExpectBegin()
				mock.ExpectQuery(`(?s)INSERT INTO usage_billing_dedup .*RETURNING id`).
					WithArgs(cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(`(?s)SELECT request_fingerprint\s+FROM usage_billing_dedup\s+WHERE request_id = \$1 AND api_key_id = \$2`).
					WithArgs(cmd.RequestID, cmd.APIKeyID).
					WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}).AddRow(cmd.RequestFingerprint))
				mock.ExpectCommit()

				repeated, err := repo.Apply(context.Background(), cmd)
				require.NoError(t, err)
				require.False(t, repeated.Applied)
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}

func TestUsageBillingRepositoryApply_ActiveAPIKeyStillCountsUsage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cmd := &service.UsageBillingCommand{
		RequestID:           "active-key-control",
		APIKeyID:            7,
		UserID:              42,
		BalanceCost:         10,
		APIKeyQuotaCost:     10,
		APIKeyRateLimitCost: 10,
	}
	expectDeletedKeyBillingClaim(mock, cmd)
	expectDeletedKeyBalanceDeduction(mock)
	mock.ExpectQuery(deletedKeyQuotaSQL).
		WithArgs(10.0, int64(7), service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted).
		WillReturnRows(sqlmock.NewRows([]string{"exhausted"}).AddRow(true))
	mock.ExpectExec(deletedKeyRateLimitSQL).
		WithArgs(10.0, int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := (&usageBillingRepository{db: db}).Apply(context.Background(), cmd)
	require.NoError(t, err)
	require.True(t, result.Applied)
	require.True(t, result.APIKeyQuotaExhausted)
	require.NotNil(t, result.NewBalance)
	require.InDelta(t, 90, *result.NewBalance, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageBillingRepositoryApply_APIKeyDatabaseErrorsStillRollback(t *testing.T) {
	for _, counter := range []string{"quota", "rate_limit"} {
		t.Run(counter, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = db.Close() }()

			cmd := &service.UsageBillingCommand{
				RequestID:   "key-database-error-" + counter,
				APIKeyID:    7,
				UserID:      42,
				BalanceCost: 10,
			}
			if counter == "quota" {
				cmd.APIKeyQuotaCost = 10
			} else {
				cmd.APIKeyRateLimitCost = 10
			}
			expectDeletedKeyBillingClaim(mock, cmd)
			expectDeletedKeyBalanceDeduction(mock)
			if counter == "quota" {
				mock.ExpectQuery(deletedKeyQuotaSQL).
					WithArgs(10.0, int64(7), service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted).
					WillReturnError(sql.ErrConnDone)
			} else {
				mock.ExpectExec(deletedKeyRateLimitSQL).
					WithArgs(10.0, int64(7)).
					WillReturnError(sql.ErrConnDone)
			}
			mock.ExpectRollback()

			result, err := (&usageBillingRepository{db: db}).Apply(context.Background(), cmd)
			require.ErrorIs(t, err, sql.ErrConnDone)
			require.Nil(t, result)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func expectDeletedKeyBillingClaim(mock sqlmock.Sqlmock, cmd *service.UsageBillingCommand) {
	cmd.Normalize()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO usage_billing_dedup .*RETURNING id`).
		WithArgs(cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(`(?s)SELECT request_fingerprint\s+FROM usage_billing_dedup_archive\s+WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs(cmd.RequestID, cmd.APIKeyID).
		WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}))
}

func expectDeletedKeyBalanceDeduction(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(usageBillingUserBalanceSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
	mock.ExpectQuery(usageBillingOverdraftUpdateSQL).
		WithArgs(10.0, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(90.0))
}
