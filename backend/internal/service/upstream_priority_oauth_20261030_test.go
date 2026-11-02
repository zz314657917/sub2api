package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Route the production Forward request through a real localhost HTTP connection,
// retaining its original URL for endpoint assertions.
type upstreamPriorityOAuthHTTP struct {
	client      *http.Client
	target      *url.URL
	originalURL string
	calls       int
}

func (u *upstreamPriorityOAuthHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.calls++
	u.originalURL = req.URL.String()
	routed := req.Clone(req.Context())
	routedURL := *req.URL
	routedURL.Scheme, routedURL.Host = u.target.Scheme, u.target.Host
	routed.URL, routed.Host = &routedURL, u.target.Host
	return u.client.Do(routed)
}

func (u *upstreamPriorityOAuthHTTP) DoWithTLS(req *http.Request, proxy string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, accountID, concurrency)
}

func upstreamPriorityOAuthServer(t *testing.T, body, contentType string) (*httptest.Server, <-chan []byte) {
	t.Helper()
	requests := make(chan []byte, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests <- requestBody
		// A nil header value suppresses net/http's automatic content sniffing.
		w.Header()["Content-Type"] = nil
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set(openAICodexTurnStateHeader, "priority-turn")
		w.Header().Set("X-Request-Id", "priority-request")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server, requests
}

func upstreamPriorityOAuthContext(path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("session_id", "priority-session")
	c.Set("api_key", &APIKey{ID: 77})
	return c, rec
}

func upstreamPriorityOAuthAccount(passthrough bool) *Account {
	return &Account{
		ID: 701, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1,
		Credentials: map[string]any{"access_token": "local-test", "chatgpt_account_id": "local-account"},
		Extra: map[string]any{
			"openai_passthrough":                        passthrough,
			"openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff,
		},
	}
}

func upstreamPriorityOAuthFrame(payload string) string {
	return "data: " + payload + "\n\n"
}

func upstreamPriorityOAuthTerminal(eventType, status, output string) string {
	return upstreamPriorityOAuthFrame(fmt.Sprintf(`{"type":%q,"response":{"id":"resp_priority","object":"response","model":"mapped-model","status":%q,"incomplete_details":{"reason":"max_output_tokens"},"output":%s,"usage":{"input_tokens":19,"output_tokens":7,"input_tokens_details":{"cached_tokens":5}}}}`, eventType, status, output))
}

func upstreamPriorityOAuthRead(t *testing.T, server *httptest.Server) *http.Response {
	t.Helper()
	resp, err := server.Client().Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

type upstreamPriorityOAuthResult struct {
	usage      *OpenAIUsage
	responseID string
	images     int
	sizes      []string
}

func upstreamPriorityOAuthHandle(svc *OpenAIGatewayService, c *gin.Context, a *Account, resp *http.Response, passthrough bool) (*upstreamPriorityOAuthResult, error) {
	if passthrough {
		r, err := svc.handleNonStreamingResponsePassthrough(c.Request.Context(), resp, c, a, "client-model", "mapped-model")
		if r == nil {
			return nil, err
		}
		return &upstreamPriorityOAuthResult{r.usage, r.responseID, r.imageCount, r.imageOutputSizes}, err
	}
	r, err := svc.handleNonStreamingResponse(c.Request.Context(), resp, c, a, "client-model", "mapped-model")
	if r == nil {
		return nil, err
	}
	return &upstreamPriorityOAuthResult{r.usage, r.responseID, r.imageCount, r.imageOutputSizes}, err
}

func TestUpstreamPriorityOAuthForwardClientStream(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, stream := range []string{"false", "omitted", "true"} {
			for _, contentType := range []string{"text/event-stream", ""} {
				t.Run(fmt.Sprintf("pass=%v/stream=%s/header=%s", passthrough, stream, contentType), func(t *testing.T) {
					payload := map[string]any{"model": "gpt-5.5", "instructions": "test", "input": "hello"}
					if stream != "omitted" {
						payload["stream"] = stream == "true"
					}
					body, err := json.Marshal(payload)
					require.NoError(t, err)
					sse := upstreamPriorityOAuthTerminal("response.completed", "completed", `[{"id":"msg","type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]}]`)
					server, requests := upstreamPriorityOAuthServer(t, sse, contentType)
					target, err := url.Parse(server.URL)
					require.NoError(t, err)
					upstream := &upstreamPriorityOAuthHTTP{client: server.Client(), target: target}
					svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
					c, rec := upstreamPriorityOAuthContext("/v1/responses", body)
					result, err := svc.Forward(context.Background(), c, upstreamPriorityOAuthAccount(passthrough), body)
					require.NoError(t, err)
					require.NotNil(t, result)
					sent := <-requests
					require.Equal(t, 1, upstream.calls)
					require.Equal(t, "https://chatgpt.com/backend-api/codex/responses", upstream.originalURL)
					require.True(t, gjson.GetBytes(sent, "stream").Bool(), "OAuth upstream must remain streaming")
					require.Equal(t, 19, result.Usage.InputTokens)
					require.Equal(t, 7, result.Usage.OutputTokens)
					require.Equal(t, stream == "true", result.Stream)
					if stream == "true" {
						require.Contains(t, rec.Header().Get("Content-Type"), "text/event-stream")
						require.Contains(t, rec.Body.String(), "data:")
					} else {
						require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
						require.True(t, json.Valid(rec.Body.Bytes()), rec.Body.String())
						require.Equal(t, "resp_priority", gjson.Get(rec.Body.String(), "id").String())
					}
				})
			}
		}
	}
}

func TestUpstreamPriorityOAuthForwardCompact(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, stream := range []string{"false", "omitted", "true"} {
			t.Run(fmt.Sprintf("pass=%v/stream=%s", passthrough, stream), func(t *testing.T) {
				payload := map[string]any{"model": "gpt-5.5", "instructions": "test", "input": []any{map[string]any{"type": "message", "role": "user", "content": "hello"}}, "store": true}
				if stream != "omitted" {
					payload["stream"] = stream == "true"
				}
				body, err := json.Marshal(payload)
				require.NoError(t, err)
				server, requests := upstreamPriorityOAuthServer(t, `{"id":"resp_compact","output":[{"id":"cmp","type":"compaction","encrypted_content":"opaque"}],"usage":{"input_tokens":2,"output_tokens":1}}`, "application/json")
				target, err := url.Parse(server.URL)
				require.NoError(t, err)
				upstream := &upstreamPriorityOAuthHTTP{client: server.Client(), target: target}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				c, rec := upstreamPriorityOAuthContext("/v1/responses/compact", body)
				if stream == "true" {
					MarkOpenAICompactClientStream(c)
				}
				if !passthrough {
					// The ordinary compact Forward path receives the handler's
					// normalized body; its stream marker carries the client intent.
					body, _, err = NormalizeOpenAICompactRequestBodyForTest(body)
					require.NoError(t, err)
				}
				result, err := svc.Forward(context.Background(), c, upstreamPriorityOAuthAccount(passthrough), body)
				require.NoError(t, err)
				require.NotNil(t, result)
				sent := <-requests
				require.Equal(t, 1, upstream.calls)
				require.Equal(t, "https://chatgpt.com/backend-api/codex/responses/compact", upstream.originalURL)
				require.False(t, gjson.GetBytes(sent, "stream").Exists())
				require.False(t, gjson.GetBytes(sent, "store").Exists())
				require.Equal(t, "hello", gjson.GetBytes(sent, "input.0.content").String())
				if stream == "true" {
					require.Contains(t, rec.Header().Get("Content-Type"), "text/event-stream")
					events := parseCompactBridgeEvents(t, rec.Body.String())
					require.Len(t, events, 2)
					require.Equal(t, "opaque", gjson.Get(events[0][1], "item.encrypted_content").String())
				} else {
					require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
					require.Equal(t, "resp_compact", gjson.Get(rec.Body.String(), "id").String())
				}
			})
		}
	}
}

