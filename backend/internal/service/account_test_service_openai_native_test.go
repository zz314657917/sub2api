package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type accountNativeCloseTracker struct {
	io.Reader
	closed int
}

func (r *accountNativeCloseTracker) Close() error {
	r.closed++
	return nil
}

func accountNativeContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/1/test", nil)
	return c, rec
}

func TestAccountTestService_OpenAIImageNativeMappedModelMIMEAndHeaders(t *testing.T) {
	for _, tt := range []struct {
		name       string
		payload    string
		wantMIME   string
		wantPrompt string
	}{
		{"item format", `{"output_format":"webp","data":[{"b64_json":"aGVsbG8=","output_format":"png","revised_prompt":"item prompt"}]}`, "image/png", "item prompt"},
		{"root format", `{"output_format":"webp","data":[{"b64_json":"aGVsbG8=","revised_prompt":"root prompt"}]}`, "image/webp", "root prompt"},
		{"default format", `{"data":[{"b64_json":"aGVsbG8=","revised_prompt":"default prompt"}]}`, "image/png", "default prompt"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := accountNativeContext()
			body := &accountNativeCloseTracker{Reader: strings.NewReader(tt.payload)}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: body}}
			proxyID := int64(9)
			account := &Account{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ProxyID: &proxyID, Proxy: &Proxy{Protocol: "http", Host: "proxy.example", Port: 8080}, Credentials: map[string]any{
				"access_token":       "token-value",
				"user_agent":         "native-test-agent",
				"chatgpt_account_id": "chatgpt-account",
				"model_mapping":      map[string]any{"gpt-image-2": "gpt-image-2.5-flare"},
			}}

			err := (&AccountTestService{httpUpstream: upstream}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "keep this prompt")
			require.NoError(t, err)
			require.Equal(t, "/backend-api/codex/images/generations", upstream.lastReq.URL.Path)
			require.Equal(t, "application/json", upstream.lastReq.Header.Get("Accept"))
			require.Equal(t, "Bearer token-value", upstream.lastReq.Header.Get("Authorization"))
			require.True(t, strings.HasPrefix(upstream.lastReq.Header.Get("User-Agent"), "codex_cli_rs/"))
			require.Equal(t, "chatgpt-account", upstream.lastReq.Header.Get("chatgpt-account-id"))
			require.Equal(t, "http://proxy.example:8080", upstream.lastProxyURL)
			require.Equal(t, "gpt-image-2.5-flare", gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, "keep this prompt", gjson.GetBytes(upstream.lastBody, "prompt").String())
			require.Equal(t, 1, body.closed)
			events := parseOpenAIImageTestSSEEvents(rec.Body.String())
			require.Len(t, events, 5)
			require.Equal(t, "test_start", gjson.Get(events[0].Data, "type").String())
			require.Equal(t, "gpt-image-2", gjson.Get(events[0].Data, "model").String())
			require.Equal(t, "content", gjson.Get(events[1].Data, "type").String())
			require.Equal(t, "content", gjson.Get(events[2].Data, "type").String())
			require.Contains(t, events[2].Data, tt.wantPrompt)
			require.Equal(t, "image", gjson.Get(events[3].Data, "type").String())
			require.Equal(t, "data:"+tt.wantMIME+";base64,aGVsbG8=", gjson.Get(events[3].Data, "image_url").String())
			require.Equal(t, "test_complete", gjson.Get(events[4].Data, "type").String())
			require.True(t, gjson.Get(events[4].Data, "success").Bool())
		})
	}
}

func TestAccountTestService_OpenAIImageNativeFallbackClosesBodiesOnce(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			c, rec := accountNativeContext()
			firstBody := &accountNativeCloseTracker{Reader: strings.NewReader(`{"error":{"message":"missing native endpoint"}}`)}
			secondBody := &accountNativeCloseTracker{Reader: strings.NewReader("data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"image_generation_call\",\"result\":\"aGVsbG8=\"}}\n\ndata: [DONE]\n\n")}
			upstream := &httpUpstreamRecorder{responses: []*http.Response{{StatusCode: status, Header: http.Header{}, Body: firstBody}, {StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: secondBody}}}
			account := &Account{ID: 8, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "token"}}
			err := (&AccountTestService{httpUpstream: upstream}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "fallback")
			require.NoError(t, err)
			require.Len(t, upstream.requests, 2)
			require.Equal(t, "/backend-api/codex/images/generations", upstream.requests[0].URL.Path)
			require.Equal(t, "/backend-api/codex/responses", upstream.requests[1].URL.Path)
			require.Equal(t, 1, firstBody.closed)
			require.Equal(t, 1, secondBody.closed)
			events := parseOpenAIImageTestSSEEvents(rec.Body.String())
			require.Len(t, events, 4)
			require.Equal(t, "test_complete", gjson.Get(events[3].Data, "type").String())
		})
	}
}

func TestAccountTestService_OpenAIImageNativeErrorDoesNotComplete(t *testing.T) {
	for _, payload := range []string{`{"data":[]}`, `{"data":[{"b64_json":"aGVsbG8="}]`} {
		c, rec := accountNativeContext()
		body := &accountNativeCloseTracker{Reader: bytes.NewBufferString(payload)}
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: body}}
		account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "token"}}
		err := (&AccountTestService{httpUpstream: upstream}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "empty")
		require.Error(t, err)
		require.Equal(t, 1, body.closed)
		require.NotContains(t, rec.Body.String(), `"success":true`)
	}
}

