package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type createLimitsRepo struct {
	APIKeyRepository
	count               int64
	countErr, createErr error
	created             []*APIKey
	exists              bool
}

func (r *createLimitsRepo) CountByUserID(context.Context, int64) (int64, error) {
	return r.count, r.countErr
}
func (r *createLimitsRepo) ExistsByKey(context.Context, string) (bool, error) {
	return r.exists, nil
}
func (r *createLimitsRepo) Create(_ context.Context, key *APIKey) error {
	if r.createErr != nil {
		return r.createErr
	}
	key.ID = int64(len(r.created) + 1)
	r.created = append(r.created, key)
	r.count++
	return nil
}
func (r *createLimitsRepo) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	return r.created[0].Key, r.created[0].UserID, nil
}
func (r *createLimitsRepo) DeleteWithAudit(context.Context, int64) error { return nil }
func (r *createLimitsRepo) Delete(context.Context, int64) error          { return nil }
func (r *createLimitsRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *createLimitsRepo) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *r.created[0]
	return &clone, nil
}
func (r *createLimitsRepo) Update(_ context.Context, key *APIKey, _ APIKeyUpdateFields) error {
	r.created[0] = key
	return nil
}

type createLimitsUser struct{ UserRepository }

func (*createLimitsUser) GetByID(_ context.Context, id int64) (*User, error) {
	return &User{ID: id, Status: StatusActive}, nil
}

type createLimitsLegacyCache struct {
	APIKeyCache
	failures, resets int
}

func (c *createLimitsLegacyCache) GetCreateAttemptCount(context.Context, int64) (int, error) {
	return c.failures, nil
}
func (c *createLimitsLegacyCache) IncrementCreateAttemptCount(context.Context, int64) error {
	c.failures++
	return nil
}
func (c *createLimitsLegacyCache) DeleteCreateAttemptCount(context.Context, int64) error {
	c.failures = 0
	c.resets++
	return nil
}
func (*createLimitsLegacyCache) DeleteAuthCache(context.Context, string) error { return nil }
func (*createLimitsLegacyCache) PublishAuthCacheInvalidation(context.Context, string) error {
	return nil
}

type createLimitsCache struct {
	createLimitsLegacyCache
	mu       sync.Mutex
	attempts int64
	err      error
}

func (c *createLimitsCache) IncrementAPIKeyCreateCount(context.Context, int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.attempts++
	return c.attempts, c.err
}
func createLimitsService(r *createLimitsRepo, cache APIKeyCache, maxKeys, hourly int) *APIKeyService {
	return &APIKeyService{apiKeyRepo: r, userRepo: &createLimitsUser{}, cache: cache,
		cfg: &config.Config{Default: config.DefaultConfig{APIKeyPrefix: "sk-"},
			APIKeyCreate: config.APIKeyCreateConfig{MaxActivePerUser: maxKeys, MaxPerUserPerHour: hourly}}}
}

func TestAPIKeyCreateLimitsCustomAndGeneratedShareQuota(t *testing.T) {
	r, cache := &createLimitsRepo{}, &createLimitsCache{}
	s := createLimitsService(r, cache, 200, 2)
	custom := "custom-key-123456789"
	_, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{CustomKey: &custom})
	require.NoError(t, err)
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{})
	require.NoError(t, err)
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrAPIKeyCreateRateLimited)
	require.Len(t, r.created, 2)
	require.EqualValues(t, 3, cache.attempts)
}

func TestAPIKeyCreateLimitsCountAndIndependentDisable(t *testing.T) {
	for _, tc := range []struct {
		name        string
		count       int64
		max, hourly int
		countErr    error
		want        error
		attempts    int64
	}{
		{"count-at-cap", 2, 2, 60, nil, ErrAPIKeyCountExceeded, 0},
		{"count-above-cap", 3, 2, 60, nil, ErrAPIKeyCountExceeded, 0},
		{"count-error", 0, 2, 60, errors.New("count failed"), nil, 0},
		{"count-disabled", 500, 0, 60, nil, nil, 1},
		{"hourly-disabled", 0, 2, 0, nil, nil, 0},
		{"both-disabled", 500, 0, 0, nil, nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &createLimitsRepo{count: tc.count, countErr: tc.countErr}
			c := &createLimitsCache{}
			_, err := createLimitsService(r, c, tc.max, tc.hourly).Create(context.Background(), 1, CreateAPIKeyRequest{})
			if tc.countErr != nil {
				require.ErrorIs(t, err, tc.countErr)
			} else if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
			} else {
				require.NoError(t, err)
				require.Len(t, r.created, 1)
			}
			require.Equal(t, tc.attempts, c.attempts)
		})
	}
}

