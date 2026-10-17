package handler

import (
	"context"
	"errors"
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
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const retryHandlerFailure = "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"Upstream request failed\"}}}\n\n"
const retryHandlerSuccess = "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_success\"}}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_success\",\"output\":[]}}\n\n"

type retryHandlerUpstream struct {
	service.HTTPUpstream
	ids     []int64
	times   []time.Time
	closed  int
	respond func(*http.Request, int64, int) *http.Response
}

func (u *retryHandlerUpstream) Do(req *http.Request, _ string, id int64, _ int) (*http.Response, error) {
	u.ids = append(u.ids, id)
	u.times = append(u.times, time.Now())
	resp := u.respond(req, id, len(u.ids))
	resp.Body = &retryHandlerBody{ReadCloser: resp.Body, close: func() { u.closed++ }}
	return resp, nil
}

type retryHandlerBody struct {
	io.ReadCloser
	close func()
}

func (b *retryHandlerBody) Close() error { b.close(); return b.ReadCloser.Close() }

func retryHandlerResponse(body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body))}
}

type retryHandlerSlots struct {
	concurrencyCacheMock
	mu                                         sync.Mutex
	held                                       map[int64]int
	acquired, released, waitAdded, waitRemoved int
	releaseTimes                               []time.Time
	acquire                                    func(context.Context, int64) (bool, error)
	onWait                                     func()
}

func (s *retryHandlerSlots) AcquireAccountSlot(ctx context.Context, id int64, _ int, _ string) (bool, error) {
	ok, err := true, error(nil)
	if s.acquire != nil {
		ok, err = s.acquire(ctx, id)
	}
	if ok && err == nil {
		s.mu.Lock()
		s.acquired++
		s.held[id]++
		s.mu.Unlock()
	}
	return ok, err
}
func (s *retryHandlerSlots) ReleaseAccountSlot(_ context.Context, id int64, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.released++
	s.held[id]--
	s.releaseTimes = append(s.releaseTimes, time.Now())
	return nil
}
func (s *retryHandlerSlots) IncrementAccountWaitCount(_ context.Context, _ int64, _ int) (bool, error) {
	s.mu.Lock()
	s.waitAdded++
	s.mu.Unlock()
	if s.onWait != nil {
		s.onWait()
	}
	return true, nil
}
func (s *retryHandlerSlots) DecrementAccountWaitCount(context.Context, int64) error {
	s.mu.Lock()
	s.waitRemoved++
	s.mu.Unlock()
	return nil
}

type retryHandlerFixture struct {
	h         *OpenAIGatewayHandler
	cfg       *config.Config
	repo      *openAIWSFailoverHandlerAccountRepoStub
	up        *retryHandlerUpstream
	slots     *retryHandlerSlots
	router    *gin.Engine
	usageRepo *openAIWSUsageHandlerUsageLogRepoStub
	onWrite   func([]byte)
}

