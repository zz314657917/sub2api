package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	quickstartTutorialConfigVersion = 1
	quickstartTutorialMaxBytes      = 24 * 1024
	quickstartTutorialMaxPlatforms  = 4
	quickstartTutorialMaxErrors     = 16
	quickstartTutorialMaxTiles      = 6
)

// QuickstartTutorialConfig is public, plain-text content for the tutorial
// landing page. It intentionally contains no HTML, links, or credentials.
type QuickstartTutorialConfig struct {
	Version         int                          `json:"version"`
	Header          QuickstartTutorialHeader     `json:"header"`
	Platforms       []QuickstartTutorialPlatform `json:"platforms"`
	Facts           QuickstartTutorialFacts      `json:"facts"`
	Desktop         QuickstartTutorialDesktop    `json:"desktop"`
	API             QuickstartTutorialSection    `json:"api"`
	APIHint         string                       `json:"api_hint"`
	Troubleshooting QuickstartTutorialSection    `json:"troubleshooting"`
	Errors          []QuickstartTutorialError    `json:"errors"`
}

type QuickstartTutorialHeader struct {
	Kicker               string `json:"kicker"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	LibraryActionLabel   string `json:"library_action_label"`
	KeysActionLabel      string `json:"keys_action_label"`
	PlatformControlLabel string `json:"platform_control_label"`
	TerminalControlLabel string `json:"terminal_control_label"`
}

type QuickstartTutorialPlatform struct {
	ID                 string `json:"id"`
	Label              string `json:"label"`
	ClientName         string `json:"client_name"`
	BaseURL            string `json:"base_url"`
	BaseURLDescription string `json:"base_url_description"`
	AuthHint           string `json:"auth_hint"`
	Protocol           string `json:"protocol"`
	ModelHint          string `json:"model_hint"`
}

type QuickstartTutorialFacts struct {
	BaseURLLabel        string `json:"base_url_label"`
	AuthLabel           string `json:"auth_label"`
	AuthDescription     string `json:"auth_description"`
	ProtocolLabel       string `json:"protocol_label"`
	ProtocolDescription string `json:"protocol_description"`
	ModelLabel          string `json:"model_label"`
	ModelDescription    string `json:"model_description"`
}

type QuickstartTutorialSection struct {
	Kicker      string `json:"kicker"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type QuickstartTutorialDesktop struct {
	QuickstartTutorialSection
	Tiles []QuickstartTutorialDesktopTile `json:"tiles"`
}

type QuickstartTutorialDesktopTile struct {
	Number      string `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type QuickstartTutorialError struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func DefaultQuickstartTutorialConfig() *QuickstartTutorialConfig {
	return &QuickstartTutorialConfig{
		Version: quickstartTutorialConfigVersion,
		Header: QuickstartTutorialHeader{
			Kicker:               "QUICK START",
			Title:                "使用文档",
			Description:          "选择模型平台和终端环境，按步骤完成第一次接入。",
			LibraryActionLabel:   "查看完整教程",
			KeysActionLabel:      "查看密钥",
			PlatformControlLabel: "模型平台",
			TerminalControlLabel: "系统 / 终端",
		},
		Platforms: []QuickstartTutorialPlatform{
			{
				ID:                 "codex",
				Label:              "ChatGPT / Codex",
				ClientName:         "ChatGPT / Codex",
				BaseURL:            "https://ai.3zapi.top",
				BaseURLDescription: "ChatGPT / Codex 使用根地址，无需追加 /v1。",
				AuthHint:           "Bearer / OPENAI_API_KEY",
				Protocol:           "OpenAI Responses",
				ModelHint:          "Codex 模型列表",
			},
			{
				ID:                 "claude",
				Label:              "Claude",
				ClientName:         "Claude",
				BaseURL:            "https://ai.3zapi.top",
				BaseURLDescription: "Claude 兼容工具使用根地址。",
				AuthHint:           "Bearer / ANTHROPIC_AUTH_TOKEN",
				Protocol:           "Anthropic Messages",
				ModelHint:          "Claude 模型列表",
			},
		},
		Facts: QuickstartTutorialFacts{
			BaseURLLabel:        "Base URL",
			AuthLabel:           "鉴权",
			AuthDescription:     "密钥来自控制台的 API 密钥页面。",
			ProtocolLabel:       "协议",
			ProtocolDescription: "端点必须和客户端的协议模式一致。",
			ModelLabel:          "模型",
			ModelDescription:    "将示例模型替换为当前账号可用的模型 ID。",
		},
		Desktop: QuickstartTutorialDesktop{
			QuickstartTutorialSection: QuickstartTutorialSection{
				Kicker:      "DESKTOP",
				Title:       "接入桌面端",
				Description: "桌面端、CLI 和配置文件共用同一套 Base URL 与 API Key。",
			},
			Tiles: []QuickstartTutorialDesktopTile{
				{Number: "01", Title: "完成 CLI 配置", Description: "先按上方步骤写入 Base URL 和鉴权。"},
				{Number: "02", Title: "复用 auth.json", Description: "同一份本地密钥可以供 CLI、桌面端和兼容插件使用。"},
				{Number: "03", Title: "启动桌面端", Description: "打开项目后发送简单问题，确认模型响应和额度均正常。"},
			},
		},
		API: QuickstartTutorialSection{
			Kicker:      "API",
			Title:       "在你自己的程序里调用 API",
			Description: "使用当前平台对应的兼容协议发起一次最小请求。",
		},
		APIHint: "多轮对话请按实际上游能力决定是否使用 `store` 与 `previous_response_id`。",
		Troubleshooting: QuickstartTutorialSection{
			Kicker:      "TROUBLESHOOTING",
			Title:       "常见错误码",
			Description: "先确认鉴权、模型、余额和请求协议，再查看服务端返回信息。",
		},
		Errors: []QuickstartTutorialError{
			{Code: "400", Title: "请求格式错误", Description: "检查 JSON、模型参数和所使用的 API 协议是否匹配。"},
			{Code: "401", Title: "密钥无效", Description: "确认 API Key 完整、未过期，并使用正确的鉴权字段。"},
			{Code: "402", Title: "余额或权益不足", Description: "补充余额、兑换权益码或确认当前账号仍有有效套餐。"},
			{Code: "403", Title: "无模型或分组权限", Description: "当前密钥可能未绑定可用分组，或模型不在允许列表。"},
			{Code: "404", Title: "端点或模型不存在", Description: "确认 Claude 使用 Messages，Codex 使用 Responses，模型以列表为准。"},
			{Code: "429", Title: "请求过快", Description: "降低并发，等待 Retry-After 后重试，避免重复提交。"},
			{Code: "CAPACITY", Title: "官方算力不足", Description: "Selected model is at capacity. Please try a different model. 请切换其他模型后重试。"},
			{Code: "5xx", Title: "链路异常", Description: "系统会自动回落；持续出现时记录请求 ID 并联系管理员。"},
		},
	}
}

func (s *SettingService) GetQuickstartTutorialConfig(ctx context.Context) (*QuickstartTutorialConfig, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultQuickstartTutorialConfig(), nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyQuickstartTutorialConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return s.defaultQuickstartTutorialConfig(ctx), nil
		}
		return nil, fmt.Errorf("get quickstart tutorial config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return s.defaultQuickstartTutorialConfig(ctx), nil
	}
	return ParseQuickstartTutorialConfig(raw)
}

func (s *SettingService) SetQuickstartTutorialConfig(ctx context.Context, cfg *QuickstartTutorialConfig) (*QuickstartTutorialConfig, error) {
	normalized, err := NormalizeQuickstartTutorialConfig(cfg)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal quickstart tutorial config: %w", err)
	}
	if len(data) > quickstartTutorialMaxBytes {
		return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_CONFIG_TOO_LARGE", "quickstart tutorial config is too large")
	}
	if s == nil || s.settingRepo == nil {
		return nil, fmt.Errorf("setting repository is unavailable")
	}
	if err := s.settingRepo.Set(ctx, SettingKeyQuickstartTutorialConfig, string(data)); err != nil {
		return nil, fmt.Errorf("set quickstart tutorial config: %w", err)
	}
	s.notifyUpdate()
	return normalized, nil
}

func (s *SettingService) ResetQuickstartTutorialConfig(ctx context.Context) (*QuickstartTutorialConfig, error) {
	return s.SetQuickstartTutorialConfig(ctx, s.defaultQuickstartTutorialConfig(ctx))
}

// defaultQuickstartTutorialConfig uses the site-wide API endpoint when one is
// already configured, so a new deployment does not repeat the built-in sample
// domain before an administrator opens the quick-start editor for the first time.
func (s *SettingService) defaultQuickstartTutorialConfig(ctx context.Context) *QuickstartTutorialConfig {
	cfg := DefaultQuickstartTutorialConfig()
	if s == nil || s.settingRepo == nil {
		return cfg
	}
	apiBaseURL, err := s.settingRepo.GetValue(ctx, SettingKeyAPIBaseURL)
	if err != nil {
		return cfg
	}
	claudeBaseURL, codexBaseURL, ok := quickstartBaseURLsFromAPIBaseURL(apiBaseURL)
	if !ok {
		return cfg
	}
	for i := range cfg.Platforms {
		switch cfg.Platforms[i].ID {
		case "claude":
			cfg.Platforms[i].BaseURL = claudeBaseURL
		case "codex":
			cfg.Platforms[i].BaseURL = codexBaseURL
		}
	}
	return cfg
}

func quickstartBaseURLsFromAPIBaseURL(raw string) (claudeBaseURL, codexBaseURL string, ok bool) {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", "", false
	}
	if strings.HasSuffix(parsed.Path, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	return baseURL, baseURL, baseURL != ""
}

func ParseQuickstartTutorialConfig(raw string) (*QuickstartTutorialConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return DefaultQuickstartTutorialConfig(), nil
	}
	var cfg QuickstartTutorialConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_CONFIG_INVALID_JSON", "quickstart tutorial config JSON is invalid")
	}
	return NormalizeQuickstartTutorialConfig(&cfg)
}

func NormalizeQuickstartTutorialConfig(cfg *QuickstartTutorialConfig) (*QuickstartTutorialConfig, error) {
	if cfg == nil {
		return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_CONFIG_EMPTY", "quickstart tutorial config is required")
	}
	out := *cfg
	out.Version = quickstartTutorialConfigVersion
	if err := normalizeQuickstartHeader(&out.Header); err != nil {
		return nil, err
	}
	if len(out.Platforms) == 0 || len(out.Platforms) > quickstartTutorialMaxPlatforms {
		return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_PLATFORMS_INVALID", "quickstart tutorial requires between 1 and 4 platforms")
	}
	platforms := make([]QuickstartTutorialPlatform, 0, len(out.Platforms))
	platformIDs := make(map[string]struct{}, len(out.Platforms))
	for _, platform := range out.Platforms {
		normalized, err := normalizeQuickstartPlatform(platform)
		if err != nil {
			return nil, err
		}
		if _, exists := platformIDs[normalized.ID]; exists {
			return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_PLATFORM_DUPLICATE", "quickstart tutorial platform id must be unique")
		}
		platformIDs[normalized.ID] = struct{}{}
		platforms = append(platforms, normalized)
	}
	out.Platforms = platforms
	if err := normalizeQuickstartFacts(&out.Facts); err != nil {
		return nil, err
	}
	if err := normalizeQuickstartDesktop(&out.Desktop); err != nil {
		return nil, err
	}
	if err := normalizeQuickstartSection(&out.API, "api"); err != nil {
		return nil, err
	}
	if err := normalizeQuickstartText(&out.APIHint, "api_hint", 500); err != nil {
		return nil, err
	}
	if err := normalizeQuickstartSection(&out.Troubleshooting, "troubleshooting"); err != nil {
		return nil, err
	}
	if len(out.Errors) == 0 || len(out.Errors) > quickstartTutorialMaxErrors {
		return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_ERRORS_INVALID", "quickstart tutorial requires between 1 and 16 troubleshooting entries")
	}
	errorsOut := make([]QuickstartTutorialError, 0, len(out.Errors))
	errorCodes := make(map[string]struct{}, len(out.Errors))
	for _, item := range out.Errors {
		item.Code = strings.TrimSpace(item.Code)
		if err := validateQuickstartText(item.Code, "error code", 32); err != nil {
			return nil, err
		}
		if _, exists := errorCodes[item.Code]; exists {
			return nil, infraerrors.BadRequest("QUICKSTART_TUTORIAL_ERROR_DUPLICATE", "quickstart tutorial error code must be unique")
		}
		errorCodes[item.Code] = struct{}{}
		if err := normalizeQuickstartText(&item.Title, "error title", 96); err != nil {
			return nil, err
		}
		if err := normalizeQuickstartText(&item.Description, "error description", 500); err != nil {
			return nil, err
		}
		errorsOut = append(errorsOut, item)
	}
	out.Errors = errorsOut
	return &out, nil
}

func normalizeQuickstartHeader(header *QuickstartTutorialHeader) error {
	if header == nil {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_HEADER_INVALID", "quickstart tutorial header is required")
	}
	for _, field := range []struct {
		value *string
		name  string
		limit int
	}{
		{&header.Kicker, "header kicker", 32},
		{&header.Title, "header title", 96},
		{&header.Description, "header description", 500},
		{&header.LibraryActionLabel, "library action label", 64},
		{&header.KeysActionLabel, "keys action label", 64},
		{&header.PlatformControlLabel, "platform control label", 64},
		{&header.TerminalControlLabel, "terminal control label", 64},
	} {
		if err := normalizeQuickstartText(field.value, field.name, field.limit); err != nil {
			return err
		}
	}
	return nil
}

func normalizeQuickstartPlatform(platform QuickstartTutorialPlatform) (QuickstartTutorialPlatform, error) {
	platform.ID = strings.ToLower(strings.TrimSpace(platform.ID))
	if platform.ID != "codex" && platform.ID != "claude" {
		return platform, infraerrors.BadRequest("QUICKSTART_TUTORIAL_PLATFORM_ID_INVALID", "quickstart tutorial platform id must be codex or claude")
	}
	for _, field := range []struct {
		value *string
		name  string
		limit int
	}{
		{&platform.Label, "platform label", 64},
		{&platform.ClientName, "platform client name", 96},
		{&platform.BaseURLDescription, "base url description", 300},
		{&platform.AuthHint, "auth hint", 160},
		{&platform.Protocol, "protocol", 96},
		{&platform.ModelHint, "model hint", 160},
	} {
		if err := normalizeQuickstartText(field.value, field.name, field.limit); err != nil {
			return platform, err
		}
	}
	platform.BaseURL = strings.TrimRight(strings.TrimSpace(platform.BaseURL), "/")
	if len(platform.BaseURL) > 2048 {
		return platform, infraerrors.BadRequest("QUICKSTART_TUTORIAL_BASE_URL_TOO_LONG", "quickstart tutorial base URL is too long")
	}
	parsed, err := url.ParseRequestURI(platform.BaseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return platform, infraerrors.BadRequest("QUICKSTART_TUTORIAL_BASE_URL_INVALID", "quickstart tutorial base URL must be an absolute http(s) URL")
	}
	return platform, nil
}

func normalizeQuickstartFacts(facts *QuickstartTutorialFacts) error {
	if facts == nil {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_FACTS_INVALID", "quickstart tutorial facts are required")
	}
	for _, field := range []struct {
		value *string
		name  string
		limit int
	}{
		{&facts.BaseURLLabel, "base URL label", 64},
		{&facts.AuthLabel, "auth label", 64},
		{&facts.AuthDescription, "auth description", 300},
		{&facts.ProtocolLabel, "protocol label", 64},
		{&facts.ProtocolDescription, "protocol description", 300},
		{&facts.ModelLabel, "model label", 64},
		{&facts.ModelDescription, "model description", 300},
	} {
		if err := normalizeQuickstartText(field.value, field.name, field.limit); err != nil {
			return err
		}
	}
	return nil
}

func normalizeQuickstartDesktop(desktop *QuickstartTutorialDesktop) error {
	if desktop == nil {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_DESKTOP_INVALID", "quickstart tutorial desktop section is required")
	}
	if err := normalizeQuickstartSection(&desktop.QuickstartTutorialSection, "desktop"); err != nil {
		return err
	}
	if len(desktop.Tiles) == 0 || len(desktop.Tiles) > quickstartTutorialMaxTiles {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_DESKTOP_TILES_INVALID", "quickstart tutorial requires between 1 and 6 desktop tiles")
	}
	for i := range desktop.Tiles {
		tile := &desktop.Tiles[i]
		if err := normalizeQuickstartText(&tile.Number, "desktop tile number", 16); err != nil {
			return err
		}
		if err := normalizeQuickstartText(&tile.Title, "desktop tile title", 96); err != nil {
			return err
		}
		if err := normalizeQuickstartText(&tile.Description, "desktop tile description", 500); err != nil {
			return err
		}
	}
	return nil
}

func normalizeQuickstartSection(section *QuickstartTutorialSection, name string) error {
	if section == nil {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_SECTION_INVALID", "quickstart tutorial section is required")
	}
	if err := normalizeQuickstartText(&section.Kicker, name+" kicker", 32); err != nil {
		return err
	}
	if err := normalizeQuickstartText(&section.Title, name+" title", 96); err != nil {
		return err
	}
	return normalizeQuickstartText(&section.Description, name+" description", 500)
}

func normalizeQuickstartText(value *string, field string, limit int) error {
	if value == nil {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_TEXT_INVALID", "quickstart tutorial "+field+" is required")
	}
	*value = strings.TrimSpace(*value)
	return validateQuickstartText(*value, field, limit)
}

func validateQuickstartText(value, field string, limit int) error {
	if value == "" {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_TEXT_REQUIRED", "quickstart tutorial "+field+" is required")
	}
	if len([]rune(value)) > limit {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_TEXT_TOO_LONG", "quickstart tutorial "+field+" is too long")
	}
	if strings.ContainsAny(value, "<>") {
		return infraerrors.BadRequest("QUICKSTART_TUTORIAL_HTML_FORBIDDEN", "quickstart tutorial only supports plain text")
	}
	return nil
}
