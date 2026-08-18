### DONE: upstream-cn-provider-responses-drain-s229-d

# Controller Result

## Task ID

upstream-cn-provider-responses-drain-s229-d

## Status

done

## Summary

- Responses to native-Anthropic streaming now continue draining upstream SSE after the
  first downstream write failure.
- The conversion state machine continues to process all events, preserving `message_start`
  input usage and terminal `message_delta` output usage.
- Finalize events are emitted only while the client remains connected and use the same
  tool-name restoration path as regular events.
- Existing interval timeout and normal connected-stream behavior remain unchanged.

## Changed Files

- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/openai_gateway_anthropic_native_pump_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-responses-drain-s229-d-result.md`

## Commands Run

```text
backend/go test ./internal/service -run "TestResponsesStreamingFromNativeAnthropic_ClientDisconnectDrainsUsage|TestResponsesStreamingFromNativeAnthropic_HangTimesOut|TestResponsesStreamingFromNativeAnthropic_HappyPathStillConverts" -count=10 -> PASS (15.811s)
backend/go test ./internal/service -count=1 -> PASS (65.382s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (5.506s)
backend/gofmt -d internal/service/openai_gateway_responses_anthropic_native.go internal/service/openai_gateway_anthropic_native_pump_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 10c8b7020 upstream/main -> PASS
```

## Risks

- Validation uses local pipe/failing-writer fixtures only; real provider, Redis, database,
  container, deployment, and push operations are excluded by contract.
- An initial gofmt probe used worktree-root-relative paths and failed before execution;
  the corrected command ran from `backend/` and passed.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
