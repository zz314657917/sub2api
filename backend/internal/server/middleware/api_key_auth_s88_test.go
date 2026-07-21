package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestS88ResolveAPIKeyForModelRequestRejectsIncompatibleDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(1)
	apiKey := &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:                   groupID,
			Platform:             service.PlatformOpenAI,
			Status:               service.StatusActive,
			Hydrated:             true,
			RoutingScope:         service.GroupRoutingScopeImage,
			AllowImageGeneration: true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: groupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{})

	resolved, ok := ResolveAPIKeyForModelRequest(c, apiKeyService, apiKey, "gpt-5.4", false)

	require.False(t, ok)
	require.Nil(t, resolved)
	require.True(t, c.IsAborted())
	require.Equal(t, http.StatusForbidden, recorder.Code)
	var body ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "NO_MATCHING_GROUP_ROUTE", body.Code)
}
