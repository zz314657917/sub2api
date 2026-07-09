package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentCreateOrderRejectsGroupBuyOrderType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42, Concurrency: 1})
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/orders",
		bytes.NewBufferString(`{"amount":128,"payment_type":"alipay","order_type":"`+payment.OrderTypeGroupBuy+`"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	NewPaymentHandler(nil, nil, nil).CreateOrder(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "GROUP_BUY_ORDER_ROUTE_REQUIRED")
}
