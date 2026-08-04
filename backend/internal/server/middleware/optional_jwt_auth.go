package middleware

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewOptionalJWTAuthMiddleware accepts anonymous requests, while still using
// the strict session-bound JWT validator whenever an Authorization header is
// supplied. Invalid credentials must not silently become an anonymous view.
func NewOptionalJWTAuthMiddleware(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	auditService *service.AuditLogService,
) OptionalJWTAuthMiddleware {
	strict := jwtAuthWithSessionBinding(authService, userService, userService, settingService, auditService)
	return OptionalJWTAuthMiddleware(func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader("Authorization")) == "" {
			c.Next()
			return
		}
		strict(c)
	})
}
