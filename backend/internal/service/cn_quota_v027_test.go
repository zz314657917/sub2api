package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const v027KimiQuotaExhaustedBody = `{"error":{"message":"You've reached your weekly usage limit. Your quota will reset when the current window ends.","type":"access_terminated_error"}}`

type v027CNQuotaRepo struct {
	AccountRepository
	rateLimitErr error
	tempErr      error
	rateCalls    int
	tempCalls    int
	errorCalls   int
	rateUntil    time.Time
	tempUntil    time.Time
	tempReason   string
}

type v027CNQuotaRuntimeBlocker struct {
	accounts []*Account
	untils   []time.Time
	reasons  []string
}

func (b *v027CNQuotaRuntimeBlocker) BlockAccountScheduling(account *Account, until time.Time, reason string) {
	b.accounts = append(b.accounts, account)
	b.untils = append(b.untils, until)
	b.reasons = append(b.reasons, reason)
}

func (b *v027CNQuotaRuntimeBlocker) ClearAccountSchedulingBlock(int64) {}

func newV027CNQuotaRepo() *v027CNQuotaRepo {
	return &v027CNQuotaRepo{}
}

func (r *v027CNQuotaRepo) SetRateLimited(_ context.Context, _ int64, until time.Time) error {
	r.rateCalls++
	r.rateUntil = until
	return r.rateLimitErr
}

func (r *v027CNQuotaRepo) SetTempUnschedulable(_ context.Context, _ int64, until time.Time, reason string) error {
	r.tempCalls++
	r.tempUntil = until
	r.tempReason = reason
	return r.tempErr
}

func (r *v027CNQuotaRepo) SetError(_ context.Context, _ int64, _ string) error {
	r.errorCalls++
	return nil
}

func newV027CNQuotaAccount(fiveHourReset, weeklyReset time.Time) *Account {
	extra := map[string]any{}
	if !fiveHourReset.IsZero() {
		extra[cnExtraKey(PlatformKimi, cnExtraSuffix5hReset)] = fiveHourReset.UTC().Format(time.RFC3339)
	}
	if !weeklyReset.IsZero() {
		extra[cnExtraKey(PlatformKimi, cnExtraSuffixWeeklyReset)] = weeklyReset.UTC().Format(time.RFC3339)
	}
	return &Account{
		ID:          2701,
		Platform:    PlatformKimi,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"account_mode": AccountModeCoding},
		Extra:       extra,
	}
}

func runV027CNQuota403(repo *v027CNQuotaRepo, account *Account, body string) bool {
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	return service.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(body))
}

func runV027CNQuota403WithBlocker(repo *v027CNQuotaRepo, blocker *v027CNQuotaRuntimeBlocker, account *Account, body string) bool {
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	service.SetAccountRuntimeBlocker(blocker)
	return service.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(body))
}

func TestV027CNQuotaFutureSnapshotPersistsRateLimit(t *testing.T) {
	now := time.Now()
	repo := newV027CNQuotaRepo()
	account := newV027CNQuotaAccount(now.Add(2*time.Hour), now.Add(7*24*time.Hour))

	require.True(t, runV027CNQuota403(repo, account, v027KimiQuotaExhaustedBody))
	require.Zero(t, repo.errorCalls)
	require.Equal(t, 1, repo.rateCalls)
	require.Zero(t, repo.tempCalls)
	require.WithinDuration(t, now.Add(2*time.Hour), repo.rateUntil, time.Second)
}

func TestV027CNQuotaTextSignalPersistsRateLimit(t *testing.T) {
	repo := newV027CNQuotaRepo()
	account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Now().Add(7*24*time.Hour))
	body := `{"error":{"message":"You've reached your weekly usage limit. Your quota will reset when the current window ends."}}`

	require.True(t, runV027CNQuota403(repo, account, body))
	require.Zero(t, repo.errorCalls)
	require.Equal(t, 1, repo.rateCalls)
	require.Zero(t, repo.tempCalls)
}

func TestV027CNQuotaStaleOrMissingSnapshotUsesBoundedTempPause(t *testing.T) {
	for _, tc := range []struct {
		name    string
		account *Account
	}{
		{"stale", newV027CNQuotaAccount(time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))},
		{"missing", newV027CNQuotaAccount(time.Time{}, time.Time{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newV027CNQuotaRepo()
			before := time.Now()
			require.True(t, runV027CNQuota403(repo, tc.account, v027KimiQuotaExhaustedBody))
			require.Zero(t, repo.errorCalls)
			require.Zero(t, repo.rateCalls)
			require.Equal(t, 1, repo.tempCalls)
			require.Contains(t, repo.tempReason, cnQuotaExhaustedReasonPrefix)
			require.WithinDuration(t, before.Add(time.Duration(openAI403CooldownMinutesDefault)*time.Minute), repo.tempUntil, time.Second)
		})
	}
}

