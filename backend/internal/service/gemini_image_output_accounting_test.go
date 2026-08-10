package service

import (
	"bytes"
	"context"
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

const geminiTestPNG = "iVBORw0KGgoAAAANSUhEUg=="

func newGeminiImageTestContext(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost,
		"/v1beta/models/nana-banana-2:generateContent", strings.NewReader("{}"))
	return c
}

func geminiImageResponse(parts string) string {
	return `{"candidates":[{"content":{"role":"model","parts":[` + parts + `]},"finishReason":"STOP"}]}`
}

func TestCountGeminiInlineImageOutputs(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    int
	}{
		{
			name:    "camelCase inlineData",
			payload: geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`),
			want:    1,
		},
		{
			name:    "snake_case inline_data",
			payload: geminiImageResponse(`{"inline_data":{"mime_type":"image/png","data":"` + geminiTestPNG + `"}}`),
			want:    1,
		},
		{
			name: "multiple images",
			payload: geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}},` +
				`{"inlineData":{"mimeType":"image/webp","data":"` + geminiTestPNG + `"}}`),
			want: 2,
		},
		{
			name:    "uppercase mime type",
			payload: geminiImageResponse(`{"inlineData":{"mimeType":"IMAGE/PNG","data":"` + geminiTestPNG + `"}}`),
			want:    1,
		},
		{
			name:    "non image mime type",
			payload: geminiImageResponse(`{"inlineData":{"mimeType":"audio/mpeg","data":"` + geminiTestPNG + `"}}`),
			want:    0,
		},
		{
			name:    "empty data",
			payload: geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":""}}`),
			want:    0,
		},
		{name: "text only", payload: geminiImageResponse(`{"text":"no image"}`), want: 0},
		{name: "invalid json", payload: "not-json", want: 0},
		{name: "error response", payload: `{"error":{"code":429}}`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, countGeminiInlineImageOutputs([]byte(tt.payload)))
		})
	}
}

func TestObserveGeminiImageOutputs_CumulativeChunksDoNotDoubleCount(t *testing.T) {
	c := newGeminiImageTestContext(t)
	beginGeminiImageOutputObservation(c)
	oneImage := geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`)
	for range 4 {
		observeGeminiImageOutputs(c, []byte(oneImage))
	}
	require.Equal(t, 1, observedGeminiImageOutputs(c))
}

func TestObserveGeminiImageOutputs_KeepsLargestChunk(t *testing.T) {
	c := newGeminiImageTestContext(t)
	beginGeminiImageOutputObservation(c)
	observeGeminiImageOutputs(c, []byte(geminiImageResponse(`{"text":"working"}`)))
	observeGeminiImageOutputs(c, []byte(geminiImageResponse(
		`{"inlineData":{"mimeType":"image/png","data":"`+geminiTestPNG+`"}},`+
			`{"inlineData":{"mimeType":"image/png","data":"`+geminiTestPNG+`"}}`)))
	observeGeminiImageOutputs(c, []byte(`{"usageMetadata":{"promptTokenCount":9}}`))
	require.Equal(t, 2, observedGeminiImageOutputs(c))
}

func TestBeginGeminiImageOutputObservation_ResetsPerForward(t *testing.T) {
	c := newGeminiImageTestContext(t)
	oneImage := []byte(geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`))
	beginGeminiImageOutputObservation(c)
	observeGeminiImageOutputs(c, oneImage)
	require.Equal(t, 1, observedGeminiImageOutputs(c))
	beginGeminiImageOutputObservation(c)
	require.Equal(t, 0, observedGeminiImageOutputs(c))
}

