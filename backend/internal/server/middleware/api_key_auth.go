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
// /v1/usage 和 /v1/sub2api/billing 端点只需鉴权，不需要计费执行。
// 前者允许过期/配额耗尽的 Key 查询自身用量，后者用于读取当前 Key 的倍率声明。
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
		// 注意：错误信息只暴露当前客户端 IP，避免暴露具体的 IP 限制机制
		if len(apiKey.IPWhitelist) > 0 || len(apiKey.IPBlacklist) > 0 {
			forwardedIPSettings := cfg.ForwardedClientIPSettings()
			clientIP := ip.GetSecurityClientIPWithHeaders(c, forwardedIPSettings.TrustForwardedIP, forwardedIPSettings.Headers)
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
		if apiKey.IsCafeRoomManaged() && apiKey.PinnedAccountID <= 0 {
			AbortWithError(c, http.StatusForbidden, "CAFE_ACCOUNT_UNAVAILABLE", "the cafe account is temporarily unavailable")
			return
		}
		billingInfoRequest := c.Request.URL.Path == "/v1/sub2api/billing"
		if cfg.RunMode != config.RunModeSimple && c.Request.URL.Path != "/v1/usage" && !billingInfoRequest {
			apiKey = withUnavailableSubscriptionRouteGroups(c.Request.Context(), apiKey, subscriptionService)
		}
		if !billingInfoRequest {
			apiKey = resolveAPIKeyForRequest(c, apiKeyService, apiKey)
			if apiKey == nil {
				AbortWithError(c, http.StatusForbidden, "NO_MATCHING_GROUP_ROUTE", "No available group route matches the request")
				return
			}
		}
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
			if !billingInfoRequest {
				_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			}
			c.Next()
			if !billingInfoRequest {
				applyAPIKeyRouteCooldownAfterRequest(c, apiKeyService, currentAPIKeyFromContext(c, apiKey))
			}
			return
		}

		// ── 5. 加载订阅（订阅模式时始终加载） ───────────────────────

		// 已创建的异步生图任务必须可由原 API Key 查询，即使生成已耗尽余额或额度。
		skipBilling := c.Request.URL.Path == "/v1/usage" || billingInfoRequest || isAsyncImageTaskRead(c.Request.Method, c.Request.URL.Path)

		var subscription *service.UserSubscription
		deferGroupBilling := shouldDeferGroupBilling(c, apiKey)
		isSubscriptionType := apiKey.Group != nil && apiKey.Group.IsSubscriptionType()

		if !deferGroupBilling && isSubscriptionType && subscriptionService != nil && !billingInfoRequest {
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
				abortWithAPIKeyQuotaError(c)
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
				abortWithAPIKeyQuotaError(c)
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
		if !billingInfoRequest {
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
		}

		c.Next()
		if !billingInfoRequest {
			applyAPIKeyRouteCooldownAfterRequest(c, apiKeyService, currentAPIKeyFromContext(c, apiKey))
		}
	}
}

func withUnavailableSubscriptionRouteGroups(ctx context.Context, apiKey *service.APIKey, subscriptionService *service.SubscriptionService) *service.APIKey {
	if apiKey == nil || subscriptionService == nil || len(apiKey.MultiGroupRoutes) == 0 {
		return apiKey
	}
	userID := apiKey.UserID
	if userID <= 0 && apiKey.User != nil {
		userID = apiKey.User.ID
	}
	if userID <= 0 {
		return apiKey
	}

	groups := make(map[int64]*service.Group, len(apiKey.MultiGroupRouteGroups)+1)
	if service.IsGroupContextValid(apiKey.Group) {
		groups[apiKey.Group.ID] = apiKey.Group
	}
	for _, group := range apiKey.MultiGroupRouteGroups {
		if service.IsGroupContextValid(group) {
			groups[group.ID] = group
		}
	}

	groupIDs := make(map[int64]struct{}, len(apiKey.MultiGroupRoutes)+1)
	for _, route := range apiKey.MultiGroupRoutes {
		if route.Enabled && route.GroupID > 0 {
			groupIDs[route.GroupID] = struct{}{}
		}
	}
	if apiKey.GroupID != nil && *apiKey.GroupID > 0 {
		groupIDs[*apiKey.GroupID] = struct{}{}
	}

	unavailable := make(map[int64]struct{})
	for groupID := range groupIDs {
		group := groups[groupID]
		if group == nil || !group.IsActive() || !group.IsSubscriptionType() {
			continue
		}
		subscription, err := subscriptionService.GetActiveSubscription(ctx, userID, groupID)
		if err != nil {
			if isSkippableSubscriptionRouteError(err) {
				unavailable[groupID] = struct{}{}
			}
			continue
		}
		_, err = subscriptionService.ValidateAndCheckLimits(subscription, group)
		if isSkippableSubscriptionRouteError(err) {
			unavailable[groupID] = struct{}{}
		}
	}
	return apiKey.WithUnavailableRouteGroups(unavailable)
}

