package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesGeminiImageRepresentations(t *testing.T) {
	for _, tc := range []struct {
		name, parts string
		want        int
	}{
		{"camel URI", `{"fileData":{"mimeType":"image/png","fileUri":"https://example.test/a.png"}}`, 1},
		{"snake URI", `{"file_data":{"mime_type":" IMAGE/WEBP ","file_uri":"gs://bucket/a.webp"}}`, 1},
		{"mixed spellings", `{"fileData":{"mime_type":"image/png","file_uri":"gs://a"}}`, 1},
		{"inline camel", `{"inlineData":{"mimeType":"image/png","data":"abc"}}`, 1},
		{"inline snake", `{"inline_data":{"mime_type":"image/png","data":"abc"}}`, 1},
		{"one per part", `{"inlineData":{"mimeType":"image/png","data":"abc"},"fileData":{"mimeType":"image/png","fileUri":"gs://a"}}`, 1},
		{"fallback representation", `{"inlineData":{"mimeType":"image/png","data":""},"fileData":{"mimeType":"image/png","fileUri":"gs://a"}}`, 1},
		{"non string inline number", `{"inlineData":{"mimeType":"image/png","data":123}}`, 0},
		{"non string inline object", `{"inline_data":{"mime_type":"image/png","data":{}}}`, 0},
		{"empty inline", `{"inlineData":{"mimeType":"image/png","data":"  "}}`, 0},
		{"non string URI", `{"fileData":{"mimeType":"image/png","fileUri":123}}`, 0},
		{"null URI", `{"fileData":{"mimeType":"image/png","fileUri":null}}`, 0},
		{"empty URI", `{"file_data":{"mime_type":"image/png","file_uri":"  "}}`, 0},
		{"missing URI", `{"fileData":{"mimeType":"image/png"}}`, 0},
		{"PDF", `{"fileData":{"mimeType":"application/pdf","fileUri":"gs://a"}}`, 0},
		{"missing MIME", `{"fileData":{"fileUri":"gs://a"}}`, 0},
		{"not image MIME", `{"inlineData":{"mimeType":"audio/mpeg","data":"abc"}}`, 0},
		{"multiple", `{"fileData":{"mimeType":"image/png","fileUri":"gs://a"}},{"inline_data":{"mime_type":"image/png","data":"abc"}},{"file_data":{"mime_type":"image/webp","file_uri":"gs://b"}}`, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, countGeminiInlineImageOutputs([]byte(geminiImageResponse(tc.parts))))
		})
	}
	require.Zero(t, countGeminiInlineImageOutputs([]byte("not json")))
}

func TestUpstreamAFixesGeminiImageCumulativeResetFallback(t *testing.T) {
	c := newGeminiImageTestContext(t)
	beginGeminiImageOutputObservation(c)
	one := geminiImageResponse(`{"fileData":{"mimeType":"image/png","fileUri":"gs://a"}}`)
	two := geminiImageResponse(`{"fileData":{"mimeType":"image/png","fileUri":"gs://a"}},{"file_data":{"mime_type":"image/png","file_uri":"gs://b"}}`)
	for _, frame := range []string{one, two, two, one, `{"usageMetadata":{"promptTokenCount":8}}`} {
		observeGeminiImageOutputs(c, []byte(frame))
	}
	require.Equal(t, 2, resolveGeminiImageCount(c, "alias", "mapped-alias"))
	beginGeminiImageOutputObservation(c)
	require.Zero(t, observedGeminiImageOutputs(c))
	require.Zero(t, resolveGeminiImageCount(c, "alias", "mapped-alias"))
	require.Equal(t, 1, resolveGeminiImageCount(c, "gemini-2.5-flash-image", "alias"))
	require.Equal(t, 1, resolveGeminiImageCount(c, "alias", "gemini-2.5-flash-image"))
}

func TestUpstreamAFixesGeminiImageForwardAccounting(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "nonstream", true: "stream"}[stream], func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			body := []byte(`{"contents":[{"role":"user","parts":[{"text":"draw"}]}]}`)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/custom-image-alias:generateContent", bytes.NewReader(body))
			payload := geminiImageResponse(`{"file_data":{"mime_type":"image/png","file_uri":"gs://bucket/a"}},{"fileData":{"mimeType":"image/webp","fileUri":"https://example.test/b.webp"}}`)
			responseBody := payload
			ct := "application/json"
			action := "generateContent"
			if stream {
				responseBody = "data: " + payload + "\n\ndata: " + payload + "\n\ndata: {\"usageMetadata\":{\"promptTokenCount\":9}}\n\n"
				ct = "text/event-stream"
				action = "streamGenerateContent"
			}
			upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{ct}},
				Body: io.NopCloser(strings.NewReader(responseBody)),
			}}
			svc := &GeminiMessagesCompatService{httpUpstream: upstream, cfg: &config.Config{}}
			account := &Account{ID: 7805, Platform: PlatformGemini, Type: AccountTypeAPIKey, Concurrency: 1,
				Credentials: map[string]any{"api_key": "test-key", "model_mapping": map[string]any{"custom-image-alias": "custom-upstream-image-alias"}}}
			// A previous attempt must not leak its output count into this forward.
			beginGeminiImageOutputObservation(c).count = 7
			result, err := svc.ForwardNative(context.Background(), c, account, "custom-image-alias", action, stream, body)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, 2, result.ImageCount)
			require.Equal(t, "custom-upstream-image-alias", result.UpstreamModel)
			groupID := int64(17805)
			billing := &GatewayService{billingService: NewBillingService(&config.Config{}, nil),
				resolver: newOpenAIImageChannelPricingResolverForTest(t, groupID, result.UpstreamModel, 0.25)}
			cost := billing.calculateRecordUsageCost(context.Background(), result,
				&APIKey{GroupID: &groupID, Group: &Group{ID: groupID}}, result.UpstreamModel, 0.15, 0.15, 1.0, nil)
			require.NotNil(t, cost)
			require.Equal(t, string(BillingModeImage), cost.BillingMode)
			require.InDelta(t, 0.50, cost.TotalCost, 1e-12)
			require.InDelta(t, 0.50, cost.ActualCost, 1e-12)
		})
	}
}
