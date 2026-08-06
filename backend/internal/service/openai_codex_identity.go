package service

import (
	"net/http"
	"strings"
	"sync/atomic"

	openaipkg "github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

const openAIDefaultCodexOriginator = "codex_cli_rs"

// codexUpstreamMinVersion 上游 /backend-api/codex 接受的最低 version 头：
// 若请求携带 version 且低于该值，上游直接 404。
const codexUpstreamMinVersion = "0.144.0"

// codexOriginatorNormalization controls whether the shared egress boundary rewrites load-shed
// Codex identities to the CLI identity. The gateway constructor publishes the startup config snapshot.
var codexOriginatorNormalization = func() *atomic.Bool {
	v := &atomic.Bool{}
	v.Store(true)
	return v
}()

// SetCodexOriginatorNormalizationEnabled publishes the Codex load-shed identity normalization switch.
func SetCodexOriginatorNormalizationEnabled(enabled bool) {
	codexOriginatorNormalization.Store(enabled)
}

// enforceCodexIdentityHeaders 收口 OAuth（ChatGPT 内部接口）出站请求的客户端身份头。
// 仅对携带 originator 的请求生效；compat messages bridge 故意不带 originator，保持原样。
func enforceCodexIdentityHeaders(headers http.Header) {
	if headers == nil {
		return
	}
	originator := strings.TrimSpace(headers.Get("originator"))
	if originator == "" {
		return
	}
	pairedOriginator, pairedUA, ok := openaipkg.PairCodexClientIdentity(headers.Get("user-agent"))
	if !ok {
		pairedOriginator, pairedUA = openaipkg.CodexCLIOriginator, codexCLIUserAgent
	}
	if codexOriginatorNormalization.Load() {
		pairedOriginator, pairedUA, _ = openaipkg.NormalizeCodexClientIdentityToCLI(pairedOriginator, pairedUA)
	}
	headers.Set("originator", pairedOriginator)
	headers.Set("user-agent", pairedUA)
	if version := strings.TrimSpace(headers.Get("version")); version != "" && CompareVersions(version, codexUpstreamMinVersion) < 0 {
		headers.Set("version", codexCLIVersion)
	}
}
