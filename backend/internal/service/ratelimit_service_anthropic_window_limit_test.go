//go:build unit

package service

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type anthropicWindowLimitRepo struct {
	mockAccountRepoForGemini
	rateLimitCalls          int
	tempUnschedCalls        int
	lastRateLimitReset      time.Time
	modelRateLimitCalls     int
	lastModelRateLimitScope string
	lastModelRateLimitReset time.Time
	sessionWindowCalls      int
	lastExtraUpdates        map[string]any
	extraHistory            []map[string]any
}

func (r *anthropicWindowLimitRepo) SetRateLimited(_ context.Context, _ int64, resetAt time.Time) error {
	r.rateLimitCalls++
	r.lastRateLimitReset = resetAt
	return nil
}

func (r *anthropicWindowLimitRepo) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, _ string) error {
	r.tempUnschedCalls++
	return nil
}

func (r *anthropicWindowLimitRepo) SetModelRateLimit(_ context.Context, _ int64, scope string, resetAt time.Time) error {
	r.modelRateLimitCalls++
	r.lastModelRateLimitScope = scope
	r.lastModelRateLimitReset = resetAt
	return nil
}

func (r *anthropicWindowLimitRepo) UpdateSessionWindow(_ context.Context, _ int64, _, _ *time.Time, _ string) error {
	r.sessionWindowCalls++
	return nil
}

func (r *anthropicWindowLimitRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.lastExtraUpdates = updates
	r.extraHistory = append(r.extraHistory, updates)
	return nil
}

func TestUpdateSessionWindowPreservesIndependentWeeklyUsage(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	for _, tc := range []struct {
		name      string
		regular   int64
		fable     int64
		keep7d    bool
		keepFable bool
	}{
		{"regular valid and fable at boundary", now.Unix() + 3600, now.Unix(), true, false},
		{"regular expired and fable valid", now.Unix() - 1, now.Unix() + 3600, false, true},
		{"both invalid", 0, -1, false, false},
		{"both valid", now.Unix() + 7200, now.Unix() + 3600, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &anthropicWindowLimitRepo{}
			svc := NewRateLimitService(repo, nil, nil, nil, nil)
			oldEnd := now.Add(-time.Hour)
			account := &Account{ID: 1, Type: AccountTypeSetupToken, Platform: PlatformAnthropic, SessionWindowEnd: &oldEnd,
				Extra: map[string]any{"passive_usage_7d_reset": float64(tc.regular), "passive_usage_7d_oi_reset": tc.fable}}
			headers := http.Header{}
			headers.Set("anthropic-ratelimit-unified-5h-status", "allowed")
			headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.15")
			headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.3")
			svc.UpdateSessionWindow(context.Background(), account, headers)
			require.Equal(t, 1, repo.sessionWindowCalls)
			require.Len(t, repo.extraHistory, 2)
			clear, sample := repo.extraHistory[0], repo.extraHistory[1]
			for key, kept := range map[string]bool{
				"session_window_utilization": false, "passive_usage_7d_reset": tc.keep7d,
				"passive_usage_7d_utilization": tc.keep7d, "passive_usage_7d_oi_reset": tc.keepFable,
				"passive_usage_7d_oi_utilization": tc.keepFable, "passive_usage_sampled_at": tc.keep7d || tc.keepFable,
			} {
				value, wrote := clear[key]
				require.Equal(t, !kept, wrote, key)
				require.Nil(t, value, key)
			}
			require.Equal(t, 0.15, sample["session_window_utilization"])
			require.Equal(t, 0.3, sample["passive_usage_7d_utilization"])
			require.NotEmpty(t, sample["passive_usage_sampled_at"])
		})
	}
}

func TestAnthropicWindowResetExtraClearsExactlyAtReset(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	extra := map[string]any{"passive_usage_7d_reset": now.Unix(), "passive_usage_7d_oi_reset": now.Unix() + 1}
	updates := anthropicWindowResetExtra(extra, now)
	require.Contains(t, updates, "passive_usage_7d_reset")
	require.NotContains(t, updates, "passive_usage_7d_oi_reset")
}

