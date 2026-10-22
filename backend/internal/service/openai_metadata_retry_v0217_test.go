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
	"github.com/stretchr/testify/require"
)

const metadataRetryPreamble = "event: response.queued\ndata: {\"type\":\"response.queued\",\"response\":{\"id\":\"failed_metadata\",\"status\":\"queued\"}}\n\nevent: response.metadata\ndata: {\"type\":\"response.metadata\",\"response_id\":\"failed_metadata\",\"metadata\":{\"moderation\":{}}}\n\n"

func TestOpenAIStreamMetadataClassificationV0217(t *testing.T) {
	for _, event := range []string{"response.created", " response.queued ", "response.metadata", "response.in_progress"} {
		require.True(t, openAIStreamEventIsPreamble(event))
		require.False(t, openAIStreamDataStartsClientOutput(`{"value":1}`, event))
	}
	for _, event := range []string{"response.output_text.delta", "response.function_call_arguments.delta", "unknown", "keepalive"} {
		require.False(t, openAIStreamEventIsPreamble(event))
		require.True(t, openAIStreamDataStartsClientOutput(`{"value":1}`, event))
	}
}

func TestOpenAIStreamMetadataRetryV0217(t *testing.T) {
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for _, ending := range []string{"overload", "eof", "read_error"} {
			t.Run(mode+"/"+ending, func(t *testing.T) {
				c, rec := streamRetryContext()
				cfg := &config.Config{}
				cfg.Gateway.MaxLineSize = 64 * 1024
				if mode == "ticker" {
					cfg.Gateway.StreamKeepaliveInterval = 1
				}
				preamble := ": heartbeat\n\n" + strings.Repeat(metadataRetryPreamble, 30)
				require.Greater(t, len(preamble), 4096)
				var body io.Reader = strings.NewReader(preamble)
				if ending == "overload" {
					body = io.MultiReader(body, strings.NewReader("data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"server_is_overloaded\",\"message\":\"overloaded\"}}}\n\n"))
				} else if ending == "read_error" {
					body = io.MultiReader(body, errReadCloser{err: io.ErrUnexpectedEOF})
				}
				svc := &OpenAIGatewayService{cfg: cfg}
				err := runStreamRetryProcessor(svc, c, io.NopCloser(body), mode == "passthrough")
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, ending == "overload", failover.RetryableOnSameAccount)
				require.NotContains(t, rec.Body.String(), "failed_metadata")
				require.NotContains(t, rec.Body.String(), "response.failed")
				require.False(t, openAIStreamClientOutputStarted(c, false))
				require.NoError(t, runStreamRetryProcessor(svc, c, io.NopCloser(strings.NewReader(streamRetrySuccess)), mode == "passthrough"))
				require.Contains(t, rec.Body.String(), "resp_success")
				require.NotContains(t, rec.Body.String(), "failed_metadata")
			})
		}
	}
}

func TestOpenAIStreamMetadataPreambleBoundV0217(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, extra := range []string{"", "\n"} {
			t.Run(fmt.Sprintf("%v/extra=%d", passthrough, len(extra)), func(t *testing.T) {
				c, rec := streamRetryContext()
				cfg := &config.Config{}
				cfg.Gateway.MaxLineSize = len(metadataRetryPreamble)
				err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: cfg}, c,
					io.NopCloser(strings.NewReader(metadataRetryPreamble+extra+streamRetryFailed)), passthrough)
				var failover *UpstreamFailoverError
				if extra == "" {
					require.ErrorAs(t, err, &failover)
				} else {
					require.ErrorContains(t, err, "preamble too large")
					require.False(t, errors.As(err, &failover))
				}
				require.Empty(t, rec.Body.String())
			})
		}
	}
}

func TestOpenAIStreamMetadataOutputBoundaryV0217(t *testing.T) {
	for _, event := range []string{
		`{"type":"response.output_text.delta","delta":"paid-output"}`,
		`{"type":"response.function_call_arguments.delta","delta":"{}"}`,
		`{"type":"arbitrary_event","value":"unknown"}`,
	} {
		for _, passthrough := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/%s", passthrough, event), func(t *testing.T) {
				c, rec := streamRetryContext()
				body := metadataRetryPreamble + "data: " + event + "\n\n" +
					"data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"server_is_overloaded\"}}}\n\n"
				err := runStreamRetryProcessor(&OpenAIGatewayService{}, c, io.NopCloser(strings.NewReader(body)), passthrough)
				require.Error(t, err)
				var failover *UpstreamFailoverError
				require.False(t, errors.As(err, &failover))
				require.Contains(t, rec.Body.String(), "failed_metadata")
				require.True(t, openAIStreamClientOutputStarted(c, false))
			})
		}
	}
}

func TestOpenAIStreamMetadataCancellationV0217(t *testing.T) {
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
			body := io.MultiReader(strings.NewReader(metadataRetryPreamble), streamRetryCancelEOF{cancel})
			err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: cfg}, c, io.NopCloser(body), mode == "passthrough")
			require.ErrorIs(t, err, context.Canceled)
			var failover *UpstreamFailoverError
			require.False(t, errors.As(err, &failover))
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestOpenAIStreamMetadataFirstOutputDeadlineV0217(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			c, rec := streamRetryContext()
			ctx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
			pr, pw := io.Pipe()
			defer pr.Close()
			finished := make(chan struct{})
			go func() {
				defer close(finished)
				defer pw.Close()
				_, _ = io.WriteString(pw, metadataRetryPreamble)
				<-ctx.Done()
			}()
			err := runStreamRetryProcessor(&OpenAIGatewayService{}, c, pr, passthrough)
			<-finished
			require.ErrorIs(t, err, context.DeadlineExceeded)
			var failover *UpstreamFailoverError
			require.False(t, errors.As(err, &failover))
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestOpenAIStreamMetadataOnlyFirstTokenV0217(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			c, rec := streamRetryContext()
			svc := &OpenAIGatewayService{}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(metadataRetryPreamble))}
			if passthrough {
				result, err := svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
				require.Error(t, err)
				require.NotNil(t, result)
				require.Nil(t, result.firstTokenMs)
			} else {
				result, err := svc.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "gpt-5", "gpt-5")
				require.Error(t, err)
				require.NotNil(t, result)
				require.Nil(t, result.firstTokenMs)
			}
			require.Empty(t, rec.Body.String())
		})
	}
}
