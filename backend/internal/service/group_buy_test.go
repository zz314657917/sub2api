package service

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestGroupBuyAdminCreatePlanRequiresCompleteTierMapping(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "groupbuy_plan_tier_mapping")
	groupRepo := newGroupBuyGroupRepoStub()
	svc := newGroupBuyTestService(client, groupRepo, nil)

	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	groupRepo.groups[groupID] = &Group{ID: groupID, Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription}

	incomplete := map[string]int64{}
	for i := 1; i <= 9; i++ {
		incomplete[strconv.Itoa(i)] = groupID
	}

	_, err := svc.AdminCreatePlan(ctx, GroupBuyPlanInput{
		Title:              "Token拼拼拼 10 份团",
		TotalShares:        10,
		PricePerShare:      12,
		QuotaPerShareLabel: "每份 50 USD/月",
		MaxSharesPerUser:   10,
		TierGroupIDs:       incomplete,
		ValidityDays:       30,
		TimeoutMinutes:     60,
		Status:             GroupBuyPlanStatusActive,
	})
	require.ErrorIs(t, err, ErrGroupBuyTierMappingInvalid)

	complete := groupBuyTestTierMap(groupID)
	plan, err := svc.AdminCreatePlan(ctx, GroupBuyPlanInput{
		Title:              "Token拼拼拼 10 份团",
		TotalShares:        10,
		PricePerShare:      12,
		QuotaPerShareLabel: "每份 50 USD/月",
		MaxSharesPerUser:   10,
		TierGroupIDs:       complete,
		ValidityDays:       30,
		TimeoutMinutes:     60,
		LaunchMode:         GroupBuyLaunchModeManual,
		Status:             GroupBuyPlanStatusActive,
	})
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, groupID, plan.TargetGroupID)
	require.Equal(t, 10, len(plan.TierGroupIDs))
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

type groupBuyUserRepoStub struct {
	users map[int64]*User
}

func (s *groupBuyUserRepoStub) Create(context.Context, *User) error { panic("unexpected Create call") }
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
func (s *groupBuyUserRepoStub) Update(context.Context, *User) error { panic("unexpected Update call") }
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
func (s *groupBuyUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}
func (s *groupBuyUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
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
func (s *groupBuyAPIKeyRepoStub) Update(_ context.Context, key *APIKey) error {
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
