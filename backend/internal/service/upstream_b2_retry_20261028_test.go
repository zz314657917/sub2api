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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const upstreamB2Overload = `{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}`
const upstreamB2Failed = "data: " + `{"type":"response.failed","response":{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}}` + "\n\n"

func TestUpstreamB2StreamBoundaries(t *testing.T) {
	empty := []string{
		`{"type":"response.output_text.delta","delta":""}`,
		`{"type":"response.reasoning_text.delta","delta":""}`,
		`{"type":"response.reasoning_summary_text.delta","delta":""}`,
		`{"type":"response.refusal.delta","delta":""}`,
		`{"type":"response.function_call_arguments.delta","delta":""}`,
		`{"type":"response.custom_tool_call_input.delta","delta":""}`,
		`{"type":"response.output_text.done","text":""}`,
		`{"type":"response.reasoning_text.done","text":""}`,
		`{"type":"response.reasoning_summary_text.done","text":""}`,
		`{"type":"response.refusal.done","refusal":""}`,
		`{"type":"response.content_part.added","part":{"type":"output_text","text":"","annotations":[]}}`,
		`{"type":"response.content_part.done","part":{"type":"refusal","refusal":""}}`,
		`{"type":"response.reasoning_summary_part.added","part":{"type":"summary_text","text":""}}`,
		`{"type":"response.reasoning_summary_part.done","part":{"type":"summary_text","text":""}}`,
		`{"type":"error","error":{"code":"server_is_overloaded"}}`,
		`{"type":"error","error":{"message":"temporary upstream problem"}}`,
	}
	boundaries := []string{
		`{"type":"response.function_call_arguments.done","arguments":""}`,
		`{"type":"response.custom_tool_call_input.done","input":""}`,
		`{"type":"response.output_item.added","item":{"type":"function_call","id":"exposed","name":"tool","arguments":""}}`,
		`{"type":"response.output_item.added","item":{"type":"custom_tool_call","id":"exposed","name":"tool","input":""}}`,
		`{"type":"response.output_item.done","item":{"type":"function_call","arguments":""}}`,
		`{"type":"response.output_item.done","item":{"type":"custom_tool_call","input":""}}`,
		`{"type":"response.reasoning.encrypted","encrypted_content":""}`,
		`{"type":"future.event","delta":""}`,
		`{"type":"response.output_text.delta","delta":`,
		`{"type":"response.content_part.done","part":null}`,
		`{"type":"response.content_part.done","part":{"type":"future_part","text":""}}`,
		`{"type":"response.content_part.done","part":{"type":"text","text":""}}`,
		`{"type":"response.content_part.done","part":{"type":"summary_text","text":""}}`,
		`{"type":"response.reasoning_summary_part.done","part":{"type":"output_text","text":""}}`,
		`{"type":"response.output_text.delta","delta":"","encrypted_content":"secret"}`,
		`{"type":"response.content_part.done","part":{"type":"output_text","text":"","encrypted_content":"secret"}}`,
		`{"type":"error","error":{"message":"permanent ordinary failure"}}`,
		`null`,
	}
	// Every known string field must reject absence, null, numbers, objects,
	// arrays, whitespace and actual text in all three production modes.
	for _, spec := range []struct{ event, field string }{
		{"response.output_text.delta", "delta"}, {"response.reasoning_text.delta", "delta"},
		{"response.reasoning_summary_text.delta", "delta"}, {"response.refusal.delta", "delta"},
		{"response.function_call_arguments.delta", "delta"}, {"response.custom_tool_call_input.delta", "delta"},
		{"response.output_text.done", "text"}, {"response.reasoning_text.done", "text"},
		{"response.reasoning_summary_text.done", "text"}, {"response.refusal.done", "refusal"},
	} {
		for _, value := range []string{"MISSING", "null", "1", "{}", "[]", `" "`, `"\n"`, `"paid"`} {
			body := fmt.Sprintf(`{"type":%q}`, spec.event)
			if value != "MISSING" {
				body = fmt.Sprintf(`{"type":%q,%q:%s}`, spec.event, spec.field, value)
			}
			boundaries = append(boundaries, body)
		}
	}
	for _, event := range []string{"response.content_part.added", "response.content_part.done", "response.reasoning_summary_part.added", "response.reasoning_summary_part.done"} {
		for _, value := range []string{"MISSING", "null", "1", "{}", "[]", `" "`, `"paid"`} {
			part := `{"type":"output_text"}`
			if strings.Contains(event, "summary") {
				part = `{"type":"summary_text"}`
			}
			if value != "MISSING" {
				part = strings.TrimSuffix(part, "}") + `,"text":` + value + "}"
			}
			boundaries = append(boundaries, fmt.Sprintf(`{"type":%q,"part":%s}`, event, part))
		}
	}
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for safe, cases := range map[bool][]string{true: empty, false: boundaries} {
			for n, event := range cases {
				t.Run(fmt.Sprintf("%s/safe=%v/%03d", mode, safe, n), func(t *testing.T) {
					c, rec := streamRetryContext()
					cfg := &config.Config{}
					cfg.Gateway.MaxLineSize = 64 * 1024
					if mode == "ticker" {
						cfg.Gateway.StreamKeepaliveInterval = 1
					}
					err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: cfg}, c, io.NopCloser(strings.NewReader("data: "+event+"\n\n"+upstreamB2Failed)), mode == "passthrough")
					var failover *UpstreamFailoverError
					require.Equal(t, safe, errors.As(err, &failover), "event: %s; err: %v", event, err)
					if safe {
						require.Empty(t, rec.Body.String())
						require.True(t, failover.RequestScopedTransient)
						require.Equal(t, 3, failover.SameAccountRetryLimit)
						require.Equal(t, -1, OpenAIStreamWrittenSize(c))
					} else {
						require.NotEmpty(t, rec.Body.String(), "conservative output must be retained")
					}
				})
			}
		}
	}
}

