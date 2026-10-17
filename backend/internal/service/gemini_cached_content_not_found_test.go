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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const geminiCachedContentNotFoundBody = `{"error":{"code":403,"message":"CachedContent not found (or permission denied)","status":"PERMISSION_DENIED"}}`
const geminiCachedContentRequestBody = `{"cachedContent":"cachedContents/missing","contents":[{"role":"user","parts":[{"text":"hi"}]}]}`

// Any unexpected repository method panics through the nil embedded interface.
// Track all error handling writes explicitly, including cooldowns.
type geminiCacheAccountRepo struct {
	AccountRepository
	errors, temp, limited, overloaded, updates int
}

func (r *geminiCacheAccountRepo) SetError(context.Context, int64, string) error {
	r.errors++
	return nil
}
func (r *geminiCacheAccountRepo) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	r.temp++
	return nil
}
func (r *geminiCacheAccountRepo) SetRateLimited(context.Context, int64, time.Time) error {
	r.limited++
	return nil
}
func (r *geminiCacheAccountRepo) SetOverloaded(context.Context, int64, time.Time) error {
	r.overloaded++
	return nil
}
func (r *geminiCacheAccountRepo) Update(context.Context, *Account) error {
	r.updates++
	return nil
}

func newGeminiCacheTestService(status int, body string) (*GeminiMessagesCompatService, *geminiCompatHTTPUpstreamStub, *geminiCacheAccountRepo) {
	repo := &geminiCacheAccountRepo{}
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: status,
		Header: http.Header{
			"Content-Type":      []string{"application/json; charset=utf-8"},
			"X-Goog-Request-Id": []string{"cache-request-1"},
			"Retry-After":       []string{"17"},
			"Cache-Control":     []string{"no-store"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}}
	cfg := &config.Config{}
	return &GeminiMessagesCompatService{
		httpUpstream: upstream, cfg: cfg, accountRepo: repo,
		tokenProvider:    &GeminiTokenProvider{},
		rateLimitService: NewRateLimitService(repo, nil, cfg, nil, nil),
	}, upstream, repo
}

func geminiCacheTestAccount() *Account {
	return &Account{ID: 710, Platform: PlatformGemini, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-key"}}
}

func TestGeminiForwardNative_CachedContentNotFoundRequestError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, auth := range []string{"api_key", "oauth_direct", "oauth_code_assist"} {
		for _, action := range []string{"generateContent", "streamGenerateContent", "countTokens"} {
			for _, policy := range []string{"default", "custom_hit", "custom_miss", "temp", "combined", "nil_service"} {
				t.Run(auth+"/"+action+"/"+policy, func(t *testing.T) {
					account := geminiCacheTestAccount()
					body := geminiCachedContentNotFoundBody
					if auth != "api_key" {
						account.Type = AccountTypeOAuth
						account.Credentials["access_token"] = "test-oauth-token"
					}
					if auth == "oauth_code_assist" {
						account.Credentials["project_id"] = "test-project"
						body = `{"response":` + body + `}`
					}
					if policy == "custom_hit" || policy == "combined" || policy == "custom_miss" {
						account.Credentials["custom_error_codes_enabled"] = true
						code := float64(403)
						if policy == "custom_miss" {
							code = 429
						}
						account.Credentials["custom_error_codes"] = []any{code}
					}
					if policy == "temp" || policy == "combined" {
						account.Credentials["temp_unschedulable_enabled"] = true
						account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
							"error_code": float64(403), "keywords": []any{"permission denied"},
							"duration_minutes": float64(10),
						}}
					}
					svc, upstream, repo := newGeminiCacheTestService(http.StatusForbidden, body)
					if policy == "nil_service" {
						svc.rateLimitService = nil
					}
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-test:"+action, nil)
					result, err := svc.ForwardNative(context.Background(), c, account, "gemini-test", action,
						action == "streamGenerateContent", []byte(geminiCachedContentRequestBody))
					require.Nil(t, result)
					require.Error(t, err)
					var failover *UpstreamFailoverError
					require.False(t, errors.As(err, &failover))
					require.Equal(t, 1, upstream.calls)
					require.Equal(t, http.StatusForbidden, rec.Code)
					require.Equal(t, geminiCachedContentNotFoundBody, rec.Body.String(), "only OAuth response wrapper is normalized")
					require.Equal(t, []string{"cache-request-1"}, rec.Header().Values("X-Request-Id"))
					require.Equal(t, "17", rec.Header().Get("Retry-After"))
					require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
					require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
					require.Equal(t, &geminiCacheAccountRepo{}, repo, "zero account writes or cooldowns")
					require.NotNil(t, upstream.lastReq)
					sent, readErr := io.ReadAll(upstream.lastReq.Body)
					require.NoError(t, readErr)
					require.Contains(t, string(sent), `"cachedContent":"cachedContents/missing"`)
				})
			}
		}
	}
}

