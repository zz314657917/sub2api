package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingPaymentEnabled      = "payment_enabled"
	SettingMinRechargeAmount   = "MIN_RECHARGE_AMOUNT"
	SettingMaxRechargeAmount   = "MAX_RECHARGE_AMOUNT"
	SettingDailyRechargeLimit  = "DAILY_RECHARGE_LIMIT"
	SettingOrderTimeoutMinutes = "ORDER_TIMEOUT_MINUTES"
	SettingMaxPendingOrders    = "MAX_PENDING_ORDERS"
	SettingEnabledPaymentTypes = "ENABLED_PAYMENT_TYPES"
	SettingLoadBalanceStrategy = "LOAD_BALANCE_STRATEGY"
	SettingBalancePayDisabled  = "BALANCE_PAYMENT_DISABLED"
	SettingBalanceRechargeMult = "BALANCE_RECHARGE_MULTIPLIER"
	SettingRechargeFeeRate     = "RECHARGE_FEE_RATE"
	SettingRechargePackages    = "PAYMENT_RECHARGE_PACKAGES"
	SettingPaymentFAQItems     = "PAYMENT_FAQ_ITEMS"
	SettingProductNamePrefix   = "PRODUCT_NAME_PREFIX"
	SettingProductNameSuffix   = "PRODUCT_NAME_SUFFIX"
	SettingHelpImageURL        = "PAYMENT_HELP_IMAGE_URL"
	SettingHelpText            = "PAYMENT_HELP_TEXT"
	SettingCancelRateLimitOn   = "CANCEL_RATE_LIMIT_ENABLED"
	SettingCancelRateLimitMax  = "CANCEL_RATE_LIMIT_MAX"
	SettingCancelWindowSize    = "CANCEL_RATE_LIMIT_WINDOW"
	SettingCancelWindowUnit    = "CANCEL_RATE_LIMIT_UNIT"
	SettingCancelWindowMode    = "CANCEL_RATE_LIMIT_WINDOW_MODE"
)

// Default values for payment configuration settings.
const (
	defaultOrderTimeoutMin   = 30
	defaultMaxPendingOrders  = 3
	defaultRechargePackageID = "pkg-5"
	minRechargePackageAmount = 5.0
)

// PaymentConfig holds the payment system configuration.
type PaymentConfig struct {
	Enabled                   bool              `json:"enabled"`
	MinAmount                 float64           `json:"min_amount"`
	MaxAmount                 float64           `json:"max_amount"`
	DailyLimit                float64           `json:"daily_limit"`
	OrderTimeoutMin           int               `json:"order_timeout_minutes"`
	MaxPendingOrders          int               `json:"max_pending_orders"`
	EnabledTypes              []string          `json:"enabled_payment_types"`
	BalanceDisabled           bool              `json:"balance_disabled"`
	BalanceRechargeMultiplier float64           `json:"balance_recharge_multiplier"`
	RechargeFeeRate           float64           `json:"recharge_fee_rate"`
	RechargePackages          []RechargePackage `json:"recharge_packages"`
	FAQItems                  []PaymentFAQItem  `json:"faq_items"`
	LoadBalanceStrategy       string            `json:"load_balance_strategy"`
	ProductNamePrefix         string            `json:"product_name_prefix"`
	ProductNameSuffix         string            `json:"product_name_suffix"`
	HelpImageURL              string            `json:"help_image_url"`
	HelpText                  string            `json:"help_text"`
	StripePublishableKey      string            `json:"stripe_publishable_key,omitempty"`

	// Cancel rate limit settings
	CancelRateLimitEnabled bool   `json:"cancel_rate_limit_enabled"`
	CancelRateLimitMax     int    `json:"cancel_rate_limit_max"`
	CancelRateLimitWindow  int    `json:"cancel_rate_limit_window"`
	CancelRateLimitUnit    string `json:"cancel_rate_limit_unit"`
	CancelRateLimitMode    string `json:"cancel_rate_limit_window_mode"`
}

