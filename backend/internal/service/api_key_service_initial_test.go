//go:build unit

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

type initialAPIKeyRepoStub struct {
	authRepoStub
	count           int64
	countErr        error
	created         []*APIKey
	backfillGroupID int64
	backfillKeys    []string
	backfillErr     error
}

func (s *initialAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return s.count, nil
}

func (s *initialAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	key.ID = int64(len(s.created) + 1)
	s.created = append(s.created, key)
	return nil
}

func (s *initialAPIKeyRepoStub) BackfillDefaultKeyFallbackGroup(_ context.Context, groupID int64) ([]string, error) {
	s.backfillGroupID = groupID
	return append([]string(nil), s.backfillKeys...), s.backfillErr
}

func TestAPIKeyService_EnsureInitialKey_CreatesSharedKeyWithStudioBridgeRoutes(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 10, Email: "user@test.com", Status: StatusActive}}
	settingsRaw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		DefaultChatGroup:     "101",
		DefaultImageGroup:    "102",
		DefaultVideoGroup:    "103",
		DefaultFallbackGroup: "104",
	})
	require.NoError(t, err)
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})
	svc.groupRepo = &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		101: {ID: 101, Name: "chat", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
		102: {ID: 102, Name: "image", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeImage, AllowImageGeneration: true, Hydrated: true},
		103: {ID: 103, Name: "video", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeVideo, Hydrated: true},
		104: {ID: 104, Name: "fallback", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, ModelMatchPatterns: []string{"*"}, IsExclusive: true, Hydrated: true},
	}}
	svc.SetStudioBridgeDefaultRouteSettingsReader(NewSettingService(&studioBridgeSettingRepoStub{
		values: map[string]string{SettingKeyStudioBridgeLuoyeAI: settingsRaw},
	}, &config.Config{}))

	key, created, err := svc.EnsureInitialKey(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, key)
	require.Len(t, repo.created, 1)

	createdKey := repo.created[0]
	require.Equal(t, int64(10), createdKey.UserID)
	require.Equal(t, DefaultAPIKeyName, createdKey.Name)
	require.NotNil(t, createdKey.GroupID)
	require.Equal(t, int64(104), *createdKey.GroupID)
	require.Equal(t, []domain.APIKeyMultiGroupRoute{
		{GroupID: 101, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, TextOnly: true},
		{GroupID: 102, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ImageOnly: true},
		{GroupID: 103, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true},
	}, createdKey.MultiGroupRoutes)
	require.Equal(t, AccountPoolStrategySharedOnly, createdKey.AccountPoolStrategy)
	require.True(t, strings.HasPrefix(createdKey.Key, "test-"))
	require.Equal(t, StatusActive, createdKey.Status)
}

func TestAPIKeyService_EnsureInitialKey_CreatesSharedKeyWithConfiguredStudioBridgeRoutes(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 14, Email: "user@test.com", Status: StatusActive}}
	settingsRaw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		DefaultAPIRoutes: []StudioBridgeDefaultAPIRoute{
			{
				GroupID:         "201",
				Priority:        1,
				Weight:          2,
				CooldownSeconds: 15,
				Enabled:         true,
				TextOnly:        true,
			},
			{
				GroupID:         "201",
				Priority:        1,
				Weight:          1,
				CooldownSeconds: 15,
				Enabled:         true,
				ModelPatterns:   []string{"gpt-image-*"},
			},
		},
	})
	require.NoError(t, err)
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})
	svc.groupRepo = &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		201: {ID: 201, Name: "multi", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
	}}
	svc.SetStudioBridgeDefaultRouteSettingsReader(NewSettingService(&studioBridgeSettingRepoStub{
		values: map[string]string{SettingKeyStudioBridgeLuoyeAI: settingsRaw},
	}, &config.Config{}))

	_, created, err := svc.EnsureInitialKey(context.Background(), 14)
	require.NoError(t, err)
	require.True(t, created)
	require.Len(t, repo.created, 1)
	require.Equal(t, []domain.APIKeyMultiGroupRoute{
		{GroupID: 201, Priority: 1, Weight: 2, CooldownSeconds: 15, Enabled: true, TextOnly: true},
		{GroupID: 201, Priority: 1, Weight: 1, CooldownSeconds: 15, Enabled: true},
	}, repo.created[0].MultiGroupRoutes)
}

func TestAPIKeyService_EnsureInitialKey_CreatesUngroupedSharedKeyWhenNoStudioBridgeRoutes(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 12, Email: "user@test.com", Status: StatusActive}}
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})

	key, created, err := svc.EnsureInitialKey(context.Background(), 12)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, key)
	require.Len(t, repo.created, 1)
	require.Empty(t, repo.created[0].MultiGroupRoutes)
}

