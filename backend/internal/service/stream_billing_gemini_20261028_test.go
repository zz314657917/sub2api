package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type streamBillingGeminiWriter struct {
	gin.ResponseWriter
	failMatch         string
	failOccurrence    int
	occurrences       int
	partial           bool
	failed            bool
	afterFailure      int
	committed         bool
	shortWithoutError bool
}

func (w *streamBillingGeminiWriter) Committed() bool { return w.committed }
func (w *streamBillingGeminiWriter) Write(p []byte) (int, error) {
	if w.failed {
		w.afterFailure++
		return 0, io.ErrClosedPipe
	}
	if w.failMatch != "" && strings.Contains(string(p), w.failMatch) {
		w.occurrences++
		if w.occurrences == w.failOccurrence {
			w.failed = true
			if w.partial {
				n, _ := w.ResponseWriter.Write(p[:len(p)/2])
				if w.shortWithoutError {
					return n, nil
				}
				return n, io.ErrClosedPipe
			}
			return 0, io.ErrClosedPipe
		}
	}
	return w.ResponseWriter.Write(p)
}
func (w *streamBillingGeminiWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }

func TestStreamBillingGeminiImageDelivery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sse     string
		failAt  int
		partial bool
		bill    bool
		count   int
	}{
		{"zero_byte_failure", streamBillingGeminiImage, 1, false, false, 0},
		{"partial_write", streamBillingGeminiImage, 1, true, false, 0},
		{"delivered_then_zero_failure", streamBillingGeminiImage + streamBillingGeminiImage, 2, false, true, 1},
		{"delivered_then_partial_failure", streamBillingGeminiImage + streamBillingGeminiImage, 2, true, true, 1},
		{"usage_then_failed_image", streamBillingGeminiUsage + streamBillingGeminiImage, 1, false, true, 0},
		{"usage_then_partial_image", streamBillingGeminiUsage + streamBillingGeminiImage, 1, true, true, 0},
		{"final_failed_image", streamBillingGeminiUsage + strings.Replace(streamBillingGeminiImage, `{"content":`, `{"finishReason":"STOP","content":`, 1), 1, false, true, 0},
		{"delivered_image", streamBillingGeminiImage, 0, false, true, 1},
		{"no_image_fallback_unfinished", streamBillingGeminiUsage, 0, false, true, 0},
		{"finished_image_fallback", streamBillingGeminiUsage + streamBillingGeminiFinal, 0, false, true, 1},
		{"URI_A1", strings.ReplaceAll(strings.ReplaceAll(streamBillingGeminiImage, "inlineData", "fileData"), `"data":"aGVsbG8="`, `"fileUri":"https://example.test/image.png"`), 0, false, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, c, stub, body, _ := streamBillingGeminiSetup(t, "native", "gemini-3-pro-image", tc.sse, context.Canceled, true)
			w := &streamBillingGeminiWriter{ResponseWriter: c.Writer, failMatch: "inlineData", failOccurrence: tc.failAt, partial: tc.partial}
			c.Writer = w
			result, err := streamBillingGeminiForward(svc, c, "native", "gemini-3-pro-image")
			require.Equal(t, 1, stub.calls)
			require.Equal(t, 1, body.closed)
			require.Zero(t, w.afterFailure, "no further client writes after failed delivery")
			if !tc.bill {
				require.Error(t, err)
				require.Nil(t, result)
				return
			}
			require.NoError(t, err)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, tc.count, result.ImageCount)
		})
	}
}

func TestStreamBillingGeminiCompatibilityDoesNotDeliverImages(t *testing.T) {
	for _, route := range []string{"messages", "chat"} {
		for _, withUsage := range []bool{false, true} {
			t.Run(route+map[bool]string{false: "/without_usage", true: "/with_usage"}[withUsage], func(t *testing.T) {
				sse := streamBillingGeminiImage
				if withUsage {
					sse = streamBillingGeminiUsage + sse
				}
				svc, c, _, _, _ := streamBillingGeminiSetup(t, route, "gemini-3-pro-image", sse, context.Canceled, true)
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-3-pro-image")
				if !withUsage {
					require.Error(t, err)
					require.Nil(t, result)
					return
				}
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				require.Zero(t, result.ImageCount)
			})
		}
	}
}

func TestStreamBillingGeminiCompletedMessagesImageObservation(t *testing.T) {
	sse := strings.Replace(streamBillingGeminiImage, `{"inlineData":`, `{"fileData":{"mimeType":"image/png","fileUri":"gs://first"}},{"inlineData":`, 1)
	for _, readErr := range []error{io.EOF, context.Canceled} {
		for _, finished := range []bool{false, true} {
			name := readErr.Error() + map[bool]string{false: "/unfinished", true: "/finished"}[finished]
			t.Run(name, func(t *testing.T) {
				input := sse + streamBillingGeminiUsage
				if finished {
					input += streamBillingGeminiFinal
				}
				svc, c, _, _, _ := streamBillingGeminiSetup(t, "messages", "custom-alias", input, readErr, readErr != io.EOF)
				result, err := streamBillingGeminiForward(svc, c, "messages", "custom-alias")
				require.NoError(t, err)
				if readErr == io.EOF || finished {
					require.Equal(t, 2, result.ImageCount, "retain existing complete-stream observation")
				} else {
					require.Zero(t, result.ImageCount, "unfinished compatibility streams delivered no images")
				}
			})
		}
	}
}

