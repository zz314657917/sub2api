### DONE: upstream-anthropic-grok-usage-s66b

# Worker Result

## Task ID

`upstream-anthropic-grok-usage-s66b`

## Status

`done`

## Summary

- 适配上游 `0d28f7f90`、`83f169e4f`：Anthropic -> Responses 非流/流式输出保留 `cache_creation_input_tokens`，Responses -> Anthropic 非流/流式输出从总输入中扣除并保留 cache creation tokens。
- 适配上游 `0fa1eb85e`：Grok Responses 结果复用共享 `extractOpenAIReasoningEffortFromBody`，兼容嵌套 `reasoning.effort` 与扁平 `reasoning_effort`。
- 适配上游 `5a0dd510e`：删除改造后未使用的 `ptrStringOrNil` helper。
- 未修改计费公式、模型映射、请求体字段、路由、重试或其他 worker 文件。

## Changed Files

- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- `backend/internal/pkg/apicompat/responses_anthropic_cache_creation_test.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `docs/workflow/worker-results/upstream-anthropic-grok-usage-s66b-result.md`

## Commands Run

```text
go test ./internal/pkg/apicompat -run "Test.*CacheCreation|Test.*Anthropic.*Usage|Test.*Responses.*Anthropic" -count=1 -> PASS
go test ./internal/service -run "Test.*Grok.*ReasoningEffort|TestForwardGrokResponsesStreaming" -count=1 -> PASS package compile; [no tests to run] because Grok test file has //go:build unit
go test ./internal/service -run "^TestExtractOpenAIReasoningEffortFromBody$" -count=1 -v -> PASS, including flat reasoning_effort case
go test -tags unit ./internal/service -run "Test.*Grok.*ReasoningEffort|TestForwardGrokResponsesStreaming" -count=1 -v -> BLOCKED by pre-existing unit-suite compile drift
git diff --check -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.816s
ok github.com/Wei-Shaw/sub2api/internal/service 5.512s [no tests to run]
--- PASS: TestExtractOpenAIReasoningEffortFromBody
PASS
ok github.com/Wei-Shaw/sub2api/internal/service 0.073s
```

The supplemental `-tags unit` run did not compile because existing unit tests reference stale APIs: duplicate `stringPtr`, obsolete billing method signatures/fields, and removed Grok runtime-block members. Those files are outside Allowed Paths and were not changed.

## Risks

- 新增的 Grok forward 断言未独立执行；默认 service package 编译和共享 extractor 的扁平 `reasoning_effort` 非 unit 测试已通过。
- 未运行完整 backend test suite；本 contract 仅要求定向 apicompat/Grok 验收。

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `partial`（实现完成；Grok unit-tag 新断言受既有全包编译漂移影响未独立执行）
- stop_rules_triggered: `no`

## Blocked Reason

- None for implementation. Supplemental unit-tag verification is blocked by pre-existing compile drift outside Allowed Paths.
