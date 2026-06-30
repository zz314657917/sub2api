### DONE: upstream-main-v0139-responses-stream-output-s29

# Worker Result

## Task ID
upstream-main-v0139-responses-stream-output-s29

## Status
done

## Summary
- Ported OpenAI Responses streaming terminal output normalization from upstream `e9a2db8e80`.
- Accumulated streaming output deltas and image outputs during `handleStreamingResponse`.
- Normalized terminal `response.output:null` to reconstructed output, or to `[]` when no output was accumulated.
- Kept existing non-empty terminal output unchanged and preserved model replacement after normalization.
- Confirmed adjacent S29 candidates are already local-equivalent and skipped them.

## Changed Files
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-main-v0139-responses-stream-output-s29.md`

## Commands Run
```text
gofmt -w internal/service/openai_gateway_service.go internal/service/openai_gateway_service_test.go -> pass
go test ./internal/service -run "TestOpenAIStreamingNormalizesTerminalOutputFromDeltas|TestOpenAIStreamingNormalizesTerminalOutputToEmptyArray|TestOpenAIStreamingPreambleKeepaliveUsesDownstreamIdle|TestOpenAIStreamingPolicyResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient" -count=1 -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/service	7.484s
```

## Risks
- Live OpenAI Responses streaming traffic was not exercised; verification is unit-level around local streaming handling.
- Passthrough streaming path was not changed in S29 because upstream `e9a2db8e80` only targeted the local `handleStreamingResponse` path.
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
