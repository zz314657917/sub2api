package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type cafeMembershipFixture struct {
	client       *dbent.Client
	service      *CafeRoomActivationService
	public       *CafePublicService
	round        *dbent.GroupBuyRound
	room         *dbent.CafeRoom
	plan         *dbent.GroupBuyPlan
	account      *dbent.Account
	unknown      *dbent.Account
	memberships  []*dbent.CafeRoundMembership
	paymentBatch []*dbent.GroupBuySeat
	now          time.Time
}

func newCafeMembershipFixture(t *testing.T, name string) cafeMembershipFixture {
	t.Helper()
	ctx := context.Background()
	client := newGroupBuyTestClient(t, name)
	now := time.Date(2026, 8, 24, 15, 0, 0, 0, time.UTC)
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	_, err := client.Group.UpdateOneID(groupID).SetAccessMode(CafeRoomGroupAccessMode).Save(ctx)
	require.NoError(t, err)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 3)
	plan, err = client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetFulfillmentMode(CafeRoomFulfillmentMode).
		SetAutoCreateRoomKey(true).
		SetSubscriptionTier("plus").
		SetMaxBuyers(2).
		SetMaxSharesPerUser(3).
		SetFulfillmentTimeoutMinutes(1440).
		SetRoomKeyQuotaUsd(10).
		SetRoomKeyRateLimit5h(2).
		SetRoomKeyRateLimit1d(4).
		SetRoomKeyRateLimit7d(8).
		Save(ctx)
	require.NoError(t, err)
	room, err := client.CafeRoom.Create().SetCode("CAFE-S252").SetName("S252 Plus 包间").SetPlanID(plan.ID).SetStatus(CafeRoomStatusEnabled).SetZoneKey("featured").SetThemeKey("warm_wood").Save(ctx)
	require.NoError(t, err)
	deadline := now.Add(24 * time.Hour)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetCafeRoomID(room.ID).
		SetStatus(GroupBuyRoundStatusAwaitingAccount).
		SetCafeFulfillmentVersion("membership_share").
		SetSubscriptionTier("plus").
		SetMaxBuyers(2).
		SetMaxSharesPerUser(3).
		SetFulfillmentTimeoutMinutes(1440).
		SetValidityDaysSnapshot(30).
		SetTargetGroupIDSnapshot(groupID).
		SetPlatformSnapshot(PlatformOpenAI).
		SetQuotaPerShareSnapshot(10).
		SetRateLimit5hPerShareSnapshot(2).
		SetRateLimit1dPerShareSnapshot(4).
		SetRateLimit7dPerShareSnapshot(8).
		SetFulfillmentDeadlineAt(deadline).
		SetTotalShares(3).
		SetTotalSeats(3).
		SetPaidShares(3).
		SetPaidSeats(3).
		SetDeadlineAt(now.Add(-time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	users := []*User{
		createGroupBuyTestUser(t, ctx, client, "s252-first@example.com"),
		createGroupBuyTestUser(t, ctx, client, "s252-second@example.com"),
	}
	memberships := make([]*dbent.CafeRoundMembership, 0, 2)
	for index, shares := range []int{2, 1} {
		membership, createErr := client.CafeRoundMembership.Create().SetRoundID(round.ID).SetUserID(users[index].ID).SetStatus(GroupBuySeatStatusPaid).SetPaidShares(shares).Save(ctx)
		require.NoError(t, createErr)
		memberships = append(memberships, membership)
	}
	batches := make([]*dbent.GroupBuySeat, 0, 3)
	for index, membership := range []*dbent.CafeRoundMembership{memberships[0], memberships[0], memberships[1]} {
		batch, createErr := client.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(membership.UserID).SetMembershipID(membership.ID).SetStatus(GroupBuySeatStatusPaid).SetShareCount(1).SetPaidAt(now).Save(ctx)
		require.NoError(t, createErr, index)
		batches = append(batches, batch)
	}
	account, err := client.Account.Create().SetName("Cafe Plus Account").SetPlatform(PlatformOpenAI).SetType("oauth").SetStatus(StatusActive).SetCredentials(map[string]any{"plan_type": "plus", "email": "owner@example.com", "access_token": "must-not-leak"}).AddGroupIDs(groupID).Save(ctx)
	require.NoError(t, err)
	unknown, err := client.Account.Create().SetName("Unknown Account").SetPlatform(PlatformOpenAI).SetType("oauth").SetStatus(StatusActive).SetCredentials(map[string]any{"plan_type": "team"}).AddGroupIDs(groupID).Save(ctx)
	require.NoError(t, err)
	apiKeyRepo := &cafeActivationAPIKeyRepo{client: client}
	apiKeySvc := NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, &cafeManagedKeyCacheStub{}, nil)
	activation := NewCafeRoomActivationService(client, apiKeySvc, apiKeyRepo)
	activation.now = func() time.Time { return now }
	sequence := 0
	activation.generateKey = func() (string, error) {
		sequence++
		return fmt.Sprintf("sk-s252-%032d", sequence), nil
	}
	public := NewCafePublicService(client, cafePublicSettingsStub{enabled: true})
	public.now = func() time.Time { return now }
	return cafeMembershipFixture{client: client, service: activation, public: public, round: round, room: room, plan: plan, account: account, unknown: unknown, memberships: memberships, paymentBatch: batches, now: now}
}