func TestUpstreamB2StreamEmptyTerminalEOFAndCancel(t *testing.T) {
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for _, ending := range []string{"success", "eof", "read_error", "cancel"} {
			t.Run(mode+"/"+ending, func(t *testing.T) {
				c, rec := streamRetryContext()
				cfg := &config.Config{}
				cfg.Gateway.MaxLineSize = 64 * 1024
				if mode == "ticker" {
					cfg.Gateway.StreamKeepaliveInterval = 1
				}
				preamble := ": keepalive\n\ndata: " + `{"type":"response.output_text.delta","delta":""}` + "\n\n"
				var reader io.Reader = strings.NewReader(preamble)
				ctx, cancel := context.WithCancel(c.Request.Context())
				defer cancel()
				c.Request = c.Request.WithContext(ctx)
				switch ending {
				case "success":
					reader = io.MultiReader(reader, strings.NewReader(streamRetrySuccess))
				case "read_error":
					reader = io.MultiReader(reader, errReadCloser{err: io.ErrUnexpectedEOF})
				case "cancel":
					reader = io.MultiReader(reader, streamRetryCancelEOF{cancel: cancel})
				}
				err := runStreamRetryProcessor(&OpenAIGatewayService{cfg: cfg}, c, io.NopCloser(reader), mode == "passthrough")
				var failover *UpstreamFailoverError
				switch ending {
				case "success":
					require.NoError(t, err)
					require.Contains(t, rec.Body.String(), "response.completed")
				case "cancel":
					require.ErrorIs(t, err, context.Canceled)
					require.False(t, errors.As(err, &failover))
				default:
					require.ErrorAs(t, err, &failover)
					require.NotContains(t, rec.Body.String(), "response.output_text.delta")
				}
				if ending != "cancel" {
					require.Contains(t, rec.Body.String(), ": keepalive")
				}
			})
		}
	}
}