func TestResolveGeminiImageCount(t *testing.T) {
	oneImage := []byte(geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`))

	t.Run("custom model bills by observed images", func(t *testing.T) {
		c := newGeminiImageTestContext(t)
		beginGeminiImageOutputObservation(c)
		observeGeminiImageOutputs(c, oneImage)
		require.False(t, isImageGenerationModel("nana-banana-2"))
		require.Equal(t, 1, resolveGeminiImageCount(c, "nana-banana-2", "nana-banana-2"))
	})

	t.Run("falls back to mapped model", func(t *testing.T) {
		c := newGeminiImageTestContext(t)
		beginGeminiImageOutputObservation(c)
		require.Equal(t, 1, resolveGeminiImageCount(c, "my-image-alias", "gemini-2.5-flash-image"))
	})

	t.Run("text model stays unbilled", func(t *testing.T) {
		c := newGeminiImageTestContext(t)
		beginGeminiImageOutputObservation(c)
		require.Equal(t, 0, resolveGeminiImageCount(c, "gemini-2.5-pro", "gemini-2.5-pro"))
	})
}

func TestHandleNativeNonStreamingResponse_FeedsImageCounter(t *testing.T) {
	c := newGeminiImageTestContext(t)
	beginGeminiImageOutputObservation(c)
	body := geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	svc := &GeminiMessagesCompatService{}
	usage, err := svc.handleNativeNonStreamingResponse(c, resp, false)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, 1, observedGeminiImageOutputs(c))
}

func TestGeminiMessagesCompatServiceForward_BillsCustomModelByObservedImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"custom-image-alias","max_tokens":16,"messages":[{"role":"user","content":"draw"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))

	responseBody := geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}}`)
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}}
	svc := &GeminiMessagesCompatService{httpUpstream: upstream, cfg: &config.Config{}}
	account := &Account{
		ID: 701, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "test-key",
			"model_mapping": map[string]any{
				"custom-image-alias": "custom-upstream-image-alias",
			},
		},
	}

	result, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, "custom-upstream-image-alias", result.UpstreamModel)
}

func TestGeminiMessagesCompatServiceForwardNative_BillsActualImageCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"contents":[{"role":"user","parts":[{"text":"draw two"}]}]}`)
	c.Request = httptest.NewRequest(http.MethodPost,
		"/v1beta/models/custom-image-alias:generateContent", bytes.NewReader(body))

	responseBody := geminiImageResponse(`{"inlineData":{"mimeType":"image/png","data":"` + geminiTestPNG + `"}},` +
		`{"inlineData":{"mimeType":"image/webp","data":"` + geminiTestPNG + `"}}`)
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}}
	svc := &GeminiMessagesCompatService{httpUpstream: upstream, cfg: &config.Config{}}
	account := &Account{
		ID: 702, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "test-key",
			"model_mapping": map[string]any{
				"custom-image-alias": "custom-upstream-image-alias",
			},
		},
	}

	result, err := svc.ForwardNative(context.Background(), c, account,
		"custom-image-alias", "generateContent", false, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.ImageCount)
	require.Equal(t, "custom-upstream-image-alias", result.UpstreamModel)
}

type s207RateLimitAccountRepoStub struct {
	AccountRepository
	rateLimitCalls     int
	lastRateLimitID    int64
	lastRateLimitReset time.Time
}

func (r *s207RateLimitAccountRepoStub) SetRateLimited(_ context.Context, id int64, resetAt time.Time) error {
	r.rateLimitCalls++
	r.lastRateLimitID = id
	r.lastRateLimitReset = resetAt
	return nil
}

func TestHandleGeminiUpstreamError_PoolMode429(t *testing.T) {
	body := []byte(`{"error":{"code":429,"message":"You have exhausted your capacity on this model."}}`)
	tests := []struct {
		name              string
		account           *Account
		expectRateLimited bool
	}{
		{
			name: "pool mode API key stays schedulable",
			account: &Account{ID: 600, Platform: PlatformGemini, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"pool_mode": true}},
		},
		{
			name: "custom error code match overrides pool mode",
			account: &Account{ID: 601, Platform: PlatformGemini, Type: AccountTypeAPIKey,
				Credentials: map[string]any{
					"pool_mode": true, "custom_error_codes_enabled": true,
					"custom_error_codes": []any{float64(429)},
				}},
			expectRateLimited: true,
		},
		{
			name: "custom error code miss skips",
			account: &Account{ID: 602, Platform: PlatformGemini, Type: AccountTypeAPIKey,
				Credentials: map[string]any{
					"pool_mode": true, "custom_error_codes_enabled": true,
					"custom_error_codes": []any{float64(500)},
				}},
		},
		{
			name:              "non pool API key remains rate limited",
			account:           &Account{ID: 603, Platform: PlatformGemini, Type: AccountTypeAPIKey},
			expectRateLimited: true,
		},
		{
			name: "OAuth ignores pool mode flag",
			account: &Account{ID: 604, Platform: PlatformGemini, Type: AccountTypeOAuth,
				Credentials: map[string]any{"pool_mode": true}},
			expectRateLimited: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &s207RateLimitAccountRepoStub{}
			svc := &GeminiMessagesCompatService{
				accountRepo:      repo,
				rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil),
			}
			svc.handleGeminiUpstreamError(context.Background(), tt.account, http.StatusTooManyRequests, http.Header{}, body)
			if tt.expectRateLimited {
				require.Equal(t, 1, repo.rateLimitCalls)
				require.Equal(t, tt.account.ID, repo.lastRateLimitID)
				return
			}
			require.Zero(t, repo.rateLimitCalls)
		})
	}
}
