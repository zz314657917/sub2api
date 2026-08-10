package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type s207MissingWebSearchSettingRepo struct {
	SettingRepository
}

func (s *s207MissingWebSearchSettingRepo) GetValue(context.Context, string) (string, error) {
	return "", ErrSettingNotFound
}

func TestLoadWebSearchConfigFromDB_MissingSetting(t *testing.T) {
	webSearchEmulationCache.Store(&cachedWebSearchEmulationConfig{
		config:    &WebSearchEmulationConfig{},
		expiresAt: 0,
	})
	defer webSearchEmulationCache.Store(&cachedWebSearchEmulationConfig{
		config:    &WebSearchEmulationConfig{},
		expiresAt: 0,
	})

	svc := NewSettingService(&s207MissingWebSearchSettingRepo{}, &config.Config{})
	before := time.Now().Add(webSearchEmulationCacheTTL - time.Second).UnixNano()
	cfg, err := svc.loadWebSearchConfigFromDB()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.False(t, cfg.Enabled)
	require.Empty(t, cfg.Providers)

	cached, ok := webSearchEmulationCache.Load().(*cachedWebSearchEmulationConfig)
	require.True(t, ok)
	require.NotNil(t, cached)
	require.Same(t, cfg, cached.config)
	require.Greater(t, cached.expiresAt, before)
}

func TestGrokOAuthServiceExchangeCodeRejectsMissingClientWithoutConsumingSession(t *testing.T) {
	svc := NewGrokOAuthService(nil, nil)
	defer svc.Stop()

	auth, err := svc.GenerateAuthURL(context.Background(), nil, "")
	require.NoError(t, err)
	_, err = svc.ExchangeCode(context.Background(), &GrokExchangeCodeInput{
		SessionID: auth.SessionID,
		Code:      "valid-code",
		State:     auth.State,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "GROK_OAUTH_CLIENT_NOT_CONFIGURED")
	_, ok := svc.sessionStore.Get(auth.SessionID)
	require.True(t, ok)
}

func TestGrokOAuthServiceRefreshTokenRejectsMissingClient(t *testing.T) {
	svc := NewGrokOAuthService(nil, nil)
	defer svc.Stop()

	_, err := svc.RefreshToken(context.Background(), "refresh-token", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "GROK_OAUTH_CLIENT_NOT_CONFIGURED")
}

type s207GeminiPolicyRepo struct {
	AccountRepository
	tempCalls      int
	rateLimitCalls int
}

func (r *s207GeminiPolicyRepo) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	r.tempCalls++
	return nil
}

func (r *s207GeminiPolicyRepo) SetRateLimited(context.Context, int64, time.Time) error {
	r.rateLimitCalls++
	return nil
}

func TestGeminiChatCompletions_TempUnscheduled429DoesNotSetAccountRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"capacity exhausted"}}`)),
	}}
	repo := &s207GeminiPolicyRepo{}
	svc := &GeminiMessagesCompatService{
		httpUpstream:     upstream,
		accountRepo:      repo,
		rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil),
		cfg:              &config.Config{},
	}
	account := &Account{
		ID: 703, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
		Credentials: map[string]any{
			"api_key":                    "test-key",
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{
				map[string]any{
					"error_code": float64(429), "keywords": []any{"capacity"},
					"duration_minutes": float64(10),
				},
			},
		},
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, 1, repo.tempCalls)
	require.Zero(t, repo.rateLimitCalls)
}

func TestGeminiMessagesCompat_TempUnscheduled429DoesNotSetAccountRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"gemini-2.5-flash","max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))

	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"capacity exhausted"}}`)),
	}}
	repo := &s207GeminiPolicyRepo{}
	svc := &GeminiMessagesCompatService{
		httpUpstream:     upstream,
		accountRepo:      repo,
		rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil),
		cfg:              &config.Config{},
	}
	account := s207GeminiTempUnscheduledAccount(704)

	result, err := svc.Forward(context.Background(), c, account, body)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, 1, repo.tempCalls)
	require.Zero(t, repo.rateLimitCalls)
}

func TestGeminiNative_TempUnscheduled429DoesNotSetAccountRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)
	c.Request = httptest.NewRequest(http.MethodPost,
		"/v1beta/models/gemini-2.5-flash:generateContent", bytes.NewReader(body))

	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"capacity exhausted"}}`)),
	}}
	repo := &s207GeminiPolicyRepo{}
	svc := &GeminiMessagesCompatService{
		httpUpstream:     upstream,
		accountRepo:      repo,
		rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil),
		cfg:              &config.Config{},
	}
	account := s207GeminiTempUnscheduledAccount(705)

	result, err := svc.ForwardNative(context.Background(), c, account,
		"gemini-2.5-flash", "generateContent", false, body)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, 1, repo.tempCalls)
	require.Zero(t, repo.rateLimitCalls)
}

func TestGeminiChatCompletions_PoolMode429DoesNotSetAccountRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"capacity exhausted"}}`)),
	}}
	repo := &s207GeminiPolicyRepo{}
	svc := &GeminiMessagesCompatService{
		httpUpstream:     upstream,
		accountRepo:      repo,
		rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil),
		cfg:              &config.Config{},
	}
	account := &Account{
		ID: 706, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
		Credentials: map[string]any{"api_key": "test-key", "pool_mode": true},
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body)
	require.Error(t, err)
	require.Nil(t, result)
	require.Zero(t, repo.tempCalls)
	require.Zero(t, repo.rateLimitCalls)
}

func s207GeminiTempUnscheduledAccount(id int64) *Account {
	return &Account{
		ID: id, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
		Credentials: map[string]any{
			"api_key":                    "test-key",
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{
				map[string]any{
					"error_code": float64(429), "keywords": []any{"capacity"},
					"duration_minutes": float64(10),
				},
			},
		},
	}
}