func TestUpstreamPriorityOAuthTerminalObjects(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, terminal := range []string{"completed", "done", "incomplete", "cancelled", "canceled"} {
			for _, contentType := range []string{"text/event-stream", ""} {
				for _, outputMode := range []string{"reconstruct", "terminal_image"} {
					t.Run(fmt.Sprintf("pass=%v/%s/header=%s/%s", passthrough, terminal, contentType, outputMode), func(t *testing.T) {
						preamble := upstreamPriorityOAuthFrame(`{"type":"response.output_item.done","item":{"id":"msg","type":"message","role":"assistant","content":[{"type":"output_text","text":"partial data: and event:"}]}}`) +
							upstreamPriorityOAuthFrame(`{"type":"response.output_item.done","item":{"id":"fc","type":"function_call","call_id":"call_1","name":"alias_tool","arguments":"{\"x\":1}{\"x\":1}"}}`) +
							upstreamPriorityOAuthFrame(`{"type":"response.output_item.done","item":{"id":"img","type":"image_generation_call","result":"aGVsbG8=","size":"1024x1024"}}`)
						output := "[]"
						if outputMode == "terminal_image" {
							preamble = ""
							output = `[{"id":"img","type":"image_generation_call","result":"aGVsbG8=","size":"1024x1024"}]`
						}
						status := terminal
						if terminal == "done" {
							status = "completed"
						}
						server, _ := upstreamPriorityOAuthServer(t, preamble+upstreamPriorityOAuthTerminal("response."+terminal, status, output), contentType)
						svc := &OpenAIGatewayService{cfg: &config.Config{}}
						c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
						setCodexToolNameReverse(c, map[string]string{"alias_tool": "client_tool"})
						notes := 0
						svc.openaiCodexTurnStateNoteHook = func(noteContext *gin.Context) {
							require.True(t, noteContext.Writer.Written())
							notes++
						}
						a := upstreamPriorityOAuthAccount(passthrough)
						result, err := upstreamPriorityOAuthHandle(svc, c, a, upstreamPriorityOAuthRead(t, server), passthrough)
						require.NoError(t, err)
						require.NotNil(t, result)
						require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
						require.True(t, json.Valid(rec.Body.Bytes()), rec.Body.String())
						require.True(t, gjson.ParseBytes(rec.Body.Bytes()).IsObject())
						require.Equal(t, status, gjson.Get(rec.Body.String(), "status").String())
						require.Equal(t, "max_output_tokens", gjson.Get(rec.Body.String(), "incomplete_details.reason").String())
						require.Equal(t, "client-model", gjson.Get(rec.Body.String(), "model").String())
						require.Equal(t, "resp_priority", result.responseID)
						require.Equal(t, 19, result.usage.InputTokens)
						require.Equal(t, 7, result.usage.OutputTokens)
						require.Equal(t, 5, result.usage.CacheReadInputTokens)
						require.Equal(t, 1, result.images)
						require.Equal(t, []string{"1024x1024"}, result.sizes)
						if outputMode == "reconstruct" {
							require.Equal(t, "partial data: and event:", gjson.Get(rec.Body.String(), "output.0.content.0.text").String())
							require.Equal(t, "client_tool", gjson.Get(rec.Body.String(), "output.1.name").String())
							require.Equal(t, `{"x":1}`, gjson.Get(rec.Body.String(), "output.1.arguments").String())
						}
						require.Equal(t, "priority-turn", rec.Header().Get(openAICodexTurnStateHeader))
						origin, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
						require.True(t, exists)
						require.Equal(t, a.ID, origin.(openAICodexTurnStateOrigin).accountID)
						require.Equal(t, 1, notes, "HTTP write, including cancellation terminals, commits provenance")
					})
				}
			}
		}
	}
}

