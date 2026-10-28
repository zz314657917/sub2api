package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const streamBillingOpenAITerminal = `data: {"type":"response.completed","response":{"id":"resp_billing","model":"gpt-5.2-observed","usage":{"input_tokens":21,"output_tokens":9,"input_tokens_details":{"cached_tokens":6}}}}` + "\n\n"
const streamBillingOpenAIPartial = `data: {"type":"response.in_progress","response":{"id":"resp_billing","model":"gpt-5.2-observed","usage":{"input_tokens":21,"output_tokens":9,"input_tokens_details":{"cached_tokens":6}}}}` + "\n\n"
const streamBillingOpenAIOutput = `data: {"type":"response.output_text.delta","delta":"paid output"}` + "\n\n"

type streamBillingOpenAIBody struct {
	io.ReadCloser
	closes atomic.Int32
}

func (b *streamBillingOpenAIBody) Close() error {
	b.closes.Add(1)
	return b.ReadCloser.Close()
}

type streamBillingOpenAIUpstream struct {
	HTTPUpstream
	body    *streamBillingOpenAIBody
	after   func()
	url     string
	request *http.Request
	calls   int
}

func (u *streamBillingOpenAIUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.calls++
	u.request = req
	var resp *http.Response
	if u.url != "" {
		loopReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, u.url, nil)
		if err != nil {
			return nil, err
		}
		resp, err = http.DefaultClient.Do(loopReq)
		if err != nil {
			return nil, err
		}
		u.body = &streamBillingOpenAIBody{ReadCloser: resp.Body}
		resp.Body = u.body
	} else {
		resp = &http.Response{StatusCode: http.StatusOK, Header: http.Header{
			"Content-Type": {"text/event-stream"}, "X-Request-Id": {"rid_billing"},
		}, Body: u.body}
	}
	if u.after != nil {
		u.after()
	}
	return resp, nil
}

func streamBillingOpenAIContext(ctx context.Context) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.1.0")
	return c, rec
}

func streamBillingOpenAIService(up *streamBillingOpenAIUpstream, passthrough bool) (*OpenAIGatewayService, *Account) {
	s := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: up}
	a := &Account{
		ID: 123, Name: "billing-test", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Concurrency: 1, Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "test"},
		Extra:  map[string]any{"openai_passthrough": passthrough, "openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff},
		Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1),
	}
	return s, a
}

func streamBillingOpenAIForward(s *OpenAIGatewayService, c *gin.Context, a *Account) (*OpenAIForwardResult, error) {
	return s.Forward(c.Request.Context(), c, a, []byte(`{"model":"gpt-5.2","stream":true,"instructions":"test","input":"hello","service_tier":"priority","reasoning":{"effort":"high"}}`))
}

func TestStreamBillingOpenAILoopbackCancelBeforeFirstByte(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			release := make(chan struct{})
			serverExited := make(chan struct{})
			var closeRelease sync.Once
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(serverExited)
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("X-Request-Id", "rid_loopback")
				w.(http.Flusher).Flush() // No upstream body byte exists when cancellation occurs.
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				// Make terminal usage arrive after cancellation, rather than
				// merely racing a terminal already queued in the socket.
				select {
				case <-time.After(10 * time.Millisecond):
				case <-r.Context().Done():
					return
				}
				_, _ = io.WriteString(w, streamBillingOpenAITerminal)
				w.(http.Flusher).Flush()
				<-r.Context().Done() // Terminal must close the body instead of waiting for EOF.
			}))
			defer server.Close()
			defer closeRelease.Do(func() { close(release) })
			up := &streamBillingOpenAIUpstream{url: server.URL, after: func() {
				cancel()
				closeRelease.Do(func() { close(release) })
			}}
			s, a := streamBillingOpenAIService(up, passthrough)
			c, rec := streamBillingOpenAIContext(ctx)
			result, err := streamBillingOpenAIForward(s, c, a)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 21, result.Usage.InputTokens)
			require.Equal(t, 9, result.Usage.OutputTokens)
			require.Equal(t, 6, result.Usage.CacheReadInputTokens)
			require.Equal(t, "gpt-5.2", result.Model)
			if passthrough {
				require.Empty(t, result.UpstreamModel, "unmapped OAuth passthrough keeps its established empty mapping field")
			} else {
				require.Equal(t, "gpt-5.2", result.UpstreamModel)
			}
			require.Equal(t, "gpt-5.2-observed", result.UpstreamResponseModel)
			require.Equal(t, "resp_billing", result.ResponseID)
			require.Equal(t, "rid_loopback", result.RequestID)
			require.NotNil(t, result.ServiceTier)
			require.Equal(t, "priority", *result.ServiceTier)
			require.NotNil(t, result.ReasoningEffort)
			require.Equal(t, "high", *result.ReasoningEffort)
			require.Positive(t, result.Duration)
			require.Empty(t, rec.Body.String())
			require.EqualValues(t, 1, up.body.closes.Load())
			require.Equal(t, 1, up.calls)
			require.NoError(t, up.request.Context().Err(), "ordinary client cancellation must leave upstream detached")
			select {
			case <-serverExited:
			case <-time.After(time.Second):
				t.Fatal("terminal completion did not close the loopback connection")
			}
		})
	}
}

