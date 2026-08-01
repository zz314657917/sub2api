package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiV1BetaInvalidModelPathSegments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &GatewayHandler{}
	groupID := int64(1)
	apiKey := &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformGemini,
		},
	}

	invalidModels := []string{
		"../x",
		"gemini-2.5-pro/../../x",
		"gemini-2.5-pro?alt=sse",
		"gemini-2.5-pro%2f..",
		" gemini-2.5-pro",
		"gemini-2.5-pro ",
		"...",
		"\u6a21\u578b",
	}
	for _, model := range invalidModels {
		t.Run("get_model_"+model, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models/invalid", nil)
			c.Params = gin.Params{{Key: "model", Value: model}}
			middleware.SetAPIKeyContext(c, apiKey)

			h.GeminiV1BetaGetModel(c)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "Invalid model in URL")
		})

		t.Run("models_action_"+model, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/invalid:generateContent", strings.NewReader(`{}`))
			c.Params = gin.Params{{Key: "modelAction", Value: "/" + model + ":generateContent"}}
			middleware.SetAPIKeyContext(c, apiKey)
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})

			h.GeminiV1BetaModels(c)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "Invalid model in URL")
		})
	}
}
