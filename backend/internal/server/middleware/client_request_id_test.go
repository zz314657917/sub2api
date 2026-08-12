package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestClientRequestIDGeneratesAndExposesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ClientRequestID())
	router.GET("/", func(c *gin.Context) {
		value, _ := c.Request.Context().Value(ctxkey.ClientRequestID).(string)
		startedAt, _ := c.Request.Context().Value(ctxkey.RequestStartedAt).(time.Time)
		require.False(t, startedAt.IsZero())
		c.String(http.StatusOK, value)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String())
	require.Equal(t, w.Body.String(), w.Header().Get(clientRequestIDHeader))
}

func TestClientRequestIDPreservesExistingContextID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ClientRequestID())
	router.GET("/", func(c *gin.Context) {
		value, _ := c.Request.Context().Value(ctxkey.ClientRequestID).(string)
		c.String(http.StatusOK, value)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.ClientRequestID, "existing-client-request-id"))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "existing-client-request-id", w.Body.String())
	require.Equal(t, "existing-client-request-id", w.Header().Get(clientRequestIDHeader))
}

func TestClientRequestIDPreservesExistingRequestStartedAt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ClientRequestID())
	want := time.Date(2026, time.August, 12, 20, 0, 0, 0, time.FixedZone("test", 8*60*60))
	router.GET("/", func(c *gin.Context) {
		got, _ := c.Request.Context().Value(ctxkey.RequestStartedAt).(time.Time)
		require.Equal(t, want, got)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.RequestStartedAt, want))
	router.ServeHTTP(httptest.NewRecorder(), req)
}
