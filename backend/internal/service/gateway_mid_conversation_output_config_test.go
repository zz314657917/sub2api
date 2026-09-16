package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSanitizeAnthropicBodyForBetaTokens_MidConversationOutputConfig(t *testing.T) {
	body := []byte(`{"output_config":{"effort":"high"},"messages":[{"role":"system","content":[],"output_config":{"effort":"high"}},{"role":"system","content":"keep","output_config":{"effort":"high"}},{"role":"assistant","content":[{"type":"image"}],"output_config":{"effort":"high"}},{"role":"user","content":"hello"}]}`)
	out, changed := sanitizeAnthropicBodyForBetaTokens(body, claude.BetaOAuth)
	require.True(t, changed)
	require.Equal(t, "high", gjson.GetBytes(out, "output_config.effort").String())
	msgs := gjson.GetBytes(out, "messages").Array()
	require.Len(t, msgs, 3)
	require.Equal(t, "system", msgs[0].Get("role").String())
	require.Equal(t, "keep", msgs[0].Get("content").String())
	require.False(t, msgs[0].Get("output_config").Exists())
	require.Equal(t, "assistant", msgs[1].Get("role").String())
	require.Equal(t, "user", msgs[2].Get("role").String())

	for _, content := range []string{"null", `""`, "[]", `[{"type":"text","text":""}]`} {
		candidate := []byte(`{"messages":[{"role":"system","content":` + content + `,"output_config":{}},{"role":"user","content":"ok"}]}`)
		sanitized, didChange := sanitizeAnthropicBodyForBetaTokens(candidate, claude.BetaOAuth)
		require.True(t, didChange)
		require.Len(t, gjson.GetBytes(sanitized, "messages").Array(), 1)
	}
	nonStandardText := []byte(`{"messages":[{"role":"system","content":[{"type":"text","text":0}],"output_config":{}},{"role":"user","content":"ok"}]}`)
	out, changed = sanitizeAnthropicBodyForBetaTokens(nonStandardText, claude.BetaOAuth)
	require.True(t, changed)
	require.Len(t, gjson.GetBytes(out, "messages").Array(), 2, "non-string text payload must be retained conservatively")
	require.False(t, gjson.GetBytes(out, "messages.0.output_config").Exists())

	noField := []byte(`{"output_config":{"effort":"high"},"messages":[{"role":"system","content":[]}]}`)
	out, changed = sanitizeAnthropicBodyForBetaTokens(noField, claude.BetaOAuth)
	require.False(t, changed)
	require.True(t, bytes.Equal(noField, out))
	out, changed = sanitizeAnthropicBodyForBetaTokens(body, claude.BetaOAuth+","+claude.BetaMidConversationOutputConfig)
	require.False(t, changed)
	require.True(t, bytes.Equal(body, out))
	malformed := []byte(`{"messages":[{"role":"system","output_config":{}}`)
	out, changed = sanitizeAnthropicBodyForBetaTokens(malformed, claude.BetaOAuth)
	require.False(t, changed)
	require.True(t, bytes.Equal(malformed, out))
	malformedWithCompleteMessages := []byte(`{"messages":[{"role":"system","content":[],"output_config":{}}],"truncated":`)
	out, changed = sanitizeAnthropicBodyForBetaTokens(malformedWithCompleteMessages, claude.BetaOAuth)
	require.False(t, changed)
	require.True(t, bytes.Equal(malformedWithCompleteMessages, out))

	legacy := []byte(`{"context_management":{},"thinking":{"block_binding":{}},"fallbacks":[],"fallback_credit_token":"x","messages":[{"role":"user","content":"ok"}]}`)
	out, changed = sanitizeAnthropicBodyForBetaTokens(legacy, claude.BetaOAuth)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(out, "context_management").Exists())
	require.False(t, gjson.GetBytes(out, "thinking.block_binding").Exists())
	require.False(t, gjson.GetBytes(out, "fallbacks").Exists())
	require.False(t, gjson.GetBytes(out, "fallback_credit_token").Exists())
}

func TestBuildUpstreamRequestOAuthMimic_MidConversationOutputConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		drop bool
		keep bool
	}{{"default", false, true}, {"policy_drop", true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c := newMidConversationContext(t, "/v1/messages", "")
			if tc.drop {
				c.Set(betaPolicyFilterSetKey, map[string]struct{}{claude.BetaMidConversationOutputConfig: {}})
			}
			req, err := (&GatewayService{cfg: &config.Config{}}).buildUpstreamRequest(context.Background(), c, newMidConversationOAuthAccount(), midConversationBody(), "token", "oauth", "claude-opus-5", false, true)
			require.NoError(t, err)
			assertMidConversationWireBody(t, readMidConversationBody(t, req), getHeaderRaw(req.Header, "anthropic-beta"), tc.keep)
		})
	}
}

func TestBuildUpstreamRequestAnthropicAPIKeyPassthrough_MidConversationOutputConfig(t *testing.T) {
	for _, tc := range []struct {
		name, beta string
		keep       bool
	}{{"client_has_beta", claude.BetaMidConversationOutputConfig, true}, {"client_missing_beta", claude.BetaOAuth, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c := newMidConversationContext(t, "/v1/messages", tc.beta)
			req, err := (&GatewayService{cfg: &config.Config{}}).buildUpstreamRequestAnthropicAPIKeyPassthrough(context.Background(), c, newMidConversationAPIKeyAccount(), midConversationBody(), "token")
			require.NoError(t, err)
			assertMidConversationWireBody(t, readMidConversationBody(t, req), getHeaderRaw(req.Header, "anthropic-beta"), tc.keep)
		})
	}
}

