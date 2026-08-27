package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestS272ResolveAPIKeyForModelRequestRejectsSingleGroupModelMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(41)
	apiKey := s272SingleGroupAPIKey(groupID, []string{"gpt-5.6-sol", "gpt-5.6-terra"})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{})

	resolved, ok := ResolveAPIKeyForModelRequest(c, apiKeyService, apiKey, "gpt-5.6-luna", false)

	require.False(t, ok)
	require.Nil(t, resolved)
	require.True(t, c.IsAborted())
	require.Equal(t, http.StatusForbidden, recorder.Code)
	var body ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "NO_MATCHING_GROUP_ROUTE", body.Code)
}

func TestS272ResolveAPIKeyForModelRequestAllowsSingleGroupModelMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(41)
	apiKey := s272SingleGroupAPIKey(groupID, []string{"gpt-5.6-*"})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{})

	resolved, ok := ResolveAPIKeyForModelRequest(c, apiKeyService, apiKey, "gpt-5.6-luna", false)

	require.True(t, ok)
	require.Same(t, apiKey, resolved)
	require.False(t, c.IsAborted())
	require.Equal(t, http.StatusOK, recorder.Code)
}

func s272SingleGroupAPIKey(groupID int64, modelMatchPatterns []string) *service.APIKey {
	return &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:                 groupID,
			Platform:           service.PlatformOpenAI,
			Status:             service.StatusActive,
			Hydrated:           true,
			RoutingScope:       service.GroupRoutingScopeInference,
			ModelMatchPatterns: modelMatchPatterns,
		},
	}
}