func isSkippableSubscriptionRouteError(err error) bool {
	return errors.Is(err, service.ErrSubscriptionNotFound) ||
		errors.Is(err, service.ErrSubscriptionExpired) ||
		errors.Is(err, service.ErrSubscriptionSuspended)
}

func abortWithAPIKeyQuotaError(c *gin.Context) {
	const message = "API key 额度已用完"
	if isOpenAIResponsesAPIKeyRequest(c) {
		abortWithOpenAIQuotaError(c, http.StatusTooManyRequests, message)
		return
	}
	AbortWithError(c, http.StatusTooManyRequests, "API_KEY_QUOTA_EXHAUSTED", message)
}

func isOpenAIResponsesAPIKeyRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	path := strings.TrimRight(strings.ToLower(strings.TrimSpace(c.Request.URL.Path)), "/")
	if !isModelAwareBillingEndpoint(path) {
		return false
	}
	for _, root := range []string{
		"/v1/responses",
		"/responses",
		"/backend-api/codex/responses",
	} {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

// SetAPIKeyContext stores the effective API key and updates request-scoped
// group/account-pool context used by downstream services.
func SetAPIKeyContext(c *gin.Context, apiKey *service.APIKey) {
	if c == nil || apiKey == nil {
		return
	}
	c.Set(string(ContextKeyAPIKey), apiKey)
	setAPIKeyAccountPoolContext(c, apiKey)
	if apiKey.PinnedAccountID > 0 {
		ctx := context.WithValue(c.Request.Context(), ctxkey.APIKeyPinnedAccountID, apiKey.PinnedAccountID)
		c.Request = c.Request.WithContext(ctx)
	}
	setGroupContext(c, apiKey.Group)
}

func isAsyncImageTaskRead(method, path string) bool {
	if method != http.MethodGet {
		return false
	}
	return strings.HasPrefix(path, "/v1/images/tasks/") || strings.HasPrefix(path, "/images/tasks/")
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

// ResolveAPIKeyForModelRequest applies model-aware routing for a parsed request.
// Gateway route wrappers and protocol handlers may both call this helper; the
// first successful decision is memoized in gin.Context so breaker leases and
// the effective platform remain stable for the whole request.
func ResolveAPIKeyForModelRequest(c *gin.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey, requestedModel string, imageIntent bool) (*service.APIKey, bool) {
	if c == nil || apiKeyService == nil || apiKey == nil {
		return apiKey, true
	}
	if resolved, ok := cachedAPIKeyModelResolution(c, apiKey, requestedModel); ok {
		return resolved, true
	}
	if !service.IsGroupContextValid(apiKey.Group) && len(apiKey.MultiGroupRoutes) == 0 && apiKey.PinnedAccountID <= 0 {
		// Direct handler callers may carry the legacy, partially hydrated single
		// group snapshot. Authentication owns availability validation; do not let
		// this compatibility case bypass multi-group or pinned route resolution.
		cacheAPIKeyModelResolution(c, apiKey, requestedModel)
		return apiKey, true
	}
	forcePlatform, _ := GetForcePlatformFromContext(c)
	resolved := apiKeyService.ResolveForModelRequest(c.Request.Context(), apiKey, c.Request.URL.Path, forcePlatform, requestedModel, imageIntent)
	if resolved == nil {
		AbortWithError(c, http.StatusForbidden, "NO_MATCHING_GROUP_ROUTE", "No available group route matches the requested model or request type")
		return nil, false
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
	cacheAPIKeyModelResolution(c, resolved, requestedModel)
	return resolved, true
}

const contextKeyAPIKeyModelResolution = "api_key_model_resolution"

type apiKeyModelResolution struct {
	apiKeyID       int64
	apiKey         string
	requestedModel string
	resolved       *service.APIKey
}

// cachedAPIKeyModelResolution returns the request's first successful
// model-aware route selection. Gateway route wrappers and protocol handlers
// both call ResolveAPIKeyForModelRequest after parsing the same body; reusing
// that decision keeps breaker leases and platform dispatch request-scoped.
func cachedAPIKeyModelResolution(c *gin.Context, apiKey *service.APIKey, requestedModel string) (*service.APIKey, bool) {
	if c == nil || apiKey == nil {
		return nil, false
	}
	value, exists := c.Get(contextKeyAPIKeyModelResolution)
	if !exists {
		return nil, false
	}
	entry, ok := value.(apiKeyModelResolution)
	if !ok || entry.resolved == nil {
		return nil, false
	}
	if entry.apiKeyID != apiKey.ID ||
		entry.apiKey != apiKey.Key ||
		entry.requestedModel != strings.TrimSpace(requestedModel) {
		return nil, false
	}
	return entry.resolved, true
}

func cacheAPIKeyModelResolution(c *gin.Context, apiKey *service.APIKey, requestedModel string) {
	if c == nil || apiKey == nil {
		return
	}
	c.Set(contextKeyAPIKeyModelResolution, apiKeyModelResolution{
		apiKeyID:       apiKey.ID,
		apiKey:         apiKey.Key,
		requestedModel: strings.TrimSpace(requestedModel),
		resolved:       apiKey,
	})
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
	if IsAPIKeyRouteCooldownMarked(c) || shouldCooldownAPIKeyRoute(c.Writer.Status()) {
		apiKeyService.MarkRouteGroupCooldown(c.Request.Context(), apiKey, groupID, cooldownSeconds)
	}
	if routeBreakerFailureStatus(c) {
		apiKeyService.RecordAPIKeyRouteBreakerFailure(c.Request.Context(), apiKey)
		return
	}
	if c.Writer.Status() < http.StatusBadRequest {
		apiKeyService.ClearRouteGroupCooldown(c.Request.Context(), apiKey, groupID)
		apiKeyService.RecordAPIKeyRouteBreakerSuccess(c.Request.Context(), apiKey)
		return
	}
	// Business and unclassified client failures must not strand a half-open probe.
	apiKeyService.ReleaseAPIKeyRouteBreakerProbe(c.Request.Context(), apiKey)
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

func routeBreakerFailureStatus(c *gin.Context) bool {
	if c == nil {
		return false
	}
	// A successful final response always clears the selected route, even when
	// an intermediate failover left an upstream status in the Ops context.
	finalStatus := c.Writer.Status()
	markedStatus, marked := APIKeyRouteBreakerMarkedStatus(c)
	if !marked && finalStatus < http.StatusBadRequest {
		return false
	}
	if upstreamValue, exists := c.Get(service.OpsUpstreamStatusCodeKey); exists {
		if upstreamStatus, ok := upstreamValue.(int); ok && upstreamStatus > 0 {
			return shouldCooldownAPIKeyRoute(upstreamStatus)
		}
	}
	if marked {
		return shouldCooldownAPIKeyRoute(markedStatus)
	}
	switch finalStatus {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, 529:
		return true
	default:
		return false
	}
}
