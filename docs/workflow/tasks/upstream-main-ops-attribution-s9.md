# 上游合成 Sprint：upstream-main-ops-attribution-s9

## Summary

- Task ID: upstream-main-ops-attribution-s9
- Role: Developer Worker
- Branch: codex/upstream-main-ops-attribution-s9
- Worktree: E:/codex-worktrees/sub2api/upstream-main-ops-attribution-s9
- Baseline: local main=d6c7e4c69, upstream/main=635ad81cd
- Goal: 合成下一批 Ops 归因与 business-limited 标记小修，不直接 merge upstream/main。

## Scope

本轮只移植后端本地拒绝 / fast-policy / whitelist denial 的 Ops attribution 修复：

- OpenAI local gateway denial 标记为 client business-limited，而不是 provider/upstream error。
- OpenAI fast-policy 各兼容入口标记 business-limited。
- Antigravity whitelist denial 标记 business-limited。

不纳入 billing、stream hot path、OpenAI messages failover、大网关重构、frontend、Ent/migration、邮件、DingTalk、Channel Monitor API mode。

## Candidate Commits

按顺序 cherry-pick -x 或手工等价移植：

1. 5d7df678b fix(openai): mark local gateway denials business-limited
2. 9c56fe0b0 fix(openai): mark fast-policy entrypoints business-limited
3. 47fe90eab fix(antigravity): mark whitelist denials business-limited

已判定本轮记录为等价或延后：

- 03ae510c6: APPLIED_EQUIVALENT. Current baseline already excludes count_tokens from Ops metrics error counts and has coverage.
- 00eb3abbe / bd1e98ec2: APPLIED_EQUIVALENT. Current baseline already marks API key and Google group denials business-limited.
- a9c7a3a09: APPLIED_EQUIVALENT. Current baseline already strips Bedrock context_management when beta is absent.
- bf3787de1: APPLIED_EQUIVALENT. Current baseline already allows Claude Code count_tokens by UA.
- 8a999f438: APPLIED_EQUIVALENT. Current baseline already excludes terminal events from OpenAI WS first-token detection.
- 32ea9cfe1: APPLIED_EQUIVALENT. Current baseline already falls back to SSE body for API key responses.
- e9a2db8e8 / 8e27ff20a / 86d9b6bff / 1e2e8b1d6: DEFERRED. Higher behavior blast radius; keep for dedicated OpenAI stream/billing Sprint.

## Allowed Paths

- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_service_test.go
- backend/internal/service/openai_gateway_chat_completions.go
- backend/internal/service/openai_gateway_chat_completions_test.go
- backend/internal/service/openai_gateway_chat_completions_raw.go
- backend/internal/service/openai_gateway_messages.go
- backend/internal/service/openai_gateway_messages_test.go
- backend/internal/service/openai_ws_forwarder.go
- backend/internal/service/openai_ws_forwarder_test.go
- backend/internal/service/openai_ws_v2_passthrough_adapter.go
- backend/internal/service/antigravity_gateway_service.go
- backend/internal/service/antigravity_gateway_service_test.go
- backend/internal/service/ops_metrics_collector.go
- backend/internal/service/ops_metrics_collector_test.go
- docs/workflow/tasks/upstream-main-ops-attribution-s9.md
- docs/workflow/worker-results/upstream-main-ops-attribution-s9-result.md
- docs/workflow/qa-reports/upstream-main-ops-attribution-s9-qa.md
- docs/workflow/main-log.md

## Denied Paths

- frontend/**
- backend/ent/**
- backend/migrations/**
- deploy/**
- knowledge/**
- .github/**
- assets/**
- README*
- docs/workflow/status.md
- docs/workflow/spec.md

## Constraints

- 不新增数据库字段、migration、Ent schema、前端 API 或配置项。
- 不直接 merge upstream/main。
- 优先保留本地更完整实现；若上游提交已被等价吸收，记录 APPLIED_EQUIVALENT，不重复覆盖。
- 若某候选需要 denied paths、migration、frontend、billing 重算、stream hot path 重构或新公开 contract 字段，标记 DEFERRED 并停止该候选，不扩大 Sprint。
- 当前主工作区有用户未提交 knowledge 改动；实施必须停留在隔离 worktree。

## Public APIs / Interfaces

- 不新增公开 DTO 字段或配置项。
- 行为变化限定为 Ops attribution：
  - 本地策略拒绝、功能门禁、白名单拒绝不再被归为上游/provider 错误。
  - 这些错误在 Ops SLA / business-limited 统计中按 client business-limited 处理。

## Acceptance Commands

在 backend/ 目录执行：

```powershell
go test ./internal/service -run "Ops|BusinessLimited|FastPolicy|OpenAI|Antigravity|Whitelist|ImageGeneration|Passthrough" -count=1
go test ./internal/handler ./internal/server/middleware ./internal/service -run "Ops|SLA|BusinessLimited|Denied|Gateway|OpenAI|Antigravity" -count=1
go test ./internal/service ./internal/handler ./internal/server/middleware -count=1
```

基础检查：

```powershell
git status --short --branch
git diff --check
```

路径审计必须确认无 denied paths 改动。

## Output

- docs/workflow/worker-results/upstream-main-ops-attribution-s9-result.md
- docs/workflow/qa-reports/upstream-main-ops-attribution-s9-qa.md
- docs/workflow/main-log.md 追加 S9 记录

## Stop Rules

- 任何候选要求修改 denied paths，停止该候选并记录 DEFERRED。
- cherry-pick 冲突若涉及 billing/stream hot path 重构或本地产品线覆盖，停止该候选并记录 DEFERRED。
- 测试失败必须先归因；不能在未解释失败原因时合回 main。
