package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type s283SparkModelRateLimitRepo struct {
	stubOpenAIAccountRepo
	modelRateLimitCalls int
	modelRateLimitKey   string
	modelRateLimitUntil time.Time
	rateLimitCalls      int
}

func (r *s283SparkModelRateLimitRepo) SetModelRateLimit(_ context.Context, _ int64, scope string, resetAt time.Time) error {
	r.modelRateLimitCalls++
	r.modelRateLimitKey = scope
	r.modelRateLimitUntil = resetAt
	return nil
}

func (r *s283SparkModelRateLimitRepo) SetRateLimited(_ context.Context, _ int64, _ time.Time) error {
	r.rateLimitCalls++
	return nil
}

func s283Spark429Headers(resetSeconds, windowMinutes string) http.Header {
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", resetSeconds)
	headers.Set("x-codex-primary-window-minutes", windowMinutes)
	headers.Set("x-codex-secondary-used-percent", "20")
	headers.Set("x-codex-secondary-reset-after-seconds", "3600")
	headers.Set("x-codex-secondary-window-minutes", "300")
	return headers
}

func TestS283SparkOAuth429IsModelScoped(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2831, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests,
		s283Spark429Headers("18000", "300"), nil, "gpt-5.3-codex-spark-high",
	)

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Greater(t, time.Until(repo.modelRateLimitUntil), 4*time.Hour)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS283SparkOAuth429HonorsAccountModelMapping(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{
		ID:       2832,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"spark-alias": "gpt-5.3-codex-spark"},
		},
	}

	svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests,
		s283Spark429Headers("604800", "10080"), nil, "spark-alias",
	)

	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Zero(t, repo.rateLimitCalls)
}

func TestS283SparkOAuth429FallbackUsesModelCooldown(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2837, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests, nil,
		[]byte(`{"error":{"type":"rate_limit_error"}}`), "gpt-5.3-codex-spark",
	)

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.InDelta(t, 5, time.Until(repo.modelRateLimitUntil).Seconds(), 1)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS283SparkOAuth429BodyResetUsesModelCooldown(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2840, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests, nil,
		[]byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":60}}`), "gpt-5.3-codex-spark",
	)

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Greater(t, time.Until(repo.modelRateLimitUntil), 45*time.Second)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS283SparkModelPathRejectsNonOAuth(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	headers := s283Spark429Headers("18000", "300")

	apiKey := &Account{ID: 2838, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	setupToken := &Account{ID: 2839, Platform: PlatformOpenAI, Type: AccountTypeSetupToken}
	require.False(t, rateLimits.HandleOpenAICodexSparkRateLimit(context.Background(), apiKey, "gpt-5.3-codex-spark", http.StatusTooManyRequests, headers, nil))
	require.False(t, rateLimits.HandleOpenAICodexSparkRateLimit(context.Background(), setupToken, "gpt-5.3-codex-spark", http.StatusTooManyRequests, headers, nil))
	require.Zero(t, repo.modelRateLimitCalls)
}

func TestS283SparkShadow429StaysModelScoped(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	parentID := int64(2833)
	shadow := &Account{
		ID:              2834,
		Platform:        PlatformOpenAI,
		Type:            AccountTypeOAuth,
		ParentAccountID: &parentID,
		QuotaDimension:  QuotaDimensionSpark,
	}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), shadow, http.StatusTooManyRequests,
		s283Spark429Headers("604800", "10080"), nil, "gpt-5.3-codex-spark",
	)

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(shadow))
}

func TestS283NonSparkOAuth429RetainsS282RetryWindow(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2835, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests,
		http.Header{}, []byte(`{"error":{"type":"rate_limit_error"}}`), "gpt-5.3-codex",
	)

	require.False(t, shouldDisable)
	require.Zero(t, repo.modelRateLimitCalls)
	require.Zero(t, repo.rateLimitCalls)
	require.True(t, svc.ShouldRetryOpenAIOAuth429(account, nil, nil))
}

func TestS283SparkPersistWSRateLimitSignalPassesModel(t *testing.T) {
	repo := &s283SparkModelRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2836, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	svc.persistOpenAIWSRateLimitSignal(
		context.Background(), account, s283Spark429Headers("604800", "10080"), nil,
		"rate_limit_exceeded", "rate_limit_error", "usage limit reached", "gpt-5.3-codex-spark",
	)

	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Zero(t, repo.rateLimitCalls)
}

func TestS283SparkModelRateLimitLookupNormalizesAliases(t *testing.T) {
	resetAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	account := &Account{
		Platform: PlatformOpenAI,
		Extra: map[string]any{
			modelRateLimitsKey: map[string]any{
				"gpt-5.3-codex-spark": map[string]any{"rate_limit_reset_at": resetAt},
			},
		},
	}

	require.True(t, account.isModelRateLimitedWithContext(context.Background(), "gpt-5.3-codex-spark-high"))
	require.Greater(t, account.GetModelRateLimitRemainingTimeWithContext(context.Background(), "gpt-5.3-codex-spark-high"), time.Minute)
}
