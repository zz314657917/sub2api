---
type: contract-review
scope: repository
status: approved
task_id: cf3577a3c-behavior-port
verdict: PASS
base_commit: c1833f66c5ca5777ae0d11c0645d5e433ba1502e
reviewer: evaluator
last_verified: 2026-09-12
---

### PASS: cf3577a3c-behavior-port

# Contract Review

## Verdict

`PASS` — 可进入隔离 worktree 的行为级实现；不得将本结论当作最终 QA 或发布批准。

## Findings

- `cf3577a3c-coverage.md` 已逐条列出上游全部 41 个路径（24 个生产、17 个测试），并将每项对应到本地 owner、专用回归测试或明确 `not-applicable` 状态。基线 HEAD 已核对为合同的 `c1833f66c5ca5777ae0d11c0645d5e433ba1502e`。
- `openai_alpha_search.go` 与其测试明确为 `not-applicable`：本地没有 alpha/search gateway route。不得因上游仅新增 `store` sanitizer 而导入此前不存在的约 671 行依赖链。
- 产品 allowlist 已改为精确路径，禁止通配符；同时明确禁止 upstream split gateway（forward/passthrough/request_body/response_handling/upstream_errors）、split WS ingress/v2 和 alpha/search 的整体导入。五个当前不存在但被精确允许的新 helper 文件，只能是本地最小 helper，不能复制上游拆分层或替换单体 gateway owner。
- 合同明确要求回归保护 `allowed_tools` 对象及子级工具选择，并保留 GPT-6 Astra 的 `reasoning.mode=pro`；上游旧模型的 `pro -> max` 归一化不得覆盖这一更晚本地语义。
- 精确的 backend build、真实生产源文件加 `cfport_*_test.go` 的 HTTP/WS fake-transport 命令、headers 测试、gofmt、diff/conflict 检查均已列明。整体服务测试的既知 fixture 漂移须单独报告，不能替代 mock runtime 验收。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`（实现新增测试文件后执行）
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes`
- openspec_traceable: `not-applicable`
- upstream_41_path_inventory_complete: `yes`
- local_Astra_allowed_tools_preservation_proven: `contract-required; pending implementation QA`

## Approval

可派发 Generator。实现中若需要不在精确 allowlist 的路径、schema/migration/config/provider 依赖，或只有整体搬运上游 split 文件才能实现，必须停止并重新审约。最终 QA 必须实际验证 Astra/`allowed_tools`、HTTP、bridge、WebSocket 的 alias/ID/retry/Ops 行为，且不得把未完成路径标为等价。
