package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCafeAdminRoutesRegisterPendingFulfillmentWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/api/v1/admin")
	handlers := &handler.Handlers{Admin: &handler.AdminHandlers{CafeRoom: adminhandler.NewCafeRoomHandler(nil)}}
	registerCafeRoomRoutes(admin, handlers)

	registered := make(map[string]struct{})
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}
	for _, expected := range []string{
		http.MethodGet + " /api/v1/admin/cafe/layout",
		http.MethodPut + " /api/v1/admin/cafe/layout",
		http.MethodGet + " /api/v1/admin/cafe/rounds/pending",
		http.MethodGet + " /api/v1/admin/cafe/rounds/:id/account-options",
		http.MethodPost + " /api/v1/admin/cafe/rounds/:id/assign-account",
	} {
		_, ok := registered[expected]
		require.True(t, ok, expected)
	}
}
