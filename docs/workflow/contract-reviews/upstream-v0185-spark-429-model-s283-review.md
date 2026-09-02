---
type: contract-review
scope: project
status: approved
task_id: upstream-v0185-spark-429-model-s283
verdict: PASS
base_commit: f48b4b77f
reviewer: final-evaluator
last_verified: 2026-09-01
---

### PASS: upstream-v0185-spark-429-model-s283

# Contract Review

## Task ID

`upstream-v0185-spark-429-model-s283`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-spark-429-model-s283.md`

## Findings

- 上游提交的拆分 owner 在本地不存在；contract 已将行为收敛到本地统一 OpenAI 错误入口、RateLimitService、WS forwarder/HTTP bridge/v2 passthrough 和既有模型限流存储 API。
- Spark OAuth 429 先于 S282 账号级 runtime 分类，命中后只写 `model_rate_limits`；同一账号的非 Spark 模型保持可调度，Spark shadow 的 global quota header 不会升级为账号级状态。
- contract 明确复用本地固定四参数 `SetModelRateLimit`，禁止修改 repository/Ent/migration；上游 variadic reason 不应移植。
- WS 调用点允许补 canonical/mapped model，覆盖标准 forwarder、ingress acquire/error、prewarm、HTTP bridge 和 v2 passthrough 握手/error event；客户端写出、failover budget、连接池和 header 隔离均冻结。
- 单体 HTTP/SSE passthrough 已由 S282 传递 requested model，本 Sprint 仅验证该依赖，不扩大 `openai_gateway_service.go` allowlist。
- API key、setup-token、非 OpenAI、非 Spark 429 与 S282 retry window 保持边界；未纳入 count-tokens 等直接错误路径被明确记录为刻意范围外。

## Gate Checks

- success_criteria_testable: `PASS`
- allowed_paths_explicit: `PASS`
- denied_paths_explicit: `PASS`
- acceptance_commands_executable: `PASS`
- worker_model_confirmed: `PASS` (`gpt-5.6-terra`)
- base_commit_confirmed: `PASS` (`f48b4b77f`)
- openspec_traceable: `not-applicable`

## Approval

Contract approved. Implementation may proceed only within the listed allowlist; all stop rules remain binding.
