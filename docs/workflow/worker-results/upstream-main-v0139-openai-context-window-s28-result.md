### DONE: upstream-main-v0139-openai-context-window-s28

# Worker Result

## Task ID
upstream-main-v0139-openai-context-window-s28

## Status
done

## Summary
- Ported OpenAI context-window error classification from upstream `7cbf82ed6` into the local OpenAI gateway.
- Prevented HTTP and stream `response.failed` context-window errors from becoming `UpstreamFailoverError`.
- Preserved failover behavior for transient failures such as `server_is_overloaded`.
- Skipped upstream `openai_account_runtime_block_fastpath.go` hunk because this local repo has no matching file or compiled entry.

## Changed Files
- `backend/internal/service/error_passthrough_runtime_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-main-v0139-openai-context-window-s28.md`

## Commands Run
```text
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_chat_completions.go backend/internal/service/openai_gateway_chat_completions_test.go backend/internal/service/error_passthrough_runtime_test.go backend/internal/service/openai_gateway_service_codex_cli_only_test.go backend/internal/service/openai_gateway_service_test.go -> pass
go test ./internal/service -run "TestIsOpenAIContextWindowError|TestShouldFailoverOpenAIUpstreamResponseContextWindow502|TestOpenAIHandleErrorResponse_ContextWindow502KeepsMessageWithoutFailover|TestForwardAsChatCompletions_BufferedContextWindowResponseFailedReturnsErrorWithoutFailover|TestForwardAsChatCompletions_BufferedTransientResponseFailedTriggersFailover|TestForwardAsChatCompletions_StreamContextWindowResponseFailedReturnsErrorWithoutFailover|TestOpenAIStreamingContextWindowResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover" -count=1 -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/service	5.854s
```

## Risks
- Live OpenAI OAuth/APIKey traffic was not exercised; verification is unit-level and handler-path-level.
- The upstream fastpath hunk was intentionally not ported because local code does not have the corresponding file.
- Existing unrelated dirty proxy/account/frontend/knowledge files remain outside this sprint.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
