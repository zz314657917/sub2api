package admin

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpsRealtimeRequestCanceled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	require.True(t, isOpsRealtimeRequestCanceled(nil, context.Canceled))
	require.True(t, isOpsRealtimeRequestCanceled(nil, errors.New("pq: canceling statement due to user request")))
	require.False(t, isOpsRealtimeRequestCanceled(nil, errors.New("database unavailable")))
	require.False(t, isOpsRealtimeRequestCanceled(nil, nil))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest("GET", "/ops", nil).WithContext(ctx)
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = req

	require.True(t, isOpsRealtimeRequestCanceled(ginCtx, errors.New("wrapped query failed")))
}