func TestStreamBillingGeminiNormalCompletion(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		for _, finalLine := range []string{"\n\n", ""} {
			t.Run(route+map[bool]string{false: "/newline", true: "/EOF_tail"}[finalLine == ""], func(t *testing.T) {
				sse := streamBillingGeminiUsage + strings.TrimSuffix(streamBillingGeminiFinal, "\n\n") + finalLine
				svc, c, _, body, _ := streamBillingGeminiSetup(t, route, "gemini-3-pro-image", sse, io.EOF, false)
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-3-pro-image")
				require.NoError(t, err)
				require.False(t, result.ClientDisconnect)
				require.Equal(t, 7, result.Usage.InputTokens)
				require.Equal(t, 7, result.Usage.OutputTokens, "latest cumulative value, not sum")
				require.Equal(t, 3, result.Usage.CacheReadInputTokens)
				require.Equal(t, 1, result.ImageCount, "preserve normal image fallback")
				require.Equal(t, 1, body.closed)
			})
		}
	}
}

func TestStreamBillingGeminiFirstResponseUncommitted(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		for _, writerState := range []string{"uncommitted", "unknown", "committed"} {
			t.Run(route+"/"+writerState, func(t *testing.T) {
				svc, c, _, _, _ := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", streamBillingGeminiUsage, context.Canceled, true)
				c.Request = c.Request.WithContext(ctxkey.WithKeyRouteFirstResponseAttempt(c.Request.Context()))
				if writerState != "unknown" {
					c.Writer = &streamBillingGeminiWriter{ResponseWriter: c.Writer, committed: writerState == "committed"}
				}
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
				require.True(t, c.Writer.Written(), "buffered prelude alone must not establish commit")
				require.Positive(t, c.Writer.Size())
				if writerState != "committed" {
					require.Error(t, err)
					require.Nil(t, result)
					return
				}
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 2, result.Usage.OutputTokens)
			})
		}
	}
}

func TestStreamBillingGeminiCompatibilityWriteFailure(t *testing.T) {
	for _, route := range []string{"messages", "chat"} {
		for _, partial := range []bool{false, true} {
			t.Run(route+map[bool]string{false: "/zero", true: "/partial"}[partial], func(t *testing.T) {
				svc, c, _, body, _ := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", streamBillingGeminiUsage+streamBillingGeminiFinal, io.EOF, false)
				w := &streamBillingGeminiWriter{ResponseWriter: c.Writer, failMatch: "hello", failOccurrence: 1, partial: partial}
				c.Writer = w
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
				require.True(t, w.failed)
				require.Zero(t, w.afterFailure)
				require.False(t, body.ended, "stop reading immediately; Gemini has no detached drain")
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 2, result.Usage.OutputTokens, "snapshot from current frame must precede write")
				require.Equal(t, 1, body.closed)
			})
		}
	}
}

func TestStreamBillingGeminiShortWriteWithoutError(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		t.Run(route, func(t *testing.T) {
			sse, failMatch := streamBillingGeminiUsage, "hello"
			if route == "native" {
				sse += streamBillingGeminiImage
				failMatch = "inlineData"
			}
			svc, c, _, body, _ := streamBillingGeminiSetup(t, route, "gemini-3-pro-image", sse, io.EOF, false)
			w := &streamBillingGeminiWriter{ResponseWriter: c.Writer, failMatch: failMatch, failOccurrence: 1, partial: true, shortWithoutError: true}
			c.Writer = w
			result, err := streamBillingGeminiForward(svc, c, route, "gemini-3-pro-image")
			require.NoError(t, err)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 2, result.Usage.OutputTokens)
			require.Zero(t, result.ImageCount)
			require.Zero(t, w.afterFailure)
			require.Equal(t, 1, body.closed)
		})
	}
}

