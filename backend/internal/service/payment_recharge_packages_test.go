package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestResolveRechargePackageForOrderRequiresEnabledPackage(t *testing.T) {
	t.Parallel()

	svc := &PaymentService{}
	cfg := &PaymentConfig{
		MinAmount: 5,
		RechargePackages: []RechargePackage{
			{ID: "pkg-5", Enabled: true, PayAmount: 5, CreditedAmount: 5},
			{ID: "pkg-20", Enabled: false, PayAmount: 20, CreditedAmount: 23},
		},
	}

	_, _, err := svc.resolveRechargePackageForOrder(context.Background(), CreateOrderRequest{
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
	}, cfg)
	require.Error(t, err)
	require.Equal(t, "RECHARGE_PACKAGE_REQUIRED", infraerrors.Reason(err))

	_, _, err = svc.resolveRechargePackageForOrder(context.Background(), CreateOrderRequest{
		UserID:            1,
		OrderType:         payment.OrderTypeBalance,
		RechargePackageID: "pkg-20",
	}, cfg)
	require.Error(t, err)
	require.Equal(t, "RECHARGE_PACKAGE_NOT_AVAILABLE", infraerrors.Reason(err))

	pkg, available, err := svc.resolveRechargePackageForOrder(context.Background(), CreateOrderRequest{
		UserID:            1,
		OrderType:         payment.OrderTypeBalance,
		RechargePackageID: "pkg-5",
	}, cfg)
	require.NoError(t, err)
	require.True(t, available)
	require.Equal(t, "pkg-5", pkg.ID)
}

func TestRechargePackageCheckoutViewsReflectFirstRechargeBonusStatus(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "checkout")
	now := time.Date(2026, 6, 10, 8, 0, 0, 0, time.Local)
	svc := &PaymentService{
		entClient: client,
		now:       func() time.Time { return now },
	}
	packages := []RechargePackage{
		{ID: "pkg-50", Label: "50", Enabled: true, PayAmount: 50, CreditedAmount: 60, SortOrder: 10},
	}

	views, claimed, claimedAt, err := svc.RechargePackageCheckoutViews(ctx, user.ID, packages)
	require.NoError(t, err)
	require.False(t, claimed)
	require.Empty(t, claimedAt)
	require.Len(t, views, 1)
	require.Equal(t, 60.0, views[0].EffectiveCreditedAmount)
	require.Equal(t, 10.0, views[0].EffectiveBonusAmount)

	_, err = client.RedeemCode.Create().
		SetCode(firstRechargePackageBonusRedeemCode(user.ID)).
		SetType(RedeemTypeMonthlyRecharge).
		SetValue(10).
		SetStatus(StatusUsed).
		SetUsedBy(user.ID).
		SetUsedAt(now.UTC()).
		Save(ctx)
	require.NoError(t, err)

	views, claimed, claimedAt, err = svc.RechargePackageCheckoutViews(ctx, user.ID, packages)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NotEmpty(t, claimedAt)
	require.Equal(t, 50.0, views[0].EffectiveCreditedAmount)
	require.Equal(t, 0.0, views[0].EffectiveBonusAmount)
}

func TestAdjustBalanceOrderForRechargePackageClaimsOncePerUser(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "claim-once")
	svc := &PaymentService{entClient: client}

	first := createRechargePackageTestOrder(t, ctx, client, user, "first", 50, 60, "202606")
	adjusted, err := svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, first)
	require.NoError(t, err)
	require.Equal(t, 60.0, adjusted.Amount)
	require.True(t, firstRechargePackageBonusCodeExists(t, ctx, client, user.ID))
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, first.ID, monthlyRechargeBonusAudit))
	require.Equal(t, 0, monthlyRechargeBonusAuditCount(t, ctx, client, first.ID, welfareFirstRechargeBonusAudit))

	second := createRechargePackageTestOrder(t, ctx, client, user, "second", 50, 60, "202606")
	adjusted, err = svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, second)
	require.NoError(t, err)
	require.Equal(t, 50.0, adjusted.Amount)
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, second.ID, "MONTHLY_RECHARGE_BONUS_SKIPPED"))

	nextMonth := createRechargePackageTestOrder(t, ctx, client, user, "next-month", 50, 60, "202607")
	adjusted, err = svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, nextMonth)
	require.NoError(t, err)
	require.Equal(t, 50.0, adjusted.Amount)
	require.False(t, monthlyRechargeBonusCodeExists(t, ctx, client, user.ID, "202607"))
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, nextMonth.ID, "MONTHLY_RECHARGE_BONUS_SKIPPED"))
}

