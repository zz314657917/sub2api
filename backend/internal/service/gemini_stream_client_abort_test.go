package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type streamBillingGeminiBody struct {
	data   *bytes.Reader
	err    error
	onEnd  func()
	closed int
	ended  bool
}

func (b *streamBillingGeminiBody) Read(p []byte) (int, error) {
	n, err := b.data.Read(p)
	if errors.Is(err, io.EOF) {
		b.ended = true
		if b.onEnd != nil {
			b.onEnd()
		}
		return n, b.err
	}
	return n, err
}

func (b *streamBillingGeminiBody) Close() error { b.closed++; return nil }

const (
	streamBillingGeminiText  = `data: {"candidates":[{"content":{"parts":[{"text":"hello"}]}}]}` + "\n\n"
	streamBillingGeminiUsage = `data: {"modelVersion":"gemini-2.5-flash","candidates":[{"content":{"parts":[{"text":"hello"}]}}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2,"cachedContentTokenCount":3}}` + "\n\n"
	streamBillingGeminiFinal = `data: {"candidates":[{"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5,"cachedContentTokenCount":3,"thoughtsTokenCount":2}}` + "\n\n"
	streamBillingGeminiImage = `data: {"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]}}]}` + "\n\n"
)

func streamBillingGeminiSetup(t *testing.T, route, model, sse string, readErr error, cancelOnEnd bool) (*GeminiMessagesCompatService, *gin.Context, *geminiCompatHTTPUpstreamStub, *streamBillingGeminiBody, context.CancelFunc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("stream_billing_gemini_recorder", rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)
	body := &streamBillingGeminiBody{data: bytes.NewReader([]byte(sse)), err: readErr}
	if cancelOnEnd {
		body.onEnd = cancel
	}
	stub := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}, "X-Goog-Request-Id": {"gemini-stream-billing"}},
		Body:       body,
	}}
	svc := &GeminiMessagesCompatService{httpUpstream: stub, cfg: &config.Config{}}
	return svc, c, stub, body, cancel
}

func streamBillingGeminiForward(s *GeminiMessagesCompatService, c *gin.Context, route, model string) (*ForwardResult, error) {
	account := &Account{ID: 101, Platform: PlatformGemini, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "test-key"}, Concurrency: 1}
	switch route {
	case "native":
		return s.ForwardNative(c.Request.Context(), c, account, model, "streamGenerateContent", true,
			[]byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
	case "messages":
		return s.Forward(c.Request.Context(), c, account,
			[]byte(`{"model":"`+model+`","stream":true,"max_tokens":100,"messages":[{"role":"user","content":"hi"}]}`))
	default:
		return s.ForwardAsChatCompletions(c.Request.Context(), c, account,
			[]byte(`{"model":"`+model+`","stream":true,"messages":[{"role":"user","content":"hi"}]}`))
	}
}

func TestStreamBillingGeminiForwardCancellation(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		for _, tc := range []struct {
			name    string
			sse     string
			readErr error
			cancel  bool
			bill    bool
			output  int
		}{
			{"partial", streamBillingGeminiUsage, context.Canceled, true, true, 2},
			{"final", streamBillingGeminiUsage + streamBillingGeminiFinal, context.Canceled, true, true, 7},
			{"canceled_wrapped_read_error", streamBillingGeminiUsage, io.ErrUnexpectedEOF, true, true, 2},
			{"canceled_EOF", streamBillingGeminiUsage, io.EOF, true, true, 2},
			{"canceled_EOF_tail", streamBillingGeminiUsage + strings.TrimSuffix(streamBillingGeminiFinal, "\n\n"), io.EOF, true, true, 7},
			{"no_usage", streamBillingGeminiText, context.Canceled, true, false, 0},
			{"no_usage_EOF", streamBillingGeminiText, io.EOF, true, false, 0},
			{"upstream_read_error", streamBillingGeminiUsage, io.ErrUnexpectedEOF, false, false, 0},
			{"malformed_frame", "data: {invalid}\n\n", context.Canceled, true, false, 0},
			{"truncated_usage_tail", streamBillingGeminiUsage + strings.TrimSuffix(streamBillingGeminiFinal, "\n\n"), context.Canceled, true, true, 2},
			{"truncated_live_tail", streamBillingGeminiUsage + strings.TrimSuffix(streamBillingGeminiFinal, "\n\n"), io.ErrUnexpectedEOF, false, false, 0},
		} {
			t.Run(route+"/"+tc.name, func(t *testing.T) {
				svc, c, stub, body, _ := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", tc.sse, tc.readErr, tc.cancel)
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
				require.Equal(t, 1, stub.calls)
				require.Equal(t, 1, body.closed, "forwarder must close this attempt")
				if !tc.bill {
					require.Error(t, err)
					require.Nil(t, result)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 7, result.Usage.InputTokens)
				require.Equal(t, tc.output, result.Usage.OutputTokens)
				require.Equal(t, 3, result.Usage.CacheReadInputTokens)
				require.Equal(t, "gemini-stream-billing", result.RequestID)
				require.Equal(t, "gemini-2.5-flash", result.Model)
				require.Equal(t, "gemini-2.5-flash", result.UpstreamModel)
				require.Equal(t, "gemini-2.5-flash", result.UpstreamResponseModel)
				require.False(t, result.UpstreamResponseModelConflict)
				require.NotNil(t, result.FirstTokenMs)
				require.Zero(t, result.ImageCount)
				require.Error(t, stub.lastReq.Context().Err(), "upstream remains bound to the client")
				output := c.MustGet("stream_billing_gemini_recorder").(*httptest.ResponseRecorder).Body.String()
				if route == "messages" {
					require.NotContains(t, output, "event: message_stop")
				}
				if route == "chat" {
					require.NotContains(t, output, "data: [DONE]")
				}
			})
		}
	}
}

func TestStreamBillingGeminiCancellationDuringDo(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		t.Run(route, func(t *testing.T) {
			svc, c, stub, body, cancel := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", "", io.EOF, false)
			stub.err = context.Canceled
			cancel()
			result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
			require.ErrorIs(t, err, context.Canceled)
			require.Nil(t, result)
			require.Equal(t, 1, stub.calls, "canceled stream must not consume retry budget")
			require.Zero(t, body.closed, "no response body was returned by Do")
		})
	}
}

func TestStreamBillingGeminiNonstreamReadFailure(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		t.Run(route, func(t *testing.T) {
			svc, c, _, body, _ := streamBillingGeminiSetup(t, route, "gemini-2.5-flash",
				`{"candidates":[{"content":{"parts":[{"text":"hello"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5}}`,
				context.Canceled, true)
			account := &Account{ID: 101, Platform: PlatformGemini, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test-key"}, Concurrency: 1}
			var result *ForwardResult
			var err error
			switch route {
			case "native":
				result, err = svc.ForwardNative(c.Request.Context(), c, account, "gemini-2.5-flash", "generateContent", false,
					[]byte(`{"contents":[{"parts":[{"text":"hi"}]}]}`))
			case "messages":
				result, err = svc.Forward(c.Request.Context(), c, account,
					[]byte(`{"model":"gemini-2.5-flash","max_tokens":100,"messages":[{"role":"user","content":"hi"}]}`))
			case "chat":
				result, err = svc.ForwardAsChatCompletions(c.Request.Context(), c, account,
					[]byte(`{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`))
			}
			require.Error(t, err)
			require.Nil(t, result, "nonstream read failures must retain existing error semantics")
			require.Equal(t, 1, body.closed)
		})
	}
}