func TestBuildCountTokensRequestOAuthMimic_MidConversationOutputConfig(t *testing.T) {
	for _, tc := range []struct {
		name       string
		drop, keep bool
	}{{"default", false, true}, {"policy_drop", true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c := newMidConversationContext(t, "/v1/messages/count_tokens", "")
			if tc.drop {
				c.Set(betaPolicyFilterSetKey, map[string]struct{}{claude.BetaMidConversationOutputConfig: {}})
			}
			req, err := (&GatewayService{cfg: &config.Config{}}).buildCountTokensRequest(context.Background(), c, newMidConversationOAuthAccount(), midConversationBody(), "token", "oauth", "claude-opus-5", true)
			require.NoError(t, err)
			assertMidConversationWireBody(t, readMidConversationBody(t, req), getHeaderRaw(req.Header, "anthropic-beta"), tc.keep)
		})
	}
}

func TestBuildCountTokensRequestAnthropicAPIKeyPassthrough_MidConversationOutputConfig(t *testing.T) {
	for _, tc := range []struct {
		name, beta string
		keep       bool
	}{{"client_has_beta", claude.BetaMidConversationOutputConfig, true}, {"client_missing_beta", claude.BetaOAuth, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c := newMidConversationContext(t, "/v1/messages/count_tokens", tc.beta)
			req, err := (&GatewayService{cfg: &config.Config{}}).buildCountTokensRequestAnthropicAPIKeyPassthrough(context.Background(), c, newMidConversationAPIKeyAccount(), midConversationBody(), "token")
			require.NoError(t, err)
			assertMidConversationWireBody(t, readMidConversationBody(t, req), getHeaderRaw(req.Header, "anthropic-beta"), tc.keep)
		})
	}
}

func TestBuildUpstreamRequestOAuthMimicEnabledCCH_MidConversationOutputConfig(t *testing.T) {
	previous, _ := gatewayForwardingCache.Load().(*cachedGatewayForwardingSettings)
	t.Cleanup(func() {
		if previous != nil {
			gatewayForwardingCache.Store(previous)
			return
		}
		gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{expiresAt: time.Now().Add(-time.Second).UnixNano()})
	})
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	c := newMidConversationContext(t, "/v1/messages", "")
	c.Set(betaPolicyFilterSetKey, map[string]struct{}{claude.BetaMidConversationOutputConfig: {}})
	svc := &GatewayService{cfg: &config.Config{}, settingService: NewSettingService(&gatewayTTLSettingRepo{data: map[string]string{SettingKeyEnableCCHSigning: "true"}}, &config.Config{})}
	body := []byte(`{"model":"claude-opus-5","system":[{"type":"text","text":"x-anthropic-billing-header: cc_version=2.1.258; cch=00000;"}],"output_config":{"effort":"high"},"messages":[{"role":"system","content":[],"output_config":{"effort":"high"}},{"role":"user","content":"hello"}]}`)
	req, err := svc.buildUpstreamRequest(context.Background(), c, newMidConversationOAuthAccount(), body, "token", "oauth", "claude-opus-5", false, true)
	require.NoError(t, err)
	out := readMidConversationBody(t, req)
	assertMidConversationWireBody(t, out, getHeaderRaw(req.Header, "anthropic-beta"), false)
	re := regexp.MustCompile(`cch=[0-9a-f]{5}`)
	require.True(t, re.Match(out))
	restored := re.ReplaceAll(out, []byte("cch=00000"))
	require.True(t, bytes.Equal(out, signBillingHeaderCCH(restored)), "CCH must sign the final sanitized body")
}

func newMidConversationContext(t *testing.T, path, beta string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, path, nil)
	if beta != "" {
		c.Request.Header.Set("Anthropic-Beta", beta)
	}
	return c
}

func newMidConversationOAuthAccount() *Account {
	return &Account{ID: 971, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "token"}, Status: StatusActive, Schedulable: true}
}
func newMidConversationAPIKeyAccount() *Account {
	return &Account{ID: 972, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "token"}, Status: StatusActive, Schedulable: true}
}
func midConversationBody() []byte {
	return []byte(`{"model":"claude-opus-5","output_config":{"effort":"high"},"messages":[{"role":"system","content":[],"output_config":{"effort":"high"}},{"role":"user","content":"hello"}]}`)
}

func readMidConversationBody(t *testing.T, req *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	return body
}
func assertMidConversationWireBody(t *testing.T, body []byte, beta string, keep bool) {
	t.Helper()
	require.Equal(t, keep, anthropicBetaTokensContains(beta, claude.BetaMidConversationOutputConfig))
	require.Equal(t, "high", gjson.GetBytes(body, "output_config.effort").String())
	msgs := gjson.GetBytes(body, "messages").Array()
	want := 1
	if keep {
		want = 2
	}
	require.Len(t, msgs, want)
	if keep {
		require.True(t, msgs[0].Get("output_config").Exists())
	}
	require.Equal(t, "user", msgs[len(msgs)-1].Get("role").String())
	require.Equal(t, "hello", msgs[len(msgs)-1].Get("content").String())
}