func TestAPIKeyService_EnsureInitialKey_SkipsInvalidStudioBridgeGroups(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 13, Email: "user@test.com", Status: StatusActive}}
	settingsRaw, err := json.Marshal(StudioBridgeAppSettings{
		DefaultChatGroup:  "101",
		DefaultImageGroup: "999",
		DefaultVideoGroup: "bad",
	})
	require.NoError(t, err)
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})
	svc.groupRepo = &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		101: {ID: 101, Name: "chat", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
	}}
	svc.SetStudioBridgeDefaultRouteSettingsReader(NewSettingService(&studioBridgeSettingRepoStub{
		values: map[string]string{SettingKeyStudioBridgeLuoyeAI: string(settingsRaw)},
	}, &config.Config{}))

	key, created, err := svc.EnsureInitialKey(context.Background(), 13)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, key)
	require.Len(t, repo.created, 1)
	require.Equal(t, []domain.APIKeyMultiGroupRoute{
		{GroupID: 101, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, TextOnly: true},
	}, repo.created[0].MultiGroupRoutes)
}

func TestAPIKeyService_EnsureInitialKey_SkipsWhenUserAlreadyHasKey(t *testing.T) {
	repo := &initialAPIKeyRepoStub{count: 1}
	userRepo := &userRepoStub{user: &User{ID: 11, Email: "user@test.com", Status: StatusActive}}
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{})

	key, created, err := svc.EnsureInitialKey(context.Background(), 11)
	require.NoError(t, err)
	require.False(t, created)
	require.Nil(t, key)
	require.Empty(t, repo.created)
}

func TestAPIKeyService_Create_StillChecksBaseGroupPermissionForUserCreatedKey(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 15, Email: "user@test.com", Status: StatusActive}}
	groupID := int64(301)
	svc := NewAPIKeyService(repo, userRepo, &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		groupID: {ID: groupID, Name: "exclusive", Status: StatusActive, IsExclusive: true, Hydrated: true},
	}}, nil, nil, nil, &config.Config{Default: config.DefaultConfig{APIKeyPrefix: "test-"}})

	_, err := svc.Create(context.Background(), 15, CreateAPIKeyRequest{
		Name:                "user key",
		GroupID:             &groupID,
		AccountPoolStrategy: AccountPoolStrategySharedOnly,
	})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Empty(t, repo.created)
}

func TestAPIKeyService_BackfillDefaultKeyFallbackGroup_UsesConfiguredActiveGroupAndInvalidatesKeys(t *testing.T) {
	repo := &initialAPIKeyRepoStub{backfillKeys: []string{"sk-one", "sk-two"}}
	cache := &authCacheStub{}
	svc := NewAPIKeyService(repo, nil, &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		401: {ID: 401, Name: "fallback", Status: StatusActive, Hydrated: true},
	}}, nil, nil, cache, &config.Config{})
	settingsRaw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{DefaultFallbackGroup: "401"})
	require.NoError(t, err)
	svc.SetStudioBridgeDefaultRouteSettingsReader(NewSettingService(&studioBridgeSettingRepoStub{
		values: map[string]string{SettingKeyStudioBridgeLuoyeAI: settingsRaw},
	}, &config.Config{}))

	groupID, updated, err := svc.BackfillDefaultKeyFallbackGroup(context.Background())

	require.NoError(t, err)
	require.Equal(t, int64(401), groupID)
	require.Equal(t, 2, updated)
	require.Equal(t, int64(401), repo.backfillGroupID)
	require.Len(t, cache.deleteAuthKeys, 2)
}

func TestAPIKeyService_BackfillDefaultKeyFallbackGroup_RejectsMissingOrInactiveSetting(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fallback string
		group    *Group
		wantErr  error
	}{
		{name: "missing", wantErr: ErrDefaultKeyFallbackGroupRequired},
		{name: "inactive", fallback: "402", group: &Group{ID: 402, Status: StatusDisabled}, wantErr: ErrDefaultKeyFallbackGroupInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &initialAPIKeyRepoStub{}
			groups := map[int64]*Group{}
			if tc.group != nil {
				groups[tc.group.ID] = tc.group
			}
			svc := NewAPIKeyService(repo, nil, &groupRepoStubForStudioBridgeGateway{groups: groups}, nil, nil, nil, &config.Config{})
			settingsRaw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{DefaultFallbackGroup: tc.fallback})
			require.NoError(t, err)
			svc.SetStudioBridgeDefaultRouteSettingsReader(NewSettingService(&studioBridgeSettingRepoStub{
				values: map[string]string{SettingKeyStudioBridgeLuoyeAI: settingsRaw},
			}, &config.Config{}))

			_, _, err = svc.BackfillDefaultKeyFallbackGroup(context.Background())

			require.ErrorIs(t, err, tc.wantErr)
			require.Zero(t, repo.backfillGroupID)
		})
	}
}
