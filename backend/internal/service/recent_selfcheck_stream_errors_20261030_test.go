package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const recentSelfcheckAnthropicText = "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n" +
	"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"selfcheck-visible-text\"}}\n\n"
const recentSelfcheckAnthropicError = "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"busy\"}}\n\n"
const recentSelfcheckAnthropicZeroStart = "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_zero\",\"model\":\"claude-sonnet-4-5\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"usage\":{\"input_tokens\":0,\"output_tokens\":0}}}\n\n"

// Size includes buffered prelude bytes, while Committed remains authoritative.
// Unlike the recorder, this writer never sends those bytes to the client.
type recentSelfcheckBufferedWriter struct {
	gin.ResponseWriter
	buffer strings.Builder
}

func (w *recentSelfcheckBufferedWriter) Committed() bool { return false }
func (w *recentSelfcheckBufferedWriter) Size() int       { return w.buffer.Len() }
func (w *recentSelfcheckBufferedWriter) Written() bool   { return w.Size() > 0 }
func (w *recentSelfcheckBufferedWriter) Flush()          {}
func (w *recentSelfcheckBufferedWriter) WriteHeaderNow() {}
func (w *recentSelfcheckBufferedWriter) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}
func (w *recentSelfcheckBufferedWriter) WriteString(p string) (int, error) {
	return w.buffer.WriteString(p)
}

type recentSelfcheckErrorWriteFailure struct {
	gin.ResponseWriter
	failures int
}

func (w *recentSelfcheckErrorWriteFailure) Write(p []byte) (int, error) {
	if strings.Contains(string(p), `"stream_incomplete"`) || strings.Contains(string(p), `"response.failed"`) ||
		strings.Contains(string(p), `"type":"error"`) {
		w.failures++
		return 0, io.ErrClosedPipe
	}
	return w.ResponseWriter.Write(p)
}
func (w *recentSelfcheckErrorWriteFailure) WriteString(p string) (int, error) {
	return w.Write([]byte(p))
}

func recentSelfcheckRequireClaudeUsage(t *testing.T, result *ForwardResult) {
	t.Helper()
	require.NotNil(t, result)
	require.Equal(t, 11, result.Usage.InputTokens)
	require.Equal(t, 9, result.Usage.OutputTokens)
	require.Equal(t, 7, result.Usage.CacheReadInputTokens)
	require.Equal(t, 5, result.Usage.CacheCreationInputTokens)
	require.Equal(t, 2, result.Usage.CacheCreation5mTokens)
	require.Equal(t, 3, result.Usage.CacheCreation1hTokens)
	require.Equal(t, "rid-billing", result.RequestID)
}

