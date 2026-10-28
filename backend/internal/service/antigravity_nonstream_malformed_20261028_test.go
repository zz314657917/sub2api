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

func TestUpstreamAFixesAntigravityNonstreamMalformedWriter(t *testing.T) {
	tool := `{"candidates":[{"content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}`
	malformed := `{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":4}}`
	for _, wrapped := range []bool{false, true} {
		for _, tc := range []struct {
			name   string
			frames []string
		}{
			{"signature only", []string{`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"thoughtSignature":"sig"}]}}]}`}},
			{"failure after tool without parts", []string{tool, malformed}},
			{"failure before usage", []string{tool, malformed, `{"usageMetadata":{"promptTokenCount":12}}`}},
			{"failure before later success", []string{malformed, tool}},
		} {
			name := tc.name
			if wrapped {
				name = "wrapped/" + name
			} else {
				name = "direct/" + name
			}
			t.Run(name, func(t *testing.T) {
				var frames strings.Builder
				for _, frame := range tc.frames {
					if wrapped {
						frame = `{"response":` + frame + `}`
					}
					frames.WriteString("data: " + frame + "\n\n")
				}
				gin.SetMode(gin.TestMode)
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				svc := newAntigravityTestService(&config.Config{})
				resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(frames.String()))}
				result, err := svc.handleClaudeStreamToNonStreaming(c, resp, time.Now(), "test-model")
				require.Error(t, err)
				require.Nil(t, result, "failed collection must not supply success usage for billing")
				require.Equal(t, http.StatusBadGateway, recorder.Code)
				require.Contains(t, recorder.Body.String(), `"type":"malformed_function_call"`)
				require.Contains(t, recorder.Body.String(), "MALFORMED_FUNCTION_CALL")
				require.NotContains(t, recorder.Body.String(), `"type":"message"`)
				require.True(t, IsResponseCommitted(c))
				var failover *UpstreamFailoverError
				require.NotErrorAs(t, err, &failover, "committed error must not trigger retry")
			})
		}
	}
}

func TestUpstreamAFixesAntigravityForwardMalformedHasNoBillableResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":16,"stream":false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	frames := `data: {"response":{"candidates":[{"content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}}` + "\n\n" +
		`data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":4}}}` + "\n\n"
	upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(frames)),
	}}}
	svc := &AntigravityGatewayService{
		settingService: NewSettingService(&antigravitySettingRepoStub{}, &config.Config{}),
		tokenProvider:  &AntigravityTokenProvider{}, httpUpstream: upstream,
	}
	account := &Account{ID: 7807, Platform: PlatformAntigravity, Type: AccountTypeOAuth,
		Status: StatusActive, Concurrency: 1, Credentials: map[string]any{"access_token": "test-token"}}
	result, err := svc.Forward(context.Background(), c, account, body, false)
	require.Error(t, err)
	require.Nil(t, result, "billing callers must not receive a successful ForwardResult")
	require.Equal(t, 1, upstream.callCount)
	require.True(t, IsResponseCommitted(c))
	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"type":"malformed_function_call"`)
}

func TestUpstreamAFixesAntigravityNormalCollectorAndTrueStream(t *testing.T) {
	for _, part := range []string{`{"text":"hello"}`, `{"functionCall":{"name":"read","args":{}}}`} {
		for _, stream := range []bool{false, true} {
			name := part
			if stream {
				name = "stream/" + name
			}
			t.Run(name, func(t *testing.T) {
				gin.SetMode(gin.TestMode)
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				body := `data: {"response":{"candidates":[{"finishReason":"STOP","content":{"parts":[` + part + `]}}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":4}}}` + "\n\n"
				svc := newAntigravityTestService(&config.Config{})
				resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}
				handle := svc.handleClaudeStreamToNonStreaming
				if stream {
					handle = svc.handleClaudeStreamingResponse
				}
				result, err := handle(c, resp, time.Now(), "test-model")
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.usage)
				require.Equal(t, http.StatusOK, recorder.Code)
				require.NotContains(t, recorder.Body.String(), "malformed_function_call")
				if stream {
					require.Contains(t, recorder.Body.String(), "message_stop")
				}
			})
		}
	}
}
