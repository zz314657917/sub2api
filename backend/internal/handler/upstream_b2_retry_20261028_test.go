package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const upstreamB2HandlerOverload = `{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}`
const upstreamB2HandlerFailed = "data: " + `{"type":"response.failed","response":{"error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded"}}}` + "\n\n"
const upstreamB2HandlerSuccess = "data: " + `{"type":"response.completed","response":{"id":"b2-success","status":"completed","output":[{"id":"m1","type":"message","role":"assistant","content":[{"type":"output_text","text":"paid"}]}],"usage":{"input_tokens":13,"output_tokens":7,"input_tokens_details":{"cached_tokens":3}}}}` + "\n\n"

type upstreamB2HandlerUsers struct{ service.UserRepository }

func (*upstreamB2HandlerUsers) GetByID(context.Context, int64) (*service.User, error) {
	return &service.User{ID: 7, Balance: 1000, Status: service.StatusActive}, nil
}

type upstreamB2HandlerCharges struct {
	commands []*service.UsageBillingCommand
}

func (b *upstreamB2HandlerCharges) Apply(_ context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error) {
	b.commands = append(b.commands, cmd)
	return &service.UsageBillingApplyResult{Applied: true}, nil
}

type upstreamB2HandlerSettings struct {
	oauthPendingFlowSettingRepoStub
}

func (r *upstreamB2HandlerSettings) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

type upstreamB2HandlerAccounts struct {
	*openAIWSFailoverHandlerAccountRepoStub
}

func (r *upstreamB2HandlerAccounts) ListSchedulableByGroupIDAndPlatform(ctx context.Context, group int64, platform string) ([]service.Account, error) {
	accounts, err := r.ListSchedulableByPlatform(ctx, platform)
	out := make([]service.Account, 0, len(accounts))
	for _, account := range accounts {
		for _, id := range account.GroupIDs {
			if id == group {
				out = append(out, account)
				break
			}
		}
	}
	return out, err
}

func upstreamB2HandlerFixture(t *testing.T, accounts, switches int, passthrough bool) (*retryHandlerFixture, *upstreamB2HandlerCharges) {
	t.Helper()
	f := newRetryHandlerFixture(t, accounts, switches, passthrough)
	t.Cleanup(f.h.billingCacheService.Stop)
	f.cfg.RunMode = "" // Exercise actual RecordUsage billing, including Apply.
	f.cfg.Gateway.OpenAIWS.LBTopK = 1
	f.cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Priority = 1
	users := &upstreamB2HandlerUsers{}
	charges := &upstreamB2HandlerCharges{}
	settingsRepo := &upstreamB2HandlerSettings{oauthPendingFlowSettingRepoStub{values: map[string]string{}}}
	settings := service.NewSettingService(settingsRepo, f.cfg)
	view, err := settings.GetAllSettings(context.Background())
	require.NoError(t, err)
	view.OpenAIAdvancedSchedulerEnabled = true
	require.NoError(t, settings.UpdateSettings(context.Background(), view))
	t.Cleanup(func() {
		view.OpenAIAdvancedSchedulerEnabled = false
		require.NoError(t, settings.UpdateSettings(context.Background(), view))
	})
	cache := service.NewBillingCacheService(nil, users, nil, nil, nil, nil, f.cfg)
	t.Cleanup(cache.Stop)
	concurrency := service.NewConcurrencyService(f.slots)
	rateLimit := service.NewRateLimitService(nil, nil, f.cfg, nil, nil)
	rateLimit.SetSettingService(settings)
	gateway := service.NewOpenAIGatewayService(&upstreamB2HandlerAccounts{f.repo}, f.usageRepo, charges, users, nil, nil, nil, f.cfg, nil, concurrency,
		service.NewBillingService(f.cfg, nil), rateLimit, cache, f.up, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil)
	f.h.gatewayService = gateway
	f.h.billingCacheService = cache
	groupID := int64(100)
	key := &service.APIKey{ID: 1, GroupID: &groupID, User: &service.User{ID: 7, Balance: 1000, Status: service.StatusActive},
		Group: &service.Group{ID: 100, Platform: service.PlatformOpenAI, Status: service.StatusActive, RateMultiplier: 1, AllowImageGeneration: true, AllowMessagesDispatch: true}}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), key)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7, Concurrency: 1})
		c.Next()
	})
	router.POST("/v1/responses", f.h.Responses)
	router.POST("/v1/chat/completions", f.h.ChatCompletions)
	router.POST("/v1/messages", f.h.Messages)
	router.POST("/v1/images/generations", f.h.Images)
	f.router = router
	require.Nil(t, f.h.usageRecordWorkerPool, "nil pool selects production synchronous submitter")
	return f, charges
}

