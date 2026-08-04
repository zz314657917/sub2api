package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type cafeActivationAPIKeyRepo struct {
	APIKeyRepository
	client *dbent.Client
}

type cafeActivationFailOnceAPIKeyRepo struct {
	*cafeActivationAPIKeyRepo
	failNextCreate bool
}

func (r *cafeActivationFailOnceAPIKeyRepo) Create(ctx context.Context, key *APIKey) error {
	if r.failNextCreate {
		r.failNextCreate = false
		return errors.New("injected managed key persistence failure")
	}
	return r.cafeActivationAPIKeyRepo.Create(ctx, key)
}

func (r *cafeActivationAPIKeyRepo) Create(ctx context.Context, key *APIKey) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	created, err := client.APIKey.Create().
		SetUserID(key.UserID).
		SetKey(key.Key).
		SetName(key.Name).
		SetNillableGroupID(key.GroupID).
		SetAccountPoolStrategy(key.AccountPoolStrategy).
		SetStatus(key.Status).
		SetQuota(key.Quota).
		SetNillableExpiresAt(key.ExpiresAt).
		SetRateLimit5h(key.RateLimit5h).
		SetRateLimit1d(key.RateLimit1d).
		SetRateLimit7d(key.RateLimit7d).
		SetMultiGroupRoutes(key.MultiGroupRoutes).
		SetManagedSourceType(key.ManagedSourceType).
		SetNillableManagedSourceID(key.ManagedSourceID).
		Save(ctx)
	if err != nil {
		return err
	}
	key.ID = created.ID
	return nil
}

type cafeManagedKeyRepoStub struct {
	APIKeyRepository
	key              *APIKey
	updated          *APIKey
	managedUpdated   *APIKey
	managedStatus    string
	managedUpdateErr error
	managedCalls     int
}

func (r *cafeManagedKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *r.key
	return &clone, nil
}

func (r *cafeManagedKeyRepoStub) Update(_ context.Context, key *APIKey) error {
	clone := *key
	r.updated = &clone
	return nil
}

func (r *cafeManagedKeyRepoStub) UpdateCafeManagedAPIKey(_ context.Context, key *APIKey, desiredStatus string, _ time.Time) error {
	r.managedCalls++
	r.managedStatus = desiredStatus
	if r.managedUpdateErr != nil {
		return r.managedUpdateErr
	}
	clone := *key
	r.managedUpdated = &clone
	r.key = &clone
	return nil
}

func (r *cafeManagedKeyRepoStub) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return []APIKey{}, &pagination.PaginationResult{}, nil
}

type cafeManagedKeyCacheStub struct {
	APIKeyCache
	deleted   []string
	published []string
}

func (s *cafeManagedKeyCacheStub) DeleteAuthCache(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}

func (s *cafeManagedKeyCacheStub) PublishAuthCacheInvalidation(_ context.Context, key string) error {
	s.published = append(s.published, key)
	return nil
}

