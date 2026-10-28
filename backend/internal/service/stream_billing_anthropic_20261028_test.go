package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const streamBillingAnthropicStart = "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_billing\",\"model\":\"claude-sonnet-4-5\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"usage\":{\"input_tokens\":11,\"cache_read_input_tokens\":7,\"cache_creation_input_tokens\":5,\"cache_creation\":{\"ephemeral_5m_input_tokens\":2,\"ephemeral_1h_input_tokens\":3}}}}\n\n"
const streamBillingAnthropicDelta = "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":null},\"usage\":{\"output_tokens\":9}}\n\n"
const streamBillingAnthropicStop = "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"

type streamBillingCountBody struct {
	io.ReadCloser
	closes atomic.Int32
	reads  atomic.Int32
}

func (b *streamBillingCountBody) Read(p []byte) (int, error) {
	b.reads.Add(1)
	return b.ReadCloser.Read(p)
}

func (b *streamBillingCountBody) Close() error {
	b.closes.Add(1)
	return b.ReadCloser.Close()
}

func streamBillingAnthropicService(body io.ReadCloser) *GatewayService {
	return &GatewayService{
		cfg: &config.Config{},
		httpUpstream: &anthropicHTTPUpstreamRecorder{resp: &http.Response{
			StatusCode: 200, Header: http.Header{"X-Request-Id": {"rid-billing"}, "Content-Type": {"text/event-stream"}}, Body: body,
		}},
		rateLimitService: &RateLimitService{}, deferredService: &DeferredService{},
	}
}

func streamBillingAnthropicForward(t *testing.T, kind string, svc *GatewayService, c *gin.Context, ctx context.Context) (*ForwardResult, error) {
	t.Helper()
	account := newAnthropicAPIKeyAccountForTest()
	switch kind {
	case "chat":
		body := []byte(`{"model":"claude-sonnet-4-5","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
		return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
	case "responses":
		body := []byte(`{"model":"claude-sonnet-4-5","stream":true,"input":"hi"}`)
		return svc.ForwardAsResponses(ctx, c, account, body, nil)
	default:
		if kind == "native" {
			account.Extra["anthropic_passthrough"] = false
		}
		body := []byte(`{"model":"claude-sonnet-4-5","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
		parsed, err := ParseGatewayRequest(body, PlatformAnthropic)
		require.NoError(t, err)
		return svc.Forward(ctx, c, account, parsed)
	}
}

func streamBillingAnthropicContext(ctx context.Context) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil).WithContext(ctx)
	return c, rec
}

func TestStreamBillingAnthropicForwardPartial(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		for _, end := range []string{"eof", "read_error", "terminal"} {
			t.Run(kind+"/"+end, func(t *testing.T) {
				payload := streamBillingAnthropicStart + streamBillingAnthropicDelta
				if end == "terminal" {
					payload += streamBillingAnthropicStop
				}
				var reader io.ReadCloser = io.NopCloser(strings.NewReader(payload))
				if end == "read_error" {
					reader = &streamReadCloser{payload: []byte(payload), err: io.ErrUnexpectedEOF}
				}
				body := &streamBillingCountBody{ReadCloser: reader}
				c, rec := streamBillingAnthropicContext(context.Background())
				result, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, c.Request.Context())
				if end == "terminal" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.NotContains(t, rec.Body.String(), "[DONE]")
					require.NotContains(t, rec.Body.String(), "response.completed")
				}
				require.NotNil(t, result)
				require.Equal(t, 11, result.Usage.InputTokens)
				require.Equal(t, 9, result.Usage.OutputTokens)
				require.Equal(t, 7, result.Usage.CacheReadInputTokens)
				require.Equal(t, 5, result.Usage.CacheCreationInputTokens)
				require.Equal(t, 2, result.Usage.CacheCreation5mTokens)
				require.Equal(t, 3, result.Usage.CacheCreation1hTokens)
				require.Equal(t, "rid-billing", result.RequestID)
				require.Equal(t, "claude-sonnet-4-5", result.Model)
				require.NotNil(t, result.FirstTokenMs)
				require.EqualValues(t, 1, body.closes.Load(), "actual Forward must close its body exactly once")
			})
		}
	}
}

type streamBillingPreludeWriter struct {
	gin.ResponseWriter
	committed bool
}

func (w *streamBillingPreludeWriter) Committed() bool { return w.committed }

func TestStreamBillingAnthropicFirstResponseCancellation(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		for _, committed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/committed=%t", kind, committed), func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
				c, _ := streamBillingAnthropicContext(ctx)
				c.Writer = &streamBillingPreludeWriter{ResponseWriter: c.Writer, committed: committed}
				_, _ = c.Writer.WriteString(": prelude\n\n")
				cancel()
				body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(streamBillingAnthropicStart + streamBillingAnthropicDelta + streamBillingAnthropicStop))}
				result, _ := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, ctx)
				if !committed {
					require.Nil(t, result, "buffered prelude is not a billable first response")
				} else {
					require.NotNil(t, result)
					require.True(t, result.ClientDisconnect)
				}
				require.EqualValues(t, 1, body.closes.Load())
			})
		}
	}
}

func TestStreamBillingAnthropicNoUsageAndFailover(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		for _, payload := range []string{"event: ping\ndata: {\"type\":\"ping\"}\n\n", streamBillingAnthropicStart + "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"busy\"}}\n\n"} {
			t.Run(kind+"/"+fmt.Sprint(len(payload)), func(t *testing.T) {
				body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
				c, _ := streamBillingAnthropicContext(context.Background())
				result, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, c.Request.Context())
				require.Error(t, err)
				require.Nil(t, result)
				if strings.Contains(payload, "overloaded_error") {
					var failover *UpstreamFailoverError
					require.ErrorAs(t, err, &failover)
				}
				require.EqualValues(t, 1, body.closes.Load())
			})
		}
	}
}

func TestStreamBillingAnthropicObservedTokensAndWrappedFailover(t *testing.T) {
	for _, usage := range []ClaudeUsage{{InputTokens: 1}, {OutputTokens: 1}, {CacheCreationInputTokens: 1}, {CacheReadInputTokens: 1}, {CacheCreation5mTokens: 1}, {CacheCreation1hTokens: 1}, {ImageOutputTokens: 1}} {
		require.True(t, usage.hasObservedTokens())
		c, _ := streamBillingAnthropicContext(context.Background())
		result := partialStreamUsageResult(context.Background(), c, &http.Response{Header: http.Header{}}, &streamingResult{usage: &usage}, "m", "u", time.Now(), fmt.Errorf("wrapped: %w", &UpstreamFailoverError{}))
		require.Nil(t, result)
	}
	require.False(t, (*ClaudeUsage)(nil).hasObservedTokens())
	require.False(t, (&ClaudeUsage{}).hasObservedTokens())
	require.False(t, StreamBillingFirstResponseUncommitted(context.Background(), nil))
}