func upstreamB2HandlerServe(t *testing.T, f *retryHandlerFixture, ctx context.Context, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	f.router.ServeHTTP(rec, req)
	require.Eventually(t, func() bool {
		f.slots.mu.Lock()
		defer f.slots.mu.Unlock()
		return f.slots.acquired == f.slots.released
	}, time.Second, time.Millisecond)
	require.Equal(t, len(f.up.ids), f.up.closed)
	require.Equal(t, f.slots.waitAdded, f.slots.waitRemoved)
	return rec
}

func upstreamB2HandlerHTTPError(status int) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}, "Retry-After-Ms": {"2"}, "X-Request-Id": {"b2-overload"}},
		Body: io.NopCloser(strings.NewReader(upstreamB2HandlerOverload))}
}

func upstreamB2AssertUsage(t *testing.T, f *retryHandlerFixture, charges *upstreamB2HandlerCharges, count, input, output, cached, images int) {
	t.Helper()
	// ServeHTTP has returned and the production nil-pool submitter has run
	// synchronously to completion. There are no pending async jobs to race.
	require.Nil(t, f.h.usageRecordWorkerPool)
	require.Len(t, f.usageRepo.created, count, "all completed submissions, not an early queue observation")
	require.Len(t, charges.commands, count, "each usage must settle exactly once")
	if count == 0 {
		return
	}
	log := <-f.usageRepo.created
	require.Equal(t, input, log.InputTokens+log.CacheReadTokens)
	require.Equal(t, output, log.OutputTokens)
	require.Equal(t, cached, log.CacheReadTokens)
	require.Equal(t, images, log.ImageCount)
	require.Equal(t, input, charges.commands[0].InputTokens+charges.commands[0].CacheReadTokens)
	require.Equal(t, output, charges.commands[0].OutputTokens)
	require.Equal(t, cached, charges.commands[0].CacheReadTokens)
	require.Equal(t, images, charges.commands[0].ImageCount)
	require.Empty(t, f.usageRepo.created, "no second record after completed synchronous submit")
	require.NotEmpty(t, charges.commands[0].RequestID)
	require.Positive(t, charges.commands[0].BalanceCost, "successful measured usage must reach a real charge command")
}

func TestUpstreamB2ResponsesHandlerRetryAndExhaustion(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, stream := range []bool{false, true} {
			for _, success := range []bool{false, true} {
				t.Run(fmt.Sprintf("passthrough=%v/stream=%v/success=%v", passthrough, stream, success), func(t *testing.T) {
					f, charges := upstreamB2HandlerFixture(t, 1, 0, passthrough)
					f.up.respond = func(_ *http.Request, _ int64, n int) *http.Response {
						require.Equal(t, n, f.slots.acquired)
						require.Equal(t, n-1, f.slots.released)
						if n > 1 {
							require.GreaterOrEqual(t, f.up.times[n-1].Sub(f.slots.releaseTimes[n-2]), time.Millisecond,
								"header wait occurs after release")
						}
						if success && n == 2 {
							return retryHandlerResponse(upstreamB2HandlerSuccess)
						}
						if stream {
							r := retryHandlerResponse("data: " + `{"type":"response.output_text.delta","delta":""}` + "\n\n" +
								"data: " + `{"type":"error","error":{"code":"server_is_overloaded"}}` + "\n\n" + upstreamB2HandlerFailed)
							r.Header.Set("Retry-After-Ms", "2")
							return r
						}
						return upstreamB2HandlerHTTPError(503)
					}
					rec := upstreamB2HandlerServe(t, f, context.Background(), "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`)
					if success {
						require.Equal(t, []int64{1, 1}, f.up.ids, rec.Body.String())
						require.Contains(t, rec.Body.String(), "b2-success")
						require.NotContains(t, rec.Body.String(), "server_is_overloaded")
						upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
						require.Equal(t, 1, f.h.gatewayService.SnapshotOpenAIAccountSchedulerMetrics().RuntimeStatsAccountCount)
					} else {
						require.Equal(t, []int64{1, 1, 1, 1}, f.up.ids, rec.Body.String())
						upstreamB2AssertUsage(t, f, charges, 0, 0, 0, 0, 0)
						require.Zero(t, f.h.gatewayService.SnapshotOpenAIAccountSchedulerMetrics().RuntimeStatsAccountCount,
							"failed overload may not sample scheduler failure or success")
					}
					require.Equal(t, len(f.up.ids), f.slots.acquired)
					require.Equal(t, len(f.up.ids), f.slots.released)
					require.Empty(t, f.repo.rateLimitedIDs)
				})
			}
		}
	}
}

