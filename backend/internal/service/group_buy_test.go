package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestGroupBuyAdminCreatePlanAcceptsTierRanges(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_plan_tier_mapping")
	groupRepo := newGroupBuyGroupRepoStub()
	svc := newGroupBuyTestService(client, groupRepo, nil)

	groupLow := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	groupHigh := createGroupBuyTestGroup(t, ctx, client, 2, 300)
	groupRepo.groups[groupLow] = &Group{ID: groupLow, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription}
	groupRepo.groups[groupHigh] = &Group{ID: groupHigh, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription}

	_, err := svc.AdminCreatePlan(ctx, GroupBuyPlanInput{
		Title:              "Token拼拼拼 10 份团",
		TotalShares:        10,
		PricePerShare:      12,
		QuotaPerShareLabel: "每份 50 USD/月",
		MaxSharesPerUser:   10,
		TierRules: []GroupBuyTierInput{
			{MinShares: 1, MaxShares: 4, TargetGroupID: groupLow, Label: "基础档"},
			{MinShares: 6, MaxShares: 10, TargetGroupID: groupHigh, Label: "高阶档"},
		},
		ValidityDays:   30,
		TimeoutMinutes: 60,
		Status:         GroupBuyPlanStatusActive,
	})
	require.ErrorIs(t, err, ErrGroupBuyTierMappingInvalid)

	plan, err := svc.AdminCreatePlan(ctx, GroupBuyPlanInput{
		Title:              "Token拼拼拼 10 份团",
		TotalShares:        10,
		PricePerShare:      12,
		QuotaPerShareLabel: "每份 50 USD/月",
		MaxSharesPerUser:   10,
		TierRules: []GroupBuyTierInput{
			{MinShares: 1, MaxShares: 4, TargetGroupID: groupLow, Label: "基础档"},
			{MinShares: 5, MaxShares: 10, TargetGroupID: groupHigh, Label: "高阶档"},
		},
		ValidityDays:   30,
		TimeoutMinutes: 60,
		LaunchMode:     GroupBuyLaunchModeManual,
		Status:         GroupBuyPlanStatusActive,
	})
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, groupHigh, plan.TargetGroupID)
	require.Len(t, plan.TierRules, 2)
	require.Equal(t, "基础档", plan.TierRules[0].Label)
	require.Equal(t, "高阶档", plan.TierRules[1].Label)
	require.Equal(t, 10, len(plan.TierGroupIDs))
}

func TestGroupBuyTierResolverRejectsGapsOverlapsAndInvalidGroups(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_tier_resolver")
	groupRepo := newGroupBuyGroupRepoStub()
	svc := newGroupBuyTestService(client, groupRepo, nil)

	groupA := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	groupB := createGroupBuyTestGroup(t, ctx, client, 5, 500)
	groupRepo.groups[groupA] = &Group{ID: groupA, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription}
	groupRepo.groups[groupB] = &Group{ID: groupB, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription}

	valid := []GroupBuyTierInput{
		{MinShares: 1, MaxShares: 3, TargetGroupID: groupA, Label: "1-3 份"},
		{MinShares: 4, MaxShares: 10, TargetGroupID: groupB, Label: "4-10 份"},
	}
	require.NoError(t, validateTierRuleShape(valid, 10))
	rule, ok := resolveTierRuleForShareCount(tierRuleInputsToDomain(valid), 3)
	require.True(t, ok)
	require.Equal(t, groupA, rule.TargetGroupID)
	rule, ok = resolveTierRuleForShareCount(tierRuleInputsToDomain(valid), 4)
	require.True(t, ok)
	require.Equal(t, groupB, rule.TargetGroupID)
	require.NoError(t, svc.validateTierRules(ctx, valid))

	require.ErrorIs(t, validateTierRuleShape([]GroupBuyTierInput{
		{MinShares: 1, MaxShares: 3, TargetGroupID: groupA},
		{MinShares: 5, MaxShares: 10, TargetGroupID: groupB},
	}, 10), ErrGroupBuyTierMappingInvalid)
	require.ErrorIs(t, validateTierRuleShape([]GroupBuyTierInput{
		{MinShares: 1, MaxShares: 5, TargetGroupID: groupA},
		{MinShares: 5, MaxShares: 10, TargetGroupID: groupB},
	}, 10), ErrGroupBuyTierMappingInvalid)
	require.ErrorIs(t, svc.validateTierRules(ctx, []GroupBuyTierInput{
		{MinShares: 1, MaxShares: 10, TargetGroupID: 99999},
	}), ErrGroupBuyTargetGroupInvalid)
}

func TestGroupBuyManualPlanWithoutOpenRoundRejectsShareOrder(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_manual_no_round")
	user := createGroupBuyTestUser(t, ctx, client, "manual-no-round@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	paymentSvc := &PaymentService{
		entClient:     client,
		configService: &PaymentConfigService{},
		userRepo:      &groupBuyUserRepoStub{users: map[int64]*User{user.ID: user}},
		now:           time.Now,
	}
	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.paymentSvc = paymentSvc

	_, _, _, err := svc.lockSharesAndCreateOrder(ctx, CreateOrderRequest{
		UserID:      user.ID,
		PaymentType: payment.TypeStripe,
		OrderType:   payment.OrderTypeGroupBuy,
		PlanID:      plan.ID,
		ClientIP:    "127.0.0.1",
		SrcHost:     "example.test",
	}, user, plan, 1, &PaymentConfig{
		Enabled:          true,
		OrderTimeoutMin:  30,
		MaxPendingOrders: 3,
	}, 0, 12, &payment.InstanceSelection{
		InstanceID:  "stripe-test",
		ProviderKey: payment.TypeStripe,
		Config:      map[string]string{"currency": payment.DefaultPaymentCurrency},
	})
	require.ErrorIs(t, err, ErrGroupBuyRoundUnavailable)

	roundCount, err := client.GroupBuyRound.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, roundCount)
}

