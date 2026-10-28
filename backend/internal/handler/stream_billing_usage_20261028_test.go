package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func streamBillingPool(t *testing.T, policy string, sample int) *service.UsageRecordWorkerPool {
	t.Helper()
	p := service.NewUsageRecordWorkerPoolWithOptions(service.UsageRecordWorkerPoolOptions{
		WorkerCount: 1, QueueSize: 1, TaskTimeout: time.Second,
		OverflowPolicy: policy, OverflowSamplePercent: sample,
	})
	t.Cleanup(p.Stop)
	return p
}

func streamBillingFillPool(t *testing.T, p *service.UsageRecordWorkerPool) func() {
	t.Helper()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	require.Equal(t, service.UsageRecordSubmitModeEnqueued, p.Submit(func(context.Context) {
		close(entered)
		<-release
	}))
	<-entered
	require.Equal(t, service.UsageRecordSubmitModeEnqueued, p.Submit(func(context.Context) {}))
	return unblock
}

func TestStreamBillingHandlerStoppedFallbackDetachedAndPanic(t *testing.T) {
	for _, name := range []string{"gateway", "openai", "mandatory"} {
		t.Run(name, func(t *testing.T) {
			p := streamBillingPool(t, "drop", 0)
			p.Stop()
			g, o := &GatewayHandler{usageRecordWorkerPool: p}, &OpenAIGatewayHandler{usageRecordWorkerPool: p}
			submit := g.submitUsageRecordTask
			if name == "openai" {
				submit = o.submitUsageRecordTask
			} else if name == "mandatory" {
				submit = o.submitMandatoryUsageRecordTask
			}
			parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey.RequestID, "stopped-request"))
			cancel()
			calls, releases := 0, 0
			mode := submit(parent, func(ctx context.Context) {
				defer func() { releases++ }()
				calls++
				require.NoError(t, ctx.Err())
				deadline, ok := ctx.Deadline()
				require.True(t, ok)
				require.WithinDuration(t, time.Now().Add(10*time.Second), deadline, time.Second)
				require.Equal(t, "stopped-request", ctx.Value(ctxkey.RequestID))
				panic("recorder failed after reservation handoff")
			})
			require.Equal(t, service.UsageRecordSubmitModeSync, mode, "recovered task still owns the reservation")
			require.Equal(t, 1, calls)
			require.Equal(t, 1, releases)
			require.EqualValues(t, 1, p.Stats().DroppedPoolStopped)
		})
	}
}

func TestStreamBillingHandlerExplicitOverflowPreserved(t *testing.T) {
	for _, name := range []string{"gateway", "openai"} {
		for _, policy := range []string{"drop", "sample"} {
			t.Run(name+"/"+policy, func(t *testing.T) {
				p := streamBillingPool(t, policy, 1)
				unblock := streamBillingFillPool(t, p)
				submit := (&GatewayHandler{usageRecordWorkerPool: p}).submitUsageRecordTask
				if name == "openai" {
					submit = (&OpenAIGatewayHandler{usageRecordWorkerPool: p}).submitUsageRecordTask
				}
				calls := 0
				for range 100 {
					submit(context.Background(), func(context.Context) { calls++ })
				}
				want := 0
				if policy == "sample" {
					want = 1
				}
				require.Equal(t, want, calls, "only configured sample tasks run inline")
				require.EqualValues(t, 100-want, p.Stats().DroppedQueueFull)
				require.Zero(t, p.Stats().DroppedPoolStopped)
				unblock()
			})
		}
	}
}

func TestStreamBillingHandlerMandatoryImagesBothDropReasons(t *testing.T) {
	for _, stopped := range []bool{false, true} {
		t.Run(fmt.Sprint(stopped), func(t *testing.T) {
			p := streamBillingPool(t, "drop", 0)
			if stopped {
				p.Stop()
			} else {
				defer streamBillingFillPool(t, p)()
			}
			h := &OpenAIGatewayHandler{usageRecordWorkerPool: p}
			calls := 0
			mode := h.submitOpenAIUsageRecordTask(context.Background(), &service.OpenAIForwardResult{ImageCount: 1}, func(ctx context.Context) {
				require.NoError(t, ctx.Err())
				calls++
			})
			require.Equal(t, service.UsageRecordSubmitModeSync, mode)
			require.Equal(t, 1, calls)
		})
	}
}

