### PASS: upstream-v027-cn-quota

# Contract Review

合同只适配 `db8692d67` 的 CN Coding Plan 配额耗尽 403 冷却。它正确指定既有
`cnAccountIsCodingPlan` 为只读判定来源，并将可修改范围固定在 CN 限流 helper、403
入口和一个聚焦测试文件。成功标准分别约束目标/非目标账户、普通及并发 403、未来/过期/
缺失快照、两类持久化失败和禁止 `SetError`，不会把通用 circuit breaker 语义交给
Developer 重构。

`TestV027CNQuota` 必须经 `HandleUpstreamError` 及 fake repository 运行，因此同时
验证分类与落库分支。验收命令包含聚焦测试、完整 backend build、格式、diff、冲突和
路径审计，且不需要数据库、容器或外部 provider。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes` (`d6e6c34718e6eee6388391b346f40e1492e81d57`)
- topology_confirmed: `yes`

## Approval

允许独立 Terra Developer 和独立 Terra QA 在本合同 allowlist 内执行。
