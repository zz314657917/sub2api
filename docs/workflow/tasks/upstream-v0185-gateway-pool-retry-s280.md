---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-gateway-pool-retry-s280
worker_model: gpt-5.6-terra
base_commit: 817b8e0f28
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 Gateway Pool Same-Account Retry S280

## Task ID

`upstream-v0185-gateway-pool-retry-s280`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑手工适配上游 `b1e60ba45`：Anthropic 兼容转发的 Chat Completions 与 Responses 路径遇到可 failover 的上游错误时，保留 rate-limit 处理结果，并仅在账号仍可调度且池模式状态码允许时设置同账号重试。禁止整体 merge/cherry-pick。

## Success Criteria

- 两条兼容路径都接收 `HandleUpstreamError` 的 `shouldDisable` 返回值，并把 `mappedModel` 传给现有可变参数，保持模型级错误处理上下文。
- `UpstreamFailoverError.RetryableOnSameAccount` 仅在 `!shouldDisable && account.IsPoolMode() && account.IsPoolModeRetryableStatus(status)` 时为 true。
- Chat Completions 与 Responses 的默认池模式 429 都返回 failover error、保留响应体/状态且允许同账号重试；不得提前向客户端写错误体。
- 非池模式 429，以及显式空 `pool_mode_retry_status_codes` 的池账号，均不允许同账号重试。
- 不改变 `shouldFailoverUpstreamError`、默认重试码、重试次数、账号状态机、scheduler、response conversion、usage/billing 或其他网关路径。
- 定向测试重复 10 轮、完整 service 包、server 编译、gofmt、diff/conflict 和受保护脏文件哈希门禁通过。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `b1e60ba4535ea1fc2cc57a181a92bab03a2e0782`，包含在当前可见上游历史中。
- Local `RateLimitService.HandleUpstreamError` already returns `shouldDisable` and accepts optional `requestedModel`; do not change its API.
- Local `Account.IsPoolModeRetryableStatus` already preserves nil/default, explicit-empty and custom-list semantics; reuse it directly.
- Controller baseline: `HEAD=817b8e0f28`; aggregate protected dirty diff hash is `0e467987fd7aec5fc451983bdb8f8216f97ba69c`.

## Allowed Paths

- `backend/internal/service/gateway_forward_as_chat_completions.go`（仅 failover error block）
- `backend/internal/service/gateway_forward_as_responses.go`（仅 failover error block）
- `backend/internal/service/gateway_pool_mode_retry_test.go`
- `docs/workflow/worker-results/upstream-v0185-gateway-pool-retry-s280-result.md`
- `docs/workflow/qa-reports/upstream-v0185-gateway-pool-retry-s280-qa.md`

## Denied Paths

- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/account.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/repository/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/server/middleware/**`
- `frontend/**`
- `frontend/pnpm-lock.yaml`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/**`
- `docs/workflow/contract-reviews/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器、部署或数据文件

## Constraints

- 保持最小手工适配，不导入上游其他提交或抽象，不做无关重构、重命名或全文件格式化。
- `shouldDisable` 默认 false；存在 `rateLimitService` 时才以其返回值覆盖。不得为了测试构造新的生产 helper。
- 测试必须通过真实的两个公开转发方法验证 failover error；可复用现有 `queuedHTTPUpstreamStub`，不得复制生产状态机。
- 受保护文件及 `outputs/**` 不得修改、暂存、提交或清理。
- 不 commit、push、调用真实 provider、操作数据库、容器、部署或共享数据。

## Protected Baseline

- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`: `3EF121902FA6707467F9449F4B7F35EB30DBFC9EA9144B2E70EA7D89C46D6488`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`: `ECC2B47B60BA2D0BFA39631867164CFDB95EB8ED4033502401D367348B6C37F7`
- `backend/internal/pkg/apicompat/types.go`: `A0BAD7ABE0F6DCF0F6D5F515E1AA394A1FBC6495D6A608D62C8ADF1D64B3BDDC`
- `backend/internal/service/admin_service.go`: `858C0CFCEC7CB0AC27ED03694B3536C3AF5E2C695E5644A281A2955745CB3841`
- `frontend/pnpm-lock.yaml`: `8B545157E34CC0DDC1866A43B7147326B91549879EE6C3360F094DB300CE135E`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`: `4999C1582056C3BC5E1B15EECC1EEB1DCA945EB5BC00CB5D97D17A6248D8EBB7`

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/service -run '^TestGatewayCompatPoolModeRetry$' -count=10
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w internal/service/gateway_forward_as_chat_completions.go internal/service/gateway_forward_as_responses.go internal/service/gateway_pool_mode_retry_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/service/gateway_forward_as_chat_completions.go backend/internal/service/gateway_forward_as_responses.go backend/internal/service/gateway_pool_mode_retry_test.go
git diff --name-only --diff-filter=U
git status --short
```

Also inspect: only the two target failover blocks and focused test/report changed; both positive paths and both negative account configurations are asserted; all Protected Baseline SHA-256 values and aggregate dirty diff hash remain unchanged; no denied path is staged.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-gateway-pool-retry-s280-result.md` using `C:/Users/Administrator/.codex/templates/worker-result.md`.
- First line must be `### DONE: upstream-v0185-gateway-pool-retry-s280`, `### BLOCKED: upstream-v0185-gateway-pool-retry-s280` or `### FAILED: upstream-v0185-gateway-pool-retry-s280`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-gateway-pool-retry-s280-qa.md`.

## Stop Rules

- If implementation requires account/rate-limit/scheduler/repository/schema/migration/frontend changes, stop with `BLOCKED`.
- If any protected baseline changes, a target file gains unrelated edits, or concurrent work overlaps the two failover blocks, stop and report; do not overwrite it.
- Record failures truthfully; do not weaken assertions, expand allowed paths, commit or push.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
