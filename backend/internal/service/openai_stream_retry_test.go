package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const streamRetryFailed = "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"Upstream request failed\"}}}\n\n"
const streamRetrySuccess = "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_success\"}}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_success\",\"output\":[]}}\n\n"

func runStreamRetryProcessor(s *OpenAIGatewayService, c *gin.Context, body io.ReadCloser, passthrough bool) error {
	resp := &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	if passthrough {
		_, err := s.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
		return err
	}
	_, err := s.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
	return err
}

func streamRetryContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	return c, rec
}

func TestOpenAIStreamRetryPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, payload       string
		pool, generic, same bool
	}{
		{name: "disconnect", generic: true},
		{name: "generic", payload: `{"response":{"error":{"message":"Upstream request failed"}}}`, generic: true},
		{name: "pool_generic", pool: true},
		{name: "capacity", payload: `{"error":{"message":"Selected model is at capacity. Please try a different model."}}`, same: true},
		{name: "overloaded", payload: `{"error":{"code":"server_is_overloaded"}}`, same: true},
		{name: "rate_limit", payload: `{"error":{"code":"rate_limit_exceeded"}}`},
		{name: "pool_rate_limit", payload: `{"error":{"code":"rate_limit_exceeded"}}`, pool: true, same: true},
		{name: "auth", payload: `{"error":{"code":"invalid_api_key"}}`},
		{name: "account_state", payload: `{"error":{"code":"account_disabled"}}`},
		{name: "quota", payload: `{"error":{"code":"insufficient_quota"}}`},
		{name: "model", payload: `{"error":{"code":"model_not_found"}}`},
		{name: "invalid_parameter", payload: `{"error":{"code":"invalid_parameter"}}`},
		{name: "bad_request", payload: `{"error":{"type":"invalid_request_error"}}`},
		{name: "403", payload: `{"error":{"status_code":403}}`},
		{name: "409", payload: `{"error":{"status_code":409}}`},
		{name: "413", payload: `{"error":{"status_code":413}}`},
		{name: "policy", payload: `{"error":{"message":"content_policy violation"}}`},
		{name: "context", payload: `{"error":{"code":"context_length_exceeded","message":"maximum context length"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := streamRetryContext()
			s := &OpenAIGatewayService{}
			account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"pool_mode": tc.pool}}
			msg := extractOpenAISSEErrorMessage([]byte(tc.payload))
			err := s.newResponsesStreamFailoverError(c, account, false, "", []byte(tc.payload), msg)
			var tagged *OpenAIStreamRetryError
			require.Equal(t, tc.generic, errors.As(err, &tagged))
			var base *UpstreamFailoverError
			require.ErrorAs(t, err, &base)
			require.Equal(t, tc.same, base.RetryableOnSameAccount)
			// Shared constructor remains unchanged for buffered/non-stream adapters.
			shared := s.newOpenAIStreamFailoverError(c, account, false, "", nil, "Upstream request failed")
			require.False(t, errors.As(shared, &tagged))
		})
	}
	require.Equal(t, []time.Duration{500 * time.Millisecond, time.Second, 1500 * time.Millisecond},
		[]time.Duration{OpenAIStreamRetryDelay(1), OpenAIStreamRetryDelay(2), OpenAIStreamRetryDelay(3)})
}

func TestOpenAIStreamRetryPreamble(t *testing.T) {
	preamble := strings.Repeat("data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 100)
	require.Greater(t, len(preamble), 4096)
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for _, ending := range []string{"missing_terminal", "read_error", "response_failed", "comment"} {
			t.Run(mode+"/"+ending, func(t *testing.T) {
				c, rec := streamRetryContext()
				cfg := &config.Config{}
				cfg.Gateway.MaxLineSize = 64 * 1024
				if mode == "ticker" {
					cfg.Gateway.StreamKeepaliveInterval = 1
				}
				s := &OpenAIGatewayService{cfg: cfg}
				body := preamble
				if ending == "response_failed" || ending == "comment" {
					body += streamRetryFailed
				}
				if ending == "comment" {
					body = ": upstream heartbeat\n\n" + body
				}
				var reader io.Reader = strings.NewReader(body)
				if ending == "read_error" {
					reader = io.MultiReader(reader, errReadCloser{err: io.ErrUnexpectedEOF})
				}
				err := runStreamRetryProcessor(s, c, io.NopCloser(reader), mode == "passthrough")
				var tagged *OpenAIStreamRetryError
				require.ErrorAs(t, err, &tagged)
				require.NotContains(t, rec.Body.String(), "resp_failed")
				require.NotContains(t, rec.Body.String(), "response.failed")
				require.Equal(t, -1, OpenAIStreamWrittenSize(c))
				require.NoError(t, runStreamRetryProcessor(s, c, io.NopCloser(strings.NewReader(streamRetrySuccess)), mode == "passthrough"))
				require.Contains(t, rec.Body.String(), "resp_success")
				require.NotContains(t, rec.Body.String(), "resp_failed")
			})
		}
	}
}

func TestOpenAIStreamRetryPreambleBound(t *testing.T) {
	preamble := strings.Repeat("data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 10)
	for _, passthrough := range []bool{false, true} {
		for _, extra := range []string{"", "\n"} {
			t.Run(fmt.Sprintf("passthrough=%v/extra=%d", passthrough, len(extra)), func(t *testing.T) {
				c, rec := streamRetryContext()
				cfg := &config.Config{}
				cfg.Gateway.MaxLineSize = len(preamble)
				s := &OpenAIGatewayService{cfg: cfg}
				err := runStreamRetryProcessor(s, c, io.NopCloser(strings.NewReader(preamble+extra+streamRetryFailed)), passthrough)
				var tagged *OpenAIStreamRetryError
				var failover *UpstreamFailoverError
				if extra == "" {
					require.ErrorAs(t, err, &tagged)
				} else {
					require.ErrorContains(t, err, "preamble too large")
					require.False(t, errors.As(err, &failover))
				}
				require.Empty(t, rec.Body.String())
			})
		}
	}
}

func TestOpenAIStreamRetryTickerCommentThenSuccess(t *testing.T) {
	c, rec := streamRetryContext()
	cfg := &config.Config{}
	cfg.Gateway.StreamKeepaliveInterval = 1
	cfg.Gateway.MaxLineSize = 64 * 1024
	s := &OpenAIGatewayService{cfg: cfg}
	pr, pw := io.Pipe()
	defer pr.Close()
	go func() {
		defer pw.Close()
		_, _ = io.WriteString(pw, strings.Repeat("data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 100))
		time.Sleep(1150 * time.Millisecond)
		_, _ = io.WriteString(pw, streamRetryFailed)
	}()
	err := runStreamRetryProcessor(s, c, pr, false)
	var tagged *OpenAIStreamRetryError
	require.ErrorAs(t, err, &tagged)
	require.Contains(t, rec.Body.String(), ":\n\n")
	require.NotContains(t, rec.Body.String(), "resp_failed")
	require.NoError(t, runStreamRetryProcessor(s, c, io.NopCloser(strings.NewReader(streamRetrySuccess)), false))
	require.Contains(t, rec.Body.String(), "resp_success")
}

type streamRetryCancelEOF struct{ cancel context.CancelFunc }

func (r streamRetryCancelEOF) Read([]byte) (int, error) { r.cancel(); return 0, io.EOF }

func TestOpenAIStreamRetryCancellationWinsEOF(t *testing.T) {
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		t.Run(mode, func(t *testing.T) {
			c, rec := streamRetryContext()
			ctx, cancel := context.WithCancel(c.Request.Context())
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
			cfg := &config.Config{}
			if mode == "ticker" {
				cfg.Gateway.StreamKeepaliveInterval = 1
			}
			err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: cfg}, c, io.NopCloser(streamRetryCancelEOF{cancel}), mode == "passthrough")
			require.ErrorIs(t, err, context.Canceled)
			require.ErrorContains(t, err, "stream usage incomplete")
			var failover *UpstreamFailoverError
			require.False(t, errors.As(err, &failover))
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestOpenAIStreamRetryOutputBoundary(t *testing.T) {
	for _, event := range []string{
		`{"type":"response.output_text.delta","delta":"text"}`,
		`{"type":"response.reasoning_summary_text.delta","delta":"reasoning"}`,
		`{"type":"response.function_call_arguments.delta","delta":"{}"}`,
		`{"type":"response.output_item.done","item":{"type":"image_generation_call","result":"image"}}`,
		`{"type":"response.completed","response":{"id":"resp_ok","output":[]}}`,
	} {
		for _, passthrough := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/%s", passthrough, event), func(t *testing.T) {
				c, rec := streamRetryContext()
				body := io.MultiReader(strings.NewReader("data: "+event+"\n\n"), errReadCloser{err: io.ErrUnexpectedEOF})
				err := runStreamRetryProcessor(&OpenAIGatewayService{}, c, io.NopCloser(body), passthrough)
				var failover *UpstreamFailoverError
				require.False(t, errors.As(err, &failover))
				require.NotEmpty(t, rec.Body.String())
			})
		}
	}
}

func TestOpenAIStreamRetryCommentAccounting(t *testing.T) {
	c, _ := streamRetryContext()
	c.Set(openAICompactClientStreamKey, true)
	stop := StartOpenAICompactSSEKeepalive(c, time.Hour)
	defer stop()
	value, ok := c.Get(openAICompactSSEKeepaliveKey)
	require.True(t, ok)
	require.True(t, value.(*openAICompactSSEKeepalive).beat())
	require.NoError(t, writeOpenAIStreamComment(c, ": ordinary\n\n"))
	require.Equal(t, -1, OpenAIStreamWrittenSize(c))
	require.False(t, openAIStreamClientOutputStarted(c, false))
	_, err := c.Writer.WriteString("data: semantic\n\n")
	require.NoError(t, err)
	require.Greater(t, OpenAIStreamWrittenSize(c), 0)
	require.True(t, openAIStreamClientOutputStarted(c, false))
	for _, unknown := range []string{"", "partial"} {
		c, _ := streamRetryContext()
		if unknown == "" {
			c.Writer.WriteHeaderNow()
		} else {
			_, _ = c.Writer.WriteString(unknown)
		}
		require.NoError(t, writeOpenAIStreamComment(c, ": heartbeat\n\n"))
		require.True(t, openAIStreamClientOutputStarted(c, false))
	}
	c, _ = streamRetryContext()
	c.Set(openAICompactClientStreamKey, true)
	stopUnknown := StartOpenAICompactSSEKeepalive(c, time.Hour)
	defer stopUnknown()
	c.Writer.WriteHeaderNow()
	require.True(t, openAIStreamClientOutputStarted(c, false))
	c, _ = streamRetryContext()
	c.Writer = &streamRetryPartialCommentWriter{ResponseWriter: c.Writer}
	require.Error(t, writeOpenAIStreamComment(c, ": heartbeat\n\n"))
	require.True(t, openAIStreamClientOutputStarted(c, false))
}

type streamRetryPartialCommentWriter struct{ gin.ResponseWriter }

func (w *streamRetryPartialCommentWriter) WriteString(s string) (int, error) {
	n, _ := w.ResponseWriter.WriteString(s[:1])
	return n, io.ErrClosedPipe
}

func TestOpenAIStreamRetryPinAndPoolRestrictions(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyPinnedAccountID, int64(9))
	require.Equal(t, int64(9), OpenAIStreamRetryContext(ctx, 1, 7).Value(ctxkey.APIKeyPinnedAccountID))
	ctx = OpenAIStreamRetryContext(context.Background(), 1, 7)
	privateOwner := int64(8)
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, OwnerUserID: &privateOwner}
	require.False(t, isOpenAIAccountEligibleForRequest(ctx, account, "", false, "", AccountCapabilityChat))
	privateOwner = 7
	ctx = withAccountPoolStrategy(ctx, AccountPoolStrategyPrivateOnly)
	require.True(t, isOpenAIAccountEligibleForRequest(ctx, account, "", false, "", AccountCapabilityChat))
	ctx = withAccountPoolStrategy(ctx, AccountPoolStrategySharedOnly)
	require.False(t, isOpenAIAccountEligibleForRequest(ctx, account, "", false, "", AccountCapabilityChat))
}

type streamRetryCancelOnOutputWriter struct {
	gin.ResponseWriter
	cancel  context.CancelFunc
	written chan struct{}
	once    sync.Once
}

func (w *streamRetryCancelOnOutputWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if err == nil && strings.Contains(string(p), "paid-output") {
		w.once.Do(func() { w.cancel(); close(w.written) })
	}
	return n, err
}
func (w *streamRetryCancelOnOutputWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// The second upstream read waits for a real downstream semantic write. This
// avoids ticker read-ahead turning a post-output cancellation into a preamble
// cancellation and proves the detached/readable upstream is still drained.
type streamRetryReadableAfterCancel struct {
	first, terminal string
	written         <-chan struct{}
}

func (r *streamRetryReadableAfterCancel) Read(p []byte) (int, error) {
	if r.first != "" {
		n := copy(p, r.first)
		r.first = r.first[n:]
		return n, nil
	}
	if r.terminal != "" {
		<-r.written
		n := copy(p, r.terminal)
		r.terminal = r.terminal[n:]
		return n, nil
	}
	return 0, io.EOF
}

func TestOpenAIStreamRetryCancellationAfterOutputDrainsUsage(t *testing.T) {
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for _, image := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/image=%v", mode, image), func(t *testing.T) {
				c, _ := streamRetryContext()
				ctx, cancel := context.WithCancel(c.Request.Context())
				defer cancel()
				c.Request = c.Request.WithContext(ctx)
				written := make(chan struct{})
				c.Writer = &streamRetryCancelOnOutputWriter{ResponseWriter: c.Writer, cancel: cancel, written: written}
				first := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"paid-output\"}\n\n"
				if image {
					first = "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"image_generation_call\",\"id\":\"img_1\",\"result\":\"paid-output\"}}\n\n"
				}
				body := &streamRetryReadableAfterCancel{first: first, written: written, terminal: "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_usage\",\"output\":[],\"usage\":{\"input_tokens\":13,\"output_tokens\":7,\"input_tokens_details\":{\"cached_tokens\":3}}}}\n\n"}
				resp := &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(body)}
				cfg := &config.Config{}
				if mode == "ticker" {
					cfg.Gateway.StreamKeepaliveInterval = 1
				}
				s := &OpenAIGatewayService{cfg: cfg}
				account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
				var usage *OpenAIUsage
				var imageCount int
				if mode == "passthrough" {
					result, err := s.handleStreamingResponsePassthrough(ctx, resp, c, account, time.Now(), "gpt-5", "gpt-5")
					require.NoError(t, err)
					usage, imageCount = result.usage, result.imageCount
				} else {
					result, err := s.handleStreamingResponse(ctx, resp, c, account, time.Now(), "gpt-5", "gpt-5")
					require.NoError(t, err)
					usage, imageCount = result.usage, result.imageCount
				}
				require.ErrorIs(t, ctx.Err(), context.Canceled)
				require.Equal(t, 13, usage.InputTokens)
				require.Equal(t, 7, usage.OutputTokens)
				require.Equal(t, 3, usage.CacheReadInputTokens)
				if image {
					require.Equal(t, 1, imageCount)
				}
				require.Empty(t, body.terminal, "terminal usage must have been drained after cancellation")
			})
		}
	}
}
