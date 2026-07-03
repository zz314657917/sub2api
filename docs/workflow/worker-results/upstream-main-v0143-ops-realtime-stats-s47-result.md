### DONE: upstream-main-v0143-ops-realtime-stats-s47

## Summary

- Ported upstream `3f2ef6046` into the isolated S47 worktree.
- Added `ListOpsAccountsForStats` as an ops-only reduced account query with platform and group filters applied in the DB query.
- Updated ops concurrency and account availability stats to use the stats-specific repository path when available, while keeping the paginated `ListWithFilters` fallback for alternate repositories/mocks.
- Added canceled realtime request suppression in admin ops realtime handlers.

## Changed Files

- `backend/internal/handler/admin/ops_realtime_handler.go`
- `backend/internal/handler/admin/ops_realtime_handler_test.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `backend/internal/service/ops_account_availability.go`
- `backend/internal/service/ops_concurrency.go`
- `backend/internal/service/ops_concurrency_test.go`
- `docs/workflow/tasks/upstream-main-v0143-ops-realtime-stats-s47.md`
- `docs/workflow/worker-results/upstream-main-v0143-ops-realtime-stats-s47-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-ops-realtime-stats-s47-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run

```powershell
go test ./internal/service -run "TestListAllAccountsForOps|TestOps.*Concurrency|TestOps.*Availability" -count=1
go test ./internal/handler/admin -run "TestOpsRealtimeRequestCanceled|TestOpsRealtime|TestGetConcurrencyStats|TestGetAccountAvailability|TestGetRealtimeTrafficSummary" -count=1
go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListOpsAccountsForStats|TestAccountRepoSuite/TestListWithFilters" -count=1
git diff --check
```

## Test Output

- `internal/service`: PASS
- `internal/handler/admin`: PASS
- `internal/repository` integration subset: PASS
- `git diff --check`: PASS, with existing line-ending warnings for workflow docs only.

## Risks

- `ListOpsAccountsForStats` intentionally selects a reduced account column set. It should remain scoped to ops realtime statistics and not be reused for general account management views.
- Group filter is now applied at load time. This reduces DB and concurrency load for filtered dashboards, but multi-group accounts still expose their loaded groups for display and aggregation according to existing `accountsToService` behavior.

## Knowledge Candidates

- None.