func TestUpstreamPriorityOAuthImagesStopAtFirstTerminal(t *testing.T) {
	image := `{"id":"first_image","type":"image_generation_call","result":"aGVsbG8=","size":"1024x1024"}`
	lateImage := `{"id":"late_image","type":"image_generation_call","result":"bGF0ZQ==","size":"1536x1024"}`
	for _, passthrough := range []bool{false, true} {
		for _, terminal := range []string{"completed", "done", "incomplete", "cancelled", "canceled"} {
			for _, framing := range []string{"lf", "crlf", "multiline_crlf", "coalesced"} {
				for _, outputMode := range []string{"repeated_image", "terminal_image", "empty_with_prior_image", "empty_without_prior_image"} {
					t.Run(fmt.Sprintf("pass=%v/%s/%s/%s", passthrough, terminal, framing, outputMode), func(t *testing.T) {
						frame := func(payload string) string {
							switch framing {
							case "crlf":
								return "data: " + payload + "\r\n\r\n"
							case "multiline_crlf":
								// Split at the outer type field, preserving valid multiline JSON.
								payload = strings.Replace(payload, `,"`, ",\n\"", 1)
								return "data: " + strings.ReplaceAll(payload, "\n", "\r\ndata: ") + "\r\n\r\n"
							case "coalesced":
								// Compatibility framing: independent JSON payloads without blank lines.
								return "data: " + payload + "\n"
							default:
								return upstreamPriorityOAuthFrame(payload)
							}
						}
						firstOutput, preamble, wantImages := "["+image+"]", "", 1
						if outputMode == "repeated_image" || outputMode == "empty_with_prior_image" {
							item := frame(`{"type":"response.output_item.done","item":` + image + `}`)
							preamble = item + item
						}
						if strings.HasPrefix(outputMode, "empty_") {
							firstOutput = "[]"
						}
						if outputMode == "empty_without_prior_image" {
							wantImages = 0
						}
						status := terminal
						if terminal == "done" {
							status = "completed"
						}
						terminalFrame := func(kind, responseStatus, output string) string {
							raw := strings.TrimSuffix(strings.TrimPrefix(upstreamPriorityOAuthTerminal("response."+kind, responseStatus, output), "data: "), "\n\n")
							return frame(raw)
						}
						body := preamble + terminalFrame(terminal, status, firstOutput) +
							frame(`{"type":"response.output_item.done","item":`+lateImage+`}`) +
							frame(`{"type":"response.output_item.done","item":{"id":"late_compaction","type":"compaction","encrypted_content":"late"}}`) +
							terminalFrame("completed", "completed", "["+lateImage+"]")
						server, _ := upstreamPriorityOAuthServer(t, body, "")
						c, rec := upstreamPriorityOAuthContext("/v1/responses/compact", nil)
						result, err := upstreamPriorityOAuthHandle(&OpenAIGatewayService{cfg: &config.Config{}}, c,
							upstreamPriorityOAuthAccount(passthrough), upstreamPriorityOAuthRead(t, server), passthrough)
						require.NoError(t, err)
						require.NotNil(t, result)
						require.Equal(t, wantImages, result.images)
						if wantImages > 0 {
							require.Equal(t, []string{"1024x1024"}, result.sizes)
							require.Equal(t, "first_image", gjson.Get(rec.Body.String(), "output.0.id").String())
						} else {
							require.Empty(t, result.sizes)
							require.Empty(t, gjson.Get(rec.Body.String(), "output").Array())
						}
						require.NotContains(t, rec.Body.String(), "late_image")
						require.NotContains(t, rec.Body.String(), "late_compaction")
						require.Equal(t, status, gjson.Get(rec.Body.String(), "status").String())
						require.Equal(t, 19, result.usage.InputTokens)
						require.Equal(t, 7, result.usage.OutputTokens)
					})
				}
			}
		}
	}
}