func TestGroupBuyDisabledBlocksUserSurfaceButKeepsAdminPlans(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_disabled")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.settingSvc = NewSettingService(&groupBuySettingRepoStub{
		values: map[string]string{SettingKeyGroupBuyEnabled: "false"},
	}, nil)

	_, err := svc.ListPlans(ctx, false)
	require.ErrorIs(t, err, ErrGroupBuyDisabled)

	adminPlans, err := svc.ListPlans(ctx, true)
	require.NoError(t, err)
	require.Len(t, adminPlans, 1)
}

func TestGroupBuyAdminPlanViewExposesFulfillmentMode(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_plan_fulfillment_mode")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 5)
	_, err := client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetFulfillmentMode("room_subscription").
		Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	plans, err := svc.ListPlans(ctx, true)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	require.Equal(t, "room_subscription", plans[0].FulfillmentMode)
}

func TestGroupBuyUserEndpointsExcludeRoomSubscriptionPlans(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_user_excludes_room_subscription")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 3)
	_, err := client.GroupBuyPlan.UpdateOneID(plan.ID).SetFulfillmentMode(CafeRoomFulfillmentMode).Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	userPlans, err := svc.ListPlans(ctx, false)
	require.NoError(t, err)
	require.Empty(t, userPlans)
	_, err = svc.loadAvailablePlan(ctx, plan.ID)
	require.ErrorIs(t, err, ErrGroupBuyPlanNotFound)

	adminPlans, err := svc.ListPlans(ctx, true)
	require.NoError(t, err)
	require.Len(t, adminPlans, 1)
	require.Equal(t, CafeRoomFulfillmentMode, adminPlans[0].FulfillmentMode)
}

func TestRefreshExpiredEntitlementsIgnoresCafeSeats(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_refresh_excludes_cafe")
	activatedAt := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, activatedAt, 1)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(fixture.groupID), nil)
	svc.now = func() time.Time { return activatedAt.AddDate(0, 0, fixture.plan.ValidityDays) }
	count, err := svc.RefreshExpiredEntitlements(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	_, err = svc.RefreshUserEntitlement(ctx, seat.UserID)
	require.NoError(t, err)

	subscriptions, err := client.UserSubscription.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, subscriptions)
}

func TestGroupBuyRefreshUserEntitlementAggregatesSharesAndExpiresToInactive(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_entitlement_refresh")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "entitlement@example.com")
	group2 := createGroupBuyTestGroup(t, ctx, client, 2, 100)
	group3 := createGroupBuyTestGroup(t, ctx, client, 3, 150)
	tierMap := groupBuyTestTierMap(group2)
	tierMap["3"] = group3
	plan := createGroupBuyTestPlanWithTierMap(t, ctx, client, tierMap, GroupBuyLaunchModeManual, 10)
	round1 := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	round2 := createGroupBuyTestRound(t, ctx, client, plan.ID, now.Add(time.Hour), 10)

	createGroupBuyTestActiveSeat(t, ctx, client, plan.ID, round1.ID, user.ID, 1, now.Add(-time.Hour), now.AddDate(0, 0, 10))
	createGroupBuyTestActiveSeat(t, ctx, client, plan.ID, round2.ID, user.ID, 2, now.Add(-30*time.Minute), now.AddDate(0, 0, 20))

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroups(map[int64]*Group{
		group2: {ID: group2, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
		group3: {ID: group3, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
	}), nil)
	svc.now = func() time.Time { return now }

	ent, err := svc.RefreshUserEntitlement(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, ent)
	require.Equal(t, GroupBuyEntitlementStatusActive, ent.Status)
	require.Equal(t, 3, ent.ActiveShareCount)
	require.NotNil(t, ent.TargetGroupID)
	require.Equal(t, group3, *ent.TargetGroupID)
	require.NotNil(t, ent.SubscriptionID)

	sub, err := client.UserSubscription.Get(ctx, *ent.SubscriptionID)
	require.NoError(t, err)
	require.Equal(t, group3, sub.GroupID)
	require.Equal(t, SubscriptionStatusActive, sub.Status)

	svc.now = func() time.Time { return now.AddDate(0, 0, 30) }
	inactive, err := svc.RefreshUserEntitlement(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, inactive)
	require.Equal(t, GroupBuyEntitlementStatusInactive, inactive.Status)
	require.Equal(t, 0, inactive.ActiveShareCount)
	require.Nil(t, inactive.TargetGroupID)
	require.Nil(t, inactive.SubscriptionID)

	expiredSub, err := client.UserSubscription.Get(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, SubscriptionStatusExpired, expiredSub.Status)
	require.False(t, expiredSub.ExpiresAt.After(svc.now()))
}

func TestGroupBuyRefreshUserEntitlementUsesLatestSeatPolicySnapshot(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_entitlement_policy_snapshot")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "snapshot@example.com")
	oldGroup := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	newGroup := createGroupBuyTestGroup(t, ctx, client, 2, 100)
	plan := createGroupBuyTestPlanWithTierRules(t, ctx, client, []domain.GroupBuyTierRule{
		{MinShares: 1, MaxShares: 10, TargetGroupID: newGroup, Label: "当前计划"},
	}, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	_, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetStatus(GroupBuySeatStatusActive).
		SetShareCount(2).
		SetActivatedAt(now.Add(-time.Hour)).
		SetExpiresAt(now.AddDate(0, 0, 10)).
		SetPolicySnapshot(domain.GroupBuyPolicySnapshot{
			ProductKey:   GroupBuyProductTokenPinPinPin,
			PlanID:       plan.ID,
			TotalShares:  10,
			ValidityDays: 30,
			RefundMode:   GroupBuyRefundModeBalanceCredit,
			TierRules:    []domain.GroupBuyTierRule{{MinShares: 1, MaxShares: 10, TargetGroupID: oldGroup, Label: "购买时策略"}},
		}).Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroups(map[int64]*Group{
		oldGroup: {ID: oldGroup, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
		newGroup: {ID: newGroup, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
	}), nil)
	svc.now = func() time.Time { return now }

	ent, err := svc.RefreshUserEntitlement(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, ent.TargetGroupID)
	require.Equal(t, oldGroup, *ent.TargetGroupID)
}

func TestGroupBuyRefreshUserEntitlementDoesNotOverwriteNormalSubscription(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_normal_subscription_isolation")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "normal-subscription@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlanWithTierRules(t, ctx, client, []domain.GroupBuyTierRule{
		{MinShares: 1, MaxShares: 10, TargetGroupID: groupID, Label: "通用权益"},
	}, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	createGroupBuyTestActiveSeat(t, ctx, client, plan.ID, round.ID, user.ID, 2, now.Add(-time.Hour), now.AddDate(0, 0, 10))

	normalSub, err := client.UserSubscription.Create().
		SetUserID(user.ID).
		SetGroupID(groupID).
		SetStartsAt(now.Add(-24 * time.Hour)).
		SetExpiresAt(now.AddDate(0, 1, 0)).
		SetStatus(SubscriptionStatusActive).
		SetSourceType("standard").
		SetNotes("normal paid subscription").
		Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.now = func() time.Time { return now }

	ent, err := svc.RefreshUserEntitlement(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, ent.SubscriptionID)
	require.NotEqual(t, normalSub.ID, *ent.SubscriptionID)

	reloadedNormal, err := client.UserSubscription.Get(ctx, normalSub.ID)
	require.NoError(t, err)
	require.Equal(t, SubscriptionStatusActive, reloadedNormal.Status)
	require.Equal(t, "standard", reloadedNormal.SourceType)
	require.False(t, reloadedNormal.ManagedByGroupBuy)

	managed, err := client.UserSubscription.Get(ctx, *ent.SubscriptionID)
	require.NoError(t, err)
	require.Equal(t, "group_buy", managed.SourceType)
	require.True(t, managed.ManagedByGroupBuy)
	require.NotNil(t, managed.SourceID)
	require.Equal(t, ent.ID, *managed.SourceID)
}

func TestGroupBuyRefundProcessingIsIdempotent(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_refund_idempotent")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "refund@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(24).
		SetPayAmount(24).
		SetRechargeCode("gb-refund").
		SetOutTradeNo("gb_refund_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("trade").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.test").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetShareCount(2).
		Save(ctx)
	require.NoError(t, err)

	userRepo := &groupBuyUserRepoStub{users: map[int64]*User{user.ID: user}}
	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.userRepo = userRepo
	svc.now = func() time.Time { return now }

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, outcome)
	outcome, err = svc.processSeatRefund(ctx, plan, seat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, outcome)
	require.Equal(t, 1, userRepo.balanceUpdates)
	require.Equal(t, 24.0, userRepo.balanceTotal)

	refundCount, err := client.GroupBuyRefund.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, refundCount)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, reloadedSeat.Status)
}

type groupBuyRefundExecution struct {
	orderStatus string
	result      *RefundResult
	err         error
}

type groupBuyPaymentRefundStub struct {
	client          *dbent.Client
	executions      []groupBuyRefundExecution
	queryStatus     string
	hasPendingAudit bool
	prepareCalls    int
	executeCalls    int
	queryCalls      int
}

func (s *groupBuyPaymentRefundStub) PrepareRefund(ctx context.Context, orderID int64, amount float64, reason string, _, _ bool) (*RefundPlan, *RefundResult, error) {
	s.prepareCalls++
	order, err := s.client.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}
	return &RefundPlan{OrderID: orderID, Order: order, RefundAmount: amount, Reason: reason}, nil, nil
}