func TestCafeRoomActivationCreatesActiveManagedKeysAndBindingsIdempotently(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_activation_idempotent")
	now := time.Date(2026, 8, 3, 14, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 2)

	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
	require.NotNil(t, round.ActivationToken)
	require.Equal(t, now, *round.ActivatedAt)
	require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), *round.EntitlementExpiresAt)

	keys, err := client.APIKey.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 2)
	expectedCacheKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		require.Equal(t, StatusAPIKeyActive, key.Status)
		require.Equal(t, APIKeyManagedSourceCafeRoomSeat, key.ManagedSourceType)
		require.NotNil(t, key.ManagedSourceID)
		require.Equal(t, fixture.groupID, *key.GroupID)
		require.Equal(t, fixture.plan.RoomKeyQuotaUsd, key.Quota)
		require.Equal(t, fixture.plan.RoomKeyRateLimit5h, key.RateLimit5h)
		require.Equal(t, fixture.plan.RoomKeyRateLimit1d, key.RateLimit1d)
		require.Equal(t, fixture.plan.RoomKeyRateLimit7d, key.RateLimit7d)
		require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), *key.ExpiresAt)
		expectedCacheKeys = append(expectedCacheKeys, fixture.apiKeyService.authCacheKey(key.Key))
	}
	require.ElementsMatch(t, expectedCacheKeys, fixture.cache.deleted)
	require.ElementsMatch(t, expectedCacheKeys, fixture.cache.published)

	bindings, err := client.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 2)
	for _, binding := range bindings {
		require.True(t, binding.StrictMode)
		require.Equal(t, fixture.accountID, binding.AccountID)
		require.Equal(t, fixture.room.ID, binding.CafeRoomID)
		require.Equal(t, fixture.round.ID, binding.RoundID)
		require.Equal(t, now, binding.StartsAt)
		require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), binding.ExpiresAt)
	}

	seats, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, seats, 2)
	for _, seat := range seats {
		require.Equal(t, GroupBuySeatStatusActive, seat.Status)
		require.NotNil(t, seat.BoundAPIKeyID)
		require.Equal(t, now, *seat.ActivatedAt)
		require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), *seat.ExpiresAt)
	}
	subscriptions, err := client.UserSubscription.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, subscriptions)
}

func TestCafeRoomActivationKeepsUnfilledRoundOpen(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_activation_pending")
	fixture := newCafeActivationFixture(t, ctx, client, time.Date(2026, 8, 3, 14, 0, 0, 0, time.UTC), 2)

	_, err := client.GroupBuySeat.Delete().Where(groupbuyseat.RoundIDEQ(fixture.round.ID), groupbuyseat.SeatNoEQ(2)).Exec(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(fixture.round.ID).SetPaidSeats(1).SetPaidShares(1).Save(ctx)
	require.NoError(t, err)

	err = fixture.service.ActivateRound(ctx, fixture.round.ID)
	require.ErrorIs(t, err, ErrCafeActivationPending)
	round, err := client.GroupBuyRound.Query().Where(groupbuyround.IDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusOpen, round.Status)
	keyCount, err := client.APIKey.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, keyCount)
}

func TestCafeRoomActivationRetriesAfterManagedKeyFailureWithOriginalClaim(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_activation_retry")
	now := time.Date(2026, 8, 3, 14, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	failingRepo := &cafeActivationFailOnceAPIKeyRepo{
		cafeActivationAPIKeyRepo: fixture.service.apiKeyRepo.(*cafeActivationAPIKeyRepo),
		failNextCreate:           true,
	}
	fixture.service.apiKeyRepo = failingRepo

	err := fixture.service.ActivateRound(ctx, fixture.round.ID)
	require.ErrorIs(t, err, ErrCafeActivationFailed)
	claimed, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActivating, claimed.Status)
	require.NotNil(t, claimed.ActivationToken)
	require.Equal(t, now, *claimed.ActivatedAt)
	require.Equal(t, now.AddDate(0, 0, fixture.plan.ValidityDays), *claimed.EntitlementExpiresAt)

	require.NoError(t, fixture.service.ActivateRound(ctx, fixture.round.ID))
	activated, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, activated.Status)
	require.Equal(t, *claimed.ActivationToken, *activated.ActivationToken)
	require.Equal(t, *claimed.ActivatedAt, *activated.ActivatedAt)
	require.Equal(t, *claimed.EntitlementExpiresAt, *activated.EntitlementExpiresAt)
	keyCount, err := client.APIKey.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, keyCount)
	bindingCount, err := client.APIKeyAccountBinding.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, bindingCount)
}