// UpdatePaymentConfigRequest contains fields to update payment configuration.
type UpdatePaymentConfigRequest struct {
	Enabled                   *bool             `json:"enabled"`
	MinAmount                 *float64          `json:"min_amount"`
	MaxAmount                 *float64          `json:"max_amount"`
	DailyLimit                *float64          `json:"daily_limit"`
	OrderTimeoutMin           *int              `json:"order_timeout_minutes"`
	MaxPendingOrders          *int              `json:"max_pending_orders"`
	EnabledTypes              []string          `json:"enabled_payment_types"`
	BalanceDisabled           *bool             `json:"balance_disabled"`
	BalanceRechargeMultiplier *float64          `json:"balance_recharge_multiplier"`
	RechargeFeeRate           *float64          `json:"recharge_fee_rate"`
	RechargePackages          []RechargePackage `json:"recharge_packages"`
	FAQItems                  []PaymentFAQItem  `json:"faq_items"`
	LoadBalanceStrategy       *string           `json:"load_balance_strategy"`
	ProductNamePrefix         *string           `json:"product_name_prefix"`
	ProductNameSuffix         *string           `json:"product_name_suffix"`
	HelpImageURL              *string           `json:"help_image_url"`
	HelpText                  *string           `json:"help_text"`

	// Cancel rate limit settings
	CancelRateLimitEnabled *bool   `json:"cancel_rate_limit_enabled"`
	CancelRateLimitMax     *int    `json:"cancel_rate_limit_max"`
	CancelRateLimitWindow  *int    `json:"cancel_rate_limit_window"`
	CancelRateLimitUnit    *string `json:"cancel_rate_limit_unit"`
	CancelRateLimitMode    *string `json:"cancel_rate_limit_window_mode"`

	VisibleMethodAlipaySource  *string `json:"payment_visible_method_alipay_source"`
	VisibleMethodWxpaySource   *string `json:"payment_visible_method_wxpay_source"`
	VisibleMethodAlipayEnabled *bool   `json:"payment_visible_method_alipay_enabled"`
	VisibleMethodWxpayEnabled  *bool   `json:"payment_visible_method_wxpay_enabled"`
}

func (r UpdatePaymentConfigRequest) HasRechargePackages() bool {
	return r.RechargePackages != nil
}

func (r UpdatePaymentConfigRequest) HasFAQItems() bool {
	return r.FAQItems != nil
}

type RechargePackage struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	Enabled        bool    `json:"enabled"`
	PayAmount      float64 `json:"pay_amount"`
	CreditedAmount float64 `json:"credited_amount"`
	SortOrder      int     `json:"sort_order"`
}

type PaymentFAQItem struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type RechargePackageCheckoutView struct {
	ID                      string  `json:"id"`
	Label                   string  `json:"label"`
	PayAmount               float64 `json:"pay_amount"`
	CreditedAmount          float64 `json:"credited_amount"`
	BonusAmount             float64 `json:"bonus_amount"`
	EffectiveCreditedAmount float64 `json:"effective_credited_amount"`
	EffectiveBonusAmount    float64 `json:"effective_bonus_amount"`
	SortOrder               int     `json:"sort_order"`
}

// MethodLimits holds per-payment-type limits.
type MethodLimits struct {
	PaymentType string  `json:"payment_type"`
	Currency    string  `json:"currency"`
	FeeRate     float64 `json:"fee_rate"`
	DailyLimit  float64 `json:"daily_limit"`
	SingleMin   float64 `json:"single_min"`
	SingleMax   float64 `json:"single_max"`
}

// MethodLimitsResponse is the full response for the user-facing /limits API.
// It includes per-method limits and the global widest range (union of all methods).
type MethodLimitsResponse struct {
	Methods   map[string]MethodLimits `json:"methods"`
	GlobalMin float64                 `json:"global_min"` // 0 = no minimum
	GlobalMax float64                 `json:"global_max"` // 0 = no maximum
}

type CreateProviderInstanceRequest struct {
	ProviderKey     string            `json:"provider_key"`
	Name            string            `json:"name"`
	Config          map[string]string `json:"config"`
	SupportedTypes  []string          `json:"supported_types"`
	Enabled         bool              `json:"enabled"`
	PaymentMode     string            `json:"payment_mode"`
	SortOrder       int               `json:"sort_order"`
	Limits          string            `json:"limits"`
	RefundEnabled   bool              `json:"refund_enabled"`
	AllowUserRefund bool              `json:"allow_user_refund"`
}

