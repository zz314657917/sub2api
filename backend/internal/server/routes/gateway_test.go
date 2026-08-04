package routes

import (
	"bytes"
	"io"
	"mime/multipart"
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

func newGatewayRoutesTestRouter(platforms ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	platform := service.PlatformOpenAI
	if len(platforms) > 0 && strings.TrimSpace(platforms[0]) != "" {
		platform = strings.TrimSpace(platforms[0])
	}
	openAIGatewayHandler := &handler.OpenAIGatewayHandler{}
	asyncImageHandler := handler.NewAsyncImageHandler(service.NewImageTaskService(nil), openAIGatewayHandler)

	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:       &handler.GatewayHandler{},
			OpenAIGateway: openAIGatewayHandler,
			AsyncImage:    asyncImageHandler,
		},
		servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
			groupID := int64(1)
			c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
				GroupID: &groupID,
				Group: &service.Group{
					Platform:              platform,
					AllowMessagesDispatch: true,
				},
			})
			c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 1, Concurrency: 1})
			c.Next()
		}),
		nil,
		nil,
		nil,
		nil,
		&config.Config{
			Gateway: config.GatewayConfig{MaxBodySize: 1024 * 1024},
		},
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

func TestGatewayRoutesResponsesSubpathRejectsNonConformingSubpaths(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	paths := []string{
		"/v1/responses/../../x/y",
		"/v1/responses/..%2f..%2fx/y",
		"/v1/responses/%2e%2e/%2e%2e/x",
		"/responses/%2e%2e%2fx",
		"/backend-api/codex/responses/..%2f..%2fx",
		`/v1/responses/..\..\x`,
		"/v1/responses/%3fa=b",
		"/v1/responses/x%23frag",
		"/v1/responses/compact%2f..",
		"/v1/responses/compact%20",
		"/v1/responses/" + strings.Repeat("a", 129),
		"/v1/responses/a/b/c/d/e/f/g/h/i",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-5"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusNotFound, w.Code, "path=%s must be rejected at the edge", path)
			require.Contains(t, w.Body.String(), "Unsupported responses subpath", "path=%s", path)
		})
	}
}

func TestGatewayRoutesOpenAIImagesPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{
		"/v1/images/generations",
		"/v1/images/edits",
		"/v1/midjourney/generations",
		"/images/generations",
		"/images/edits",
		"/midjourney/generations",
	} {
		model := "gpt-image-2"
		if strings.Contains(path, "midjourney") {
			model = "midjourney"
		}
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"`+model+`","prompt":"draw a cat"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI images handler", path)
	}
}

func TestGatewayRoutesAsyncImagePathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()
	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /v1/images/generations/async",
		"POST /v1/images/edits/async",
		"GET /v1/images/tasks/:task_id",
		"POST /images/generations/async",
		"POST /images/edits/async",
		"GET /images/tasks/:task_id",
	} {
		require.True(t, routes[route], "missing route %s", route)
	}
}

func TestResolveAPIKeyRouteForJSONModelReadsMultipartImageModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "gpt-image-2"))
	require.NoError(t, writer.WriteField("prompt", "edit this image"))
	part, err := writer.CreateFormFile("image", "input.png")
	require.NoError(t, err)
	_, err = io.WriteString(part, "png")
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	defaultGroupID := int64(1)
	imageGroupID := int64(2)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", bytes.NewReader(body.Bytes()))
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &defaultGroupID,
		Group: &service.Group{
			ID:       defaultGroupID,
			Platform: service.PlatformOpenAI,
			Status:   service.StatusActive,
			Hydrated: true,
		},
		MultiGroupRouteGroups: []*service.Group{
			{
				ID:                   imageGroupID,
				Platform:             service.PlatformOpenAI,
				Status:               service.StatusActive,
				RoutingScope:         service.GroupRoutingScopeImage,
				AllowImageGeneration: true,
				ModelMatchPatterns:   []string{"gpt-image-*"},
				Hydrated:             true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	})

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/images/edits", true)

	require.True(t, ok)
	resolved, exists := servermiddleware.GetAPIKeyFromContext(c)
	require.True(t, exists)
	require.NotNil(t, resolved.Group)
	require.Equal(t, imageGroupID, *resolved.GroupID)
}

func TestGatewayRoutesOpenAIVideosPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/v1/videos/generations", strings.NewReader(`{"model":"doubao-seedance-2.0","prompt":"make a video"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusNotFound, w.Code)

	taskReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_123?language=zh", nil)
	taskW := httptest.NewRecorder()

	router.ServeHTTP(taskW, taskReq)
	require.NotEqual(t, http.StatusNotFound, taskW.Code)
}

