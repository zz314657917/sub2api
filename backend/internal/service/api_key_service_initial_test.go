//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
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

func TestAPIKeyService_EnsureInitialKey_CreatesUngroupedSharedKey(t *testing.T) {
	repo := &initialAPIKeyRepoStub{}
	userRepo := &userRepoStub{user: &User{ID: 10, Email: "user@test.com", Status: StatusActive}}
	svc := NewAPIKeyService(repo, userRepo, nil, nil, nil, nil, &config.Config{
		Default: config.DefaultConfig{APIKeyPrefix: "test-"},
	})

	key, created, err := svc.EnsureInitialKey(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, key)
	require.Len(t, repo.created, 1)

	createdKey := repo.created[0]
	require.Equal(t, int64(10), createdKey.UserID)
	require.Equal(t, initialAPIKeyName, createdKey.Name)
	require.Nil(t, createdKey.GroupID)
	require.Empty(t, createdKey.MultiGroupRoutes)
	require.Equal(t, AccountPoolStrategySharedOnly, createdKey.AccountPoolStrategy)
	require.True(t, strings.HasPrefix(createdKey.Key, "test-"))
	require.Equal(t, StatusActive, createdKey.Status)
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