type UpdateProviderInstanceRequest struct {
	Name            *string           `json:"name"`
	Config          map[string]string `json:"config"`
	SupportedTypes  []string          `json:"supported_types"`
	Enabled         *bool             `json:"enabled"`
	PaymentMode     *string           `json:"payment_mode"`
	SortOrder       *int              `json:"sort_order"`
	Limits          *string           `json:"limits"`
	RefundEnabled   *bool             `json:"refund_enabled"`
	AllowUserRefund *bool             `json:"allow_user_refund"`
}
type CreatePlanRequest struct {
	GroupID       int64    `json:"group_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	ValidityDays  int      `json:"validity_days"`
	ValidityUnit  string   `json:"validity_unit"`
	Features      string   `json:"features"`
	ProductName   string   `json:"product_name"`
	ForSale       bool     `json:"for_sale"`
	SortOrder     int      `json:"sort_order"`
}

type UpdatePlanRequest struct {
	GroupID       *int64   `json:"group_id"`
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Price         *float64 `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	ValidityDays  *int     `json:"validity_days"`
	ValidityUnit  *string  `json:"validity_unit"`
	Features      *string  `json:"features"`
	ProductName   *string  `json:"product_name"`
	ForSale       *bool    `json:"for_sale"`
	SortOrder     *int     `json:"sort_order"`
}

// PaymentConfigService manages payment configuration and CRUD for
// provider instances, channels, and subscription plans.
type PaymentConfigService struct {
	entClient     *dbent.Client
	settingRepo   SettingRepository
	encryptionKey []byte
}

// NewPaymentConfigService creates a new PaymentConfigService.
func NewPaymentConfigService(entClient *dbent.Client, settingRepo SettingRepository, encryptionKey []byte) *PaymentConfigService {
	return &PaymentConfigService{entClient: entClient, settingRepo: settingRepo, encryptionKey: encryptionKey}
}

// IsPaymentEnabled returns whether the payment system is enabled.
func (s *PaymentConfigService) IsPaymentEnabled(ctx context.Context) bool {
	val, err := s.settingRepo.GetValue(ctx, SettingPaymentEnabled)
	if err != nil {
		return false
	}
	return val == "true"
}

// GetPaymentConfig returns the full payment configuration.
func (s *PaymentConfigService) GetPaymentConfig(ctx context.Context) (*PaymentConfig, error) {
	keys := []string{
		SettingPaymentEnabled, SettingMinRechargeAmount, SettingMaxRechargeAmount,
		SettingDailyRechargeLimit, SettingOrderTimeoutMinutes, SettingMaxPendingOrders,
		SettingEnabledPaymentTypes, SettingBalancePayDisabled, SettingBalanceRechargeMult, SettingRechargeFeeRate, SettingLoadBalanceStrategy,
		SettingRechargePackages, SettingPaymentFAQItems, SettingProductNamePrefix, SettingProductNameSuffix,
		SettingHelpImageURL, SettingHelpText,
		SettingCancelRateLimitOn, SettingCancelRateLimitMax,
		SettingCancelWindowSize, SettingCancelWindowUnit, SettingCancelWindowMode,
		SettingPaymentVisibleMethodAlipayEnabled, SettingPaymentVisibleMethodAlipaySource,
		SettingPaymentVisibleMethodWxpayEnabled, SettingPaymentVisibleMethodWxpaySource,
	}
	vals, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get payment config settings: %w", err)
	}
	cfg := s.parsePaymentConfig(vals)
	// Load Stripe publishable key from the first enabled Stripe provider instance
	cfg.StripePublishableKey = s.getStripePublishableKey(ctx)
	return cfg, nil
}

