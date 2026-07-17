package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayHandlerResponses_GrokPassiveImageNamespaceBypassesPermissionGate(t *testing.T) {
	body := `{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen","tools":[{"type":"function","name":"imagegen"}]}],"tool_choice":"auto","input":"write code"}`
	rec := runOpenAIResponsesImagePermissionGateTest(t, service.PlatformGrok, body)

	require.NotEqual(t, http.StatusForbidden, rec.Code)
	require.NotContains(t, rec.Body.String(), service.ImageGenerationPermissionMessage())
}

func TestOpenAIGatewayHandlerResponses_GrokExplicitImageSignalsStillHitPermissionGate(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "native image_generation tool",
			body: `{"model":"grok-4.5","tools":[{"type":"image_generation"}],"input":"draw"}`,
		},
		{
			name: "explicit namespace tool choice",
			body: `{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":{"type":"namespace","name":"image_gen"},"input":"draw"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := runOpenAIResponsesImagePermissionGateTest(t, service.PlatformGrok, tt.body)
			require.Equal(t, http.StatusForbidden, rec.Code)
			require.Contains(t, rec.Body.String(), service.ImageGenerationPermissionMessage())
		})
	}
}

func TestOpenAIGatewayHandlerResponses_OpenAIPassiveImageNamespacePreservesLegacyPermissionGate(t *testing.T) {
	body := `{"model":"gpt-5.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":"auto","input":"write code"}`
	rec := runOpenAIResponsesImagePermissionGateTest(t, service.PlatformOpenAI, body)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), service.ImageGenerationPermissionMessage())
}

func TestOpenAIGatewayHandlerChatCompletions_OpenAIPassiveImageNamespacePreservesLegacyPermissionGate(t *testing.T) {
	body := `{"model":"gpt-5.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":"auto","messages":[{"role":"user","content":"write code"}]}`
	rec := runOpenAIChatCompletionsImagePermissionGateTest(t, service.PlatformOpenAI, body)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), service.ImageGenerationPermissionMessage())
}

func runOpenAIResponsesImagePermissionGateTest(t *testing.T, platform string, body string) *httptest.ResponseRecorder {
	return runOpenAIImagePermissionGateTest(t, platform, "/v1/responses", body, func(h *OpenAIGatewayHandler, c *gin.Context) {
		h.Responses(c)
	})
}

func runOpenAIChatCompletionsImagePermissionGateTest(t *testing.T, platform string, body string) *httptest.ResponseRecorder {
	return runOpenAIImagePermissionGateTest(t, platform, "/v1/chat/completions", body, func(h *OpenAIGatewayHandler, c *gin.Context) {
		h.ChatCompletions(c)
	})
}

func runOpenAIImagePermissionGateTest(
	t *testing.T,
	platform string,
	path string,
	body string,
	invoke func(*OpenAIGatewayHandler, *gin.Context),
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	groupID := int64(6301)
	userID := int64(6302)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID:      6303,
		GroupID: &groupID,
		Group: &service.Group{
			ID:                   groupID,
			Platform:             platform,
			AllowImageGeneration: false,
		},
		User: &service.User{ID: userID, Status: service.StatusActive},
	})
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID, Concurrency: 1})

	h := &OpenAIGatewayHandler{
		gatewayService:      &service.OpenAIGatewayService{},
		billingCacheService: service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, &config.Config{RunMode: config.RunModeSimple}),
		apiKeyService:       &service.APIKeyService{},
		concurrencyHelper: &ConcurrencyHelper{concurrencyService: service.NewConcurrencyService(
			&helperConcurrencyCacheStub{userSeq: []bool{true}},
		)},
		cfg:          &config.Config{},
		imageLimiter: &imageConcurrencyLimiter{},
	}

	invoke(h, c)
	return rec
}
