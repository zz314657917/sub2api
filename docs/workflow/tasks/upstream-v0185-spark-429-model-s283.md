---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-spark-429-model-s283
worker_model: gpt-5.6-terra
base_commit: f48b4b77f
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 Spark 429 model isolation S283

## Task ID

`upstream-v0185-spark-429-model-s283`

## Role

你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `5d9c7abed`：OpenAI OAuth 的 Codex Spark 429 配额只写入模型级限流，不再把 Spark 配额耗尽误当成账号级全局 cooldown。Spark 影子账号继续隔离于母账号的 global Codex quota，普通 OAuth 429、API key、非 OpenAI 账号和既有 S282 运行时阻断语义保持不变。

## Success Criteria

- `RateLimitService` 能识别 OpenAI OAuth 的 Spark 请求（含 Spark shadow），按请求模型或账号映射后的模型判定并以归一化 Spark key 写入现有 `model_rate_limits`/`SetModelRateLimit`，reset 头、body reset 或既有秒级 fallback 均生效；不写账号级 `SetRateLimited`，不建立账号级 runtime block。
- 同一 OAuth 账号的非 Spark 模型不受 Spark 模型限流影响；API key、setup-token、非 OpenAI 账号与普通非 Spark 429 不误命中该路径。
- OpenAI HTTP/WS 错误入口在调用通用 OAuth 429 账号级逻辑前传递 canonical/mapped model；WS HTTP bridge、标准 forwarder、Responses WS v2 passthrough 的握手与后续 error event 均覆盖，且不改变客户端错误输出、failover 和连接复用语义。
- 复用本地既有 `AccountRepository.SetModelRateLimit(ctx, id, scope, resetAt)` 签名，不改 repository、Ent、migration 或 scheduler 持久化结构；保留 S282 的普通 OAuth 429 retry window、shadow global header 隔离和最长 runtime block/clear 语义。
- 新增定向测试覆盖：Spark OAuth 模型级写入、5h/7d/body reset/fallback、Spark shadow、同账号非 Spark 模型、API key/非 OAuth 边界及 WS model 透传；既有 model transient 与 WS rate-limit 测试继续通过。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `5d9c7abed59a5a53e36cd6cd62257807a3099ab7` (`fix(openai): scope Spark quota 429 to model`)
- Local base: `f48b4b77f` (S282 OAuth 429 quota runtime handling)
- Local owner differs from upstream split files: generic OpenAI error handling is in `openai_images_responses.go`; WS signal persistence is in `openai_ws_forwarder.go`, `openai_ws_http_bridge.go` and `openai_ws_v2_passthrough_adapter.go`; model-level repository API already exists with a fixed four-argument signature.
- The monolithic HTTP/SSE passthrough in `openai_gateway_service.go` already passes the requested model through the S282 error path; it is verified as an existing dependency but is not modified in S283.
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`.

## Allowed Paths

- `backend/internal/service/model_rate_limit.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_spark_quota_model_test.go` (新增)
- `docs/workflow/worker-results/upstream-v0185-spark-429-model-s283-result.md`
- `docs/workflow/qa-reports/upstream-v0185-spark-429-model-s283-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/internal/service/admin_service.go`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/openai_ws_v2/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `frontend/**`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/**`（本 contract 文件除外）
- `docs/workflow/contract-reviews/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器或数据文件

## Constraints

- 不 cherry-pick、merge、rebase 或照搬上游拆分文件；按本地 owner 手工适配。
- Spark 模型判断必须复用 `isCodexSparkModel`/`normalizeCodexModel` 与现有 model-rate-limit key 解析，并同时覆盖请求模型和映射后 key；不新增数据库字段、migration 或修改 `AccountRepository` 接口。
- Spark shadow 的 global `/responses` quota header 不得触发账号级 `RateLimitResetAt` 或账号级 runtime block；Spark 请求本身的 429 可以只写 Spark 模型 key，即使 header 来自 global quota，且不得升级为账号级状态。
- 429 处理必须先执行 Spark model scope，再进入 S282 的 OAuth 账号级分类；非 Spark 路径的 retry window、fallback、API key transient、pool-mode、Grok 和其他平台语义不变。
- WS 调用点只补 canonical/mapped model 参数，不改变握手 quota header 隔离、客户端写出顺序、failover budget、连接池生命周期或错误状态码；未纳入的 count-tokens 等直接错误路径保持原状。
- 保留所有既有保护脏文件和 `outputs/**`，其 aggregate dirty diff hash 必须保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`。
- 不执行真实 provider、数据库、容器、部署或 push。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/service -run 'Test(S283|OpenAI.*Spark|Spark.*Quota|OpenAIWS.*RateLimit)' -count=10
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -d internal/service/model_rate_limit.go internal/service/ratelimit_service.go internal/service/openai_images_responses.go internal/service/openai_ws_forwarder.go internal/service/openai_ws_http_bridge.go internal/service/openai_ws_v2_passthrough_adapter.go internal/service/openai_spark_quota_model_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/service
git diff --name-only --diff-filter=U
git status --short
```

Also inspect exact allowlist/denied-path audit and the protected aggregate hash; `go test -tags unit ./internal/service` may expose unrelated baseline fixture/API drift and must be recorded without weakening tests.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-spark-429-model-s283-result.md`; first line must be `### DONE: upstream-v0185-spark-429-model-s283`, `### BLOCKED: ...` or `### FAILED: ...`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-spark-429-model-s283-qa.md` with first line `### PASS/FAIL/BLOCKED: upstream-v0185-spark-429-model-s283`.

## Stop Rules

- If implementation requires repository/Ent/migration/handler/frontend changes, semantic WS header isolation, or a new model-rate-limit persistence API, stop with `BLOCKED` and report the topology boundary.
- If any allowed target has unrelated concurrent edits or a protected baseline changes, stop without overwriting.
- Do not weaken tests for baseline unit-tag failures; do not modify denied paths or `outputs/**`.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