func TestUpstreamB2RetryAfterParsing(t *testing.T) {
	now := time.Date(2026, 10, 28, 1, 2, 3, 0, time.UTC)
	for _, tc := range []struct {
		name string
		h    http.Header
		body string
		want time.Duration
	}{
		{"seconds", http.Header{"Retry-After": {"2"}}, "", 2 * time.Second},
		{"decimal", http.Header{"Retry-After": {"0.002"}}, "", 2 * time.Millisecond},
		{"lowercase", http.Header{"retry-after": {"0.003"}}, "", 3 * time.Millisecond},
		{"first_valid", http.Header{"Retry-After": {"NaN", "-1", "0", "0.004", "7"}}, "", 4 * time.Millisecond},
		{"header_seconds_priority", http.Header{"Retry-After": {"1"}, "Retry-After-Ms": {"2"}}, `{"error":{"headers":{"retry-after":"3"}}}`, time.Second},
		{"header_ms_priority", http.Header{"Retry-After": {"invalid"}, "Retry-After-Ms": {"2"}}, `{"error":{"headers":{"retry-after":"3"}}}`, 2 * time.Millisecond},
		{"nested_seconds", nil, `{"error":{"headers":{"Retry-After":2}}}`, 2 * time.Second},
		{"nested_ms", nil, `{"response":{"error":{"headers":{"retry-after-ms":"3"}}}}`, 3 * time.Millisecond},
		{"nested_array", nil, `{"error":{"headers":{"retry-after":["NaN",0,"0.004"]}}}`, 4 * time.Millisecond},
		{"date", http.Header{"Retry-After": {now.Add(4 * time.Second).Format(http.TimeFormat)}}, "", 4 * time.Second},
		{"nested_date", nil, fmt.Sprintf(`{"response":{"error":{"headers":{"retry-after":%q}}}}`, now.Add(4*time.Second).Format(http.TimeFormat)), 4 * time.Second},
		{"past_date", http.Header{"Retry-After": {now.Add(-time.Second).Format(http.TimeFormat)}}, "", 0},
		{"clamp_seconds", http.Header{"Retry-After": {"1e200"}}, "", 8 * time.Second},
		{"clamp_ms", http.Header{"Retry-After-Ms": {"1e300"}}, "", 8 * time.Second},
		{"clamp_date", http.Header{"Retry-After": {now.Add(time.Hour).Format(http.TimeFormat)}}, "", 8 * time.Second},
		{"overflow_fallback", http.Header{"Retry-After": {"1e309"}, "Retry-After-Ms": {"2"}}, "", 2 * time.Millisecond},
		{"invalid_nested_type", nil, `{"error":{"headers":{"retry-after":true,"retry-after-ms":null}}}`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, openAIRequestRetryDelay(tc.h, []byte(tc.body), now))
		})
	}
	for _, value := range []string{"", "0", "-1", "NaN", "Inf", "+Inf", "-Inf", "Infinity", "1e309", "-1e309", "word", "1s"} {
		t.Run("invalid/"+value, func(t *testing.T) {
			require.Zero(t, openAIRequestRetryDelay(http.Header{"Retry-After": {value}, "Retry-After-Ms": {value}}, nil, now))
		})
	}
}