type streamBillingOpenAITail struct {
	err    error
	cancel func()
}

func (r streamBillingOpenAITail) Read([]byte) (int, error) {
	if r.cancel != nil {
		r.cancel()
	}
	return 0, r.err
}

type streamBillingOpenAIAttemptWriter struct {
	gin.ResponseWriter
	committed bool
	size      int
	cancel    func()
}

func (w *streamBillingOpenAIAttemptWriter) Committed() bool { return w.committed }
func (w *streamBillingOpenAIAttemptWriter) Size() int       { return w.size }
func (w *streamBillingOpenAIAttemptWriter) Written() bool   { return w.size >= 0 }
func (w *streamBillingOpenAIAttemptWriter) Flush()          {}
func (w *streamBillingOpenAIAttemptWriter) Write(p []byte) (int, error) {
	w.size += len(p)
	if w.cancel != nil {
		w.cancel()
	}
	return len(p), nil
}
func (w *streamBillingOpenAIAttemptWriter) WriteString(p string) (int, error) {
	return w.Write([]byte(p))
}

func TestStreamBillingOpenAIMarkedAttemptCommitBoundary(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, committed := range []bool{false, true} {
			t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough]+map[bool]string{false: "/uncommitted", true: "/committed"}[committed], func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
				defer cancel()
				up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: io.NopCloser(io.MultiReader(
					strings.NewReader(streamBillingOpenAIPartial),
					streamBillingOpenAITail{err: context.Canceled, cancel: cancel},
				))}}
				s, a := streamBillingOpenAIService(up, passthrough)
				c, rec := streamBillingOpenAIContext(ctx)
				writer := &streamBillingOpenAIAttemptWriter{ResponseWriter: c.Writer, committed: committed}
				_, _ = writer.WriteString(": buffered prelude\n\n")
				c.Writer = writer
				require.Positive(t, writer.Size())
				require.True(t, writer.Written())
				result, err := streamBillingOpenAIForward(s, c, a)
				require.ErrorIs(t, err, context.Canceled)
				if committed {
					require.NotNil(t, result)
					require.Equal(t, 21, result.Usage.InputTokens)
					require.True(t, result.ClientDisconnect)
				} else {
					require.Nil(t, result, "an uncommitted timed-out attempt must never be billed")
				}
				require.Empty(t, rec.Body.String())
				require.EqualValues(t, 1, up.body.closes.Load())
				require.Equal(t, 1, up.calls)
				require.ErrorIs(t, up.request.Context().Err(), context.Canceled, "marked attempts must retain upstream cancellation")
			})
		}
	}
}