func TestStreamBillingHandlerShutdownRaceRecordOnce(t *testing.T) {
	for _, gateway := range []bool{false, true} {
		t.Run(fmt.Sprint(gateway), func(t *testing.T) {
			p := streamBillingPool(t, "sync", 0)
			submit := (&OpenAIGatewayHandler{usageRecordWorkerPool: p}).submitUsageRecordTask
			if gateway {
				submit = (&GatewayHandler{usageRecordWorkerPool: p}).submitUsageRecordTask
			}
			const count = 96
			var records [count]atomic.Int32
			start := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(count + 1)
			go func() { defer wg.Done(); <-start; p.Stop() }()
			for i := range count {
				go func() {
					defer wg.Done()
					<-start
					mode := submit(context.Background(), func(ctx context.Context) {
						if ctx.Err() == nil {
							records[i].Add(1)
						}
					})
					if mode != service.UsageRecordSubmitModeSync && mode != service.UsageRecordSubmitModeEnqueued {
						t.Errorf("billing lost during shutdown: %q", mode)
					}
				}()
			}
			close(start)
			wg.Wait()
			p.Stop()
			for i := range records {
				require.EqualValues(t, 1, records[i].Load(), "submission %d must record once", i)
			}
		})
	}
}

type streamBillingGroupRepo struct {
	service.GroupRepository
	group *service.Group
}

type streamBillingAccounts struct {
	*openAIWSFailoverHandlerAccountRepoStub
}

type streamBillingUpstream struct{ *retryHandlerUpstream }

func (u *streamBillingUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

func (r *streamBillingAccounts) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, _ int64, platforms []string) ([]service.Account, error) {
	var result []service.Account
	for _, platform := range platforms {
		accounts, err := r.ListSchedulableByPlatform(ctx, platform)
		if err != nil {
			return nil, err
		}
		result = append(result, accounts...)
	}
	return result, nil
}

func (r *streamBillingGroupRepo) GetByID(context.Context, int64) (*service.Group, error) {
	return r.group, nil
}
func (r *streamBillingGroupRepo) GetByIDLite(context.Context, int64) (*service.Group, error) {
	return r.group, nil
}

func newStreamBillingGatewayFixture(t *testing.T, passthrough bool, repos ...service.UsageLogRepository) (*retryHandlerFixture, *GatewayHandler, *upstreamB2HandlerCharges) {
	t.Helper()
	f := newRetryHandlerFixture(t, 1, 0, false)
	t.Cleanup(f.h.billingCacheService.Stop)
	f.cfg.RunMode = ""
	f.repo.accounts[0].Platform = service.PlatformAnthropic
	f.repo.accounts[0].Extra = map[string]any{"anthropic_passthrough": passthrough}
	group := &service.Group{ID: 100, Platform: service.PlatformAnthropic, Status: service.StatusActive, RateMultiplier: 1}
	groups := &streamBillingGroupRepo{group: group}
	users := &upstreamB2HandlerUsers{}
	charges := &upstreamB2HandlerCharges{}
	cache := service.NewBillingCacheService(nil, users, nil, nil, nil, nil, f.cfg)
	t.Cleanup(cache.Stop)
	concurrency := service.NewConcurrencyService(f.slots)
	var usageRepo service.UsageLogRepository = f.usageRepo
	if len(repos) > 0 {
		usageRepo = repos[0]
	}
	svc := service.NewGatewayService(&streamBillingAccounts{f.repo}, groups, usageRepo, charges, users, nil, nil, nil, f.cfg,
		nil, concurrency, service.NewBillingService(f.cfg, nil), service.NewRateLimitService(nil, nil, f.cfg, nil, nil),
		cache, nil, &streamBillingUpstream{f.up}, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	h := &GatewayHandler{gatewayService: svc, billingCacheService: cache, cfg: f.cfg,
		concurrencyHelper: NewConcurrencyHelper(concurrency, SSEPingFormatNone, 0)}
	key := &service.APIKey{ID: 1, UserID: 7, GroupID: &group.ID, Group: group,
		User: &service.User{ID: 7, Balance: 1000, Status: service.StatusActive}}
	f.router = gin.New()
	f.router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), key)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7, Concurrency: 1})
		c.Next()
	})
	f.router.POST("/v1/messages", h.Messages)
	f.router.POST("/v1/chat/completions", h.ChatCompletions)
	f.router.POST("/v1/responses", h.Responses)
	return f, h, charges
}