func TestUpstreamB2ChatMessagesAndFallbackHandlers(t *testing.T) {
	for _, branch := range []string{"chat_responses", "chat_raw", "messages", "responses_chat"} {
		t.Run(branch, func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 1, 0, false)
			path, body := "/v1/chat/completions", `{"model":"gpt-5.1","stream":true,"messages":[{"role":"user","content":"hi"}]}`
			if branch == "chat_raw" || branch == "responses_chat" {
				f.repo.accounts[0].Extra["openai_responses_supported"] = false
			}
			if branch == "messages" {
				path, body = "/v1/messages", `{"model":"gpt-5.1","stream":true,"max_tokens":32,"messages":[{"role":"user","content":"hi"}]}`
			}
			if branch == "responses_chat" {
				path, body = "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`
			}
			f.up.respond = func(req *http.Request, _ int64, n int) *http.Response {
				if branch == "chat_raw" || branch == "responses_chat" {
					require.True(t, strings.HasSuffix(req.URL.Path, "/chat/completions"))
				} else {
					require.True(t, strings.HasSuffix(req.URL.Path, "/responses"))
				}
				if n == 1 {
					return upstreamB2HandlerHTTPError(503)
				}
				if branch == "chat_raw" || branch == "responses_chat" {
					return retryHandlerResponse("data: " + `{"id":"b2-chat","choices":[{"index":0,"delta":{"content":"paid"}}]}` + "\n\n" +
						"data: " + `{"choices":[],"usage":{"prompt_tokens":13,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3}}}` + "\n\ndata: [DONE]\n\n")
				}
				return retryHandlerResponse(upstreamB2HandlerSuccess)
			}
			rec := upstreamB2HandlerServe(t, f, context.Background(), path, body)
			require.Equal(t, []int64{1, 1}, f.up.ids, rec.Body.String())
			require.Equal(t, 2, f.slots.acquired)
			require.Equal(t, 2, f.slots.released)
			upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
		})
	}
}

