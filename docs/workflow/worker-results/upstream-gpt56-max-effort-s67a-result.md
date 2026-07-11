### DONE: upstream-gpt56-max-effort-s67a

# Worker Result

## Task ID

`upstream-gpt56-max-effort-s67a`

## Status

`done`

## Summary

- 适配上游 `80b3d4c1f`：GPT-5.6 Sol/Terra/Luna（含 provider、alias 和后缀形式）的显式或后缀 `max` 保持为 `max`，其他模型继续归一为 `xhigh`。
- OpenAI OAuth `/responses/compact` 的 GPT-5.6 `reasoning.effort=max` 定向降级为 `xhigh`；普通 Responses、API Key compact 和其他平台 OAuth 不变。
- 适配上游 `c3ae5fc3c`：effort 提取支持有序模型候选，HTTP、raw Chat、WS v2、WS ingress 和 WS HTTP bridge 在映射/规范化后仍可从原请求后缀恢复 metadata。
- 适配上游 `b9b013a08`：WS passthrough 首帧及多轮 `response.create` 将映射后模型作为首候选。
- 未修改 S67b 独占的 messages/fallback 文件，也未修改价格、计费公式、路由或重试行为。

## Changed Files

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_model_alias.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/service/openai_gpt56_max_test.go`
- `backend/internal/service/openai_reasoning_effort_candidates_test.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `backend/internal/service/openai_model_alias_test.go`
- `docs/workflow/worker-results/upstream-gpt56-max-effort-s67a-result.md`

## Commands Run

```text
go test ./internal/service -run "Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|TestExtractOpenAIReasoningEffortFromBody|Test.*WSPassthrough.*Effort" -count=1 -> PASS
go test ./internal/service -run "TestOpenAIGatewayServiceForward.*GPT56|TestNormalizeOpenAICodexCompactReasoningEffort" -count=1 -> PASS
go test ./internal/service -run "TestForwardAsRawChatCompletionsGPT56MappedMaxEffort" -count=1 -> PASS
go test ./internal/service -run "^$" -count=1 -> PASS package compile
go test -tags=unit ./internal/service -run "TestForwardAsRawChatCompletions_PreservesMappedGPT56MaxEffort" -count=1 -> BLOCKED by pre-existing unit-suite compile drift; test was moved to an untagged S67a file and passed above
git diff --check -> PASS
allowed-path audit -> PASS
S67b ownership audit -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.654s
ok github.com/Wei-Shaw/sub2api/internal/service 5.626s
ok github.com/Wei-Shaw/sub2api/internal/service 5.743s
```

## Risks

- 未连接真实 OpenAI/Codex 上游验证 GPT-5.6 compact；HTTP/WS 行为由本地 recorder 和 metadata 单元测试覆盖。
- 完整 `-tags=unit` service 套件仍被既有重复 `stringPtr`、旧 billing API 和已移除 Grok runtime-block 测试引用阻断；这些文件不在 Allowed Paths，本任务的 raw Chat 回归已迁至默认测试并通过。

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
