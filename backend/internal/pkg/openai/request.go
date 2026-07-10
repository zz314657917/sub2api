package openai

import (
	"regexp"
	"strings"
)

// CodexCLIUserAgentPrefixes matches Codex CLI User-Agent patterns
// Examples: "codex_vscode/1.0.0", "codex_cli_rs/0.1.2"
var CodexCLIUserAgentPrefixes = []string{
	"codex_vscode/",
	"codex_cli_rs/",
}

// codexOfficialClientUAPrefixes：Codex 官方客户端家族 User-Agent 前缀（均含下划线/连字符，
// 每项都是确定字面量；不含会被 TrimSpace 退化成裸 "codex" 的空格前缀）。
// 用途：OpenAI OAuth `codex_cli_only` 访问限制判定 + passthrough 的「非官方 UA 安全兜底」
// （IsCodexOfficialClientRequest 命中即视为官方真实 UA，逐字透传、不改写）。
//
// Cursor/VSCode 扩展两种 UA：默认 `codex_vscode/`、GitHub Copilot 集成模式 `codex_vscode_copilot/`
// （取证 extension.js `IS="codex_vscode_copilot"` 经 env 注入）；交互式 TUI 自报 `codex-tui/`
// （连字符，2026-06-23 审计抽样约占真实流量 35%，必须显式列出）。`Codex Desktop/` 等 `Codex `
// 前缀家族由 codexOfficialClientFamilyPrefix 单独处理（保留空格，避免退化为裸 codex 的宽松兜底）。
var codexOfficialClientUAPrefixes = []string{
	"codex_cli_rs/",
	"codex-tui/",
	"codex_vscode/",
	"codex_vscode_copilot/",
	"codex_app/",
	"codex_chatgpt_desktop/",
	"codex_atlas/",
	"codex_exec/",
	"codex_sdk_ts/",
}

// codexOfficialClientFamilyPrefix 覆盖 `Codex ` 前缀家族（Codex Desktop 等），对应 codex-rs
// is_first_party_originator 的 starts_with("Codex ")。**保留尾随空格**，并以 HasPrefix 直接比对
// 已归一化（小写 + 去首尾空格）的值——绝不能再经 normalizeCodexClientHeader 处理本前缀，否则
// 空格被 TrimSpace 去掉、退化成裸 "codex" 而把任何含 codex 的串都放行。
const codexOfficialClientFamilyPrefix = "codex "

// codexOfficialClientOriginators：Codex 官方客户端家族 originator 精确集合。
// app-server `initialize` 把 originator 设为 clientInfo.name 逐字值（codex-rs default_client.rs），
// 故官方集合是这些确定字面量；镜像 is_first_party_originator / is_first_party_chat_originator
// 并叠加 sub2api 已取证变体。用精确匹配而非「含 codex_/codex」的宽松兜底，避免 evil-codex_ 之类
// 伪造绕过（gate 仍需 UA 双因子佐证）。新官方/合作客户端经 allowed_client.go 命名预设放行，
// 或在 bump context/codex 时同步补入本集合。
var codexOfficialClientOriginators = map[string]bool{
	"codex_cli_rs":          true,
	"codex-tui":             true,
	"codex_vscode":          true,
	"codex_vscode_copilot":  true,
	"codex_app":             true,
	"codex_chatgpt_desktop": true,
	"codex_atlas":           true,
	"codex_exec":            true,
	"codex_sdk_ts":          true,
}

// IsBrowserUserAgent 判断 User-Agent 是否来自浏览器（Chrome/Firefox/Safari/Edge/Opera 等）。
// 所有现代浏览器的 UA 均以 "Mozilla/" 作为前缀，CLI 工具（codex/claude/curl/postman/python-requests 等）不会。
// 该判定用于避免 Cloudflare 对浏览器型 UA 在 OpenAI 上游接口上触发 JS 质询。
func IsBrowserUserAgent(userAgent string) bool {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(ua), "mozilla/")
}

// IsCodexCLIRequest checks if the User-Agent indicates a Codex CLI request
func IsCodexCLIRequest(userAgent string) bool {
	ua := normalizeCodexClientHeader(userAgent)
	if ua == "" {
		return false
	}
	return matchCodexClientHeaderPrefixes(ua, CodexCLIUserAgentPrefixes)
}

// IsCodexOfficialClientRequest checks if the User-Agent indicates a Codex 官方客户端请求。
// 与 IsCodexCLIRequest 解耦，避免影响历史兼容逻辑。宽松版：官方 UA 前缀集允许 Contains 子串兜底，
// 供 passthrough（IsCodexOfficialClientByHeaders）等历史路径使用，行为不变。
func IsCodexOfficialClientRequest(userAgent string) bool {
	return isCodexOfficialClientRequest(userAgent, false)
}

// IsCodexOfficialClientRequestStrict 同 IsCodexOfficialClientRequest，但官方 UA 前缀集只做前缀
// 匹配（HasPrefix），不退化为 Contains 子串兜底——专供 codex_cli_only 访问门，收窄「浏览器前缀 +
// 中段 codex token」之类的伪造面。`Codex ` 家族前缀与 UA 尾部兜底保持一致；passthrough 仍用宽松版。
func IsCodexOfficialClientRequestStrict(userAgent string) bool {
	return isCodexOfficialClientRequest(userAgent, true)
}