func (s *PaymentConfigService) parsePaymentConfig(vals map[string]string) *PaymentConfig {
	cfg := &PaymentConfig{
		Enabled:                   vals[SettingPaymentEnabled] == "true",
		MinAmount:                 pcParseFloat(vals[SettingMinRechargeAmount], minRechargePackageAmount),
		MaxAmount:                 pcParseFloat(vals[SettingMaxRechargeAmount], 0),
		DailyLimit:                pcParseFloat(vals[SettingDailyRechargeLimit], 0),
		OrderTimeoutMin:           pcParseInt(vals[SettingOrderTimeoutMinutes], defaultOrderTimeoutMin),
		MaxPendingOrders:          pcParseInt(vals[SettingMaxPendingOrders], defaultMaxPendingOrders),
		BalanceDisabled:           vals[SettingBalancePayDisabled] == "true",
		BalanceRechargeMultiplier: normalizeBalanceRechargeMultiplier(pcParseFloat(vals[SettingBalanceRechargeMult], defaultBalanceRechargeMultiplier)),
		RechargeFeeRate:           pcParseFloat(vals[SettingRechargeFeeRate], 0),
		RechargePackages:          parseRechargePackages(vals[SettingRechargePackages]),
		FAQItems:                  ParsePaymentFAQItems(vals[SettingPaymentFAQItems]),
		LoadBalanceStrategy:       vals[SettingLoadBalanceStrategy],
		ProductNamePrefix:         vals[SettingProductNamePrefix],
		ProductNameSuffix:         vals[SettingProductNameSuffix],
		HelpImageURL:              vals[SettingHelpImageURL],
		HelpText:                  vals[SettingHelpText],

		CancelRateLimitEnabled: vals[SettingCancelRateLimitOn] == "true",
		CancelRateLimitMax:     pcParseInt(vals[SettingCancelRateLimitMax], 10),
		CancelRateLimitWindow:  pcParseInt(vals[SettingCancelWindowSize], 1),
		CancelRateLimitUnit:    vals[SettingCancelWindowUnit],
		CancelRateLimitMode:    vals[SettingCancelWindowMode],
	}
	if cfg.LoadBalanceStrategy == "" {
		cfg.LoadBalanceStrategy = payment.DefaultLoadBalanceStrategy
	}
	if raw := vals[SettingEnabledPaymentTypes]; raw != "" {
		types := make([]string, 0, len(strings.Split(raw, ",")))
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
		cfg.EnabledTypes = NormalizeVisibleMethods(types)
	}
	return cfg
}

func DefaultRechargePackages() []RechargePackage {
	return []RechargePackage{
		{
			ID:             defaultRechargePackageID,
			Label:          "5",
			Enabled:        true,
			PayAmount:      minRechargePackageAmount,
			CreditedAmount: minRechargePackageAmount,
			SortOrder:      10,
		},
	}
}

func defaultRechargePackagesJSON() string {
	raw, err := json.Marshal(DefaultRechargePackages())
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func DefaultPaymentFAQItems() []PaymentFAQItem {
	return []PaymentFAQItem{
		{
			Title: "额度与计费规则",
			Body:  "灵活积分按实际调用消耗；订阅套餐按配置的周期额度和倍率计费，具体以当前站点配置为准。",
		},
		{
			Title: "灵活积分说明",
			Body:  "购买后的积分进入账户积分，在积分用完前持续有效，可用于未被套餐覆盖的调用。",
		},
		{
			Title: "大量使用是否可以联系管理员获得额外折扣？",
			Body:  "如果您需要大量使用我们的服务，可以联系我们的管理员团队获取企业级定制解决方案和额外的折扣优惠。",
		},
		{
			Title: "如何升级套餐？",
			Body:  "选择更高档套餐并完成支付后，系统会按当前订阅规则刷新可用额度和有效期。",
		},
		{
			Title: "额度恢复机制",
			Body:  "订阅额度按套餐周期自动重置；灵活积分不会周期清零，只随调用扣减。",
		},
		{
			Title: "套餐变更说明",
			Body:  "同一订阅分组再次购买通常视为续费或延长，具体生效方式由后台套餐配置决定。",
		},
		{
			Title: "订阅额度与灵活积分",
			Body:  "优先使用订阅套餐覆盖的额度；超出或未覆盖部分可继续使用灵活积分支付。",
		},
	}
}

func defaultPaymentFAQItemsJSON() string {
	raw, err := json.Marshal(DefaultPaymentFAQItems())
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func ParsePaymentFAQItems(raw string) []PaymentFAQItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultPaymentFAQItems()
	}
	var items []PaymentFAQItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return DefaultPaymentFAQItems()
	}
	normalized := NormalizePaymentFAQItems(items)
	if len(normalized) == 0 {
		return DefaultPaymentFAQItems()
	}
	return normalized
}