func TestV027CNQuotaRateLimitPersistenceFailureFallsBackToTempPause(t *testing.T) {
	repo := newV027CNQuotaRepo()
	repo.rateLimitErr = errors.New("rate limit write failed")
	account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Now().Add(7*24*time.Hour))

	require.True(t, runV027CNQuota403(repo, account, v027KimiQuotaExhaustedBody))
	require.Zero(t, repo.errorCalls)
	require.Equal(t, 1, repo.rateCalls)
	require.Equal(t, 1, repo.tempCalls)
}

func TestV027CNQuotaRuntimeBlockerFollowsSuccessfulPersistence(t *testing.T) {
	t.Run("future snapshot", func(t *testing.T) {
		now := time.Now()
		repo := newV027CNQuotaRepo()
		blocker := &v027CNQuotaRuntimeBlocker{}
		account := newV027CNQuotaAccount(now.Add(time.Hour), now.Add(7*24*time.Hour))

		require.True(t, runV027CNQuota403WithBlocker(repo, blocker, account, v027KimiQuotaExhaustedBody))
		require.Equal(t, 1, repo.rateCalls)
		require.Zero(t, repo.tempCalls)
		require.Len(t, blocker.accounts, 1)
		require.Same(t, account, blocker.accounts[0])
		require.WithinDuration(t, repo.rateUntil, blocker.untils[0], time.Second)
		require.Equal(t, cnQuotaExhaustedReasonPrefix, blocker.reasons[0])
	})

	t.Run("missing snapshot", func(t *testing.T) {
		repo := newV027CNQuotaRepo()
		blocker := &v027CNQuotaRuntimeBlocker{}
		account := newV027CNQuotaAccount(time.Time{}, time.Time{})

		require.True(t, runV027CNQuota403WithBlocker(repo, blocker, account, v027KimiQuotaExhaustedBody))
		require.Zero(t, repo.rateCalls)
		require.Equal(t, 1, repo.tempCalls)
		require.Len(t, blocker.accounts, 1)
		require.WithinDuration(t, repo.tempUntil, blocker.untils[0], time.Second)
		require.Equal(t, cnQuotaExhaustedReasonPrefix, blocker.reasons[0])
	})

	t.Run("both writes fail", func(t *testing.T) {
		repo := newV027CNQuotaRepo()
		repo.rateLimitErr = errors.New("rate limit write failed")
		repo.tempErr = errors.New("temp pause write failed")
		blocker := &v027CNQuotaRuntimeBlocker{}
		account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Now().Add(7*24*time.Hour))

		require.True(t, runV027CNQuota403WithBlocker(repo, blocker, account, v027KimiQuotaExhaustedBody))
		require.Equal(t, 1, repo.rateCalls)
		require.Equal(t, 1, repo.tempCalls)
		require.Empty(t, blocker.accounts)
	})
}

func TestV027CNQuotaTempPausePersistenceFailureDoesNotSetError(t *testing.T) {
	repo := newV027CNQuotaRepo()
	repo.tempErr = errors.New("temp pause write failed")
	account := newV027CNQuotaAccount(time.Time{}, time.Time{})

	require.True(t, runV027CNQuota403(repo, account, v027KimiQuotaExhaustedBody))
	require.Zero(t, repo.errorCalls)
	require.Zero(t, repo.rateCalls)
	require.Equal(t, 1, repo.tempCalls)
}

func TestV027CNQuotaNonTargetAndConcurrencyKeepExisting403Paths(t *testing.T) {
	t.Run("non coding plan quota body", func(t *testing.T) {
		repo := newV027CNQuotaRepo()
		account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Now().Add(7*24*time.Hour))
		account.Credentials = map[string]any{"account_mode": AccountModePayG}

		require.True(t, runV027CNQuota403(repo, account, v027KimiQuotaExhaustedBody))
		require.Zero(t, repo.rateCalls)
		require.Equal(t, 1, repo.errorCalls)
	})

	t.Run("concurrency message", func(t *testing.T) {
		repo := newV027CNQuotaRepo()
		account := newV027CNQuotaAccount(time.Now().Add(time.Hour), time.Now().Add(7*24*time.Hour))
		body := `{"error":{"message":"You've reached your concurrent request limit. Please wait for your ongoing requests to finish and try again.","type":"access_terminated_error"}}`

		require.True(t, runV027CNQuota403(repo, account, body))
		require.Zero(t, repo.errorCalls)
		require.Zero(t, repo.rateCalls)
		require.Equal(t, 1, repo.tempCalls)
		require.Contains(t, repo.tempReason, cnConcurrencyLimitReasonPrefix)
	})
}