func TestUpstreamB2EligibilityAndConflicts(t *testing.T) {
	for _, typ := range []string{AccountTypeAPIKey, AccountTypeOAuth, AccountTypeSetupToken} {
		for _, status := range []int{200, 400, 500, 502, 503, 529, 401, 402, 403, 404, 408, 409, 413, 429, 501, 504} {
			t.Run(fmt.Sprintf("%s/%d", typ, status), func(t *testing.T) {
				c, _ := streamRetryContext()
				a := &Account{ID: 1, Platform: PlatformOpenAI, Type: typ}
				require.Equal(t, openAIRequestOverloadHTTPStatus(status), openAIRequestScopedOverload(c, a, status, []byte(upstreamB2Overload), false))
				require.Equal(t, status == 200 || openAIRequestOverloadHTTPStatus(status), openAIRequestScopedOverload(c, a, status, []byte(upstreamB2Overload), true))
			})
		}
	}
	for _, tc := range []struct{ name, body string }{
		{"auth_type", `{"error":{"code":"server_is_overloaded","type":"authentication_error"}}`},
		{"mixed_auth", `{"error":{"code":"slow_down"},"response":{"error":{"code":"invalid_api_key"}}}`},
		{"root_429", `{"status_code":429,"error":{"code":"slow_down"}}`},
		{"root_401", `{"status":401,"error":{"code":"slow_down"}}`},
		{"nested_403", `{"response":{"error":{"code":"slow_down","status_code":"403"}}}`},
		{"nested_409", `{"error":{"code":"server_is_overloaded","status":409}}`},
		{"quota", `{"error":{"code":"slow_down","type":"insufficient_quota"}}`},
		{"rate", `{"code":"rate_limit_exceeded","error":{"code":"slow_down"}}`},
		{"billing", `{"error":{"code":"slow_down","message":"billing hard limit reached"}}`},
		{"permission", `{"error":{"code":"slow_down","type":"permission_denied"}}`},
		{"invalid", `{"error":{"code":"slow_down","type":"invalid_request_error"}}`},
		{"unsupported", `{"error":{"code":"slow_down","message":"unsupported parameter"}}`},
		{"model", `{"error":{"code":"slow_down"},"response":{"code":"model_not_found"}}`},
		{"capacity", `{"error":{"code":"slow_down","message":"Selected model is at capacity. Please try a different model."}}`},
		{"context", `{"error":{"code":"slow_down","message":"maximum context length exceeded"}}`},
		{"policy", `{"error":{"code":"slow_down","type":"content_policy_violation"}}`},
		{"safety", `{"error":{"code":"slow_down","message":"safety rejection"}}`},
		{"cyber", `{"error":{"code":"slow_down","message":"high-risk cyber request"}}`},
		{"account", `{"error":{"code":"slow_down"},"code":"account_disabled"}`},
		{"quoted", `{"request":{"error":{"code":"server_is_overloaded"}}}`},
		{"malformed", `{"error":{"code":"server_is_overloaded"}`},
		{"generic", `{"error":{"message":"internal server error"}}`},
		{"null", `null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := streamRetryContext()
			a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			for _, status := range []int{200, 400, 500, 502, 503, 529} {
				require.False(t, openAIRequestScopedOverload(c, a, status, []byte(tc.body), true), "%d %s", status, tc.body)
			}
			fo := (&OpenAIGatewayService{}).newOpenAIStreamFailoverError(c, a, false, "", []byte(tc.body), extractOpenAISSEErrorMessage([]byte(tc.body)))
			require.False(t, fo.RequestScopedTransient)
		})
	}
	for _, excluded := range []string{"nil_context", "nil_request", "nil_account", "anthropic", "gemini", "grok", "kimi", "upstream_type", "ws", "unknown_transport", "upgrade", "native", "other_endpoint"} {
		t.Run(excluded, func(t *testing.T) {
			c, _ := streamRetryContext()
			a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			switch excluded {
			case "nil_context":
				c = nil
			case "nil_request":
				c.Request = nil
			case "nil_account":
				a = nil
			case "upstream_type":
				a.Type = AccountTypeUpstream
			case "ws":
				SetOpenAIClientTransport(c, OpenAIClientTransportWS)
			case "unknown_transport":
				c.Set(openAIClientTransportContextKey, "future")
			case "upgrade":
				c.Request.Header.Set("Upgrade", "WebSocket")
			case "native":
				SetActualOpenAIUpstreamEndpoint(c, "/v1/messages")
			case "other_endpoint":
				SetActualOpenAIUpstreamEndpoint(c, "/v1/embeddings")
			default:
				a.Platform = excluded
			}
			require.False(t, openAIRequestScopedOverload(c, a, 503, []byte(upstreamB2Overload), false))
		})
	}
	for _, body := range []string{upstreamB2Overload, `{"type":"error","code":"slow_down"}`, `{"type":"response.failed","response":{"error":{"code":"slow_down"}}}`, "Our servers are currently overloaded."} {
		c, _ := streamRetryContext()
		require.True(t, openAIRequestScopedOverload(c, &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, 503, []byte(body), false))
	}
}

type upstreamB2HTTP struct {
	HTTPUpstream
	requests []*http.Request
	bodies   [][]byte
	closed   int
	respond  func(int) *http.Response
}

type upstreamB2Body struct {
	io.ReadCloser
	close func()
}

func (b *upstreamB2Body) Close() error { b.close(); return b.ReadCloser.Close() }
func (u *upstreamB2HTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	u.requests = append(u.requests, req)
	u.bodies = append(u.bodies, body)
	resp := u.respond(len(u.requests))
	resp.Body = &upstreamB2Body{ReadCloser: resp.Body, close: func() { u.closed++ }}
	return resp, nil
}

func upstreamB2HTTPResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}, "Retry-After-Ms": {"2"}, "X-Request-Id": {"b2-original"}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestUpstreamB2RealForwardOwners(t *testing.T) {
	for _, branch := range []string{"responses", "passthrough", "chat_to_responses", "raw_chat", "responses_to_chat", "messages", "image_apikey", "image_oauth"} {
		t.Run(branch, func(t *testing.T) {
			c, _ := streamRetryContext()
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			a := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test", "base_url": "http://127.0.0.1:1"},
				Extra:       map[string]any{"openai_responses_supported": true}}
			up := &upstreamB2HTTP{respond: func(int) *http.Response { return upstreamB2HTTPResponse(503, upstreamB2Overload) }}
			s := &OpenAIGatewayService{cfg: cfg, httpUpstream: up}
			var result *OpenAIForwardResult
			var err error
			body := []byte(`{"model":"gpt-5.1","input":"hi","stream":true}`)
			switch branch {
			case "passthrough":
				a.Extra["openai_passthrough"] = true
				fallthrough
			case "responses":
				result, err = s.Forward(context.Background(), c, a, body)
			case "chat_to_responses":
				result, err = s.ForwardAsChatCompletions(context.Background(), c, a, []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"hi"}]}`), "", "")
			case "raw_chat":
				result, err = s.forwardAsRawChatCompletions(context.Background(), c, a, []byte(`{"model":"gpt-5.1","messages":[{"role":"user","content":"hi"}]}`), "")
			case "responses_to_chat":
				a.Extra["openai_responses_supported"] = false
				result, err = s.Forward(context.Background(), c, a, body)
			case "messages":
				result, err = s.ForwardAsAnthropic(context.Background(), c, a, []byte(`{"model":"gpt-5.1","max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`), "", "")
			default:
				if branch == "image_oauth" {
					a.Type = AccountTypeOAuth
					a.Credentials["access_token"] = "test"
				}
				body = []byte(`{"model":"gpt-image-2","prompt":"draw","n":1}`)
				c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
				parsed, parseErr := s.ParseOpenAIImagesRequest(c, body)
				require.NoError(t, parseErr)
				result, err = s.ForwardImages(withOpenAIImagesForceResponses(context.Background()), c, a, body, parsed, "")
			}
			require.Nil(t, result)
			var fo *UpstreamFailoverError
			require.ErrorAs(t, err, &fo)
			require.True(t, fo.RequestScopedTransient)
			require.Equal(t, 2*time.Millisecond, fo.SameAccountRetryDelay)
			require.Equal(t, "b2-original", fo.ResponseHeaders.Get("X-Request-Id"))
			require.Equal(t, 3, fo.SameAccountRetryLimit)
			require.True(t, fo.RetryableOnSameAccount)
			require.Equal(t, 1, up.closed)
			require.Len(t, up.requests, 1)
			require.True(t, strings.HasSuffix(up.requests[0].URL.Path, "/responses") ||
				strings.HasSuffix(up.requests[0].URL.Path, "/chat/completions") ||
				strings.HasSuffix(up.requests[0].URL.Path, "/images/generations"))
		})
	}
}