func NormalizePaymentFAQItems(items []PaymentFAQItem) []PaymentFAQItem {
	normalized := make([]PaymentFAQItem, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		body := strings.TrimSpace(item.Body)
		if title == "" || body == "" {
			continue
		}
		normalized = append(normalized, PaymentFAQItem{
			Title: title,
			Body:  body,
		})
	}
	return normalized
}

func parseRechargePackages(raw string) []RechargePackage {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultRechargePackages()
	}
	var packages []RechargePackage
	if err := json.Unmarshal([]byte(raw), &packages); err != nil {
		return DefaultRechargePackages()
	}
	normalized, err := NormalizeRechargePackages(packages)
	if err != nil || len(normalized) == 0 {
		return DefaultRechargePackages()
	}
	return normalized
}

func NormalizeRechargePackages(packages []RechargePackage) ([]RechargePackage, error) {
	normalized := make([]RechargePackage, 0, len(packages))
	seen := make(map[string]bool, len(packages))
	enabledCount := 0
	for idx, pkg := range packages {
		pkg.ID = strings.TrimSpace(pkg.ID)
		if pkg.ID == "" {
			pkg.ID = fmt.Sprintf("pkg-%d", idx+1)
		}
		if seen[pkg.ID] {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PACKAGE", "duplicate recharge package id").
				WithMetadata(map[string]string{"id": pkg.ID})
		}
		seen[pkg.ID] = true

		pkg.PayAmount = normalizePaymentPackageAmount(pkg.PayAmount)
		pkg.CreditedAmount = normalizePaymentPackageAmount(pkg.CreditedAmount)
		if pkg.PayAmount < minRechargePackageAmount {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PACKAGE", "recharge package pay amount must be at least 5").
				WithMetadata(map[string]string{"id": pkg.ID, "min": fmt.Sprintf("%.2f", minRechargePackageAmount)})
		}
		if pkg.CreditedAmount < pkg.PayAmount {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PACKAGE", "recharge package credited amount cannot be less than pay amount").
				WithMetadata(map[string]string{"id": pkg.ID})
		}
		if strings.TrimSpace(pkg.Label) == "" {
			pkg.Label = formatRechargePackageLabel(pkg.PayAmount)
		} else {
			pkg.Label = strings.TrimSpace(pkg.Label)
		}
		if pkg.SortOrder == 0 {
			pkg.SortOrder = (idx + 1) * 10
		}
		if pkg.Enabled {
			enabledCount++
		}
		normalized = append(normalized, pkg)
	}
	if len(normalized) == 0 || enabledCount == 0 {
		return nil, infraerrors.BadRequest("INVALID_RECHARGE_PACKAGE", "at least one enabled recharge package is required")
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].SortOrder == normalized[j].SortOrder {
			return normalized[i].PayAmount < normalized[j].PayAmount
		}
		return normalized[i].SortOrder < normalized[j].SortOrder
	})
	return normalized, nil
}

func normalizePaymentPackageAmount(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	return math.Round(value*100) / 100
}

