package handler

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService           *service.SettingService
	apiKeyService            *service.APIKeyService
	notificationEmailService *service.NotificationEmailService
	version                  string
}

// SetAPIKeyService attaches the catalog access service without changing the
// constructor signature used by existing tests.
func (h *SettingHandler) SetAPIKeyService(apiKeyService *service.APIKeyService) {
	h.apiKeyService = apiKeyService
}

// SetNotificationEmailService attaches the public notification email service without
// changing the constructor signature used by existing tests.
func (h *SettingHandler) SetNotificationEmailService(notificationEmailService *service.NotificationEmailService) {
	h.notificationEmailService = notificationEmailService
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PublicSettings{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		PasskeyEnabled:                   settings.PasskeyEnabled,
		LoginAgreementEnabled:            settings.LoginAgreementEnabled,
		LoginAgreementMode:               settings.LoginAgreementMode,
		LoginAgreementUpdatedAt:          settings.LoginAgreementUpdatedAt,
		LoginAgreementRevision:           settings.LoginAgreementRevision,
		LoginAgreementDocuments:          publicLoginAgreementDocumentsToDTO(settings.LoginAgreementDocuments),
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		SupportPopupTitle:                settings.SupportPopupTitle,
		SupportPopupDescription:          settings.SupportPopupDescription,
		SupportPopupFooter:               settings.SupportPopupFooter,
		SupportPopupItems:                dto.ParseSupportPopupItems(settings.SupportPopupItems),
		DocURL:                           settings.DocURL,
		HomeContent:                      settings.HomeContent,
		HideCcsImportButton:              settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:          settings.PurchaseSubscriptionURL,
		TableDefaultPageSize:             settings.TableDefaultPageSize,
		TablePageSizeOptions:             settings.TablePageSizeOptions,
		CustomMenuItems:                  dto.ParseUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  dto.ParseCustomEndpoints(settings.CustomEndpoints),
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		OIDCOAuthEnabled:                 settings.OIDCOAuthEnabled,
		OIDCOAuthProviderName:            settings.OIDCOAuthProviderName,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		PaymentFAQItems:                  dto.PaymentFAQItemsFromService(service.ParsePaymentFAQItems(settings.PaymentFAQItems)),
		Version:                          h.version,
		ServerTimezone:                   timezone.Name(),
		ServerUTCOffset:                  timezone.UTCOffset(),
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled:   settings.AvailableChannelsEnabled,
		ModelPlazaEnabled:          settings.ModelPlazaEnabled,
		ModelPlazaRequireAuth:      settings.ModelPlazaRequireAuth,
		GroupBuyEnabled:            settings.GroupBuyEnabled,
		GroupBuyProductName:        settings.GroupBuyProductName,
		GroupBuyDescription:        settings.GroupBuyDescription,
		PixelCafeEnabled:           settings.PixelCafeEnabled,
		PixelCafeTitle:             settings.PixelCafeTitle,
		PixelCafeDescription:       settings.PixelCafeDescription,
		PixelCafeHeaderVisible:     settings.PixelCafeHeaderVisible,
		PixelCafeWorkstationLayout: settings.PixelCafeWorkstationLayout,

		AffiliateEnabled: settings.AffiliateEnabled,

		AccountShareEnabled:              settings.AccountShareEnabled,
		AccountShareChannelStatusVisible: settings.AccountShareChannelStatusVisible,
		ExternalCapacityReferenceEnabled: settings.ExternalCapacityReferenceEnabled,

		RiskControlEnabled:         settings.RiskControlEnabled,
		AllowUserViewErrorRequests: settings.AllowUserViewErrorRequests,

		WelfareEnabled:               settings.WelfareEnabled,
		WelfareDailyCheckinEnabled:   settings.WelfareDailyCheckinEnabled,
		WelfareRechargeEnabled:       settings.WelfareRechargeEnabled,
		WelfareVIPEnabled:            settings.WelfareVIPEnabled,
		WelfareNewUserTrialEnabled:   settings.WelfareNewUserTrialEnabled,
		LeaderboardMinAccountAgeDays: settings.LeaderboardMinAccountAgeDays,
	})
}

