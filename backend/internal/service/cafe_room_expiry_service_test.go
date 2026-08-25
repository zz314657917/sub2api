package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/stretchr/testify/require"
)

type cafeExpiryCacheInvalidatorStub struct {
	keys []string
}

func (s *cafeExpiryCacheInvalidatorStub) InvalidateAuthCacheByKey(_ context.Context, key string) {
	s.keys = append(s.keys, key)
}

func TestCafeRoomExpiryReclaimsManagedFactsAndCompletesRound(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_expiry_reclaim")
	activatedAt := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, activatedAt, 2)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))

	expiresAt := activatedAt.AddDate(0, 0, fixture.plan.ValidityDays)
	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomExpiryService(client, invalidator)
	svc.now = func() time.Time { return expiresAt }

	count, err := svc.ExpireCafeRounds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusCompleted, round.Status)
	require.NotNil(t, round.CompletedAt)
	require.Equal(t, expiresAt, *round.CompletedAt)

	seats, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, seats, 2)
	for _, seat := range seats {
		require.Equal(t, GroupBuySeatStatusExpired, seat.Status)
	}
	bindings, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 2)
	for _, binding := range bindings {
		require.Equal(t, apiKeyAccountBindingStatusExpired, binding.Status)
	}
	keys, err := client.APIKey.Query().Where(apikey.ManagedSourceTypeEQ(APIKeyManagedSourceCafeRoomSeat)).All(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 2)
	for _, key := range keys {
		require.Equal(t, StatusAPIKeyExpired, key.Status)
		require.NotNil(t, key.ExpiresAt)
		require.Equal(t, expiresAt, *key.ExpiresAt)
	}
	require.ElementsMatch(t, []string{keys[0].Key, keys[1].Key}, invalidator.keys)

	events, err := client.GroupBuyEvent.Query().Where(groupbuyevent.RoundIDEQ(fixture.round.ID), groupbuyevent.EventTypeEQ(groupBuyEventRoundCompleted)).All(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	count, err = svc.ExpireCafeRounds(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	require.Len(t, invalidator.keys, 2)
}

func TestCafeRoomExpiryReclaimsMembershipOnceAndExpiresAllPaymentBatches(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_expiry")
	_, err := fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	expiresAt := fixture.now.AddDate(0, 0, 30)
	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomExpiryService(fixture.client, invalidator)
	svc.now = func() time.Time { return expiresAt }

	count, err := svc.ExpireCafeRounds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Len(t, invalidator.keys, 2)
	round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusCompleted, round.Status)
	memberships, err := fixture.client.CafeRoundMembership.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, memberships, 2)
	for _, membership := range memberships {
		require.Equal(t, GroupBuySeatStatusExpired, membership.Status)
	}
	batches, err := fixture.client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, batches, 3)
	for _, batch := range batches {
		require.Equal(t, GroupBuySeatStatusExpired, batch.Status)
	}
	bindings, err := fixture.client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 2)
	for _, binding := range bindings {
		require.Equal(t, apiKeyAccountBindingStatusExpired, binding.Status)
		require.Nil(t, binding.SeatID)
		require.NotNil(t, binding.MembershipID)
	}
	keys, err := fixture.client.APIKey.Query().Where(apikey.ManagedSourceTypeEQ(APIKeyManagedSourceCafeRoomMembership)).All(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 2)
	for _, key := range keys {
		require.Equal(t, StatusAPIKeyExpired, key.Status)
	}

	count, err = svc.ExpireCafeRounds(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	require.Len(t, invalidator.keys, 2)
}

func TestCafeRoomExpiryRollsBackInconsistentRoundWithoutCacheInvalidation(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_expiry_inconsistent")
	activatedAt := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, activatedAt, 1)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))

	binding, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKeyAccountBinding.DeleteOneID(binding.ID).Exec(ctx))

	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomExpiryService(client, invalidator)
	svc.now = func() time.Time { return activatedAt.AddDate(0, 0, fixture.plan.ValidityDays) }

	count, err := svc.ExpireCafeRounds(ctx)
	require.ErrorIs(t, err, ErrCafeExpiryInconsistent)
	require.Zero(t, count)
	require.Empty(t, invalidator.keys)

	round, err := client.GroupBuyRound.Query().Where(groupbuyround.IDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusActive, seat.Status)
	key, err := client.APIKey.Query().Where(apikey.IDEQ(*seat.BoundAPIKeyID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, StatusAPIKeyActive, key.Status)
}