func TestRecentSelfcheckAnthropicOwnerErrors(t *testing.T) {
	for _, kind := range []string{"chat", "responses"} {
		for _, timeout := range []bool{false, true} {
			for _, mode := range []string{
				"committed_usage", "eof_control", "buffered_usage", "zero_output",
				"committed_zero_usage", "error_write_failure", "already_disconnected", "already_canceled",
			} {
				t.Run(fmt.Sprintf("%s/timeout=%t/%s", kind, timeout, mode), func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					c, rec := streamBillingAnthropicContext(ctx)
					payload := streamBillingAnthropicStart + recentSelfcheckAnthropicText + streamBillingAnthropicDelta
					var buffered *recentSelfcheckBufferedWriter
					var failed *recentSelfcheckErrorWriteFailure
					var disconnected *streamBillingDisconnectWriter
					switch mode {
					case "buffered_usage":
						payload = streamBillingAnthropicStart + streamBillingAnthropicDelta
						buffered = &recentSelfcheckBufferedWriter{ResponseWriter: c.Writer}
						c.Writer = buffered
					case "zero_output":
						payload = ""
					case "committed_zero_usage":
						payload = recentSelfcheckAnthropicZeroStart + recentSelfcheckAnthropicText
					case "error_write_failure":
						failed = &recentSelfcheckErrorWriteFailure{ResponseWriter: c.Writer}
						c.Writer = failed
					case "already_disconnected":
						disconnected = &streamBillingDisconnectWriter{ResponseWriter: c.Writer}
						c.Writer = disconnected
					case "already_canceled":
						cancel()
					}
					if mode != "eof_control" {
						payload += recentSelfcheckAnthropicError
					}
					body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
					svc := streamBillingAnthropicService(body)
					if timeout {
						svc.cfg.Gateway.StreamDataIntervalTimeout = 1
					}
					result, err := streamBillingAnthropicForward(t, kind, svc, c, ctx)
					require.Error(t, err)
					require.EqualValues(t, 1, body.closes.Load())
					require.NotContains(t, rec.Body.String(), "[DONE]")
					require.NotContains(t, rec.Body.String(), "response.completed")
					var failover *UpstreamFailoverError
					if mode == "buffered_usage" || mode == "zero_output" {
						require.ErrorAs(t, err, &failover)
						require.Nil(t, result, "an uncommitted retry must not bill observed upstream usage")
						require.Empty(t, rec.Body.String())
						require.False(t, IsResponseCommitted(c), "failure handling must not cross the commit barrier")
						if buffered != nil {
							require.Positive(t, buffered.Size())
							require.True(t, buffered.Written())
							require.False(t, buffered.Committed())
							require.NotContains(t, buffered.buffer.String(), "stream_incomplete")
							require.NotContains(t, buffered.buffer.String(), "response.failed")
						}
						return
					}
					require.False(t, errors.As(err, &failover), "delivered output/disconnect forbids replay")
					if mode == "committed_zero_usage" {
						require.Nil(t, result, "no observed tokens must not create a billable result")
					} else {
						recentSelfcheckRequireClaudeUsage(t, result)
						require.Equal(t, mode == "error_write_failure" || mode == "already_disconnected" || mode == "already_canceled", result.ClientDisconnect)
					}
					if mode != "eof_control" {
						var upstreamError *sseStreamErrorEventError
						require.ErrorAs(t, err, &upstreamError)
						require.Contains(t, upstreamError.RawData, "overloaded_error")
						require.Contains(t, upstreamError.RawData, "busy")
					}
					if mode == "already_canceled" {
						require.Empty(t, rec.Body.String())
					} else if disconnected != nil {
						require.Equal(t, 1, disconnected.writes, "never write again after client disconnect")
						require.Empty(t, rec.Body.String())
					} else {
						require.Contains(t, rec.Body.String(), "selfcheck-visible-text")
						if failed != nil {
							require.Equal(t, 1, failed.failures, "only one attempt to write the error")
						} else if kind == "responses" {
							require.Equal(t, 1, strings.Count(rec.Body.String(), "event: response.failed"))
						} else {
							require.Equal(t, 1, strings.Count(rec.Body.String(), `"stream_incomplete"`))
						}
					}
				})
			}
		}
	}
}

func TestRecentSelfcheckAnthropicFailureBarrier(t *testing.T) {
	for _, mode := range []string{"unwritten", "buffered_size", "written", "headers_written", "committed_overrides_size", "disconnected", "write_failure", "zero_usage"} {
		t.Run(mode, func(t *testing.T) {
			c, _ := streamRetryContext()
			result := &ForwardResult{Usage: ClaudeUsage{InputTokens: 11, OutputTokens: 9}}
			switch mode {
			case "buffered_size":
				w := &recentSelfcheckBufferedWriter{ResponseWriter: c.Writer}
				c.Writer = w
				_, _ = w.WriteString(": prelude\n\n")
			case "written", "write_failure", "zero_usage":
				_, _ = c.Writer.WriteString("data: delivered\n\n")
			case "headers_written":
				c.Writer.WriteHeaderNow()
			case "committed_overrides_size":
				c.Writer = &streamBillingPreludeWriter{ResponseWriter: c.Writer, committed: true}
				require.Equal(t, -1, c.Writer.Size())
			case "disconnected":
				result.ClientDisconnect = true
			}
			if mode == "zero_usage" {
				result.Usage = ClaudeUsage{}
			}
			writes := 0
			original := &sseStreamErrorEventError{RawData: `{"type":"error","error":{"message":"busy"}}`}
			got, err := anthropicCompatBillingFailure(c, result, original, func() bool {
				writes++
				MarkResponseCommitted(c)
				return mode == "write_failure"
			})
			var failover *UpstreamFailoverError
			if mode == "unwritten" || mode == "buffered_size" {
				require.ErrorAs(t, err, &failover)
				require.Nil(t, got)
				require.Zero(t, writes)
				require.False(t, IsResponseCommitted(c))
				return
			}
			require.False(t, errors.As(err, &failover))
			require.Same(t, original, err, "preserve the original upstream error")
			if mode == "zero_usage" {
				require.Nil(t, got)
			} else {
				require.Same(t, result, got)
				require.Equal(t, mode == "disconnected" || mode == "write_failure", got.ClientDisconnect)
			}
			if mode == "disconnected" {
				require.Zero(t, writes)
			} else {
				require.Equal(t, 1, writes)
			}
		})
	}
}

