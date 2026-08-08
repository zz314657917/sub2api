package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAICodexRoutingHintCanonicalizesOfficialServiceTiers(t *testing.T) {
	oauthAccount := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	tests := []struct {
		name        string
		model       string
		serviceTier string
		want        string
	}{
		{name: "fast alias", model: "gpt-5.6", serviceTier: " fast ", want: "model=gpt-5.6;tier=priority"},
		{name: "priority", model: "gpt-5.6", serviceTier: "priority", want: "model=gpt-5.6;tier=priority"},
		{name: "flex", model: "gpt-5.6", serviceTier: "flex", want: "model=gpt-5.6;tier=flex"},
		{name: "default", model: "gpt-5.6", serviceTier: "default", want: "model=gpt-5.6"},
		{name: "omitted", model: "gpt-5.6", want: "model=gpt-5.6"},
		{name: "unknown", model: "gpt-5.6", serviceTier: "turbo", want: "model=gpt-5.6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(http.Header)
			setOpenAICodexRoutingHint(headers, oauthAccount, tt.model, tt.serviceTier)
			require.Equal(t, tt.want, headers.Get(openAICodexRoutingHintHeader))
		})
	}

	for _, model := range []string{"gpt-5.6;evil", "gpt=5.6", "gpt-5.6\ninvalid"} {
		t.Run("unsafe model "+model, func(t *testing.T) {
			headers := make(http.Header)
			setOpenAICodexRoutingHint(headers, oauthAccount, model, "priority")
			require.Empty(t, headers.Get(openAICodexRoutingHintHeader))
		})
	}
}

func TestOpenAICodexRoutingHintIsGatewayOwnedAndOAuthOnly(t *testing.T) {
	headers := make(http.Header)
	headers[openAICodexRoutingHintHeader] = []string{"lowercase-spoof"}
	headers["X-Codex-Routing-Hint"] = []string{"canonical-spoof"}
	setOpenAICodexRoutingHint(headers, &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, "gpt-5.6", "priority")
	for key := range headers {
		require.False(t, strings.EqualFold(key, openAICodexRoutingHintHeader))
	}

	headers[openAICodexRoutingHintHeader] = []string{"model=spoof;tier=flex"}
	setOpenAICodexRoutingHint(headers, &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, "gpt-5.6", "priority")
	require.Equal(t, "model=gpt-5.6;tier=priority", headers.Get(openAICodexRoutingHintHeader))
	require.Len(t, headers, 1)
}

func TestOpenAIOAuthHTTPBuildersSendRoutingHintFromFinalBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oauthAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"chatgpt_account_id": "test-account",
		},
	}
	svc := &OpenAIGatewayService{cfg: &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{Enabled: false},
		},
	}}
	body := []byte(`{"model":"gpt-5.6-codex","service_tier":"fast"}`)

	for _, passthrough := range []bool{false, true} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))

		var req *http.Request
		var err error
		if passthrough {
			req, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, oauthAccount, body, "test-token")
		} else {
			req, err = svc.buildUpstreamRequest(context.Background(), c, oauthAccount, body, "test-token", false, "", true)
		}
		require.NoError(t, err)
		require.Equal(t, "model=gpt-5.6-codex;tier=priority", req.Header.Get(openAICodexRoutingHintHeader))
	}
}

func TestOpenAIHTTPBuildersStripOnlyOAuthLegacyResponsesBeta(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{cfg: &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{Enabled: false},
		},
	}}
	body := []byte(`{"model":"gpt-5.6-codex"}`)
	oauth := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	apiKey := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-key"}}

	build := func(t *testing.T, account *Account, passthrough bool) http.Header {
		t.Helper()
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
		c.Request.Header.Add("OpenAI-Beta", "responses=experimental, future_feature=v1")
		var req *http.Request
		var err error
		if passthrough {
			req, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, body, "test-token")
		} else {
			req, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "test-token", false, "", true)
		}
		require.NoError(t, err)
		return req.Header
	}

	for _, passthrough := range []bool{false, true} {
		mode := "normal"
		if passthrough {
			mode = "passthrough"
		}
		t.Run(mode, func(t *testing.T) {
			if passthrough {
				require.Equal(t, []string{"future_feature=v1"}, build(t, oauth, true).Values("OpenAI-Beta"))
				require.Equal(t, []string{"responses=experimental, future_feature=v1"}, build(t, apiKey, true).Values("OpenAI-Beta"))
				return
			}
			require.Empty(t, build(t, oauth, false).Values("OpenAI-Beta"))
			require.Empty(t, build(t, apiKey, false).Values("OpenAI-Beta"))
		})
	}
}

func TestOpenAIWSHeadersSendOAuthRoutingHintOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	svc := &OpenAIGatewayService{}
	decision := OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}

	oauth := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	headers, _ := svc.buildOpenAIWSHeaders(context.Background(), c, oauth, "test-token", decision, true, "", "", "", "gpt-5.6-codex", "fast")
	require.Equal(t, "model=gpt-5.6-codex;tier=priority", headers.Get(openAICodexRoutingHintHeader))

	apiKey := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	headers, _ = svc.buildOpenAIWSHeaders(context.Background(), c, apiKey, "test-token", decision, true, "", "", "", "gpt-5.6-codex", "priority")
	require.Empty(t, headers.Get(openAICodexRoutingHintHeader))
}