func (s *groupBuyPaymentRefundStub) ExecuteRefund(ctx context.Context, plan *RefundPlan) (*RefundResult, error) {
	index := s.executeCalls
	s.executeCalls++
	if index >= len(s.executions) {
		return nil, errors.New("unexpected provider refund execution")
	}
	execution := s.executions[index]
	if execution.orderStatus != "" {
		if _, err := s.client.PaymentOrder.UpdateOneID(plan.OrderID).SetStatus(execution.orderStatus).Save(ctx); err != nil {
			return nil, err
		}
	}
	return execution.result, execution.err
}

func (s *groupBuyPaymentRefundStub) QueryAndFinalizeRefund(ctx context.Context, orderID int64) (*RefundResult, error) {
	s.queryCalls++
	if s.queryStatus != "" {
		if _, err := s.client.PaymentOrder.UpdateOneID(orderID).SetStatus(s.queryStatus).Save(ctx); err != nil {
			return nil, err
		}
	}
	succeeded := s.queryStatus == OrderStatusRefunded
	return &RefundResult{Success: succeeded}, nil
}

func (s *groupBuyPaymentRefundStub) hasAuditLog(_ context.Context, _ int64, action string) bool {
	return action == "REFUND_PENDING" && s.hasPendingAudit
}

func TestGroupBuyProviderRefundSuccessIsIdempotent(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_provider_refund_success")
	plan, seat, order := createGroupBuyProviderRefundFixture(t, ctx, client, "success")
	stub := &groupBuyPaymentRefundStub{
		client: client,
		executions: []groupBuyRefundExecution{{
			orderStatus: OrderStatusRefunded,
			result:      &RefundResult{Success: true},
		}},
	}
	svc := newGroupBuyTestService(client, nil, nil)
	svc.refundSvc = stub

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, outcome)

	reloadedSeat, err := client.GroupBuySeat.Query().Where(groupbuyseat.IDEQ(seat.ID)).WithOrder().Only(ctx)
	require.NoError(t, err)
	outcome, err = svc.processSeatRefund(ctx, plan, reloadedSeat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, outcome)
	require.Equal(t, 1, stub.executeCalls)
	require.Equal(t, 1, stub.prepareCalls)

	reloadedOrder, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, reloadedOrder.Status)
	reloadedSeat, err = client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, reloadedSeat.Status)
	refundCount, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, refundCount)
}