func TestAPIKeyServiceProtectsCafeManagedKeyLifecycleFields(t *testing.T) {
	managedSourceID := int64(99)
	repo := &cafeManagedKeyRepoStub{key: &APIKey{
		ID:                1,
		UserID:            7,
		Key:               "sk-cafe-managed-service-test",
		Name:              "Original managed key",
		Status:            StatusAPIKeyActive,
		ManagedSourceType: APIKeyManagedSourceCafeRoomSeat,
		ManagedSourceID:   &managedSourceID,
	}}
	svc := &APIKeyService{apiKeyRepo: repo, cafeManagedKeyUpdater: repo}
	groupID := int64(3)
	quota := 1.0
	expiresAt := time.Now().Add(time.Hour)
	reset := true
	poolStrategy := AccountPoolStrategyPrivateOnly
	for name, request := range map[string]UpdateAPIKeyRequest{
		"group":            {GroupID: &groupID},
		"routes":           {MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{}},
		"pool":             {AccountPoolStrategy: &poolStrategy},
		"quota":            {Quota: &quota},
		"expiry":           {ExpiresAt: &expiresAt},
		"clear_expiry":     {ClearExpiration: true},
		"reset_quota":      {ResetQuota: &reset},
		"rate_5h":          {RateLimit5h: &quota},
		"rate_1d":          {RateLimit1d: &quota},
		"rate_7d":          {RateLimit7d: &quota},
		"reset_rate_usage": {ResetRateLimitUsage: &reset},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), 1, 7, request)
			require.ErrorIs(t, err, ErrCafeManagedKeyProtected)
		})
	}
	require.Zero(t, repo.managedCalls)

	_, err := svc.Update(context.Background(), 1, 8, UpdateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrInsufficientPerms)

	invalidStatus := StatusAPIKeyDisabled
	_, err = svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{Status: &invalidStatus})
	require.ErrorIs(t, err, ErrCafeManagedKeyStatusInvalid)
	require.Zero(t, repo.managedCalls)

	name := "Allowed managed key name"
	updated, err := svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{Name: &name, IPWhitelist: &[]string{"127.0.0.1"}})
	require.NoError(t, err)
	require.Equal(t, name, updated.Name)
	require.NotNil(t, repo.updated)
	require.Equal(t, []string{"127.0.0.1"}, repo.updated.IPWhitelist)

	inactiveStatus := "inactive"
	inactiveName := "Paused managed key"
	updated, err = svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{
		Name:        &inactiveName,
		Status:      &inactiveStatus,
		IPBlacklist: &[]string{"10.0.0.0/8"},
	})
	require.NoError(t, err)
	require.Equal(t, inactiveStatus, updated.Status)
	require.Equal(t, inactiveName, updated.Name)
	require.Equal(t, inactiveStatus, repo.managedStatus)
	require.Equal(t, []string{"10.0.0.0/8"}, repo.managedUpdated.IPBlacklist)
	require.Equal(t, 1, repo.managedCalls)

	repo.managedUpdateErr = ErrCafeManagedKeyEnableUnavailable
	activeStatus := StatusAPIKeyActive
	failedName := "Must not persist"
	_, err = svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{Name: &failedName, Status: &activeStatus})
	require.ErrorIs(t, err, ErrCafeManagedKeyEnableUnavailable)
	require.Equal(t, inactiveName, repo.key.Name)
	require.Equal(t, inactiveStatus, repo.key.Status)
	require.Equal(t, 2, repo.managedCalls)
	require.ErrorIs(t, svc.Delete(context.Background(), 1, 7), ErrCafeManagedKeyProtected)
}

func TestAPIKeyServiceManagedKeyStatusInvalidatesAuthCaches(t *testing.T) {
	managedSourceID := int64(99)
	repo := &cafeManagedKeyRepoStub{key: &APIKey{
		ID:                1,
		UserID:            7,
		Key:               "sk-cafe-managed-cache-test",
		Name:              "Managed cache key",
		Status:            StatusAPIKeyActive,
		ManagedSourceType: APIKeyManagedSourceCafeRoomSeat,
		ManagedSourceID:   &managedSourceID,
	}}
	cache := &cafeManagedKeyCacheStub{}
	svc := NewAPIKeyService(repo, nil, nil, nil, nil, cache, &config.Config{
		APIKeyAuth: config.APIKeyAuthCacheConfig{
			L1Size:       1000,
			L1TTLSeconds: 60,
			L2TTLSeconds: 60,
		},
	})
	require.NotNil(t, svc.authCacheL1)
	cacheKey := svc.authCacheKey(repo.key.Key)
	svc.setAuthCacheL1(cacheKey, &APIKeyAuthCacheEntry{NotFound: true})
	svc.authCacheL1.Wait()
	_, found := svc.authCacheL1.Get(cacheKey)
	require.True(t, found)

	inactiveStatus := "inactive"
	_, err := svc.Update(context.Background(), repo.key.ID, repo.key.UserID, UpdateAPIKeyRequest{Status: &inactiveStatus})
	require.NoError(t, err)
	require.Equal(t, []string{cacheKey}, cache.deleted)
	require.Equal(t, []string{cacheKey}, cache.published)
	require.Eventually(t, func() bool {
		_, ok := svc.authCacheL1.Get(cacheKey)
		return !ok
	}, time.Second, 10*time.Millisecond)
}