func TestStreamBillingOpenAIMarkedBufferedOutputNeverBills(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough], func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
			defer cancel()
			up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: io.NopCloser(strings.NewReader(
				streamBillingOpenAIPartial + streamBillingOpenAIOutput + streamBillingOpenAITerminal,
			))}}
			s, a := streamBillingOpenAIService(up, passthrough)
			c, rec := streamBillingOpenAIContext(ctx)
			writer := &streamBillingOpenAIAttemptWriter{ResponseWriter: c.Writer, cancel: cancel}
			c.Writer = writer
			result, err := streamBillingOpenAIForward(s, c, a)
			require.ErrorIs(t, err, context.Canceled)
			require.Nil(t, result, "a local output barrier cannot make buffered, uncommitted bytes billable")
			require.Positive(t, writer.Size())
			require.True(t, writer.Written())
			require.False(t, writer.Committed())
			require.Empty(t, rec.Body.String())
			require.EqualValues(t, 1, up.body.closes.Load())
		})
	}
}

// Embedding the Gin interface forwards Write/Size/Written while hiding optional
// methods on the wrapped value, as ordinary response-writer decorators can do.
type streamBillingOpenAIOpaqueWriter struct {
	gin.ResponseWriter
}

func TestStreamBillingOpenAIMarkedOpaqueWriterNeverBills(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough], func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
			defer cancel()
			up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: io.NopCloser(io.MultiReader(
				strings.NewReader(streamBillingOpenAIPartial+streamBillingOpenAIOutput),
				streamBillingOpenAITail{err: context.Canceled},
			))}}
			s, a := streamBillingOpenAIService(up, passthrough)
			c, rec := streamBillingOpenAIContext(ctx)
			buffered := &streamBillingOpenAIAttemptWriter{ResponseWriter: c.Writer}
			_, _ = buffered.WriteString(": buffered prelude\n\n")
			buffered.cancel = cancel
			c.Writer = &streamBillingOpenAIOpaqueWriter{ResponseWriter: buffered}
			_, hasCommitted := c.Writer.(interface{ Committed() bool })
			require.False(t, hasCommitted, "the wrapper must actually hide the optional commit interface")
			require.Positive(t, c.Writer.Size())
			require.True(t, c.Writer.Written())

			result, err := streamBillingOpenAIForward(s, c, a)
			require.ErrorIs(t, err, context.Canceled)
			require.Nil(t, result, "observed usage cannot make an opaque, unproven commit billable")
			require.True(t, openAIStreamUncommittedAttempt(ctx, c, true), "localStarted=true is a replay barrier, not commit proof")
			require.Positive(t, c.Writer.Size())
			require.True(t, c.Writer.Written())
			require.False(t, buffered.Committed())
			require.Empty(t, rec.Body.String())
			require.EqualValues(t, 1, up.body.closes.Load())
			require.Equal(t, 1, up.calls)
		})
	}
}

func TestStreamBillingOpenAIObservedImagePartial(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough], func(t *testing.T) {
			payload := `data: {"type":"response.output_item.done","item":{"type":"image_generation_call","id":"img_partial","result":"aW1n","size":"1024x1024"}}` + "\n\n"
			up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: io.NopCloser(io.MultiReader(
				strings.NewReader(payload), streamBillingOpenAITail{err: io.ErrUnexpectedEOF},
			))}}
			s, a := streamBillingOpenAIService(up, passthrough)
			c, rec := streamBillingOpenAIContext(context.Background())
			result, err := streamBillingOpenAIForward(s, c, a)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
			require.NotNil(t, result, "observed image usage must survive without token usage")
			require.Equal(t, 1, result.ImageCount)
			require.Zero(t, result.Usage.InputTokens)
			require.Zero(t, result.Usage.OutputTokens)
			require.Contains(t, rec.Body.String(), "img_partial")
			require.Equal(t, 1, up.calls)
			require.EqualValues(t, 1, up.body.closes.Load())
		})
	}
}

