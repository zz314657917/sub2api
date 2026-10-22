package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Count repository reads as well as writes: a second None evaluation has no
// extra state write but still reads the stored 401 pause again.
type geminiSingleCheckRepo struct {
	geminiCacheAccountRepo
	reads      int
	tempReason string
}

func (r *geminiSingleCheckRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	r.reads++
	return &Account{ID: id, TempUnschedulableReason: r.tempReason}, nil
}

func (r *geminiSingleCheckRepo) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, reason string) error {
	r.temp++
	r.tempReason = reason
	return nil
}

type geminiSingleCheckResponse struct {
	status int
	body   string
}

type geminiSingleCheckUpstream struct {
	responses []geminiSingleCheckResponse
	requests  []string
}

func (s *geminiSingleCheckUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	response := s.responses[min(len(s.requests), len(s.responses)-1)]
	s.requests = append(s.requests, string(body))
	contentType := "application/json"
	if strings.HasPrefix(response.body, "data: ") {
		contentType = "text/event-stream"
	}
	return &http.Response{
		StatusCode: response.status,
		Header: http.Header{
			"Content-Type":      []string{contentType},
			"X-Goog-Request-Id": []string{"single-check-request"},
			"Retry-After":       []string{"17"},
		},
		Body: io.NopCloser(strings.NewReader(response.body)),
	}, nil
}

func (s *geminiSingleCheckUpstream) DoWithTLS(req *http.Request, proxyURL string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxyURL, id, concurrency)
}

func newGeminiSingleCheckService(responses ...geminiSingleCheckResponse) (*GeminiMessagesCompatService, *geminiSingleCheckUpstream, *geminiSingleCheckRepo) {
	repo := &geminiSingleCheckRepo{}
	upstream := &geminiSingleCheckUpstream{responses: responses}
	cfg := &config.Config{}
	return &GeminiMessagesCompatService{
		httpUpstream: upstream, accountRepo: repo, cfg: cfg,
		rateLimitService: NewRateLimitService(repo, nil, cfg, nil, nil),
	}, upstream, repo
}

var geminiSingleCheckEntrypoints = []string{"messages", "native", "chat"}

