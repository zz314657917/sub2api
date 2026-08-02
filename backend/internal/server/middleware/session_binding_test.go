package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSessionBindingContextUsesConfiguredHeaderAndRequestSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(true, []string{"X-Client-IP"})

	r := gin.New()
	r.Use(SessionBindingContext(cfg))
	var binding *service.SessionBinding
	var securityIP string
	r.GET("/t", func(c *gin.Context) {
		binding = service.SessionBindingFromContext(c.Request.Context())
		// A hot update must not change the already-captured request snapshot.
		cfg.SetForwardedClientIPSettings(false, []string{"X-New-Client-IP"})
		securityIP = SecurityClientIP(c)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Client-IP", "1.2.3.4")
	req.Header.Set("X-New-Client-IP", "4.4.4.4")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.NotNil(t, binding)
	require.Equal(t, "1.2.3.4", binding.IP)
	require.Equal(t, "1.2.3.4", securityIP)
}

func TestSessionBindingContextKeepsRawForwardedHeadersDisabledByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	r := gin.New()
	r.Use(SessionBindingContext(cfg))
	r.GET("/t", func(c *gin.Context) {
		binding := service.SessionBindingFromContext(c.Request.Context())
		require.NotNil(t, binding)
		require.Equal(t, "9.9.9.9", binding.IP)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Client-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestSecurityClientIPFallbackUsesExplicitForwardedSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t", func(c *gin.Context) {
		// This simulates a narrowly composed route that does not mount the
		// SessionBindingContext middleware itself.
		ip.SetForwardedIPSettings(c, true, []string{"X-Client-IP"})
		require.Equal(t, "1.2.3.4", SecurityClientIP(c))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Client-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAuditUsesTheSameRequestSnapshotAsSessionBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(true, []string{"X-Client-IP"})
	repository := &auditCaptureRepository{}
	auditService := service.NewAuditLogService(repository, nil)
	auditService.Start()

	r := gin.New()
	r.Use(SessionBindingContext(cfg))
	r.Use(gin.HandlerFunc(NewAuditLogMiddleware(auditService)))
	r.POST("/t", func(c *gin.Context) {
		cfg.SetForwardedClientIPSettings(false, []string{"X-New-Client-IP"})
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Client-IP", "1.2.3.4")
	req.Header.Set("X-New-Client-IP", "4.4.4.4")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	auditService.Stop()

	require.Equal(t, http.StatusCreated, rec.Code)
	repository.mu.Lock()
	logs := append([]*service.AuditLog(nil), repository.logs...)
	repository.mu.Unlock()
	require.Len(t, logs, 1)
	require.Equal(t, "1.2.3.4", logs[0].ClientIP)
}
