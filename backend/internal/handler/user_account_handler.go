package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserAccountHandler struct {
	userAccountService *service.UserAccountService
	accountUsage       *service.AccountUsageService
	accountTest        *service.AccountTestService
	oauthService       *service.OAuthService
	openaiOAuth        *service.OpenAIOAuthService
	geminiOAuth        *service.GeminiOAuthService
	antigravityOAuth   *service.AntigravityOAuthService
}

func NewUserAccountHandler(
	userAccountService *service.UserAccountService,
	accountUsage *service.AccountUsageService,
	accountTest *service.AccountTestService,
	oauthService *service.OAuthService,
	openaiOAuth *service.OpenAIOAuthService,
	geminiOAuth *service.GeminiOAuthService,
	antigravityOAuth *service.AntigravityOAuthService,
) *UserAccountHandler {
	return &UserAccountHandler{
		userAccountService: userAccountService,
		accountUsage:       accountUsage,
		accountTest:        accountTest,
		oauthService:       oauthService,
		openaiOAuth:        openaiOAuth,
		geminiOAuth:        geminiOAuth,
		antigravityOAuth:   antigravityOAuth,
	}
}

type userAccountCreateRequest struct {
	Name               string         `json:"name" binding:"required"`
	Notes              *string        `json:"notes"`
	Platform           string         `json:"platform" binding:"required"`
	Type               string         `json:"type" binding:"required,oneof=oauth setup-token apikey upstream"`
	Credentials        map[string]any `json:"credentials" binding:"required"`
	Extra              map[string]any `json:"extra"`
	ExpiresAt          *int64         `json:"expires_at"`
	AutoPauseOnExpired *bool          `json:"auto_pause_on_expired"`
}

type userAccountUpdateRequest struct {
	Name               *string        `json:"name"`
	Notes              *string        `json:"notes"`
	Credentials        map[string]any `json:"credentials"`
	Extra              map[string]any `json:"extra"`
	ExpiresAt          *int64         `json:"expires_at"`
	AutoPauseOnExpired *bool          `json:"auto_pause_on_expired"`
}

type userAccountShareModeRequest struct {
	ShareMode string `json:"share_mode" binding:"required,oneof=private public"`
}

type userAccountImportRequest struct {
	Format             string         `json:"format"`
	Name               string         `json:"name"`
	Notes              *string        `json:"notes"`
	Platform           string         `json:"platform" binding:"required"`
	Type               string         `json:"type" binding:"required,oneof=oauth setup-token apikey upstream"`
	Credentials        map[string]any `json:"credentials" binding:"required"`
	Extra              map[string]any `json:"extra"`
	ExpiresAt          *int64         `json:"expires_at"`
	AutoPauseOnExpired *bool          `json:"auto_pause_on_expired"`
}

type userAccountGenerateAuthURLRequest struct {
	Platform    string `json:"platform"`
	Method      string `json:"method"`
	RedirectURI string `json:"redirect_uri"`
	ProjectID   string `json:"project_id"`
	OAuthType   string `json:"oauth_type"`
	TierID      string `json:"tier_id"`
}

type userAccountExchangeCodeRequest struct {
	Platform    string  `json:"platform" binding:"required"`
	Method      string  `json:"method"`
	SessionID   string  `json:"session_id" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	State       string  `json:"state"`
	RedirectURI string  `json:"redirect_uri"`
	OAuthType   string  `json:"oauth_type"`
	TierID      string  `json:"tier_id"`
	Name        string  `json:"name"`
	Notes       *string `json:"notes"`
}

type userAccountSessionImportRequest struct {
	Platform   string  `json:"platform" binding:"required"`
	Method     string  `json:"method"`
	SessionKey string  `json:"session_key"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Notes      *string `json:"notes"`
}

func (h *UserAccountHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	accounts, result, err := h.userAccountService.List(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.Account, 0, len(accounts))
	for i := range accounts {
		out = append(out, *dto.AccountFromService(&accounts[i]))
	}
	response.Paginated(c, out, result.Total, result.Page, result.PageSize)
}

