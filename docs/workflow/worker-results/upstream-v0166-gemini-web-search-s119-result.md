### DONE: upstream-v0166-gemini-web-search-s119

# Worker Result

## Task ID
upstream-v0166-gemini-web-search-s119

## Status
done

## Summary

- Changed Gemini built-in search recognition to rely only on explicit tool
  types. Ordinary Chat Completions functions named `web_search` now remain
  client-side function declarations.
- Added an HTTP forwarding regression with `web_search` and `read_file` and
  retained the existing explicit `web_search_20250305` conversion coverage.

## Changed Files

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `docs/workflow/**` S119 evidence files

## Commands Run

```text
gofmt -w targeted Gemini files -> PASS
go test ./internal/service -run "^TestGeminiForwardAsChatCompletions_FunctionNamedWebSearchStaysClientSide$" -count=1 -> PASS (5.474s)
go test ./internal/service -run "TestGemini" -count=1 -> PASS (4.553s)
go test ./... -run "^$" -> PASS (38.5s)
gofmt -d targeted Gemini files -> PASS (no output)
git diff --check -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.474s
ok github.com/Wei-Shaw/sub2api/internal/service 4.553s
go test ./... -run "^$" completed successfully for all packages
```

## Risks

- No real Gemini upstream request or Hermes client session was run.
- No deployment, container update, or primary-worktree integration was run.

## Knowledge Candidates

- none

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason

- none
