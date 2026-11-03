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
