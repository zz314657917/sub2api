package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestWelfareOverviewIncludesRechargeFirstBonusStatus(t *testing.T) {
	claimedAt := time.Date(2026, 6, 10, 2, 3, 4, 0, time.UTC)
	userID := int64(42)
	usedBy := userID
	redeemRepo := &welfareRedeemRepoStub{created: []RedeemCode{
		{
			Code:   firstRechargeBonusRedeemCode(userID),
			Type:   RedeemTypeFirstRechargeBonus,
			Value:  5,
			Status: StatusUsed,
			UsedBy: &usedBy,
			UsedAt: &claimedAt,
		},
	}}
	svc := NewWelfareService(&welfareRepoStub{}, &welfareUserRepoStub{}, redeemRepo, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	got, err := svc.GetOverview(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, got.Recharge)
	require.True(t, got.Modules.Recharge)
	require.True(t, got.Recharge.Enabled)
	require.Equal(t, 5.0, got.Recharge.FirstBonusAmount)
	require.True(t, got.Recharge.FirstBonusClaimed)
	require.Equal(t, welfareReasonAlreadyClaimed, got.Recharge.Reason)
	require.Equal(t, claimedAt.Format(time.RFC3339), got.Recharge.FirstBonusClaimedAt)
}

func TestWelfareGrantFirstRechargeBonusSkipsWhenDisabledOrZeroAmount(t *testing.T) {
	cases := []struct {
		name            string
		welfareEnabled  bool
		rechargeEnabled bool
		amount          float64
	}{
		{name: "welfare disabled", welfareEnabled: false, rechargeEnabled: true, amount: 5},
		{name: "recharge disabled", welfareEnabled: true, rechargeEnabled: false, amount: 5},
		{name: "zero amount", welfareEnabled: true, rechargeEnabled: true, amount: 0},
	}

	for idx, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			user := createFirstRechargeBonusUser(t, ctx, client, fmt.Sprintf("skip-%d", idx))
			order := createFirstRechargeBonusOrder(t, ctx, client, user, 10, payment.OrderTypeBalance, OrderStatusPaid, fmt.Sprintf("skip-%d", idx))
			userRepo := &welfareUserRepoStub{}
			redeemRepo := &welfareRedeemRepoStub{}
			svc := NewWelfareService(&welfareRepoStub{}, userRepo, redeemRepo, firstRechargeBonusSettingRepo(tt.welfareEnabled, tt.rechargeEnabled, tt.amount), client, nil, nil)

			granted, err := svc.GrantFirstRechargeBonusForOrder(ctx, order.ID)

			require.NoError(t, err)
			require.False(t, granted)
			require.Empty(t, redeemRepo.created)
			require.Empty(t, userRepo.grants)
			require.False(t, firstRechargeBonusAuditExists(t, ctx, client, order.ID))
		})
	}
}

func TestWelfareGrantFirstRechargeBonusGrantsFirstBalanceOrderOnce(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createFirstRechargeBonusUser(t, ctx, client, "grant-once")
	order := createFirstRechargeBonusOrder(t, ctx, client, user, 12.5, payment.OrderTypeBalance, OrderStatusPaid, "grant-once")
	userRepo := &welfareUserRepoStub{}
	redeemRepo := &welfareRedeemRepoStub{}
	authInvalidator := &welfareAuthInvalidatorStub{}
	balanceInvalidator := &welfareBalanceInvalidatorStub{}
	svc := NewWelfareService(&welfareRepoStub{}, userRepo, redeemRepo, firstRechargeBonusSettingRepo(true, true, 5), client, authInvalidator, nil)
	svc.billingCacheInvalidator = balanceInvalidator
	grantTime := time.Date(2026, 6, 10, 8, 9, 10, 0, time.UTC)
	svc.now = func() time.Time { return grantTime }

	granted, err := svc.GrantFirstRechargeBonusForOrder(ctx, order.ID)

	require.NoError(t, err)
	require.True(t, granted)
	require.Equal(t, []float64{5}, userRepo.grants)
	require.Empty(t, userRepo.updates)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, firstRechargeBonusRedeemCode(user.ID), redeemRepo.created[0].Code)
	require.Equal(t, RedeemTypeFirstRechargeBonus, redeemRepo.created[0].Type)
	require.Equal(t, StatusUsed, redeemRepo.created[0].Status)
	require.Equal(t, 5.0, redeemRepo.created[0].Value)
	require.Equal(t, user.ID, *redeemRepo.created[0].UsedBy)
	require.NotNil(t, redeemRepo.created[0].UsedAt)
	require.Equal(t, grantTime, *redeemRepo.created[0].UsedAt)
	require.Equal(t, []int64{user.ID}, authInvalidator.userIDs)
	require.Equal(t, []int64{user.ID}, balanceInvalidator.userIDs)

	hasBonus, err := svc.OrderHasFirstRechargeBonus(ctx, order.ID)
	require.NoError(t, err)
	require.True(t, hasBonus)
	require.Equal(t, 1, firstRechargeBonusAuditCount(t, ctx, client, order.ID))

	granted, err = svc.GrantFirstRechargeBonusForOrder(ctx, order.ID)

	require.NoError(t, err)
	require.False(t, granted)
	require.Equal(t, []float64{5}, userRepo.grants)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, 1, firstRechargeBonusAuditCount(t, ctx, client, order.ID))
}

