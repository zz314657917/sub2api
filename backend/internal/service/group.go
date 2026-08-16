package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

type OpenAIMessagesDispatchModelConfig = domain.OpenAIMessagesDispatchModelConfig
type GroupModelsListConfig = domain.GroupModelsListConfig

type Group struct {
	ID             int64
	Name           string
	Description    string
	Platform       string
	RateMultiplier float64
	// 高峰时段倍率：peak_rate_enabled 为 true 且当前时刻处于 [PeakStart, PeakEnd) 时，
	// token 计费倍率额外乘以 PeakRateMultiplier。详见 PeakMultiplierAt。
	PeakRateEnabled    bool
	PeakStart          string
	PeakEnd            string
	PeakRateMultiplier float64
	IsExclusive        bool
	Status             string
	Hydrated           bool // indicates the group was loaded from a trusted repository source
	// DuplicateOperationID is internal persistence metadata used only to recover
	// an already committed one-click copy. It must never be mapped to API DTOs.
	DuplicateOperationID string

	SubscriptionType          string
	RoutingScope              string
	DailyLimitUSD             *float64
	WeeklyLimitUSD            *float64
	MonthlyLimitUSD           *float64
	DefaultValidityDays       int
	LongContextPricingEnabled bool
	ModelPricing              []ChannelModelPricing

	// 图片生成计费配置（antigravity 和 gemini 平台使用）
	AllowImageGeneration bool
	ImageRateIndependent bool
	ImageRateMultiplier  float64
	ImagePrice1K         *float64
	ImagePrice2K         *float64
	ImagePrice4K         *float64

	// Claude Code 客户端限制
	ClaudeCodeOnly  bool
	FallbackGroupID *int64
	// 无效请求兜底分组（仅 anthropic 平台使用）
	FallbackGroupIDOnInvalidRequest *int64

	// 模型路由配置
	// key: 模型匹配模式（支持 * 通配符，如 "claude-opus-*"）
	// value: 优先账号 ID 列表
	ModelRouting        map[string][]int64
	ModelRoutingEnabled bool

	// ModelMatchPatterns are administrator-owned request model rules for API-key
	// group routing. "*" explicitly permits every model.
	ModelMatchPatterns []string

	// MCP XML 协议注入开关（仅 antigravity 平台使用）
	MCPXMLInject bool

	// 支持的模型系列（仅 antigravity 平台使用）
	// 可选值: claude, gemini_text, gemini_image
	SupportedModelScopes []string

	// 分组排序
	SortOrder int

	// OpenAI Messages 调度配置（仅 openai 平台使用）
	AllowMessagesDispatch       bool
	AllowLive                   bool
	RequireOAuthOnly            bool // 仅允许非 apikey 类型账号关联（OpenAI/Antigravity/Anthropic/Gemini）
	RequirePrivacySet           bool // 调度时仅允许 privacy 已成功设置的账号（OpenAI/Antigravity/Anthropic/Gemini）
	DefaultMappedModel          string
	MessagesDispatchModelConfig OpenAIMessagesDispatchModelConfig
	ModelsListConfig            GroupModelsListConfig

	// RPMLimit 分组级每分钟请求数上限（0 = 不限制）。
	// 一旦设置即接管该分组用户的限流（覆盖用户级 rpm_limit），可被 user-group rpm_override 进一步覆盖。
	RPMLimit int

	CreatedAt time.Time
	UpdatedAt time.Time

	AccountGroups           []AccountGroup
	AccountCount            int64
	ActiveAccountCount      int64
	RateLimitedAccountCount int64
}

func (g *Group) IsActive() bool {
	return g.Status == StatusActive
}

func (g *Group) IsSubscriptionType() bool {
	return g.SubscriptionType == SubscriptionTypeSubscription
}

func (g *Group) EffectiveRoutingScope() string {
	if g == nil {
		return GroupRoutingScopeInference
	}
	return NormalizeGroupRoutingScope(g.RoutingScope, g.AllowImageGeneration)
}