func TestGroupBuyProviderRefundPendingIsReconciled(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_provider_refund_pending")
	plan, seat, order := createGroupBuyProviderRefundFixture(t, ctx, client, "pending")
	stub := &groupBuyPaymentRefundStub{
		client:          client,
		hasPendingAudit: true,
		queryStatus:     OrderStatusRefunded,
		executions: []groupBuyRefundExecution{{
			orderStatus: OrderStatusRefundPending,
			result:      &RefundResult{Success: false, Warning: "pending"},
		}},
	}
	svc := newGroupBuyTestService(client, nil, nil)
	svc.refundSvc = stub

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusPendingProvider, outcome)

	finalized, err := svc.ReconcilePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, finalized)
	require.Equal(t, 1, stub.queryCalls)

	reloadedOrder, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, reloadedOrder.Status)
	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, refund.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, reloadedSeat.Status)
}

func TestGroupBuyProviderPartialRefundIsQuarantined(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_provider_partial_refund")
	plan, seat, _ := createGroupBuyProviderRefundFixture(t, ctx, client, "partial")
	stub := &groupBuyPaymentRefundStub{
		client: client,
		executions: []groupBuyRefundExecution{{
			orderStatus: OrderStatusPartiallyRefunded,
			result:      &RefundResult{Success: true},
		}},
	}
	svc := newGroupBuyTestService(client, nil, nil)
	svc.refundSvc = stub

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.Error(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, outcome)
	require.Equal(t, 1, stub.prepareCalls)
	require.Equal(t, 1, stub.executeCalls)

	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, refund.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundProcessing, reloadedSeat.Status)
}

func TestGroupBuyPendingProviderPartialRefundIsQuarantined(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_pending_provider_partial_refund")
	plan, seat, order := createGroupBuyProviderRefundFixture(t, ctx, client, "pending-partial")
	stub := &groupBuyPaymentRefundStub{
		client:          client,
		hasPendingAudit: true,
		queryStatus:     OrderStatusPartiallyRefunded,
		executions: []groupBuyRefundExecution{{
			orderStatus: OrderStatusRefundPending,
			result:      &RefundResult{Success: false, Warning: "pending"},
		}},
	}
	svc := newGroupBuyTestService(client, nil, nil)
	svc.refundSvc = stub

	_, err := svc.processSeatRefund(ctx, plan, seat)
	require.NoError(t, err)
	finalized, err := svc.ReconcilePendingProviderRefunds(ctx)
	require.Error(t, err)
	require.Equal(t, 1, finalized)
	require.Equal(t, 1, stub.queryCalls)

	reloadedOrder, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPartiallyRefunded, reloadedOrder.Status)
	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, refund.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundProcessing, reloadedSeat.Status)
}

func TestGroupBuyLatePaymentForReleasedSeatQueuesRefund(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_late_payment_refund")
	_, seat, _ := createGroupBuyProviderRefundFixture(t, ctx, client, "late-payment")
	now := time.Date(2026, 7, 30, 18, 30, 0, 0, time.UTC)
	_, err := client.GroupBuySeat.UpdateOneID(seat.ID).SetStatus(GroupBuySeatStatusReleased).Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, nil, nil)
	svc.now = func() time.Time { return now }
	require.NoError(t, svc.HandleGroupBuyOrderPaid(ctx, *seat.OrderID))

	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundPending, reloadedSeat.Status)
	require.NotNil(t, reloadedSeat.PaidAt)
	require.Equal(t, now, *reloadedSeat.PaidAt)
	require.Contains(t, *reloadedSeat.RefundNote, "待原路退款")
	queued, err := client.GroupBuyEvent.Query().
		Where(groupbuyevent.SeatIDEQ(seat.ID), groupbuyevent.EventTypeEQ(groupBuyEventRefundQueued)).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, "late_payment_after_round_timeout", queued.Metadata["reason"])
	refunds, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Count(ctx)
	require.NoError(t, err)
	require.Zero(t, refunds)
}

func TestGroupBuyProviderRefundFailureCanRetry(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_provider_refund_retry")
	plan, seat, _ := createGroupBuyProviderRefundFixture(t, ctx, client, "retry")
	stub := &groupBuyPaymentRefundStub{
		client: client,
		executions: []groupBuyRefundExecution{
			{orderStatus: OrderStatusRefundFailed, err: errors.New("provider refund failed")},
			{orderStatus: OrderStatusRefunded, result: &RefundResult{Success: true}},
		},
	}
	svc := newGroupBuyTestService(client, nil, nil)
	svc.refundSvc = stub

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.Error(t, err)
	require.Equal(t, GroupBuyRefundStatusFailed, outcome)

	reloadedSeat, err := client.GroupBuySeat.Query().Where(groupbuyseat.IDEQ(seat.ID)).WithOrder().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundPending, reloadedSeat.Status)
	outcome, err = svc.processSeatRefund(ctx, plan, reloadedSeat)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, outcome)
	require.Equal(t, 2, stub.prepareCalls)
	require.Equal(t, 2, stub.executeCalls)

	refundCount, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, refundCount)
}

func TestGroupBuyHistoricalBalanceRefundIsQuarantined(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_refund_needs_review")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "refund-review@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	_, err := client.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetStatus(GroupBuySeatStatusRefundProcessing).
		SetShareCount(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRefund.Create().
		SetSeatID(seat.ID).
		SetUserID(user.ID).
		SetMode(GroupBuyRefundModeBalanceCredit).
		SetStatus(GroupBuyRefundStatusProcessing).
		SetAmount(plan.PricePerShare).
		SetIdempotencyKey("historical-processing-" + strconv.FormatInt(seat.ID, 10)).
		Save(ctx)
	require.NoError(t, err)

	userRepo := &groupBuyUserRepoStub{users: map[int64]*User{user.ID: user}}
	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.userRepo = userRepo
	svc.now = func() time.Time { return now }

	result, err := svc.AdminProcessRefunds(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, 1, result.Failed)
	require.Zero(t, result.Processed)
	require.Len(t, result.Failures, 1)
	require.Equal(t, 0, userRepo.balanceUpdates)

	review, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, review.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundProcessing, reloadedSeat.Status)
}