func TestCafeRoomExpiryRejectsBindingAndKeyFromOtherGroup(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_expiry_wrong_group")
	activatedAt := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, activatedAt, 1)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))

	otherGroupID := createGroupBuyTestGroup(t, ctx, client, 9, 100)
	binding, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKeyAccountBinding.UpdateOneID(binding.ID).SetGroupID(otherGroupID).Exec(ctx))
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKey.UpdateOneID(*seat.BoundAPIKeyID).SetGroupID(otherGroupID).Exec(ctx))

	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomExpiryService(client, invalidator)
	svc.now = func() time.Time { return activatedAt.AddDate(0, 0, fixture.plan.ValidityDays) }

	count, err := svc.ExpireCafeRounds(ctx)
	require.ErrorIs(t, err, ErrCafeExpiryInconsistent)
	require.Zero(t, count)
	require.Empty(t, invalidator.keys)

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
}

func TestCafeRoomExpiryContinuesAfterOneInconsistentRoom(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_expiry_isolated")
	activatedAt := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	broken := newCafeActivationFixture(t, ctx, client, activatedAt, 1)
	require.NoError(t, broken.service.ActivateRound(ctx, broken.round.ID))

	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(broken.round.ID)).Only(ctx)
	require.NoError(t, err)
	binding, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKeyAccountBinding.DeleteOneID(binding.ID).Exec(ctx))

	healthyRoundID := createSecondExpiredCafeRound(t, ctx, client, activatedAt, seat.UserID)
	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomExpiryService(client, invalidator)
	svc.now = func() time.Time { return activatedAt.AddDate(0, 0, broken.plan.ValidityDays) }

	count, err := svc.ExpireCafeRounds(ctx)
	require.ErrorIs(t, err, ErrCafeExpiryInconsistent)
	require.Equal(t, 1, count)
	require.Len(t, invalidator.keys, 1)

	brokenRound, err := client.GroupBuyRound.Get(ctx, broken.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, brokenRound.Status)
	healthyRound, err := client.GroupBuyRound.Get(ctx, healthyRoundID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusCompleted, healthyRound.Status)
}

func createSecondExpiredCafeRound(t *testing.T, ctx context.Context, client *dbent.Client, activatedAt time.Time, userID int64) int64 {
	t.Helper()
	groupID := createGroupBuyTestGroup(t, ctx, client, 2, 100)
	_, err := client.Group.UpdateOneID(groupID).SetAccessMode(CafeRoomGroupAccessMode).Save(ctx)
	require.NoError(t, err)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, 1)
	plan, err = client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetFulfillmentMode(CafeRoomFulfillmentMode).
		SetAutoCreateRoomKey(true).
		SetRoomKeyQuotaUsd(25).
		Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("cafe-expiry-second-account").
		SetPlatform(PlatformOpenAI).
		SetType("api_key").
		SetStatus(StatusActive).
		AddGroupIDs(groupID).
		Save(ctx)
	require.NoError(t, err)
	room, round := createCafeRoomOrderRoom(t, ctx, client, plan.ID, account.ID, activatedAt, 1, 992)
	_, err = client.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(plan.ID).
		SetUserID(userID).
		SetSeatNo(1).
		SetStatus(GroupBuySeatStatusPaid).
		SetShareCount(1).
		SetPaidAt(activatedAt).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(round.ID).
		SetPaidSeats(1).
		SetPaidShares(1).
		Save(ctx)
	require.NoError(t, err)

	activation := NewCafeRoomActivationService(client, &APIKeyService{}, &cafeActivationAPIKeyRepo{client: client})
	activation.now = func() time.Time { return activatedAt }
	activation.generateKey = func() (string, error) { return fmt.Sprintf("sk-cafe-expiry-%d", room.ID), nil }
	require.NoError(t, activation.ActivateRound(ctx, round.ID))
	return round.ID
}