func TestGeminiCachedContentNotFoundRecognition(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"direct", 403, geminiCachedContentNotFoundBody, true},
		{"oauth_wrapper", 403, `{"response":` + geminiCachedContentNotFoundBody + `}`, true},
		{"singleton_array", 403, `[` + geminiCachedContentNotFoundBody + `]`, true},
		{"wrong_status", 400, geminiCachedContentNotFoundBody, false},
		{"not_found_status", 404, geminiCachedContentNotFoundBody, false},
		{"credential", 403, `{"error":{"code":403,"status":"PERMISSION_DENIED","message":"Permission denied: Consumer has been suspended."}}`, false},
		{"mention_only", 403, `{"error":{"code":403,"status":"PERMISSION_DENIED","message":"Invalid credentials; previous error CachedContent not found (or permission denied)"}}`, false},
		{"wrong_code", 403, `{"error":{"code":401,"status":"PERMISSION_DENIED","message":"CachedContent not found (or permission denied)"}}`, false},
		{"wrong_error_status", 403, `{"error":{"code":403,"status":"UNAUTHENTICATED","message":"CachedContent not found (or permission denied)"}}`, false},
		{"metadata_only", 403, `{"metadata":{"message":"CachedContent not found (or permission denied)"}}`, false},
		{"mixed_array", 403, `[` + geminiCachedContentNotFoundBody + `,{"error":{"message":"invalid credentials"}}]`, false},
		{"malformed", 403, geminiCachedContentNotFoundBody + `!`, false},
		{"plain_text", 403, `CachedContent not found (or permission denied)`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isGeminiCachedContentNotFound(tc.status, []byte(tc.body)))
		})
	}
}

type geminiCacheProbeReader struct {
	io.Reader
	bytes  int
	closed bool
}

func (r *geminiCacheProbeReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytes += n
	return n, err
}
func (r *geminiCacheProbeReader) Close() error { r.closed = true; return nil }

type geminiCacheReadError struct{}

func (geminiCacheReadError) Read([]byte) (int, error) { return 0, errors.New("fixture read failure") }

func TestGeminiCachedContentNotFoundBoundedProbe(t *testing.T) {
	svc := &GeminiMessagesCompatService{}
	for _, tc := range []struct {
		name   string
		reader io.Reader
		status int
		want   bool
	}{
		{"small_valid", strings.NewReader(geminiCachedContentNotFoundBody), 403, true},
		{"large_valid_json", strings.NewReader(geminiCachedContentNotFoundBody + strings.Repeat(" ", 3<<20)), 403, false},
		{"large_truncated_json", strings.NewReader(`{"padding":"` + strings.Repeat("x", 3<<20) + `",` + geminiCachedContentNotFoundBody[1:]), 403, false},
		{"read_failure_after_valid_json", io.MultiReader(strings.NewReader(geminiCachedContentNotFoundBody), geminiCacheReadError{}), 403, false},
		{"read_failure", geminiCacheReadError{}, 403, false},
		{"non_403_unread", strings.NewReader(geminiCachedContentNotFoundBody), 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &geminiCacheProbeReader{Reader: tc.reader}
			resp := &http.Response{StatusCode: tc.status, Header: http.Header{"X-Request-Id": []string{"probe"}}, Body: reader}
			found, rebuilt := svc.cachedContentNotFoundInLoop(resp)
			require.Equal(t, tc.want, found)
			require.LessOrEqual(t, reader.bytes, 2<<20)
			require.Equal(t, resp.Header, rebuilt.Header)
			if tc.status != 403 {
				require.Same(t, resp, rebuilt)
				require.Zero(t, reader.bytes)
				require.False(t, reader.closed)
				return
			}
			require.True(t, reader.closed)
			restored, err := io.ReadAll(rebuilt.Body)
			require.NoError(t, err)
			require.Len(t, restored, reader.bytes)
			if tc.want {
				require.Equal(t, geminiCachedContentNotFoundBody, string(restored))
			}
			require.NoError(t, rebuilt.Body.Close())
		})
	}
}