func TestCafeMembershipFulfillmentCreatesOneScaledKeyPerBuyerAtomically(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_activation")

	options, page, err := fixture.service.ListRoundAccountOptions(ctx, fixture.round.ID, CafeRoundAccountOptionParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, options, 1)
	require.Equal(t, fixture.account.ID, options[0].ID)
	require.Equal(t, "plus", options[0].PlanType)
	require.Equal(t, "o***r@example.com", options[0].EmailMasked)

	result, err := fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, result.Status)
	require.Equal(t, fixture.room.ID, result.RoomID)
	require.Equal(t, 2, result.JoinedBuyers)

	round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
	require.Equal(t, fixture.account.ID, *round.AssignedAccountID)
	require.Equal(t, fixture.now, *round.ActivatedAt)
	require.Equal(t, fixture.now.AddDate(0, 0, 30), *round.EntitlementExpiresAt)

	keys, err := fixture.client.APIKey.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 2)
	keyByMembership := make(map[int64]*dbent.APIKey, 2)
	for _, key := range keys {
		require.Equal(t, APIKeyManagedSourceCafeRoomMembership, key.ManagedSourceType)
		require.NotNil(t, key.ManagedSourceID)
		keyByMembership[*key.ManagedSourceID] = key
	}
	require.Equal(t, 20.0, keyByMembership[fixture.memberships[0].ID].Quota)
	require.Equal(t, 4.0, keyByMembership[fixture.memberships[0].ID].RateLimit5h)
	require.Equal(t, 16.0, keyByMembership[fixture.memberships[0].ID].RateLimit7d)
	require.Equal(t, 10.0, keyByMembership[fixture.memberships[1].ID].Quota)

	bindings, err := fixture.client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.StatusEQ("active")).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 2)
	for _, binding := range bindings {
		require.Nil(t, binding.SeatID)
		require.NotNil(t, binding.MembershipID)
	}
	batches, err := fixture.client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, batches, 3)
	require.Equal(t, batches[0].BoundAPIKeyID, batches[1].BoundAPIKeyID)

	_, err = fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	keyCount, err := fixture.client.APIKey.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, keyCount)
}

func TestCafeMembershipFulfillmentRejectsUnknownTierAndRollsBackKeyFailure(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_activation_rollback")

	_, err := fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.unknown.ID)
	require.ErrorIs(t, err, ErrCafeAccountTierMismatch)
	require.Equal(t, "CAFE_ACCOUNT_TIER_MISMATCH", infraerrors.Reason(err))

	fixture.service.apiKeyRepo = &cafeActivationFailAlwaysAPIKeyRepo{cafeActivationAPIKeyRepo: fixture.service.apiKeyRepo.(*cafeActivationAPIKeyRepo)}
	_, err = fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.ErrorIs(t, err, ErrCafeActivationFailed)
	round, loadErr := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, loadErr)
	require.Equal(t, GroupBuyRoundStatusAwaitingAccount, round.Status)
	require.Nil(t, round.AssignedAccountID)
	keyCount, loadErr := fixture.client.APIKey.Query().Count(ctx)
	require.NoError(t, loadErr)
	require.Zero(t, keyCount)
	bindingCount, loadErr := fixture.client.APIKeyAccountBinding.Query().Count(ctx)
	require.NoError(t, loadErr)
	require.Zero(t, bindingCount)
}