func TestUpstreamPriorityOAuthJSONAndAPIKeyCompatibility(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, accountType := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
			for _, contentType := range []string{"application/json", "", "text/plain"} {
				t.Run(fmt.Sprintf("pass=%v/%s/header=%s", passthrough, accountType, contentType), func(t *testing.T) {
					body := `{"id":"resp_json","model":"mapped-model","output":[{"type":"message","content":[{"type":"output_text","text":"data: literal\nevent: literal"}]}],"usage":{"input_tokens":9,"output_tokens":4}}`
					server, _ := upstreamPriorityOAuthServer(t, body, contentType)
					c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
					a := upstreamPriorityOAuthAccount(passthrough)
					a.Type = accountType
					result, err := upstreamPriorityOAuthHandle(&OpenAIGatewayService{cfg: &config.Config{}}, c, a, upstreamPriorityOAuthRead(t, server), passthrough)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.True(t, json.Valid(rec.Body.Bytes()))
					require.Equal(t, "resp_json", result.responseID)
					require.Equal(t, 9, result.usage.InputTokens)
					require.Equal(t, "data: literal\nevent: literal", gjson.Get(rec.Body.String(), "output.0.content.0.text").String())
				})
			}
		}
		for _, contentType := range []string{"", "text/plain", "text/event-stream"} {
			t.Run(fmt.Sprintf("api-key-sse/pass=%v/header=%s", passthrough, contentType), func(t *testing.T) {
				sse := upstreamPriorityOAuthTerminal("response.completed", "completed", "[]")
				server, _ := upstreamPriorityOAuthServer(t, sse, contentType)
				c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
				a := upstreamPriorityOAuthAccount(passthrough)
				a.Type = AccountTypeAPIKey
				result, err := upstreamPriorityOAuthHandle(&OpenAIGatewayService{cfg: &config.Config{}}, c, a, upstreamPriorityOAuthRead(t, server), passthrough)
				require.NoError(t, err)
				require.Equal(t, 19, result.usage.InputTokens)
				if passthrough && contentType != "text/event-stream" {
					require.Equal(t, sse, rec.Body.String(), "keep API-key passthrough's existing wrong-header behavior")
				} else {
					require.True(t, json.Valid(rec.Body.Bytes()), rec.Body.String())
					require.Equal(t, "resp_priority", result.responseID)
				}
			})
		}
	}
}