func TestAccountTestService_OpenAIImageNativeDoesNotReplaySecondFallbackOrOtherFailures(t *testing.T) {
	for _, tt := range []struct {
		name     string
		statuses []int
		err      error
		requests int
	}{
		{"second fallback", []int{http.StatusNotFound, http.StatusMethodNotAllowed}, nil, 2},
		{"other HTTP", []int{http.StatusInternalServerError}, nil, 1},
		{"transport", nil, io.ErrUnexpectedEOF, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := accountNativeContext()
			bodies := make([]*accountNativeCloseTracker, 0, len(tt.statuses))
			responses := make([]*http.Response, 0, len(tt.statuses))
			for _, status := range tt.statuses {
				body := &accountNativeCloseTracker{Reader: strings.NewReader(`{"error":{"message":"server failure"}}`)}
				bodies = append(bodies, body)
				responses = append(responses, &http.Response{StatusCode: status, Body: body})
			}
			upstream := &httpUpstreamRecorder{responses: responses, err: tt.err}
			account := &Account{ID: 10, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "token"}}
			err := (&AccountTestService{httpUpstream: upstream}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "failure")
			require.Error(t, err)
			require.Len(t, upstream.requests, tt.requests)
			for _, body := range bodies {
				require.Equal(t, 1, body.closed)
			}
			require.NotContains(t, rec.Body.String(), `"success":true`)
		})
	}
}

func TestAccountTestService_OpenAIImageNativePreservesSetupTokenShadowAndAgentIdentityAuth(t *testing.T) {
	t.Run("setup token uses default user agent", func(t *testing.T) {
		c, _ := accountNativeContext()
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"aGVsbG8="}]}`))}}
		account := &Account{ID: 11, Platform: PlatformOpenAI, Type: AccountTypeSetupToken, Credentials: map[string]any{"access_token": "setup-token"}}
		require.NoError(t, (&AccountTestService{httpUpstream: upstream}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "setup"))
		require.True(t, strings.HasPrefix(upstream.lastReq.Header.Get("User-Agent"), "codex_cli_rs/"))
		require.Equal(t, "Bearer setup-token", upstream.lastReq.Header.Get("Authorization"))
	})

	t.Run("shadow resolves parent credentials", func(t *testing.T) {
		c, _ := accountNativeContext()
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"aGVsbG8="}]}`))}}
		parent := &Account{ID: 12, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "parent-token", "chatgpt_account_id": "parent-chatgpt"}}
		parentID := parent.ID
		shadow := &Account{ID: 13, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: &parentID, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2.5-flare"}}}
		repo := &agentIdentityCredentialsRepo{account: parent}
		require.NoError(t, (&AccountTestService{httpUpstream: upstream, accountRepo: repo}).testOpenAIImageOAuth(c, context.Background(), shadow, "gpt-image-2", "shadow"))
		require.Equal(t, "Bearer parent-token", upstream.lastReq.Header.Get("Authorization"))
		require.Equal(t, "parent-chatgpt", upstream.lastReq.Header.Get("chatgpt-account-id"))
	})

	t.Run("agent identity uses assertion without bearer token", func(t *testing.T) {
		c, _ := accountNativeContext()
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"aGVsbG8="}]}`))}}
		key, privateKey := newTestAgentIdentityKey(t)
		account := &Account{ID: 14, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"auth_mode": OpenAIAuthModeAgentIdentity, "agent_runtime_id": key.runtimeID, "agent_private_key": privateKey, "task_id": "existing-task"}}
		repo := &agentIdentityCredentialsRepo{account: account}
		svc := &AccountTestService{httpUpstream: upstream, accountRepo: repo, agentIdentityTaskMu: sync.Mutex{}}
		require.NoError(t, svc.testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "identity"))
		require.NotEmpty(t, upstream.lastReq.Header.Get("Authorization"))
		require.NotEqual(t, "Bearer token", upstream.lastReq.Header.Get("Authorization"))
		require.NotContains(t, upstream.lastReq.Header.Get("Authorization"), "agent_private_key")
	})

	t.Run("agent identity redacts a sensitive error body", func(t *testing.T) {
		c, rec := accountNativeContext()
		key, privateKey := newTestAgentIdentityKey(t)
		body := &accountNativeCloseTracker{Reader: strings.NewReader(`{"error":{"message":"` + privateKey + `"}}`)}
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusInternalServerError, Header: http.Header{}, Body: body}}
		account := &Account{ID: 15, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"auth_mode": OpenAIAuthModeAgentIdentity, "agent_runtime_id": key.runtimeID, "agent_private_key": privateKey, "task_id": "existing-task"}}
		repo := &agentIdentityCredentialsRepo{account: account}
		err := (&AccountTestService{httpUpstream: upstream, accountRepo: repo}).testOpenAIImageOAuth(c, context.Background(), account, "gpt-image-2", "identity error")
		require.Error(t, err)
		require.Equal(t, 1, body.closed)
		require.NotContains(t, err.Error(), privateKey)
		require.NotContains(t, rec.Body.String(), privateKey)
		require.NotContains(t, rec.Body.String(), `"success":true`)
	})
}
