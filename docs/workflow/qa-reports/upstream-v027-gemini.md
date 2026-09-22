### PASS: upstream-v027-gemini

## QA 范围

- 独立审阅契约 allowlist 的 7 个 Gemini 文件：
  `antigravity_gateway_service.go`、`antigravity_gemini_thinking_variant.go`、
  `gemini_sse_comment_compat.go`、`gemini_messages_compat_service.go`、
  `gemini_v1beta_handler.go` 及两份 `*_v027_test.go`。
- 未修改业务代码、未提交、未推送、未部署；并行任务的其余脏改动未纳入本结论。

## 结果

- PASS - 初始 QA 曾因缺少精确 identity mapping 回归而判 FAIL。仅测试补充后，
  `TestV027GeminiThinkingVariantRespectsMappings` 现在断言
  `"gemini-3.8-flash": "gemini-3.8-flash"` 且同时存在 `-low/-high` 变体候选时，
  `resolveGeminiThinkingVariant` 返回 `ok=false` 和空值。`ForwardGemini` 因而进入
  既有 `getMappedModel` 回退，保留 explicit identity mapping；独立重跑通过。
- PASS - Thinking variant：仅 bare `gemini-*` 请求进入候选变体逻辑；请求显式
  `-low/-medium/-high/-tiered` 后缀、非 Gemini 模型、精确 custom 映射和精确
  identity 映射均保留既有映射。`thinkingLevel` 的有效值优先于
  `thinkingBudget`；无效/缺失值按 high，budget 按 `<=1024` low、`<=8192`
  medium、其余 high 分档。
- PASS - `ForwardGemini` fake HTTP 运行用例验证包装请求和 `UpstreamModel` 使用
  解析出的变体，而非仅验证辅助函数。
- PASS - SSE：fake stream 证实普通客户端仍接收 `:\n\n` keepalive；
  `google-genai-sdk/... gl-go/` 与 `gl-python/` 不接收注释。源码同时检查
  `User-Agent` 与 `X-Goog-Api-Client`，保留取消、错误和 usage 流程。
- PASS - model listing：普通 Gemini 路径保留 custom group list 优先；强制
  Antigravity 路径先于 custom list。普通路径仅纳入 `mixed_scheduling=true`
  的可调度 Antigravity 账号，强制路径不要求 opt-in；枚举仅来自对应 group 的
  Gemini mapping key，排除非 Gemini key。
- PASS - native `/v1beta/models`：合并仅发生在 200 且存在额外 Antigravity
  模型时；保留上游模型对象字段、`nextPageToken`、响应头和非 200 错误体。没有
  可调度且可见的 Antigravity Gemini mapping 时不会添加模型。
- N/A - 本地 `service.Group` 没有上游的 model allowlist 字段/API，无法实施或
  验证上游 allowlist 过滤；代码未虚构该能力，后续若引入该配置需补充列表过滤测试。

## 已执行命令与证据

在 `backend`：

```powershell
go test ./internal/service ./internal/handler -run '^TestV027Gemini' -count=1
# PASS: internal/service 2.483s; internal/handler 0.083s

# identity mapping 测试补充后的独立重跑
go test ./internal/service ./internal/handler -run '^TestV027Gemini' -count=1
# PASS: internal/service 2.477s; internal/handler 0.075s

go test -tags unit ./internal/handler -run '^(TestGeminiV1BetaListModels_CustomGroupListUsesNativeResponse|TestGeminiV1BetaListModels_ForcedAntigravityIgnoresCustomGroupList)$' -count=1
# PASS: internal/handler 0.076s

go build ./...
# PASS

gofmt -d <7 allowlist Go files>
# PASS: 无输出
```

在仓库根目录：

```powershell
git diff HEAD --check
# PASS

git diff --name-only --diff-filter=U
# PASS: 空（NO_UNMERGED_PATHS）
```

`TestV027Gemini*` 覆盖 fake HTTP 的最终请求模型、level/budget 优先级、explicit
custom/identity mapping、显式 suffix、非目标模型、fake SSE heartbeat、mixed opt-out、
group 隔离、native metadata/pagination 和错误透传。既有 unit-tag 两例确认 custom
list 与 forced Antigravity 的优先级。

## 未验证边界

- 未访问真实 Gemini 或 Antigravity provider；真实上游认证、模型可用性、实际 SDK
  网络行为、数据库调度、容器、部署均未测。
