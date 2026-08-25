package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCodexAccountIdentityNamespaceScopesBodyAndSession(t *testing.T) {
	first := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"chatgpt_account_id": "acct-first", "chatgpt_user_id": "user-first"}}
	second := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"chatgpt_account_id": "acct-second", "chatgpt_user_id": "user-second"}}
	require.NotEmpty(t, codexAccountIdentityNamespace(first))
	require.NotEqual(t, isolateOpenAIUpstreamSessionID(7, first, "client-session"), isolateOpenAIUpstreamSessionID(7, second, "client-session"))
	require.Equal(t, isolateOpenAIUpstreamSessionID(7, first, "client-session"), isolateOpenAIUpstreamSessionID(7, first, "client-session"))

	body := []byte(`{"prompt_cache_key":"client-session","client_metadata":{"session_id":"client-session","thread_id":"thread","x-codex-turn-metadata":"{\"session_id\":\"client-session\"}"}}`)
	rewritten, changed, err := applyCodexAccountIdentityClientMetadataRaw(body, first, 7)
	require.NoError(t, err)
	require.True(t, changed)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(rewritten, &decoded))
	metadata := decoded["client_metadata"].(map[string]any)
	require.Equal(t, metadata["session_id"], decoded["prompt_cache_key"], "matching prompt cache and session use one namespace")
	require.NotEqual(t, "client-session", metadata["session_id"])
}

func TestCodexAccountIdentityKeepsAPIKeySessionIsolation(t *testing.T) {
	apiKey := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	require.Equal(t, isolateOpenAISessionID(9, "session"), isolateOpenAIUpstreamSessionID(9, apiKey, "session"))
}

func TestCodexAccountIdentityRawPassthroughBuildsNamespacedOutboundBodyAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "account-a", "chatgpt_user_id": "user-a"}}
	body := []byte(`{"prompt_cache_key":"client-session","client_metadata":{"session_id":"client-session","thread_id":"thread-a"}}`)
	scoped, changed, err := applyCodexAccountIdentityClientMetadataRaw(body, account, 5)
	require.NoError(t, err)
	require.True(t, changed)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("session_id", "client-session")
	c.Request.Header.Set("conversation_id", "client-session")
	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, scoped, "test-token")
	require.NoError(t, err)
	got, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, scopeCodexAccountIdentityValue(account, 5, "session", "client-session"), gjson.GetBytes(got, "client_metadata.session_id").String())
	require.Equal(t, gjson.GetBytes(got, "client_metadata.session_id").String(), gjson.GetBytes(got, "prompt_cache_key").String())
	require.Equal(t, isolateOpenAIUpstreamSessionID(0, account, "client-session"), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAIUpstreamSessionID(0, account, "client-session"), req.Header.Get("conversation_id"))
}

func TestCodexAccountIdentityRawPassthroughUsesOriginalPromptCacheKeyWithoutHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{
		"access_token": "test-token", "chatgpt_account_id": "account-a", "chatgpt_user_id": "user-a",
	}}
	original := "client-session"
	body := []byte(`{"prompt_cache_key":"client-session","client_metadata":{"session_id":"client-session"}}`)
	scoped, changed, err := applyCodexAccountIdentityClientMetadataRaw(body, account, 5)
	require.NoError(t, err)
	require.True(t, changed)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	stageCodexAccountIdentityOriginalPromptCacheKey(c, original)

	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, scoped, "test-token")
	require.NoError(t, err)
	require.Equal(t, isolateOpenAIUpstreamSessionID(0, account, original), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAIUpstreamSessionID(0, account, original), req.Header.Get("conversation_id"))
	require.NotEqual(t, isolateOpenAIUpstreamSessionID(0, account, scopeCodexAccountIdentityValue(account, 5, "session", original)), req.Header.Get("session_id"))
}