func TestAdjustBalanceOrderForRechargePackageZeroBonusStillConsumesFirstRecharge(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "zero-bonus-first")
	svc := &PaymentService{entClient: client}

	first := createRechargePackageTestOrder(t, ctx, client, user, "zero-bonus-first", 5, 5, "202606")
	adjusted, err := svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, first)
	require.NoError(t, err)
	require.Equal(t, 5.0, adjusted.Amount)
	require.True(t, firstRechargePackageBonusCodeExists(t, ctx, client, user.ID))
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, first.ID, monthlyRechargeBonusAudit))
	require.Equal(t, 0, monthlyRechargeBonusAuditCount(t, ctx, client, first.ID, welfareFirstRechargeBonusAudit))

	second := createRechargePackageTestOrder(t, ctx, client, user, "bonus-second", 50, 60, "202606")
	adjusted, err = svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, second)
	require.NoError(t, err)
	require.Equal(t, 50.0, adjusted.Amount)
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, second.ID, "MONTHLY_RECHARGE_BONUS_SKIPPED"))
}

func TestRechargePackageLegacyMonthlyClaimBlocksFirstRechargeBonus(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "legacy-monthly-claim")
	now := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	_, err := client.RedeemCode.Create().
		SetCode(monthlyRechargeBonusRedeemCode(user.ID, "202606")).
		SetType(RedeemTypeMonthlyRecharge).
		SetValue(10).
		SetStatus(StatusUsed).
		SetUsedBy(user.ID).
		SetUsedAt(now).
		Save(ctx)
	require.NoError(t, err)
	svc := &PaymentService{entClient: client}

	order := createRechargePackageTestOrder(t, ctx, client, user, "legacy-monthly-claim", 50, 60, "202607")
	adjusted, err := svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, order)

	require.NoError(t, err)
	require.Equal(t, 50.0, adjusted.Amount)
	require.False(t, firstRechargePackageBonusCodeExists(t, ctx, client, user.ID))
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, order.ID, "MONTHLY_RECHARGE_BONUS_SKIPPED"))
}

func TestMonthlyRechargeBonusClaimReleasedWhenBalanceRedeemFails(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "release-on-failure")
	order := createRechargePackageTestOrder(t, ctx, client, user, "release-on-failure", 50, 60, "202606")
	userRepo := &rechargePackageUserRepo{
		user:             &User{ID: user.ID, Email: user.Email, Username: user.Username, Balance: 0},
		updateBalanceErr: infraerrors.BadRequest("REDEEM_FAILED", "redeem failed"),
	}
	redeemRepo := &rechargePackageRedeemRepo{client: client}
	redeemService := NewRedeemService(redeemRepo, userRepo, nil, nil, nil, client, nil, nil)
	svc := &PaymentService{
		entClient:     client,
		redeemService: redeemService,
		userRepo:      userRepo,
	}

	err := svc.doBalance(ctx, order)

	require.Error(t, err)
	require.False(t, firstRechargePackageBonusCodeExists(t, ctx, client, user.ID))
	require.Equal(t, 0, monthlyRechargeBonusAuditCount(t, ctx, client, order.ID, monthlyRechargeBonusAudit))
	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, 60.0, reloaded.Amount)
}

func TestMonthlyRechargeBonusClaimRetrySameOrderKeepsCreditedAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "same-order-retry")
	order := createRechargePackageTestOrder(t, ctx, client, user, "same-order-retry", 50, 60, "202606")
	svc := &PaymentService{entClient: client}

	adjusted, err := svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, order)
	require.NoError(t, err)
	require.Equal(t, 60.0, adjusted.Amount)

	adjusted, err = svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, adjusted)
	require.NoError(t, err)
	require.Equal(t, 60.0, adjusted.Amount)
	require.Equal(t, 1, monthlyRechargeBonusAuditCount(t, ctx, client, order.ID, monthlyRechargeBonusAudit))
}

func TestMonthlyRechargeBonusDoesNotMarkOrderAsFirstRechargeBonus(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := createRechargePackageTestUser(t, ctx, client, "not-first-bonus")
	order := createRechargePackageTestOrder(t, ctx, client, user, "not-first-bonus", 50, 60, "202606")
	svc := &PaymentService{entClient: client}

	_, err := svc.adjustBalanceOrderForMonthlyRechargePackage(ctx, order)
	require.NoError(t, err)

	welfareSvc := NewWelfareService(nil, nil, nil, nil, client, nil, nil)
	hasBonus, err := welfareSvc.OrderHasFirstRechargeBonus(ctx, order.ID)
	require.NoError(t, err)
	require.False(t, hasBonus)
}

