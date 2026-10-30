package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type streamBillingDisconnectWriter struct {
	gin.ResponseWriter
	writes int
}

func (w *streamBillingDisconnectWriter) Write(p []byte) (int, error) {
	w.writes++
	return 0, io.ErrClosedPipe
}

func (w *streamBillingDisconnectWriter) WriteString(p string) (int, error) {
	return w.Write([]byte(p))
}

type streamBillingLoopbackUpstream struct {
	target *url.URL
	body   *streamBillingCountBody
}

func (u *streamBillingLoopbackUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	req = req.Clone(req.Context())
	copiedURL := *req.URL
	req.URL = &copiedURL
	req.URL.Scheme, req.URL.Host = u.target.Scheme, u.target.Host
	req.Host = u.target.Host
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		u.body = &streamBillingCountBody{ReadCloser: resp.Body}
		resp.Body = u.body
	}
	return resp, err
}

func (u *streamBillingLoopbackUpstream) DoWithTLS(req *http.Request, proxy string, account int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, account, concurrency)
}

func TestStreamBillingAnthropicLoopbackDrain(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		for _, mode := range []string{"failed_write", "cancel_before_first_byte", "cancel_after_usage"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				serverExited := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer close(serverExited)
					w.Header().Set("Content-Type", "text/event-stream")
					w.Header().Set("X-Request-Id", "rid-loopback")
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
					if mode == "cancel_before_first_byte" {
						cancel()
					}
					_, _ = io.WriteString(w, streamBillingAnthropicStart)
					w.(http.Flusher).Flush()
					if mode == "cancel_after_usage" {
						cancel()
					}
					_, _ = io.WriteString(w, streamBillingAnthropicDelta+streamBillingAnthropicStop)
					w.(http.Flusher).Flush()
					// Do not send EOF. The terminal must close the network body and reader.
					<-r.Context().Done()
				}))
				defer server.Close()
				target, err := url.Parse(server.URL)
				require.NoError(t, err)
				upstream := &streamBillingLoopbackUpstream{target: target}
				svc := streamBillingAnthropicService(nil)
				svc.httpUpstream = upstream
				c, rec := streamBillingAnthropicContext(ctx)
				var failed *streamBillingDisconnectWriter
				if mode == "failed_write" {
					failed = &streamBillingDisconnectWriter{ResponseWriter: c.Writer}
					c.Writer = failed
				}
				result, err := streamBillingAnthropicForward(t, kind, svc, c, ctx)
				require.NoError(t, err)
				require.NotNil(t, result)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 11, result.Usage.InputTokens)
				require.Equal(t, 9, result.Usage.OutputTokens)
				require.Equal(t, 7, result.Usage.CacheReadInputTokens)
				require.Equal(t, 2, result.Usage.CacheCreation5mTokens)
				require.Equal(t, 3, result.Usage.CacheCreation1hTokens)
				require.EqualValues(t, 1, upstream.body.closes.Load())
				require.Positive(t, upstream.body.reads.Load())
				if failed != nil {
					require.Equal(t, 1, failed.writes, "no second write after a failed write")
				}
				if mode == "cancel_before_first_byte" {
					require.Empty(t, rec.Body.String())
				}
				select {
				case <-serverExited:
				case <-time.After(2 * time.Second):
					t.Fatal("upstream connection/reader outlived terminal")
				}
			})
		}
	}
}