func (g *Group) AllowsRoutingScope(scope string) bool {
	if g == nil {
		return true
	}
	return g.EffectiveRoutingScope() == NormalizeGroupRoutingScope(scope, false)
}

// NormalizeGroupModelMatchPatterns canonicalizes administrator-owned model
// rules. Empty input is retained as an empty slice so migration preflight can
// identify groups that still need configuration.
func NormalizeGroupModelMatchPatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(patterns))
	seen := make(map[string]struct{}, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if _, exists := seen[pattern]; exists {
			continue
		}
		seen[pattern] = struct{}{}
		out = append(out, pattern)
	}
	sort.Strings(out)
	return out
}

// ValidateGroupModelMatchPatterns normalizes rules and rejects an empty rule
// set at the administrator write boundary.
func ValidateGroupModelMatchPatterns(patterns []string) ([]string, error) {
	normalized := NormalizeGroupModelMatchPatterns(patterns)
	if len(normalized) == 0 {
		return nil, ErrGroupModelMatchPatternsRequired
	}
	return normalized, nil
}

// MatchesModel reports whether this group allows a requested model. An empty
// group rule set is intentionally a non-match; it is a migration blocker.
func (g *Group) MatchesModel(requestedModel string) bool {
	if g == nil {
		return false
	}
	model := strings.ToLower(strings.TrimSpace(requestedModel))
	for _, pattern := range g.ModelMatchPatterns {
		if matchModelPatternAny(pattern, model) {
			return true
		}
	}
	return false
}

func matchModelPatternAny(pattern, model string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == model
	}
	parts := strings.Split(pattern, "*")
	remaining := model
	if first := parts[0]; first != "" {
		if !strings.HasPrefix(remaining, first) {
			return false
		}
		remaining = remaining[len(first):]
	}
	for _, part := range parts[1 : len(parts)-1] {
		if part == "" {
			continue
		}
		idx := strings.Index(remaining, part)
		if idx < 0 {
			return false
		}
		remaining = remaining[idx+len(part):]
	}
	last := parts[len(parts)-1]
	return last == "" || strings.HasSuffix(remaining, last)
}

func (g *Group) HasDailyLimit() bool {
	return g.DailyLimitUSD != nil && *g.DailyLimitUSD > 0
}

func (g *Group) HasWeeklyLimit() bool {
	return g.WeeklyLimitUSD != nil && *g.WeeklyLimitUSD > 0
}

func (g *Group) HasMonthlyLimit() bool {
	return g.MonthlyLimitUSD != nil && *g.MonthlyLimitUSD > 0
}

// GetImagePrice 根据 image_size 返回对应的图片生成价格
// 如果分组未配置价格，返回 nil（调用方应使用默认值）
func (g *Group) GetImagePrice(imageSize string) *float64 {
	switch imageSize {
	case "1K":
		return g.ImagePrice1K
	case "2K":
		return g.ImagePrice2K
	case "4K":
		return g.ImagePrice4K
	default:
		// 未知尺寸默认按 2K 计费
		return g.ImagePrice2K
	}
}

// IsGroupContextValid reports whether a group from context has the fields required for routing decisions.
func IsGroupContextValid(group *Group) bool {
	if group == nil {
		return false
	}
	if group.ID <= 0 {
		return false
	}
	if !group.Hydrated {
		return false
	}
	if group.Platform == "" || group.Status == "" {
		return false
	}
	return true
}

func NormalizeGroupRoutingScope(scope string, allowImageGeneration bool) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case GroupRoutingScopeInference, "text", "chat", "reasoning", "reason", "completion", "completions":
		return GroupRoutingScopeInference
	case GroupRoutingScopeImage, "images", "image_generation", "image-generation":
		return GroupRoutingScopeImage
	case GroupRoutingScopeVideo, "videos", "video_generation", "video-generation":
		return GroupRoutingScopeVideo
	case GroupRoutingScopeEmbedding, "embeddings", "embed":
		return GroupRoutingScopeEmbedding
	default:
		if allowImageGeneration {
			return GroupRoutingScopeImage
		}
		return GroupRoutingScopeInference
	}
}

