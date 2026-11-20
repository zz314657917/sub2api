package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGroupFast20261103HTTPBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, ingress := range []string{"native-apikey", "native-oauth", "passthrough-oauth", "chat", "messages"} {
		for _, tier := range []string{"", "default", "flex", "ultrafast", "unknown"} {
			for _, action := range []string{BetaPolicyActionPass, BetaPolicyActionFilter, BetaPolicyActionBlock} {
				t.Run(ingress+"/"+tier+"/"+action, func(t *testing.T) {
					ctx := groupFast20261103Context(PlatformComposite, true)
					body := map[string]any{"model": "gpt-5.5", "stream": false, "instructions": "test", "input": "hello"}
					if ingress == "chat" || ingress == "messages" {
						body["messages"] = []any{map[string]any{"role": "user", "content": "hello"}}
						body["max_tokens"] = 10
						delete(body, "input")
					}
					if tier != "" {
						body["service_tier"] = tier
					}
					raw, err := json.Marshal(body)
					require.NoError(t, err)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(raw)).WithContext(ctx)
					c.Request.Header.Set("Content-Type", "application/json")
					c.Request.Header.Set("User-Agent", "codex_cli_rs/0.144.1")
					upstream := &httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
						Body: io.NopCloser(strings.NewReader("event: response.completed\ndata: " +
							`{"type":"response.completed","response":{"id":"resp_group_fast","object":"response","model":"gpt-5.5","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}` + "\n\n")),
					}}
					svc := newOpenAIGatewayServiceWithSettings(t, &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{
						ServiceTier: OpenAIFastTierPriority, Action: action, Scope: BetaPolicyScopeAll,
					}}})
					svc.cfg = &config.Config{}
					svc.httpUpstream = upstream
					account := &Account{
						ID: 123, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1,
						Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1),
						Credentials: map[string]any{"access_token": "test", "chatgpt_account_id": "test", "api_key": "test"},
						Extra:       map[string]any{"openai_passthrough": false, "openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff},
					}
					if ingress == "native-apikey" {
						account.Type = AccountTypeAPIKey
					}
					if ingress == "passthrough-oauth" {
						account.Extra["openai_passthrough"] = true
					}
					var result *OpenAIForwardResult
					switch ingress {
					case "chat":
						result, err = svc.ForwardAsChatCompletions(ctx, c, account, raw, "", "")
					case "messages":
						result, err = svc.ForwardAsAnthropic(ctx, c, account, raw, "", "")
					default:
						result, err = svc.Forward(ctx, c, account, raw)
					}
					if action == BetaPolicyActionBlock {
						var blocked *OpenAIFastBlockedError
						require.True(t, errors.As(err, &blocked), "error=%v", err)
						require.Nil(t, upstream.lastReq, "blocked request must never reach transport")
						require.Equal(t, http.StatusForbidden, rec.Code)
						return
					}
					require.NoError(t, err)
					require.NotNil(t, result)
					require.NotNil(t, upstream.lastReq)
					require.Equal(t, "gpt-5.5", gjson.GetBytes(upstream.lastBody, "model").String())
					if action == BetaPolicyActionFilter {
						require.False(t, gjson.GetBytes(upstream.lastBody, "service_tier").Exists())
						require.Nil(t, result.ServiceTier)
					} else {
						require.Equal(t, "priority", gjson.GetBytes(upstream.lastBody, "service_tier").String())
						require.NotNil(t, result.ServiceTier)
						require.Equal(t, "priority", *result.ServiceTier)
					}
				})
			}
		}
	}
}

func TestGroupUltrafastHTTPPassthroughWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, ingress := range []struct {
		name, clientPath, upstreamPath string
		chatOnly                       bool
	}{
		{"native", "/v1/responses", "/v1/responses", false},
		{"chat", "/v1/chat/completions", "/v1/responses", false},
		{"messages", "/v1/messages", "/v1/responses", false},
		{"raw-chat", "/v1/chat/completions", "/v1/chat/completions", true},
		{"responses-to-raw-chat", "/v1/responses", "/v1/chat/completions", true},
	} {
		for _, tc := range []struct {
			name, requested, source string
			filter, reject          bool
		}{
			{"flag-off", "gpt-5.5", "gpt-5.5", false, false},
			{"flag-on", "gpt-5.5", "gpt-5.5", true, false},
			{"wildcard-source", "gpt-5.5", "gpt-5.*", true, false},
			{"empty-mapping", "gpt-5.5", "", true, false},
			{"mapping-target-is-not-a-source", "gpt-6.1-sol", "gpt-5.5", true, true},
			{"allowed-sol", "gpt-6.1-sol", "gpt-6.1-sol", true, false},
		} {
			t.Run(ingress.name+"/"+tc.name, func(t *testing.T) {
				ctx := groupFast20261103Context(PlatformComposite, true)
				ctx.Value(ctxkey.Group).(*Group).ForceOpenAIUltrafast = true
				body := map[string]any{"model": tc.requested, "stream": false}
				if ingress.clientPath == "/v1/responses" {
					body["instructions"], body["input"] = "test", "hello"
				} else {
					body["messages"] = []map[string]any{{"role": "user", "content": "hello"}}
					body["max_tokens"] = 10
				}
				raw, err := json.Marshal(body)
				require.NoError(t, err)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, ingress.clientPath, bytes.NewReader(raw)).WithContext(ctx)
				c.Request.Header.Set("Content-Type", "application/json")

				mapping := map[string]any{}
				if tc.source != "" {
					mapping[tc.source] = "gpt-6.1-sol"
				}
				account := &Account{
					ID: 123, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1,
					Status: StatusActive, Schedulable: true,
					Credentials: map[string]any{"api_key": "test", "model_mapping": mapping},
					Extra:       map[string]any{"openai_passthrough": true, "openai_passthrough_model_filter": tc.filter},
				}
				wantModel := tc.requested
				if !tc.filter && ingress.name != "native" {
					wantModel = "gpt-6.1-sol"
				}
				wantTier := "priority"
				if wantModel == "gpt-6.1-sol" {
					wantTier = "ultrafast"
				}
				contentType := "text/event-stream"
				upstreamBody := `event: response.completed` + "\ndata: " +
					`{"type":"response.completed","response":{"id":"resp_qa","object":"response","model":"` + wantModel + `","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}}` + "\n\n"
				if ingress.chatOnly {
					account.Extra["openai_responses_mode"] = "force_chat_completions"
					contentType = "application/json"
					upstreamBody = `{"id":"chatcmpl-qa","object":"chat.completion","model":"` + wantModel + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{contentType}},
					Body:       io.NopCloser(strings.NewReader(upstreamBody)),
				}}
				svc := newOpenAIGatewayServiceWithSettings(t, DefaultOpenAIFastPolicySettings())
				svc.cfg = &config.Config{}
				svc.httpUpstream = upstream
				var result *OpenAIForwardResult
				switch ingress.clientPath {
				case "/v1/chat/completions":
					result, err = svc.ForwardAsChatCompletions(ctx, c, account, raw, "", "")
				case "/v1/messages":
					result, err = svc.ForwardAsAnthropic(ctx, c, account, raw, "", "")
				default:
					result, err = svc.Forward(ctx, c, account, raw)
				}
				if tc.reject {
					require.ErrorContains(t, err, "model is not supported by the selected account")
					require.Equal(t, http.StatusBadRequest, rec.Code)
					require.Nil(t, result)
					require.Nil(t, upstream.lastReq, "an already selected account must still reject a target-only model")
					return
				}
				require.NoError(t, err)
				require.NotNil(t, upstream.lastReq)
				require.Equal(t, ingress.upstreamPath, upstream.lastReq.URL.Path)
				require.Equal(t, wantModel, gjson.GetBytes(upstream.lastBody, "model").String())
				require.Equal(t, wantTier, gjson.GetBytes(upstream.lastBody, "service_tier").String())
				require.NotNil(t, result)
				if ingress.name == "native" {
					// Native passthrough only records an explicit override when needed.
					require.Empty(t, result.UpstreamModel)
				} else {
					require.Equal(t, wantModel, result.UpstreamModel)
				}
				require.NotNil(t, result.ServiceTier)
				require.Equal(t, wantTier, *result.ServiceTier)
			})
		}
	}
}

func TestResponsesRawChatFallbackAuthorizesOutboundModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, model, wantModel, wantError string
		filter                            bool
	}{
		{"allowed-normalized-model", "  gpt-5.5  ", "gpt-5.5", "", true},
		{"unlisted-model", "gpt-5.4", "", "model is not supported", true},
		{"blank-model", "  ", "", "single non-empty string", true},
		{"filter-disabled-retains-mapping", "gpt-5.5", "gpt-6.1-sol", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"model": tc.model, "input": "hello"})
			require.NoError(t, err)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			account := &Account{
				ID: 123, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test", "model_mapping": map[string]any{"gpt-5.5": "gpt-6.1-sol"}},
				Extra: map[string]any{
					"openai_passthrough": true, "openai_passthrough_model_filter": tc.filter,
					"openai_responses_mode": "force_chat_completions",
				},
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"chatcmpl-test","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			var result *OpenAIForwardResult
			if tc.name == "blank-model" {
				result, err = svc.Forward(c.Request.Context(), c, account, body)
			} else {
				// Exercise the final authorization at the fallback boundary rather
				// than relying on Forward's earlier eligibility check.
				result, err = svc.forwardResponsesViaRawChatCompletions(c.Request.Context(), c, account, body)
			}
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				require.Equal(t, http.StatusBadRequest, rec.Code)
				require.Nil(t, result)
				require.Nil(t, upstream.lastReq)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, upstream.lastReq)
			require.Equal(t, tc.wantModel, result.UpstreamModel)
			require.Equal(t, tc.wantModel, gjson.GetBytes(upstream.lastBody, "model").String())
		})
	}
}

func TestGroupFast20261103HTTPDisabledAndUntrusted(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "body-spoof-without-context"}[enabled], func(t *testing.T) {
			ctx := context.Background()
			if !enabled {
				ctx = groupFast20261103Context(PlatformOpenAI, false)
			}
			raw := []byte(`{"model":"gpt-5.5","stream":false,"input":"hello","group_id":7,"force_openai_fast":true}`)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(raw)).WithContext(ctx)
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200,
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"id":"test","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`))}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			_, err := svc.Forward(ctx, c, &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test"}, Extra: map[string]any{"openai_passthrough": false}}, raw)
			require.NoError(t, err)
			require.NotNil(t, upstream.lastReq)
			require.False(t, gjson.GetBytes(upstream.lastBody, "service_tier").Exists())
		})
	}
}