func TestUpstreamPriorityOAuthMalformedAndTerminalAuthority(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, terminal := range []string{"completed", "done", "incomplete", "cancelled", "canceled"} {
			for _, response := range []string{"[]", "null", `"text"`, "1", "true", "MISSING", `{"id":"broken"`} {
				t.Run(fmt.Sprintf("pass=%v/%s/response=%s", passthrough, terminal, response), func(t *testing.T) {
					payload := fmt.Sprintf(`{"type":"response.%s","response":%s}`, terminal, response)
					if response == "MISSING" {
						payload = fmt.Sprintf(`{"type":"response.%s"}`, terminal)
					}
					server, _ := upstreamPriorityOAuthServer(t, upstreamPriorityOAuthFrame(payload), "text/event-stream")
					c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
					svc := &OpenAIGatewayService{cfg: &config.Config{}}
					result, err := upstreamPriorityOAuthHandle(svc, c, upstreamPriorityOAuthAccount(passthrough), upstreamPriorityOAuthRead(t, server), passthrough)
					require.Error(t, err)
					require.Nil(t, result)
					require.Equal(t, http.StatusBadGateway, rec.Code)
					require.True(t, json.Valid(rec.Body.Bytes()))
					require.Equal(t, "upstream_error", gjson.Get(rec.Body.String(), "error.type").String())
					_, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
					require.False(t, exists)
				})
			}
		}
		for _, terminal := range []string{"response.failed", "error", "response.incomplete", "response.cancelled", "response.canceled", "response.completed"} {
			t.Run(fmt.Sprintf("first-terminal/pass=%v/%s", passthrough, terminal), func(t *testing.T) {
				first := upstreamPriorityOAuthTerminal(terminal, strings.TrimPrefix(terminal, "response."), "[]")
				failure := terminal == "response.failed" || terminal == "error"
				if failure {
					first = upstreamPriorityOAuthFrame(fmt.Sprintf(`{"type":%q,"error":{"message":"permanent request rejection"}}`, terminal))
				}
				second := upstreamPriorityOAuthTerminal("response.completed", "later-success", "[]")
				if terminal == "response.completed" {
					second = upstreamPriorityOAuthFrame(`{"type":"error","error":{"message":"late error"}}`)
				}
				server, _ := upstreamPriorityOAuthServer(t, first+second, "text/event-stream")
				c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
				svc := &OpenAIGatewayService{cfg: &config.Config{}}
				result, err := upstreamPriorityOAuthHandle(svc, c, upstreamPriorityOAuthAccount(passthrough), upstreamPriorityOAuthRead(t, server), passthrough)
				if failure {
					require.Error(t, err)
					require.Nil(t, result)
					if terminal == "response.failed" {
						// Local response.failed remains eligible for failover,
						// without committing any downstream response.
						var failover *UpstreamFailoverError
						require.ErrorAs(t, err, &failover)
						require.False(t, c.Writer.Written())
						require.Empty(t, rec.Body.String())
					} else {
						require.Equal(t, http.StatusBadGateway, rec.Code)
					}
				} else {
					require.NoError(t, err)
					require.Equal(t, strings.TrimPrefix(terminal, "response."), gjson.Get(rec.Body.String(), "status").String())
				}
			})
		}
	}
}

