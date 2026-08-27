package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type contentModerationHandlerProxyRepo struct {
	service.ProxyRepository
	proxy *service.Proxy
}

func (r *contentModerationHandlerProxyRepo) GetByID(context.Context, int64) (*service.Proxy, error) {
	return r.proxy, nil
}

func TestContentModerationHandlerUpdateConfigMapsThresholdsAndProxyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settings := newTestSettingRepo()
	proxyRepo := &contentModerationHandlerProxyRepo{proxy: &service.Proxy{
		ID:       7,
		Status:   service.StatusActive,
		Protocol: "http",
		Host:     "127.0.0.1",
		Port:     8080,
	}}
	handler := NewContentModerationHandler(service.NewContentModerationService(
		settings, nil, nil, nil, nil, nil, nil, proxyRepo,
	))
	router := gin.New()
	router.PUT("/risk-control/config", handler.UpdateConfig)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/risk-control/config", strings.NewReader(`{"proxy_id":7,"thresholds":{"sexual":0.72}}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var saved service.ContentModerationConfig
	require.NoError(t, json.Unmarshal([]byte(settings.values[service.SettingKeyContentModerationConfig]), &saved))
	require.NotNil(t, saved.ProxyID)
	require.Equal(t, int64(7), *saved.ProxyID)
	require.Equal(t, 0.72, saved.Thresholds["sexual"])
}