func TestStreamBillingOpenAIForwardPartialAndFailover(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, tc := range []struct {
			name     string
			payload  string
			tail     error
			cancel   bool
			wantErr  bool
			partial  bool
			failover bool
		}{
			{"read_error_after_output", streamBillingOpenAIPartial + streamBillingOpenAIOutput, io.ErrUnexpectedEOF, false, true, true, false},
			{"missing_terminal_after_output", streamBillingOpenAIPartial + streamBillingOpenAIOutput, io.EOF, false, true, true, false},
			{"cancel_without_output", streamBillingOpenAIPartial, context.Canceled, true, true, true, false},
			{"no_usage", streamBillingOpenAIOutput, io.ErrUnexpectedEOF, false, true, false, false},
			{"no_output_failover_with_usage", streamBillingOpenAIPartial, io.ErrUnexpectedEOF, false, true, false, true},
			{"terminal_then_cancel", streamBillingOpenAITerminal, context.Canceled, true, false, true, false},
			{"terminal_then_read_error", streamBillingOpenAITerminal, io.ErrUnexpectedEOF, false, false, true, false},
			{"failed_before_output_with_usage", `data: {"type":"response.failed","response":{"usage":{"input_tokens":21},"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}}` + "\n\n", io.EOF, false, true, false, true},
			{"failed_after_output_with_usage", streamBillingOpenAIPartial + streamBillingOpenAIOutput + `data: {"type":"response.failed","response":{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}}` + "\n\n", io.EOF, false, true, true, false},
			{"in_band_error_after_output", streamBillingOpenAIPartial + streamBillingOpenAIOutput + `data: {"type":"error","error":{"message":"upstream error"}}` + "\n\n", io.EOF, false, true, true, false},
			{"cancel_without_usage", "", context.Canceled, true, true, false, false},
		} {
			t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough]+"/"+tc.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				tail := streamBillingOpenAITail{err: tc.tail}
				if tc.cancel {
					tail.cancel = cancel
				}
				up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: io.NopCloser(io.MultiReader(strings.NewReader(tc.payload), tail))}}
				s, a := streamBillingOpenAIService(up, passthrough)
				c, _ := streamBillingOpenAIContext(ctx)
				result, err := streamBillingOpenAIForward(s, c, a)
				require.Equal(t, tc.wantErr, err != nil, "%v", err)
				var failover *UpstreamFailoverError
				require.Equal(t, tc.failover, errors.As(err, &failover), "%v", err)
				if tc.partial {
					require.NotNil(t, result)
					require.Equal(t, 21, result.Usage.InputTokens)
					require.Equal(t, 9, result.Usage.OutputTokens)
					require.Equal(t, 6, result.Usage.CacheReadInputTokens)
					require.Equal(t, "gpt-5.2-observed", result.UpstreamResponseModel)
					require.Equal(t, "resp_billing", result.ResponseID)
				} else {
					require.Nil(t, result)
				}
				require.Equal(t, 1, up.calls, "a visible stream must never be replayed")
				require.EqualValues(t, 1, up.body.closes.Load())
			})
		}
	}
}

func TestStreamBillingOpenAILoopbackAbortedUpstream(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, observed := range []bool{false, true} {
			t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough]+map[bool]string{false: "/no_usage", true: "/observed_usage"}[observed], func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/event-stream")
					if observed {
						_, _ = io.WriteString(w, streamBillingOpenAIPartial)
					}
					_, _ = io.WriteString(w, streamBillingOpenAIOutput)
					w.(http.Flusher).Flush()
					conn, _, err := w.(http.Hijacker).Hijack()
					if err == nil {
						_ = conn.Close() // Abort chunked HTTP without the final zero chunk.
					}
				}))
				defer server.Close()
				up := &streamBillingOpenAIUpstream{url: server.URL}
				s, a := streamBillingOpenAIService(up, passthrough)
				c, rec := streamBillingOpenAIContext(context.Background())
				result, err := streamBillingOpenAIForward(s, c, a)
				require.ErrorIs(t, err, io.ErrUnexpectedEOF)
				var failover *UpstreamFailoverError
				require.False(t, errors.As(err, &failover))
				require.Contains(t, rec.Body.String(), "paid output")
				if observed {
					require.NotNil(t, result)
					require.Equal(t, 21, result.Usage.InputTokens)
					require.Equal(t, 6, result.Usage.CacheReadInputTokens)
					require.Equal(t, "gpt-5.2-observed", result.UpstreamResponseModel)
					require.False(t, result.ClientDisconnect, "upstream failure is not a downstream disconnect")
				} else {
					require.Nil(t, result, "no observed usage must not produce a phantom partial bill")
				}
				require.Equal(t, 1, up.calls)
				require.EqualValues(t, 1, up.body.closes.Load())
			})
		}
	}
}

