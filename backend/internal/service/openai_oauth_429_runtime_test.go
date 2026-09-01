package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// s282OAuth429Repo overrides only the repository methods exercised by the
// OpenAI 429 path; the embedded test repository supplies the remaining
// AccountRepository surface without changing production interfaces.
type s282OAuth429Repo struct {
	stubOpenAIAccountRepo
	rateLimitCalls int
	lastRateLimit  time.Time
	tempBlockCalls int
	lastTempBlock  time.Time
	updateExtra    []map[string]any
}

func (r *s282OAuth429Repo) SetRateLimited(_ context.Context, _ int64, resetAt time.Time) error {
	r.rateLimitCalls++
	r.lastRateLimit = resetAt
	return nil
}

func (r *s282OAuth429Repo) SetTempUnschedulable(_ context.Context, _ int64, until time.Time, _ string) error {
	r.tempBlockCalls++
	r.lastTempBlock = until
	return nil
}

func (r *s282OAuth429Repo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	copied := make(map[string]any, len(updates))
	for key, value := range updates {
		copied[key] = value
	}
	r.updateExtra = append(r.updateExtra, copied)
	return nil
}

func TestS282ClassifyOpenAIOAuth429(t *testing.T) {
	sevenDay := http.Header{}
	sevenDay.Set("x-codex-primary-used-percent", "100")
	sevenDay.Set("x-codex-primary-reset-after-seconds", "604800")
	sevenDay.Set("x-codex-primary-window-minutes", "10080")
	sevenDay.Set("x-codex-secondary-used-percent", "20")
	sevenDay.Set("x-codex-secondary-reset-after-seconds", "3600")
	sevenDay.Set("x-codex-secondary-window-minutes", "300")
	disposition, resetAt := classifyOpenAIOAuth429(sevenDay, nil)
	require.Equal(t, openAIOAuth429Quota7d, disposition)
	require.NotNil(t, resetAt)
	require.Greater(t, time.Until(*resetAt), 6*24*time.Hour)

	fiveHour := sevenDay.Clone()
	fiveHour.Set("x-codex-primary-used-percent", "100")
	fiveHour.Set("x-codex-primary-reset-after-seconds", "18000")
	fiveHour.Set("x-codex-primary-window-minutes", "300")
	fiveHour.Set("x-codex-secondary-used-percent", "20")
	fiveHour.Set("x-codex-secondary-reset-after-seconds", "604800")
	fiveHour.Set("x-codex-secondary-window-minutes", "10080")
	disposition, resetAt = classifyOpenAIOAuth429(fiveHour, nil)
	require.Equal(t, openAIOAuth429Quota5h, disposition)
	require.NotNil(t, resetAt)
	require.Greater(t, time.Until(*resetAt), 4*time.Hour)

	bodyReset := []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":60}}`)
	disposition, resetAt = classifyOpenAIOAuth429(nil, bodyReset)
	require.Equal(t, openAIOAuth429QuotaReset, disposition)
	require.NotNil(t, resetAt)

	disposition, resetAt = classifyOpenAIOAuth429(nil, []byte(`{"error":{"type":"rate_limit_error","message":"try again"}}`))
	require.Equal(t, openAIOAuth429Transient, disposition)
	require.Nil(t, resetAt)
}

func TestS282TransientOAuth429SkipsDurableCooldown(t *testing.T) {
	repo := &s282OAuth429Repo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2801, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests, http.Header{},
		[]byte(`{"error":{"type":"rate_limit_error","message":"try again"}}`),
	)

	require.False(t, shouldDisable)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.True(t, svc.shouldRetryOpenAIOAuth429OnSameAccount(account, http.StatusTooManyRequests, false))

	setupToken := &Account{ID: 2802, Platform: PlatformOpenAI, Type: AccountTypeSetupToken}
	require.False(t, svc.shouldRetryOpenAIOAuth429OnSameAccount(setupToken, http.StatusTooManyRequests, false))
}