// GetRoutingAccountIDs 根据请求模型获取路由账号 ID 列表
// 返回匹配的优先账号 ID 列表，如果没有匹配规则则返回 nil
func (g *Group) GetRoutingAccountIDs(requestedModel string) []int64 {
	if !g.ModelRoutingEnabled || len(g.ModelRouting) == 0 || requestedModel == "" {
		return nil
	}

	// 1. 精确匹配优先
	if accountIDs, ok := g.ModelRouting[requestedModel]; ok && len(accountIDs) > 0 {
		return accountIDs
	}

	// 2. 通配符匹配（前缀匹配）
	for pattern, accountIDs := range g.ModelRouting {
		if matchModelPattern(pattern, requestedModel) && len(accountIDs) > 0 {
			return accountIDs
		}
	}

	return nil
}

// matchModelPattern 检查模型是否匹配模式
// 支持 * 通配符，如 "claude-opus-*" 匹配 "claude-opus-4-20250514"
func matchModelPattern(pattern, model string) bool {
	if pattern == model {
		return true
	}

	// 处理 * 通配符（仅支持末尾通配符）
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(model, prefix)
	}

	return false
}

// parseMinutes 把 "HH:MM" 解析为当日分钟数（0..1439），格式非法返回 (0,false)。
// 手工解析而非 time.Parse：本函数位于每请求的计费热路径（PeakMultiplierAt），
// 避免对静态配置字符串重复走 layout 解析与 time.Time 分配。
// 接受集与 time.Parse("15:04", s) 完全一致（存量数据按旧解析写入，不得收窄）：
// 小时 1–2 位数字（0..23，允许不补零如 "1:30"），分钟固定 2 位数字（00..59）。
func parseMinutes(hhmm string) (int, bool) {
	colon := strings.IndexByte(hhmm, ':')
	if (colon != 1 && colon != 2) || len(hhmm)-colon-1 != 2 {
		return 0, false
	}
	h := 0
	for i := 0; i < colon; i++ {
		d := hhmm[i] - '0'
		if d > 9 {
			return 0, false
		}
		h = h*10 + int(d)
	}
	m1, m2 := hhmm[colon+1]-'0', hhmm[colon+2]-'0'
	if m1 > 9 || m2 > 9 {
		return 0, false
	}
	m := int(m1)*10 + int(m2)
	if h > 23 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// PeakMultiplierAt 返回指定时刻 now 的分时计费因子。
//   - 未启用 / 未配置 / 配置非法（start>=end 或格式错误） / 非配置时段 → 返回 1.0（安全降级）
//   - 区间为左闭右开 [PeakStart, PeakEnd)，仅支持当日区间，不支持跨天（如 22:00-次日02:00）
//   - 时刻基于全局系统时区（timezone.Location）判定
//
// 该方法是纯函数，不读取任何外部状态，便于单测。
func (g *Group) PeakMultiplierAt(now time.Time) float64 {
	if g == nil || !g.PeakRateEnabled || g.PeakStart == "" || g.PeakEnd == "" {
		return 1.0
	}
	if !isFinitePeakRateMultiplier(g.PeakRateMultiplier) || g.PeakRateMultiplier < 0 ||
		(g.SubscriptionType != SubscriptionTypeSubscription && g.PeakRateMultiplier == 0) {
		return 1.0
	}
	start, ok1 := parseMinutes(g.PeakStart)
	end, ok2 := parseMinutes(g.PeakEnd)
	if !ok1 || !ok2 || start >= end {
		return 1.0
	}
	t := now.In(timezone.Location())
	cur := t.Hour()*60 + t.Minute()
	if cur >= start && cur < end {
		return g.PeakRateMultiplier
	}
	return 1.0
}

func isFinitePeakRateMultiplier(multiplier float64) bool {
	return !math.IsNaN(multiplier) && !math.IsInf(multiplier, 0)
}

// ValidatePeakRateConfig 是分时倍率配置的唯一校验来源，供 handler 与 service 层共用。
// enabled=true 时要求 start/end 合法且 end>start（不支持跨天）。标准分组 multiplier>0，
// 订阅分组保留 multiplier=0 的兼容行为，可用于免费时段。
// enabled=false 时放行（不关心类型）。subscriptionType 为空按 standard 处理。
func ValidatePeakRateConfig(subscriptionType string, enabled bool, start, end string, multiplier float64) error {
	if !enabled {
		return nil
	}
	if start == "" || end == "" {
		return errors.New("peak_rate_enabled 为 true 时 peak_start 与 peak_end 必填")
	}
	st, okStart := parseMinutes(start)
	if !okStart {
		return fmt.Errorf("peak_start 格式应为 HH:MM，got %q", start)
	}
	en, okEnd := parseMinutes(end)
	if !okEnd {
		return fmt.Errorf("peak_end 格式应为 HH:MM，got %q", end)
	}
	if st >= en {
		return errors.New("peak_end 必须大于 peak_start（不支持跨天区间，如 22:00-02:00）")
	}
	if !isFinitePeakRateMultiplier(multiplier) {
		return errors.New("peak_rate_multiplier 必须是有限数")
	}
	if multiplier < 0 {
		return errors.New("peak_rate_multiplier 不能为负")
	}
	if subscriptionType != SubscriptionTypeSubscription && multiplier == 0 {
		return errors.New("标准分组的 peak_rate_multiplier 必须大于 0")
	}
	return nil
}

// NormalizePeakRateConfig 归一化最终落库的分时配置，CreateGroup 与 UpdateGroup 两条写路径共用（唯一收口）：
//   - 标准和订阅分组都保留合法配置，便于临时停用后再启用；
//   - 关闭时清掉非法/跨天窗口，并按分组类型修复非法倍率，避免脏数据入库。
//
// 与 ValidatePeakRateConfig 的分工：enabled=true 时校验已保证各字段合法，本函数为无操作；
// enabled=false 时校验放行，由本函数兜底清洗。调用顺序为先归一化、后校验。
func NormalizePeakRateConfig(subscriptionType string, enabled bool, start, end string, multiplier float64) (bool, string, string, float64) {
	if !enabled {
		startMinutes, startOK := parseMinutes(start)
		endMinutes, endOK := parseMinutes(end)
		if !startOK || !endOK || startMinutes >= endMinutes {
			start, end = "", ""
		}
		if !isFinitePeakRateMultiplier(multiplier) || multiplier < 0 ||
			(subscriptionType != SubscriptionTypeSubscription && multiplier == 0) {
			multiplier = 1.0
		}
	}
	return enabled, start, end, multiplier
}

func billingReferenceTime(requestStartedAt time.Time) time.Time {
	if requestStartedAt.IsZero() {
		return timezone.Now()
	}
	return requestStartedAt
}

// computePeakAwareMultipliers 把"基础 token 倍率 base"（已含系统/分组/用户级倍率，但不含高峰）
// 拆分为最终 token 倍率与图片按次倍率：图片按次倍率基于 base 现算、不受高峰影响；token 倍率在 base 上叠加高峰因子。
// gateway_service.recordUsageCore 与 openai_gateway_service.RecordUsage 共用此函数，
// 锁死"高峰因子只乘入 token 倍率、图片按次倍率不受影响"这一叠加顺序——任何调换都会被 group_peak_rate_test 覆盖。
func computePeakAwareMultipliers(apiKey *APIKey, base float64, now time.Time) (text, image float64) {
	image = resolveImageRateMultiplier(apiKey, base)
	peak := 1.0
	if apiKey != nil && apiKey.Group != nil {
		peak = apiKey.Group.PeakMultiplierAt(now)
	}
	text = base * peak
	return
}

func usageRateMultiplier(mode string, imageCount int, token, perRequest, image float64) float64 {
	if imageCount > 0 && (mode == string(BillingModeImage) || mode == string(BillingModePerRequest)) {
		return image
	}
	if mode == string(BillingModePerRequest) || mode == string(BillingModeImage) {
		return perRequest
	}
	return token
}