func formatRechargePackageLabel(amount float64) string {
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

// getStripePublishableKey finds the publishable key from the first enabled Stripe provider instance.
func (s *PaymentConfigService) getStripePublishableKey(ctx context.Context) string {
	if s.entClient == nil {
		return ""
	}
	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(
			paymentproviderinstance.EnabledEQ(true),
			paymentproviderinstance.ProviderKeyEQ(payment.TypeStripe),
		).Limit(1).All(ctx)
	if err != nil || len(instances) == 0 {
		return ""
	}
	cfg, err := s.decryptConfig(instances[0].Config)
	if err != nil || cfg == nil {
		return ""
	}
	return cfg[payment.ConfigKeyPublishableKey]
}

// UpdatePaymentConfig updates the payment configuration settings.
// NOTE: This function exceeds 30 lines because each field requires an independent
// nil-check before serialisation — this is inherent to patch-style update patterns
// and cannot be meaningfully decomposed without introducing unnecessary abstraction.
func (s *PaymentConfigService) UpdatePaymentConfig(ctx context.Context, req UpdatePaymentConfigRequest) error {
	if req.BalanceRechargeMultiplier != nil {
		if math.IsNaN(*req.BalanceRechargeMultiplier) || math.IsInf(*req.BalanceRechargeMultiplier, 0) || *req.BalanceRechargeMultiplier <= 0 {
			return infraerrors.BadRequest("INVALID_BALANCE_RECHARGE_MULTIPLIER", "balance recharge multiplier must be greater than 0")
		}
	}
	if req.RechargeFeeRate != nil {
		v := *req.RechargeFeeRate
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 100 {
			return infraerrors.BadRequest("INVALID_RECHARGE_FEE_RATE", "recharge fee rate must be between 0 and 100")
		}
		// Enforce max 2 decimal places
		if math.Round(v*100) != v*100 {
			return infraerrors.BadRequest("INVALID_RECHARGE_FEE_RATE", "recharge fee rate allows at most 2 decimal places")
		}
	}
	var rechargePackagesRaw string
	if req.HasRechargePackages() {
		packages, err := NormalizeRechargePackages(req.RechargePackages)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(packages)
		if err != nil {
			return fmt.Errorf("marshal recharge packages: %w", err)
		}
		rechargePackagesRaw = string(raw)
	}
	var faqItemsRaw string
	if req.HasFAQItems() {
		items := NormalizePaymentFAQItems(req.FAQItems)
		if len(items) == 0 {
			items = DefaultPaymentFAQItems()
		}
		raw, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("marshal payment faq items: %w", err)
		}
		faqItemsRaw = string(raw)
	}
	m := map[string]string{
		SettingPaymentEnabled:                    formatBoolOrEmpty(req.Enabled),
		SettingMinRechargeAmount:                 formatPositiveFloat(req.MinAmount),
		SettingMaxRechargeAmount:                 formatPositiveFloat(req.MaxAmount),
		SettingDailyRechargeLimit:                formatPositiveFloat(req.DailyLimit),
		SettingOrderTimeoutMinutes:               formatPositiveInt(req.OrderTimeoutMin),
		SettingMaxPendingOrders:                  formatPositiveInt(req.MaxPendingOrders),
		SettingBalancePayDisabled:                formatBoolOrEmpty(req.BalanceDisabled),
		SettingBalanceRechargeMult:               formatPositiveFloat(req.BalanceRechargeMultiplier),
		SettingRechargeFeeRate:                   formatNonNegativeFloat(req.RechargeFeeRate),
		SettingLoadBalanceStrategy:               derefStr(req.LoadBalanceStrategy),
		SettingProductNamePrefix:                 derefStr(req.ProductNamePrefix),
		SettingProductNameSuffix:                 derefStr(req.ProductNameSuffix),
		SettingHelpImageURL:                      derefStr(req.HelpImageURL),
		SettingHelpText:                          derefStr(req.HelpText),
		SettingCancelRateLimitOn:                 formatBoolOrEmpty(req.CancelRateLimitEnabled),
		SettingCancelRateLimitMax:                formatPositiveInt(req.CancelRateLimitMax),
		SettingCancelWindowSize:                  formatPositiveInt(req.CancelRateLimitWindow),
		SettingCancelWindowUnit:                  derefStr(req.CancelRateLimitUnit),
		SettingCancelWindowMode:                  derefStr(req.CancelRateLimitMode),
		SettingPaymentVisibleMethodAlipaySource:  derefStr(req.VisibleMethodAlipaySource),
		SettingPaymentVisibleMethodWxpaySource:   derefStr(req.VisibleMethodWxpaySource),
		SettingPaymentVisibleMethodAlipayEnabled: formatBoolOrEmpty(req.VisibleMethodAlipayEnabled),
		SettingPaymentVisibleMethodWxpayEnabled:  formatBoolOrEmpty(req.VisibleMethodWxpayEnabled),
	}
	if req.HasRechargePackages() {
		m[SettingRechargePackages] = rechargePackagesRaw
	}
	if req.HasFAQItems() {
		m[SettingPaymentFAQItems] = faqItemsRaw
	}
	if req.EnabledTypes != nil {
		m[SettingEnabledPaymentTypes] = strings.Join(req.EnabledTypes, ",")
	} else {
		m[SettingEnabledPaymentTypes] = ""
	}
	return s.settingRepo.SetMultiple(ctx, m)
}

