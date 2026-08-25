package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/stretchr/testify/require"
)

func TestCafeRoomMigrationReplacesBindingsAndInvalidatesCache(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_migration_success")
	now := time.Date(2026, 8, 3, 19, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 2)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	newAccount := createCafeMigrationAccount(t, ctx, client, fixture.groupID, "cafe-migration-new-account")
	invalidator := &cafeExpiryCacheInvalidatorStub{}
	service := NewCafeRoomMigrationService(client, invalidator)
	service.now = func() time.Time { return now.Add(time.Minute) }

	result, err := service.MigrateActiveRound(ctx, fixture.round.ID, newAccount.ID, "认证故障")
	require.NoError(t, err)
	require.Equal(t, fixture.accountID, result.OldAccountID)
	require.Equal(t, newAccount.ID, result.NewAccountID)
	require.Equal(t, 2, result.MigratedBindings)
	require.False(t, result.Noop)
	require.Len(t, invalidator.keys, 2)

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, newAccount.ID, *round.AssignedAccountID)
	room, err := client.CafeRoom.Get(ctx, fixture.room.ID)
	require.NoError(t, err)
	require.Equal(t, newAccount.ID, *room.AccountID)

	bindings, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 4)
	activeBindings, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).All(ctx)
	require.NoError(t, err)
	require.Len(t, activeBindings, 2)
	for _, binding := range bindings {
		if binding.Status == cafeBindingStatusMigrated {
			require.NotNil(t, binding.ReplacedByBindingID)
			require.NotNil(t, binding.MigratedAt)
		} else {
			require.Equal(t, newAccount.ID, binding.AccountID)
			require.True(t, binding.StrictMode)
		}
	}
	keys, err := client.APIKey.Query().Where(apikey.ManagedSourceTypeEQ(APIKeyManagedSourceCafeRoomSeat)).All(ctx)
	require.NoError(t, err)
	for _, key := range keys {
		require.Equal(t, StatusAPIKeyActive, key.Status)
	}

	second, err := service.MigrateActiveRound(ctx, fixture.round.ID, newAccount.ID, "重复请求")
	require.NoError(t, err)
	require.True(t, second.Noop)
	require.Zero(t, second.MigratedBindings)
	require.Len(t, invalidator.keys, 2)
}

func TestCafeRoomMigrationReplacesMembershipBindingsWithoutRebindingRoom(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_migration")
	_, err := fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	newAccount, err := fixture.client.Account.Create().
		SetName("Cafe Plus Replacement").
		SetPlatform(PlatformOpenAI).
		SetType("oauth").
		SetStatus(StatusActive).
		SetCredentials(map[string]any{"plan_type": "plus", "email": "replacement@example.com"}).
		AddGroupIDs(fixture.plan.TargetGroupID).
		Save(ctx)
	require.NoError(t, err)
	invalidator := &cafeExpiryCacheInvalidatorStub{}
	svc := NewCafeRoomMigrationService(fixture.client, invalidator)
	svc.now = func() time.Time { return fixture.now.Add(time.Minute) }

	result, err := svc.MigrateActiveRound(ctx, fixture.round.ID, newAccount.ID, "membership account rotation")
	require.NoError(t, err)
	require.Equal(t, 2, result.MigratedBindings)
	require.Len(t, invalidator.keys, 2)
	round, err := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, newAccount.ID, *round.AssignedAccountID)
	room, err := fixture.client.CafeRoom.Get(ctx, fixture.room.ID)
	require.NoError(t, err)
	require.Nil(t, room.AccountID, "membership rounds keep the account assignment on the round")
	activeBindings, err := fixture.client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).All(ctx)
	require.NoError(t, err)
	require.Len(t, activeBindings, 2)
	for _, binding := range activeBindings {
		require.Equal(t, newAccount.ID, binding.AccountID)
		require.Nil(t, binding.SeatID)
		require.NotNil(t, binding.MembershipID)
	}
	report, err := svc.CheckConsistency(ctx)
	require.NoError(t, err)
	require.Empty(t, report.Issues)
}

