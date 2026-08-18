### DONE: upstream-cn-provider-partial-usage-s229-e

# Controller Result

## Task ID

upstream-cn-provider-partial-usage-s229-e

## Status

done

## Summary

- Chat Completions, Responses, and Messages now submit non-failover partial results
  through the existing usage worker path before returning an error.
- Messages client-disconnect results are submitted before the disconnect return.
- Failover errors remain excluded by a shared helper, preserving retry/switch behavior.
- Existing request metadata, channel mapping, quota platform, session, and trial
  release context are reused for partial usage records.

## Changed Files

- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_partial_usage_contract_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-partial-usage-s229-e-result.md`

## Commands Run

```text
backend/go test ./internal/handler -run "TestShouldSubmitOpenAIPartialUsage|TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=10 -> PASS (5.528s)
backend/go test ./internal/handler -count=1 -> PASS (27.285s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (5.515s)
backend/gofmt -d internal/handler/openai_chat_completions.go internal/handler/openai_gateway_handler.go internal/handler/openai_partial_usage_contract_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 10c8b7020 upstream/main -> PASS
```

## Risks

- Validation uses handler package tests and local service dependencies only; real provider,
  Redis, database, container, deployment, and push operations are excluded by contract.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
