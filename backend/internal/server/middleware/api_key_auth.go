package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewAPIKeyAuthMiddleware 创建 API Key 认证中间件
func NewAPIKeyAuthMiddleware(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) APIKeyAuthMiddleware {
	return APIKeyAuthMiddleware(apiKeyAuthWithSubscription(apiKeyService, subscriptionService, cfg))
}

// apiKeyAuthWithSubscription API Key认证中间件（支持订阅验证）
//
// 中间件职责分为两层：
//   - 鉴权（Authentication）：验证 Key 有效性、用户状态、IP 限制 —— 始终执行
//   - 计费执行（Billing Enforcement）：过期/配额/订阅/余额检查 —— skipBilling 时整块跳过
//
// /v1/usage 端点只需鉴权，不需要计费执行（允许过期/配额耗尽的 Key 查询自身用量）。
func apiKeyAuthWithSubscription(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. 提取 API Key ──────────────────────────────────────────

		queryKey := strings.TrimSpace(c.Query("key"))
		queryApiKey := strings.TrimSpace(c.Query("api_key"))
		if queryKey != "" || queryApiKey != "" {
			AbortWithError(c, 400, "api_key_in_query_deprecated", "API key in query parameter is deprecated. Please use Authorization header instead.")
			return
		}

		// 尝试从Authorization header中提取API key (Bearer scheme)
		authHeader := c.GetHeader("Authorization")
		var apiKeyString string

		if authHeader != "" {
			// 验证Bearer scheme
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				apiKeyString = strings.TrimSpace(parts[1])
			}
		}

		// 如果Authorization header中没有，尝试从x-api-key header中提取
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-api-key")
		}

		// 如果x-api-key header中没有，尝试从x-goog-api-key header中提取（Gemini CLI兼容）
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-goog-api-key")
		}

		// 如果所有header都没有API key
		if apiKeyString == "" {
			AbortWithError(c, 401, "API_KEY_REQUIRED", "API key is required in Authorization header (Bearer scheme), x-api-key header, or x-goog-api-key header")
			return
		}

		// ── 2. 验证 Key 存在 ─────────────────────────────────────────

		apiKey, err := apiKeyService.GetByKey(c.Request.Context(), apiKeyString)
		if err != nil {
			if errors.Is(err, service.ErrAPIKeyNotFound) {
				AbortWithError(c, 401, "INVALID_API_KEY", "Invalid API key")
				return
			}
			AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to validate API key")
			return
		}

		// ── 3. 基础鉴权（始终执行） ─────────────────────────────────

		// disabled / 未知状态 → 无条件拦截（expired 和 quota_exhausted 留给计费阶段）
		if !apiKey.IsActive() &&
			apiKey.Status != service.StatusAPIKeyExpired &&
			apiKey.Status != service.StatusAPIKeyQuotaExhausted {
			AbortWithError(c, 401, "API_KEY_DISABLED", "API key is disabled")
			return
		}

		// 检查 IP 限制（白名单/黑名单）
		// 注意：错误信息故意模糊，避免暴露具体的 IP 限制机制
		if len(apiKey.IPWhitelist) > 0 || len(apiKey.IPBlacklist) > 0 {
			clientIP := ip.GetTrustedClientIP(c)
			allowed, _ := ip.CheckIPRestrictionWithCompiledRules(clientIP, apiKey.CompiledIPWhitelist, apiKey.CompiledIPBlacklist)
			if !allowed {
				if clientIP == "" {
					clientIP = "unknown"
				}
				service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonIPRestriction)
				AbortWithError(c, 403, "ACCESS_DENIED", fmt.Sprintf("Access denied. Your IP is %s", clientIP))
				return
			}
		}

		// 检查关联的用户
		if apiKey.User == nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "User associated with API key not found")
			return
		}

		// 检查用户状态
		if !apiKey.User.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
			return
		}
		apiKey = resolveAPIKeyForRequest(c, apiKeyService, apiKey)
		if abortIfAPIKeyGroupUnavailable(c, apiKey) {
			return
		}
		if abortIfAPIKeyGroupNotAllowed(c, apiKey) {
			return
		}

		// ── 4. SimpleMode → early return ─────────────────────────────

		if cfg.RunMode == config.RunModeSimple {
			SetAPIKeyContext(c, apiKey)
			c.Set(string(ContextKeyUser), AuthSubject{
				UserID:      apiKey.User.ID,
				Concurrency: apiKey.User.Concurrency,
			})
			c.Set(string(ContextKeyUserRole), apiKey.User.Role)
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			c.Next()
			applyAPIKeyRouteCooldownAfterRequest(c, apiKeyService, currentAPIKeyFromContext(c, apiKey))
			return
		}

		// ── 5. 加载订阅（订阅模式时始终加载） ───────────────────────

		// skipBilling: /v1/usage 只需鉴权，跳过所有计费执行
		skipBilling := c.Request.URL.Path == "/v1/usage"

		var subscription *service.UserSubscription
		deferGroupBilling := shouldDeferGroupBilling(c, apiKey)
		isSubscriptionType := apiKey.Group != nil && apiKey.Group.IsSubscriptionType()

		if !deferGroupBilling && isSubscriptionType && subscriptionService != nil {
			sub, subErr := subscriptionService.GetActiveSubscription(
				c.Request.Context(),
				apiKey.User.ID,
				apiKey.Group.ID,
			)
			if subErr != nil {
				if !skipBilling {
					AbortWithError(c, 403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
					return
				}
				// skipBilling: 订阅不存在也放行，handler 会返回可用的数据
			} else {
				subscription = sub
			}
		}

		// ── 6. 计费执行（skipBilling 时整块跳过） ────────────────────

		if !skipBilling {
			// Key 状态检查
			switch apiKey.Status {
			case service.StatusAPIKeyQuotaExhausted:
				AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
				return
			case service.StatusAPIKeyExpired:
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}

			// 运行时过期/配额检查（即使状态是 active，也要检查时间和用量）
			if apiKey.IsExpired() {
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}
			if apiKey.IsQuotaExhausted() {
				AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
				return
			}

			if !deferGroupBilling && !enforceGroupBilling(c, apiKey, subscription, subscriptionService, skipBilling) {
				return
			}
		}

		// ── 7. 设置上下文 → Next ─────────────────────────────────────

		if subscriptionService != nil {
			c.Set(string(ContextKeySubscriptionService), subscriptionService)
		}
		if deferGroupBilling {
			c.Set(string(ContextKeyDeferredGroupBilling), true)
		}
		if subscription != nil {
			c.Set(string(ContextKeySubscription), subscription)
		}
		SetAPIKeyContext(c, apiKey)
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      apiKey.User.ID,
			Concurrency: apiKey.User.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), apiKey.User.Role)
		_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)

		c.Next()
		applyAPIKeyRouteCooldownAfterRequest(c, apiKeyService, currentAPIKeyFromContext(c, apiKey))
	}
}