func (h *UserAccountHandler) GetByID(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	account, err := h.userAccountService.GetByID(c.Request.Context(), subject.UserID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *UserAccountHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req userAccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	svcReq := service.CreateAccountRequest{
		Name:               strings.TrimSpace(req.Name),
		Notes:              req.Notes,
		Platform:           strings.ToLower(strings.TrimSpace(req.Platform)),
		Type:               strings.ToLower(strings.TrimSpace(req.Type)),
		Credentials:        req.Credentials,
		Extra:              req.Extra,
		ExpiresAt:          unixSecondsToTime(req.ExpiresAt),
		AutoPauseOnExpired: req.AutoPauseOnExpired,
	}
	executeUserIdempotentJSON(c, "user.accounts.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		account, err := h.userAccountService.Create(ctx, subject.UserID, svcReq)
		if err != nil {
			return nil, err
		}
		return dto.AccountFromService(account), nil
	})
}

func (h *UserAccountHandler) Import(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req userAccountImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = defaultUserAccountName(req.Platform, req.Type, req.Credentials)
	}
	svcReq := service.CreateAccountRequest{
		Name:               name,
		Notes:              req.Notes,
		Platform:           strings.ToLower(strings.TrimSpace(req.Platform)),
		Type:               strings.ToLower(strings.TrimSpace(req.Type)),
		Credentials:        req.Credentials,
		Extra:              req.Extra,
		ExpiresAt:          unixSecondsToTime(req.ExpiresAt),
		AutoPauseOnExpired: req.AutoPauseOnExpired,
	}
	executeUserIdempotentJSON(c, "user.accounts.import", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		account, err := h.userAccountService.Create(ctx, subject.UserID, svcReq)
		if err != nil {
			return nil, err
		}
		return dto.AccountFromService(account), nil
	})
}