func TestGroupBuyAlreadyRefundedBalanceOrderIsPersistedForReview(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_refunded_order_needs_review")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "refunded-order-review@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	_, err := client.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(12).
		SetPayAmount(12).
		SetRechargeCode("gb-refunded-review").
		SetOutTradeNo("gb_refunded_review_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("trade-refunded-review").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusRefunded).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.test").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetShareCount(1).
		SetPolicySnapshot(buildGroupBuyPolicySnapshot(plan, now)).
		Save(ctx)
	require.NoError(t, err)

	userRepo := &groupBuyUserRepoStub{users: map[int64]*User{user.ID: user}}
	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.userRepo = userRepo
	svc.now = func() time.Time { return now }

	result, err := svc.AdminProcessRefunds(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, 1, result.Failed)
	require.Zero(t, result.Processed)
	require.Equal(t, 0, userRepo.balanceUpdates)

	review, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundModeBalanceCredit, review.Mode)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, review.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundProcessing, reloadedSeat.Status)
}

func TestGroupBuyHistoricalProviderRefundIsQuarantinedBeforeRetry(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_provider_refund_needs_review")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "provider-review@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(12).
		SetPayAmount(12).
		SetRechargeCode("gb-provider-review").
		SetOutTradeNo("gb_provider_review_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("trade-provider-review").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.test").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundProcessing).
		SetShareCount(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRefund.Create().
		SetSeatID(seat.ID).
		SetOrderID(order.ID).
		SetUserID(user.ID).
		SetMode(GroupBuyRefundModeProviderRefund).
		SetStatus(GroupBuyRefundStatusProcessing).
		SetAmount(order.Amount).
		SetIdempotencyKey("provider-processing-" + strconv.FormatInt(seat.ID, 10)).
		Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.now = func() time.Time { return now }

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.Error(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, outcome)
	review, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusNeedsReview, review.Status)
}

func TestGroupBuyBalanceRefundRollsBackWhenBalanceGrantFails(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_refund_transaction_rollback")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "refund-rollback@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, now, 10)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetShareCount(1).
		Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.userRepo = &groupBuyFailingBalanceRepo{}
	svc.now = func() time.Time { return now }

	outcome, err := svc.processSeatRefund(ctx, plan, seat)
	require.Error(t, err)
	require.Equal(t, GroupBuyRefundStatusFailed, outcome)

	reloadedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, reloadedUser.Balance)
	refundCount, err := client.GroupBuyRefund.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, refundCount)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundPending, reloadedSeat.Status)
}

type groupBuyFailingBalanceRepo struct {
	UserRepository
}

func (r *groupBuyFailingBalanceRepo) AddBalance(ctx context.Context, userID int64, amount float64) error {
	tx := dbent.TxFromContext(ctx)
	if tx == nil {
		return errors.New("transaction context is missing")
	}
	_, err := tx.Client().User.Update().Where(user.IDEQ(userID)).AddBalance(amount).Save(ctx)
	if err != nil {
		return err
	}
	return errors.New("forced balance grant failure")
}