// GetModelMarketCatalog 获取公开模型市场目录。
// GET /api/v1/model-market/catalog
func (h *SettingHandler) GetModelMarketCatalog(c *gin.Context) {
	if h.settingService == nil {
		response.NotFound(c, "Model plaza is not enabled")
		return
	}
	runtime := h.settingService.GetModelPlazaRuntime(c.Request.Context())
	if !runtime.Enabled {
		response.NotFound(c, "Model plaza is not enabled")
		return
	}

	subject, authenticated := servermiddleware.GetAuthSubjectFromContext(c)
	if runtime.RequireAuth && !authenticated {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var allowedExclusive map[int64]struct{}
	if authenticated {
		if h.apiKeyService == nil {
			response.ErrorFrom(c, fmt.Errorf("model market access service is unavailable"))
			return
		}
		var err error
		allowedExclusive, err = h.apiKeyService.GetUserAllowedGroupIDSet(c.Request.Context(), subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}
	catalog, err := h.settingService.GetModelMarketCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, modelMarketPublicCatalog{
		ModelMarketCatalog: filterModelMarketCatalog(catalog, allowedExclusive),
		Description:        runtime.Description,
	})
}

type modelMarketPublicCatalog struct {
	*service.ModelMarketCatalog
	Description string `json:"description,omitempty"`
}

// filterModelMarketCatalog preserves the local catalog's existing unscoped
// pricing cards, while hiding exclusive account groups unless the current user
// is explicitly authorized. A group-scoped card with no visible account group
// is omitted entirely.
func filterModelMarketCatalog(catalog *service.ModelMarketCatalog, allowedExclusive map[int64]struct{}) *service.ModelMarketCatalog {
	if catalog == nil {
		return nil
	}
	filtered := *catalog
	filtered.Groups = make([]service.ModelMarketGroup, 0, len(catalog.Groups))
	for _, group := range catalog.Groups {
		if len(group.SupportedGroupIDs) == 0 {
			filtered.Groups = append(filtered.Groups, group)
			continue
		}
		visible := make([]service.ModelMarketAccountGroup, 0, len(group.SupportedGroups))
		visibleIDs := make([]int64, 0, len(group.SupportedGroups))
		for _, accountGroup := range group.SupportedGroups {
			if accountGroup.Exclusive {
				if allowedExclusive == nil {
					continue
				}
				if _, allowed := allowedExclusive[accountGroup.ID]; !allowed {
					continue
				}
			}
			visible = append(visible, accountGroup)
			visibleIDs = append(visibleIDs, accountGroup.ID)
		}
		if len(visible) == 0 {
			continue
		}
		group.SupportedGroups = visible
		group.SupportedGroupIDs = visibleIDs
		filtered.Groups = append(filtered.Groups, group)
	}
	return &filtered
}

// GetQuickstartTutorialConfig returns the public quick-start tutorial content.
// GET /api/v1/tutorials/quickstart-config
func (h *SettingHandler) GetQuickstartTutorialConfig(c *gin.Context) {
	cfg, err := h.settingService.GetQuickstartTutorialConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UnsubscribeNotificationEmail handles optional notification email opt-outs.
// GET /api/v1/settings/email-unsubscribe?token=...
func (h *SettingHandler) UnsubscribeNotificationEmail(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}
	result, err := h.notificationEmailService.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	body := "<!doctype html><html><head><meta charset=\"utf-8\"><title>Unsubscribed</title></head><body style=\"font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;padding:32px;\"><h1>Unsubscribed</h1><p>You have unsubscribed <strong>" + html.EscapeString(result.Email) + "</strong> from <strong>" + html.EscapeString(result.Event) + "</strong> emails.</p></body></html>"
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(body))
}

func publicLoginAgreementDocumentsToDTO(items []service.LoginAgreementDocument) []dto.LoginAgreementDocument {
	result := make([]dto.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LoginAgreementDocument{
			ID:        item.ID,
			Title:     item.Title,
			ContentMD: item.ContentMD,
		})
	}
	return result
}