func TestStreamBillingAnthropicRealHTTPClientAbort(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		t.Run(kind, func(t *testing.T) {
			clientGone := make(chan struct{})
			providerExited := make(chan struct{})
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(providerExited)
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, streamBillingAnthropicStart)
				w.(http.Flusher).Flush()
				select {
				case <-clientGone:
				case <-r.Context().Done():
					return
				}
				_, _ = io.WriteString(w, streamBillingAnthropicDelta+streamBillingAnthropicStop)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer provider.Close()
			target, err := url.Parse(provider.URL)
			require.NoError(t, err)
			upstream := &streamBillingLoopbackUpstream{target: target}
			svc := streamBillingAnthropicService(nil)
			svc.httpUpstream = upstream
			type outcome struct {
				result *ForwardResult
				err    error
			}
			completed := make(chan outcome, 1)
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, _ := gin.CreateTestContext(w)
				c.Request = r
				watcherExited := make(chan struct{})
				go func() {
					defer close(watcherExited)
					<-r.Context().Done()
					close(clientGone)
				}()
				result, err := streamBillingAnthropicForward(t, kind, svc, c, r.Context())
				<-watcherExited
				completed <- outcome{result: result, err: err}
			}))
			defer gateway.Close()
			resp, err := http.Get(gateway.URL)
			require.NoError(t, err)
			// Close a real downstream socket before consuming the streamed response.
			require.NoError(t, resp.Body.Close())
			select {
			case got := <-completed:
				require.NoError(t, got.err)
				require.NotNil(t, got.result)
				require.True(t, got.result.ClientDisconnect)
				require.Equal(t, 11, got.result.Usage.InputTokens)
				require.Equal(t, 9, got.result.Usage.OutputTokens)
				require.EqualValues(t, 1, upstream.body.closes.Load())
			case <-time.After(3 * time.Second):
				gateway.CloseClientConnections()
				provider.CloseClientConnections()
				t.Fatal("real downstream abort failed to drain and release the provider")
			}
			select {
			case <-providerExited:
			case <-time.After(time.Second):
				t.Fatal("provider outlived Forward")
			}
		})
	}
}

type streamBillingCancelOnWrite struct {
	*streamBillingPreludeWriter
	cancel context.CancelFunc
}

func (w *streamBillingCancelOnWrite) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.cancel()
	return n, err
}

func (w *streamBillingCancelOnWrite) WriteString(p string) (int, error) {
	return w.Write([]byte(p))
}

func TestStreamBillingAnthropicFirstResponseAfterUsageCancellation(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		for _, committed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/committed=%t", kind, committed), func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
				defer cancel()
				c, _ := streamBillingAnthropicContext(ctx)
				c.Writer = &streamBillingCancelOnWrite{streamBillingPreludeWriter: &streamBillingPreludeWriter{ResponseWriter: c.Writer, committed: committed}, cancel: cancel}
				body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(streamBillingAnthropicStart + streamBillingAnthropicDelta + streamBillingAnthropicStop))}
				result, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, ctx)
				if committed {
					require.NoError(t, err)
					require.NotNil(t, result)
					require.True(t, result.ClientDisconnect)
					require.Equal(t, 9, result.Usage.OutputTokens)
				} else {
					require.ErrorIs(t, err, context.Canceled)
					require.Nil(t, result)
				}
				require.EqualValues(t, 1, body.closes.Load())
			})
		}
	}
}

type streamBillingBlockedBody struct {
	payload  io.Reader
	closed   chan struct{}
	returned chan struct{}
	once     sync.Once
	closes   atomic.Int32
}

func newStreamBillingBlockedBody(payload string) *streamBillingBlockedBody {
	return &streamBillingBlockedBody{payload: strings.NewReader(payload), closed: make(chan struct{}), returned: make(chan struct{})}
}

func (b *streamBillingBlockedBody) Read(p []byte) (int, error) {
	if n, err := b.payload.Read(p); n > 0 || err != io.EOF {
		return n, err
	}
	<-b.closed
	b.once.Do(func() { close(b.returned) })
	return 0, io.ErrClosedPipe
}

func (b *streamBillingBlockedBody) Close() error {
	if b.closes.Add(1) == 1 {
		close(b.closed)
	}
	return nil
}