func TestGroupBuyExpiredLockedSeatReleasesSharesAndExpiresOrder(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_expired_lock_release")
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	user := createGroupBuyTestUser(t, ctx, client, "expired-lock@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetStatus(GroupBuyRoundStatusOpen).
		SetTotalShares(10).
		SetPaidShares(0).
		SetReservedShares(3).
		SetTotalSeats(10).
		SetPaidSeats(0).
		SetReservedSeats(3).
		SetDeadlineAt(now.Add(time.Hour)).
		SetStartedAt(now.Add(-time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(36).
		SetPayAmount(36).
		SetRechargeCode("gb-expire").
		SetOutTradeNo("gb_expire_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusPending).
		SetExpiresAt(now.Add(-time.Minute)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.test").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusLocked).
		SetShareCount(3).
		SetLockedUntil(now.Add(-time.Minute)).
		Save(ctx)
	require.NoError(t, err)

	svc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	svc.now = func() time.Time { return now }
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	released, err := svc.releaseExpiredLockedSeatsForRoundTx(dbent.NewTxContext(ctx, tx), tx, round.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, released)

	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusReleased, reloadedSeat.Status)
	reloadedOrder, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusExpired, reloadedOrder.Status)
	reloadedRound, err := client.GroupBuyRound.Get(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, 0, reloadedRound.ReservedShares)
	require.Equal(t, 0, reloadedRound.ReservedSeats)
}

type groupBuyUserRepoStub struct {
	users          map[int64]*User
	balanceUpdates int
	balanceTotal   float64
}

func (s *groupBuyUserRepoStub) Create(context.Context, *User) error { panic("unexpected Create call") }
func (s *groupBuyUserRepoStub) CreateWithEmailAliasGuard(context.Context, *User) error {
	panic("unexpected CreateWithEmailAliasGuard call")
}
func (s *groupBuyUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	if user := s.users[id]; user != nil {
		cp := *user
		return &cp, nil
	}
	return nil, ErrUserNotFound
}
func (s *groupBuyUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}
func (s *groupBuyUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}
func (s *groupBuyUserRepoStub) Update(context.Context, *User, UserUpdateFields) error {
	panic("unexpected Update call")
}
func (s *groupBuyUserRepoStub) Delete(context.Context, int64) error { panic("unexpected Delete call") }
func (s *groupBuyUserRepoStub) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}
func (s *groupBuyUserRepoStub) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}
func (s *groupBuyUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}
func (s *groupBuyUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *groupBuyUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *groupBuyUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}
func (s *groupBuyUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}
func (s *groupBuyUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}
func (s *groupBuyUserRepoStub) UpdateBalance(_ context.Context, _ int64, amount float64) error {
	s.balanceUpdates++
	s.balanceTotal += amount
	return nil
}
func (s *groupBuyUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}
func (s *groupBuyUserRepoStub) AdjustBalance(_ context.Context, id int64, delta float64) (BalanceChange, error) {
	user := s.users[id]
	if user == nil {
		return BalanceChange{}, ErrUserNotFound
	}
	change := BalanceChange{Old: user.Balance, New: user.Balance + delta}
	if change.New < 0 {
		return change, ErrBalanceNegative
	}
	user.Balance = change.New
	return change, nil
}
func (s *groupBuyUserRepoStub) SetBalance(_ context.Context, id int64, value float64) (BalanceChange, error) {
	user := s.users[id]
	if user == nil {
		return BalanceChange{}, ErrUserNotFound
	}
	if value < 0 {
		return BalanceChange{}, ErrBalanceNegative
	}
	change := BalanceChange{Old: user.Balance, New: value}
	user.Balance = value
	return change, nil
}
func (s *groupBuyUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}
func (s *groupBuyUserRepoStub) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchSetConcurrency call")
}
func (s *groupBuyUserRepoStub) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchAddConcurrency call")
}
func (s *groupBuyUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}
func (s *groupBuyUserRepoStub) ExistsByEmailAlias(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmailAlias call")
}
func (s *groupBuyUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}
func (s *groupBuyUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}
func (s *groupBuyUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}
func (s *groupBuyUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}
func (s *groupBuyUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}
func (s *groupBuyUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}
func (s *groupBuyUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}
func (s *groupBuyUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

type groupBuySettingRepoStub struct {
	values map[string]string
}

func (s *groupBuySettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *groupBuySettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *groupBuySettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *groupBuySettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *groupBuySettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *groupBuySettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *groupBuySettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type groupBuyGroupRepoStub struct {
	groups map[int64]*Group
}

func newGroupBuyGroupRepoStub() *groupBuyGroupRepoStub {
	return &groupBuyGroupRepoStub{groups: map[int64]*Group{}}
}

func newGroupBuyGroupRepoStubWithGroup(groupID int64) *groupBuyGroupRepoStub {
	return newGroupBuyGroupRepoStubWithGroups(map[int64]*Group{
		groupID: {ID: groupID, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription},
	})
}

func newGroupBuyGroupRepoStubWithGroups(groups map[int64]*Group) *groupBuyGroupRepoStub {
	return &groupBuyGroupRepoStub{groups: groups}
}

func (s *groupBuyGroupRepoStub) Create(context.Context, *Group) error {
	panic("unexpected Create call")
}
func (s *groupBuyGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	if group := s.groups[id]; group != nil {
		cp := *group
		return &cp, nil
	}
	return nil, ErrGroupNotFound
}
func (s *groupBuyGroupRepoStub) GetByIDLite(ctx context.Context, id int64) (*Group, error) {
	return s.GetByID(ctx, id)
}
func (s *groupBuyGroupRepoStub) Update(context.Context, *Group) error {
	panic("unexpected Update call")
}
func (s *groupBuyGroupRepoStub) Delete(context.Context, int64) error { panic("unexpected Delete call") }
func (s *groupBuyGroupRepoStub) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}
func (s *groupBuyGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *groupBuyGroupRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *groupBuyGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	out := make([]Group, 0, len(s.groups))
	for _, group := range s.groups {
		if group.Status == StatusActive {
			cp := *group
			out = append(out, cp)
		}
	}
	return out, nil
}
func (s *groupBuyGroupRepoStub) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}
func (s *groupBuyGroupRepoStub) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}
func (s *groupBuyGroupRepoStub) GetAccountCount(context.Context, int64) (int64, int64, error) {
	return 0, 0, nil
}
func (s *groupBuyGroupRepoStub) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}
func (s *groupBuyGroupRepoStub) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}
func (s *groupBuyGroupRepoStub) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}
func (s *groupBuyGroupRepoStub) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

type groupBuyUserSubRepoStub struct {
	client *dbent.Client
}

func (s *groupBuyUserSubRepoStub) Create(context.Context, *UserSubscription) error {
	panic("unexpected Create call")
}
func (s *groupBuyUserSubRepoStub) GetByID(context.Context, int64) (*UserSubscription, error) {
	panic("unexpected GetByID call")
}
func (s *groupBuyUserSubRepoStub) GetByIDForUpdate(context.Context, int64) (*UserSubscription, error) {
	panic("unexpected GetByIDForUpdate call")
}
func (s *groupBuyUserSubRepoStub) GetByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	panic("unexpected GetByUserIDAndGroupID call")
}
func (s *groupBuyUserSubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	sub, err := s.client.UserSubscription.Query().
		Where(
			usersubscription.UserIDEQ(userID),
			usersubscription.GroupIDEQ(groupID),
			usersubscription.StatusEQ(SubscriptionStatusActive),
			usersubscription.ExpiresAtGT(time.Now()),
		).
		WithGroup().
		Only(ctx)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}
	return groupBuyEntSubToService(sub), nil
}
func (s *groupBuyUserSubRepoStub) Update(context.Context, *UserSubscription) error {
	panic("unexpected Update call")
}
func (s *groupBuyUserSubRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}
func (s *groupBuyUserSubRepoStub) ListByUserID(context.Context, int64) ([]UserSubscription, error) {
	panic("unexpected ListByUserID call")
}
func (s *groupBuyUserSubRepoStub) ListActiveByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) {
	subs, err := s.client.UserSubscription.Query().
		Where(
			usersubscription.UserIDEQ(userID),
			usersubscription.StatusEQ(SubscriptionStatusActive),
			usersubscription.ExpiresAtGT(time.Now()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserSubscription, 0, len(subs))
	for _, sub := range subs {
		out = append(out, *groupBuyEntSubToService(sub))
	}
	return out, nil
}
func (s *groupBuyUserSubRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (s *groupBuyUserSubRepoStub) List(context.Context, pagination.PaginationParams, *int64, *int64, string, string, string, string) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *groupBuyUserSubRepoStub) ExistsByUserIDAndGroupID(context.Context, int64, int64) (bool, error) {
	panic("unexpected ExistsByUserIDAndGroupID call")
}
func (s *groupBuyUserSubRepoStub) ExtendExpiry(context.Context, int64, time.Time) error {
	panic("unexpected ExtendExpiry call")
}
func (s *groupBuyUserSubRepoStub) UpdateStatus(context.Context, int64, string) error {
	panic("unexpected UpdateStatus call")
}
func (s *groupBuyUserSubRepoStub) UpdateNotes(context.Context, int64, string) error {
	panic("unexpected UpdateNotes call")
}
func (s *groupBuyUserSubRepoStub) ActivateWindows(context.Context, int64, time.Time) error {
	panic("unexpected ActivateWindows call")
}
func (s *groupBuyUserSubRepoStub) ResetDailyUsage(context.Context, int64, time.Time) error {
	panic("unexpected ResetDailyUsage call")
}
func (s *groupBuyUserSubRepoStub) ResetWeeklyUsage(context.Context, int64, time.Time) error {
	panic("unexpected ResetWeeklyUsage call")
}
func (s *groupBuyUserSubRepoStub) ResetMonthlyUsage(context.Context, int64, time.Time) error {
	panic("unexpected ResetMonthlyUsage call")
}
func (s *groupBuyUserSubRepoStub) IncrementUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementUsage call")
}
func (s *groupBuyUserSubRepoStub) BatchUpdateExpiredStatus(context.Context) (int64, error) {
	panic("unexpected BatchUpdateExpiredStatus call")
}

