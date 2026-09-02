---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-ws-later-turn-429-s285
worker_model: gpt-5.6-terra
base_commit: 65bf61f5a
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 WebSocket later-turn 429 failover replay S285

## Task ID

`upstream-v0185-ws-later-turn-429-s285`

## Role

你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑手工适配上游 `82cbe6aff`：WebSocket HTTP bridge 已完成的后续 turn 在上游 HTTP/SSE 429 且尚未向客户端写出任何数据时，换号并重放当前 turn；重放必须携带足够的累计 input 上下文、移除旧 `previous_response_id`，并重新走新账号的模型映射。直接上游 WS 没有可用于跨账号重建的完整输出历史，本 Sprint 保持其现状。

## Success Criteria

- WS HTTP bridge 的后续 HTTP 429，及连接后尚未写出下游数据的 SSE `error`/`response.failed` 429，返回带 current-turn retry payload 的可识别 failover；handler 换号前使用该 payload，而不是重新发送首轮消息。
- 首轮、直接上游 WS、非 429 与已写出下游数据的错误行为保持不变。
- retry payload 合并已确认的历史输入与当前 turn 输入，移除 `previous_response_id`，恢复客户端模型标识后在新账号上重新映射；payload 必须是副本，不能被后续修改污染。
- 携带 `function_call_output` 但缺少对应工具调用上下文或 `item_reference` 无法安全重建时，不切换账号，fail-close 并保持客户端连接安全；不得丢失工具续链语义。
- 不改变非 429 failover、首轮 failover、客户端写出顺序、billing/usage、session sticky、连接池生命周期、S282/S283/S284 限流处理；Grok/CN 与非 WS 路径行为不变。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `82cbe6aff7d963e5096d26df73671236a707ad24` (`fix(openai): resume later websocket turns after 429`)
- Local base: `65bf61f5a` (S284 semantic 429 header isolation)
- 上游拆分的 `openai_ws_forwarder_ingress.go`/`openai_ws_forwarder_payload.go` 在本地不存在；本地 owner 是 `backend/internal/service/openai_ws_forwarder.go`。HTTP bridge owner 是 `openai_ws_http_bridge.go`；handler owner 是 `openai_gateway_handler.go`。
- 本地已有 S243 replay helpers（`buildOpenAIWSReplayInputSequence`、工具上下文分析）和 S284 429 处理；本 Sprint 只为 HTTP bridge 增加跨账号 current-turn payload 传递及后续 429 failover。

## Allowed Paths

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_ws_failover_resume_test.go`（新增）
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_http_bridge_resume_test.go`（新增）
- `backend/internal/service/openai_ws_forwarder_resume_test.go`（新增）
- `docs/workflow/worker-results/upstream-v0185-ws-later-turn-429-s285-result.md`
- `docs/workflow/qa-reports/upstream-v0185-ws-later-turn-429-s285-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/internal/service/admin_service.go`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/handler/**`（仅允许本 contract 列出的 handler 文件）
- `backend/internal/service/**`（仅允许本 contract 列出的 service 文件）
- `frontend/**`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/**`（本 contract 文件除外）
- `docs/workflow/contract-reviews/**`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器或数据文件

## Constraints

- 不 cherry-pick、merge、rebase 或照搬上游拆分文件；只在本地 owner 手工适配。
- 仅对 HTTP bridge 后续 turn 的 429 且尚未写出下游数据触发跨账号 failover/replay；首轮、直接上游 WS 和非 429 行为沿用现有策略。
- retry payload 必须由服务层构造并返回副本；不得把上游已映射模型直接当作客户端首包，避免 handler 二次映射错误。
- 任何 `function_call_output` 必须有可验证的历史工具调用上下文或完整 `item_reference` 覆盖；否则不构造重放 payload，直接 fail-close。
- 不通过修改 repository/Ent/migration/API 接口实现；不改变 S282/S283/S284 的账号/模型限流 side effects。
- 保留所有保护脏文件与 `outputs/**`，六个受保护文件 aggregate dirty diff hash 必须保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`。
- 不执行真实 provider、数据库、容器、部署或 push。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/handler -run '^TestOpenAIWSNextAttemptMessage' -count=10
go test ./internal/service -run 'Test(OpenAIWSCurrentTurn|OpenAIWSHTTPBridge.*LaterTurn|OpenAIWS.*429|S284|S283)' -count=10
go test ./internal/service
go test ./internal/handler
go test ./cmd/server -run '^$' -count=1
gofmt -d internal/handler/openai_gateway_handler.go internal/handler/openai_gateway_ws_failover_resume_test.go internal/service/openai_gateway_service.go internal/service/openai_ws_forwarder.go internal/service/openai_ws_http_bridge.go internal/service/openai_ws_http_bridge_resume_test.go internal/service/openai_ws_forwarder_resume_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/handler backend/internal/service
git diff --name-only --diff-filter=U
git status --short
```

Also inspect exact allowlist/denied-path audit and protected aggregate hash; `go test -tags unit ./internal/service` may expose unrelated baseline fixture/API drift and must be recorded without weakening tests.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-ws-later-turn-429-s285-result.md`; first line must be `### DONE: upstream-v0185-ws-later-turn-429-s285`, `### BLOCKED: ...` or `### FAILED: ...`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-ws-later-turn-429-s285-qa.md` with first line `### PASS/FAIL/BLOCKED: upstream-v0185-ws-later-turn-429-s285`.

## Stop Rules

- If implementation requires changes outside the listed handler/service/test paths, repository/Ent/migration/frontend changes, direct-upstream-WS replay, or semantic changes beyond HTTP-bridge later-turn 429 replay, stop with `BLOCKED`.
- If safe current-turn payload cannot be proven due missing tool context, do not guess; fail-close and report the case.
- If any allowed target has unrelated concurrent edits or a protected baseline changes, stop without overwriting.
- Do not weaken tests for baseline unit-tag failures; do not modify denied paths or `outputs/**`.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
