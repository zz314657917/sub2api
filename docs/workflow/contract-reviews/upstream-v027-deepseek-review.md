### PASS: upstream-v027-deepseek

# Contract Review

合同将范围限定为 `881ab1b0c` 与 `acc05620c` 的 DeepSeek Responses 请求重写。
本地实际 owner 已固定为 `openai_gateway_service.go:7354`，apicompat helper 和两个
新增测试路径均在 allowlist 内。成功标准明确保留 `store=false`、
`previous_response_id` 删除、Anthropic 既有行为及非目标账户行为，并要求通过
`httptest` 捕获实际 localhost 转发体覆盖并行 tool output、图片后移、提示项、精度和
no-op。

`TestV027DeepSeek` 前缀、apicompat 全包回归和 `go build ./...` 可独立执行；diff、
冲突、格式和精确路径审计已列为 QA gate。与 OpenAI 合同共享单体文件的写入被明确规定
为串行，故不存在并发 ownership 歧义。

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
`upstream-v027-openai` 同时修改 `openai_gateway_service.go`。
