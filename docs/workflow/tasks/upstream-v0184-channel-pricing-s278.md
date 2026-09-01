---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0184-channel-pricing-s278
worker_model: gpt-5.6-sol
base_commit: f81bb2a55
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.184 Channel Pricing Normalization S278

## Task ID

`upstream-v0184-channel-pricing-s278`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `eb4237a2b`：渠道定价字面模型查找失败后，对已知 OpenAI/Codex 模型做归一化重试，让带 effort/date 后缀的请求继续使用渠道配置而非官方兜底价。禁止整段合并或 cherry-pick 上游历史。

## Success Criteria

- `ModelPricingResolver.Resolve` 和 token/channel override 路径均使用 literal-first 的 normalized channel lookup；请求 `gpt-5.6-luna-high`、已知日期后缀等在渠道仅配置 `gpt-5.6-luna` 时命中渠道价。
- 同时配置具体变体和基名时，具体变体字面价优先；未知 OpenAI 变体、非 OpenAI 模型和不相关模型不会误命中基名渠道价。
- 默认/订阅分组的实际 `RecordUsage` 计费回归证明输入成本选中渠道价，官方兜底价仅在无相关渠道配置时使用；不改变持久化字段、倍率、余额扣除或 billing 算法。
- 定向 Go 回归、受影响 service 包、`cmd/server` 编译、gofmt、diff/conflict 检查通过；既有 `-tags unit` 全包编译漂移若仍存在必须原样记录，不得修改无关测试。

## Context

- Repo: `F:\mcplugins\sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`
- Upstream provenance: `eb4237a2b`.
- Local owner: `backend/internal/service/model_pricing_resolver.go`.
- Existing usage-test helpers live in `backend/internal/service/openai_gateway_record_usage_test.go`; a focused regression may reuse them without changing that file.
- Current protected dirty paths: all existing `backend/**` edits outside the owner, `frontend/pnpm-lock.yaml`, Pixel Cafe, `outputs/**`, and workflow/knowledge files except the explicit evidence paths below.

## Allowed Paths

- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/model_pricing_resolver_channel_normalized_test.go`
- `docs/workflow/worker-results/upstream-v0184-channel-pricing-s278-result.md`
- `docs/workflow/qa-reports/upstream-v0184-channel-pricing-s278-qa.md`

## Denied Paths

- `backend/internal/service/admin_service.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/repository/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_route_breaker_test.go`
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
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、生产配置、容器、部署或数据文件

## Constraints

- `gpt-5.6-terra` Developer 连续两次在接口层失败（HTTP 524、HTTP 503）；用户已明确授权本 S278 的 Developer 与独立 QA 改用 `gpt-5.6-sol`。该例外不修改全局 Agent Matrix。
- S278 产品提交为 `43d109581`，实际父提交为并行会话的 `f81bb2a55`；后者的 17 个文件全部不属于 S278，不得修改、重提或作为本 Sprint 证据。QA 必须审 `43d109581^..43d109581` 加当前 allowlist follow-up diff，而不是把 `e5ff9b299..HEAD` 当作 S278 范围。
- literal lookup must run first; only retry when `normalizeKnownOpenAICodexModel` returns a different non-empty name.
- Do not normalize arbitrary strings, non-OpenAI models, channel mapping targets, or unrelated channel entries.
- Keep usage, billing, failover, provider request, persistence and multiplier semantics unchanged; only channel pricing selection may differ for the covered aliases.
- Do not alter `openai_gateway_record_usage_test.go`; reuse its helpers or create self-contained helpers in the allowlisted test file with unique names.
- Do not commit, push, call providers, touch database/container/deployment/shared data, or modify the current protected dirty paths.

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/service -run '^TestChannelPricing_' -count=10
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w internal/service/model_pricing_resolver.go internal/service/model_pricing_resolver_channel_normalized_test.go
```

If the full service package is attempted with `-tags unit`, record the existing compile baseline separately; it is not a substitute for the default-tag gates above.

From repo root:

```powershell
git diff --check -- backend/internal/service/model_pricing_resolver.go backend/internal/service/model_pricing_resolver_channel_normalized_test.go
git diff --name-only --diff-filter=U
```

Also inspect: changed paths are exactly the allowlist plus the worker report; literal-first and unrelated-model negative cases are present; protected dirty paths and `outputs/**` have unchanged status/hash; no provider/database/container/deployment/push action occurred.

## Output

- Write `docs/workflow/worker-results/upstream-v0184-channel-pricing-s278-result.md` using `C:/Users/Administrator/.codex/templates/worker-result.md`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0184-channel-pricing-s278-qa.md` using `C:/Users/Administrator/.codex/templates/qa-report.md`.
- First line must be `### DONE: upstream-v0184-channel-pricing-s278`, `### BLOCKED: upstream-v0184-channel-pricing-s278` or `### FAILED: upstream-v0184-channel-pricing-s278`.
- Report changed files, commands, test summaries, source provenance status, risks, contract compliance and `knowledge_candidates`.

## Stop Rules

- If the behavior requires billing algorithm, repository/schema/migration, provider, frontend or denied-path changes, stop with `BLOCKED`.
- If existing helpers cannot be reused without modifying denied files, create unique self-contained test helpers or stop; do not widen allowlist.
- Record all failures truthfully; do not delete tests, weaken assertions or absorb unrelated dirty edits.

## Budget

- worker_model: `gpt-5.6-sol`（用户授权的 S278-only 例外）
- qa_worker_model: `gpt-5.6-sol`（用户授权的 S278-only 例外）
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
