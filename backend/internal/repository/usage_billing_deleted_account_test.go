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

func TestGapFixDeletedAccountSettlement(t *testing.T) {
	for _, billing := range []string{"balance", "subscription"} {
		for _, accountType := range []string{service.AccountTypeAPIKey, service.AccountTypeBedrock} {
			for _, state := range []string{"deleted", "active", "database_error"} {
				t.Run(billing+"/"+accountType+"/"+state, func(t *testing.T) {
					db, mock, err := sqlmock.New()
					require.NoError(t, err)
					defer db.Close()
					cmd := &service.UsageBillingCommand{
						RequestID: "gap-account-" + billing + accountType + state,
						APIKeyID:  7, UserID: 42, AccountID: 99, AccountType: accountType,
						AccountQuotaCost: 10, AccountShareOwnerUserID: 101, AccountShareOwnerRatePercent: 50,
					}
					if billing == "balance" {
						cmd.BalanceCost = 10
					} else {
						id := int64(23)
						cmd.SubscriptionID, cmd.SubscriptionCost = &id, 10
					}
					expectDeletedKeyBillingClaim(mock, cmd)
					if billing == "balance" {
						expectDeletedKeyBalanceDeduction(mock)
					} else {
						mock.ExpectExec(`(?s)UPDATE user_subscriptions us.*WHERE us.id = \$2`).
							WithArgs(10.0, int64(23)).WillReturnResult(sqlmock.NewResult(0, 1))
					}
					q := mock.ExpectQuery(`(?s)UPDATE accounts SET extra = .*WHERE id = \$2 AND deleted_at IS NULL.*RETURNING`).
						WithArgs(10.0, int64(99))
					switch state {
					case "deleted":
						q.WillReturnRows(sqlmock.NewRows([]string{"quota_used"}))
					case "database_error":
						q.WillReturnError(sql.ErrConnDone)
					default:
						q.WillReturnRows(sqlmock.NewRows([]string{
							"quota_used", "quota_limit", "quota_daily_used", "quota_daily_limit",
							"quota_weekly_used", "quota_weekly_limit", "quota_monthly_used", "quota_monthly_limit",
						}).AddRow(10, 100, 10, 100, 10, 100, 10, 100))
					}
					if state == "database_error" {
						mock.ExpectRollback()
					} else {
						// Ordered expectations prove the share ledger survives missing accounts
						// and is still written after account counters, before commit.
						mock.ExpectExec(`(?s)INSERT INTO account_share_ledger`).
							WithArgs(int64(101), int64(99), int64(42), cmd.RequestID, int64(7),
								10.0, 50.0, 5.0, 5.0, cmd.AccountShareFreezeHours).
							WillReturnResult(sqlmock.NewResult(1, 1))
						mock.ExpectCommit()
					}
					repo := &usageBillingRepository{db: db}
					result, err := repo.Apply(context.Background(), cmd)
					if state == "database_error" {
						require.ErrorIs(t, err, sql.ErrConnDone)
						require.Nil(t, result)
					} else {
						require.NoError(t, err)
						require.True(t, result.Applied)
						if state == "deleted" {
							require.Nil(t, result.QuotaState)
						} else {
							require.Equal(t, 10.0, result.QuotaState.TotalUsed)
						}
						if billing == "balance" {
							require.Equal(t, 90.0, *result.NewBalance)
						} else {
							require.Nil(t, result.NewBalance)
						}
						mock.ExpectBegin()
						mock.ExpectQuery(`(?s)INSERT INTO usage_billing_dedup .*RETURNING id`).
							WithArgs(cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).WillReturnError(sql.ErrNoRows)
						mock.ExpectQuery(`(?s)SELECT request_fingerprint.*FROM usage_billing_dedup\s+WHERE`).
							WithArgs(cmd.RequestID, cmd.APIKeyID).
							WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}).AddRow(cmd.RequestFingerprint))
						mock.ExpectCommit()
						repeated, err := repo.Apply(context.Background(), cmd)
						require.NoError(t, err)
						require.False(t, repeated.Applied, "retry must not repeat any charge, counter or share ledger")
					}
					require.NoError(t, mock.ExpectationsWereMet())
				})
			}
		}
	}
}