type groupBuyAPIKeyRepoStub struct {
	keys map[int64]*APIKey
}

func (s *groupBuyAPIKeyRepoStub) Create(context.Context, *APIKey) error {
	panic("unexpected Create call")
}
func (s *groupBuyAPIKeyRepoStub) GetByID(_ context.Context, id int64) (*APIKey, error) {
	if key := s.keys[id]; key != nil {
		cp := *key
		return &cp, nil
	}
	return nil, ErrAPIKeyNotFound
}
func (s *groupBuyAPIKeyRepoStub) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}
func (s *groupBuyAPIKeyRepoStub) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}
func (s *groupBuyAPIKeyRepoStub) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}
func (s *groupBuyAPIKeyRepoStub) Update(_ context.Context, key *APIKey, _ APIKeyUpdateFields) error {
	cp := *key
	s.keys[key.ID] = &cp
	return nil
}
func (s *groupBuyAPIKeyRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}
func (s *groupBuyAPIKeyRepoStub) DeleteWithAudit(context.Context, int64) error {
	panic("unexpected DeleteWithAudit call")
}
func (s *groupBuyAPIKeyRepoStub) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}
func (s *groupBuyAPIKeyRepoStub) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}
func (s *groupBuyAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}
func (s *groupBuyAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	panic("unexpected ExistsByKey call")
}
func (s *groupBuyAPIKeyRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (s *groupBuyAPIKeyRepoStub) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}
func (s *groupBuyAPIKeyRepoStub) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}
func (s *groupBuyAPIKeyRepoStub) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}
func (s *groupBuyAPIKeyRepoStub) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}
func (s *groupBuyAPIKeyRepoStub) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}
func (s *groupBuyAPIKeyRepoStub) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}
func (s *groupBuyAPIKeyRepoStub) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}
func (s *groupBuyAPIKeyRepoStub) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}
func (s *groupBuyAPIKeyRepoStub) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}
func (s *groupBuyAPIKeyRepoStub) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}
func (s *groupBuyAPIKeyRepoStub) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

func newGroupBuyTestClient(t *testing.T, name string) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newGroupBuyTestService(client *dbent.Client, groupRepo GroupRepository, apiKeyRepo APIKeyRepository) *GroupBuyService {
	if groupRepo == nil {
		groupRepo = newGroupBuyGroupRepoStub()
	}
	userSubRepo := &groupBuyUserSubRepoStub{client: client}
	subSvc := NewSubscriptionService(groupRepo, userSubRepo, nil, client, nil)
	var apiKeySvc *APIKeyService
	if apiKeyRepo != nil {
		apiKeySvc = NewAPIKeyService(apiKeyRepo, &groupBuyUserRepoStub{users: map[int64]*User{}}, groupRepo, userSubRepo, nil, nil, nil)
	}
	return &GroupBuyService{
		entClient:       client,
		subscriptionSvc: subSvc,
		apiKeySvc:       apiKeySvc,
		groupRepo:       groupRepo,
		now:             time.Now,
	}
}

func createGroupBuyTestUser(t *testing.T, ctx context.Context, client *dbent.Client, email string) *User {
	t.Helper()
	entUser, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetUsername(email).
		SetStatus(StatusActive).
		Save(ctx)
	require.NoError(t, err)
	return &User{ID: entUser.ID, Email: entUser.Email, Username: entUser.Username, Status: entUser.Status}
}

func createGroupBuyTestGroup(t *testing.T, ctx context.Context, client *dbent.Client, share int, monthlyLimit float64) int64 {
	t.Helper()
	group, err := client.Group.Create().
		SetName("Token拼拼拼 " + strconv.Itoa(share) + " 份档").
		SetPlatform(PlatformOpenAI).
		SetStatus(StatusActive).
		SetSubscriptionType(domain.SubscriptionTypeSubscription).
		SetMonthlyLimitUsd(monthlyLimit).
		Save(ctx)
	require.NoError(t, err)
	return group.ID
}

func createGroupBuyTestPlan(t *testing.T, ctx context.Context, client *dbent.Client, groupID int64, launchMode string, totalShares int) *dbent.GroupBuyPlan {
	t.Helper()
	return createGroupBuyTestPlanWithTierMap(t, ctx, client, groupBuyTestTierMap(groupID), launchMode, totalShares)
}

func createGroupBuyTestPlanWithTierMap(t *testing.T, ctx context.Context, client *dbent.Client, tierMap map[string]int64, launchMode string, totalShares int) *dbent.GroupBuyPlan {
	t.Helper()
	targetGroupID := tierMap[strconv.Itoa(totalShares)]
	if targetGroupID <= 0 {
		targetGroupID = tierMap["10"]
	}
	plan, err := client.GroupBuyPlan.Create().
		SetTitle("Token拼拼拼 10 份团").
		SetProductKey(GroupBuyProductTokenPinPinPin).
		SetTotalShares(totalShares).
		SetSeatCount(totalShares).
		SetPricePerShare(12).
		SetPricePerSeat(12).
		SetQuotaPerShareLabel("每份 50 USD/月").
		SetQuotaLabel("每份 50 USD/月").
		SetMaxSharesPerUser(10).
		SetTargetGroupID(targetGroupID).
		SetTierGroupIds(tierMap).
		SetValidityDays(30).
		SetTimeoutMinutes(60).
		SetLaunchMode(launchMode).
		SetRefundMode(GroupBuyRefundModeBalanceCredit).
		SetStatus(GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)
	return plan
}

