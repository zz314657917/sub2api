package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type defaultKeyFallbackAPIKeyRepoStub struct {
	APIKeyRepository
	created         []*APIKey
	backfillGroupID int64
	backfillKeys    []string
}

func (s *defaultKeyFallbackAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	return int64(len(s.created)), nil
}

func (s *defaultKeyFallbackAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	key.ID = int64(len(s.created) + 1)
	s.created = append(s.created, key)
	return nil
}

func (s *defaultKeyFallbackAPIKeyRepoStub) BackfillDefaultKeyFallbackGroup(_ context.Context, groupID int64) ([]string, error) {
	s.backfillGroupID = groupID
	return append([]string(nil), s.backfillKeys...), nil
}

type defaultKeyFallbackUserRepoStub struct {
	UserRepository
	user *User
}

func (s *defaultKeyFallbackUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return s.user, nil
}

type defaultKeyFallbackGroupRepoStub struct {
	GroupRepository
	groups map[int64]*Group
}

func (s *defaultKeyFallbackGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group := s.groups[id]
	if group == nil {
		return nil, ErrGroupNotFound
	}
	clone := *group
	return &clone, nil
}

type defaultKeyFallbackSettingsReaderStub struct {
	settings *StudioBridgeAppSettings
}

func (s *defaultKeyFallbackSettingsReaderStub) GetStudioBridgeLuoyeAISettings(context.Context) (*StudioBridgeAppSettings, error) {
	return s.settings, nil
}

type defaultKeyFallbackCacheStub struct {
	APIKeyCache
	deleted []string
}

func (s *defaultKeyFallbackCacheStub) DeleteAuthCache(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}

func (s *defaultKeyFallbackCacheStub) PublishAuthCacheInvalidation(context.Context, string) error {
	return nil
}

func TestDefaultKeyFallbackS93CreatesSystemDefaultWithBaseGroup(t *testing.T) {
	repo := &defaultKeyFallbackAPIKeyRepoStub{}
	groupRepo := &defaultKeyFallbackGroupRepoStub{groups: map[int64]*Group{
		10: {
			ID:                 10,
			Name:               "fallback",
			Status:             StatusActive,
			Platform:           PlatformOpenAI,
			RoutingScope:       GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"*"},
			IsExclusive:        true,
			Hydrated:           true,
		},
	}}
	svc := NewAPIKeyService(
		repo,
		&defaultKeyFallbackUserRepoStub{user: &User{ID: 7, Status: StatusActive}},
		groupRepo,
		nil,
		nil,
		nil,
		&config.Config{Default: config.DefaultConfig{APIKeyPrefix: "test-"}},
	)
	svc.SetStudioBridgeDefaultRouteSettingsReader(&defaultKeyFallbackSettingsReaderStub{
		settings: &StudioBridgeAppSettings{DefaultFallbackGroup: "10"},
	})

	key, created, err := svc.EnsureInitialKey(context.Background(), 7)

	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, key.GroupID)
	require.Equal(t, int64(10), *key.GroupID)
	require.Equal(t, AccountPoolStrategySharedOnly, key.AccountPoolStrategy)
	require.Len(t, repo.created, 1)
}

func TestDefaultKeyFallbackS93BackfillUsesSavedGroupAndInvalidatesChangedKeys(t *testing.T) {
	repo := &defaultKeyFallbackAPIKeyRepoStub{backfillKeys: []string{"sk-one", "sk-two"}}
	cache := &defaultKeyFallbackCacheStub{}
	svc := NewAPIKeyService(
		repo,
		nil,
		&defaultKeyFallbackGroupRepoStub{groups: map[int64]*Group{
			20: {ID: 20, Name: "fallback", Status: StatusActive, Hydrated: true},
		}},
		nil,
		nil,
		cache,
		&config.Config{},
	)
	svc.SetStudioBridgeDefaultRouteSettingsReader(&defaultKeyFallbackSettingsReaderStub{
		settings: &StudioBridgeAppSettings{DefaultFallbackGroup: "20"},
	})

	groupID, updated, err := svc.BackfillDefaultKeyFallbackGroup(context.Background())

	require.NoError(t, err)
	require.Equal(t, int64(20), groupID)
	require.Equal(t, 2, updated)
	require.Equal(t, int64(20), repo.backfillGroupID)
	require.Len(t, cache.deleted, 2)
}