func TestUpstreamB2ImagesSplitPartial(t *testing.T) {
	for _, oauth := range []bool{false, true} {
		for _, partial := range []bool{false, true} {
			t.Run(fmt.Sprintf("oauth=%v/partial=%v", oauth, partial), func(t *testing.T) {
				c, _ := streamRetryContext()
				cfg := &config.Config{}
				cfg.Security.URLAllowlist.AllowInsecureHTTP = true
				a := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
					Credentials: map[string]any{"api_key": "test", "base_url": "http://127.0.0.1:1", "access_token": "test"}}
				if oauth {
					a.Type = AccountTypeOAuth
				}
				up := &upstreamB2HTTP{respond: func(n int) *http.Response {
					if !partial || n == 2 {
						return upstreamB2HTTPResponse(503, upstreamB2Overload)
					}
					if oauth {
						r := upstreamB2HTTPResponse(200, "data: "+`{"type":"response.completed","response":{"usage":{"input_tokens":13,"output_tokens":7},"output":[{"type":"image_generation_call","result":"aW1n","size":"1024x1024"}]}}`+"\n\n")
						r.Header.Set("Content-Type", "text/event-stream")
						return r
					}
					return upstreamB2HTTPResponse(200, `{"created":1,"usage":{"input_tokens":13,"output_tokens":7},"data":[{"b64_json":"aW1n"}]}`)
				}}
				s := &OpenAIGatewayService{cfg: cfg, httpUpstream: up}
				body := []byte(`{"model":"gpt-image-2","prompt":"draw","n":3}`)
				c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
				parsed, err := s.ParseOpenAIImagesRequest(c, body)
				require.NoError(t, err)
				result, err := s.ForwardImages(withOpenAIImagesForceResponses(context.Background()), c, a, body, parsed, "")
				var fo *UpstreamFailoverError
				require.ErrorAs(t, err, &fo)
				require.True(t, fo.RequestScopedTransient)
				require.Equal(t, 2*time.Millisecond, fo.SameAccountRetryDelay)
				require.Equal(t, "b2-original", fo.ResponseHeaders.Get("X-Request-Id"))
				if partial {
					require.NotNil(t, result)
					require.Equal(t, 1, result.ImageCount)
					require.Equal(t, 13, result.Usage.InputTokens)
					require.Equal(t, 7, result.Usage.OutputTokens)
					require.Len(t, up.requests, 2)
				} else {
					require.Nil(t, result)
					require.Len(t, up.requests, 1)
				}
				require.Equal(t, len(up.requests), up.closed)
			})
		}
	}
}