func TestUpstreamB2VisibleOutputNoReplayAndUsagePolicies(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, event := range []string{
			`{"type":"response.output_text.delta","delta":"paid"}`,
			`{"type":"response.function_call_arguments.done","arguments":""}`,
			`{"type":"response.custom_tool_call_input.done","input":""}`,
			`{"type":"future.event","delta":""}`,
			`{"type":"response.output_text.delta","delta":null}`,
		} {
			t.Run(fmt.Sprintf("%v/%s", passthrough, event), func(t *testing.T) {
				f, charges := upstreamB2HandlerFixture(t, 2, 10, passthrough)
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					return retryHandlerResponse("data: " + event + "\n\n" + upstreamB2HandlerFailed)
				}
				upstreamB2HandlerServe(t, f, context.Background(), "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`)
				require.Equal(t, []int64{1}, f.up.ids)
				// Responses Forward wrappers intentionally return nil,error.
				upstreamB2AssertUsage(t, f, charges, 0, 0, 0, 0, 0)
			})
		}
	}
	t.Run("raw_chat_existing_partial", func(t *testing.T) {
		f, charges := upstreamB2HandlerFixture(t, 2, 10, false)
		f.repo.accounts[0].Extra["openai_responses_supported"] = false
		f.up.respond = func(*http.Request, int64, int) *http.Response {
			r := retryHandlerResponse("data: " + `{"id":"b2-partial","choices":[{"index":0,"delta":{"content":"paid"}}]}` + "\n\n" +
				"data: " + `{"choices":[],"usage":{"prompt_tokens":13,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3}}}` + "\n\n")
			r.Body = io.NopCloser(io.MultiReader(r.Body, retryHandlerReadError{}))
			return r
		}
		upstreamB2HandlerServe(t, f, context.Background(), "/v1/chat/completions", `{"model":"gpt-5.1","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
		require.Equal(t, []int64{1}, f.up.ids)
		upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
	})
}

func TestUpstreamB2ImagesPartialHandlerNoRestart(t *testing.T) {
	f, charges := upstreamB2HandlerFixture(t, 2, 10, false)
	f.up.respond = func(_ *http.Request, _ int64, n int) *http.Response {
		if n == 1 {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"b2-partial-image"}},
				Body: io.NopCloser(strings.NewReader(`{"created":1,"usage":{"input_tokens":13,"output_tokens":7},"data":[{"b64_json":"aW1n"}]}`))}
		}
		return upstreamB2HandlerHTTPError(503)
	}
	rec := upstreamB2HandlerServe(t, f, context.Background(), "/v1/images/generations", `{"model":"gpt-image-2","prompt":"draw","n":3}`)
	require.Equal(t, []int64{1, 1}, f.up.ids, rec.Body.String())
	require.Equal(t, 1, f.slots.acquired, "split subcalls use the existing outer slot")
	require.Equal(t, 1, f.slots.released)
	require.Zero(t, f.h.gatewayService.SnapshotOpenAIAccountSchedulerMetrics().RuntimeStatsAccountCount,
		"partial result plus overload must not turn into a synthetic success sample")
	upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 0, 1)
}

func TestUpstreamB2SharedLoopBudgetsAndNeutrality(t *testing.T) {
	for _, status := range []int{400, 502, 503, 529} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			state := NewFailoverState(1, false)
			temp := &mockTempUnscheduler{}
			fo := &service.UpstreamFailoverError{StatusCode: status, RetryableOnSameAccount: true,
				SameAccountRetryLimit: 3, SameAccountRetryBackoffBase: time.Second,
				SameAccountRetryDelay: time.Nanosecond, RequestScopedTransient: true}
			for n := 0; n < 3; n++ {
				require.Equal(t, FailoverContinue, state.HandleFailoverError(context.Background(), temp, 1, service.PlatformOpenAI, fo))
				require.Zero(t, state.SwitchCount)
			}
			require.Equal(t, FailoverContinue, state.HandleFailoverError(context.Background(), temp, 1, service.PlatformOpenAI, fo))
			require.Equal(t, 1, state.SwitchCount)
			require.Contains(t, state.FailedAccountIDs, int64(1))
			require.Empty(t, temp.calls)
			for n := 0; n < 3; n++ {
				require.Equal(t, FailoverContinue, state.HandleFailoverError(context.Background(), temp, 2, service.PlatformOpenAI, fo))
			}
			require.Equal(t, FailoverExhausted, state.HandleFailoverError(context.Background(), temp, 2, service.PlatformOpenAI, fo))
			require.Empty(t, temp.calls)
			require.Equal(t, 3, state.SameAccountRetryCount[1])
			require.Equal(t, 3, state.SameAccountRetryCount[2])
		})
	}
	require.Equal(t, 5, sameAccountRetryLimit(&service.UpstreamFailoverError{SameAccountRetryLimit: 5}, 3))
	require.Equal(t, 3, service.OpenAIStreamRetryLimit)
	require.Equal(t, 3*time.Second, sameAccountRetryDelay(&service.UpstreamFailoverError{SameAccountRetryBackoffBase: time.Second}, 3))
	require.Equal(t, 8*time.Second, sameAccountRetryDelay(&service.UpstreamFailoverError{SameAccountRetryDelay: 8 * time.Second, SameAccountRetryBackoffBase: time.Second}, 3))
	temp := &mockTempUnscheduler{}
	state := NewFailoverState(0, false)
	state.SameAccountRetryCount[1] = 3
	require.Equal(t, FailoverExhausted, state.HandleFailoverError(context.Background(), temp, 1, service.PlatformOpenAI,
		&service.UpstreamFailoverError{StatusCode: 429, RetryableOnSameAccount: true}))
	require.Len(t, temp.calls, 1, "non-request-scoped policy retains old side effects")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Equal(t, FailoverCanceled, NewFailoverState(0, false).HandleFailoverError(ctx, temp, 1, service.PlatformOpenAI,
		&service.UpstreamFailoverError{RetryableOnSameAccount: true, RequestScopedTransient: true, SameAccountRetryDelay: time.Second}))
}

func TestUpstreamB2CancelAndAcquireFailures(t *testing.T) {
	for _, mode := range []string{"cancel_delay", "acquire_error", "wait_error", "cancel_wait"} {
		t.Run(mode, func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 1, 0, false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				if mode == "cancel_delay" {
					time.AfterFunc(20*time.Millisecond, cancel)
				} else {
					f.slots.acquire = func(context.Context, int64) (bool, error) {
						if mode == "acquire_error" || mode == "wait_error" && f.slots.waitAdded > 0 {
							return false, errors.New("b2 acquire failure")
						}
						return false, nil
					}
					if mode == "cancel_wait" {
						f.slots.onWait = cancel
					}
				}
				r := upstreamB2HandlerHTTPError(503)
				if mode == "cancel_delay" {
					r.Header.Set("Retry-After", "1")
				}
				return r
			}
			upstreamB2HandlerServe(t, f, ctx, "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`)
			require.Equal(t, []int64{1}, f.up.ids)
			require.Equal(t, 1, f.slots.acquired)
			require.Equal(t, 1, f.slots.released, "never release an unacquired slot")
			require.Equal(t, 1, f.up.closed)
			upstreamB2AssertUsage(t, f, charges, 0, 0, 0, 0, 0)
		})
	}
}