// SetAPIKeyContext stores the effective API key and updates request-scoped
// group/account-pool context used by downstream services.
func SetAPIKeyContext(c *gin.Context, apiKey *service.APIKey) {
	if c == nil || apiKey == nil {
		return
	}
	c.Set(string(ContextKeyAPIKey), apiKey)
	setAPIKeyAccountPoolContext(c, apiKey)
	setGroupContext(c, apiKey.Group)
}

// GetAPIKeyFromContext 从上下文中获取API key
func GetAPIKeyFromContext(c *gin.Context) (*service.APIKey, bool) {
	value, exists := c.Get(string(ContextKeyAPIKey))
	if !exists {
		return nil, false
	}
	apiKey, ok := value.(*service.APIKey)
	return apiKey, ok
}

// GetSubscriptionFromContext 从上下文中获取订阅信息
func GetSubscriptionFromContext(c *gin.Context) (*service.UserSubscription, bool) {
	value, exists := c.Get(string(ContextKeySubscription))
	if !exists {
		return nil, false
	}
	subscription, ok := value.(*service.UserSubscription)
	return subscription, ok
}

func setSubscriptionContext(c *gin.Context, subscription *service.UserSubscription) {
	if c == nil {
		return
	}
	if subscription == nil {
		c.Set(string(ContextKeySubscription), nil)
		return
	}
	c.Set(string(ContextKeySubscription), subscription)
}

// ResolveAPIKeyForModelRequest re-routes an API key after a handler has parsed
// the request model/body. It updates gin.Context and reloads subscription
// context when the effective group changes.
func ResolveAPIKeyForModelRequest(c *gin.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey, requestedModel string, imageIntent bool) (*service.APIKey, bool) {
	if c == nil || apiKeyService == nil || apiKey == nil || len(apiKey.MultiGroupRoutes) == 0 {
		return apiKey, true
	}
	forcePlatform, _ := GetForcePlatformFromContext(c)
	resolved := apiKeyService.ResolveForModelRequest(c.Request.Context(), apiKey, c.Request.URL.Path, forcePlatform, requestedModel, imageIntent)
	if resolved == nil {
		resolved = apiKey
	}
	if abortIfAPIKeyGroupUnavailable(c, resolved) {
		return nil, false
	}
	SetAPIKeyContext(c, resolved)
	deferredBilling := shouldEnforceDeferredGroupBilling(c)
	if deferredBilling {
		subscriptionService, _ := subscriptionServiceFromContext(c)
		if !enforceGroupBilling(c, resolved, nil, subscriptionService, false) {
			return nil, false
		}
	}
	if !deferredBilling && !refreshSubscriptionContextForResolvedAPIKey(c, resolved) {
		return nil, false
	}
	return resolved, true
}