func TestAPIKeyCreateLimitsPersistenceFailureConsumesAttempt(t *testing.T) {
	r, c := &createLimitsRepo{createErr: errors.New("database unavailable")}, &createLimitsCache{}
	s := createLimitsService(r, c, 200, 1)
	_, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{})
	require.ErrorIs(t, err, r.createErr)
	r.createErr = nil
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrAPIKeyCreateRateLimited)
	require.Empty(t, r.created)
}

func TestAPIKeyCreateLimitsCacheFailOpen(t *testing.T) {
	for _, cache := range []APIKeyCache{nil, &createLimitsLegacyCache{}, &createLimitsCache{err: errors.New("redis unavailable")}} {
		r := &createLimitsRepo{}
		s := createLimitsService(r, cache, 200, 1)
		for i := 0; i < 2; i++ {
			_, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{})
			require.NoError(t, err)
		}
		require.Len(t, r.created, 2)
	}
}

func TestAPIKeyCreateLimitsNilConfigDisablesGuards(t *testing.T) {
	r, c := &createLimitsRepo{count: 500, countErr: errors.New("must not count")}, &createLimitsCache{}
	s := createLimitsService(r, c, 1, 1)
	s.cfg = nil
	custom := "nil-config-key-12345"
	_, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{CustomKey: &custom})
	require.NoError(t, err)
	require.Zero(t, c.attempts)
	require.Len(t, r.created, 1)
}

func TestAPIKeyCreateLimitsValidationAndFailedCustomCounter(t *testing.T) {
	r, c := &createLimitsRepo{exists: true}, &createLimitsCache{}
	s := createLimitsService(r, c, 200, 1)
	custom := "duplicate-key-123456"
	_, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{CustomKey: &custom})
	require.ErrorIs(t, err, ErrAPIKeyExists)
	require.Equal(t, 1, c.failures)
	require.Zero(t, c.attempts)
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{Quota: -1})
	require.Error(t, err)
	require.Zero(t, c.attempts)
	c.failures = apiKeyMaxErrorsPerHour
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{CustomKey: &custom})
	require.ErrorIs(t, err, ErrAPIKeyRateLimited)
	require.Zero(t, c.attempts)
}

func TestAPIKeyCreateLimitsEnsureInitialKeyConsumesOnlyCreation(t *testing.T) {
	r, c := &createLimitsRepo{}, &createLimitsCache{}
	s := createLimitsService(r, c, 200, 1)
	key, created, err := s.EnsureInitialKey(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, created)
	require.True(t, key.IsDefault)
	require.EqualValues(t, 1, c.attempts)
	_, created, err = s.EnsureInitialKey(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, created)
	require.EqualValues(t, 1, c.attempts)
	r.count = 0
	_, created, err = s.EnsureInitialKey(context.Background(), 1)
	require.ErrorIs(t, err, ErrAPIKeyCreateRateLimited)
	require.False(t, created)
}

func TestAPIKeyCreateLimitsUpdateDeleteDoNotRefund(t *testing.T) {
	r, c := &createLimitsRepo{}, &createLimitsCache{}
	s := createLimitsService(r, c, 200, 1)
	key, err := s.Create(context.Background(), 1, CreateAPIKeyRequest{Name: "ordinary"})
	require.NoError(t, err)
	name := "changed"
	_, err = s.Update(context.Background(), key.ID, 1, UpdateAPIKeyRequest{Name: &name})
	require.NoError(t, err)
	require.EqualValues(t, 1, c.attempts)
	require.NoError(t, s.Delete(context.Background(), key.ID, 1))
	require.Equal(t, 1, c.resets)
	require.EqualValues(t, 1, c.attempts)
	_, err = s.Create(context.Background(), 1, CreateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrAPIKeyCreateRateLimited)
}
