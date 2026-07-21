### DONE: multi-group-route-fail-closed-s88

# Worker Result

## Task ID

`multi-group-route-fail-closed-s88`

## Status

`done`

## Summary

- Model-aware multi-group routing now validates the default group's platform,
  routing scope, and enabled explicit model rules before using it as fallback.
- An incompatible fallback resolves to `nil`; middleware rejects it with HTTP
  403 and code `NO_MATCHING_GROUP_ROUTE` before refreshing subscription context
  or continuing toward account scheduling.
- Compatible defaults, matching configured routes, single-group keys, and the
  pre-body routing pass retain their existing behavior.

## Changed Files

- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_routing_s88_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s88_test.go`
- S88 workflow spec, contract, status, log, result, QA, and current-task files.

## Commands Run

```text
gofmt -w <four S88 Go paths> -> PASS
go test ./internal/service ./internal/server/middleware -list "^TestS88" -> PASS, 9 tests discovered
go test ./internal/service -run "^TestS88" -count=10 -> PASS
go test ./internal/server/middleware -run "^TestS88ResolveAPIKeyForModelRequest" -count=10 -> PASS
go test ./internal/service -run "^(TestAPIKeyResolveForRequest|TestAPIKeyResolveForModelRequest)" -count=1 -> PASS
go test ./internal/handler ./internal/server/routes -run "^$" -> PASS
go test ./internal/service -run "^TestPeakMultiplier" -count=1 -> PASS
go test ./internal/service ./internal/server/middleware -count=1 -> middleware PASS; service baseline peak-rate failures only
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service (S88 count=10)
ok github.com/Wei-Shaw/sub2api/internal/server/middleware (S88 count=10)
ok github.com/Wei-Shaw/sub2api/internal/handler [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/server/routes [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/service (isolated TestPeakMultiplier*)
```

## Risks

- The aggregate service package still fails seven existing
  `group_peak_rate_test.go` assertions after another test changes global
  timezone state. The same failures were reproduced in a clean worktree at
  baseline `96021f068`; the peak tests pass in isolation.
- No live API request, push, deployment, or container update was performed.

## Knowledge Candidates

- None. The change is task-specific and is recorded in workflow/current-task
  evidence only.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `partial` (all S88 behavior gates pass; aggregate
  service command retains a proven baseline-only failure)
- stop_rules_triggered: `no`

## Blocked Reason

- None.
