package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIImagesResponsesDriverAndImageModels(t *testing.T) {
	for _, override := range []string{"", "  ", " gpt-5.6-sol "} {
		t.Run(fmt.Sprintf("override=%q", override), func(t *testing.T) {
			t.Setenv("SUB2API_IMAGES_MAIN_MODEL", override)
			driver := strings.TrimSpace(override)
			if driver == "" {
				driver = "gpt-5.6-luna"
			}
			for _, model := range []string{"gpt-image-2", "gpt-image-2.5-flare", "gpt-image-2.5-sunburst", "gpt-image-2.5-flare-2026-09-08"} {
				parsed := &OpenAIImagesRequest{Endpoint: openAIImagesGenerationsEndpoint, Model: model, Prompt: "draw a cup", Quality: "xhigh", Size: "1536x864", N: 1}
				body, err := buildOpenAIImagesResponsesRequest(parsed, model)
				require.NoError(t, err)
				require.Equal(t, driver, gjson.GetBytes(body, "model").String())
				require.Equal(t, model, gjson.GetBytes(body, "tools.0.model").String())

				req := map[string]any{"model": model, "input": "draw a cup"}
				require.True(t, normalizeOpenAIResponsesImageOnlyModel(req))
				require.Equal(t, driver, req["model"])
				require.Equal(t, model, req["tools"].([]any)[0].(map[string]any)["model"])
			}
		})
	}
}

type openAIImagesPlanGatedRepoStub struct {
	AccountRepository
	modelRateLimitCalls []openAIImagesModelRateLimitCall
	tempCalls           int
}

type openAIImagesModelRateLimitCall struct {
	id      int64
	scope   string
	resetAt time.Time
}

func (r *openAIImagesPlanGatedRepoStub) SetModelRateLimit(_ context.Context, id int64, scope string, resetAt time.Time) error {
	r.modelRateLimitCalls = append(r.modelRateLimitCalls, openAIImagesModelRateLimitCall{id: id, scope: scope, resetAt: resetAt})
	return nil
}

func (r *openAIImagesPlanGatedRepoStub) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, _ string) error {
	r.tempCalls++
	return nil
}

func TestOpenAIImagesRejectedDriverDoesNotCoolImageModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("SUB2API_IMAGES_MAIN_MODEL", "gpt-5.4-mini")
	repo := &openAIImagesPlanGatedRepoStub{}
	svc := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test"}}

	for _, tc := range []struct {
		rejected      string
		wantCooldowns int
		wantUpstream  bool
	}{
		{rejected: "gpt-5.4-mini", wantCooldowns: 0, wantUpstream: true},
		{rejected: "gpt-image-2.5-flare", wantCooldowns: 1},
	} {
		t.Run(tc.rejected, func(t *testing.T) {
			repo.modelRateLimitCalls = nil
			repo.tempCalls = 0
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, openAIImagesGenerationsEndpoint, nil)
			body := fmt.Sprintf(`{"error":{"message":"The '%s' model is not supported when using Codex with a ChatGPT account.","type":"invalid_request_error"}}`, tc.rejected)
			resp := &http.Response{StatusCode: 400, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
			_, err := svc.handleOpenAIImagesErrorResponse(context.Background(), resp, c, account, "gpt-image-2.5-flare")
			require.Error(t, err)
			if tc.wantUpstream {
				var upstreamErr *OpenAIImagesUpstreamError
				require.ErrorAs(t, err, &upstreamErr)
				require.Contains(t, upstreamErr.Message, tc.rejected)
			} else {
				var failoverErr *UpstreamFailoverError
				require.ErrorAs(t, err, &failoverErr)
			}
			require.Len(t, repo.modelRateLimitCalls, tc.wantCooldowns)
			require.Zero(t, repo.tempCalls)
			if tc.wantCooldowns == 1 {
				require.Equal(t, "gpt-image-2.5-flare", repo.modelRateLimitCalls[0].scope)
				require.WithinDuration(t, time.Now().Add(upstreamModelNotFoundCooldown), repo.modelRateLimitCalls[0].resetAt, 5*time.Second)
			}
		})
	}
}

func TestGPTImage25PricingDoesNotUseLegacyImageRates(t *testing.T) {
	for _, model := range []string{"gpt-image-2.5-flare", "gpt-image-2.5-sunburst", "gpt-image-2.5-flare-2026-09-08", "gpt-image-2.5-sunburst-2026-09-08"} {
		svc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{"gpt-image-2": {InputCostPerToken: 2.5e-6, OutputCostPerImageToken: 15e-6}}}
		pricing := svc.GetModelPricing(model)
		require.NotNil(t, pricing)
		require.Equal(t, 5e-6, pricing.InputCostPerToken)
		require.Equal(t, 8e-6, pricing.InputCostPerImageToken)
		require.Equal(t, 30e-6, pricing.OutputCostPerImageToken)
		require.Equal(t, 1.25e-6, pricing.CacheReadInputTokenCost)
		custom := &LiteLLMModelPricing{InputCostPerToken: 7e-6}
		svc.pricingData[model] = custom
		require.Same(t, custom, svc.GetModelPricing(model))
	}
}

func TestGPTImage25AccountModelPermissions(t *testing.T) {
	for _, model := range []string{"gpt-image-2.5-flare", "gpt-image-2.5-sunburst"} {
		account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{}}
		require.True(t, account.IsModelSupported(model))
		restricted := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2"}}}
		require.False(t, restricted.IsModelSupported(model))
	}
}