func recentSelfcheckOpenAIProcess(s *OpenAIGatewayService, c *gin.Context, body io.ReadCloser, headers http.Header, passthrough bool) (*OpenAIUsage, bool, error) {
	resp := &http.Response{StatusCode: http.StatusOK, Body: body, Header: headers}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	if passthrough {
		result, err := s.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
		if result == nil {
			return nil, false, err
		}
		return result.usage, result.clientDisconnected, err
	}
	result, err := s.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
	if result == nil {
		return nil, false, err
	}
	return result.usage, result.clientDisconnected, err
}

func recentSelfcheckOpenAIService(timeout bool) *OpenAIGatewayService {
	cfg := &config.Config{}
	if timeout {
		cfg.Gateway.StreamDataIntervalTimeout = 1
		cfg.Gateway.StreamKeepaliveInterval = 1
	}
	return &OpenAIGatewayService{cfg: cfg}
}

func TestRecentSelfcheckOpenAIBareClassification(t *testing.T) {
	cases := []struct {
		name, payload            string
		headers                  http.Header
		failover, scoped, tagged bool
		limit                    int
		delay                    time.Duration
		status                   int
	}{
		{name: "overload", payload: `{"type":"error","error":{"code":"server_is_overloaded","headers":{"Retry-After":"2"}}}`, failover: true, scoped: true, limit: 3, delay: 2 * time.Second},
		{name: "slow_down_cap", payload: `{"type":"error","error":{"code":"slow_down","headers":{"retry-after":"99"}}}`, failover: true, scoped: true, limit: 3, delay: 8 * time.Second},
		{name: "http_header_priority", payload: `{"type":"error","error":{"code":"server_is_overloaded","headers":{"Retry-After":"2"}}}`, headers: http.Header{"Retry-After": {"99"}}, failover: true, scoped: true, limit: 3, delay: 8 * time.Second},
		{name: "failed_control", payload: `{"type":"response.failed","response":{"error":{"code":"server_is_overloaded","headers":{"Retry-After":"2"}}}}`, failover: true, scoped: true, limit: 3, delay: 2 * time.Second},
		{name: "temporary", payload: `{"type":"error","error":{"message":"temporary upstream problem"}}`, failover: true, tagged: true},
		{name: "processing", payload: `{"type":"error","error":{"message":"An error occurred while processing your request"}}`, failover: true, tagged: true},
		{name: "please_retry", payload: `{"type":"error","error":{"message":"please retry"}}`, failover: true, tagged: true},
		{name: "capacity", payload: `{"type":"error","error":{"message":"Selected model is at capacity. Please try a different model."}}`, failover: true, limit: 5},
		{name: "auth", payload: `{"type":"error","error":{"code":"invalid_api_key"}}`, failover: true},
		{name: "rate_limit", payload: `{"type":"error","error":{"code":"rate_limit_exceeded"}}`, failover: true, status: 429},
		{name: "account", payload: `{"type":"error","error":{"code":"account_disabled"}}`, failover: true},
		{name: "overload_auth_conflict", payload: `{"type":"error","error":{"code":"server_is_overloaded","type":"authentication_error"}}`, failover: true, limit: 3},
		{name: "permission", payload: `{"type":"error","error":{"code":"permission_denied","message":"permission denied"}}`},
		{name: "context", payload: `{"type":"error","error":{"code":"context_length_exceeded","message":"maximum context length"}}`},
		{name: "policy", payload: `{"type":"error","error":{"type":"content_policy_violation","message":"content_policy violation"}}`},
		{name: "unknown", payload: `{"type":"error","error":{"message":"permanent ordinary failure"}}`},
		{name: "empty", payload: `{"type":"error","error":{}}`},
		{name: "malformed", payload: `{"type":"error","error":{"message":"permanent ordinary failure"`},
	}
	for _, passthrough := range []bool{false, true} {
		for _, timeout := range []bool{false, true} {
			for _, ending := range []string{"eof", "done"} {
				for _, tc := range cases {
					t.Run(fmt.Sprintf("passthrough=%t/timeout=%t/%s/%s", passthrough, timeout, ending, tc.name), func(t *testing.T) {
						c, rec := streamRetryContext()
						// Both lifecycle and explicit empty output remain behind the barrier.
						preamble := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"failed-attempt\"}}\n\n" +
							"data: {\"type\":\"response.output_text.delta\",\"delta\":\"\"}\n\n"
						payload := preamble + "data: " + tc.payload + "\n\n"
						if ending == "done" {
							payload += "data: [DONE]\n\n"
						}
						body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
						usage, disconnected, err := recentSelfcheckOpenAIProcess(recentSelfcheckOpenAIService(timeout), c, body, tc.headers, passthrough)
						require.Error(t, err, "DONE must never turn a bare error into success")
						require.EqualValues(t, 1, body.closes.Load())
						require.False(t, disconnected)
						require.NotNil(t, usage)
						require.Zero(t, usage.InputTokens)
						require.Zero(t, usage.OutputTokens)
						var failover *UpstreamFailoverError
						var tagged *OpenAIStreamRetryError
						require.Equal(t, tc.failover, errors.As(err, &failover))
						require.Equal(t, tc.tagged, errors.As(err, &tagged))
						if !tc.failover {
							require.NotEmpty(t, rec.Body.String(), "nonretryable errors keep their existing client terminal")
							require.NotContains(t, rec.Body.String(), "[DONE]")
							return
						}
						require.Empty(t, rec.Body.String())
						require.Equal(t, -1, OpenAIStreamWrittenSize(c))
						require.Equal(t, tc.scoped, failover.RequestScopedTransient)
						require.Equal(t, tc.limit, failover.SameAccountRetryLimit)
						require.Equal(t, tc.delay, failover.SameAccountRetryDelay)
						status := tc.status
						if status == 0 {
							status = http.StatusBadGateway
						}
						require.Equal(t, status, failover.StatusCode)
						require.Equal(t, tc.limit > 0, failover.RetryableOnSameAccount)
						if tc.headers != nil {
							require.Equal(t, tc.headers.Get("Retry-After"), failover.ResponseHeaders.Get("Retry-After"))
						}
					})
				}
			}
		}
	}
}