const streamBillingAnthropicStart = "event: message_start\ndata: " + `{"type":"message_start","message":{"id":"msg_billing","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[],"usage":{"input_tokens":13,"cache_read_input_tokens":3,"cache_creation_input_tokens":4,"cache_creation":{"ephemeral_5m_input_tokens":4}}}}` + "\n\n"
const streamBillingAnthropicDelta = "event: content_block_start\ndata: " + `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
	"event: content_block_delta\ndata: " + `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"paid-output"}}` + "\n\n" +
	"event: message_delta\ndata: " + `{"type":"message_delta","delta":{},"usage":{"output_tokens":7}}` + "\n\n"

var streamBillingRequests = []struct{ path, body string }{
	{"/v1/messages", `{"model":"claude-sonnet-4-6","stream":true,"max_tokens":32,"messages":[{"role":"user","content":"hello"}]}`},
	{"/v1/chat/completions", `{"model":"claude-sonnet-4-6","stream":true,"messages":[{"role":"user","content":"hello"}]}`},
	{"/v1/responses", `{"model":"claude-sonnet-4-6","stream":true,"input":"hello"}`},
}

func TestStreamBillingGatewayHandlerObservedPartialRecordOnce(t *testing.T) {
	for _, route := range streamBillingRequests {
		for _, passthrough := range []bool{false, true} {
			for _, terminal := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/pass=%v/terminal=%v", route.path, passthrough, terminal), func(t *testing.T) {
					f, _, charges := newStreamBillingGatewayFixture(t, passthrough)
					f.up.respond = func(*http.Request, int64, int) *http.Response {
						body := streamBillingAnthropicStart + streamBillingAnthropicDelta
						if terminal {
							body += "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
						}
						resp := retryHandlerResponse(body)
						resp.Header.Set("X-Request-Id", "handler-billing-request")
						return resp
					}
					rec := upstreamB2HandlerServe(t, f, context.Background(), route.path, route.body)
					require.Len(t, f.up.ids, 1, rec.Body.String())
					require.EqualValues(t, 1, atomic.LoadInt32(&f.slots.releaseUserCalled))
					require.Len(t, charges.commands, 1, rec.Body.String())
					require.Len(t, f.usageRepo.created, 1, rec.Body.String())
					log := <-f.usageRepo.created
					require.Equal(t, 13, log.InputTokens)
					require.Equal(t, 7, log.OutputTokens)
					require.Equal(t, 3, log.CacheReadTokens)
					require.Equal(t, 4, log.CacheCreationTokens)
					require.Equal(t, "handler-billing-request", log.RequestID)
					require.Equal(t, "claude-sonnet-4-6", log.RequestedModel)
					require.NotNil(t, log.DurationMs)
					require.NotNil(t, log.FirstTokenMs)
					require.Positive(t, charges.commands[0].BalanceCost)
					require.Empty(t, f.usageRepo.created)
					if !terminal {
						require.NotContains(t, rec.Body.String(), "response.completed")
						require.NotContains(t, rec.Body.String(), "[DONE]")
						errorFrames := 0
						for _, line := range strings.Split(rec.Body.String(), "\n") {
							if strings.HasPrefix(line, "data:") && (strings.Contains(line, `"error"`) || strings.Contains(line, `"response.failed"`)) {
								errorFrames++
							}
						}
						require.Equal(t, 1, errorFrames, "one terminal error, without a handler duplicate")
					}
				})
			}
		}
	}
}

func TestStreamBillingGatewayHandlerZeroUsageAndFailoverNotBilled(t *testing.T) {
	for _, route := range streamBillingRequests {
		for _, failure := range []string{"zero_observed", "failover"} {
			t.Run(route.path+"/"+failure, func(t *testing.T) {
				f, _, charges := newStreamBillingGatewayFixture(t, true)
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					if failure == "failover" {
						return &http.Response{StatusCode: 503, Header: http.Header{"Content-Type": {"application/json"}},
							Body: io.NopCloser(strings.NewReader(`{"error":{"message":"unavailable"}}`))}
					}
					return retryHandlerResponse(": ping\n\n")
				}
				upstreamB2HandlerServe(t, f, context.Background(), route.path, route.body)
				require.Len(t, f.up.ids, 1)
				require.Empty(t, f.usageRepo.created)
				require.Empty(t, charges.commands)
			})
		}
	}
}

func TestStreamBillingGatewayHandlerLoopbackPartialUsage(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Request-Id", "loopback-partial")
		_, _ = io.WriteString(w, streamBillingAnthropicStart+streamBillingAnthropicDelta)
		w.(http.Flusher).Flush()
		// End the real HTTP/SSE body without a terminal event.
	}))
	defer upstream.Close()
	f, _, charges := newStreamBillingGatewayFixture(t, true)
	f.up.respond = func(*http.Request, int64, int) *http.Response {
		resp, err := upstream.Client().Get(upstream.URL)
		require.NoError(t, err)
		return resp
	}
	route := streamBillingRequests[0]
	upstreamB2HandlerServe(t, f, context.Background(), route.path, route.body)
	require.Len(t, f.up.ids, 1)
	require.Len(t, f.usageRepo.created, 1)
	require.Len(t, charges.commands, 1)
	log := <-f.usageRepo.created
	require.Equal(t, "loopback-partial", log.RequestID)
	require.Equal(t, 13, log.InputTokens)
	require.Equal(t, 7, log.OutputTokens)
}

const streamBillingOpenAIPartial = "data: " + `{"type":"response.output_text.delta","delta":"paid-output"}` + "\n\n" +
	"data: " + `{"type":"response.in_progress","response":{"id":"resp_billing","usage":{"input_tokens":13,"output_tokens":7,"input_tokens_details":{"cached_tokens":3}}}}` + "\n\n"

func streamBillingOpenAIStream(f *retryHandlerFixture, path string) string {
	if path == "/v1/messages" {
		// This adapter's established usage path receives usage on a terminal
		// response; a later reader error must not turn it into a second record.
		return upstreamB2HandlerSuccess
	}
	if path == "/v1/chat/completions" {
		f.repo.accounts[0].Extra["openai_responses_supported"] = false
		return "data: " + `{"id":"chat_billing","choices":[{"index":0,"delta":{"content":"paid-output"}}]}` + "\n\n" +
			"data: " + `{"choices":[],"usage":{"prompt_tokens":13,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3}}}` + "\n\n"
	}
	return streamBillingOpenAIPartial
}

func TestStreamBillingOpenAIHandlerPartialUsageRecordOnce(t *testing.T) {
	for _, path := range []string{"/v1/responses", "/v1/chat/completions", "/v1/messages"} {
		t.Run(path, func(t *testing.T) {
			f, charges := upstreamB2HandlerFixture(t, 1, 0, false)
			f.h.usageRecordWorkerPool = streamBillingPool(t, "drop", 0)
			f.h.usageRecordWorkerPool.Stop()
			stream := streamBillingOpenAIStream(f, path)
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				resp := retryHandlerResponse(stream)
				resp.Header.Set("X-Request-Id", "openai-handler-partial")
				resp.Body = io.NopCloser(io.MultiReader(resp.Body, retryHandlerReadError{}))
				return resp
			}
			body := `{"model":"gpt-5.1","stream":true,"input":"hello","max_tokens":32,"messages":[{"role":"user","content":"hello"}]}`
			rec := upstreamB2HandlerServe(t, f, context.Background(), path, body)
			require.Len(t, f.up.ids, 1, rec.Body.String())
			require.Len(t, f.usageRepo.created, 1, rec.Body.String())
			require.Len(t, charges.commands, 1)
			log := <-f.usageRepo.created
			require.Equal(t, 13, log.InputTokens+log.CacheReadTokens)
			require.Equal(t, 7, log.OutputTokens)
			require.Equal(t, 3, log.CacheReadTokens)
			require.Equal(t, "openai-handler-partial", log.RequestID)
			require.NotContains(t, rec.Body.String(), "response.completed")
			require.NotContains(t, rec.Body.String(), "[DONE]")
		})
	}
}

type streamBillingCommitWriter struct {
	gin.ResponseWriter
	committed bool
	onWrite   func()
}

func (w *streamBillingCommitWriter) Committed() bool { return w.committed }

func (w *streamBillingCommitWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if n > 0 && w.onWrite != nil {
		w.onWrite()
	}
	return n, err
}

func (w *streamBillingCommitWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

type streamBillingCancelEOF struct{ cancel context.CancelFunc }

func (r streamBillingCancelEOF) Read([]byte) (int, error) {
	r.cancel()
	return 0, io.EOF
}

func TestStreamBillingHandlersFirstResponseCommitBoundary(t *testing.T) {
	for _, platform := range []string{"anthropic", "openai"} {
		for _, route := range streamBillingRequests {
			for _, tc := range []struct {
				name              string
				marked, committed bool
				want              int
			}{
				{"buffered_attempt", true, false, 0},
				{"committed_attempt", true, true, 1},
				{"ordinary_client_cancel", false, false, 1},
			} {
				t.Run(platform+route.path+"/"+tc.name, func(t *testing.T) {
					var f *retryHandlerFixture
					var gateway *GatewayHandler
					var charges *upstreamB2HandlerCharges
					body := route.body
					stream := streamBillingAnthropicStart + streamBillingAnthropicDelta
					if platform == "anthropic" {
						f, gateway, charges = newStreamBillingGatewayFixture(t, true)
					} else {
						f, charges = upstreamB2HandlerFixture(t, 1, 0, false)
						body = strings.ReplaceAll(body, "claude-sonnet-4-6", "gpt-5.1")
						stream = streamBillingOpenAIStream(f, route.path)
					}
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					if tc.marked {
						ctx = ctxkey.WithKeyRouteFirstResponseAttempt(ctx)
					}
					var attemptWriter *streamBillingCommitWriter
					f.up.respond = func(*http.Request, int64, int) *http.Response {
						// Prelude arrives after routing, while this attempt is
						// waiting for its first business output.
						_, _ = attemptWriter.WriteString(": buffered-prelude\n\n")
						attemptWriter.onWrite = cancel
						resp := retryHandlerResponse(stream)
						resp.Body = io.NopCloser(io.MultiReader(resp.Body, streamBillingCancelEOF{cancel: cancel}))
						return resp
					}
					router := gin.New()
					router.Use(func(c *gin.Context) {
						attemptWriter = &streamBillingCommitWriter{ResponseWriter: c.Writer, committed: tc.committed}
						c.Writer = attemptWriter
						group := &service.Group{ID: 100, Platform: f.repo.accounts[0].Platform, Status: service.StatusActive,
							RateMultiplier: 1, AllowMessagesDispatch: true}
						key := &service.APIKey{ID: 1, UserID: 7, GroupID: &group.ID, Group: group,
							User: &service.User{ID: 7, Balance: 1000, Status: service.StatusActive}}
						c.Set(string(middleware.ContextKeyAPIKey), key)
						c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7, Concurrency: 1})
						c.Next()
					})
					dispatch := f.h.Responses
					if route.path == "/v1/messages" {
						dispatch = f.h.Messages
					} else if route.path == "/v1/chat/completions" {
						dispatch = f.h.ChatCompletions
					}
					if gateway != nil {
						dispatch = gateway.Responses
						if route.path == "/v1/messages" {
							dispatch = gateway.Messages
						} else if route.path == "/v1/chat/completions" {
							dispatch = gateway.ChatCompletions
						}
					}
					router.POST(route.path, dispatch)
					f.router = router
					rec := upstreamB2HandlerServe(t, f, ctx, route.path, body)
					require.ErrorIs(t, ctx.Err(), context.Canceled, "upstream calls=%v response=%s", f.up.ids, rec.Body.String())
					require.Contains(t, rec.Body.String(), "buffered-prelude")
					require.Len(t, f.up.ids, 1)
					require.Len(t, f.usageRepo.created, tc.want, rec.Body.String())
					require.Len(t, charges.commands, tc.want)
				})
			}
		}
	}
}

type streamBillingTrialRepo struct {
	service.WelfareRepository
	begun, canceled atomic.Int32
}

func (r *streamBillingTrialRepo) BeginNewUserTrial(_ context.Context, userID int64, quota float64, _, requestID string, _ time.Time, _ int) (*service.WelfareNewUserTrial, error) {
	r.begun.Add(1)
	return &service.WelfareNewUserTrial{ID: 9, UserID: userID, QuotaAmount: quota, LastRequestID: requestID}, nil
}
func (r *streamBillingTrialRepo) CancelNewUserTrial(context.Context, int64, string) error {
	r.canceled.Add(1)
	return nil
}

type streamBillingPoorUserRepo struct{ service.UserRepository }

func (*streamBillingPoorUserRepo) GetByID(context.Context, int64) (*service.User, error) {
	return &service.User{ID: 7, Status: service.StatusActive}, nil
}

type streamBillingPanicUsageRepo struct{ service.UsageLogRepository }

func (*streamBillingPanicUsageRepo) Create(context.Context, *service.UsageLog) (bool, error) {
	panic("recorder repository panic")
}

func TestStreamBillingGatewayHandlerTrialReleaseOnce(t *testing.T) {
	for _, route := range streamBillingRequests {
		for _, mode := range []string{"sync", "stopped", "queued", "drop", "panic"} {
			t.Run(route.path+"/"+mode, func(t *testing.T) {
				var repos []service.UsageLogRepository
				if mode == "panic" {
					repos = append(repos, &streamBillingPanicUsageRepo{})
				}
				f, h, _ := newStreamBillingGatewayFixture(t, true, repos...)
				// Execute the actual trial admission and recorder, using simple-mode
				// persistence to avoid pretending an in-memory test is SQL settlement.
				f.cfg.RunMode = config.RunModeSimple
				trials := &streamBillingTrialRepo{}
				settings := &oauthPendingFlowSettingRepoStub{values: map[string]string{
					service.SettingKeyWelfareEnabled: "true", service.SettingKeyWelfareNewUserTrialEnabled: "true",
					service.SettingKeyWelfareNewUserTrialQuotaAmount:            "1",
					service.SettingKeyWelfareNewUserTrialDailySiteQuotaAmount:   "0",
					service.SettingKeyWelfareNewUserTrialDailyIPActivationLimit: "0",
				}}
				cache := service.NewBillingCacheService(nil, &streamBillingPoorUserRepo{}, nil, nil, nil, nil, &config.Config{})
				t.Cleanup(cache.Stop)
				h.billingCacheService = cache
				h.welfareService = service.NewWelfareService(trials, nil, nil, settings, nil, nil, nil)
				var unblock func()
				if mode == "queued" || mode == "drop" || mode == "stopped" {
					h.usageRecordWorkerPool = streamBillingPool(t, "drop", 0)
					if mode == "stopped" {
						h.usageRecordWorkerPool.Stop()
					} else if mode == "drop" {
						unblock = streamBillingFillPool(t, h.usageRecordWorkerPool)
						defer unblock()
					} else {
						entered, release := make(chan struct{}), make(chan struct{})
						var once sync.Once
						unblock = func() { once.Do(func() { close(release) }) }
						t.Cleanup(unblock)
						h.usageRecordWorkerPool.Submit(func(context.Context) { close(entered); <-release })
						<-entered
					}
				}
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					return retryHandlerResponse(streamBillingAnthropicStart + streamBillingAnthropicDelta)
				}
				upstreamB2HandlerServe(t, f, context.Background(), route.path, route.body)
				require.EqualValues(t, 1, trials.begun.Load())
				if mode == "queued" {
					require.Zero(t, trials.canceled.Load(), "reservation stays held until usage executes")
					unblock()
					h.usageRecordWorkerPool.Stop()
				}
				require.EqualValues(t, 1, trials.canceled.Load(), "one terminal owner releases the trial")
				if mode == "drop" || mode == "panic" {
					require.Empty(t, f.usageRepo.created)
				} else {
					require.Len(t, f.usageRepo.created, 1)
					require.Equal(t, service.BillingTypeNewUserTrial, (<-f.usageRepo.created).BillingType)
				}
			})
		}
	}
}

func TestStreamBillingOpenAIHandlerTrialReleaseOnce(t *testing.T) {
	for _, route := range streamBillingRequests {
		t.Run(route.path, func(t *testing.T) {
			f, _ := upstreamB2HandlerFixture(t, 1, 0, false)
			f.cfg.RunMode = config.RunModeSimple
			trials := &streamBillingTrialRepo{}
			settings := &oauthPendingFlowSettingRepoStub{values: map[string]string{
				service.SettingKeyWelfareEnabled: "true", service.SettingKeyWelfareNewUserTrialEnabled: "true",
				service.SettingKeyWelfareNewUserTrialQuotaAmount:            "1",
				service.SettingKeyWelfareNewUserTrialDailySiteQuotaAmount:   "0",
				service.SettingKeyWelfareNewUserTrialDailyIPActivationLimit: "0",
			}}
			cache := service.NewBillingCacheService(nil, &streamBillingPoorUserRepo{}, nil, nil, nil, nil, &config.Config{})
			t.Cleanup(cache.Stop)
			f.h.billingCacheService = cache
			f.h.welfareService = service.NewWelfareService(trials, nil, nil, settings, nil, nil, nil)
			f.h.usageRecordWorkerPool = streamBillingPool(t, "drop", 0)
			f.h.usageRecordWorkerPool.Stop()
			stream := streamBillingOpenAIStream(f, route.path)
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				resp := retryHandlerResponse(stream)
				resp.Body = io.NopCloser(io.MultiReader(resp.Body, retryHandlerReadError{}))
				return resp
			}
			body := strings.ReplaceAll(route.body, "claude-sonnet-4-6", "gpt-5.1")
			upstreamB2HandlerServe(t, f, context.Background(), route.path, body)
			require.EqualValues(t, 1, trials.begun.Load())
			require.EqualValues(t, 1, trials.canceled.Load())
			require.Len(t, f.usageRepo.created, 1)
			require.Equal(t, service.BillingTypeNewUserTrial, (<-f.usageRepo.created).BillingType)
		})
	}
}

func TestStreamBillingTrialReleaseCallbackIdempotent(t *testing.T) {
	for _, openai := range []bool{false, true} {
		t.Run(fmt.Sprint(openai), func(t *testing.T) {
			trials := &streamBillingTrialRepo{}
			settings := &oauthPendingFlowSettingRepoStub{values: map[string]string{
				service.SettingKeyWelfareEnabled: "true", service.SettingKeyWelfareNewUserTrialEnabled: "true",
				service.SettingKeyWelfareNewUserTrialQuotaAmount:            "1",
				service.SettingKeyWelfareNewUserTrialDailySiteQuotaAmount:   "0",
				service.SettingKeyWelfareNewUserTrialDailyIPActivationLimit: "0",
			}}
			cache := service.NewBillingCacheService(nil, &streamBillingPoorUserRepo{}, nil, nil, nil, nil, &config.Config{})
			t.Cleanup(cache.Stop)
			welfare := service.NewWelfareService(trials, nil, nil, settings, nil, nil, nil)
			check := (&GatewayHandler{billingCacheService: cache, welfareService: welfare}).checkBillingEligibilityOrBeginTrial
			if openai {
				check = (&OpenAIGatewayHandler{billingCacheService: cache, welfareService: welfare}).checkBillingEligibilityOrBeginTrial
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/messages", nil)
			session, release, err := check(c, &service.APIKey{User: &service.User{ID: 7}}, nil, nil, "stream_billing_test")
			require.NoError(t, err)
			require.NotNil(t, session)
			require.NotNil(t, release)
			var wg sync.WaitGroup
			for range 16 {
				wg.Go(release)
			}
			wg.Wait()
			require.EqualValues(t, 1, trials.canceled.Load(), "including repeated WebSocket AfterTurn and request cleanup")
		})
	}
}
