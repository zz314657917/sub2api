---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-anthropic-fallback-s287
worker_model: controller
base_commit: 409a93110
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-02
---

# Upstream v0.1.185 Anthropic fallback field sanitization S287

## Goal

适配上游 `200b1406d`：客户端携带 `fallbacks` 或 `fallback_credit_token` 时，只有最终 Anthropic beta header 明确声明对应 beta 才保留字段；否则在签名和转发前清除，避免标准 Messages schema 以 `Extra inputs are not permitted` 拒绝请求。Bedrock 路径不支持这些 beta 字段，始终清除对应 body 字段。

## Allowed Paths

- `backend/internal/pkg/claude/constants.go`
- `backend/internal/service/gateway_request.go`
- `backend/internal/service/bedrock_request.go`
- `backend/internal/service/gateway_fallbacks_sanitize_test.go`
- `docs/workflow/worker-results/upstream-v0185-anthropic-fallback-s287-result.md`
- `docs/workflow/qa-reports/upstream-v0185-anthropic-fallback-s287-qa.md`

## Invariants

- `fallbacks` 保留条件仅为 `server-side-fallback-2026-07-01`。
- `fallback_credit_token` 保留条件为 `server-side-fallback-2026-07-01`、`fallback-credit-2026-07-01` 或 `fallback-credit-2026-06-01` 任一 token。
- 不把 fallback beta 注入默认 OAuth/API-key mimic header，不改变模型路由、计费或签名顺序；其他 body 字段与现有 context-management 清理保持不变。
- Bedrock 的 fallback token 不进入其白名单，因此 `fallbacks` 与 `fallback_credit_token` 始终清除；不引入上游本地不存在的 CC 专用函数。
- 不修改 Ent、migration、repository、handler、frontend、provider、数据库、容器、部署或保护性脏改。

## Acceptance

From `backend`:

```text
go test ./internal/service -run 'Test(SanitizeAnthropicBodyForBetaTokens|SanitizeBedrockFieldsForBetaTokens)' -count=10
go test ./internal/service
go test ./cmd/server -run '^$' -count=1
gofmt -d internal/pkg/claude/constants.go internal/service/gateway_request.go internal/service/bedrock_request.go internal/service/gateway_fallbacks_sanitize_test.go
```

From repository root:

```text
git diff --check -- backend/internal/pkg/claude/constants.go backend/internal/service/gateway_request.go backend/internal/service/bedrock_request.go backend/internal/service/gateway_fallbacks_sanitize_test.go
git diff --name-only --diff-filter=U
```

真实 Anthropic/Bedrock provider、数据库、容器、部署和浏览器 smoke 不属于本 Sprint；完整 `-tags unit` 基线漂移需单独记录。
