---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-oauth-429-quota-s282
worker_model: gpt-5.6-terra
base_commit: c886cdcac
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 OAuth 429 quota S282

## Task ID

`upstream-v0185-oauth-429-quota-s282`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `f1aadd48d`：区分 OpenAI OAuth 账号的普通瞬时 429 与 5h/7d/明确 reset quota exhausted 429。瞬时 429 在短期同账号重试窗口内不应落库 cooldown；明确配额耗尽应立即建立账号级运行时暂停并由现有限流服务持久化到 reset 时间。保持 Spark shadow、API key 和非 OpenAI 账号的既有语义。

## Success Criteria

- 新增可并发安全的 OpenAI/Grok 账号级运行时调度阻断；调度器的 `isOpenAIAccountRequestRuntimeBlocked` 同时尊重账号级阻断和既有 account-model transient 阻断；成功清限流时可清除运行时阻断。
- `classifyOpenAIOAuth429` 优先识别归一化 Codex 7d/5h 使用率达到 100%，其次识别 headers/body reset；普通瞬时 OAuth 429 在 retry window 内通过 `ShouldRetryOpenAIOAuth429` 让 `RateLimitService.handle429` 跳过持久化。
- quota/reset 429 立即暂停账号至有效 reset（缺失 reset 时使用既有秒级 fallback），并让 `RateLimitService` 继续执行现有 quota snapshot、`SetRateLimited` 与 body reset 逻辑；不得把 Spark shadow 的 global `/responses` quota 信号写入账号级状态。
- OpenAI 网关的 HTTP 429 入口在调用 `HandleUpstreamError` 前先分类/更新 runtime state；现有 failover 构造可获得普通 OAuth 429 的同账号重试标记，不改变容量、overload、API key transient 或非 429 行为。
- Grok 临时暂停同步到账号级运行时阻断，以修复现有 scheduler/runtime 测试的 owner 缺口；不改变 Grok 持久化冷却时长和 reason。
- 新增/扩展定向测试覆盖：transient 不写 cooldown、5h、7d、body reset 立即暂停；API key、shadow、非 quota 429 不误命中；runtime block 清除/最长 cooldown 保留；现有 model transient 测试继续通过。
- 运行定向 service 测试（必要时重复）、完整 service、server compile、gofmt、diff/conflict、保护脏文件哈希；完整 service 若受既有 unit-tag/fixture 漂移阻塞需据实记录。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `f1aadd48d5ca295cb1babc5e65412af2f46d1247`。
- 上游拆分文件在本地不存在；本地 owner 是 `openai_images_responses.go` 的统一错误入口、`ratelimit_service.go` 的 429 处理、`openai_account_model_transient.go` 的调度阻断扩展、`openai_gateway_service.go` 的服务状态字段、`openai_ws_forwarder.go` 的 WS 429 信号和 `openai_gateway_grok.go` 的 Grok 临时暂停。
- Codex header parser/Normalize、`calculateOpenAI429ResetTime`、`parseOpenAIRateLimitResetTime`、既有 model transient 与 OAuth 429 storm 统计均复用，不重复实现或重命名接口。
- Base commit: `c886cdcac`（S281 quota reset cooldown 已本地提交）。

## Allowed Paths

- `backend/internal/service/openai_account_runtime_block_fastpath.go`（新增：分类、账号级 runtime blocker、OAuth retry window）
- `backend/internal/service/openai_gateway_service.go`（仅新增/初始化账号级 runtime state 字段）
- `backend/internal/service/openai_images_responses.go`（统一 OpenAI 账号错误入口的 429 调用顺序；必要的 OAuth retry failover 标记）
- `backend/internal/service/openai_account_model_transient.go`（仅合并账号级阻断到请求兼容判断）
- `backend/internal/service/ratelimit_service.go`（runtime blocker hook 与 OAuth transient 429 持久化短路）
- `backend/internal/service/openai_gateway_grok.go`（仅将既有 Grok temp pause 镜像到 runtime blocker）
- `backend/internal/service/openai_ws_forwarder.go`（仅 WS 429 信号入口先更新 OAuth runtime state）
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_embeddings.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_oauth_429_failover.go`
- `backend/internal/service/openai_oauth_429_runtime_test.go`（新增定向测试）
- `docs/workflow/worker-results/upstream-v0185-oauth-429-quota-s282-result.md`
- `docs/workflow/qa-reports/upstream-v0185-oauth-429-quota-s282-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/openai_ws_v2/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `backend/internal/pkg/apicompat/**`
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
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器、部署或数据文件

## Constraints

- 不 cherry-pick、merge、rebase 或照搬上游拆分文件；按本地 owner 手工适配，保持既有公开接口和 test doubles 兼容。
- `classifyOpenAIOAuth429` 只对 OpenAI OAuth 且非 shadow 生效；不把 Spark shadow 的 global Codex headers 当作 Spark quota。语义流 HTTP 200/header 隔离属于后续 S284，本 Sprint 不扩大处理。
- runtime block 只影响进程内调度选择，不替代数据库状态；有效 reset 时沿用现有 `SetRateLimited`，清除操作不得缩短更长的既有 block。
- transient retry window 不得修改普通 429 fallback、Anthropic/CN/Gemini、API key、pool-mode 或 model transient 语义；没有有效 reset 的 quota signal 仍必须 fallback pause，避免热循环。
- 不新增数据库字段、迁移、provider/真实 WebSocket 调用、容器、部署或 push。
- 所有现有保护脏文件、`outputs/**` 和其 aggregate dirty diff hash 必须保持不变。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/service -run 'Test(OpenAIOAuth429|OpenAIAccountRuntime|HandleGrokAccountUpstreamError|OpenAIModelTransient)' -count=10
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w internal/service/openai_account_runtime_block_fastpath.go internal/service/openai_gateway_service.go internal/service/openai_images_responses.go internal/service/openai_account_model_transient.go internal/service/ratelimit_service.go internal/service/openai_gateway_grok.go internal/service/openai_ws_forwarder.go internal/service/openai_gateway_chat_completions.go internal/service/openai_gateway_chat_completions_raw.go internal/service/openai_gateway_messages.go internal/service/openai_gateway_responses_chat_fallback.go internal/service/openai_embeddings.go internal/service/openai_images.go internal/service/openai_oauth_429_runtime_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/service
git diff --name-only --diff-filter=U
git status --short
```

Also inspect: no denied path changes; protected dirty paths and aggregate hash remain unchanged; direct 429 failover constructors only gain the narrowly scoped transient OAuth retry metadata.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-oauth-429-quota-s282-result.md` using the worker-result template.
- First line must be `### DONE: upstream-v0185-oauth-429-quota-s282`, `### BLOCKED: upstream-v0185-oauth-429-quota-s282` or `### FAILED: upstream-v0185-oauth-429-quota-s282`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-oauth-429-quota-s282-qa.md`.

## Stop Rules

- If implementation requires handler/Ent/migration/frontend changes or semantic WS header isolation, stop with `BLOCKED` and report the topology boundary.
- If any allowed target has unrelated concurrent edits or protected baseline changes, stop without overwriting.
- Do not weaken tests to accommodate baseline fixture drift or unit-tag compile conflicts.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
