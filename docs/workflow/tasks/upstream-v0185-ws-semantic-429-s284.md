---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-ws-semantic-429-s284
worker_model: gpt-5.6-terra
base_commit: bb3d3bca6
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 WebSocket semantic 429 isolation S284

## Task ID

`upstream-v0185-ws-semantic-429-s284`

## Role

你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `571d1e1d9`：WebSocket 连接建立后的语义 429 不得把握手阶段的全局 Codex quota header 误用于普通模型的账号级 cooldown；只有 Spark OAuth 模型语义 429 保留并使用该模型的 quota header。握手本身返回 HTTP 429 时仍保留响应 header，维持 S283 的 Spark 模型级限流与 S282 的 OAuth 账号级 retry 语义。

## Success Criteria

- 标准 WS forwarder、WS HTTP bridge 和 v2 passthrough 的连接后 `error`/`response.failed` 429，在普通模型或非 OAuth 账号上清空握手 quota header 后进入既有错误处理，不因 global 5h/7d header 写入账号级 cooldown。
- Spark OAuth 模型的连接后语义 429 保留握手 quota header，并由 S283 的 model-scoped handler 写入归一化 Spark `model_rate_limits`；不建立账号级 runtime block 或 `SetRateLimited`。
- 握手阶段 HTTP 429（无已建立连接的 response body）继续保留响应 header，普通 OAuth 账号沿用既有分类，Spark OAuth 继续模型级隔离。
- 不改变客户端错误事件写出顺序、failover status/budget、WS 连接池生命周期、S283 模型映射/归一化或 S282 普通 OAuth retry window；非 429/非 WS 路径行为不变。
- 新增定向测试覆盖普通模型语义 429 忽略 header、Spark OAuth 语义 429 使用 header、握手 429 保留 header、API key/shadow 边界；既有 S282/S283、WS rate-limit 与 model transient 测试继续通过。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `571d1e1d9be93fea4e2d840f0e7d9570a442b263` (`fix(openai): isolate websocket semantic rate limits`)
- Local base: `bb3d3bca6` (S283 Spark quota model isolation)
- Upstream split owners `openai_gateway_passthrough.go`/`openai_ws_forwarder_support.go` do not exist locally; their behavior is consolidated in `openai_ws_forwarder.go` (`handleOpenAIWSTerminalTransientFailure`, `handleOpenAIWSErrorEventTransientFailure`, `persistOpenAIWSRateLimitSignal`). WS HTTP bridge and v2 passthrough already call these helpers.
- Handshake dial paths pass `responseBody=nil`; established-connection semantic events pass non-empty payload. This distinction must drive header isolation.

## Allowed Paths

- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_semantic_429_test.go` (新增)
- `docs/workflow/worker-results/upstream-v0185-ws-semantic-429-s284-result.md`
- `docs/workflow/qa-reports/upstream-v0185-ws-semantic-429-s284-qa.md`

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

- 不 cherry-pick、merge、rebase 或照搬上游拆分文件；只在本地 forwarder owner 手工适配。
- 语义事件只有在连接已建立且带非空 event payload 时才清除握手 headers；握手 HTTP 429 的 `responseBody=nil` 必须保留 headers。
- `openAIWSSemantic429Headers` 仅对 `isCodexSparkModel(model) && isOpenAIOAuthAccount(account)` 保留 headers；普通模型、API key、setup-token、Grok、Spark shadow 的 global 账号状态均不得因此升级为账号级 cooldown。Spark shadow 的模型级行为仍由 S283 负责。
- 复用 S283 的 `handleOpenAIAccountUpstreamError`、`HandleOpenAICodexSparkRateLimit` 与四参数 `SetModelRateLimit`，不改 repository/Ent/migration/接口。
- 不扩大到语义 WS header 之外的 quota 语义、不改客户端输出、failover、连接池或非 429 路径；保留保护脏文件与 `outputs/**`，aggregate dirty diff hash 必须保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`。
- 不执行真实 provider、数据库、容器、部署或 push。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/service -run 'Test(S284|OpenAIWSSemantic429|OpenAIWS.*RateLimit|S283|OpenAI.*Spark)' -count=10
go test ./internal/service
go test ./cmd/server -run '^$' -count=1
gofmt -d internal/service/openai_ws_forwarder.go internal/service/openai_ws_semantic_429_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/service
git diff --name-only --diff-filter=U
git status --short
```

Also inspect exact allowlist/denied-path audit and protected aggregate hash; `go test -tags unit ./internal/service` may expose unrelated baseline fixture/API drift and must be recorded without weakening tests.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-ws-semantic-429-s284-result.md`; first line must be `### DONE: upstream-v0185-ws-semantic-429-s284`, `### BLOCKED: ...` or `### FAILED: ...`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-ws-semantic-429-s284-qa.md` with first line `### PASS/FAIL/BLOCKED: upstream-v0185-ws-semantic-429-s284`.

## Stop Rules

- If implementation requires changes outside the listed forwarder/test paths, repository/Ent/migration/handler/frontend changes, or semantic behavior beyond header isolation, stop with `BLOCKED`.
- If any allowed target has unrelated concurrent edits or a protected baseline changes, stop without overwriting.
- Do not weaken tests for baseline unit-tag failures; do not modify denied paths or `outputs/**`.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
