package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户相关路由（需要认证）
func RegisterUserRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	internalStudioBridge := v1.Group("/internal/studio-bridge")
	{
		internalStudioBridge.POST("/redeem", h.StudioBridge.Redeem)
		internalStudioBridge.POST("/user-summary", h.StudioBridge.UserSummary)
		internalStudioBridge.POST("/charges/reserve", h.StudioBridge.Reserve)
		internalStudioBridge.POST("/charges/commit", h.StudioBridge.Commit)
		internalStudioBridge.POST("/charges/refund", h.StudioBridge.Refund)
	}

	authenticated := v1.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		// 用户接口
		user := authenticated.Group("/user")
		{
			user.GET("/profile", h.User.GetProfile)
			user.PUT("/password", h.User.ChangePassword)
			user.PUT("", h.User.UpdateProfile)
			user.GET("/aff", h.User.GetAffiliate)
			user.POST("/aff/transfer", h.User.TransferAffiliateQuota)
			user.POST("/aff/invitees/:invitee_id/api-call-reward/claim", h.User.ClaimAffiliateAPICallReward)
			user.POST("/account-bindings/email/send-code", h.User.SendEmailBindingCode)
			user.POST("/account-bindings/email", h.User.BindEmailIdentity)
			user.DELETE("/account-bindings/:provider", h.User.UnbindIdentity)
			user.POST("/auth-identities/bind/start", h.User.StartIdentityBinding)
			user.GET("/api-keys/:id/usage/daily", h.Usage.GetMyAPIKeyDailyUsage)

			studioBridge := user.Group("/studio-bridge")
			{
				studioBridge.POST("/launch", h.StudioBridge.Launch)
				studioBridge.GET("/session-probe", h.StudioBridge.SessionProbe)
			}

			imageCreator := user.Group("/image-creator")
			{
				imageCreator.POST("/tasks", h.ImageCreator.CreateTask)
				imageCreator.GET("/tasks", h.ImageCreator.ListTasks)
				imageCreator.GET("/tasks/:id", h.ImageCreator.GetTask)
				imageCreator.GET("/images", h.ImageCreator.ListImages)
				imageCreator.DELETE("/images", h.ImageCreator.DeleteImages)
				imageCreator.GET("/images/:id/file", h.ImageCreator.GetImageFile)
				imageCreator.GET("/images/:id/reference-file", h.ImageCreator.GetReferenceImageFile)
			}

			canvases := user.Group("/canvases")
			{
				canvases.GET("", h.Canvas.ListCanvases)
				canvases.POST("", h.Canvas.SaveCanvas)
				canvases.GET("/:id", h.Canvas.GetCanvas)
				canvases.PUT("/:id", h.Canvas.UpdateCanvas)
				canvases.DELETE("/:id", h.Canvas.DeleteCanvas)
			}

			canvasRuns := user.Group("/canvas-runs")
			{
				canvasRuns.GET("", h.Canvas.ListRuns)
				canvasRuns.POST("", h.Canvas.CreateRun)
				canvasRuns.GET("/:id", h.Canvas.GetRun)
				canvasRuns.POST("/:id/cancel", h.Canvas.CancelRun)
			}

			user.GET("/canvas/models", h.Canvas.ListModels)

			user.GET("/prompt-favorites", h.PromptFavorite.List)
			user.POST("/prompt-favorites", h.PromptFavorite.Save)
			user.DELETE("/prompt-favorites/:id", h.PromptFavorite.Delete)

			tickets := user.Group("/tickets")
			{
				tickets.GET("/unread-summary", h.Ticket.UnreadSummary)
				tickets.GET("", h.Ticket.List)
				tickets.POST("", h.Ticket.Create)
				tickets.GET("/:id", h.Ticket.Get)
				tickets.POST("/:id/messages", h.Ticket.AddMessage)
				tickets.POST("/:id/read", h.Ticket.MarkRead)
				tickets.POST("/:id/close", h.Ticket.Close)
			}

			welfare := user.Group("/welfare")
			{
				welfare.GET("/overview", h.Welfare.GetOverview)
				welfare.GET("/daily-checkin", h.Welfare.GetDailyCheckin)
				welfare.POST("/daily-checkin/claim", h.Welfare.ClaimDailyCheckin)
				welfare.POST("/daily-checkin/milestones/:day/claim", h.Welfare.ClaimDailyCheckinMilestone)
				welfare.POST("/new-user-trial/reward/claim", h.Welfare.ClaimNewUserTrialSuccessReward)
				welfare.POST("/recharge/first-bonus/claim", h.Welfare.ClaimFirstRechargeBonus)
			}

			proxies := user.Group("/proxies")
			{
				proxies.GET("", h.UserProxy.List)
				proxies.POST("", h.UserProxy.Create)
				proxies.PUT("/:id", h.UserProxy.Update)
				proxies.DELETE("/:id", h.UserProxy.Delete)
			}

			// 通知邮箱管理
			notifyEmail := user.Group("/notify-email")
			{
				notifyEmail.POST("/send-code", h.User.SendNotifyEmailCode)
				notifyEmail.POST("/verify", h.User.VerifyNotifyEmail)
				notifyEmail.PUT("/toggle", h.User.ToggleNotifyEmail)
				notifyEmail.DELETE("", h.User.RemoveNotifyEmail)
			}

			// TOTP 双因素认证
			totp := user.Group("/totp")
			{
				totp.GET("/status", h.Totp.GetStatus)
				totp.GET("/verification-method", h.Totp.GetVerificationMethod)
				totp.POST("/send-code", h.Totp.SendVerifyCode)
				totp.POST("/setup", h.Totp.InitiateSetup)
				totp.POST("/enable", h.Totp.Enable)
				totp.POST("/disable", h.Totp.Disable)
			}
		}

		// API Key管理
		keys := authenticated.Group("/keys")
		{
			keys.GET("", h.APIKey.List)
			keys.GET("/:id", h.APIKey.GetByID)
			keys.POST("", h.APIKey.Create)
			keys.PUT("/:id", h.APIKey.Update)
			keys.DELETE("/:id", h.APIKey.Delete)
		}

		// 用户账号共享池
		accounts := authenticated.Group("/user/accounts")
		{
			accounts.GET("", h.UserAccount.List)
			accounts.POST("", h.UserAccount.Create)
			accounts.POST("/import", h.UserAccount.Import)
			accounts.POST("/oauth/auth-url", h.UserAccount.GenerateAuthURL)
			accounts.POST("/oauth/exchange-code", h.UserAccount.ExchangeCode)
			accounts.POST("/session-import", h.UserAccount.ImportSession)
			accounts.GET("/share/summary", h.UserAccount.GetShareSummary)
			accounts.POST("/share/transfer", h.UserAccount.TransferShareToBalance)
			accounts.GET("/usage/summary", h.UserAccount.GetUsageSummary)
			accounts.GET("/capacity-pools", h.UserAccount.GetCapacityPools)
			accounts.GET("/:id", h.UserAccount.GetByID)
			accounts.PUT("/:id", h.UserAccount.Update)
			accounts.DELETE("/:id", h.UserAccount.Delete)
			accounts.POST("/:id/share-mode", h.UserAccount.UpdateShareMode)
			accounts.GET("/:id/usage", h.UserAccount.GetUsage)
			accounts.POST("/:id/test", h.UserAccount.Test)
		}

		// 用户可用分组（非管理员接口）
		groups := authenticated.Group("/groups")
		{
			groups.GET("/available", h.APIKey.GetAvailableGroups)
			groups.GET("/rates", h.APIKey.GetUserGroupRates)
		}

		// 用户可用渠道（非管理员接口）
		channels := authenticated.Group("/channels")
		{
			channels.GET("/available", h.AvailableChannel.List)
		}

		// 使用记录
		usage := authenticated.Group("/usage")
		{
			usage.GET("", h.Usage.List)
			usage.GET("/stats", h.Usage.Stats)
			// User dashboard endpoints
			usage.GET("/dashboard/stats", h.Usage.DashboardStats)
			usage.GET("/dashboard/leaderboard", h.Usage.DashboardLeaderboard)
			usage.POST("/dashboard/leaderboard/daily-reward/claim", h.Usage.ClaimDashboardLeaderboardDailyReward)
			usage.GET("/dashboard/trend", h.Usage.DashboardTrend)
			usage.GET("/dashboard/models", h.Usage.DashboardModels)
			usage.POST("/dashboard/api-keys-usage", h.Usage.DashboardAPIKeysUsage)
			usage.GET("/:id", h.Usage.GetByID)
		}

		// 公告（用户可见）
		announcements := authenticated.Group("/announcements")
		{
			announcements.GET("", h.Announcement.List)
			announcements.POST("/:id/read", h.Announcement.MarkRead)
		}

		// 卡密兑换
		redeem := authenticated.Group("/redeem")
		{
			redeem.POST("", h.Redeem.Redeem)
			redeem.GET("/history", h.Redeem.GetHistory)
		}

		// 用户订阅
		subscriptions := authenticated.Group("/subscriptions")
		{
			subscriptions.GET("", h.Subscription.List)
			subscriptions.GET("/active", h.Subscription.GetActive)
			subscriptions.GET("/progress", h.Subscription.GetProgress)
			subscriptions.GET("/summary", h.Subscription.GetSummary)
		}

		// 渠道监控（用户只读）
		monitors := authenticated.Group("/channel-monitors")
		{
			monitors.GET("", h.ChannelMonitor.List)
			monitors.GET("/:id/status", h.ChannelMonitor.GetStatus)
		}
	}
}