func TestGeminiForwardNative_OtherErrorPolicyUnchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body, policy               string
		status, errors, temp, overloaded int
	}{
		{"credential_default", `{"error":{"code":403,"message":"Consumer has been suspended.","status":"PERMISSION_DENIED"}}`, "", 403, 1, 0, 0},
		{"cache_mention_credential", `{"error":{"code":403,"message":"Invalid credentials; CachedContent not found (or permission denied)","status":"PERMISSION_DENIED"}}`, "", 403, 1, 0, 0},
		{"credential_custom_miss", `{"error":{"code":403,"message":"Invalid credentials","status":"PERMISSION_DENIED"}}`, "custom_miss", 403, 0, 0, 0},
		{"credential_temp", `{"error":{"code":403,"message":"permission denied","status":"PERMISSION_DENIED"}}`, "temp", 403, 0, 1, 0},
		{"overload_custom", `{"error":{"code":529,"message":"overloaded","status":"UNAVAILABLE"}}`, "custom_hit", 529, 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, upstream, repo := newGeminiCacheTestService(tc.status, tc.body)
			account := geminiCacheTestAccount()
			if strings.HasPrefix(tc.policy, "custom") {
				code := tc.status
				if tc.policy == "custom_miss" {
					code = 429
				}
				account.Credentials["custom_error_codes_enabled"] = true
				account.Credentials["custom_error_codes"] = []any{float64(code)}
			}
			if tc.policy == "temp" {
				account.Credentials["temp_unschedulable_enabled"] = true
				account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
					"error_code": float64(403), "keywords": []any{"permission denied"}, "duration_minutes": float64(10),
				}}
			}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			_, err := svc.ForwardNative(context.Background(), c, account, "gemini-test", "generateContent", false, []byte(geminiCachedContentRequestBody))
			var failover *UpstreamFailoverError
			require.ErrorAs(t, err, &failover)
			require.Equal(t, tc.status, failover.StatusCode)
			require.Equal(t, tc.body, string(failover.ResponseBody))
			require.Equal(t, tc.errors, repo.errors)
			require.Equal(t, tc.temp, repo.temp)
			require.Equal(t, tc.overloaded, repo.overloaded)
			require.Zero(t, repo.limited)
			require.Zero(t, repo.updates)
			require.Equal(t, 1, upstream.calls)
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestGeminiForwardNative_CachedContentProbeFailureKeepsErrorPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		reader io.Reader
	}{
		{"large_valid_json", strings.NewReader(geminiCachedContentNotFoundBody + strings.Repeat(" ", 3<<20))},
		{"read_failure_after_valid_json", io.MultiReader(strings.NewReader(geminiCachedContentNotFoundBody), geminiCacheReadError{})},
		{"read_failure", geminiCacheReadError{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, upstream, repo := newGeminiCacheTestService(403, "")
			reader := &geminiCacheProbeReader{Reader: tc.reader}
			upstream.response.Body = reader
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			_, err := svc.ForwardNative(context.Background(), c, geminiCacheTestAccount(), "gemini-test",
				"generateContent", false, []byte(geminiCachedContentRequestBody))
			var failover *UpstreamFailoverError
			require.ErrorAs(t, err, &failover, "unconfirmed cache error must follow the original account error policy")
			require.Equal(t, 403, failover.StatusCode)
			require.Equal(t, 1, repo.errors)
			require.Zero(t, repo.temp)
			require.Zero(t, repo.limited)
			require.Zero(t, repo.overloaded)
			require.Zero(t, repo.updates)
			require.Equal(t, 1, upstream.calls)
			require.LessOrEqual(t, reader.bytes, 2<<20)
			require.True(t, reader.closed)
			require.Empty(t, rec.Body.String(), "must not fall through to special request-error passthrough")
		})
	}
}