func TestStreamBillingGeminiUsageKinds(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		for _, tc := range []struct {
			name, usage string
			bill        bool
		}{
			{"input_only", `"promptTokenCount":10`, true},
			{"output_only", `"candidatesTokenCount":2`, true},
			{"cache_only", `"promptTokenCount":3,"cachedContentTokenCount":3`, true},
			{"image_tokens_only", `"candidatesTokensDetails":[{"modality":"IMAGE","tokenCount":5}]`, true},
			{"empty_usage", `"promptTokenCount":0,"candidatesTokenCount":0`, false},
		} {
			t.Run(route+"/"+tc.name, func(t *testing.T) {
				sse := `data: {"usageMetadata":{` + tc.usage + "}}\n\n"
				svc, c, _, _, _ := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", sse, context.Canceled, true)
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
				if !tc.bill {
					require.Error(t, err)
					require.Nil(t, result)
					return
				}
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				switch tc.name {
				case "input_only":
					require.Equal(t, 10, result.Usage.InputTokens)
				case "output_only":
					require.Equal(t, 2, result.Usage.OutputTokens)
				case "cache_only":
					require.Equal(t, 3, result.Usage.CacheReadInputTokens)
					require.Zero(t, result.Usage.InputTokens)
				case "image_tokens_only":
					require.Equal(t, 5, result.Usage.ImageOutputTokens)
				}
			})
		}
	}
}

type streamBillingGeminiHTTPUpstream struct {
	client *http.Client
	url    *url.URL
	closed atomic.Int32
	calls  atomic.Int32
}

type streamBillingGeminiHTTPBody struct {
	io.ReadCloser
	closed *atomic.Int32
}

func (b *streamBillingGeminiHTTPBody) Close() error {
	b.closed.Add(1)
	return b.ReadCloser.Close()
}

func (s *streamBillingGeminiHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.calls.Add(1)
	req = req.Clone(req.Context())
	req.URL = s.url
	resp, err := s.client.Do(req)
	if err == nil {
		resp.Body = &streamBillingGeminiHTTPBody{ReadCloser: resp.Body, closed: &s.closed}
	}
	return resp, err
}
func (s *streamBillingGeminiHTTPUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxy, id, concurrency)
}

type streamBillingGeminiCancelWriter struct {
	gin.ResponseWriter
	cancel context.CancelFunc
}

func (w *streamBillingGeminiCancelWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if n == len(p) && strings.Contains(string(p), "hello") {
		w.cancel()
	}
	return n, err
}
func (w *streamBillingGeminiCancelWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func TestStreamBillingGeminiLoopbackCancellation(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		for _, beforeFirstByte := range []bool{false, true} {
			t.Run(route+map[bool]string{false: "/after_usage", true: "/before_first_byte"}[beforeFirstByte], func(t *testing.T) {
				svc, c, _, _, cancel := streamBillingGeminiSetup(t, route, "gemini-2.5-flash", "", io.EOF, false)
				boundedCtx, release := context.WithTimeout(c.Request.Context(), 3*time.Second)
				defer release()
				c.Request = c.Request.WithContext(boundedCtx)
				serverCanceled := make(chan struct{})
				headersSent := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.Copy(io.Discard, r.Body)
					w.Header().Set("Content-Type", "text/event-stream")
					w.WriteHeader(http.StatusOK)
					if !beforeFirstByte {
						_, _ = io.WriteString(w, streamBillingGeminiUsage)
					}
					w.(http.Flusher).Flush()
					close(headersSent)
					<-r.Context().Done()
					close(serverCanceled)
				}))
				defer server.Close()
				t.Cleanup(server.Client().CloseIdleConnections)
				u, err := url.Parse(server.URL)
				require.NoError(t, err)
				upstream := &streamBillingGeminiHTTPUpstream{client: server.Client(), url: u}
				svc.httpUpstream = upstream
				c.Writer = &streamBillingGeminiCancelWriter{ResponseWriter: c.Writer, cancel: cancel}
				cancelDone := make(chan struct{})
				if beforeFirstByte {
					go func() { <-headersSent; cancel(); close(cancelDone) }()
				} else {
					close(cancelDone)
				}
				result, err := streamBillingGeminiForward(svc, c, route, "gemini-2.5-flash")
				<-cancelDone
				if beforeFirstByte {
					require.Error(t, err)
					require.Nil(t, result)
				} else {
					require.NoError(t, err)
					require.True(t, result.ClientDisconnect)
					require.Equal(t, 2, result.Usage.OutputTokens)
					require.Equal(t, 1, int(upstream.closed.Load()))
				}
				require.Equal(t, 1, int(upstream.calls.Load()))
				select {
				case <-serverCanceled:
				case <-time.After(2 * time.Second):
					t.Fatal("upstream request outlived client cancellation")
				}
			})
		}
	}
}

func TestStreamBillingGeminiWriteFailureWithoutUsage(t *testing.T) {
	for _, route := range []string{"native", "messages", "chat"} {
		t.Run(route, func(t *testing.T) {
			svc, c, _, body, _ := streamBillingGeminiSetup(t, route, "gemini-3-pro-image", streamBillingGeminiText, io.EOF, false)
			w := &streamBillingGeminiWriter{ResponseWriter: c.Writer, failMatch: "hello", failOccurrence: 1}
			c.Writer = w
			result, err := streamBillingGeminiForward(svc, c, route, "gemini-3-pro-image")
			require.ErrorIs(t, err, io.ErrClosedPipe)
			require.Nil(t, result)
			require.Zero(t, w.afterFailure)
			require.Equal(t, 1, body.closed)
		})
	}
}
