---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 14:09 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-ops-realtime-stats-s47

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `3f2ef6046` so admin ops realtime account statistics avoid loading every account when platform/group filters are present, and so canceled realtime polling requests do not emit noisy API errors.

## Success Criteria
- Ops realtime handlers silently return when the request context is canceled or PostgreSQL reports `canceling statement due to user request`.
- Account repository exposes a stats-specific account listing path that selects only fields required by ops realtime stats, applies platform and group filters at query time, and preserves group metadata needed by the existing UI summaries.
- Ops concurrency and account availability stats call the stats-specific repository path when available, while preserving the paginated `ListWithFilters` fallback for mocks or alternate repositories.
- Group filtering is passed into account loading for both concurrency and availability stats; account/group/platform aggregation semantics remain compatible with the existing dashboard.
- No Ent schema, migrations, frontend, i18n, deploy, README, or product-page changes are introduced.

## Allowed Paths
- `backend/internal/handler/admin/ops_realtime_handler.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/service/ops_account_availability.go`
- `backend/internal/service/ops_concurrency.go`
- `backend/internal/service/ops_concurrency_test.go`
- `backend/internal/handler/admin/ops_realtime_handler_test.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `docs/workflow/tasks/upstream-main-v0143-ops-realtime-stats-s47.md`
- `docs/workflow/worker-results/upstream-main-v0143-ops-realtime-stats-s47-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-ops-realtime-stats-s47-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of `upstream/main` or the `v0.1.143` release branch.
- Do not add migrations or regenerate Ent.
- Do not reuse the reduced stats query outside ops realtime statistics.
- Preserve the existing repository interface by using an optional narrow interface assertion rather than adding methods to `service.AccountRepository`.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44
go test ./internal/service -run "TestListAllAccountsForOps|TestOps.*Concurrency|TestOps.*Availability" -count=1
go test ./internal/handler/admin -run "TestOpsRealtimeRequestCanceled|TestOpsRealtime|TestGetConcurrencyStats|TestGetAccountAvailability|TestGetRealtimeTrafficSummary" -count=1
go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListOpsAccountsForStats|TestAccountRepoSuite/TestListWithFilters" -count=1
git diff --check
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/|deploy/|knowledge/|\\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-ops-realtime-stats-s47`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-ops-realtime-stats-s47-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-ops-realtime-stats-s47-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if the reduced stats query omits fields that make existing ops account/group summaries materially incorrect.
- Stop if the fallback repository path no longer supports existing service tests or mocks.
- Stop if canceled-request suppression hides non-cancel operational errors.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 14:09 +08:00.
- Verdict: approved.
- Reason: upstream patch is a narrow admin ops performance/noise fix, has no database schema impact, and can be isolated from local product customizations.