func TestGatewayRoutesGrokAllowsCLICompatibilityEntrypoints(t *testing.T) {
	router := newGatewayRoutesTestRouter(service.PlatformGrok)

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/v1/messages"},
		{http.MethodPost, "/v1/chat/completions"},
		{http.MethodPost, "/chat/completions"},
		{http.MethodGet, "/v1/responses"},
		{http.MethodGet, "/responses"},
		{http.MethodGet, "/backend-api/codex/responses"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"model":"grok"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "method=%s path=%s", tc.method, tc.path)
		require.NotContains(t, w.Body.String(), "not supported for Grok groups")
	}

	for _, path := range []string{
		"/v1/responses",
		"/responses",
		"/backend-api/codex/responses",
	} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"grok","input":"hi"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should still reach Responses handler", path)
	}
}

func TestGatewayRoutesGrokCountTokensIsRejectedAtPlatformGate(t *testing.T) {
	router := newGatewayRoutesTestRouter(service.PlatformGrok)

	req := httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", strings.NewReader(`{"model":"grok","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "Token counting is not supported for this platform")
}

func TestGatewayRoutesOpenAICountTokensPathIsRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter(service.PlatformOpenAI)

	req := httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", strings.NewReader(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestGatewayRoutesNonOpenAICountTokensPathStillRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter(service.PlatformAnthropic)

	req := httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusNotFound, w.Code)
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
			ID:           defaultGroupID,
			Platform:     service.PlatformAnthropic,
			Status:       service.StatusActive,
			Hydrated:     true,
			RoutingScope: service.GroupRoutingScopeInference,
		},
		MultiGroupRouteGroups: []*service.Group{
			{
				ID:                    openAIGroupID,
				Platform:              service.PlatformOpenAI,
				Status:                service.StatusActive,
				Hydrated:              true,
				RoutingScope:          service.GroupRoutingScopeInference,
				ModelMatchPatterns:    []string{"gpt-*"},
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

func TestResolveAPIKeyRouteForJSONModel_GrokImageIntentPassiveNamespaceUsesTextRoute(t *testing.T) {
	c := newGrokImageIntentRoutingContext(`{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":"auto"}`)

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/responses", false)

	require.True(t, ok)
	resolved, exists := servermiddleware.GetAPIKeyFromContext(c)
	require.True(t, exists)
	require.NotNil(t, resolved.Group)
	require.Equal(t, int64(3), *resolved.GroupID)
	require.Equal(t, service.PlatformGrok, resolved.Group.Platform)
}

func TestResolveAPIKeyRouteForJSONModel_GrokImageIntentExplicitSignalUsesImageRoute(t *testing.T) {
	c := newGrokImageIntentRoutingContext(`{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":{"type":"namespace","name":"image_gen"}}`)

	ok := resolveAPIKeyRouteForJSONModel(c, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{}), "/v1/responses", false)

	require.True(t, ok)
	resolved, exists := servermiddleware.GetAPIKeyFromContext(c)
	require.True(t, exists)
	require.NotNil(t, resolved.Group)
	require.Equal(t, int64(2), *resolved.GroupID)
	require.Equal(t, service.PlatformOpenAI, resolved.Group.Platform)
}

func newGrokImageIntentRoutingContext(body string) *gin.Context {
	groupID := int64(1)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:           groupID,
			Platform:     service.PlatformGrok,
			Status:       service.StatusActive,
			Hydrated:     true,
			RoutingScope: service.GroupRoutingScopeInference,
		},
		MultiGroupRouteGroups: []*service.Group{
			{
				ID:                   2,
				Platform:             service.PlatformOpenAI,
				Status:               service.StatusActive,
				RoutingScope:         service.GroupRoutingScopeImage,
				ModelMatchPatterns:   []string{"grok-*"},
				AllowImageGeneration: true,
				Hydrated:             true,
			},
			{
				ID:                 3,
				Platform:           service.PlatformGrok,
				Status:             service.StatusActive,
				RoutingScope:       service.GroupRoutingScopeInference,
				ModelMatchPatterns: []string{"grok-*"},
				Hydrated:           true,
			},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{
				GroupID:       2,
				Enabled:       true,
				Priority:      1,
				Weight:        1,
				ImageOnly:     true,
				ModelPatterns: []string{"grok-*"},
			},
			{
				GroupID:       3,
				Enabled:       true,
				Priority:      2,
				Weight:        1,
				TextOnly:      true,
				ModelPatterns: []string{"grok-*"},
			},
		},
	})
	return c
}