func TestCodexAccountIdentityShadowHTTPUsesParentNamespaceAndClearsGinSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	parent := newSparkShadowOAuthAccount(7001, "test-parent-token", "parent-account")
	parent.Credentials["chatgpt_user_id"] = "parent-user"
	child := newSparkShadowOAuthAccount(7002, "", "")
	child.ParentAccountID = sparkShadowAccountID(parent.ID)
	child.Credentials = nil
	child.Extra = map[string]any{"openai_passthrough": true}
	next := newSparkShadowOAuthAccount(7003, "test-next-token", "next-account")
	next.Credentials["chatgpt_user_id"] = "next-user"
	next.Extra = map[string]any{"openai_passthrough": true}
	upstream := &httpUpstreamRecorder{responses: []*http.Response{newSparkShadowHTTPResponse(), newSparkShadowHTTPResponse()}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream, accountRepo: stubOpenAIAccountRepo{accounts: []Account{parent, next}}}
	body := []byte(`{"model":"gpt-5.1","stream":false,"prompt_cache_key":"client-session","client_metadata":{"session_id":"client-session","thread_id":"client-thread"},"input":[{"type":"input_text","text":"hello"}]}`)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	require.NoError(t, func() error { _, err := svc.Forward(context.Background(), c, &child, body); return err }())
	require.NoError(t, func() error { _, err := svc.Forward(context.Background(), c, &next, body); return err }())
	require.Len(t, upstream.bodies, 2)
	parentScoped := scopeCodexAccountIdentityValue(&parent, 0, "session", "client-session")
	nextScoped := scopeCodexAccountIdentityValue(&next, 0, "session", "client-session")
	require.Equal(t, parentScoped, gjson.GetBytes(upstream.bodies[0], "client_metadata.session_id").String())
	require.Equal(t, nextScoped, gjson.GetBytes(upstream.bodies[1], "client_metadata.session_id").String())
	require.NotEqual(t, parentScoped, nextScoped)
	require.NotEmpty(t, upstream.requests[0].Header.Get("session_id"))
	require.NotEmpty(t, upstream.requests[1].Header.Get("session_id"))
	require.NotEqual(t, upstream.requests[0].Header.Get("session_id"), upstream.requests[1].Header.Get("session_id"))
}

func TestCodexAccountIdentityChatAndMessagesCaptureOutboundBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{
		"access_token": "test-oauth-token", "chatgpt_account_id": "chat-account", "chatgpt_user_id": "chat-user",
	}}
	chatBody := []byte(`{"model":"gpt-5.4","stream":false,"prompt_cache_key":"chat-session","client_metadata":{"session_id":"chat-session","thread_id":"chat-thread"},"messages":[{"role":"user","content":"hello"}]}`)
	messagesBody := []byte(`{"model":"claude-sonnet-4-5","max_tokens":16,"stream":false,"prompt_cache_key":"messages-session","client_metadata":{"session_id":"messages-session","thread_id":"messages-thread"},"messages":[{"role":"user","content":"hello"}]}`)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		openAICompatSSECompletedResponse("resp-chat-identity", "gpt-5.4"),
		openAICompatSSECompletedResponse("resp-messages-identity", "gpt-5.4"),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, httpUpstream: upstream}
	chatCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	chatCtx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(chatBody))
	_, err := svc.ForwardAsChatCompletions(context.Background(), chatCtx, account, chatBody, "chat-session", "gpt-5.4")
	require.NoError(t, err)
	messagesCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	messagesCtx.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(messagesBody))
	_, err = svc.ForwardAsAnthropic(context.Background(), messagesCtx, account, messagesBody, "messages-session", "gpt-5.4")
	require.NoError(t, err)
	require.Len(t, upstream.bodies, 2)
	require.Equal(t, generateSessionUUID(isolateOpenAIUpstreamSessionID(0, account, "chat-session")), upstream.requests[0].Header.Get("session_id"))
	require.Equal(t, generateSessionUUID(isolateOpenAIUpstreamSessionID(0, account, "messages-session")), upstream.requests[1].Header.Get("session_id"))
}
