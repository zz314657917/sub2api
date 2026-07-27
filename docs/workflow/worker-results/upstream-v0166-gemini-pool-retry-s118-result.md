### DONE: upstream-v0166-gemini-pool-retry-s118

# Worker Result

## Task ID
upstream-v0166-gemini-pool-retry-s118

## Status
done

## Summary
- Added Gemini pool-mode skipped-policy failover handling for HTTP messages and
  native messages, and preserved configured same-account retry eligibility for
  chat-completions failover errors.
- Pool 429 can retry the same account under the existing policy; pool 500
  still fails over but only gains that flag when 500 is explicitly configured.
  Pool 400 and non-pool behavior remain unchanged.

## Changed Files
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_pool_retry_test.go`
- `docs/workflow/**` P/G/E evidence files

## Commands Run
```text
gofmt -w internal/service/gemini_chat_completions_compat_service.go internal/service/gemini_messages_compat_service.go internal/service/gemini_pool_retry_test.go -> PASS
go test ./internal/service -run "TestGeminiPoolMode" -count=1 -> PASS (5.502s)
go test ./... -run "^$" -> PASS (57s)
gofmt -d targeted Gemini files -> PASS (no output)
git diff --check -> PASS
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.502s
go test ./... -run "^$" completed successfully for all packages
```

## Risks
- No real Gemini upstream request or live same-account retry loop was run.
- The known unit-tag aggregate compile drift remains outside this contract; the
  approved default-tag regression and default repository compile gates pass.

## Knowledge Candidates
- none

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- none