func TestStreamBillingOpenAILoopbackDrainTimeout(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, observed := range []bool{false, true} {
			t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough]+map[bool]string{false: "/no_first_byte", true: "/partial_usage"}[observed], func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				exited := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer close(exited)
					w.Header().Set("Content-Type", "text/event-stream")
					if observed {
						_, _ = io.WriteString(w, streamBillingOpenAIPartial)
					}
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				}))
				defer server.Close()
				up := &streamBillingOpenAIUpstream{url: server.URL, after: cancel}
				s, a := streamBillingOpenAIService(up, passthrough)
				s.cfg.Gateway.StreamDataIntervalTimeout = 1
				c, _ := streamBillingOpenAIContext(ctx)
				start := time.Now()
				result, err := streamBillingOpenAIForward(s, c, a)
				require.Error(t, err)
				require.Less(t, time.Since(start), 3*time.Second)
				require.GreaterOrEqual(t, time.Since(start), time.Second)
				var failover *UpstreamFailoverError
				require.False(t, errors.As(err, &failover))
				if observed {
					require.NotNil(t, result)
					require.Equal(t, 21, result.Usage.InputTokens)
					require.True(t, result.ClientDisconnect)
				} else {
					require.Nil(t, result)
				}
				require.EqualValues(t, 1, up.body.closes.Load())
				select {
				case <-exited:
				case <-time.After(time.Second):
					t.Fatal("timeout did not close the stalled loopback body")
				}
			})
		}
	}
}

func TestStreamBillingOpenAIMappedToolBodyClosesOnce(t *testing.T) {
	c, rec := streamBillingOpenAIContext(context.Background())
	setOpenAIResponsesClientToolMapping(c, apicompat.ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	body := &streamBillingOpenAIBody{ReadCloser: io.NopCloser(strings.NewReader(streamBillingOpenAITerminal))}
	resp := &http.Response{Body: body, Header: make(http.Header)}
	s := &OpenAIGatewayService{cfg: &config.Config{}}
	result, err := s.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, &Account{ID: 123}, time.Now(), "gpt-5.2", "gpt-5.2")
	require.NoError(t, err)
	require.Equal(t, 21, result.usage.InputTokens)
	require.Contains(t, rec.Body.String(), "response.completed")
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 1, body.closes.Load())
}

type streamBillingOpenAIFailedWriter struct {
	gin.ResponseWriter
	writes int
	cancel func()
}

func (w *streamBillingOpenAIFailedWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.cancel != nil {
		w.cancel()
	}
	return 0, io.ErrClosedPipe
}
func (w *streamBillingOpenAIFailedWriter) WriteString(p string) (int, error) {
	return w.Write([]byte(p))
}