func TestRecentSelfcheckOpenAIBareImmediateClose(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, timeout := range []bool{false, true} {
			t.Run(fmt.Sprintf("passthrough=%t/timeout=%t", passthrough, timeout), func(t *testing.T) {
				c, rec := streamRetryContext()
				// No EOF or DONE is available: the owner must terminate on the error.
				body := newStreamBillingBlockedBody("data: {\"type\":\"error\",\"error\":{\"code\":\"server_is_overloaded\",\"headers\":{\"Retry-After\":\"2\"}}}\n\n")
				guard := time.AfterFunc(750*time.Millisecond, func() { _ = body.Close() })
				defer guard.Stop()
				start := time.Now()
				_, _, err := recentSelfcheckOpenAIProcess(recentSelfcheckOpenAIService(timeout), c, body, nil, passthrough)
				elapsed := time.Since(start)
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.True(t, failover.RequestScopedTransient)
				require.Equal(t, 2*time.Second, failover.SameAccountRetryDelay)
				require.Less(t, elapsed, 500*time.Millisecond, "must not wait for EOF/idle timeout")
				require.EqualValues(t, 1, body.closes.Load())
				require.Empty(t, rec.Body.String())
			})
		}
	}
}

func TestRecentSelfcheckOpenAIAfterOutput(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, timeout := range []bool{false, true} {
			for _, mode := range []string{"usage", "zero_usage", "buffered_writer", "disconnected", "error_write_failure"} {
				t.Run(fmt.Sprintf("passthrough=%t/timeout=%t/%s", passthrough, timeout, mode), func(t *testing.T) {
					c, rec := streamRetryContext()
					var failed *recentSelfcheckErrorWriteFailure
					var disconnected *streamBillingDisconnectWriter
					if mode == "buffered_writer" {
						c.Writer = &recentSelfcheckBufferedWriter{ResponseWriter: c.Writer}
					} else if mode == "disconnected" {
						disconnected = &streamBillingDisconnectWriter{ResponseWriter: c.Writer}
						c.Writer = disconnected
					} else if mode == "error_write_failure" {
						failed = &recentSelfcheckErrorWriteFailure{ResponseWriter: c.Writer}
						c.Writer = failed
					}
					usageJSON := `,"response":{"usage":{"input_tokens":11,"output_tokens":9,"input_tokens_details":{"cached_tokens":7}}}`
					if mode == "zero_usage" {
						usageJSON = ""
					}
					// A bare error after real output cannot be replayed, including
					// local output buffered behind a writer reporting Committed=false.
					payload := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"selfcheck-visible-text\"" + usageJSON + "}\n\n" +
						"data: {\"type\":\"error\",\"error\":{\"code\":\"server_is_overloaded\"}}\n\n" +
						"data: [DONE]\n\n"
					body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
					usage, gone, err := recentSelfcheckOpenAIProcess(recentSelfcheckOpenAIService(timeout), c, body, nil, passthrough)
					require.Error(t, err)
					var failover *UpstreamFailoverError
					require.False(t, errors.As(err, &failover))
					require.NotNil(t, usage)
					if mode == "zero_usage" {
						require.Zero(t, usage.InputTokens)
						require.Zero(t, usage.OutputTokens)
					} else {
						require.Equal(t, 11, usage.InputTokens)
						require.Equal(t, 9, usage.OutputTokens)
						require.Equal(t, 7, usage.CacheReadInputTokens)
					}
					require.EqualValues(t, 1, body.closes.Load())
					require.NotContains(t, rec.Body.String(), "[DONE]")
					if disconnected != nil {
						require.True(t, gone)
						require.Equal(t, 1, disconnected.writes)
					} else if mode != "buffered_writer" {
						require.Contains(t, rec.Body.String(), "selfcheck-visible-text")
					}
					if failed != nil {
						require.True(t, gone)
						require.Equal(t, 1, failed.failures)
					}
				})
			}
		}
	}
}