func TestStreamBillingAnthropicForwardCancellationIdleBound(t *testing.T) {
	for _, kind := range []string{"native", "passthrough", "chat", "responses"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Includes cancellation before the first upstream byte.
			body := newStreamBillingBlockedBody(streamBillingAnthropicStart)
			svc := streamBillingAnthropicService(body)
			svc.cfg.Gateway.StreamDataIntervalTimeout = 1
			c, rec := streamBillingAnthropicContext(ctx)
			result, err := streamBillingAnthropicForward(t, kind, svc, c, ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), "timeout")
			require.NotNil(t, result)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 11, result.Usage.InputTokens)
			require.Empty(t, rec.Body.String())
			require.EqualValues(t, 1, body.closes.Load())
			select {
			case <-body.returned:
			default:
				t.Fatal("Forward returned while the reader was still blocked")
			}
		})
	}
}

func TestStreamBillingAnthropicPumpBounds(t *testing.T) {
	t.Run("connected_zero_disconnected_idle", func(t *testing.T) {
		body := newStreamBillingBlockedBody("")
		pump := newGatewayAnthropicLinePump(body, 4096, 0, 20*time.Millisecond, time.Second)
		st := &anthropicCompatBillingStream{pump: pump}
		require.Nil(t, pump.idleC(), "zero gap remains disabled while connected")
		st.disconnect()
		_, err := st.next()
		require.ErrorIs(t, err, errGatewayAnthropicIdle)
		pump.stop()
		require.EqualValues(t, 1, body.closes.Load())
		select {
		case <-pump.exited:
		default:
			t.Fatal("reader did not exit")
		}
	})
	t.Run("connected_positive", func(t *testing.T) {
		body := newStreamBillingBlockedBody("")
		pump := newGatewayAnthropicLinePump(body, 4096, 15*time.Millisecond, time.Second, time.Second)
		st := &anthropicCompatBillingStream{pump: pump}
		_, err := st.next()
		require.ErrorIs(t, err, errGatewayAnthropicIdle)
		pump.stop()
		require.EqualValues(t, 1, body.closes.Load())
	})
	t.Run("total_deadline_with_continuous_input", func(t *testing.T) {
		reader, writer := io.Pipe()
		body := &streamBillingCountBody{ReadCloser: reader}
		pump := newGatewayAnthropicLinePump(body, 4096, time.Second, time.Second, 30*time.Millisecond)
		st := &anthropicCompatBillingStream{pump: pump}
		st.disconnect()
		writerExited := make(chan struct{})
		go func() {
			defer close(writerExited)
			for {
				if _, err := io.WriteString(writer, ": ping\n"); err != nil {
					return
				}
			}
		}()
		for {
			_, err := st.next()
			if err != nil {
				require.ErrorIs(t, err, errGatewayAnthropicDrain)
				break
			}
		}
		pump.stop()
		_ = writer.Close()
		<-writerExited
		require.EqualValues(t, 1, body.closes.Load())
	})
	t.Run("production_limits", func(t *testing.T) {
		for _, gap := range []int{0, 2} {
			svc := streamBillingAnthropicService(nil)
			svc.cfg.Gateway.StreamDataIntervalTimeout = gap
			resp := &http.Response{Body: io.NopCloser(strings.NewReader(""))}
			pump := svc.newAnthropicBillingPump(resp)
			require.Equal(t, 180*time.Second, pump.drainLimit)
			require.Equal(t, 180*time.Second, pump.drainIdle)
			pump.disconnect()
			want := 180 * time.Second
			if gap > 0 {
				want = time.Duration(gap) * time.Second
			}
			require.Equal(t, want, pump.interval)
			pump.stop()
		}
	})
}