type upstreamB2HealthRepo struct {
	AccountRepository
	errors, rate, model, temp, cleared int
}

func (r *upstreamB2HealthRepo) SetError(context.Context, int64, string) error {
	r.errors++
	return nil
}
func (r *upstreamB2HealthRepo) SetRateLimited(context.Context, int64, time.Time) error {
	r.rate++
	return nil
}
func (r *upstreamB2HealthRepo) SetModelRateLimit(context.Context, int64, string, time.Time) error {
	r.model++
	return nil
}
func (r *upstreamB2HealthRepo) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	r.temp++
	return nil
}
func (r *upstreamB2HealthRepo) ClearError(context.Context, int64) error     { r.cleared++; return nil }
func (r *upstreamB2HealthRepo) ClearRateLimit(context.Context, int64) error { r.cleared++; return nil }

func TestUpstreamB2HealthNeutralAndGenericDefault(t *testing.T) {
	repo := &upstreamB2HealthRepo{}
	s := &OpenAIGatewayService{rateLimitService: NewRateLimitService(repo, nil, nil, nil, nil)}
	c, _ := streamRetryContext()
	a := &Account{ID: 13, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	s.rateLimitService.SetAccountRuntimeBlocker(s)
	until := time.Now().Add(time.Hour)
	s.BlockAccountScheduling(a, until, "preexisting")
	s.recordOpenAIAccountModelTransientFailure(a, "gpt-5.1", time.Now())
	state := s.getOpenAIAccountModelTransientState()
	key, _ := openAIAccountModelTransientKey(a.ID, "gpt-5.1")
	before := state.entries[key]
	for n := 0; n < 8; n++ {
		require.False(t, s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, a, 502, nil, []byte(upstreamB2Overload), "gpt-5.1"))
		(&GatewayService{accountRepo: repo}).TempUnscheduleRetryableError(context.Background(), a.ID,
			&UpstreamFailoverError{StatusCode: 502, RetryableOnSameAccount: true, RequestScopedTransient: true})
	}
	require.Equal(t, before, state.entries[key], "neither failure nor synthetic success may change preexisting state")
	block, exists := s.openaiAccountRuntimeBlockUntil.Load(a.ID)
	require.True(t, exists)
	require.Equal(t, until, block, "request overload may not clear or extend an existing runtime block")
	fresh := &Account{ID: 14, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	require.False(t, s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, fresh, 503, nil, []byte(upstreamB2Overload)))
	require.False(t, s.isOpenAIAccountRuntimeBlocked(fresh), "eligible overload may not install a runtime block")
	require.Equal(t, &upstreamB2HealthRepo{}, repo)
	// Generic helper is deliberately still penalising. This protects WS bridge
	// and indirect non-HTTP consumers from an unconditional overload skip.
	s.handleOpenAIAccountUpstreamError(context.Background(), a, 503, nil, []byte(upstreamB2Overload), "gpt-5.1")
	require.Greater(t, state.entries[key].failureStreak, before.failureStreak)
	a.Type = AccountTypeOAuth
	s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, a, 401, nil, []byte(`{"error":{"code":"invalid_api_key"}}`))
	require.Positive(t, repo.errors)
	a.Type = AccountTypeAPIKey
	s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, a, 429, nil, []byte(`{"error":{"code":"rate_limit_exceeded"}}`))
	require.Positive(t, repo.rate)
	s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, fresh, 429,
		nil, []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":60}}`))
	require.True(t, s.isOpenAIAccountRuntimeBlocked(fresh), "genuine OAuth quota 429 retains runtime blocking")
	require.Zero(t, repo.cleared)
}

