package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// This transport exercises the actual Forward and response handlers without
// provider credentials, database connections or copied gateway implementations.
type cfportHTTP struct {
	body       []byte
	sse        bool
	calls      int
	reject     string
	earlyUsage bool
}

func (u *cfportHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.calls++
	u.body, _ = io.ReadAll(req.Body)
	if u.calls == 1 && u.reject != "" {
		return &http.Response{StatusCode: 400, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(u.reject))}, nil
	}
	name := gjson.GetBytes(u.body, "tools.0.name").String()
	if name == "" {
		name = "python"
	}
	response := fmt.Sprintf(`{"id":"resp_cf","model":"gpt-5.5","output":[{"type":"function_call","id":"fc_1","call_id":"call_1","name":%q,"arguments":"{}"}],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}`, name)
	contentType := "application/json"
	if u.sse {
		contentType = "text/event-stream"
		if u.earlyUsage {
			response = strings.ReplaceAll(response, `,"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}`, "")
		}
		response = "data: {\"type\":\"response.completed\",\"response\":" + response + "}\n\ndata: [DONE]\n\n"
		if u.earlyUsage {
			response = "data: {\"type\":\"response.created\",\"response\":{\"usage\":{\"input_tokens\":17,\"output_tokens\":4}}}\n\n" + response
		}
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(response))}, nil
}
func (u *cfportHTTP) DoWithTLS(req *http.Request, p string, id int64, n int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, p, id, n)
}

func TestCFPortGatewayRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, oauth := range []bool{false, true} {
		for _, passthrough := range []bool{false, true} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("oauth=%v/pass=%v/stream=%v", oauth, passthrough, stream), func(t *testing.T) {
					u := &cfportHTTP{sse: oauth || stream}
					s := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: u, toolCorrector: NewCodexToolCorrector()}
					a := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "test", "access_token": "test"}, Extra: map[string]any{"openai_passthrough": passthrough, "openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff}}
					if oauth {
						a.Type = AccountTypeOAuth
					}
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
					body := []byte(fmt.Sprintf(`{"model":"gpt-5.5","stream":%v,"instructions":"Test instructions","tools":[{"type":"function","name":"python","parameters":{"type":"object","properties":{"text":{"type":"string","pattern":"(?=a)a"}}}}],"input":[{"type":"reasoning","id":"item_bad","encrypted_content":"opaque","summary":[]},{"role":"user","content":"hello"}],"text":{"format":{"type":"json_schema","name":"result","schema":{"properties":{"answer":{"type":"string"}},"uniqueItems":true}}}}`, stream))
					result, err := s.Forward(context.Background(), c, a, body)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, 1, u.calls)
					require.False(t, gjson.GetBytes(u.body, "input.0.id").Exists())
					require.Equal(t, "opaque", gjson.GetBytes(u.body, "input.0.encrypted_content").String())
					require.False(t, gjson.GetBytes(u.body, "tools.0.parameters.properties.text.pattern").Exists())
					require.False(t, gjson.GetBytes(u.body, "text.format.schema.uniqueItems").Exists())
					if oauth {
						require.Equal(t, "python__sub2api", gjson.GetBytes(u.body, "tools.0.name").String())
					}
					require.Contains(t, w.Body.String(), `"name":"python"`)
					require.NotContains(t, w.Body.String(), "python__sub2api")
				})
			}
		}
	}
}

func TestCFPortGatewayRejectedStatusRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			u := &cfportHTTP{reject: `{"error":{"code":"unknown_parameter","param":"input[0].status","message":"Unknown parameter: input[0].status"}}`}
			s := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: u, toolCorrector: NewCodexToolCorrector()}
			a := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test"}, Extra: map[string]any{"openai_passthrough": passthrough}}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			body := []byte(`{"model":"gpt-5.5","stream":false,"input":[{"type":"message","role":"user","content":"keep me","status":"completed"}]}`)
			_, err := s.Forward(context.Background(), c, a, body)
			require.NoError(t, err)
			require.Equal(t, 2, u.calls)
			require.False(t, gjson.GetBytes(u.body, "input.0.status").Exists())
			require.Equal(t, "keep me", gjson.GetBytes(u.body, "input.0.content").String())
		})
	}
}

func TestCFPortForwardStreamUsageAndSSEToJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("pass=%v/stream=%v", passthrough, stream), func(t *testing.T) {
				u := &cfportHTTP{sse: true, earlyUsage: stream}
				s := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: u, toolCorrector: NewCodexToolCorrector()}
				a := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test"}, Extra: map[string]any{"openai_passthrough": passthrough}}
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
				body := []byte(fmt.Sprintf(`{"model":"gpt-5.5","stream":%v,"input":"hello"}`, stream))
				result, err := s.Forward(context.Background(), c, a, body)
				require.NoError(t, err)
				if stream {
					require.Contains(t, w.Body.String(), "data:")
					require.Equal(t, 17, result.Usage.InputTokens)
					require.Equal(t, 4, result.Usage.OutputTokens)
				} else {
					require.True(t, gjson.Valid(w.Body.String()))
					require.Equal(t, "resp_cf", gjson.Get(w.Body.String(), "id").String())
					require.Equal(t, 3, result.Usage.InputTokens)
				}
			})
		}
	}
}

func TestCFPortEarlyUsageAndAuthoritativeTerminal(t *testing.T) {
	s := &OpenAIGatewayService{}
	u := &OpenAIUsage{}
	s.parseSSEUsageBytes([]byte(`{"type":"response.created","response":{"usage":{"input_tokens":17,"output_tokens":4}}}`), u)
	require.Equal(t, 17, u.InputTokens)
	require.Equal(t, 4, u.OutputTokens)
	s.parseSSEUsageBytes([]byte(`{"type":"response.in_progress","usage":{"input_tokens":0,"output_tokens":5}}`), u)
	require.Equal(t, 17, u.InputTokens)
	require.Equal(t, 5, u.OutputTokens)
	s.parseSSEUsageBytes([]byte(`{"type":"response.completed","response":{"id":"r"}}`), u)
	require.Equal(t, 17, u.InputTokens)
	s.parseSSEUsageBytes([]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":0,"output_tokens":0}}}`), u)
	require.Zero(t, u.InputTokens)
	require.Zero(t, u.OutputTokens)
}

func TestCFPortStructuredErrorsIgnoreEchoedInput(t *testing.T) {
	for _, text := range []string{"context_length_exceeded", "an error occurred while processing your request", "selected model is at capacity"} {
		body := []byte(fmt.Sprintf(`{"error":{"message":"bad field"},"input":%q}`, text))
		require.False(t, isOpenAIContextWindowError("bad field", body))
		require.False(t, isOpenAITransientProcessingError(400, "bad field", body))
	}
	require.True(t, isOpenAIContextWindowError("", []byte("maximum context length exceeded")))
	require.True(t, isOpenAITransientProcessingError(400, "", []byte(`{"response":{"error":{"message":"an error occurred while processing your request"}}}`)))
}