func TestStreamBillingOpenAIWriterDisconnectDrainsAndCloses(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "responses", true: "passthrough"}[passthrough], func(t *testing.T) {
			pr, pw := io.Pipe()
			defer pw.Close()
			failed := make(chan struct{})
			up := &streamBillingOpenAIUpstream{body: &streamBillingOpenAIBody{ReadCloser: pr}}
			s, a := streamBillingOpenAIService(up, passthrough)
			c, rec := streamBillingOpenAIContext(context.Background())
			var signal sync.Once
			writer := &streamBillingOpenAIFailedWriter{ResponseWriter: c.Writer, cancel: func() { signal.Do(func() { close(failed) }) }}
			c.Writer = writer
			producer := make(chan struct{})
			go func() {
				defer close(producer)
				_, _ = io.WriteString(pw, streamBillingOpenAIOutput)
				<-failed
				_, _ = io.WriteString(pw, streamBillingOpenAITerminal)
				// Keep the producer open: terminal must unblock the reader itself.
			}()
			result, err := streamBillingOpenAIForward(s, c, a)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 21, result.Usage.InputTokens)
			require.NotNil(t, result.FirstTokenMs)
			require.Equal(t, 1, writer.writes, "writes must stop immediately after disconnect")
			require.Empty(t, rec.Body.String())
			require.EqualValues(t, 1, up.body.closes.Load())
			select {
			case <-producer:
			case <-time.After(time.Second):
				t.Fatal("terminal reader cleanup left the producer blocked")
			}
		})
	}
}

func TestStreamBillingOpenAIReaderBoundsAndExit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		gap     time.Duration
		idle    time.Duration
		total   time.Duration
		drain   bool
		trickle bool
		want    error
	}{
		{"connected_positive_gap", 20 * time.Millisecond, time.Second, time.Second, false, false, errOpenAIStreamIdle},
		{"disconnected_zero_gap", 0, 20 * time.Millisecond, time.Second, true, false, errOpenAIStreamIdle},
		{"disconnected_positive_gap", 35 * time.Millisecond, time.Millisecond, time.Second, true, false, errOpenAIStreamIdle},
		{"total_deadline_despite_trickle", 0, 50 * time.Millisecond, 35 * time.Millisecond, true, true, errOpenAIStreamDrain},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pr, pw := io.Pipe()
			defer pw.Close()
			body := &streamBillingOpenAIBody{ReadCloser: pr}
			resp := &http.Response{Body: body}
			reader := newOpenAIStreamLineReader(resp, defaultMaxLineSize, tc.gap, tc.idle, tc.total)
			defer reader.stop()
			if tc.drain {
				reader.enableDrain()
			}
			producerDone := make(chan struct{})
			if tc.trickle {
				go func() {
					defer close(producerDone)
					ticker := time.NewTicker(2 * time.Millisecond)
					defer ticker.Stop()
					for range ticker.C {
						if _, err := io.WriteString(pw, ": upstream heartbeat\n"); err != nil {
							return
						}
					}
				}()
			}
			start := time.Now()
			for {
				_, err := reader.next(context.Background(), nil)
				if err != nil {
					require.ErrorIs(t, err, tc.want)
					break
				}
			}
			require.Less(t, time.Since(start), time.Second)
			if tc.name == "disconnected_positive_gap" {
				require.GreaterOrEqual(t, time.Since(start), tc.gap, "positive configured gap must win over the fallback idle")
			}
			// Exercise the same cleanup as both stream helpers, including the
			// outer Forward defer; inspect the joined reader before returning.
			reader.stop()
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 1, body.closes.Load())
			select {
			case <-reader.exited:
			default:
				t.Fatal("reader goroutine outlived stop")
			}
			if tc.trickle {
				select {
				case <-producerDone:
				case <-time.After(time.Second):
					t.Fatal("body close failed to release the producer")
				}
			}
		})
	}
}

func TestStreamBillingOpenAIConnectedZeroGapAndCancelWithoutBytes(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()
	body := &streamBillingOpenAIBody{ReadCloser: pr}
	reader := newOpenAIStreamLineReader(&http.Response{Body: body}, defaultMaxLineSize, 0, 20*time.Millisecond, time.Second)
	defer reader.stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	returned := make(chan error, 1)
	go func() { _, err := reader.next(ctx, nil); returned <- err }()
	select {
	case err := <-returned:
		t.Fatalf("zero gap must remain disabled while connected: %v", err)
	case <-time.After(40 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-returned:
		require.ErrorIs(t, err, errOpenAIStreamClientGone)
	case <-time.After(time.Second):
		t.Fatal("cancellation before the first byte was not observed")
	}
	_, err := reader.next(ctx, nil)
	require.ErrorIs(t, err, errOpenAIStreamIdle)
}