func TestUpstreamPriorityOAuthFailureAndRawSSEBoundaries(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, ending := range []string{"no_terminal", "failed", "error", "cyber", "committed_failure"} {
			t.Run(fmt.Sprintf("pass=%v/%s", passthrough, ending), func(t *testing.T) {
				body := upstreamPriorityOAuthFrame(`{"type":"response.output_text.delta","delta":"partial"}`)
				switch ending {
				case "failed", "committed_failure":
					body += upstreamPriorityOAuthFrame(`{"type":"response.failed","response":{"error":{"message":"Selected model is at capacity. Please try a different model."}}}`)
				case "error":
					body += upstreamPriorityOAuthFrame(`{"type":"error","error":{"message":"Temporary upstream failure, please retry"}}`)
				case "cyber":
					body += upstreamPriorityOAuthFrame(`{"type":"response.failed","response":{"error":{"code":"cyber_policy","message":"blocked by cyber policy"},"usage":{"input_tokens":2,"output_tokens":1}}}`)
				}
				server, _ := upstreamPriorityOAuthServer(t, body, "text/event-stream")
				c, rec := upstreamPriorityOAuthContext("/v1/responses", nil)
				if ending == "committed_failure" {
					MarkResponseCommitted(c)
				}
				svc := &OpenAIGatewayService{cfg: &config.Config{}}
				result, err := upstreamPriorityOAuthHandle(svc, c, upstreamPriorityOAuthAccount(passthrough), upstreamPriorityOAuthRead(t, server), passthrough)
				var failover *UpstreamFailoverError
				if ending == "no_terminal" {
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, body, rec.Body.String())
					require.Contains(t, rec.Header().Get("Content-Type"), "text/event-stream")
					_, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
					require.True(t, exists, "raw SSE fallback retains HTTP-write provenance")
					return
				}
				require.Error(t, err)
				require.Nil(t, result)
				require.False(t, gjson.Get(rec.Body.String(), "status").Exists())
				if ending == "failed" || ending == "error" {
					require.ErrorAs(t, err, &failover)
					require.False(t, c.Writer.Written())
					require.Empty(t, rec.Body.String())
				} else {
					require.False(t, errors.As(err, &failover))
					require.True(t, c.Writer.Written())
					if ending == "cyber" {
						require.ErrorIs(t, err, errOpenAICyberPolicyForwarded)
						require.Equal(t, http.StatusBadRequest, rec.Code)
					}
				}
				_, exists := svc.openaiCodexTurnStateOrigins.Load(openAICodexTurnStateSeed(c))
				require.False(t, exists, "a failure writer does not create successful-response provenance")
			})
		}
	}
}