func (h *UserAccountHandler) Update(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	var req userAccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	svcReq := service.UpdateAccountRequest{
		Name:               trimmedStringPtr(req.Name),
		Notes:              req.Notes,
		ExpiresAt:          unixSecondsToTime(req.ExpiresAt),
		AutoPauseOnExpired: req.AutoPauseOnExpired,
	}
	if req.Credentials != nil {
		svcReq.Credentials = &req.Credentials
	}
	if req.Extra != nil {
		svcReq.Extra = &req.Extra
	}
	account, err := h.userAccountService.UpdateWithShareTransition(c.Request.Context(), subject.UserID, accountID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *UserAccountHandler) Delete(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	if err := h.userAccountService.Delete(c.Request.Context(), subject.UserID, accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Account deleted successfully"})
}

func (h *UserAccountHandler) UpdateShareMode(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	var req userAccountShareModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	account, err := h.userAccountService.UpdateShareMode(c.Request.Context(), subject.UserID, accountID, req.ShareMode)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

func (h *UserAccountHandler) Test(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	if _, err := h.userAccountService.GetByID(c.Request.Context(), subject.UserID, accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if h.accountTest == nil {
		response.InternalError(c, "Account test service not configured")
		return
	}
	var req struct {
		ModelID string `json:"model_id"`
		Prompt  string `json:"prompt"`
		Mode    string `json:"mode"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.accountTest.TestAccountConnection(c, accountID, req.ModelID, req.Prompt, req.Mode); err != nil {
		return
	}
}

func (h *UserAccountHandler) GetUsage(c *gin.Context) {
	subject, accountID, ok := h.requireSubjectAndAccountID(c)
	if !ok {
		return
	}
	if _, err := h.userAccountService.GetByID(c.Request.Context(), subject.UserID, accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if h.accountUsage == nil {
		response.InternalError(c, "Account usage service not configured")
		return
	}
	source := c.DefaultQuery("source", "active")
	var (
		usage *service.UsageInfo
		err   error
	)
	if source == "passive" {
		usage, err = h.accountUsage.GetPassiveUsage(c.Request.Context(), accountID)
	} else {
		usage, err = h.accountUsage.GetUsage(c.Request.Context(), accountID)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, usage)
}

func (h *UserAccountHandler) GetShareSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	summary, err := h.userAccountService.GetShareSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func (h *UserAccountHandler) GetUsageSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	startTime, endTime, ok := parseUserAccountUsageSummaryRange(c)
	if !ok {
		return
	}
	summary, err := h.userAccountService.GetUsageSummary(c.Request.Context(), subject.UserID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func (h *UserAccountHandler) GetCapacityPools(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	pools, err := h.userAccountService.GetCapacityPools(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, pools)
}

func (h *UserAccountHandler) TransferShareToBalance(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	executeUserIdempotentJSON(c, "user.accounts.share.transfer", gin.H{"user_id": subject.UserID}, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		transferred, balance, err := h.userAccountService.TransferAvailableShareToBalance(ctx, subject.UserID)
		if err != nil {
			return nil, err
		}
		return gin.H{
			"transferred_amount": transferred,
			"balance":            balance,
		}, nil
	})
}

func parseUserAccountUsageSummaryRange(c *gin.Context) (time.Time, time.Time, bool) {
	userTZ := c.Query("timezone")
	now := timezone.NowInUserLocation(userTZ)
	startDate := strings.TrimSpace(c.Query("start_date"))
	endDate := strings.TrimSpace(c.Query("end_date"))

	startTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -6), userTZ)
	endTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)

	if startDate != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date")
			return time.Time{}, time.Time{}, false
		}
		startTime = timezone.StartOfDayInUserLocation(parsed, userTZ)
	}
	if endDate != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date")
			return time.Time{}, time.Time{}, false
		}
		endTime = timezone.StartOfDayInUserLocation(parsed.AddDate(0, 0, 1), userTZ)
	}
	if !endTime.After(startTime) {
		response.BadRequest(c, "end_date must be after or equal to start_date")
		return time.Time{}, time.Time{}, false
	}
	return startTime, endTime, true
}

func (h *UserAccountHandler) GenerateAuthURL(c *gin.Context) {
	if h.userAccountService != nil && !h.userAccountService.IsEnabled(c.Request.Context()) {
		response.ErrorFrom(c, service.ErrUserAccountShareDisabled)
		return
	}
	var req userAccountGenerateAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = userAccountGenerateAuthURLRequest{}
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	method := strings.ToLower(strings.TrimSpace(req.Method))
	switch platform {
	case service.PlatformOpenAI:
		if h.openaiOAuth == nil {
			response.InternalError(c, "OpenAI OAuth service not configured")
			return
		}
		result, err := h.openaiOAuth.GenerateAuthURL(c.Request.Context(), nil, req.RedirectURI, service.PlatformOpenAI)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, result)
	case service.PlatformAnthropic:
		if h.oauthService == nil {
			response.InternalError(c, "Claude OAuth service not configured")
			return
		}
		var (
			result *service.GenerateAuthURLResult
			err    error
		)
		if method == "setup-token" || method == service.AccountTypeSetupToken {
			result, err = h.oauthService.GenerateSetupTokenURL(c.Request.Context(), nil)
		} else {
			result, err = h.oauthService.GenerateAuthURL(c.Request.Context(), nil)
		}
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, result)
	case service.PlatformGemini:
		if h.geminiOAuth == nil {
			response.InternalError(c, "Gemini OAuth service not configured")
			return
		}
		oauthType := strings.TrimSpace(req.OAuthType)
		if oauthType == "" {
			oauthType = "code_assist"
		}
		redirectURI := req.RedirectURI
		if redirectURI == "" {
			redirectURI = deriveUserAccountGeminiRedirectURI(c)
		}
		result, err := h.geminiOAuth.GenerateAuthURL(c.Request.Context(), nil, redirectURI, req.ProjectID, oauthType, req.TierID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, result)
	case service.PlatformAntigravity:
		if h.antigravityOAuth == nil {
			response.InternalError(c, "Antigravity OAuth service not configured")
			return
		}
		result, err := h.antigravityOAuth.GenerateAuthURL(c.Request.Context(), nil)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, result)
	default:
		response.BadRequest(c, "unsupported platform")
	}
}

func (h *UserAccountHandler) ExchangeCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req userAccountExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	method := strings.ToLower(strings.TrimSpace(req.Method))
	executeUserIdempotentJSON(c, "user.accounts.oauth.exchange", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		account, err := h.createAccountFromOAuthExchange(ctx, subject.UserID, platform, method, req)
		if err != nil {
			return nil, err
		}
		return dto.AccountFromService(account), nil
	})
}

func (h *UserAccountHandler) ImportSession(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req userAccountSessionImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	executeUserIdempotentJSON(c, "user.accounts.session.import", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		account, err := h.createAccountFromSessionKey(ctx, subject.UserID, req)
		if err != nil {
			return nil, err
		}
		return dto.AccountFromService(account), nil
	})
}

func (h *UserAccountHandler) createAccountFromOAuthExchange(ctx context.Context, userID int64, platform, method string, req userAccountExchangeCodeRequest) (*service.Account, error) {
	var svcReq service.CreateAccountRequest
	switch platform {
	case service.PlatformOpenAI:
		tokenInfo, err := h.openaiOAuth.ExchangeCode(ctx, &service.OpenAIExchangeCodeInput{
			SessionID:   req.SessionID,
			Code:        req.Code,
			State:       req.State,
			RedirectURI: req.RedirectURI,
		})
		if err != nil {
			return nil, err
		}
		extra := map[string]any{}
		if tokenInfo.PrivacyMode != "" {
			extra["privacy_mode"] = tokenInfo.PrivacyMode
		}
		svcReq = service.CreateAccountRequest{
			Name:        defaultName(req.Name, tokenInfo.Email, "OpenAI OAuth Account"),
			Notes:       req.Notes,
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Credentials: h.openaiOAuth.BuildAccountCredentials(tokenInfo),
			Extra:       emptyMapToNil(extra),
		}
	case service.PlatformAnthropic:
		tokenInfo, err := h.oauthService.ExchangeCode(ctx, &service.ExchangeCodeInput{
			SessionID: req.SessionID,
			Code:      req.Code,
		})
		if err != nil {
			return nil, err
		}
		accountType := service.AccountTypeOAuth
		if method == service.AccountTypeSetupToken || strings.Contains(tokenInfo.Scope, "inference") && !strings.Contains(tokenInfo.Scope, "profile") {
			accountType = service.AccountTypeSetupToken
		}
		svcReq = service.CreateAccountRequest{
			Name:        defaultName(req.Name, tokenInfo.EmailAddress, "Claude OAuth Account"),
			Notes:       req.Notes,
			Platform:    service.PlatformAnthropic,
			Type:        accountType,
			Credentials: service.BuildClaudeAccountCredentials(tokenInfo),
		}
	case service.PlatformGemini:
		oauthType := strings.TrimSpace(req.OAuthType)
		if oauthType == "" {
			oauthType = "code_assist"
		}
		tokenInfo, err := h.geminiOAuth.ExchangeCode(ctx, &service.GeminiExchangeCodeInput{
			SessionID: req.SessionID,
			State:     req.State,
			Code:      req.Code,
			OAuthType: oauthType,
			TierID:    req.TierID,
		})
		if err != nil {
			return nil, err
		}
		svcReq = service.CreateAccountRequest{
			Name:        defaultName(req.Name, tokenInfo.ProjectID, "Gemini OAuth Account"),
			Notes:       req.Notes,
			Platform:    service.PlatformGemini,
			Type:        service.AccountTypeOAuth,
			Credentials: h.geminiOAuth.BuildAccountCredentials(tokenInfo),
			Extra:       tokenInfo.Extra,
		}
	case service.PlatformAntigravity:
		tokenInfo, err := h.antigravityOAuth.ExchangeCode(ctx, &service.AntigravityExchangeCodeInput{
			SessionID: req.SessionID,
			State:     req.State,
			Code:      req.Code,
		})
		if err != nil {
			return nil, err
		}
		extra := map[string]any{}
		if tokenInfo.PrivacyMode != "" {
			extra["privacy_mode"] = tokenInfo.PrivacyMode
		}
		svcReq = service.CreateAccountRequest{
			Name:        defaultName(req.Name, tokenInfo.Email, "Antigravity OAuth Account"),
			Notes:       req.Notes,
			Platform:    service.PlatformAntigravity,
			Type:        service.AccountTypeOAuth,
			Credentials: h.antigravityOAuth.BuildAccountCredentials(tokenInfo),
			Extra:       emptyMapToNil(extra),
		}
	default:
		return nil, service.ErrUserAccountShareInvalid
	}
	return h.userAccountService.Create(ctx, userID, svcReq)
}

func (h *UserAccountHandler) createAccountFromSessionKey(ctx context.Context, userID int64, req userAccountSessionImportRequest) (*service.Account, error) {
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	method := strings.ToLower(strings.TrimSpace(req.Method))
	sessionKey := strings.TrimSpace(req.SessionKey)
	if sessionKey == "" {
		sessionKey = strings.TrimSpace(req.Code)
	}
	if platform != service.PlatformAnthropic || sessionKey == "" {
		return nil, service.ErrUserAccountShareInvalid
	}
	scope := "full"
	accountType := service.AccountTypeOAuth
	if method == service.AccountTypeSetupToken || method == "session-key-setup-token" {
		scope = "inference"
		accountType = service.AccountTypeSetupToken
	}
	tokenInfo, err := h.oauthService.CookieAuth(ctx, &service.CookieAuthInput{
		SessionKey: sessionKey,
		Scope:      scope,
	})
	if err != nil {
		return nil, err
	}
	svcReq := service.CreateAccountRequest{
		Name:        defaultName(req.Name, tokenInfo.EmailAddress, "Claude Session Account"),
		Notes:       req.Notes,
		Platform:    service.PlatformAnthropic,
		Type:        accountType,
		Credentials: service.BuildClaudeAccountCredentials(tokenInfo),
	}
	return h.userAccountService.Create(ctx, userID, svcReq)
}

func (h *UserAccountHandler) requireSubjectAndAccountID(c *gin.Context) (middleware2.AuthSubject, int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return middleware2.AuthSubject{}, 0, false
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return middleware2.AuthSubject{}, 0, false
	}
	return subject, accountID, true
}

func unixSecondsToTime(value *int64) *time.Time {
	if value == nil || *value <= 0 {
		return nil
	}
	t := time.Unix(*value, 0)
	return &t
}

func trimmedStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func defaultName(primary, fallback, final string) string {
	if s := strings.TrimSpace(primary); s != "" {
		return s
	}
	if s := strings.TrimSpace(fallback); s != "" {
		return s
	}
	return final
}

func defaultUserAccountName(platform, accountType string, credentials map[string]any) string {
	for _, key := range []string{"email", "email_address", "project_id", "name"} {
		if value, ok := credentials[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return strings.TrimSpace(platform) + " " + strings.TrimSpace(accountType) + " Account"
}

func emptyMapToNil(value map[string]any) map[string]any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func deriveUserAccountGeminiRedirectURI(c *gin.Context) string {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin != "" {
		return strings.TrimRight(origin, "/") + "/auth/callback"
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if xfProto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); xfProto != "" {
		scheme = strings.TrimSpace(strings.Split(xfProto, ",")[0])
	}
	host := strings.TrimSpace(c.Request.Host)
	if xfHost := strings.TrimSpace(c.GetHeader("X-Forwarded-Host")); xfHost != "" {
		host = strings.TrimSpace(strings.Split(xfHost, ",")[0])
	}
	return scheme + "://" + host + "/auth/callback"
}