// Exercise the public forwarding entries, not an error-policy helper.
func geminiSingleCheckForward(t *testing.T, entry string, svc *GeminiMessagesCompatService, account *Account, stream bool) (*ForwardResult, error, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	streamJSON := "false"
	if stream {
		streamJSON = "true"
	}
	body := `{"model":"gemini-2.5-flash","max_tokens":16,"stream":` + streamJSON + `,"stream_options":{"include_usage":true},"messages":[{"role":"user","content":"hi"}]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+entry, strings.NewReader(body))
	var result *ForwardResult
	var err error
	switch entry {
	case "messages":
		result, err = svc.Forward(context.Background(), c, account, []byte(body))
	case "native":
		action := "generateContent"
		if stream {
			action = "streamGenerateContent"
		}
		result, err = svc.ForwardNative(context.Background(), c, account, "gemini-2.5-flash", action, stream,
			[]byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
	case "chat":
		result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, []byte(body))
	default:
		t.Fatalf("unknown entry %q", entry)
	}
	return result, err, rec
}

func geminiSingleCheckAccount() *Account {
	account := geminiCacheTestAccount()
	account.Credentials["temp_unschedulable_enabled"] = true
	account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
		"error_code": float64(401), "keywords": []any{"expired"}, "duration_minutes": float64(30),
	}}
	return account
}

const geminiSingleCheck401 = `{"error":{"code":401,"message":"API key expired","status":"UNAUTHENTICATED"}}`

func TestGeminiV0217ErrorPolicyNoneCheckedOnceAtRealEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range geminiSingleCheckEntrypoints {
		t.Run(entry, func(t *testing.T) {
			body := `{"error":{"code":401,"message":"invalid credential","status":"UNAUTHENTICATED"}}`
			svc, upstream, repo := newGeminiSingleCheckService(geminiSingleCheckResponse{401, body})
			_, err, _ := geminiSingleCheckForward(t, entry, svc, geminiSingleCheckAccount(), false)
			var failover *UpstreamFailoverError
			require.ErrorAs(t, err, &failover)
			require.Equal(t, 401, failover.StatusCode)
			require.JSONEq(t, body, string(failover.ResponseBody))
			require.Len(t, upstream.requests, 1)
			require.Equal(t, 1, repo.reads, "None must not be evaluated again after the loop")
			require.Equal(t, 1, repo.errors)
			require.Zero(t, repo.temp)
		})
	}
}

func TestGeminiV0217ErrorPolicyFirstAndSubsequent401AtRealEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range geminiSingleCheckEntrypoints {
		t.Run(entry, func(t *testing.T) {
			svc, upstream, repo := newGeminiSingleCheckService(geminiSingleCheckResponse{401, geminiSingleCheck401})
			account := geminiSingleCheckAccount()
			for request := 1; request <= 2; request++ {
				_, err, _ := geminiSingleCheckForward(t, entry, svc, account, false)
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, 401, failover.StatusCode)
				require.JSONEq(t, geminiSingleCheck401, string(failover.ResponseBody))
				require.Len(t, upstream.requests, request)
				require.Equal(t, request, repo.reads, "one stored-pause read per real request")
				require.Equal(t, 1, repo.temp, "first request pauses only once")
				require.Equal(t, request-1, repo.errors, "only the later request escalates")
			}
		})
	}
}

func TestGeminiV0217ErrorPolicyMatchedAndSkippedAtRealEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range geminiSingleCheckEntrypoints {
		for _, policy := range []string{"matched", "custom_miss", "pool"} {
			t.Run(entry+"/"+policy, func(t *testing.T) {
				svc, upstream, repo := newGeminiSingleCheckService(geminiSingleCheckResponse{401, geminiSingleCheck401})
				account := geminiSingleCheckAccount()
				if policy == "pool" {
					account.Credentials["pool_mode"] = true
				} else {
					account.Credentials["custom_error_codes_enabled"] = true
					code := float64(401)
					if policy == "custom_miss" {
						code = 429
					}
					account.Credentials["custom_error_codes"] = []any{code}
				}
				_, err, _ := geminiSingleCheckForward(t, entry, svc, account, false)
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, 401, failover.StatusCode)
				require.JSONEq(t, geminiSingleCheck401, string(failover.ResponseBody))
				require.Len(t, upstream.requests, 1, "non-None policy stops retries")
				wantErrors := 0
				if policy == "matched" {
					wantErrors = 1
				}
				require.Equal(t, wantErrors, repo.errors)
				require.Zero(t, repo.temp)
				require.Zero(t, repo.reads, "custom-code and pool decisions precede temp rules")
			})
		}
	}
}

func TestGeminiV0217ErrorPolicyFreshResponseAndUnchecked400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range geminiSingleCheckEntrypoints {
		for _, retry := range []bool{false, true} {
			name := entry + "/direct"
			if retry {
				name = entry + "/after503"
			}
			t.Run(name, func(t *testing.T) {
				body := `{"error":{"code":400,"message":"request rejected marker","status":"INVALID_ARGUMENT"}}`
				responses := []geminiSingleCheckResponse{{400, body}}
				if retry {
					responses = append([]geminiSingleCheckResponse{{503, `{"error":{"message":"backend busy"}}`}}, responses...)
				}
				svc, upstream, repo := newGeminiSingleCheckService(responses...)
				account := geminiSingleCheckAccount()
				account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
					"error_code": float64(400), "keywords": []any{"marker"}, "duration_minutes": float64(30),
				}}
				_, err, _ := geminiSingleCheckForward(t, entry, svc, account, false)
				var failover *UpstreamFailoverError
				if entry == "chat" {
					// Chat keeps the existing mapped 400 response for temp rules.
					require.Error(t, err)
					require.False(t, errors.As(err, &failover))
				} else {
					require.ErrorAs(t, err, &failover)
					require.Equal(t, 400, failover.StatusCode)
					require.JSONEq(t, body, string(failover.ResponseBody))
				}
				require.Len(t, upstream.requests, len(responses))
				require.Equal(t, 1, repo.temp, "the final 400 must be checked once even after an earlier None")
				require.Zero(t, repo.errors)
			})
		}
	}
}

func TestGeminiV0217ErrorPolicyNativeCachedContentBypassesPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, upstream, repo := newGeminiSingleCheckService(geminiSingleCheckResponse{403, geminiCachedContentNotFoundBody})
	account := geminiSingleCheckAccount()
	account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
		"error_code": float64(403), "keywords": []any{"permission denied"}, "duration_minutes": float64(30),
	}}
	_, err, rec := geminiSingleCheckForward(t, "native", svc, account, false)
	require.Error(t, err)
	require.Equal(t, 403, rec.Code)
	require.Equal(t, geminiCachedContentNotFoundBody, rec.Body.String())
	require.Equal(t, "single-check-request", rec.Header().Get("X-Request-Id"))
	require.Equal(t, "17", rec.Header().Get("Retry-After"))
	require.Len(t, upstream.requests, 1)
	require.Equal(t, &geminiSingleCheckRepo{}, repo, "request-scoped cache miss must bypass account policy")
}

func TestGeminiV0217ErrorPolicyMessagesSignatureRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, upstream, repo := newGeminiSingleCheckService(
		geminiSingleCheckResponse{400, `{"error":{"message":"invalid thought_signature"}}`},
		geminiSingleCheckResponse{400, `{"error":{"message":"invalid signature"}}`},
		geminiSingleCheckResponse{200, `{"candidates":[{"content":{"role":"model","parts":[{"text":"recovered"}]},"finishReason":"STOP"}]}`},
	)
	result, err, rec := geminiSingleCheckForward(t, "messages", svc, geminiSingleCheckAccount(), false)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "recovered")
	require.Len(t, upstream.requests, 3, "both signature recovery stages remain available")
	require.Equal(t, &geminiSingleCheckRepo{}, repo)
}
