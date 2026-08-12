package handler

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

func TestRequestStartedAtUsesMiddlewareTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	want := time.Date(2026, time.August, 12, 21, 59, 0, 0, time.FixedZone("test", 8*60*60))
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request = req.WithContext(context.WithValue(req.Context(), ctxkey.RequestStartedAt, want))

	require.Equal(t, want, requestStartedAt(c))
}

func TestRequestStartedAtFallsBackForInternalCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	before := time.Now()
	got := requestStartedAt(c)
	after := time.Now()
	require.False(t, got.Before(before))
	require.False(t, got.After(after))
}
