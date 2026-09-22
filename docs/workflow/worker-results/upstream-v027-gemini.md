### PASS: upstream-v027-gemini

## 已实施

- 在本地单体 `AntigravityGatewayService.ForwardGemini` 接入裸 Gemini 模型的
  `generationConfig.thinkingConfig` 变体解析。仅当账号没有显式裸模型映射时，才从
  `-low/-medium/-high/-tiered` 映射中选取变体；显式后缀、显式 identity mapping、非 Gemini
  模型与无变体映射保持既有路径。
- 在本地 Gemini 流 owner 中，对 go-genai 与 python-genai 的
  `google-genai-sdk/... gl-go/` / `gl-python/` 客户端抑制 SSE 注释心跳；其他客户端保留
  `:\n\n` keepalive。
- Gemini `/v1beta/models` 从可调度 Antigravity 账号的实际 Gemini mapping key 派生目录。
  普通 Gemini 路由只纳入 `mixed_scheduling=true` 的账号；强制 Antigravity 路由不要求该 opt-in。
  分组 custom model list 和 forced platform 的既有优先级保留，真实 native upstream 响应在合并时
  保留模型元数据、分页 token、响应头与错误响应。

## 运行证据

- `go test ./internal/service ./internal/handler -run '^TestV027Gemini' -count=1`：PASS。
  覆盖 fake HTTP `ForwardGemini` 请求包装内的最终模型、thinkingLevel 优先、预算分档、显式/
  identity mapping、非目标模型、fake SSE 的 GenAI heartbeat 兼容、mixed opt-out、分组隔离、
  native 元数据与错误透传。
- `go test -tags unit ./internal/handler -run '^(TestGeminiV1BetaListModels_CustomGroupListUsesNativeResponse|TestGeminiV1BetaListModels_ForcedAntigravityIgnoresCustomGroupList)$' -count=1`：PASS。
- `go build ./...`：PASS。
- `git diff --check`：PASS；`git diff --name-only --diff-filter=U`：空。

更正：`TestV027GeminiThinkingVariantRespectsMappings` 现明确断言
`"gemini-3.8-flash": "gemini-3.8-flash"` 的精确 identity mapping 在 low budget
请求下不选择 `-low` 变体；原有非 identity custom mapping、显式 suffix 和非目标模型断言仍保留。

## 范围与限制

- 本地 `service.Group` 不存在上游 `ModelAllowlistEnabled`、`ModelAllowlist` 或
  `filterUpstreamGeminiModelsBody` API，故 group allowlist 条目不适用；未引入虚构字段。
- 未访问真实 Gemini/Antigravity provider，未运行部署，未提交、暂存或推送。