func createGroupBuyTestPlanWithTierRules(t *testing.T, ctx context.Context, client *dbent.Client, rules []domain.GroupBuyTierRule, launchMode string, totalShares int) *dbent.GroupBuyPlan {
	t.Helper()
	targetGroupID := int64(0)
	if rule, ok := resolveTierRuleForShareCount(rules, totalShares); ok {
		targetGroupID = rule.TargetGroupID
	}
	if targetGroupID <= 0 && len(rules) > 0 {
		targetGroupID = rules[len(rules)-1].TargetGroupID
	}
	plan, err := client.GroupBuyPlan.Create().
		SetTitle("Token拼拼拼 10 份团").
		SetProductKey(GroupBuyProductTokenPinPinPin).
		SetTotalShares(totalShares).
		SetSeatCount(totalShares).
		SetPricePerShare(12).
		SetPricePerSeat(12).
		SetQuotaPerShareLabel("每份 50 USD/月").
		SetQuotaLabel("每份 50 USD/月").
		SetMaxSharesPerUser(10).
		SetTargetGroupID(targetGroupID).
		SetTierGroupIds(tierRulesToExactMapping(domainTierRulesToInputs(rules), totalShares)).
		SetTierRules(rules).
		SetValidityDays(30).
		SetTimeoutMinutes(60).
		SetLaunchMode(launchMode).
		SetRefundMode(GroupBuyRefundModeBalanceCredit).
		SetStatus(GroupBuyPlanStatusActive).
		Save(ctx)
	require.NoError(t, err)
	return plan
}

func createGroupBuyProviderRefundFixture(t *testing.T, ctx context.Context, client *dbent.Client, name string) (*dbent.GroupBuyPlan, *dbent.GroupBuySeat, *dbent.PaymentOrder) {
	t.Helper()
	user := createGroupBuyTestUser(t, ctx, client, "provider-"+name+"@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 10)
	plan, err := client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetRefundMode(GroupBuyRefundModeProviderRefund).
		Save(ctx)
	require.NoError(t, err)
	round := createGroupBuyTestRound(t, ctx, client, plan.ID, time.Now().UTC(), 10)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(12).
		SetPayAmount(12).
		SetRechargeCode("gb-provider-" + name).
		SetOutTradeNo("gb_provider_" + name).
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("trade-provider-" + name).
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(user.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetShareCount(1).
		SetPolicySnapshot(domain.GroupBuyPolicySnapshot{RefundMode: GroupBuyRefundModeProviderRefund}).
		Save(ctx)
	require.NoError(t, err)
	seat.Edges.Order = order
	return plan, seat, order
}

func createGroupBuyTestRound(t *testing.T, ctx context.Context, client *dbent.Client, planID int64, started time.Time, totalShares int) *dbent.GroupBuyRound {
	t.Helper()
	round, err := client.GroupBuyRound.Create().
		SetPlanID(planID).
		SetStatus(GroupBuyRoundStatusActive).
		SetTotalShares(totalShares).
		SetPaidShares(totalShares).
		SetReservedShares(0).
		SetTotalSeats(totalShares).
		SetPaidSeats(totalShares).
		SetReservedSeats(0).
		SetDeadlineAt(started.Add(time.Hour)).
		SetStartedAt(started).
		SetClosedAt(started.Add(time.Minute)).
		Save(ctx)
	require.NoError(t, err)
	return round
}

func createGroupBuyTestActiveSeat(t *testing.T, ctx context.Context, client *dbent.Client, planID, roundID, userID int64, shares int, activatedAt, expiresAt time.Time) {
	t.Helper()
	_, err := client.GroupBuySeat.Create().
		SetRoundID(roundID).
		SetPlanID(planID).
		SetUserID(userID).
		SetStatus(GroupBuySeatStatusActive).
		SetShareCount(shares).
		SetPaidAt(activatedAt).
		SetActivatedAt(activatedAt).
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
}

func groupBuyTestTierMap(groupID int64) map[string]int64 {
	out := make(map[string]int64, 10)
	for i := 1; i <= 10; i++ {
		out[strconv.Itoa(i)] = groupID
	}
	return out
}

func groupBuyEntSubToService(sub *dbent.UserSubscription) *UserSubscription {
	if sub == nil {
		return nil
	}
	out := &UserSubscription{
		ID:                 sub.ID,
		UserID:             sub.UserID,
		GroupID:            sub.GroupID,
		StartsAt:           sub.StartsAt,
		ExpiresAt:          sub.ExpiresAt,
		Status:             sub.Status,
		DailyWindowStart:   sub.DailyWindowStart,
		WeeklyWindowStart:  sub.WeeklyWindowStart,
		MonthlyWindowStart: sub.MonthlyWindowStart,
		DailyUsageUSD:      sub.DailyUsageUsd,
		WeeklyUsageUSD:     sub.WeeklyUsageUsd,
		MonthlyUsageUSD:    sub.MonthlyUsageUsd,
		AssignedBy:         sub.AssignedBy,
		AssignedAt:         sub.AssignedAt,
		Notes:              psStringValue(sub.Notes),
		CreatedAt:          sub.CreatedAt,
		UpdatedAt:          sub.UpdatedAt,
	}
	if group := sub.Edges.Group; group != nil {
		out.Group = &Group{
			ID:               group.ID,
			Name:             group.Name,
			Platform:         group.Platform,
			Status:           group.Status,
			SubscriptionType: group.SubscriptionType,
		}
	}
	return out
}
