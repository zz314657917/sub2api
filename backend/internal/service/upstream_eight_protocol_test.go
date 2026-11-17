package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Exercise the production native ingress loop across two real client frames.
// The existing deterministic upstream transport records exactly which turns
// reached the relay; this is not an assertion on the wrapping helper alone.
func TestUpstreamEightNativeWSFailover(t *testing.T) {
	for _, turns := range []int{1, 2} {
		t.Run(fmt.Sprint(turns), func(t *testing.T) {
			failed := []byte(`{"type":"error","error":{"type":"rate_limit_error","code":"rate_limit_exceeded","message":"limited"}}`)
			events := [][]byte{failed}
			if turns == 2 {
				events = append([][]byte{[]byte(`{"type":"response.completed","response":{"id":"resp_first","status":"completed","model":"gpt-5.1","output":[],"usage":{"input_tokens":1,"output_tokens":1}}}`)}, events...)
			}
			svc, _, _, capture, account := newMalformedWSV2ForwardTest(t, events)
			account.Credentials["base_url"] = "https://example.invalid"
			result := make(chan error, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := coderws.Accept(w, r, nil)
				if err != nil {
					result <- err
					return
				}
				defer conn.CloseNow()
				_, first, err := conn.Read(r.Context())
				if err != nil {
					result <- err
					return
				}
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = r.Clone(r.Context())
				result <- svc.ProxyResponsesWebSocketFromClient(r.Context(), c, conn, account, "sk-test", first, nil)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			client, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
			require.NoError(t, err)
			defer client.CloseNow()
			first := `{"type":"response.create","model":"gpt-5.1","input":[{"role":"user","content":"first"}]}`
			require.NoError(t, client.Write(ctx, coderws.MessageText, []byte(first)))
			if turns == 2 {
				_, message, err := client.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, "resp_first", gjson.GetBytes(message, "response.id").String())
				require.NoError(t, client.Write(ctx, coderws.MessageText, []byte(`{"type":"response.create","model":"gpt-5.1","input":[{"role":"user","content":"second"}]}`)))
			}
			select {
			case err := <-result:
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				payload, currentTurn := OpenAIWSCurrentTurnRetryPayload(err)
				require.Equal(t, turns > 1, currentTurn)
				require.Empty(t, payload, "native later turns must close rather than retry the original frame")
			case <-ctx.Done():
				t.Fatal("native ingress did not return the failover")
			}
			capture.mu.Lock()
			defer capture.mu.Unlock()
			require.Len(t, capture.writes, turns)
			require.Equal(t, "first", gjson.Get(requestToJSONString(capture.writes[0]), "input.0.content").String())
			if turns == 2 {
				require.Equal(t, "second", gjson.Get(requestToJSONString(capture.writes[1]), "input.0.content").String())
			}
		})
	}
}

func TestUpstreamEightCodexMetadataRetry(t *testing.T) {
	for _, event := range []string{"codex.rate_limits", "codex.response.metadata"} {
		for _, passthrough := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%v", event, passthrough), func(t *testing.T) {
				require.True(t, openAIStreamEventIsPreamble(event))
				require.False(t, openAIStreamDataStartsClientOutput(`{"type":"`+event+`"}`, event))
				c, rec := streamRetryContext()
				preamble := "event: " + event + "\ndata: {\"type\":\"" + event + "\",\"value\":1}\n\n"
				overload := "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"server_is_overloaded\",\"message\":\"overloaded\"}}}\n\n"
				err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: &config.Config{}}, c,
					io.NopCloser(strings.NewReader(preamble+overload)), passthrough)
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.True(t, failover.RetryableOnSameAccount)
				require.Empty(t, rec.Body.String())
				require.False(t, openAIStreamClientOutputStarted(c, false))
				require.NoError(t, runStreamRetryProcessor(&OpenAIGatewayService{}, c,
					io.NopCloser(strings.NewReader(streamRetrySuccess)), passthrough))
				require.Contains(t, rec.Body.String(), "resp_success")
				require.NotContains(t, rec.Body.String(), event)

				c, rec = streamRetryContext()
				err = runStreamRetryProcessor(&OpenAIGatewayService{}, c, io.NopCloser(strings.NewReader(
					preamble+"data: {\"type\":\"response.output_text.delta\",\"delta\":\"visible\"}\n\n"+overload)), passthrough)
				require.Error(t, err)
				require.False(t, errors.As(err, &failover))
				require.Contains(t, rec.Body.String(), "visible")
			})
		}
	}
}

func TestUpstreamEightCodexMetadataDeadline(t *testing.T) {
	for _, event := range []string{"codex.rate_limits", "codex.response.metadata"} {
		t.Run(event, func(t *testing.T) {
			c, rec := streamRetryContext()
			ctx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
			reader, writer := io.Pipe()
			defer reader.Close()
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer writer.Close()
				_, _ = fmt.Fprintf(writer, "data: {\"type\":%q}\n\n", event)
				<-ctx.Done()
			}()
			err := runStreamRetryProcessor(&OpenAIGatewayService{}, c, reader, false)
			<-done
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.Empty(t, rec.Body.String())
			require.False(t, openAIStreamClientOutputStarted(c, false))
		})
	}
}

func TestUpstreamEightSessionSeed(t *testing.T) {
	prefix := `{"model":"gpt-5.1","input":[{"role":"system","content":"A"},{"role":"developer","content":"B"}`
	for _, boundary := range []string{`{"role":"user","content":"hello"}`, `{"role":"assistant","content":"hi"}`, `{"type":"function_call_output","call_id":"c","output":"ok"}`} {
		base := prefix + "," + boundary
		first := deriveOpenAIContentSessionSeed([]byte(base + `]}`))
		require.Equal(t, first, deriveOpenAIContentSessionSeed([]byte(base+`,{"role":"system","content":"dynamic reminder"}]}`)))
		require.NotEqual(t, first, deriveOpenAIContentSessionSeed([]byte(strings.Replace(base, `"content":"A"`, `"content":"changed"`, 1)+`]}`)))
	}
	for _, request := range []string{`{"model":"gpt-5.1","input":"你好"}`, `{"model":"gpt-5.1","messages":[{"role":"system","content":"A"},{"role":"user","content":"hello"}]}`} {
		require.NotEmpty(t, deriveOpenAIContentSessionSeed([]byte(request)))
	}
}
