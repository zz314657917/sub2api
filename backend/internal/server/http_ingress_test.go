package server

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func clientIPForIngressTest(t *testing.T, serverConfig config.ServerConfig) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	configureTrustedProxies(r, serverConfig)
	r.GET("/ip", func(c *gin.Context) {
		c.String(200, c.ClientIP())
	})

	req := httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	return w.Body.String()
}

func TestConfigureTrustedProxiesMissingFailsClosed(t *testing.T) {
	require.Equal(t, "9.9.9.9", clientIPForIngressTest(t, config.ServerConfig{}))
}

func TestConfigureTrustedProxiesUnmarkedListFailsClosed(t *testing.T) {
	require.Equal(t, "9.9.9.9", clientIPForIngressTest(t, config.ServerConfig{
		TrustedProxies: []string{"9.9.9.9"},
	}))
}

func TestConfigureTrustedProxiesExplicitEmptyFailsClosed(t *testing.T) {
	require.Equal(t, "9.9.9.9", clientIPForIngressTest(t, config.ServerConfig{
		TrustedProxiesConfigured: true,
	}))
}

func TestConfigureTrustedProxiesAcceptsExplicitTrustedProxy(t *testing.T) {
	require.Equal(t, "1.2.3.4", clientIPForIngressTest(t, config.ServerConfig{
		TrustedProxiesConfigured: true,
		TrustedProxies:           []string{"9.9.9.9"},
	}))
}

func TestConfigureTrustedProxiesInvalidInputFailsClosed(t *testing.T) {
	require.Equal(t, "9.9.9.9", clientIPForIngressTest(t, config.ServerConfig{
		TrustedProxiesConfigured: true,
		TrustedProxies:           []string{"not-a-cidr"},
	}))
}
