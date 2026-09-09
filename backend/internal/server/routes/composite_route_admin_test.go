package routes

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCompositeRouteAdminRoutesRegisteredUnderAdminGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	v1 := engine.Group("/api/v1")
	admin := v1.Group("/admin")
	handlers := &handler.Handlers{Admin: &handler.AdminHandlers{
		Group: adminhandler.NewGroupHandler(nil, nil, nil),
	}}

	registerGroupRoutes(admin, handlers)

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{
		"GET /api/v1/admin/groups/:id/composite-routes",
		"POST /api/v1/admin/groups/:id/composite-routes",
		"POST /api/v1/admin/groups/:id/composite-routes/preview",
		"PUT /api/v1/admin/groups/:id/composite-routes/:route_id",
		"DELETE /api/v1/admin/groups/:id/composite-routes/:route_id",
	} {
		require.True(t, routes[expected], "missing route %s", expected)
	}
}