func TestCafeRoomMigrationRejectsMembershipAccountWithWrongTier(t *testing.T) {
	ctx := context.Background()
	fixture := newCafeMembershipFixture(t, "cafe_membership_migration_tier")
	_, err := fixture.service.AssignAccountAndActivateRound(ctx, fixture.round.ID, fixture.account.ID)
	require.NoError(t, err)
	wrongTier, err := fixture.client.Account.Create().SetName("Cafe Pro Replacement").SetPlatform(PlatformOpenAI).SetType("oauth").SetStatus(StatusActive).SetCredentials(map[string]any{"plan_type": "pro"}).AddGroupIDs(fixture.plan.TargetGroupID).Save(ctx)
	require.NoError(t, err)
	svc := NewCafeRoomMigrationService(fixture.client, &cafeExpiryCacheInvalidatorStub{})
	svc.now = func() time.Time { return fixture.now.Add(time.Minute) }

	_, err = svc.MigrateActiveRound(ctx, fixture.round.ID, wrongTier.ID, "wrong tier")
	require.ErrorIs(t, err, ErrCafeMigrationAccountInvalid)
	round, loadErr := fixture.client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, loadErr)
	require.Equal(t, fixture.account.ID, *round.AssignedAccountID)
}

func TestCafeRoomMigrationRejectsIncompatibleTargetWithoutMutation(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_migration_invalid_target")
	now := time.Date(2026, 8, 3, 19, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	invalidAccount := createCafeMigrationAccount(t, ctx, client, 0, "cafe-migration-invalid-account")
	service := NewCafeRoomMigrationService(client, &cafeExpiryCacheInvalidatorStub{})
	service.now = func() time.Time { return now.Add(time.Minute) }

	_, err := service.MigrateActiveRound(ctx, fixture.round.ID, invalidAccount.ID, "不兼容账号")
	require.ErrorIs(t, err, ErrCafeMigrationAccountInvalid)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.accountID, *round.AssignedAccountID)
	room, err := client.CafeRoom.Get(ctx, fixture.room.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.accountID, *room.AccountID)
	activeCount, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, activeCount)
}

func TestCafeRoomMigrationRollsBackWhenActiveFactsAreInconsistent(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_migration_rollback")
	now := time.Date(2026, 8, 3, 19, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 2)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	newAccount := createCafeMigrationAccount(t, ctx, client, fixture.groupID, "cafe-migration-rollback-account")
	inconsistentAccount := createCafeMigrationAccount(t, ctx, client, fixture.groupID, "cafe-migration-inconsistent-account")
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID), groupbuyseat.SeatNoEQ(2)).Only(ctx)
	require.NoError(t, err)
	binding, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.SeatIDEQ(seat.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKeyAccountBinding.UpdateOneID(binding.ID).SetAccountID(inconsistentAccount.ID).Exec(ctx))

	service := NewCafeRoomMigrationService(client, &cafeExpiryCacheInvalidatorStub{})
	service.now = func() time.Time { return now.Add(time.Minute) }
	_, err = service.MigrateActiveRound(ctx, fixture.round.ID, newAccount.ID, "一致性回滚")
	require.ErrorIs(t, err, ErrCafeMigrationInconsistent)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.accountID, *round.AssignedAccountID)
	activeCount, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.RoundIDEQ(fixture.round.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, activeCount)
}

func TestCafeRoomConsistencyAndDryRunAreReadOnly(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_migration_checker")
	now := time.Date(2026, 8, 3, 19, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	binding, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.SeatIDEQ(seat.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.APIKeyAccountBinding.DeleteOneID(binding.ID).Exec(ctx))

	service := NewCafeRoomMigrationService(client, &cafeExpiryCacheInvalidatorStub{})
	service.now = func() time.Time { return now }
	report, err := service.CheckConsistency(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, report.Issues)
	require.Contains(t, consistencyIssueCodes(report.Issues), cafeMigrationIssueSeatBindingMismatch)
	plan, err := service.PlanDryRunRepair(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, plan.Suggestions)
	require.Contains(t, repairActions(plan.Suggestions), cafeMigrationActionManualInvestigation)
	require.NotContains(t, repairActions(plan.Suggestions), cafeMigrationActionRetryActivation)
	activeRound, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, activeRound.Status)
	activeBindingCount, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).Count(ctx)
	require.NoError(t, err)
	require.Zero(t, activeBindingCount)
}

func createCafeMigrationAccount(t *testing.T, ctx context.Context, client *dbent.Client, groupID int64, name string) *dbent.Account {
	t.Helper()
	builder := client.Account.Create().SetName(name).SetPlatform(PlatformOpenAI).SetType("api_key").SetStatus(StatusActive)
	if groupID > 0 {
		builder.AddGroupIDs(groupID)
	}
	account, err := builder.Save(ctx)
	require.NoError(t, err)
	return account
}

func consistencyIssueCodes(issues []CafeConsistencyIssue) []string {
	codes := make([]string, 0, len(issues))
	for _, issue := range issues {
		codes = append(codes, issue.Code)
	}
	return codes
}

func repairActions(suggestions []CafeRepairSuggestion) []string {
	actions := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		actions = append(actions, suggestion.Action)
	}
	return actions
}