func isCodexOfficialClientRequest(userAgent string, strict bool) bool {
	ua := normalizeCodexClientHeader(userAgent)
	if ua == "" {
		return false
	}
	if strict {
		if matchCodexClientHeaderStrictPrefixes(ua, codexOfficialClientUAPrefixes) {
			return true
		}
	} else if matchCodexClientHeaderPrefixes(ua, codexOfficialClientUAPrefixes) {
		return true
	}
	if strings.HasPrefix(ua, codexOfficialClientFamilyPrefix) {
		return true
	}
	if name := codexUATrailerName(ua); name != "" {
		return IsCodexOfficialClientOriginator(name)
	}
	return false
}

func codexUATrailerName(ua string) string {
	last := strings.LastIndex(ua, "(")
	if last < 0 {
		return ""
	}
	rest := ua[last+1:]
	closeIdx := strings.Index(rest, ")")
	if closeIdx < 0 {
		return ""
	}
	inner := strings.TrimSpace(rest[:closeIdx])
	if semi := strings.Index(inner, ";"); semi >= 0 {
		inner = strings.TrimSpace(inner[:semi])
	}
	return inner
}

// IsCodexOfficialClientOriginator checks if originator indicates a Codex 官方客户端请求。
// 精确集合匹配 + `Codex ` 家族前缀；不再用「含 codex」宽松兜底（避免伪造绕过）。
func IsCodexOfficialClientOriginator(originator string) bool {
	v := normalizeCodexClientHeader(originator)
	if v == "" {
		return false
	}
	if codexOfficialClientOriginators[v] {
		return true
	}
	return strings.HasPrefix(v, codexOfficialClientFamilyPrefix)
}

// IsCodexOfficialClientByHeaders checks whether the request headers indicate an
// official Codex client family request.
func IsCodexOfficialClientByHeaders(userAgent, originator string) bool {
	return IsCodexOfficialClientRequest(userAgent) || IsCodexOfficialClientOriginator(originator)
}

func normalizeCodexClientHeader(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func matchCodexClientHeaderPrefixes(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		normalizedPrefix := normalizeCodexClientHeader(prefix)
		if normalizedPrefix == "" {
			continue
		}
		if strings.HasPrefix(value, normalizedPrefix) || strings.Contains(value, normalizedPrefix) {
			return true
		}
	}
	return false
}

func matchCodexClientHeaderStrictPrefixes(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if p := normalizeCodexClientHeader(prefix); p != "" && strings.HasPrefix(value, p) {
			return true
		}
	}
	return false
}

// PairCodexClientIdentity 由最终出站 User-Agent 推导与其配套的 originator，必要时归一化
// UA 首段，保证两者一致。
func PairCodexClientIdentity(userAgent string) (originator string, pairedUA string, ok bool) {
	ua := strings.TrimSpace(userAgent)
	slash := strings.IndexByte(ua, '/')
	if slash <= 0 {
		return "", "", false
	}
	if leading := strings.TrimSpace(ua[:slash]); isSaneCodexOriginator(leading) && IsCodexOfficialClientOriginator(leading) {
		leading = canonicalizeCodexOriginator(leading)
		return leading, leading + ua[slash:], true
	}
	if trailer := codexUATrailerName(ua); trailer != "" && !strings.ContainsRune(trailer, '/') &&
		isSaneCodexOriginator(trailer) && IsCodexOfficialClientOriginator(trailer) {
		trailer = canonicalizeCodexOriginator(trailer)
		return trailer, trailer + ua[slash:], true
	}
	return "", "", false
}

const codexOriginatorMaxLen = 64

func isSaneCodexOriginator(name string) bool {
	if name == "" || len(name) > codexOriginatorMaxLen {
		return false
	}
	for i := 0; i < len(name); i++ {
		if c := name[i]; c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func canonicalizeCodexOriginator(name string) string {
	if lower := normalizeCodexClientHeader(name); codexOfficialClientOriginators[lower] {
		return lower
	}
	return name
}

var codexEngineVersionPattern = regexp.MustCompile(`^(\d+\.\d+\.\d+)`)

// ParseCodexEngineVersion 从 codex-rs 形态 UA 取引擎版本：
// `{originator}/{X.Y.Z} (...)`，第一个 '/' 后、首个空格或 '(' 前的三段版本。
func ParseCodexEngineVersion(ua string) (string, bool) {
	ua = strings.TrimSpace(ua)
	slash := strings.IndexByte(ua, '/')
	if slash < 0 {
		return "", false
	}
	rest := ua[slash+1:]
	end := len(rest)
	for i := 0; i < len(rest); i++ {
		if rest[i] == ' ' || rest[i] == '(' {
			end = i
			break
		}
	}
	m := codexEngineVersionPattern.FindString(strings.TrimSpace(rest[:end]))
	if m == "" {
		return "", false
	}
	return m, true
}
