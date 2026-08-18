package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiSkippedWriteHTTPStub struct {
	response *http.Response
}

func (s *geminiSkippedWriteHTTPStub) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	resp := *s.response
	return &resp, nil
}

func (s *geminiSkippedWriteHTTPStub) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxyURL, accountID, accountConcurrency)
}

func newGeminiSkippedWriteService(status int, body string) *GeminiMessagesCompatService {
	return &GeminiMessagesCompatService{
		httpUpstream: &geminiSkippedWriteHTTPStub{response: &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}},
		cfg:              &config.Config{},
		rateLimitService: NewRateLimitService(nil, nil, &config.Config{}, nil, nil),
	}
}

func geminiSkippedPoolAccount() *Account {
	return &Account{
		ID:       700,
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":   "test-key",
			"pool_mode": true,
		},
	}
}

func geminiSkippedCustomAccount() *Account {
	return &Account{
		ID:       701,
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":                    "test-key",
			"custom_error_codes_enabled": true,
			"custom_error_codes":         []any{float64(http.StatusTooManyRequests)},
		},
	}
}

func geminiSkippedErrorBody() string {
	return `{"error":{"code":null,"message":"invalid Gemini function call history","param":"","type":"invalid_request_error"}}`
}

func newGeminiSkippedContext(method, path string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	return c, rec
}

func TestGeminiForwardNative_PoolModeSkipped400PassthroughRealStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstreamBody := geminiSkippedErrorBody()
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, upstreamBody)
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1beta/models/gemini-2.5-flash:generateContent", `{}`)

	result, err := svc.ForwardNative(context.Background(), c, geminiSkippedPoolAccount(), "gemini-2.5-flash", "generateContent", false, []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))

	require.Nil(t, result)
	require.Error(t, err)
	var failoverErr *UpstreamFailoverError
	require.False(t, errors.As(err, &failoverErr))
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, upstreamBody, rec.Body.String())
}

func TestGeminiForwardNative_SkippedFailoverStatusesSwitchAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newGeminiSkippedWriteService(http.StatusServiceUnavailable, `{"error":{"message":"temporarily unavailable"}}`)
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1beta/models/gemini-2.5-flash:generateContent", `{}`)

	result, err := svc.ForwardNative(context.Background(), c, geminiSkippedPoolAccount(), "gemini-2.5-flash", "generateContent", false, []byte(`{"contents":[]}`))

	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusServiceUnavailable, failoverErr.StatusCode)
	require.Zero(t, rec.Body.Len())
}

func TestGeminiForwardNative_CustomCodeMiss400HidesUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstreamBody := geminiSkippedErrorBody()
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, upstreamBody)
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1beta/models/gemini-2.5-flash:generateContent", `{}`)

	result, err := svc.ForwardNative(context.Background(), c, geminiSkippedCustomAccount(), "gemini-2.5-flash", "generateContent", false, []byte(`{"contents":[]}`))

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.NotContains(t, rec.Body.String(), "invalid Gemini function call history")
	require.Contains(t, rec.Body.String(), geminiCustomCodeSkippedClientMessage)
}

func TestGeminiForwardMessages_CustomCodeMiss400HidesUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstreamBody := geminiSkippedErrorBody()
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, upstreamBody)
	body := `{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1/messages", body)

	result, err := svc.Forward(context.Background(), c, geminiSkippedCustomAccount(), []byte(body))

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.NotContains(t, rec.Body.String(), "invalid Gemini function call history")
	require.Contains(t, rec.Body.String(), geminiCustomCodeSkippedClientMessage)
}

func TestGeminiForwardMessages_PoolMode400KeepsUpstreamMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, geminiSkippedErrorBody())
	body := `{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1/messages", body)

	result, err := svc.Forward(context.Background(), c, geminiSkippedPoolAccount(), []byte(body))

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "invalid Gemini function call history")
}

func TestGeminiForwardAsChatCompletions_CustomCodeMiss400HidesUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstreamBody := geminiSkippedErrorBody()
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, upstreamBody)
	body := `{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1/chat/completions", body)

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, geminiSkippedCustomAccount(), []byte(body))

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	errObj, ok := got["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "api_error", errObj["type"])
	require.Equal(t, geminiCustomCodeSkippedClientMessage, errObj["message"])
	require.NotContains(t, rec.Body.String(), "invalid Gemini function call history")
}

func TestGeminiForwardAsChatCompletions_PoolMode400KeepsUpstreamMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newGeminiSkippedWriteService(http.StatusBadRequest, geminiSkippedErrorBody())
	body := `{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"hi"}]}`
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1/chat/completions", body)

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, geminiSkippedPoolAccount(), []byte(body))

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	errObj, ok := got["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "invalid_request_error", errObj["type"])
	require.Equal(t, "invalid Gemini function call history", errObj["message"])
}

func TestWriteGeminiMappedError_400KeepsUpstreamMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &GeminiMessagesCompatService{cfg: &config.Config{}}
	c, rec := newGeminiSkippedContext(http.MethodPost, "/v1/messages", "")

	err := svc.writeGeminiMappedError(c, &Account{ID: 702, Platform: PlatformGemini}, http.StatusBadRequest, "req-1", []byte(geminiSkippedErrorBody()))

	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	errObj, ok := got["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "invalid Gemini function call history", errObj["message"])
}

func TestSkippedErrorPolicyFailoverError_CustomCodeMiss500HasNoSameAccountRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &GeminiMessagesCompatService{}
	c, _ := newGeminiSkippedContext(http.MethodPost, "/v1/messages", "")

	err := svc.skippedErrorPolicyFailoverError(c, geminiSkippedCustomAccount(), http.StatusInternalServerError, []byte(`{"error":"internal"}`), "req-1")

	require.NotNil(t, err)
	require.False(t, err.RetryableOnSameAccount)
}
