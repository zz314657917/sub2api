package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRouteCooldownUsesStreamErrorMarkerAtHTTP200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defaultGroupID := int64(14)
	fallbackGroupID := int64(3)
	apiKey := &service.APIKey{
		ID:      2,
		GroupID: &defaultGroupID,
		Group: &service.Group{
			ID:       defaultGroupID,
			Platform: service.PlatformOpenAI,
			Status:   service.StatusActive,
			Hydrated: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true},
			{GroupID: fallbackGroupID, Priority: 2, Weight: 1, CooldownSeconds: 30, Enabled: true},
		},
		MultiGroupRouteGroups: []*service.Group{{
			ID:       fallbackGroupID,
			Platform: service.PlatformOpenAI,
			Status:   service.StatusActive,
			Hydrated: true,
		}},
	}
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Status(http.StatusOK)
	MarkAPIKeyRouteCooldown(c, http.StatusTooManyRequests)

	applyAPIKeyRouteCooldownAfterRequest(c, apiKeyService, apiKey)
	resolved := apiKeyService.ResolveForRequest(context.Background(), apiKey, c.Request.URL.Path, "")

	require.True(t, IsAPIKeyRouteCooldownMarked(c))
	require.Equal(t, http.StatusOK, c.Writer.Status())
	require.NotNil(t, resolved)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, fallbackGroupID, *resolved.GroupID)
}