func TestUpstreamB2PinnedAndLossOfEligibility(t *testing.T) {
	for _, mode := range []string{"original_pin", "lost_ownership", "wrong_group"} {
		t.Run(mode, func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 2, 10, false)
			ctx := context.Background()
			if mode == "original_pin" {
				ctx = context.WithValue(ctx, ctxkey.APIKeyPinnedAccountID, int64(1))
			}
			if mode == "lost_ownership" {
				ctx = context.WithValue(ctx, ctxkey.APIKeyAccountPoolStrategy, service.AccountPoolStrategyPrivateFirst)
				owner := int64(7)
				f.repo.accounts[0].OwnerUserID = &owner
			}
			f.up.respond = func(_ *http.Request, _ int64, n int) *http.Response {
				if n == 1 {
					switch mode {
					case "original_pin":
						f.repo.accounts[0].Schedulable = false
					case "lost_ownership":
						owner := int64(8)
						f.repo.accounts[0].OwnerUserID = &owner
					case "wrong_group":
						f.repo.accounts[0].GroupIDs = []int64{200}
					}
					return upstreamB2HandlerHTTPError(503)
				}
				return retryHandlerResponse(upstreamB2HandlerSuccess)
			}
			rec := upstreamB2HandlerServe(t, f, ctx, "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`)
			if mode == "original_pin" {
				require.Equal(t, []int64{1}, f.up.ids, rec.Body.String())
				upstreamB2AssertUsage(t, f, charges, 0, 0, 0, 0, 0)
			} else {
				require.Equal(t, []int64{1, 2}, f.up.ids, rec.Body.String())
				upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
			}
		})
	}
}