func TestRecentSelfcheckOpenAIBareMalformedBarrier(t *testing.T) {
	const overload = `{"type":"error","error":{"code":"server_is_overloaded","headers":{"Retry-After":"2"}}}`
	cases := []struct {
		name, payload string
		failover      bool
		followError   bool
	}{
		{name: "malformed_overload", payload: `{"type":"error","error":{"code":"server_is_overloaded"`},
		{name: "valid_overload", payload: overload, failover: true},
		{name: "nonobject_array", payload: "[" + overload + "]", followError: true},
		{name: "nonobject_string", payload: `"server_is_overloaded"`, followError: true},
		{name: "nonobject_null", payload: "null", followError: true},
		{name: "nonobject_number", payload: "42", followError: true},
	}
	for _, passthrough := range []bool{false, true} {
		for _, timeout := range []bool{false, true} {
			for _, ending := range []string{"eof", "done"} {
				for _, tc := range cases {
					t.Run(fmt.Sprintf("passthrough=%t/timeout=%t/%s/%s", passthrough, timeout, ending, tc.name), func(t *testing.T) {
						c, rec := streamRetryContext()
						payload := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"malformed-barrier\"}}\n\n" +
							"data: " + tc.payload + "\n\n"
						if tc.followError {
							// A nonobject frame must commit output before a later
							// recognized error, so the request cannot be replayed.
							payload += "data: " + overload + "\n\n"
						}
						if ending == "done" {
							payload += "data: [DONE]\n\n"
						}
						body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
						svc := recentSelfcheckOpenAIService(timeout)
						svc.cfg.Gateway.LogUpstreamErrorBody = true
						usage, disconnected, err := recentSelfcheckOpenAIProcess(svc, c, body, nil, passthrough)
						require.Error(t, err, "an upstream error must not become success at DONE")
						require.EqualValues(t, 1, body.closes.Load())
						require.False(t, disconnected)
						require.NotNil(t, usage)
						require.Zero(t, usage.InputTokens)
						require.Zero(t, usage.OutputTokens)
						var failover *UpstreamFailoverError
						var tagged *OpenAIStreamRetryError
						require.Equal(t, tc.failover, errors.As(err, &failover), "malformed/nonobject frames forbid replay")
						require.False(t, errors.As(err, &tagged))
						require.NotContains(t, rec.Body.String(), "[DONE]")
						if tc.failover {
							require.Empty(t, rec.Body.String())
							require.Equal(t, -1, OpenAIStreamWrittenSize(c))
							require.Equal(t, overload, c.GetString(OpsUpstreamErrorDetailKey), "retry classification keeps the original payload for diagnostics")
							require.True(t, failover.RequestScopedTransient)
							require.True(t, failover.RetryableOnSameAccount)
							require.Equal(t, 3, failover.SameAccountRetryLimit)
							require.Equal(t, 2*time.Second, failover.SameAccountRetryDelay)
							require.Equal(t, http.StatusBadGateway, failover.StatusCode)
							return
						}
						require.Contains(t, rec.Body.String(), "data: "+tc.payload+"\n", "the original frame crosses the client output barrier")
						require.Positive(t, OpenAIStreamWrittenSize(c))
						if tc.followError {
							require.Contains(t, rec.Body.String(), "server_is_overloaded")
						}
					})
				}
			}
		}
	}
}