func refreshSubscriptionContextForResolvedAPIKey(c *gin.Context, apiKey *service.APIKey) bool {
	if c == nil || apiKey == nil || apiKey.Group == nil || !apiKey.Group.IsSubscriptionType() {
		setSubscriptionContext(c, nil)
		return true
	}
	subscriptionService, ok := subscriptionServiceFromContext(c)
	if !ok {
		return true
	}
	userID := apiKey.UserID
	if userID <= 0 && apiKey.User != nil {
		userID = apiKey.User.ID
	}
	subscription, err := subscriptionService.GetActiveSubscription(c.Request.Context(), userID, apiKey.Group.ID)
	if err != nil {
		if c.Request != nil && c.Request.URL != nil && c.Request.URL.Path == "/v1/usage" {
			setSubscriptionContext(c, nil)
			return true
		}
		AbortWithError(c, 403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
		return false
	}
	if needsMaintenance, validateErr := subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group); validateErr != nil {
		code := "SUBSCRIPTION_INVALID"
		status := 403
		if errors.Is(validateErr, service.ErrDailyLimitExceeded) ||
			errors.Is(validateErr, service.ErrWeeklyLimitExceeded) ||
			errors.Is(validateErr, service.ErrMonthlyLimitExceeded) {
			code = "USAGE_LIMIT_EXCEEDED"
			status = 429
		}
		AbortWithError(c, status, code, validateErr.Error())
		return false
	} else if needsMaintenance {
		maintenanceCopy := *subscription
		subscriptionService.DoWindowMaintenance(&maintenanceCopy)
	}
	setSubscriptionContext(c, subscription)
	return true
}

func shouldDeferGroupBilling(c *gin.Context, apiKey *service.APIKey) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil || apiKey == nil || len(apiKey.MultiGroupRoutes) == 0 {
		return false
	}
	if c.Request.URL.Path == "/v1/usage" {
		return false
	}
	return isModelAwareBillingEndpoint(c.Request.URL.Path)
}

func shouldEnforceDeferredGroupBilling(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, exists := c.Get(string(ContextKeyDeferredGroupBilling))
	if !exists {
		return false
	}
	enabled, _ := value.(bool)
	return enabled
}

func isModelAwareBillingEndpoint(path string) bool {
	path = strings.TrimRight(strings.ToLower(strings.TrimSpace(path)), "/")
	switch {
	case path == "/v1/messages":
		return true
	case path == "/v1/messages/count_tokens":
		return true
	case strings.HasPrefix(path, "/v1/responses"), strings.HasPrefix(path, "/responses"), strings.HasPrefix(path, "/backend-api/codex/responses"):
		return true
	case path == "/v1/chat/completions" || path == "/chat/completions":
		return true
	case path == "/v1/embeddings" || path == "/embeddings":
		return true
	case path == "/v1/images/generations" || path == "/images/generations":
		return true
	case path == "/v1/images/edits" || path == "/images/edits":
		return true
	case path == "/v1/midjourney/generations" || path == "/midjourney/generations":
		return true
	case strings.HasPrefix(path, "/v1beta/models/") || strings.HasPrefix(path, "/antigravity/v1beta/models/"):
		return true
	default:
		return false
	}
}

func subscriptionServiceFromContext(c *gin.Context) (*service.SubscriptionService, bool) {
	if c == nil {
		return nil, false
	}
	value, exists := c.Get(string(ContextKeySubscriptionService))
	if !exists {
		return nil, false
	}
	subscriptionService, ok := value.(*service.SubscriptionService)
	return subscriptionService, ok && subscriptionService != nil
}

func enforceGroupBilling(c *gin.Context, apiKey *service.APIKey, subscription *service.UserSubscription, subscriptionService *service.SubscriptionService, skipBilling bool) bool {
	if skipBilling {
		return true
	}
	if apiKey == nil || apiKey.User == nil {
		return true
	}
	if apiKey.Group != nil && apiKey.Group.IsSubscriptionType() && subscriptionService != nil {
		if subscription == nil {
			userID := apiKey.UserID
			if userID <= 0 {
				userID = apiKey.User.ID
			}
			sub, err := subscriptionService.GetActiveSubscription(c.Request.Context(), userID, apiKey.Group.ID)
			if err != nil {
				AbortWithError(c, 403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
				return false
			}
			subscription = sub
		}
		needsMaintenance, validateErr := subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group)
		if validateErr != nil {
			code := "SUBSCRIPTION_INVALID"
			status := 403
			if errors.Is(validateErr, service.ErrDailyLimitExceeded) ||
				errors.Is(validateErr, service.ErrWeeklyLimitExceeded) ||
				errors.Is(validateErr, service.ErrMonthlyLimitExceeded) {
				code = "USAGE_LIMIT_EXCEEDED"
				status = 429
			}
			AbortWithError(c, status, code, validateErr.Error())
			return false
		}
		if needsMaintenance {
			maintenanceCopy := *subscription
			subscriptionService.DoWindowMaintenance(&maintenanceCopy)
		}
		setSubscriptionContext(c, subscription)
		return true
	}
	setSubscriptionContext(c, nil)
	if apiKey.User.Balance <= 0 {
		AbortWithError(c, 403, "INSUFFICIENT_BALANCE", "Insufficient account balance")
		return false
	}
	return true
}

