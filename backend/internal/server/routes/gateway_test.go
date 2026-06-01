package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newGatewayRoutesTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:       &handler.GatewayHandler{},
			OpenAIGateway: &handler.OpenAIGatewayHandler{},
		},
		servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
			groupID := int64(1)
			c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
				GroupID: &groupID,
				Group:   &service.Group{Platform: service.PlatformOpenAI},
			})
			c.Next()
		}),
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
	)

	return router
}

func TestGatewayRoutesOpenAIResponsesCompactPathIsRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{
		"/v1/responses/compact",
		"/responses/compact",
		"/backend-api/codex/responses",
		"/backend-api/codex/responses/compact",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-5"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI responses handler", path)
	}
}

func TestGatewayRoutesOpenAIImagesPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{
		"/v1/images/generations",
		"/v1/images/edits",
		"/images/generations",
		"/images/edits",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-image-2","prompt":"draw a cat"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI images handler", path)
	}
}

func TestResolveAPIKeyRouteForJSONModelReroutesMessagesBeforeDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	defaultGroupID := int64(1)
	openAIGroupID := int64(2)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"gpt-5"}`))
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &defaultGroupID,
		Group: &service.Group{
			ID:       defaultGroupID,
			Platform: service.PlatformAnthropic,
			Status:   service.StatusActive,
			Hydrated: true,
		},
		MultiGroupRouteGroups: []*service.Group{
			{
				ID:                    openAIGroupID,
				Platform:              service.PlatformOpenAI,
				Status:                service.StatusActive,
				Hydrated:              true,
				AllowMessagesDispatch: true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{
				GroupID:       openAIGroupID,
				Enabled:       true,
				Priority:      1,
				Weight:        1,
				ModelPatterns: []string{"gpt-*"},
			},
		},
	})

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/messages", false)

	require.True(t, ok)
	resolved, ok := servermiddleware.GetAPIKeyFromContext(c)
	require.True(t, ok)
	require.NotNil(t, resolved.Group)
	require.Equal(t, openAIGroupID, *resolved.GroupID)
	require.Equal(t, service.PlatformOpenAI, resolved.Group.Platform)
}

func TestResolveAPIKeyRouteForJSONModelReturnsFalseWhenResolvedGroupUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(1)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-5"}`))
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			Status:   service.StatusDisabled,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Enabled: true, Priority: 1, Weight: 1},
		},
	})

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/chat/completions", false)

	require.False(t, ok)
	require.True(t, c.IsAborted())
}

func TestResolveAPIKeyRouteForJSONModelEnforcesDeferredGroupBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(1)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-5"}`))
	c.Set(string(servermiddleware.ContextKeyDeferredGroupBilling), true)
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		UserID:  9,
		GroupID: &groupID,
		User: &service.User{
			ID:      9,
			Balance: 0,
		},
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			Status:   service.StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Enabled: true, Priority: 1, Weight: 1},
		},
	})

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/chat/completions", false)

	require.False(t, ok)
	require.True(t, c.IsAborted())
	require.Equal(t, http.StatusForbidden, c.Writer.Status())
}
