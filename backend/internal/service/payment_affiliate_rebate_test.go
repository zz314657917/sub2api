package service

import (
	"context"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

type paymentRebateRepoStub struct {
	firstRechargeAffiliateRepoStub
	amount float64
	calls  int
}

func (r *paymentRebateRepoStub) AccrueQuota(_ context.Context, inviterID, inviteeID int64, amount float64, _ int, orderID *int64) (bool, error) {
	r.calls++
	r.amount = amount
	if inviterID != 1 || inviteeID != 2 || orderID == nil || *orderID != 3 {
		return false, fmt.Errorf("unexpected rebate recipient or source order")
	}
	return true, nil
}

func TestPaymentAffiliateRebateExcludesPackageBonus(t *testing.T) {
	for _, tt := range []struct {
		name               string
		amount, paid, base float64
		snapshot           map[string]any
	}{
		{"ordinary recharge", 10, 10, 10, nil},
		{"100 paid 110 credited", 110, 100, 100, map[string]any{"recharge_package_pay_amount": float64(100)}},
		{"package with processing fee", 110, 103, 100, map[string]any{"recharge_package_pay_amount": float64(100)}},
		{"package bonus already claimed", 100, 100, 100, map[string]any{"recharge_package_pay_amount": float64(100)}},
		{"legacy recharge with fee", 100, 103, 100, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
			t.Cleanup(func() { _ = client.Close() })
			inviterID := int64(1)
			repo := &paymentRebateRepoStub{firstRechargeAffiliateRepoStub: firstRechargeAffiliateRepoStub{
				summaryByUserID: map[int64]*AffiliateSummary{
					1: {UserID: 1}, 2: {UserID: 2, InviterID: &inviterID},
				},
			}}
			settings := NewSettingService(firstRechargeAffiliateSettingRepo{
				SettingKeyAffiliateEnabled:           "true",
				SettingKeyAffiliateRebateRate:        "10",
				SettingKeyAffiliateRebateFreezeHours: "0",
			}, nil)
			svc := &PaymentService{entClient: client, affiliateService: NewAffiliateService(repo, settings, nil, nil)}
			order := &dbent.PaymentOrder{ID: 3, UserID: 2, OrderType: payment.OrderTypeBalance, Amount: tt.amount, PayAmount: tt.paid, ProviderSnapshot: tt.snapshot}
			mock.ExpectBegin()
			mock.ExpectQuery("INSERT INTO payment_audit_logs").
				WithArgs("3", fmt.Sprintf(`{"baseAmount":%g,"status":"reserved"}`, tt.base)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			mock.ExpectExec(`UPDATE "payment_audit_logs"`).
				WithArgs("AFFILIATE_REBATE_APPLIED", fmt.Sprintf(`{"baseAmount":%g,"rebateAmount":%g}`, tt.base, tt.base/10), "system", "3", "AFFILIATE_REBATE_APPLIED").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			require.NoError(t, svc.applyAffiliateRebateForOrder(context.Background(), order))
			require.Equal(t, tt.base/10, repo.amount)
			require.Equal(t, 1, repo.calls)
			// A repeated callback must not accrue the rebate again.
			mock.ExpectBegin()
			mock.ExpectQuery("INSERT INTO payment_audit_logs").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectRollback()
			require.NoError(t, svc.applyAffiliateRebateForOrder(context.Background(), order))
			require.Equal(t, 1, repo.calls)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