func newRetryHandlerFixture(t *testing.T, count, switches int, passthrough bool) *retryHandlerFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.RunMode = config.RunModeSimple
	cfg.Default.RateMultiplier = 1
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.MaxLineSize = 64 * 1024
	cfg.Gateway.Scheduling.FallbackWaitTimeout = 100 * time.Millisecond
	cfg.Gateway.Scheduling.FallbackMaxWaiting = 2
	repo := &openAIWSFailoverHandlerAccountRepoStub{}
	for i := 1; i <= count; i++ {
		repo.accounts = append(repo.accounts, service.Account{
			ID: int64(i), Name: fmt.Sprintf("retry-%d", i), Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
			Status: service.StatusActive, Schedulable: true, Concurrency: 1, Priority: i, GroupIDs: []int64{100},
			Credentials: map[string]any{"api_key": "sk-local-test", "base_url": "http://127.0.0.1:1"},
			Extra:       map[string]any{"openai_passthrough": passthrough, "responses_supported": true},
		})
	}
	slots := &retryHandlerSlots{held: map[int64]int{}}
	slots.acquireUserSlotFn = func(context.Context, int64, int, string) (bool, error) { return true, nil }
	concurrency := service.NewConcurrencyService(slots)
	billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg)
	up := &retryHandlerUpstream{respond: func(*http.Request, int64, int) *http.Response { return retryHandlerResponse(retryHandlerFailure) }}
	usageRepo := &openAIWSUsageHandlerUsageLogRepoStub{created: make(chan *service.UsageLog, 32)}
	gateway := service.NewOpenAIGatewayService(repo, usageRepo, nil, nil, nil, nil, nil, cfg, nil, concurrency,
		service.NewBillingService(cfg, nil), nil, billingCache, up, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil)
	h := &OpenAIGatewayHandler{gatewayService: gateway, billingCacheService: billingCache,
		apiKeyService: &service.APIKeyService{}, concurrencyHelper: NewConcurrencyHelper(concurrency, SSEPingFormatNone, time.Second),
		maxAccountSwitches: switches, cfg: cfg}
	group := int64(100)
	key := &service.APIKey{ID: 1, GroupID: &group, User: &service.User{ID: 7, Status: service.StatusActive},
		Group: &service.Group{ID: 100, Platform: service.PlatformOpenAI, Status: service.StatusActive}}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), key)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7, Concurrency: 1})
		c.Next()
	})
	router.POST("/v1/responses", h.Responses)
	return &retryHandlerFixture{h: h, cfg: cfg, repo: repo, up: up, slots: slots, router: router, usageRepo: usageRepo}
}
func (f *retryHandlerFixture) run(t *testing.T, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"gpt-5.1","stream":true,"input":"hello"}`)).WithContext(ctx)
	var writer http.ResponseWriter = rec
	if f.onWrite != nil {
		writer = &retryHandlerNotifyWriter{ResponseRecorder: rec, onWrite: f.onWrite}
	}
	f.router.ServeHTTP(writer, req)
	require.Eventually(t, func() bool {
		f.slots.mu.Lock()
		defer f.slots.mu.Unlock()
		return f.slots.acquired == f.slots.released && atomic.LoadInt32(&f.slots.releaseUserCalled) == 1
	}, time.Second, time.Millisecond, "all acquired slots must be released exactly once")
	f.slots.mu.Lock()
	defer f.slots.mu.Unlock()
	for _, held := range f.slots.held {
		require.Zero(t, held)
	}
	require.Equal(t, f.slots.waitAdded, f.slots.waitRemoved)
	require.Equal(t, len(f.up.ids), f.up.closed, "every upstream body closes before retry")
	return rec
}

type retryHandlerNotifyWriter struct {
	*httptest.ResponseRecorder
	onWrite func([]byte)
}

func (w *retryHandlerNotifyWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseRecorder.Write(p)
	if err == nil {
		w.onWrite(p)
	}
	return n, err
}

func TestOpenAIPreOutputRetryFourAttempts(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			f := newRetryHandlerFixture(t, 1, 0, passthrough)
			// Every dispatch must have reacquired its slot; the previous one
			// must already have been released, including throughout backoff.
			f.up.respond = func(_ *http.Request, id int64, n int) *http.Response {
				f.slots.mu.Lock()
				require.Equal(t, 1, f.slots.held[id])
				require.Equal(t, n-1, f.slots.released)
				f.slots.mu.Unlock()
				require.Equal(t, n-1, f.up.closed)
				return retryHandlerResponse(retryHandlerFailure)
			}
			rec := f.run(t, context.Background())
			require.Equal(t, []int64{1, 1, 1, 1}, f.up.ids, rec.Body.String())
			for i, delay := range []time.Duration{500 * time.Millisecond, time.Second, 1500 * time.Millisecond} {
				require.GreaterOrEqual(t, f.up.times[i+1].Sub(f.up.times[i]), delay-20*time.Millisecond)
				require.GreaterOrEqual(t, f.up.times[i+1].Sub(f.slots.releaseTimes[i]), delay-20*time.Millisecond)
			}
			require.NotContains(t, rec.Body.String(), "response.created")
		})
	}
}

func TestOpenAIPreOutputRetryFourteenAttempts(t *testing.T) {
	f := newRetryHandlerFixture(t, 11, 10, false)
	f.run(t, context.Background())
	require.Equal(t, []int64{1, 1, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, f.up.ids)
	require.Equal(t, 14, f.slots.acquired)
}

func TestOpenAIPreOutputRetryPinnedSuccess(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			f := newRetryHandlerFixture(t, 2, 10, passthrough)
			f.up.respond = func(_ *http.Request, _ int64, n int) *http.Response {
				preamble := strings.Repeat("data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 100)
				if n == 1 {
					return retryHandlerResponse(": heartbeat\n\n" + preamble + retryHandlerFailure)
				}
				return retryHandlerResponse(retryHandlerSuccess)
			}
			rec := f.run(t, context.Background())
			require.Equal(t, []int64{1, 1}, f.up.ids)
			require.Contains(t, rec.Body.String(), "resp_success")
			require.NotContains(t, rec.Body.String(), "resp_failed")
			require.NotContains(t, rec.Body.String(), "response.failed")
			require.Equal(t, 2, f.slots.acquired)
		})
	}
}

func TestOpenAIPreOutputRetryInvalidTargetSharedBudget(t *testing.T) {
	f := newRetryHandlerFixture(t, 3, 10, false)
	f.up.respond = func(_ *http.Request, id int64, n int) *http.Response {
		if id == 1 {
			f.repo.accounts[0].Schedulable = false
		}
		if id == 3 {
			return retryHandlerResponse(retryHandlerSuccess)
		}
		return retryHandlerResponse(retryHandlerFailure)
	}
	rec := f.run(t, context.Background())
	require.Equal(t, []int64{1, 2, 2, 2, 3}, f.up.ids, rec.Body.String())
	require.GreaterOrEqual(t, f.up.times[1].Sub(f.up.times[0]), 480*time.Millisecond)
	require.GreaterOrEqual(t, f.up.times[2].Sub(f.up.times[1]), 980*time.Millisecond)
	require.GreaterOrEqual(t, f.up.times[3].Sub(f.up.times[2]), 1480*time.Millisecond)
}

func TestOpenAIPreOutputRetryInvalidTargetNoFreeSwitch(t *testing.T) {
	f := newRetryHandlerFixture(t, 2, 0, false)
	f.up.respond = func(*http.Request, int64, int) *http.Response {
		f.repo.accounts[0].Schedulable = false
		return retryHandlerResponse(retryHandlerFailure)
	}
	rec := f.run(t, context.Background())
	require.Equal(t, []int64{1}, f.up.ids)
	require.Contains(t, rec.Body.String(), "Upstream")
}

func TestOpenAIPreOutputRetryPoolDispatch(t *testing.T) {
	for _, tc := range []struct {
		name        string
		count       any
		statuses    []any
		setStatuses bool
		payload     string
		want        int
	}{
		{name: "zero", count: 0, payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 1},
		{name: "empty", count: 3, statuses: []any{}, setStatuses: true, payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 1},
		{name: "default", count: 2, payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 3},
		{name: "missing_count_default", payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 4},
		{name: "custom_429", count: 1, statuses: []any{429}, setStatuses: true, payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 2},
		{name: "custom_excludes_429", count: 3, statuses: []any{403}, setStatuses: true, payload: `{"error":{"code":"rate_limit_exceeded"}}`, want: 1},
		{name: "generic_default", count: 3, payload: `{"error":{"message":"Upstream request failed"}}`, want: 1},
		{name: "generic_custom_502", count: 3, statuses: []any{502}, setStatuses: true, payload: `{"error":{"message":"Upstream request failed"}}`, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newRetryHandlerFixture(t, 1, 0, false)
			creds := f.repo.accounts[0].Credentials
			creds["pool_mode"] = true
			creds["pool_mode_retry_count"] = tc.count
			if tc.setStatuses {
				creds["pool_mode_retry_status_codes"] = tc.statuses
			}
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				payload := strings.TrimSuffix(tc.payload, "}")
				return retryHandlerResponse("data: " + payload + `,"type":"response.failed"}` + "\n\n")
			}
			f.run(t, context.Background())
			require.Len(t, f.up.ids, tc.want)
		})
	}
}

func TestOpenAIPreOutputRetryCancelBackoff(t *testing.T) {
	f := newRetryHandlerFixture(t, 2, 10, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f.up.respond = func(*http.Request, int64, int) *http.Response {
		time.AfterFunc(20*time.Millisecond, cancel)
		return retryHandlerResponse(retryHandlerFailure)
	}
	f.run(t, ctx)
	require.Equal(t, []int64{1}, f.up.ids)
	require.Equal(t, 1, f.slots.released)
}

func TestOpenAIPreOutputRetrySlotFailures(t *testing.T) {
	for _, mode := range []string{"acquire_error", "wait_error", "cancel_wait"} {
		t.Run(mode, func(t *testing.T) {
			f := newRetryHandlerFixture(t, 1, 0, false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				f.slots.acquire = func(context.Context, int64) (bool, error) {
					if mode == "acquire_error" {
						return false, errors.New("test acquire failure")
					}
					if mode == "wait_error" && f.slots.waitAdded > 0 {
						return false, errors.New("test wait failure")
					}
					return false, nil
				}
				if mode == "cancel_wait" {
					f.slots.onWait = func() { cancel() }
				}
				return retryHandlerResponse(retryHandlerFailure)
			}
			f.run(t, ctx)
			require.Equal(t, []int64{1}, f.up.ids)
			if mode != "acquire_error" {
				require.Equal(t, 1, f.slots.waitAdded)
			}
		})
	}
}

func TestOpenAIPreOutputRetryOriginalPinPrivateFirst(t *testing.T) {
	for _, mode := range []string{"original_pin", "private_first", "lost_ownership", "wrong_group"} {
		t.Run(mode, func(t *testing.T) {
			f := newRetryHandlerFixture(t, 2, 10, false)
			ctx := context.Background()
			if mode == "original_pin" {
				ctx = context.WithValue(ctx, ctxkey.APIKeyPinnedAccountID, int64(1))
			}
			if mode == "private_first" || mode == "lost_ownership" {
				ctx = context.WithValue(ctx, ctxkey.APIKeyAccountPoolStrategy, service.AccountPoolStrategyPrivateFirst)
				owner := int64(7)
				f.repo.accounts[0].OwnerUserID = &owner
			}
			f.up.respond = func(_ *http.Request, id int64, n int) *http.Response {
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
					return retryHandlerResponse(retryHandlerFailure)
				}
				return retryHandlerResponse(retryHandlerSuccess)
			}
			rec := f.run(t, ctx)
			switch mode {
			case "original_pin":
				require.Equal(t, []int64{1}, f.up.ids)
			case "private_first":
				require.Equal(t, []int64{1, 1}, f.up.ids)
			default:
				require.Equal(t, []int64{1, 2}, f.up.ids, rec.Body.String())
			}
		})
	}
}

func TestOpenAIPreOutputRetryOverflowNoDispatch(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			f := newRetryHandlerFixture(t, 2, 10, passthrough)
			f.cfg.Gateway.MaxLineSize = 1024
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				return retryHandlerResponse(strings.Repeat("data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 30) + retryHandlerFailure)
			}
			rec := f.run(t, context.Background())
			require.Equal(t, []int64{1}, f.up.ids)
			require.NotContains(t, rec.Body.String(), "resp_failed")
		})
	}
}

func TestOpenAIPreOutputRetryLocalHTTP(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if calls.Add(1) == 1 {
			_, _ = io.WriteString(w, retryHandlerFailure)
		} else {
			_, _ = io.WriteString(w, retryHandlerSuccess)
		}
	}))
	defer upstream.Close()
	f := newRetryHandlerFixture(t, 2, 10, false)
	for i := range f.repo.accounts {
		f.repo.accounts[i].Credentials["base_url"] = upstream.URL
	}
	f.up.respond = func(req *http.Request, _ int64, _ int) *http.Response {
		resp, err := upstream.Client().Do(req)
		require.NoError(t, err)
		return resp
	}
	rec := f.run(t, context.Background())
	require.Equal(t, []int64{1, 1}, f.up.ids)
	require.Equal(t, int32(2), calls.Load())
	require.Contains(t, rec.Body.String(), "resp_success")
}

type retryHandlerReadError struct{}

func (retryHandlerReadError) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestOpenAIPreOutputRetryDisconnectDispatch(t *testing.T) {
	for _, mode := range []string{"missing_terminal", "read_error"} {
		for _, passthrough := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%v", mode, passthrough), func(t *testing.T) {
				f := newRetryHandlerFixture(t, 1, 0, passthrough)
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					resp := retryHandlerResponse("data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_failed\"}}\n\n")
					if mode == "read_error" {
						resp.Body = io.NopCloser(io.MultiReader(resp.Body, retryHandlerReadError{}))
					}
					return resp
				}
				rec := f.run(t, context.Background())
				require.Equal(t, []int64{1, 1, 1, 1}, f.up.ids)
				require.NotContains(t, rec.Body.String(), "resp_failed")
			})
		}
	}
}

func TestOpenAIPreOutputRetryNeverAfterOutput(t *testing.T) {
	for _, event := range []string{
		`{"type":"response.output_text.delta","delta":"text"}`,
		`{"type":"response.reasoning_summary_text.delta","delta":"reasoning"}`,
		`{"type":"response.function_call_arguments.delta","delta":"{}"}`,
		`{"type":"response.output_item.done","item":{"type":"image_generation_call","result":"image"}}`,
		`{"type":"response.completed","response":{"id":"resp_success","output":[]}}`,
	} {
		for _, passthrough := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/%s", passthrough, event), func(t *testing.T) {
				f := newRetryHandlerFixture(t, 2, 10, passthrough)
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					resp := retryHandlerResponse("data: " + event + "\n\n")
					resp.Body = io.NopCloser(io.MultiReader(resp.Body, retryHandlerReadError{}))
					return resp
				}
				f.run(t, context.Background())
				require.Equal(t, []int64{1}, f.up.ids)
			})
		}
	}
}

type retryHandlerCancelEOF struct{ cancel context.CancelFunc }

func (r retryHandlerCancelEOF) Read([]byte) (int, error) { r.cancel(); return 0, io.EOF }

func TestOpenAIPreOutputRetryCancelEOF(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			f := newRetryHandlerFixture(t, 2, 10, passthrough)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				resp := retryHandlerResponse("")
				resp.Body = io.NopCloser(retryHandlerCancelEOF{cancel})
				return resp
			}
			rec := f.run(t, ctx)
			require.Equal(t, []int64{1}, f.up.ids)
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestOpenAIPreOutputRetrySpecialPolicies(t *testing.T) {
	for _, tc := range []struct {
		name, code, message string
		want                int
	}{
		{name: "capacity", message: "Selected model is at capacity. Please try a different model.", want: 6},
		{name: "overloaded", code: "server_is_overloaded", want: 4},
		{name: "auth_existing_failover", code: "invalid_api_key", want: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newRetryHandlerFixture(t, 1, 0, false)
			if tc.name == "auth_existing_failover" {
				f = newRetryHandlerFixture(t, 2, 10, false)
			}
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				return retryHandlerResponse(fmt.Sprintf("data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":%q,\"message\":%q}}}\n\n", tc.code, tc.message))
			}
			f.run(t, context.Background())
			require.Len(t, f.up.ids, tc.want)
			if tc.name == "auth_existing_failover" {
				require.Equal(t, []int64{1, 2}, f.up.ids)
			}
			if tc.name == "overloaded" {
				for i, delay := range []time.Duration{time.Second, 2 * time.Second, 3 * time.Second} {
					require.GreaterOrEqual(t, f.up.times[i+1].Sub(f.up.times[i]), delay-20*time.Millisecond)
				}
			}
		})
	}
}

func TestOpenAIPreOutputRetryTickerKeepalive(t *testing.T) {
	f := newRetryHandlerFixture(t, 2, 10, false)
	f.cfg.Gateway.StreamKeepaliveInterval = 1
	f.up.respond = func(_ *http.Request, _ int64, n int) *http.Response {
		if n > 1 {
			return retryHandlerResponse(retryHandlerSuccess)
		}
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			_, _ = io.WriteString(pw, strings.Repeat("data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_failed\"}}\n\n", 100))
			time.Sleep(1150 * time.Millisecond)
			_, _ = io.WriteString(pw, retryHandlerFailure)
		}()
		resp := retryHandlerResponse("")
		resp.Body = pr
		return resp
	}
	rec := f.run(t, context.Background())
	require.Equal(t, []int64{1, 1}, f.up.ids)
	require.Contains(t, rec.Body.String(), ":\n\n")
	require.Contains(t, rec.Body.String(), "resp_success")
	require.NotContains(t, rec.Body.String(), "resp_failed")
	require.NotContains(t, rec.Body.String(), "response.failed")
}

func TestOpenAIPreOutputRetrySingleAccountNoAlternative(t *testing.T) {
	f := newRetryHandlerFixture(t, 1, 10, false)
	f.run(t, context.Background())
	require.Equal(t, []int64{1, 1, 1, 1}, f.up.ids)
}

func TestOpenAIPreOutputRetryOversizedLine(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			f := newRetryHandlerFixture(t, 2, 10, passthrough)
			f.up.respond = func(*http.Request, int64, int) *http.Response {
				return retryHandlerResponse("data: " + strings.Repeat("x", 128*1024) + "\n\n")
			}
			rec := f.run(t, context.Background())
			require.Equal(t, []int64{1}, f.up.ids)
			require.NotContains(t, rec.Body.String(), strings.Repeat("x", 128))
		})
	}
}

type retryHandlerCanceledRead struct{}

func (retryHandlerCanceledRead) Read([]byte) (int, error) { return 0, context.Canceled }

func TestOpenAIPreOutputRetryCancellationRecordsExistingPartialUsage(t *testing.T) {
	f := newRetryHandlerFixture(t, 2, 10, false)
	// Exercise the existing raw-chat adapter, which already returns partial
	// result+error. Do not change Responses wrappers' baseline nil,error policy.
	f.repo.accounts[0].Extra["openai_responses_supported"] = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f.onWrite = func(p []byte) {
		if strings.Contains(string(p), "paid-output") {
			cancel()
		}
	}
	f.up.respond = func(*http.Request, int64, int) *http.Response {
		body := "data: {\"id\":\"chat_usage\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"paid-output\"}}]}\n\n" +
			"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":13,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":3}}}\n\n"
		resp := retryHandlerResponse(body)
		resp.Body = io.NopCloser(io.MultiReader(resp.Body, retryHandlerCanceledRead{}))
		return resp
	}
	rec := f.run(t, ctx)
	require.ErrorIs(t, ctx.Err(), context.Canceled)
	require.Equal(t, []int64{1}, f.up.ids)
	require.NotContains(t, rec.Body.String(), "response.failed")
	select {
	case usage := <-f.usageRepo.created:
		require.Equal(t, 13, usage.InputTokens+usage.CacheReadTokens)
		require.Equal(t, 7, usage.OutputTokens)
		require.Equal(t, 3, usage.CacheReadTokens)
	default:
		t.Fatal("canceled partial usage was not submitted to RecordUsage")
	}
	require.Empty(t, f.usageRepo.created, "partial usage must be recorded exactly once")
}

func TestOpenAIPreOutputRetryCancellationAfterOutputRecordsTerminalUsage(t *testing.T) {
	for _, mode := range []string{"sync", "ticker", "passthrough"} {
		for _, image := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/image=%v", mode, image), func(t *testing.T) {
				f := newRetryHandlerFixture(t, 2, 10, mode == "passthrough")
				if mode == "ticker" {
					f.cfg.Gateway.StreamKeepaliveInterval = 1
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				written := make(chan struct{})
				var once sync.Once
				f.onWrite = func(p []byte) {
					if strings.Contains(string(p), "paid-output") {
						once.Do(func() { cancel(); close(written) })
					}
				}
				f.up.respond = func(*http.Request, int64, int) *http.Response {
					first := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"paid-output\"}\n\n"
					if image {
						first = "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"image_generation_call\",\"id\":\"img_1\",\"result\":\"paid-output\"}}\n\n"
					}
					resp := retryHandlerResponse("")
					resp.Body = io.NopCloser(&retryHandlerReadableAfterCancel{
						first: first, written: written,
						terminal: "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_usage\",\"output\":[],\"usage\":{\"input_tokens\":13,\"output_tokens\":7,\"input_tokens_details\":{\"cached_tokens\":3}}}}\n\n",
					})
					return resp
				}
				f.run(t, ctx)
				require.ErrorIs(t, ctx.Err(), context.Canceled)
				require.Equal(t, []int64{1}, f.up.ids)
				select {
				case usage := <-f.usageRepo.created:
					require.Equal(t, 13, usage.InputTokens+usage.CacheReadTokens)
					require.Equal(t, 7, usage.OutputTokens)
					require.Equal(t, 3, usage.CacheReadTokens)
					if image {
						require.Equal(t, 1, usage.ImageCount)
					}
				default:
					t.Fatal("terminal usage after output cancellation was not recorded")
				}
				require.Empty(t, f.usageRepo.created, "terminal usage must be recorded exactly once")
			})
		}
	}
}

type retryHandlerReadableAfterCancel struct {
	first, terminal string
	written         <-chan struct{}
}

func (r *retryHandlerReadableAfterCancel) Read(p []byte) (int, error) {
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