func TestWelfareGrantFirstRechargeBonusDoesNotIncreaseTotalRecharged(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createFirstRechargeBonusUser(t, ctx, client, "no-total-recharged")
	require.NoError(t, client.User.UpdateOneID(user.ID).SetTotalRecharged(12.5).Exec(ctx))
	order := createFirstRechargeBonusOrder(t, ctx, client, user, 20, payment.OrderTypeBalance, OrderStatusPaid, "no-total-recharged")
	svc := NewWelfareService(
		&welfareRepoStub{},
		&firstRechargeBonusEntUserRepo{client: client},
		&firstRechargeBonusEntRedeemRepo{client: client},
		firstRechargeBonusSettingRepo(true, true, 5),
		client,
		nil,
		nil,
	)

	granted, err := svc.GrantFirstRechargeBonusForOrder(ctx, order.ID)

	require.NoError(t, err)
	require.True(t, granted)
	reloaded, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 5.0, reloaded.Balance)
	require.Equal(t, 12.5, reloaded.TotalRecharged)
}

func TestWelfareGrantFirstRechargeBonusOnlyFirstPositiveBalanceOrder(t *testing.T) {
	cases := []struct {
		name        string
		amount      float64
		orderType   string
		withEarlier bool
	}{
		{name: "second balance order", amount: 10, orderType: payment.OrderTypeBalance, withEarlier: true},
		{name: "zero balance order", amount: 0, orderType: payment.OrderTypeBalance},
		{name: "subscription order", amount: 10, orderType: payment.OrderTypeSubscription},
	}

	for idx, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			user := createFirstRechargeBonusUser(t, ctx, client, fmt.Sprintf("only-first-%d", idx))
			if tt.withEarlier {
				createFirstRechargeBonusOrder(t, ctx, client, user, 1, payment.OrderTypeBalance, OrderStatusCompleted, fmt.Sprintf("only-first-%d-earlier", idx))
			}
			order := createFirstRechargeBonusOrder(t, ctx, client, user, tt.amount, tt.orderType, OrderStatusPaid, fmt.Sprintf("only-first-%d-current", idx))
			userRepo := &welfareUserRepoStub{}
			redeemRepo := &welfareRedeemRepoStub{}
			svc := NewWelfareService(&welfareRepoStub{}, userRepo, redeemRepo, firstRechargeBonusSettingRepo(true, true, 5), client, nil, nil)

			granted, err := svc.GrantFirstRechargeBonusForOrder(ctx, order.ID)

			require.NoError(t, err)
			require.False(t, granted)
			require.Empty(t, redeemRepo.created)
			require.Empty(t, userRepo.grants)
			require.False(t, firstRechargeBonusAuditExists(t, ctx, client, order.ID))
		})
	}
}

func TestPaymentFirstRechargeBonusRefundUnsupportedForUserAndAdmin(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createFirstRechargeBonusUser(t, ctx, client, "refund")
	order := createFirstRechargeBonusOrder(t, ctx, client, user, 10, payment.OrderTypeBalance, OrderStatusCompleted, "refund")
	require.NoError(t, writeFirstRechargeBonusTestAudit(ctx, client, order.ID))
	welfareSvc := NewWelfareService(nil, nil, nil, nil, client, nil, nil)
	svc := &PaymentService{
		entClient:      client,
		welfareService: welfareSvc,
	}

	err := svc.RequestRefund(ctx, order.ID, user.ID, "user refund")
	require.Error(t, err)
	require.Equal(t, "FIRST_RECHARGE_BONUS_REFUND_UNSUPPORTED", infraerrors.Reason(err))

	plan, result, err := svc.PrepareRefund(ctx, order.ID, 0, "admin refund", false, false)
	require.Nil(t, plan)
	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "FIRST_RECHARGE_BONUS_REFUND_UNSUPPORTED", infraerrors.Reason(err))
}