func TestS282QuotaOAuth429BlocksImmediatelyAndPersistsReset(t *testing.T) {
	repo := &s282OAuth429Repo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2803, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", "18000")
	headers.Set("x-codex-primary-window-minutes", "300")
	headers.Set("x-codex-secondary-used-percent", "20")
	headers.Set("x-codex-secondary-reset-after-seconds", "604800")
	headers.Set("x-codex-secondary-window-minutes", "10080")

	shouldDisable := svc.handleOpenAIAccountUpstreamError(
		context.Background(), account, http.StatusTooManyRequests, headers,
		[]byte(`{"error":{"type":"rate_limit_error","code":"rate_limit_exceeded"}}`),
	)

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.rateLimitCalls)
	require.NotEmpty(t, repo.updateExtra)
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.Greater(t, time.Until(repo.lastRateLimit), 4*time.Hour)
	require.False(t, svc.ShouldRetryOpenAIOAuth429(account, headers, nil))
}

func TestS282SevenDayQuotaOAuth429UsesImmediateRuntimeAndTempBlock(t *testing.T) {
	repo := &s282OAuth429Repo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2804, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", "604800")
	headers.Set("x-codex-primary-window-minutes", "10080")
	headers.Set("x-codex-secondary-used-percent", "20")
	headers.Set("x-codex-secondary-reset-after-seconds", "18000")
	headers.Set("x-codex-secondary-window-minutes", "300")

	svc.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusTooManyRequests, headers, nil)

	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.Equal(t, 1, repo.tempBlockCalls)
	require.Zero(t, repo.rateLimitCalls)
	require.Greater(t, time.Until(repo.lastTempBlock), 6*24*time.Hour)
}

func TestS282BodyResetAndAPIKeyBoundaries(t *testing.T) {
	repo := &s282OAuth429Repo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	oauth := &Account{ID: 2805, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	body := []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":60}}`)
	svc.handleOpenAIAccountUpstreamError(context.Background(), oauth, http.StatusTooManyRequests, nil, body)
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(oauth))
	require.Equal(t, 1, repo.rateLimitCalls)
	require.Greater(t, time.Until(repo.lastRateLimit), 45*time.Second)

	apiKey := &Account{ID: 2806, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc.handleOpenAIAccountUpstreamError(context.Background(), apiKey, http.StatusTooManyRequests, nil, []byte(`{"error":{"type":"rate_limit_error"}}`))
	require.Equal(t, 2, repo.rateLimitCalls)
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(apiKey))
	require.False(t, svc.shouldRetryOpenAIOAuth429OnSameAccount(apiKey, http.StatusTooManyRequests, false))
}

func TestS282SparkShadowDoesNotPersistGlobalCodex429(t *testing.T) {
	repo := &s282OAuth429Repo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	parentID := int64(2810)
	shadow := &Account{
		ID:              2811,
		Platform:        PlatformOpenAI,
		Type:            AccountTypeOAuth,
		ParentAccountID: &parentID,
		QuotaDimension:  QuotaDimensionSpark,
	}
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", "604800")
	headers.Set("x-codex-primary-window-minutes", "10080")

	svc.handleOpenAIAccountUpstreamError(context.Background(), shadow, http.StatusTooManyRequests, headers, nil)

	require.Zero(t, repo.rateLimitCalls)
	require.Zero(t, repo.tempBlockCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(shadow))
}

func TestS282RuntimeBlockKeepsLongestAndClearReenablesScheduling(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 2807, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	longUntil := time.Now().Add(10 * time.Minute)
	svc.BlockAccountScheduling(account, longUntil, "quota")
	svc.BlockAccountScheduling(account, time.Now().Add(time.Minute), "transient")

	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	value, ok := svc.openaiAccountRuntimeBlockUntil.Load(account.ID)
	require.True(t, ok)
	blockedUntil, ok := value.(time.Time)
	require.True(t, ok)
	require.WithinDuration(t, longUntil, blockedUntil, time.Second)

	svc.ClearAccountSchedulingBlock(account.ID)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS282TransientOAuth429FailoverMetadata(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 2808, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	svc.markOpenAIOAuth429RateLimited(context.Background(), account, http.Header{}, nil)
	failoverErr := svc.applyOpenAIOAuth429Retry(account, http.StatusTooManyRequests, false, http.Header{}, nil, &UpstreamFailoverError{})

	require.True(t, failoverErr.RetryableOnSameAccount)
	require.Equal(t, openAIOAuth429MaxAccountAttempts, failoverErr.SameAccountRetryLimit)
	require.Equal(t, openAIOAuth429RetryDelay, failoverErr.SameAccountRetryBackoffBase)
}