func TestCafeMembershipPublicProjectionWaitsForActivation(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_public")
	userID := fixture.memberships[0].UserID

	rooms, _, err := fixture.public.List(ctx, userID, CafePublicListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, rooms, 1)
	require.Equal(t, "plus", rooms[0].Plan.SubscriptionTier)
	require.Equal(t, 3, rooms[0].Round.PaidShares)
	require.Equal(t, 2, rooms[0].Round.JoinedBuyers)
	require.Equal(t, 2, rooms[0].MyPaidShares)
	require.Len(t, rooms[0].MemberAvatars, 2)
	require.Equal(t, "awaiting_account", rooms[0].PurchaseState)

	items, _, err := fixture.public.MyRooms(ctx, userID, CafeMyRoomsListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, 2, items[0].PaidShares)
	require.Nil(t, items[0].Account)
	require.Nil(t, items[0].ManagedAPIKey)

	_, err = fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	items, _, err = fixture.public.MyRooms(ctx, userID, CafeMyRoomsListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].Account)
	require.Equal(t, "o***r@example.com", items[0].Account.EmailMasked)
	require.NotNil(t, items[0].ManagedAPIKey)
	require.Equal(t, 20.0, items[0].ManagedAPIKey.Quota)
}

func TestCafeFulfillmentTimeoutStaysRefundingUntilEveryBatchSucceeds(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_refund")
	lifecycle := &CafeRoomLifecycleService{entClient: fixture.client, groupBuy: newGroupBuyTestService(fixture.client, nil, nil), now: func() time.Time { return fixture.now.Add(24 * time.Hour) }}

	changed, err := lifecycle.expireAwaitingAccountCafeRound(ctx, fixture.round.ID, fixture.now.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, changed)
	round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusRefunding, round.Status)

	for index, seat := range fixture.paymentBatch {
		status := GroupBuyRefundStatusSucceeded
		if index == 0 {
			status = GroupBuyRefundStatusPendingProvider
		}
		_, err = fixture.client.GroupBuySeat.UpdateOneID(seat.ID).SetStatus(GroupBuySeatStatusRefunded).Save(ctx)
		require.NoError(t, err)
		_, err = fixture.client.GroupBuyRefund.Create().SetSeatID(seat.ID).SetUserID(seat.UserID).SetMode(GroupBuyRefundModeProviderRefund).SetStatus(status).SetAmount(12).SetIdempotencyKey(fmt.Sprintf("s252-refund-%d", seat.ID)).Save(ctx)
		require.NoError(t, err)
	}
	changed, err = lifecycle.finalizeCafeRefundedRound(ctx, fixture.round.ID, fixture.now.Add(24*time.Hour))
	require.NoError(t, err)
	require.False(t, changed)
	round, err = fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusRefunding, round.Status)

	pending, err := fixture.client.GroupBuyRefund.Query().Where(groupbuyrefund.StatusEQ(GroupBuyRefundStatusPendingProvider)).Only(ctx)
	require.NoError(t, err)
	_, err = fixture.client.GroupBuyRefund.UpdateOneID(pending.ID).SetStatus(GroupBuyRefundStatusSucceeded).SetProcessedAt(fixture.now.Add(24 * time.Hour)).Save(ctx)
	require.NoError(t, err)
	changed, err = lifecycle.finalizeCafeRefundedRound(ctx, fixture.round.ID, fixture.now.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, changed)
	round, err = fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
require.Equal(t, GroupBuyRoundStatusRefunded, round.Status)
}