func firstRechargeBonusSettingRepo(welfareEnabled, rechargeEnabled bool, amount float64) *welfareSettingRepoStub {
	repo := welfareSettingRepo(welfareEnabled, true, 1, 1)
	repo.values[SettingKeyWelfareRechargeEnabled] = welfareBool(rechargeEnabled)
	repo.values[SettingKeyWelfareFirstRechargeBonusAmount] = welfareFloat(amount)
	return repo
}

func createFirstRechargeBonusUser(t *testing.T, ctx context.Context, client *dbent.Client, suffix string) *dbent.User {
	t.Helper()
	suffix = sanitizeFirstRechargeTestSuffix(suffix)
	user, err := client.User.Create().
		SetEmail("first-recharge-" + suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername("first-recharge-" + suffix).
		Save(ctx)
	require.NoError(t, err)
	return user
}

func createFirstRechargeBonusOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *dbent.User, amount float64, orderType string, status string, suffix string) *dbent.PaymentOrder {
	t.Helper()
	suffix = sanitizeFirstRechargeTestSuffix(suffix)
	now := time.Date(2026, 6, 10, 1, 2, 3, 0, time.UTC)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(amount).
		SetPayAmount(amount).
		SetFeeRate(0).
		SetRechargeCode("FRB-ORDER-" + suffix).
		SetOutTradeNo("sub2_frb_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-" + suffix).
		SetOrderType(orderType).
		SetStatus(status).
		SetExpiresAt(now.Add(time.Hour)).
		SetPaidAt(now).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	return order
}

func sanitizeFirstRechargeTestSuffix(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("/", "-", " ", "-", "_", "-", "#", "-").Replace(value)
	if len(value) > 42 {
		value = value[:42]
	}
	return value
}

func writeFirstRechargeBonusTestAudit(ctx context.Context, client *dbent.Client, orderID int64) error {
	_, err := client.PaymentAuditLog.Create().
		SetOrderID(strconv.FormatInt(orderID, 10)).
		SetAction(welfareFirstRechargeBonusAudit).
		SetDetail(`{"amount":5,"redeemCode":"FRBTEST"}`).
		SetOperator("system").
		Save(ctx)
	return err
}

func firstRechargeBonusAuditExists(t *testing.T, ctx context.Context, client *dbent.Client, orderID int64) bool {
	t.Helper()
	return firstRechargeBonusAuditCount(t, ctx, client, orderID) > 0
}

func firstRechargeBonusAuditCount(t *testing.T, ctx context.Context, client *dbent.Client, orderID int64) int {
	t.Helper()
	count, err := client.PaymentAuditLog.Query().
		Where(
			paymentauditlog.OrderIDEQ(strconv.FormatInt(orderID, 10)),
			paymentauditlog.ActionEQ(welfareFirstRechargeBonusAudit),
		).
		Count(ctx)
	require.NoError(t, err)
	return count
}

type firstRechargeBonusEntUserRepo struct {
	UserRepository
	client *dbent.Client
}

func (r *firstRechargeBonusEntUserRepo) AddBalance(ctx context.Context, id int64, amount float64) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	affected, err := client.User.Update().Where(dbuser.IDEQ(id)).AddBalance(amount).Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

type firstRechargeBonusEntRedeemRepo struct {
	RedeemCodeRepository
	client *dbent.Client
}

func (r *firstRechargeBonusEntRedeemRepo) Create(ctx context.Context, code *RedeemCode) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	created, err := client.RedeemCode.Create().
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays).
		SetNillableExpiresAt(code.ExpiresAt).
		SetNillableUsedBy(code.UsedBy).
		SetNillableUsedAt(code.UsedAt).
		SetNillableGroupID(code.GroupID).
		Save(ctx)
	if err != nil {
		return err
	}
	code.ID = created.ID
	code.CreatedAt = created.CreatedAt
	return nil
}

func (r *firstRechargeBonusEntRedeemRepo) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	found, err := client.RedeemCode.Query().Where(redeemcode.CodeEQ(code)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrRedeemCodeNotFound
		}
		return nil, err
	}
	notes := ""
	if found.Notes != nil {
		notes = *found.Notes
	}
	return &RedeemCode{
		ID:           found.ID,
		Code:         found.Code,
		Type:         found.Type,
		Value:        found.Value,
		Status:       found.Status,
		UsedBy:       found.UsedBy,
		UsedAt:       found.UsedAt,
		Notes:        notes,
		CreatedAt:    found.CreatedAt,
		ExpiresAt:    found.ExpiresAt,
		GroupID:      found.GroupID,
		ValidityDays: found.ValidityDays,
	}, nil
}
