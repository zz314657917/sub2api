//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type openAIQuotaAccountRepoStub struct {
	account *Account
	err     error
}

func (r *openAIQuotaAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.account, nil
}

type openAIQuotaProxyRepoStub struct {
	proxy *Proxy
	err   error
	calls int32
}

func (r *openAIQuotaProxyRepoStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	atomic.AddInt32(&r.calls, 1)
	if r.err != nil {
		return nil, r.err
	}
	return r.proxy, nil
}

func withOpenAIQuotaTestURLs(t *testing.T, serverURL string) {
	t.Helper()
	oldUsageURL := chatGPTUsageURL
	oldResetURL := chatGPTRateLimitResetURL
	chatGPTUsageURL = serverURL + "/backend-api/wham/usage"
	chatGPTRateLimitResetURL = serverURL + "/backend-api/wham/rate-limit-reset-credits/consume"
	t.Cleanup(func() {
		chatGPTUsageURL = oldUsageURL
		chatGPTRateLimitResetURL = oldResetURL
	})
}

func newOpenAIQuotaTestAccount() *Account {
	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	return &Account{
		ID:       42,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "access-token",
			"expires_at":         expiresAt,
			"chatgpt_account_id": "acct_123",
		},
	}
}

func newOpenAIQuotaTestService(account *Account, proxyRepo openAIQuotaProxyReader, factory PrivacyClientFactory) *OpenAIQuotaService {
	tokenProvider := NewOpenAITokenProvider(nil, newOpenAITokenCacheStub(), nil)
	return newOpenAIQuotaService(&openAIQuotaAccountRepoStub{account: account}, proxyRepo, tokenProvider, factory)
}

func TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy(t *testing.T) {
	account := newOpenAIQuotaTestAccount()
	proxyID := int64(7)
	account.ProxyID = &proxyID
	account.Proxy = &Proxy{
		Protocol: "http",
		Host:     "proxy.local",
		Port:     8080,
		Username: "u",
		Password: "p",
	}
	proxyRepo := &openAIQuotaProxyRepoStub{}

	var gotAuth string
	var gotAccountID string
	var gotOriginator string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/wham/usage", r.URL.Path)
		gotAuth = r.Header.Get("Authorization")
		gotAccountID = r.Header.Get("Chatgpt-Account-Id")
		gotOriginator = r.Header.Get("Originator")
		require.Equal(t, "zh-CN", r.Header.Get("Oai-Language"))
		require.Equal(t, "none", r.Header.Get("Sec-Fetch-Site"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"user_id":"user_1",
			"account_id":"acct_123",
			"email":"owner@example.com",
			"plan_type":"plus",
			"rate_limit_reset_credits":{"available_count":2}
		}`))
	}))
	defer server.Close()
	withOpenAIQuotaTestURLs(t, server.URL)

	var gotProxyURL string
	svc := newOpenAIQuotaTestService(account, proxyRepo, func(proxyURL string) (*req.Client, error) {
		gotProxyURL = proxyURL
		return req.C().SetTimeout(time.Second), nil
	})

	usage, err := svc.QueryUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, "user_1", usage.UserID)
	require.Equal(t, "plus", usage.PlanType)
	require.NotNil(t, usage.RateLimitResetCredits)
	require.Equal(t, 2, usage.RateLimitResetCredits.AvailableCount)
	require.Greater(t, usage.FetchedAt, int64(0))
	require.Equal(t, "Bearer access-token", gotAuth)
	require.Equal(t, "acct_123", gotAccountID)
	require.Equal(t, "Codex Desktop", gotOriginator)
	require.Equal(t, "http://u:p@proxy.local:8080", gotProxyURL)
	require.Equal(t, int32(0), atomic.LoadInt32(&proxyRepo.calls), "eager-loaded proxy should avoid repository fallback")
}

func TestOpenAIQuotaResetCreditPostsRedeemRequestIDAndFallsBackToProxyRepo(t *testing.T) {
	account := newOpenAIQuotaTestAccount()
	delete(account.Credentials, "chatgpt_account_id")
	account.Credentials["organization_id"] = "org_456"
	proxyID := int64(8)
	account.ProxyID = &proxyID
	proxyRepo := &openAIQuotaProxyRepoStub{
		proxy: &Proxy{
			Protocol: "socks5",
			Host:     "127.0.0.1",
			Port:     1080,
		},
	}

	var body map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/wham/rate-limit-reset-credits/consume", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))
		require.Equal(t, "org_456", r.Header.Get("Chatgpt-Account-Id"))
		require.Contains(t, r.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"success","windows_reset":2,"credit":{"id":"credit_1","status":"redeemed"}}`))
	}))
	defer server.Close()
	withOpenAIQuotaTestURLs(t, server.URL)

	var gotProxyURL string
	svc := newOpenAIQuotaTestService(account, proxyRepo, func(proxyURL string) (*req.Client, error) {
		gotProxyURL = proxyURL
		return req.C().SetTimeout(time.Second), nil
	})

	result, err := svc.ResetCredit(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "success", result.Code)
	require.Equal(t, 2, result.WindowsReset)
	require.NotNil(t, result.Credit)
	require.Equal(t, "credit_1", result.Credit.ID)
	require.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, body["redeem_request_id"])
	require.Equal(t, "socks5://127.0.0.1:1080", gotProxyURL)
	require.Equal(t, int32(1), atomic.LoadInt32(&proxyRepo.calls))
}

func TestOpenAIQuotaValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		account *Account
		want    string
	}{
		{
			name: "non openai",
			account: &Account{
				ID:       1,
				Platform: PlatformAnthropic,
				Type:     AccountTypeOAuth,
			},
			want: "OPENAI_QUOTA_INVALID_PLATFORM",
		},
		{
			name: "non oauth",
			account: &Account{
				ID:       2,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
			},
			want: "OPENAI_QUOTA_INVALID_TYPE",
		},
		{
			name: "missing account id",
			account: &Account{
				ID:       3,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token",
				},
			},
			want: "OPENAI_QUOTA_MISSING_ACCOUNT_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newOpenAIQuotaTestService(tt.account, nil, func(proxyURL string) (*req.Client, error) {
				return req.C(), nil
			})
			_, err := svc.QueryUsage(context.Background(), tt.account.ID)
			require.Error(t, err)
			require.Equal(t, tt.want, infraerrors.Reason(err))
		})
	}
}

func TestOpenAIQuotaUpstreamStatusMapping(t *testing.T) {
	account := newOpenAIQuotaTestAccount()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("x", 300), http.StatusUnauthorized)
	}))
	defer server.Close()
	withOpenAIQuotaTestURLs(t, server.URL)

	svc := newOpenAIQuotaTestService(account, nil, func(proxyURL string) (*req.Client, error) {
		return req.C().SetTimeout(time.Second), nil
	})

	_, err := svc.QueryUsage(context.Background(), account.ID)
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, infraerrors.Code(err))
	require.Equal(t, "OPENAI_QUOTA_UPSTREAM_ERROR", infraerrors.Reason(err))
	require.Contains(t, infraerrors.Message(err), "upstream returned 401")
}

func TestOpenAIQuotaServiceNotConfigured(t *testing.T) {
	svc := &OpenAIQuotaService{}
	_, err := svc.QueryUsage(context.Background(), 1)
	require.Error(t, err)
	require.Equal(t, "OPENAI_QUOTA_NOT_CONFIGURED", infraerrors.Reason(err))
	require.Equal(t, http.StatusInternalServerError, infraerrors.Code(err))
}

func TestOpenAIQuotaAccountNotFound(t *testing.T) {
	tokenProvider := NewOpenAITokenProvider(nil, nil, nil)
	svc := newOpenAIQuotaService(
		&openAIQuotaAccountRepoStub{err: errors.New("not found")},
		nil,
		tokenProvider,
		func(proxyURL string) (*req.Client, error) { return req.C(), nil },
	)

	_, err := svc.QueryUsage(context.Background(), 404)
	require.Error(t, err)
	require.Equal(t, "OPENAI_QUOTA_ACCOUNT_NOT_FOUND", infraerrors.Reason(err))
	require.Equal(t, http.StatusNotFound, infraerrors.Code(err))
}
