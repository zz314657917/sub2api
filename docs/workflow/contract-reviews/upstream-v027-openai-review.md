### PASS: upstream-v027-openai

# Contract Review

合同把 strict Chat role、manifest 最终 parse/validation 和取消后 affinity 持久化限制在
明确的五个 service 文件及证据文件。`bindHTTPResponseAccount` 的本地 owner 已确认在
`openai_gateway_service.go:6668`；合同明确只保留现有 `BindResponseAccount`/guard，
不引入本地没有的 response-owner 子系统。manifest 目标也限制为去除重复解析、保持
case-sensitive key 和 `[` 校验，不扩大为模型目录重构。

`TestV027OpenAI` 要求覆盖取消、值传播、deadline、nil/无效输入、localhost role payload、
未知 JSON 字段与数值精度，以及 manifest 的 null、错误类型、空数组和精确 key 场景。
`go test`、`go build`、格式、diff、冲突和路径审计均可独立执行。与 DeepSeek 合同对
`openai_gateway_service.go` 的写入明确串行；其他允许文件没有重叠所有者。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes` (`d6e6c34718e6eee6388391b346f40e1492e81d57`)
- topology_confirmed: `yes`

## Approval

允许独立 Terra Developer 和独立 Terra QA 在本合同 allowlist 内执行。不得与
`upstream-v027-deepseek` 同时修改 `openai_gateway_service.go`。
