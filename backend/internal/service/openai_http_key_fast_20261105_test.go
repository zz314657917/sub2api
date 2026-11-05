package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestKeyFast20261105HTTPIngressAndFinalMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name, policy, tier, action, want string
		group                            bool
	}{
		{"force-missing", "force", "", "pass", "priority", false},
		{"force-flex", "force", "flex", "pass", "priority", false},
		{"force-priority", "force", "priority", "pass", "priority", false},
		{"force-ultrafast", "force", "ultrafast", "pass", "priority", false},
		{"force-unknown", "force", "unknown", "pass", "priority", false},
		{"filter-priority", "filter", "priority", "pass", "", false},
		{"filter-alias", "filter", "  FaSt  ", "pass", "", false},
		{"filter-flex", "filter", "flex", "pass", "flex", false},
		{"filter-ultrafast", "filter", "ultrafast", "pass", "ultrafast", false},
		{"pass-flex", "passthrough", "flex", "pass", "flex", false},
		{"group-beats-filter", "filter", "flex", "pass", "priority", true},
		{"global-removes-force", "force", "flex", "filter", "", false},
		{"global-removes-priority", "force", "priority", "filter", "", false},
		{"global-removes-group", "filter", "priority", "filter", "", true},
		{"global-blocks-force", "force", "", "block", "", false},
		{"global-blocks-group", "filter", "flex", "block", "", true},
	}
	for _, ingress := range []string{"native-apikey", "native-oauth", "passthrough-oauth", "chat", "messages", "raw-chat", "responses-raw-chat"} {
		for _, stream := range []bool{false, true} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s/stream=%v/%s", ingress, stream, tc.name), func(t *testing.T) {
					ctx := keyFast20261105Context(tc.policy, tc.group)
					input := map[string]any{"model": "gpt-5.5", "stream": stream, "instructions": "test", "input": "hello"}
					if ingress == "chat" || ingress == "messages" || ingress == "raw-chat" {
						input["messages"] = []any{map[string]any{"role": "user", "content": "hello"}}
						input["max_tokens"] = 10
						delete(input, "input")
					}
					if tc.tier != "" {
						input["service_tier"] = tc.tier
					}
					raw, err := json.Marshal(input)
					require.NoError(t, err)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(raw)).WithContext(ctx)
					c.Request.Header.Set("Content-Type", "application/json")
					c.Request.Header.Set("User-Agent", "codex_cli_rs/0.144.1")
					if ingress == "messages" && (strings.EqualFold(strings.TrimSpace(tc.tier), "fast") || strings.EqualFold(strings.TrimSpace(tc.tier), "priority")) {
						c.Request.Header.Set("anthropic-beta", claude.BetaFastMode)
					}
					chatWire := ingress == "raw-chat" || ingress == "responses-raw-chat"
					upstream := &httpUpstreamRecorder{resp: keyFast20261105UpstreamResponse(chatWire, stream)}
					svc := newOpenAIGatewayServiceWithSettings(t, &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{
						ServiceTier: OpenAIFastTierPriority, Action: tc.action, Scope: BetaPolicyScopeAll,
					}}})
					svc.cfg = &config.Config{}
					svc.cfg.Security.URLAllowlist.AllowInsecureHTTP = true
					svc.httpUpstream = upstream
					account := &Account{ID: 123, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
						Concurrency: 1, Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1),
						Credentials: map[string]any{"access_token": "test", "chatgpt_account_id": "test", "api_key": "test"},
						Extra:       map[string]any{"openai_passthrough": false, "openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff}}
					if ingress == "native-apikey" || chatWire {
						account.Type = AccountTypeAPIKey
					}
					if ingress == "passthrough-oauth" {
						account.Extra["openai_passthrough"] = true
					}
					if chatWire {
						account.Extra[openai_compat.ExtraKeyResponsesMode] = string(openai_compat.ResponsesSupportModeForceChatCompletions)
					}
					var result *OpenAIForwardResult
					switch ingress {
					case "chat", "raw-chat":
						result, err = svc.ForwardAsChatCompletions(ctx, c, account, raw, "", "")
					case "messages":
						result, err = svc.ForwardAsAnthropic(ctx, c, account, raw, "", "")
					default:
						result, err = svc.Forward(ctx, c, account, raw)
					}
					if tc.action == "block" {
						var blocked *OpenAIFastBlockedError
						require.True(t, errors.As(err, &blocked), "error=%v", err)
						require.Nil(t, upstream.lastReq)
						require.Equal(t, http.StatusForbidden, rec.Code)
						return
					}
					require.NoError(t, err)
					require.NotNil(t, upstream.lastReq)
					require.NotNil(t, result)
					want := tc.want
					// Messages accepts Fast through anthropic-beta. Its existing
					// protocol converter ignores a body-level flex/ultrafast tier.
					if ingress == "messages" && (want == "flex" || want == "ultrafast") {
						want = ""
					}
					require.Equal(t, want, gjson.GetBytes(upstream.lastBody, "service_tier").String())
					if want == "" {
						require.False(t, gjson.GetBytes(upstream.lastBody, "service_tier").Exists())
						require.Nil(t, result.ServiceTier, "removed tier must not leak into billing metadata")
					} else {
						require.NotNil(t, result.ServiceTier)
						require.Equal(t, want, *result.ServiceTier, "metadata must match final transport")
					}
					require.Equal(t, 1, result.Usage.InputTokens)
					require.Equal(t, 1, result.Usage.OutputTokens)
					if account.IsOpenAIOAuth() {
						hint := "model=gpt-5.5"
						if want != "" {
							hint += ";tier=" + want
						}
						require.Equal(t, hint, upstream.lastReq.Header.Get(openAICodexRoutingHintHeader))
					}
					if chatWire {
						require.Equal(t, "/v1/chat/completions", upstream.lastReq.URL.Path)
					}
				})
			}
		}
	}
}

func keyFast20261105UpstreamResponse(chat, stream bool) *http.Response {
	var body, contentType string
	if chat {
		if stream {
			contentType = "text/event-stream"
			body = "data: " + `{"id":"chat_test","object":"chat.completion.chunk","model":"gpt-5.5","choices":[{"index":0,"delta":{"content":"hello"},"finish_reason":null}]}` + "\n\n" +
				"data: " + `{"id":"chat_test","object":"chat.completion.chunk","model":"gpt-5.5","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}` + "\n\ndata: [DONE]\n\n"
		} else {
			contentType = "application/json"
			body = `{"id":"chat_test","model":"gpt-5.5","choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`
		}
	} else {
		// OAuth Responses uses SSE upstream even for a nonstreaming client.
		contentType = "text/event-stream"
		body = "event: response.completed\ndata: " + `{"type":"response.completed","response":{"id":"resp_test","object":"response","model":"gpt-5.5","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}` + "\n\n"
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(body))}
}
