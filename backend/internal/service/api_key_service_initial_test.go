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
	count    int64
	countErr error
	created  []*APIKey
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

func TestAPIKeyService_EnsureInitialKey_CreatesSharedKeyWithStudioBridgeRoutes(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 10, Email: "user@test.com", Status: StatusActive}}
	settingsRaw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		DefaultChatGroup:  "101",
		DefaultImageGroup: "102",
		DefaultVideoGroup: "103",
	})
	require.NoError(t, err)
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})
	svc.groupRepo = &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		101: {ID: 101, Name: "chat", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
		102: {ID: 102, Name: "image", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeImage, AllowImageGeneration: true, Hydrated: true},
		103: {ID: 103, Name: "video", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeVideo, Hydrated: true},
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
	require.Nil(t, createdKey.GroupID)
	require.Equal(t, []domain.APIKeyMultiGroupRoute{
		{GroupID: 101, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, TextOnly: true},
		{GroupID: 102, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ImageOnly: true},
		{GroupID: 103, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ModelPatterns: []string{"doubao-seedance-*", "*-video-*"}},
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
		{GroupID: 201, Priority: 1, Weight: 1, CooldownSeconds: 15, Enabled: true, ModelPatterns: []string{"gpt-image-*"}},
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