func setGroupContext(c *gin.Context, group *service.Group) {
	if !service.IsGroupContextValid(group) {
		return
	}
	if existing, ok := c.Request.Context().Value(ctxkey.Group).(*service.Group); ok && existing != nil && existing.ID == group.ID && service.IsGroupContextValid(existing) {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.Group, group)
	c.Request = c.Request.WithContext(ctx)
}

func setAPIKeyAccountPoolContext(c *gin.Context, apiKey *service.APIKey) {
	if c == nil || apiKey == nil || c.Request == nil {
		return
	}
	userID := apiKey.UserID
	if userID <= 0 && apiKey.User != nil {
		userID = apiKey.User.ID
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.APIKeyAccountPoolStrategy, service.NormalizeAccountPoolStrategy(apiKey.AccountPoolStrategy))
	ctx = context.WithValue(ctx, ctxkey.APIKeyUserID, userID)
	c.Request = c.Request.WithContext(ctx)
}

func abortIfAPIKeyGroupUnavailable(c *gin.Context, apiKey *service.APIKey) bool {
	code, message, ok := validateAPIKeyGroupAvailable(apiKey)
	if ok {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
	AbortWithError(c, 403, code, message)
	return true
}

func abortIfAPIKeyGroupNotAllowed(c *gin.Context, apiKey *service.APIKey) bool {
	if validateAPIKeyGroupAllowed(apiKey) {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
	AbortWithError(c, 403, "GROUP_NOT_ALLOWED", "API Key 所属专属分组不再允许当前用户使用")
	return true
}

func validateAPIKeyGroupAllowed(apiKey *service.APIKey) bool {
	if apiKey == nil || apiKey.GroupID == nil || apiKey.User == nil || apiKey.Group == nil {
		return true
	}
	group := apiKey.Group
	if group.IsSubscriptionType() {
		return true
	}
	return apiKey.User.CanBindGroup(group.ID, group.IsExclusive)
}

func validateAPIKeyGroupAvailable(apiKey *service.APIKey) (string, string, bool) {
	if apiKey == nil || apiKey.GroupID == nil {
		return "", "", true
	}
	group := apiKey.Group
	if group == nil || strings.EqualFold(group.Status, "deleted") {
		return "GROUP_DELETED", "API Key 所属分组已删除", false
	}
	if !group.IsActive() {
		return "GROUP_DISABLED", "API Key 所属分组已停用", false
	}
	return "", "", true
}

func resolveAPIKeyForRequest(c *gin.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey) *service.APIKey {
	if apiKeyService == nil || apiKey == nil {
		return apiKey
	}
	forcePlatform, _ := GetForcePlatformFromContext(c)
	return apiKeyService.ResolveForRequest(c.Request.Context(), apiKey, c.Request.URL.Path, forcePlatform)
}

func applyAPIKeyRouteCooldownAfterRequest(c *gin.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey) {
	if c == nil || apiKeyService == nil || apiKey == nil || apiKey.GroupID == nil || len(apiKey.MultiGroupRoutes) == 0 {
		return
	}
	groupID := *apiKey.GroupID
	cooldownSeconds, ok := apiKey.RouteCooldownSeconds(groupID)
	if !ok {
		return
	}
	if shouldCooldownAPIKeyRoute(c.Writer.Status()) {
		apiKeyService.MarkRouteGroupCooldown(c.Request.Context(), apiKey, groupID, cooldownSeconds)
		return
	}
	if c.Writer.Status() < http.StatusBadRequest {
		apiKeyService.ClearRouteGroupCooldown(c.Request.Context(), apiKey, groupID)
	}
}

func currentAPIKeyFromContext(c *gin.Context, fallback *service.APIKey) *service.APIKey {
	if current, ok := GetAPIKeyFromContext(c); ok && current != nil {
		return current
	}
	return fallback
}

func shouldCooldownAPIKeyRoute(status int) bool {
	return status == http.StatusTooManyRequests || status == 529 || status >= http.StatusInternalServerError
}