func TestUpstreamB2LocalHTTPChatMessages(t *testing.T) {
	for _, branch := range []string{"chat", "messages"} {
		t.Run(branch, func(t *testing.T) {
			var calls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				require.True(t, strings.HasSuffix(r.URL.Path, "/responses"))
				if calls.Add(1) == 1 {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After-Ms", "2")
					w.WriteHeader(503)
					_, _ = io.WriteString(w, upstreamB2HandlerOverload)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, upstreamB2HandlerSuccess)
			}))
			defer upstream.Close()
			f, charges := upstreamB2HandlerFixture(t, 1, 0, false)
			f.repo.accounts[0].Credentials["base_url"] = upstream.URL
			f.up.respond = func(req *http.Request, _ int64, _ int) *http.Response {
				resp, err := upstream.Client().Do(req)
				require.NoError(t, err)
				return resp
			}
			path, body := "/v1/chat/completions", `{"model":"gpt-5.1","stream":true,"messages":[{"role":"user","content":"hi"}]}`
			if branch == "messages" {
				path, body = "/v1/messages", `{"model":"gpt-5.1","stream":true,"max_tokens":32,"messages":[{"role":"user","content":"hi"}]}`
			}
			rec := upstreamB2HandlerServe(t, f, context.Background(), path, body)
			require.Equal(t, []int64{1, 1}, f.up.ids, rec.Body.String())
			require.Equal(t, int32(2), calls.Load())
			require.Equal(t, 2, f.slots.acquired)
			require.Equal(t, 2, f.slots.released)
			upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
		})
	}
}

func TestUpstreamB2OtherHTTPHandlerExhaustion(t *testing.T) {
	for _, branch := range []string{"chat_responses", "chat_raw", "messages", "images"} {
		t.Run(branch, func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 1, 0, false)
			path, body := "/v1/chat/completions", `{"model":"gpt-5.1","stream":true,"messages":[{"role":"user","content":"hi"}]}`
			switch branch {
			case "chat_raw":
				f.repo.accounts[0].Extra["openai_responses_supported"] = false
			case "messages":
				path, body = "/v1/messages", `{"model":"gpt-5.1","stream":true,"max_tokens":32,"messages":[{"role":"user","content":"hi"}]}`
			case "images":
				path, body = "/v1/images/generations", `{"model":"gpt-image-2","prompt":"draw","n":1}`
			}
			f.up.respond = func(*http.Request, int64, int) *http.Response { return upstreamB2HandlerHTTPError(503) }
			rec := upstreamB2HandlerServe(t, f, context.Background(), path, body)
			require.Equal(t, []int64{1, 1, 1, 1}, f.up.ids, rec.Body.String())
			require.Equal(t, 4, f.slots.acquired)
			require.Equal(t, 4, f.slots.released)
			upstreamB2AssertUsage(t, f, charges, 0, 0, 0, 0, 0)
			require.Zero(t, f.h.gatewayService.SnapshotOpenAIAccountSchedulerMetrics().RuntimeStatsAccountCount)
		})
	}
}

func TestUpstreamB2HandlerBoundedSwitchAndPool(t *testing.T) {
	for _, pool := range []bool{false, true} {
		t.Run(fmt.Sprint(pool), func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 2, 1, false)
			f.repo.accounts[0].Credentials["pool_mode"] = pool
			f.up.respond = func(_ *http.Request, id int64, _ int) *http.Response {
				if id == 1 {
					return upstreamB2HandlerHTTPError(503)
				}
				return retryHandlerResponse(upstreamB2HandlerSuccess)
			}
			rec := upstreamB2HandlerServe(t, f, context.Background(), "/v1/responses", `{"model":"gpt-5.1","stream":true,"input":"hi"}`)
			require.Equal(t, []int64{1, 1, 1, 1, 2}, f.up.ids, rec.Body.String())
			require.Equal(t, 5, f.slots.acquired)
			require.Equal(t, 5, f.slots.released)
			metrics := f.h.gatewayService.SnapshotOpenAIAccountSchedulerMetrics()
			require.Equal(t, int64(1), metrics.AccountSwitchTotal)
			require.Equal(t, 1, metrics.RuntimeStatsAccountCount, "only the final successful account contributes health state")
			upstreamB2AssertUsage(t, f, charges, 1, 13, 7, 3, 0)
		})
	}
}