func TestHandleUpstreamError_AnthropicWindowLimitPreemptsTempUnschedRule(t *testing.T) {
	resetAt := time.Now().Add(3 * time.Hour).Truncate(time.Second)
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.02")
	headers.Set("anthropic-ratelimit-unified-5h-reset", strconv.FormatInt(resetAt.Unix(), 10))

	repo := &anthropicWindowLimitRepo{}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)
	account := &Account{
		ID:       42,
		Type:     AccountTypeOAuth,
		Platform: PlatformAnthropic,
		Credentials: map[string]any{
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{
				map[string]any{
					"error_code":       float64(http.StatusTooManyRequests),
					"keywords":         []any{"rate limit"},
					"duration_minutes": float64(10),
				},
			},
		},
	}

	svc.HandleUpstreamError(
		context.Background(),
		account,
		http.StatusTooManyRequests,
		headers,
		[]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"This request would exceed your account's rate limit. Please try again later."}}`),
	)

	require.Zero(t, repo.tempUnschedCalls, "official Anthropic window limits should not be shortened by local temp-unsched rules")
	require.Equal(t, 1, repo.rateLimitCalls)
	require.Equal(t, resetAt, repo.lastRateLimitReset)
}

func fable429Headers(reset5h, resetOI time.Time) http.Header {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-reset", strconv.FormatInt(reset5h.Unix(), 10))
	headers.Set("anthropic-ratelimit-unified-5h-status", "allowed")
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.41")
	headers.Set("anthropic-ratelimit-unified-7d-reset", strconv.FormatInt(resetOI.Unix(), 10))
	headers.Set("anthropic-ratelimit-unified-7d-status", "allowed")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.56")
	headers.Set("anthropic-ratelimit-unified-7d_oi-reset", strconv.FormatInt(resetOI.Unix(), 10))
	headers.Set("anthropic-ratelimit-unified-7d_oi-status", "rejected")
	headers.Set("anthropic-ratelimit-unified-7d_oi-surpassed-threshold", "1.0")
	headers.Set("anthropic-ratelimit-unified-7d_oi-utilization", "1.0")
	headers.Set("anthropic-ratelimit-unified-reset", strconv.FormatInt(resetOI.Unix(), 10))
	return headers
}

func TestHandleUpstreamError_Anthropic7dOiOnlyMarksModelRateLimit(t *testing.T) {
	now := time.Now()
	reset5h := now.Add(2 * time.Hour).Truncate(time.Second)
	resetOI := now.Add(80 * time.Hour).Truncate(time.Second)
	repo := &anthropicWindowLimitRepo{}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)
	account := &Account{ID: 42, Type: AccountTypeOAuth, Platform: PlatformAnthropic}

	shouldDisable := svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, fable429Headers(reset5h, resetOI), nil, "claude-fable-5-1")

	require.False(t, shouldDisable)
	require.Zero(t, repo.rateLimitCalls, "7d_oi must not rate limit the whole account")
	require.Zero(t, repo.tempUnschedCalls, "7d_oi must not trigger temporary unscheduling")
	require.Zero(t, repo.sessionWindowCalls, "7d_oi must not rewrite the 5h session window")
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, anthropicFableRateLimitKey, repo.lastModelRateLimitScope)
	require.Equal(t, resetOI, repo.lastModelRateLimitReset)
	require.Equal(t, 1.0, repo.lastExtraUpdates["passive_usage_7d_oi_utilization"])
	require.Equal(t, resetOI.Unix(), repo.lastExtraUpdates["passive_usage_7d_oi_reset"])
}

func TestHandleUpstreamError_AnthropicFableCreditsRequiredOnlyMarksModelRateLimit(t *testing.T) {
	resetAt := time.Now().Add(29 * 24 * time.Hour).Truncate(time.Second)
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-reset", strconv.FormatInt(resetAt.Unix(), 10))
	repo := &anthropicWindowLimitRepo{}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)
	account := &Account{ID: 42, Type: AccountTypeOAuth, Platform: PlatformAnthropic}
	body := []byte(`{"error":{"details":{"error_code":"credits_required","model":"claude-fable-5"}}}`)

	require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, headers, body, "claude-fable-5"))
	require.Zero(t, repo.rateLimitCalls)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, anthropicFableRateLimitKey, repo.lastModelRateLimitScope)
	require.Equal(t, resetAt, repo.lastModelRateLimitReset)
}

func TestHandleUpstreamError_AnthropicNonFableCreditsRequiredKeepsAccountLimit(t *testing.T) {
	resetAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-reset", strconv.FormatInt(resetAt.Unix(), 10))
	repo := &anthropicWindowLimitRepo{}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)
	account := &Account{ID: 42, Type: AccountTypeOAuth, Platform: PlatformAnthropic}
	body := []byte(`{"error":{"details":{"error_code":"credits_required","model":"claude-opus-5"}}}`)

	svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, headers, body, "claude-opus-5")
	require.Zero(t, repo.modelRateLimitCalls)
	require.Equal(t, 1, repo.rateLimitCalls)
	require.Equal(t, resetAt, repo.lastRateLimitReset)
}

func TestHandleUpstreamError_AnthropicAccountWindowStillWinsOver7dOi(t *testing.T) {
	now := time.Now()
	reset5h := now.Add(2 * time.Hour).Truncate(time.Second)
	resetOI := now.Add(80 * time.Hour).Truncate(time.Second)
	headers := fable429Headers(reset5h, resetOI)
	headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.0")

	repo := &anthropicWindowLimitRepo{}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)
	account := &Account{ID: 42, Type: AccountTypeOAuth, Platform: PlatformAnthropic}
	svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, headers, nil, "claude-fable-5")

	require.Equal(t, 1, repo.rateLimitCalls)
	require.Equal(t, reset5h, repo.lastRateLimitReset)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, anthropicFableRateLimitKey, repo.lastModelRateLimitScope)
}
