package server

import (
	"context"
	"log"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/routes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const frameSrcRefreshTimeout = 5 * time.Second

type studioBridgeFrameAncestorAllowlist struct {
	origins []string
	domains []string
}

// SetupRouter 配置路由器中间件和路由
func SetupRouter(
	r *gin.Engine,
	handlers *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
	redisClient *redis.Client,
) *gin.Engine {
	// 缓存 iframe 页面的 origin 列表，用于动态注入 CSP frame-src
	var cachedFrameOrigins atomic.Pointer[[]string]
	emptyOrigins := []string{}
	cachedFrameOrigins.Store(&emptyOrigins)
	var cachedStudioBridgeFrameAncestors atomic.Pointer[studioBridgeFrameAncestorAllowlist]
	cachedStudioBridgeFrameAncestors.Store(&studioBridgeFrameAncestorAllowlist{})

	refreshFrameOrigins := func() {
		ctx, cancel := context.WithTimeout(context.Background(), frameSrcRefreshTimeout)
		defer cancel()
		origins, err := settingService.GetFrameSrcOrigins(ctx)
		if err != nil {
			// 获取失败时保留已有缓存，避免 frame-src 被意外清空
			return
		}
		cachedFrameOrigins.Store(&origins)
		studioBridgeSettings, err := settingService.GetStudioBridgeLuoyeAISettings(ctx)
		if err == nil {
			cachedStudioBridgeFrameAncestors.Store(studioBridgeFrameAncestorAllowlistFromSettings(studioBridgeSettings))
		}
	}
	refreshFrameOrigins() // 启动时初始化

	// 应用中间件
	r.Use(middleware2.RequestLogger())
	r.Use(middleware2.Logger())
	r.Use(middleware2.CORS(cfg.CORS))
	r.Use(middleware2.SecurityHeadersWithOptions(cfg.Security.CSP, middleware2.SecurityHeadersOptions{
		GetFrameSrcOrigins: func() []string {
			if p := cachedFrameOrigins.Load(); p != nil {
				return *p
			}
			return nil
		},
		GetFrameAncestors: func(c *gin.Context) []string {
			if c == nil || c.Request == nil || c.Request.URL == nil || c.Request.URL.Path != "/studio-bridge/session-probe" {
				return nil
			}
			origin := frameOriginFromURL(c.Query("parent_origin"))
			if origin == "" {
				return nil
			}
			allowlist := cachedStudioBridgeFrameAncestors.Load()
			if !studioBridgeFrameAncestorAllowed(origin, allowlist) {
				return nil
			}
			return []string{origin}
		},
	}))

	// Serve embedded frontend with settings injection if available
	if web.HasEmbeddedFrontend() {
		frontendServer, err := web.NewFrontendServer(settingService)
		if err != nil {
			log.Printf("Warning: Failed to create frontend server with settings injection: %v, using legacy mode", err)
			r.Use(web.ServeEmbeddedFrontend())
			settingService.SetOnUpdateCallback(refreshFrameOrigins)
		} else {
			// Register combined callback: invalidate HTML cache + refresh frame origins
			settingService.SetOnUpdateCallback(func() {
				frontendServer.InvalidateCache()
				refreshFrameOrigins()
			})
			r.Use(frontendServer.Middleware())
		}
	} else {
		settingService.SetOnUpdateCallback(refreshFrameOrigins)
	}

	// 注册路由
	registerRoutes(r, handlers, jwtAuth, adminAuth, apiKeyAuth, apiKeyService, subscriptionService, opsService, settingService, cfg, redisClient)

	return r
}

func appendFrameOrigin(origins []string, rawURL string) []string {
	origin := frameOriginFromURL(rawURL)
	if origin == "" {
		return origins
	}
	for _, existing := range origins {
		if existing == origin {
			return origins
		}
	}
	return append(origins, origin)
}

func frameOriginFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func studioBridgeFrameAncestorAllowlistFromSettings(settings *service.StudioBridgeAppSettings) *studioBridgeFrameAncestorAllowlist {
	allowlist := &studioBridgeFrameAncestorAllowlist{}
	if settings == nil || !settings.Enabled {
		return allowlist
	}
	if origin := frameOriginFromURL(settings.LaunchReturnURL); origin != "" {
		allowlist.origins = appendFrameOrigin(allowlist.origins, origin)
	}
	for _, domain := range settings.AllowedReturnDomains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" || strings.ContainsAny(domain, "/:") {
			continue
		}
		allowlist.domains = append(allowlist.domains, domain)
	}
	return allowlist
}

func studioBridgeFrameAncestorAllowed(origin string, allowlist *studioBridgeFrameAncestorAllowlist) bool {
	origin = frameOriginFromURL(origin)
	if origin == "" || allowlist == nil {
		return false
	}
	for _, allowed := range allowlist.origins {
		if origin == allowed {
			return true
		}
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Hostname() == "" {
		return false
	}
	if parsed.Scheme != "https" && !isLoopbackFrameAncestorHost(parsed.Hostname()) {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, domain := range allowlist.domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func isLoopbackFrameAncestorHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasPrefix(host, "127.")
}

// registerRoutes 注册所有 HTTP 路由
func registerRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
	redisClient *redis.Client,
) {
	// 通用路由（健康检查、状态等）
	routes.RegisterCommonRoutes(r)

	// API v1
	v1 := r.Group("/api/v1")

	// 注册各模块路由
	routes.RegisterAuthRoutes(v1, h, jwtAuth, redisClient, settingService, cfg)
	routes.RegisterUserRoutes(v1, h, jwtAuth, settingService)
	routes.RegisterAdminRoutes(v1, h, adminAuth)
	routes.RegisterGatewayRoutes(r, h, apiKeyAuth, apiKeyService, subscriptionService, opsService, settingService, cfg)
	routes.RegisterPaymentRoutes(v1, h.Payment, h.PaymentWebhook, h.Admin.Payment, jwtAuth, adminAuth, settingService)
	routes.RegisterMembershipRoutes(v1, h.Membership, jwtAuth, adminAuth, settingService)

	handler.RegisterPageRoutes(v1, cfg.Pricing.DataDir, gin.HandlerFunc(jwtAuth), gin.HandlerFunc(adminAuth), settingService)
}
