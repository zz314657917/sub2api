### DONE: upstream-usage-breakdown-legacy-request-type-s73

# Worker Result

## Task ID

`upstream-usage-breakdown-legacy-request-type-s73`

## Status

`done`

## Summary

- Added an alias-aware legacy request-type filter while preserving the existing unaliased SQL through delegation.
- Updated `GetUserBreakdownStats` to use the `ul` alias for Sync, Stream, and WS v2 legacy fallback rows.
- Added untagged helper and SQLMock matrices covering non-empty seven-column scans, dimensions, RequestType+Stream, ordering, limits, and ordinary-breakdown inclusion.
- The configured `deepseek-v4-pro` worker returned model 404, so this implementation was completed by the authorized collaboration-agent fallback.

## Changed Files

- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `docs/workflow/worker-results/upstream-usage-breakdown-legacy-request-type-s73-result.md`

## Commands Run

```text
gofmt -w internal/repository/usage_log_repo.go internal/repository/usage_log_repo_request_type_test.go -> PASS
go test ./internal/repository -list <S73 required pattern> -> PASS, 2/2 discovered
go test ./internal/repository -run <S73 required pattern> -count=1 -> PASS
go test ./internal/repository -run <S73 request-type regression pattern> -count=1 -> PASS
go test ./internal/repository -list ^TestUsageLogRepositoryGetUserLeaderboard -> PASS, 5 discovered
go test ./internal/repository -run ^TestUsageLogRepositoryGetUserLeaderboard -count=1 -> PASS
go test ./internal/repository -run ^$ -count=1 -> PASS (compile-only)
clean committed worktree gate -> PASS
merge-base allowed-path and production-hunk audits -> PASS
git diff --check <base>..HEAD -> PASS
full contract Acceptance Commands -> PASS
```

## Test Output

```text
required discovery: 2/2
required tests: ok github.com/Wei-Shaw/sub2api/internal/repository
request-type regressions: ok github.com/Wei-Shaw/sub2api/internal/repository
leaderboard discovery: 5
leaderboard regressions: ok github.com/Wei-Shaw/sub2api/internal/repository
repository compile-only: ok github.com/Wei-Shaw/sub2api/internal/repository [no tests to run]
ACCEPTANCE_PASS base=64d2b0b7cea88da25f2f0eec52141e6036685f1f tests=2 leaderboard=5
```

## Risks

- No real PostgreSQL instance was used; query behavior is verified with SQLMock and exact captured SQL fragments.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
