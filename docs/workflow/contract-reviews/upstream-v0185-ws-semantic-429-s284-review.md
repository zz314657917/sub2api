---
type: contract-review
scope: project
status: approved
task_id: upstream-v0185-ws-semantic-429-s284
verdict: PASS
base_commit: bb3d3bca6
reviewer: final-evaluator
last_verified: 2026-09-01
---

### PASS: upstream-v0185-ws-semantic-429-s284

# Contract Review

## Task ID

`upstream-v0185-ws-semantic-429-s284`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-ws-semantic-429-s284.md`

## Findings

- 上游拆分 owner 在本地不存在；contract 已将实现收敛到 `openai_ws_forwarder.go` 的三个统一 helper，覆盖标准 WS、HTTP bridge 与 v2 passthrough 的既有调用链。
- 以 `responseBody` 是否为空区分握手 HTTP 429 与已建立连接后的语义 429：语义事件普通模型清空握手 quota header，Spark OAuth 保留并交给 S283 模型级限流；握手 429 保留 header。
- 仅允许 header 隔离，不改变客户端写出顺序、failover、连接池、状态码、S282/S283 runtime 与模型归一化行为；repository/Ent/migration/handler/frontend 均拒绝。
- 验收覆盖普通模型、Spark OAuth、握手路径、API key/shadow 边界及既有 WS/S282/S283 回归；既有 `-tags unit` 基线漂移按证据记录。

## Gate Checks

- success_criteria_testable: `PASS`
- allowed_paths_explicit: `PASS`
- denied_paths_explicit: `PASS`
- acceptance_commands_executable: `PASS`
- worker_model_confirmed: `PASS` (`gpt-5.6-terra`)
- base_commit_confirmed: `PASS` (`bb3d3bca6`)
- openspec_traceable: `not-applicable`

## Approval

Contract approved. Implementation may proceed only within the listed allowlist; all stop rules remain binding.