func TestStreamBillingAnthropicCompatFailureProtocol(t *testing.T) {
	for _, kind := range []string{"chat", "responses"} {
		for _, payload := range []string{streamBillingAnthropicStart, streamBillingAnthropicStart + "event: error\ndata: {\"type\":\"error\",\"error\":{\"message\":\"busy\"}}\n\n"} {
			t.Run(kind+"/"+fmt.Sprint(len(payload)), func(t *testing.T) {
				body := &streamBillingCountBody{ReadCloser: io.NopCloser(strings.NewReader(payload))}
				c, rec := streamBillingAnthropicContext(context.Background())
				_, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, c.Request.Context())
				require.Error(t, err)
				require.NotContains(t, rec.Body.String(), "response.completed")
				require.NotContains(t, rec.Body.String(), "[DONE]")
				if !strings.Contains(payload, `"type":"error"`) {
					if kind == "responses" {
						require.Contains(t, rec.Body.String(), "response.failed")
						require.Contains(t, rec.Body.String(), `"id":"msg_billing"`)
					} else {
						require.Contains(t, rec.Body.String(), `"error"`)
					}
				}
				require.EqualValues(t, 1, body.closes.Load())
			})
		}
	}
}

func TestStreamBillingAnthropicCanceledAttemptContextPredicate(t *testing.T) {
	ctx, cancel := context.WithCancel(ctxkey.WithKeyRouteFirstResponseAttempt(context.Background()))
	c, _ := streamBillingAnthropicContext(ctx)
	writer := &streamBillingPreludeWriter{ResponseWriter: c.Writer}
	c.Writer = writer
	_, _ = writer.WriteString(": buffered\n\n")
	require.True(t, writer.Written())
	require.False(t, StreamBillingFirstResponseUncommitted(ctx, c))
	cancel()
	require.True(t, StreamBillingFirstResponseUncommitted(ctx, c))
	writer.committed = true
	require.False(t, StreamBillingFirstResponseUncommitted(ctx, c))
	require.True(t, errors.Is(ctx.Err(), context.Canceled))
}

func TestStreamBillingAnthropicR2DisconnectedInbandError(t *testing.T) {
	const inbandError = "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"busy\"}}\n\n"
	for _, kind := range []string{"chat", "responses"} {
		for _, mode := range []string{"connected", "failed_write", "canceled"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if mode == "canceled" {
					cancel()
				}
				c, rec := streamBillingAnthropicContext(ctx)
				var failed *streamBillingDisconnectWriter
				if mode == "failed_write" {
					failed = &streamBillingDisconnectWriter{ResponseWriter: c.Writer}
					c.Writer = failed
				}
				// The reader would hang after this error without explicit body closure.
				body := newStreamBillingBlockedBody(streamBillingAnthropicStart + streamBillingAnthropicDelta + inbandError)
				result, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, ctx)
				require.Error(t, err)
				require.EqualValues(t, 1, body.closes.Load(), "actual Forward closes the reader exactly once")
				require.NotContains(t, rec.Body.String(), "[DONE]")
				require.NotContains(t, rec.Body.String(), "response.completed")
				var failover *UpstreamFailoverError
				if mode == "connected" {
					require.False(t, errors.As(err, &failover))
					require.NotNil(t, result, "the ordinary writer already committed protocol output and observed usage")
					return
				}
				require.False(t, errors.As(err, &failover), "an aborted client's observed usage must not become a retry")
				require.NotNil(t, result)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 11, result.Usage.InputTokens)
				require.Equal(t, 9, result.Usage.OutputTokens)
				require.Equal(t, 7, result.Usage.CacheReadInputTokens)
				require.Equal(t, 2, result.Usage.CacheCreation5mTokens)
				require.Equal(t, 3, result.Usage.CacheCreation1hTokens)
				require.Equal(t, "rid-billing", result.RequestID)
				var upstreamError *sseStreamErrorEventError
				require.ErrorAs(t, err, &upstreamError)
				require.Contains(t, upstreamError.RawData, "overloaded_error")
				if failed != nil {
					require.Equal(t, 1, failed.writes, "the in-band error must not write to the disconnected client")
				} else {
					require.Empty(t, rec.Body.String())
				}
			})
		}
	}
}

