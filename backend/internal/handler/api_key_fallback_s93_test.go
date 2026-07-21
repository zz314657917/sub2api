package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDefaultKeyFallbackS93HandlerRejectsMissingSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, &config.Config{})
	h := NewAPIKeyHandler(svc)
	router := gin.New()
	router.POST("/api/v1/admin/settings/default-key-fallback/backfill", h.BackfillDefaultKeyFallbackGroup)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/default-key-fallback/backfill", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "DEFAULT_KEY_FALLBACK_GROUP_REQUIRED")
}