type cafeActivationFixture struct {
	service       *CafeRoomActivationService
	apiKeyService *APIKeyService
	cache         *cafeManagedKeyCacheStub
	plan          *dbent.GroupBuyPlan
	room          *dbent.CafeRoom
	round         *dbent.GroupBuyRound
	groupID       int64
	accountID     int64
}

func newCafeActivationFixture(t *testing.T, ctx context.Context, client *dbent.Client, now time.Time, totalSeats int) cafeActivationFixture {
	t.Helper()
	firstUser := createGroupBuyTestUser(t, ctx, client, "cafe-activation-first@example.com")
	secondUser := createGroupBuyTestUser(t, ctx, client, "cafe-activation-second@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	_, err := client.Group.UpdateOneID(groupID).SetAccessMode(CafeRoomGroupAccessMode).Save(ctx)
	require.NoError(t, err)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, totalSeats)
	plan, err = client.GroupBuyPlan.UpdateOneID(plan.ID).
		SetFulfillmentMode(CafeRoomFulfillmentMode).
		SetAutoCreateRoomKey(true).
		SetRoomKeyQuotaUsd(25).
		SetRoomKeyRateLimit5h(5).
		SetRoomKeyRateLimit1d(10).
		SetRoomKeyRateLimit7d(15).
		Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("cafe-activation-account").
		SetPlatform(PlatformOpenAI).
		SetType("api_key").
		SetStatus(StatusActive).
		AddGroupIDs(groupID).
		Save(ctx)
	require.NoError(t, err)
	room, round := createCafeRoomOrderRoom(t, ctx, client, plan.ID, account.ID, now, totalSeats, 91)
	users := []*User{firstUser, secondUser}
	for seatNo := 1; seatNo <= totalSeats; seatNo++ {
		user := users[(seatNo-1)%len(users)]
		_, err := client.GroupBuySeat.Create().
			SetRoundID(round.ID).
			SetPlanID(plan.ID).
			SetUserID(user.ID).
			SetSeatNo(seatNo).
			SetStatus(GroupBuySeatStatusPaid).
			SetShareCount(1).
			SetPaidAt(now).
			Save(ctx)
		require.NoError(t, err)
	}
	_, err = client.GroupBuyRound.UpdateOneID(round.ID).
		SetPaidSeats(totalSeats).
		SetPaidShares(totalSeats).
		SetReservedSeats(0).
		SetReservedShares(0).
		Save(ctx)
	require.NoError(t, err)

	apiKeyRepo := &cafeActivationAPIKeyRepo{client: client}
	cache := &cafeManagedKeyCacheStub{}
	apiKeySvc := NewAPIKeyService(apiKeyRepo, nil, nil, nil, nil, cache, nil)
	service := NewCafeRoomActivationService(client, apiKeySvc, apiKeyRepo)
	service.now = func() time.Time { return now }
	sequence := 0
	service.generateKey = func() (string, error) {
		sequence++
		return fmt.Sprintf("sk-cafe-activation-%032d", sequence), nil
	}
	return cafeActivationFixture{
		service:       service,
		apiKeyService: apiKeySvc,
		cache:         cache,
		plan:          plan,
		room:          room,
		round:         round,
		groupID:       groupID,
		accountID:     account.ID,
	}
}