func TestStreamBillingAnthropicR2EOFUsageTail(t *testing.T) {
	const priorDelta = "event: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":2}}\n\n"
	const deltaTail = "event: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":13}}"
	const terminalTail = "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":13}}"
	cases := []struct {
		name       string
		prefix     string
		tail       string
		readErr    error
		wantInput  int
		wantOutput int
		terminal   bool
	}{
		{name: "complete_eof_no_newline", tail: deltaTail, wantInput: 11, wantOutput: 13},
		{name: "complete_eof_line_newline", tail: deltaTail + "\n", wantInput: 11, wantOutput: 13},
		{name: "complete_eof_data_only", tail: strings.TrimPrefix(deltaTail, "event: message_delta\n"), wantInput: 11, wantOutput: 13},
		{name: "only_message_start_eof", prefix: "empty", tail: strings.TrimSuffix(streamBillingAnthropicStart, "\n\n"), wantInput: 11},
		{name: "invalid_json_eof", tail: deltaTail[:len(deltaTail)-1], wantInput: 11, wantOutput: 2},
		{name: "complete_non_eof_read_error", tail: deltaTail, readErr: io.ErrUnexpectedEOF, wantInput: 11, wantOutput: 2},
		{name: "invalid_json_non_eof", tail: deltaTail[:len(deltaTail)-1], readErr: io.ErrUnexpectedEOF, wantInput: 11, wantOutput: 2},
		{name: "terminal_delta_eof", tail: terminalTail, wantInput: 11, wantOutput: 13, terminal: true},
		{name: "terminal_stop_eof", tail: strings.TrimSuffix(streamBillingAnthropicStop, "\n\n"), wantInput: 11, wantOutput: 2, terminal: true},
		{name: "terminal_done_eof", tail: "data: [DONE]", wantInput: 11, wantOutput: 2, terminal: true},
		{name: "terminal_non_eof_read_error", tail: terminalTail, readErr: io.ErrUnexpectedEOF, wantInput: 11, wantOutput: 2},
	}
	for _, kind := range []string{"chat", "responses"} {
		for _, tc := range cases {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				prefix := streamBillingAnthropicStart + priorDelta
				if tc.prefix == "empty" {
					prefix = ""
				}
				var reader io.ReadCloser = io.NopCloser(strings.NewReader(prefix + tc.tail))
				if tc.readErr != nil {
					reader = &streamReadCloser{payload: []byte(prefix + tc.tail), err: tc.readErr}
				}
				body := &streamBillingCountBody{ReadCloser: reader}
				c, rec := streamBillingAnthropicContext(context.Background())
				result, err := streamBillingAnthropicForward(t, kind, streamBillingAnthropicService(body), c, c.Request.Context())
				if tc.terminal {
					require.NoError(t, err)
					if kind == "chat" {
						require.Contains(t, rec.Body.String(), "data: [DONE]")
					} else {
						require.Contains(t, rec.Body.String(), "response.completed")
					}
				} else {
					require.Error(t, err)
					if tc.readErr != nil {
						require.ErrorIs(t, err, tc.readErr)
					} else {
						require.ErrorIs(t, err, io.EOF)
					}
					require.NotContains(t, rec.Body.String(), "data: [DONE]")
					require.NotContains(t, rec.Body.String(), "response.completed")
				}
				require.EqualValues(t, 1, body.closes.Load(), "actual Forward closes the reader exactly once")
				require.NotNil(t, result)
				require.Equal(t, tc.wantInput, result.Usage.InputTokens)
				require.Equal(t, tc.wantOutput, result.Usage.OutputTokens)
				require.Equal(t, 7, result.Usage.CacheReadInputTokens)
				require.Equal(t, 2, result.Usage.CacheCreation5mTokens)
				require.Equal(t, 3, result.Usage.CacheCreation1hTokens)
				require.Equal(t, "rid-billing", result.RequestID)
				require.False(t, result.ClientDisconnect)
			})
		}
	}
}
