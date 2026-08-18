package service

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEnforceCodexIdentityHeaders(t *testing.T) {
	const tuiUA = "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)"
	const tuiNormalizedUA = "codex_cli_rs/0.140.2 (Mac OS X 14.0; arm64) iTerm"

	tests := []struct {
		name           string
		originator     string
		userAgent      string
		version        string
		wantOriginator string
		wantUA         string
		wantVersion    string
	}{
		{
			name:           "错配 originator 按最终 UA 重配后归一化",
			originator:     "codex_cli_rs",
			userAgent:      tuiUA,
			wantOriginator: "codex_cli_rs",
			wantUA:         tuiNormalizedUA,
		},
		{
			name:           "降载身份改写为 CLI 身份",
			originator:     "codex-tui",
			userAgent:      tuiUA,
			wantOriginator: "codex_cli_rs",
			wantUA:         tuiNormalizedUA,
		},
		{
			name:           "非降载官方身份原样保留",
			originator:     "codex_vscode",
			userAgent:      "codex_vscode/1.2.3 (Ubuntu 22.4.0; x86_64) vscode (codex_vscode; 1.2.3)",
			wantOriginator: "codex_vscode",
			wantUA:         "codex_vscode/1.2.3 (Ubuntu 22.4.0; x86_64) vscode (codex_vscode; 1.2.3)",
		},
		{
			name:           "第三方 UA 整体回退默认身份",
			originator:     "opencode",
			userAgent:      "luna/1.0.0",
			wantOriginator: "codex_cli_rs",
			wantUA:         openai.CodexCLIOriginator + "/" + codexCLIVersion,
		},
		{
			name:           "UA 缺失回退默认身份",
			originator:     "codex_vscode",
			wantOriginator: "codex_cli_rs",
			wantUA:         openai.CodexCLIOriginator + "/" + codexCLIVersion,
		},
		{
			name:           "originator override UA 首段被尾部真实身份重写后归一化",
			originator:     "cccc",
			userAgent:      "cccc/0.142.0 (Ubuntu 22.4.0; x86_64) screen (codex-tui; 0.142.0)",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.142.0 (Ubuntu 22.4.0; x86_64) screen",
		},
		{
			name:           "低于门槛的 version 提升为内置版本",
			originator:     "codex_cli_rs",
			userAgent:      "codex_cli_rs/0.125.0",
			version:        "0.125.0",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.125.0",
			wantVersion:    codexCLIVersion,
		},
		{
			name:           "达标 version 原样保留",
			originator:     "codex_cli_rs",
			userAgent:      "codex_cli_rs/0.145.0",
			version:        "0.145.0",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.145.0",
			wantVersion:    "0.145.0",
		},
		{
			name:           "未携带 version 不注入",
			originator:     "codex_cli_rs",
			userAgent:      "codex_cli_rs/0.98.0",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.98.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := make(http.Header)
			if tt.originator != "" {
				h.Set("originator", tt.originator)
			}
			if tt.userAgent != "" {
				h.Set("user-agent", tt.userAgent)
			}
			if tt.version != "" {
				h.Set("version", tt.version)
			}

			enforceCodexIdentityHeaders(h)

			require.Equal(t, tt.wantOriginator, h.Get("originator"))
			require.Equal(t, tt.wantUA, h.Get("user-agent"))
			require.Equal(t, tt.wantVersion, h.Get("version"))
		})
	}
}

func TestCodexCanonicalAuthIdentityUsesPairedHeadersWithoutVersion(t *testing.T) {
	userAgent, originator := CodexCanonicalAuthIdentity()
	require.Equal(t, codexCLIUserAgent, userAgent)
	require.Equal(t, openAIDefaultCodexOriginator, originator)

	h := make(http.Header)
	h.Set("version", "0.125.0")
	ApplyCodexCanonicalAuthIdentity(h)

	require.Equal(t, userAgent, h.Get("user-agent"))
	require.Equal(t, originator, h.Get("originator"))
	require.Empty(t, h.Get("version"))
}

// 开关是进程级快照，零值 Config（测试 / 工具手工构造，不经 viper）必须落在「归一化开启」
// 一侧。开关类用例不能并行，因为它们会改写进程级状态。
func TestCodexOriginatorNormalizationZeroValueConfigKeepsItEnabled(t *testing.T) {
	var cfg config.Config
	require.False(t, cfg.Gateway.DisableCodexOriginatorNormalization)

	SetCodexOriginatorNormalizationEnabled(!cfg.Gateway.DisableCodexOriginatorNormalization)
	t.Cleanup(func() { SetCodexOriginatorNormalizationEnabled(true) })

	h := make(http.Header)
	h.Set("originator", "codex-tui")
	h.Set("user-agent", "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)")

	enforceCodexIdentityHeaders(h)

	require.Equal(t, openai.CodexCLIOriginator, h.Get("originator"))
}

func TestEnforceCodexIdentityHeaders_NormalizationDisabled(t *testing.T) {
	const tuiUA = "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)"

	SetCodexOriginatorNormalizationEnabled(false)
	t.Cleanup(func() { SetCodexOriginatorNormalizationEnabled(true) })

	h := make(http.Header)
	h.Set("originator", "codex-tui")
	h.Set("user-agent", tuiUA)

	enforceCodexIdentityHeaders(h)

	require.Equal(t, "codex-tui", h.Get("originator"))
	require.Equal(t, tuiUA, h.Get("user-agent"))
}

func TestEnforceCodexIdentityHeaders_NormalizationIsIdempotent(t *testing.T) {
	h := make(http.Header)
	h.Set("originator", "codex-tui")
	h.Set("user-agent", "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)")

	enforceCodexIdentityHeaders(h)
	first := h.Get("user-agent")
	enforceCodexIdentityHeaders(h)

	require.Equal(t, first, h.Get("user-agent"))
	require.Equal(t, openai.CodexCLIOriginator, h.Get("originator"))
}

// compat messages bridge 故意不带 originator：收口必须保持 no-op，不得注入身份头。
func TestEnforceCodexIdentityHeaders_NoOriginatorIsNoop(t *testing.T) {
	h := make(http.Header)
	h.Set("user-agent", "luna/1.0.0")

	enforceCodexIdentityHeaders(h)

	require.Empty(t, h.Get("originator"))
	require.Equal(t, "luna/1.0.0", h.Get("user-agent"))
}

func TestCompatMessagesBridgeOriginator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	require.False(t, isOpenAICompatMessagesBridgeContext(c))

	setOpenAICompatMessagesBridgeContext(c, true)

	require.True(t, isOpenAICompatMessagesBridgeContext(c))
	require.Empty(t, resolveOpenAIUpstreamOriginator(c, true))
}