func TestOpenAIWSConnPoolRoutingHintIsSoftAffinity(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 4
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 4
	pool := newOpenAIWSConnPool(cfg)
	dialer := &openAIWSCountingDialer{}
	pool.setClientDialerForTest(dialer)
	account := &Account{ID: 913, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	acquire := func(t *testing.T, hint, preferred string, forcePreferred bool) *openAIWSConnLease {
		t.Helper()
		headers := make(http.Header)
		headers.Set(openAICodexRoutingHintHeader, hint)
		lease, err := pool.Acquire(context.Background(), openAIWSAcquireRequest{
			Account:            account,
			WSURL:              "wss://example.com/v1/responses",
			Headers:            headers,
			PreferredConnID:    preferred,
			ForcePreferredConn: forcePreferred,
		})
		require.NoError(t, err)
		return lease
	}

	priority := acquire(t, "model=gpt-5.6-codex;tier=priority", "", false)
	priorityID := priority.ConnID()
	priority.Release()
	priorityAgain := acquire(t, "model=gpt-5.6-codex;tier=priority", "", false)
	require.True(t, priorityAgain.Reused())
	require.Equal(t, priorityID, priorityAgain.ConnID())
	priorityAgain.Release()

	flex := acquire(t, "model=gpt-5.6-codex;tier=flex", "", false)
	require.False(t, flex.Reused())
	flexID := flex.ConnID()
	require.NotEqual(t, priorityID, flexID)
	flex.Release()

	continued := acquire(t, "model=gpt-5.6-codex", flexID, true)
	require.True(t, continued.Reused())
	require.Equal(t, flexID, continued.ConnID())
	continued.Release()
	require.Equal(t, 2, dialer.DialCount())
}

func TestOpenAIWSConnPoolRoutingHintReplacesIdleMismatchButFallsBackWhenBusy(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
	pool := newOpenAIWSConnPool(cfg)
	dialer := &openAIWSCountingDialer{}
	pool.setClientDialerForTest(dialer)
	account := &Account{ID: 915, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	acquire := func(ctx context.Context, hint string) (*openAIWSConnLease, error) {
		headers := make(http.Header)
		headers.Set(openAICodexRoutingHintHeader, hint)
		return pool.Acquire(ctx, openAIWSAcquireRequest{
			Account: account,
			WSURL:   "wss://example.com/v1/responses",
			Headers: headers,
		})
	}

	priority, err := acquire(context.Background(), "model=gpt-5.6-codex;tier=priority")
	require.NoError(t, err)
	priorityID := priority.ConnID()

	type acquireResult struct {
		lease *openAIWSConnLease
		err   error
	}
	resultCh := make(chan acquireResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		lease, acquireErr := acquire(ctx, "model=gpt-5.6-codex;tier=flex")
		resultCh <- acquireResult{lease: lease, err: acquireErr}
	}()
	require.Eventually(t, func() bool {
		_, waiters, _ := pool.AccountPoolLoad(account.ID)
		return waiters == 1
	}, time.Second, 10*time.Millisecond)
	priority.Release()

	busyFallback := <-resultCh
	require.NoError(t, busyFallback.err)
	require.NotNil(t, busyFallback.lease)
	require.True(t, busyFallback.lease.Reused())
	require.Equal(t, priorityID, busyFallback.lease.ConnID())
	busyFallback.lease.Release()
	require.Equal(t, 1, dialer.DialCount())

	flex, err := acquire(context.Background(), "model=gpt-5.6-codex;tier=flex")
	require.NoError(t, err)
	require.False(t, flex.Reused())
	require.NotEqual(t, priorityID, flex.ConnID())
	flex.Release()
	require.Equal(t, 2, dialer.DialCount())
}

func TestOpenAIWSConnPoolGenerationRejectsStaleDial(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 2
	pool := newOpenAIWSConnPool(cfg)
	dialer := &s206GenerationDialer{started: make(chan struct{}), release: make(chan struct{})}
	pool.setClientDialerForTest(dialer)
	account := &Account{ID: 914, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	resultCh := make(chan error, 1)
	go func() {
		lease, err := pool.Acquire(context.Background(), openAIWSAcquireRequest{Account: account, WSURL: "wss://example.com/v1/responses"})
		if lease != nil {
			lease.Release()
		}
		resultCh <- err
	}()
	<-dialer.started
	pool.ClearAccount(account.ID)
	close(dialer.release)
	require.NoError(t, <-resultCh)
	require.Equal(t, int32(2), dialer.count.Load())
}

type s206GenerationDialer struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	count   atomic.Int32
}

func (d *s206GenerationDialer) Dial(context.Context, string, http.Header, string) (openAIWSClientConn, int, http.Header, error) {
	d.count.Add(1)
	d.once.Do(func() { close(d.started) })
	<-d.release
	return &openAIWSFakeConn{}, http.StatusSwitchingProtocols, nil, nil
}
