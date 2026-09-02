---
type: contract-review
scope: project
status: approved
task_id: upstream-v0185-ws-later-turn-429-s285
verdict: PASS
base_commit: 65bf61f5a
reviewer: final-evaluator
last_verified: 2026-09-01
---

### PASS: upstream-v0185-ws-later-turn-429-s285

# Contract Review

## Task ID

`upstream-v0185-ws-later-turn-429-s285`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-ws-later-turn-429-s285.md`

## Findings

- 上游 commit 将 handler、WS ingress、HTTP bridge 和 replay payload 拆在多个 owner；本 contract 已按本地 consolidated `openai_ws_forwarder.go` 及现有 S243 replay helpers 重新收敛，未要求照搬不存在的上游文件。
- later-turn 429 的 failover/replay 边界可测试：仅 HTTP bridge 且尚未写出下游数据才换号，retry payload 必须带累计上下文、去除旧 `previous_response_id` 并恢复客户端模型；工具上下文无法证明时 fail-close。
- 上游 commit 没有为直接上游 WS 构造跨账号完整输出历史，contract 已明确不扩大该路径。首轮 failover、已写出下游数据、非 429、Grok/CN 与非 WS 路径均保持现状；S282/S283/S284 限流 side effects 不在本 Sprint 重写。
- allowlist、denylist、保护 dirty hash 和 acceptance commands 明确，且未授权 repository/Ent/migration/frontend 变更。

## Gate Checks

- success_criteria_testable: `PASS`
- allowed_paths_explicit: `PASS`
- denied_paths_explicit: `PASS`
- acceptance_commands_executable: `PASS`
- worker_model_confirmed: `PASS` (`gpt-5.6-terra`)
- base_commit_confirmed: `PASS` (`65bf61f5a`)
- openspec_traceable: `not-applicable`

## Approval

Contract approved. Implementation may proceed only within the listed allowlist; all stop rules remain binding.