type rechargePackageUserRepo struct {
	UserRepository
	user             *User
	updateBalanceErr error
}

func (r *rechargePackageUserRepo) GetByID(context.Context, int64) (*User, error) {
	if r.user == nil {
		return &User{}, nil
	}
	cloned := *r.user
	return &cloned, nil
}

func (r *rechargePackageUserRepo) UpdateBalance(context.Context, int64, float64) error {
	return r.updateBalanceErr
}

type rechargePackageRedeemRepo struct {
	RedeemCodeRepository
	client *dbent.Client
}

func (r *rechargePackageRedeemRepo) Create(ctx context.Context, code *RedeemCode) error {
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

func (r *rechargePackageRedeemRepo) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
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
	return redeemCodeFromEnt(found), nil
}

func (r *rechargePackageRedeemRepo) Use(ctx context.Context, id, userID int64) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	now := time.Now().UTC()
	affected, err := client.RedeemCode.Update().
		Where(redeemcode.IDEQ(id), redeemcode.StatusEQ(StatusUnused)).
		SetStatus(StatusUsed).
		SetUsedBy(userID).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRedeemCodeUsed
	}
	return nil
}

func redeemCodeFromEnt(found *dbent.RedeemCode) *RedeemCode {
	if found == nil {
		return nil
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
	}
}

func createRechargePackageTestUser(t *testing.T, ctx context.Context, client *dbent.Client, suffix string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail("recharge-package-" + suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername("recharge-package-" + suffix).
		Save(ctx)
	require.NoError(t, err)
	return user
}

func createRechargePackageTestOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *dbent.User, suffix string, payAmount, creditedAmount float64, period string) *dbent.PaymentOrder {
	t.Helper()
	now := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(creditedAmount).
		SetPayAmount(payAmount).
		SetFeeRate(0).
		SetRechargeCode("MRB-ORDER-" + suffix).
		SetOutTradeNo("sub2_mrb_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-" + suffix).
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusRecharging).
		SetExpiresAt(now.Add(time.Hour)).
		SetPaidAt(now).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderSnapshot(map[string]any{
			"schema_version":                   2,
			"recharge_package_id":              "pkg-50",
			"recharge_package_label":           "50",
			"recharge_package_pay_amount":      payAmount,
			"recharge_package_credited_amount": creditedAmount,
			"monthly_recharge_bonus_period":    period,
			"recharge_package_sort_order":      10,
		}).
		Save(ctx)
	require.NoError(t, err)
	return order
}

func monthlyRechargeBonusCodeExists(t *testing.T, ctx context.Context, client *dbent.Client, userID int64, period string) bool {
	t.Helper()
	exists, err := client.RedeemCode.Query().
		Where(redeemcode.CodeEQ(monthlyRechargeBonusRedeemCode(userID, period))).
		Exist(ctx)
	require.NoError(t, err)
	return exists
}

func firstRechargePackageBonusCodeExists(t *testing.T, ctx context.Context, client *dbent.Client, userID int64) bool {
	t.Helper()
	exists, err := client.RedeemCode.Query().
		Where(redeemcode.CodeEQ(firstRechargePackageBonusRedeemCode(userID))).
		Exist(ctx)
	require.NoError(t, err)
	return exists
}

func monthlyRechargeBonusAuditCount(t *testing.T, ctx context.Context, client *dbent.Client, orderID int64, action string) int {
	t.Helper()
	count, err := client.PaymentAuditLog.Query().
		Where(
			paymentauditlog.OrderIDEQ(strconv.FormatInt(orderID, 10)),
			paymentauditlog.ActionEQ(action),
		).
		Count(ctx)
	require.NoError(t, err)
	return count
}

func TestBuildWeChatPaymentOAuthStartURLIncludesRechargePackageID(t *testing.T) {
	t.Parallel()

	got, err := buildWeChatPaymentOAuthStartURL(CreateOrderRequest{
		PaymentType:       payment.TypeWxpay,
		OrderType:         payment.OrderTypeBalance,
		Amount:            50,
		RechargePackageID: "pkg-50",
	}, "snsapi_base")

	require.NoError(t, err)
	require.Contains(t, got, "recharge_package_id=pkg-50")
	require.Contains(t, got, "amount=50")
}
