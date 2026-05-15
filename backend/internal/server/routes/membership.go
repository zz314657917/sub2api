package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterMembershipRoutes(
	v1 *gin.RouterGroup,
	membershipHandler *handler.MembershipHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	settingService *service.SettingService,
) {
	if membershipHandler == nil {
		return
	}

	authenticated := v1.Group("/membership")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/status", membershipHandler.GetStatus)
	}

	adminSettings := v1.Group("/admin/settings")
	adminSettings.Use(gin.HandlerFunc(adminAuth))
	{
		adminSettings.GET("/membership", membershipHandler.GetAdminSettings)
		adminSettings.PUT("/membership", membershipHandler.UpdateAdminSettings)
	}
}