func formatBoolOrEmpty(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

func formatPositiveFloat(v *float64) string {
	if v == nil || *v <= 0 {
		return "" // empty → parsePaymentConfig uses default
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func formatNonNegativeFloat(v *float64) string {
	if v == nil || *v < 0 {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func formatPositiveInt(v *int) string {
	if v == nil || *v <= 0 {
		return ""
	}
	return strconv.Itoa(*v)
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func splitTypes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func joinTypes(types []string) string {
	return strings.Join(types, ",")
}

func pcParseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func pcParseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func buildVisibleMethodSourceAvailability(instances []*dbent.PaymentProviderInstance) map[string]bool {
	available := make(map[string]bool, 4)
	for _, inst := range instances {
		switch inst.ProviderKey {
		case payment.TypeAlipay:
			if inst.SupportedTypes == "" || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeAlipay) || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeAlipayDirect) {
				available[VisibleMethodSourceOfficialAlipay] = true
			}
		case payment.TypeWxpay:
			if inst.SupportedTypes == "" || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeWxpay) || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeWxpayDirect) {
				available[VisibleMethodSourceOfficialWechat] = true
			}
		case payment.TypeEasyPay:
			for _, supportedType := range splitTypes(inst.SupportedTypes) {
				switch NormalizeVisibleMethod(supportedType) {
				case payment.TypeAlipay:
					available[VisibleMethodSourceEasyPayAlipay] = true
				case payment.TypeWxpay:
					available[VisibleMethodSourceEasyPayWechat] = true
				}
			}
		}
	}
	return available
}

func applyVisibleMethodRoutingToEnabledTypes(base []string, vals map[string]string, available map[string]bool) []string {
	shouldExpose := map[string]bool{
		payment.TypeAlipay: visibleMethodShouldBeExposed(payment.TypeAlipay, vals, available),
		payment.TypeWxpay:  visibleMethodShouldBeExposed(payment.TypeWxpay, vals, available),
	}

	seen := make(map[string]struct{}, len(base)+2)
	out := make([]string, 0, len(base)+2)
	appendType := func(paymentType string) {
		paymentType = NormalizeVisibleMethod(paymentType)
		if paymentType == "" {
			return
		}
		if _, ok := seen[paymentType]; ok {
			return
		}
		seen[paymentType] = struct{}{}
		out = append(out, paymentType)
	}

	for _, paymentType := range base {
		visibleMethod := NormalizeVisibleMethod(paymentType)
		switch visibleMethod {
		case payment.TypeAlipay, payment.TypeWxpay:
			if shouldExpose[visibleMethod] {
				appendType(visibleMethod)
			}
		default:
			appendType(visibleMethod)
		}
	}

	for _, visibleMethod := range []string{payment.TypeAlipay, payment.TypeWxpay} {
		if shouldExpose[visibleMethod] {
			appendType(visibleMethod)
		}
	}
	return out
}

func visibleMethodShouldBeExposed(method string, vals map[string]string, available map[string]bool) bool {
	enabledKey := visibleMethodEnabledSettingKey(method)
	sourceKey := visibleMethodSourceSettingKey(method)
	if enabledKey == "" || sourceKey == "" || vals[enabledKey] != "true" {
		return false
	}
	source := NormalizeVisibleMethodSource(method, vals[sourceKey])
	return source != "" && available[source]
}
