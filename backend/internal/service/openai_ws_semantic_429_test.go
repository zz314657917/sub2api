package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type s284WSRateLimitRepo struct {
	s282OAuth429Repo
	modelRateLimitCalls int
	modelRateLimitKey   string
	modelRateLimitUntil time.Time
}

func (r *s284WSRateLimitRepo) SetModelRateLimit(_ context.Context, _ int64, scope string, resetAt time.Time) error {
	r.modelRateLimitCalls++
	r.modelRateLimitKey = scope
	r.modelRateLimitUntil = resetAt
	return nil
}

func TestS284WSSemantic429OrdinaryModelIgnoresHandshakeQuotaHeaders(t *testing.T) {
	repo := &s284WSRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2841, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := s283Spark429Headers("604800", "10080")
	payload := []byte(`{"type":"error","error":{"type":"rate_limit_error","code":"rate_limit_exceeded"}}`)

	svc.persistOpenAIWSRateLimitSignal(
		context.Background(), account, headers, payload,
		"rate_limit_exceeded", "rate_limit_error", "quota exhausted", "gpt-5.3-codex",
	)

	require.Nil(t, openAIWSSemantic429Headers(account, "gpt-5.3-codex", headers))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.Zero(t, repo.rateLimitCalls)
	require.Zero(t, repo.modelRateLimitCalls)
}

func TestS284WSSemantic429SparkOAuthUsesHandshakeQuotaHeaders(t *testing.T) {
	repo := &s284WSRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2842, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := s283Spark429Headers("604800", "10080")
	payload := []byte(`{"type":"error","error":{"type":"rate_limit_error","code":"rate_limit_exceeded"}}`)

	svc.persistOpenAIWSRateLimitSignal(
		context.Background(), account, headers, payload,
		"rate_limit_exceeded", "rate_limit_error", "quota exhausted", "gpt-5.3-codex-spark",
	)

	require.Equal(t, headers, openAIWSSemantic429Headers(account, "gpt-5.3-codex-spark", headers))
	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Greater(t, time.Until(repo.modelRateLimitUntil), 6*24*time.Hour)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS284WSHandshake429RetainsQuotaHeaders(t *testing.T) {
	repo := &s284WSRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2843, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := s283Spark429Headers("604800", "10080")

	svc.persistOpenAIWSRateLimitSignal(
		context.Background(), account, headers, nil,
		"rate_limit_exceeded", "rate_limit_error", "handshake quota exhausted", "gpt-5.3-codex",
	)

	// Handshake 429 uses responseBody=nil, so persistOpenAIWSRateLimitSignal
	// retains the original headers; the isolation helper itself only preserves
	// headers for Spark OAuth semantic events.
	require.Nil(t, openAIWSSemantic429Headers(account, "gpt-5.3-codex", headers))
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.Greater(t, repo.tempBlockCalls, 0)
}

func TestS284WSTerminalSemantic429UsesModelIsolation(t *testing.T) {
	repo := &s284WSRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2844, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := s283Spark429Headers("604800", "10080")
	// A real response.failed may expose only the nested numeric status.
	payload := []byte(`{"type":"response.failed","response":{"error":{"status_code":429}}}`)

	svc.handleOpenAIWSTerminalTransientFailure(context.Background(), account, "gpt-5.3-codex-spark", headers, payload)

	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS284WSTerminalSemantic429NestedCodeUsesModelIsolation(t *testing.T) {
	repo := &s284WSRateLimitRepo{}
	rateLimits := NewRateLimitService(repo, nil, nil, nil, nil)
	svc := &OpenAIGatewayService{rateLimitService: rateLimits}
	rateLimits.SetAccountRuntimeBlocker(svc)
	account := &Account{ID: 2846, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers := s283Spark429Headers("604800", "10080")
	payload := []byte(`{"type":"response.failed","response":{"error":{"code":"rate_limit_exceeded","message":"quota exhausted"}}}`)

	svc.handleOpenAIWSTerminalTransientFailure(context.Background(), account, "gpt-5.3-codex-spark", headers, payload)

	require.Equal(t, 1, repo.modelRateLimitCalls)
	require.Equal(t, "gpt-5.3-codex-spark", repo.modelRateLimitKey)
	require.Zero(t, repo.rateLimitCalls)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestS284WSSemantic429APIKeyDropsHandshakeHeaders(t *testing.T) {
	account := &Account{ID: 2845, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	headers := s283Spark429Headers("604800", "10080")
	require.Nil(t, openAIWSSemantic429Headers(account, "gpt-5.3-codex-spark", headers))
}