func TestUpstreamB2HeaderCloneAndStreamPolicy(t *testing.T) {
	c, _ := streamRetryContext()
	s := &OpenAIGatewayService{}
	a := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	h := http.Header{"Retry-After-Ms": {"3"}, "X-Original": {"value"}}
	fo := s.newOpenAIStreamFailoverError(c, a, false, "", []byte(upstreamB2Overload), "", h)
	require.True(t, fo.RequestScopedTransient)
	require.Equal(t, 3*time.Millisecond, fo.SameAccountRetryDelay)
	h.Set("X-Original", "mutated")
	require.Equal(t, "value", fo.ResponseHeaders.Get("X-Original"))
	for _, tc := range []struct {
		body    string
		limit   int
		generic bool
	}{
		{`{"error":{"message":"Selected model is at capacity. Please try a different model."}}`, 5, false},
		{upstreamB2Overload, 3, false},
		{`{"error":{"message":"ordinary temporary upstream error"}}`, 0, true},
		{`{"error":{"code":"invalid_api_key"}}`, 0, false},
		{`{"error":{"code":"rate_limit_exceeded"}}`, 0, false},
	} {
		a.Type = AccountTypeAPIKey
		err := s.newResponsesStreamFailoverError(c, a, false, "", []byte(tc.body), extractOpenAISSEErrorMessage([]byte(tc.body)))
		var typed *UpstreamFailoverError
		require.ErrorAs(t, err, &typed)
		require.Equal(t, tc.limit, typed.SameAccountRetryLimit)
		var generic *OpenAIStreamRetryError
		require.Equal(t, tc.generic, errors.As(err, &generic))
	}
	// Ensure all metadata is ordinary serialisable state, with no aliasing of
	// upstream headers or mutation of the original failure body.
	b := []byte(upstreamB2Overload)
	resp := upstreamB2HTTPResponse(503, upstreamB2Overload)
	out := augmentOpenAIHTTPFailover(c, a, resp, b, &UpstreamFailoverError{ResponseBody: b})
	require.True(t, reflect.DeepEqual(b, out.ResponseBody))
	_, err := json.Marshal(out)
	require.NoError(t, err)
}

func TestUpstreamB2SharedHTTPAndNativeExclusions(t *testing.T) {
	for _, transport := range []string{"http", "ws", "upgrade", "native"} {
		t.Run(transport, func(t *testing.T) {
			c, _ := streamRetryContext()
			a := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			if transport == "ws" {
				SetOpenAIClientTransport(c, OpenAIClientTransportWS)
			} else if transport == "upgrade" {
				c.Request.Header.Set("Upgrade", "websocket")
			} else if transport == "native" {
				SetActualOpenAIUpstreamEndpoint(c, "/v1/messages")
			}
			s := &OpenAIGatewayService{}
			resp := upstreamB2HTTPResponse(503, upstreamB2Overload)
			err := s.failoverOpenAIUpstreamHTTPError(context.Background(), c, a, resp, []byte(upstreamB2Overload), "", "gpt-5.1")
			var fo *UpstreamFailoverError
			require.ErrorAs(t, err, &fo)
			require.Equal(t, transport == "http", fo.RequestScopedTransient)
			require.Equal(t, "b2-original", fo.ResponseHeaders.Get("X-Request-Id"))
			require.Equal(t, transport != "http", s.getOpenAIAccountModelTransientState().size() > 0,
				"excluded indirect consumers retain generic account side effects")
		})
	}
	for _, platform := range []string{PlatformKimi, PlatformDeepseek} {
		t.Run("real_native_"+platform, func(t *testing.T) {
			c, _ := streamRetryContext()
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			a := &Account{ID: 2, Platform: platform, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test", "base_url": "http://127.0.0.1:1", "api_protocol": APIProtocolAnthropic}}
			up := &upstreamB2HTTP{respond: func(int) *http.Response { return upstreamB2HTTPResponse(503, upstreamB2Overload) }}
			s := &OpenAIGatewayService{cfg: cfg, httpUpstream: up}
			_, err := s.ForwardAsAnthropic(context.Background(), c, a, []byte(`{"model":"test-model","max_tokens":8,"messages":[{"role":"user","content":"hi"}]}`), "", "")
			var fo *UpstreamFailoverError
			require.ErrorAs(t, err, &fo)
			require.False(t, fo.RequestScopedTransient)
			require.Zero(t, fo.SameAccountRetryDelay)
			require.Len(t, up.requests, 1)
			require.True(t, strings.HasSuffix(up.requests[0].URL.Path, "/messages"))
			// The excluded native adapter restores resp.Body via the existing
			// readOpenAIUpstreamError helper. Its closure policy is out of B2.
		})
	}
}
