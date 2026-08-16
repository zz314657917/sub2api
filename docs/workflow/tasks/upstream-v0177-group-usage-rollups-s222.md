---
task_id: upstream-v0177-group-usage-rollups-s222
phase: pre-reviewed-pending-s221
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Group Usage Daily Rollups S222

## Task ID

upstream-v0177-group-usage-rollups-s222

## Role

You are the independent `gpt-5.6-terra` Generator worker. Execute only the
approved contract after S220/S221 integration; do not expand scope.

## Goal

Behaviorally port group usage daily rollups from upstream `cb7b03795` plus the
timezone/test corrections in `89d826be2` and `45dcce0e4`. Replace full-history
group summary scans with persistent closed-day rollups plus the live retained
tail, using the server-configured timezone.

## Success Criteria

- Migrations 222/223 create daily rollup and state tables, publication
  watermark, retained lower bound, configured timezone, and concurrency-safe
  invalidation triggers for insert/update/delete and historical cleanup.
- Startup performs a bounded 30-minute backfill under an independent PostgreSQL
  advisory leader lock. Scheduled aggregation synchronizes rollups under a
  distinct advisory leader lock after every cycle, including early dashboard
  watermark returns. Each lock is held on a dedicated connection, released
  before that connection closes, and automatically released on process or
  connection failure.
- Rollup publication is atomic and watermark-last. A timezone change, future or
  invalid watermark, recompute, historical mutation, partition cleanup, or
  retained-history deletion invalidates/rebuilds the required range.
- Group summary returns today, yesterday, and total using server timezone and
  DST-safe natural-day boundaries. Browser timezone query parameters are no
  longer sent or trusted.
- Admin Groups UI renders yesterday between today and total without regressing
  the S220 group-pricing controls.
- Follow-up fixes save/restore the test's prior timezone. Go version,
  dependencies, lockfiles, CI/release/security workflows are unchanged.
- Database evidence uses only fresh task-owned disposable PostgreSQL fixtures;
  no shared or production DB is touched.

## Context

- Repo: `F:/mcplugins/sub2api`.
- Product base: S220/S221-integrated `main` approval commit supplied at dispatch.
- Upstream source commits: `cb7b03795daa7e569fe2f5618207af742ab3da60`,
  `89d826be2979b975400b4424897c5ab665a4fb90`, and
  `45dcce0e49b979c4eeea9002b844dc5e923bb681`.
- The upstream `usage_log_repo_trend.go` is not present locally; its summary
  method lives in `backend/internal/repository/usage_log_repo.go`.
- User authorization covers migrations 222/223 source integration and
  disposable validation, not execution against shared or production databases.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/repository/custom_group_usage_rollup_repo.go`
- `backend/internal/repository/custom_group_usage_timezone_test.go`
- `backend/internal/repository/dashboard_aggregation_group_usage_test.go`
- `backend/internal/repository/dashboard_aggregation_repo.go`
- `backend/internal/repository/group_usage_rollup_trigger_integration_test.go`
- `backend/internal/repository/usage_cleanup_repo.go`
- `backend/internal/repository/usage_cleanup_repo_test.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_group_summary_test.go`
- `backend/internal/service/custom_group_usage_rollup.go`
- `backend/internal/service/custom_group_usage_rollup_test.go`
- `backend/internal/service/dashboard_aggregation_service.go`
- `backend/internal/service/dashboard_aggregation_service_test.go`
- `backend/internal/service/dashboard_service.go`
- `backend/migrations/222_group_usage_daily_rollups.sql`
- `backend/migrations/223_group_usage_rollup_timezone.sql`
- `backend/migrations/group_usage_rollup_migration_test.go`
- `frontend/src/api/__tests__/admin.groups.usage-summary.spec.ts`
- `frontend/src/api/admin/groups.ts`
- `frontend/src/i18n/locales/en/admin/groups.ts`
- `frontend/src/i18n/locales/zh/admin/groups.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.columnSettings.spec.ts`
- `docs/workflow/worker-results/upstream-v0177-group-usage-rollups-s222-result.md`

## Denied Paths

- `.github/**`, `backend/go.mod`, `backend/go.sum`,
  `frontend/pnpm-lock.yaml`, dependency/toolchain updates, VERSION, and unrelated
  workflow/release/security files from `89d826be2`.
- User account modal/test changes, fingerprint behavior, unrelated schemas,
  providers, shared/production DB, containers, deployment, push, and `outputs/**`.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, QA reports,
  `knowledge/**`, and global memories.

## Constraints

- Work only in the isolated S222 worktree at the approved commit.
- Adapt split upstream repository files to local `usage_log_repo.go` and preserve
  newer local timezone, dashboard, cleanup, pricing, and S220 Groups UI behavior.
- Migration files are forward-only and checksum-immutable after integration.
  Validate them before main integration; never run them against user/shared data.
- Integration tests must own and clean up their exact PostgreSQL container or
  fresh task-specific database and test timezone state. Do not broadly stop
  Docker or other processes. A Docker-unavailable skip or `[no tests to run]`
  is not acceptance evidence.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-usage-rollups-s222/backend
go test -tags=unit ./migrations -run '^TestMigration22(2|3)' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 migration tests failed' }
go test ./internal/config -run '^TestLoadStandardTZEnvironmentTakesPriority$' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S222 TZ environment precedence failed' }

$serviceNames = @(
  'TestGroupUsageDateUsesConfiguredTimezoneBoundary',
  'TestGroupUsageParseDateUsesConfiguredTimezone',
  'TestGroupUsageYesterdayStartHandlesDST',
  'TestDashboardAggregationService_RunScheduledAggregationSyncsGroupUsageRollups',
  'TestDashboardAggregationService_RunScheduledAggregationSyncsGroupAfterDashboardEarlyReturn',
  'TestDashboardAggregationService_StartupGroupSyncUsesIndependentAdvisoryLeaderLock',
  'TestDashboardAggregationService_RunScheduledAggregationUsesAdvisoryLeaderLock',
  'TestDashboardAggregationService_GroupUsageSyncSkipsPeerHeldAdvisoryLock',
  'TestDashboardAggregationService_GroupUsageSyncSkipsAdvisoryLockError'
)
$service = '^(' + ($serviceNames -join '|') + ')$'
$serviceList = @(go test ./internal/service -list $service)
if ($LASTEXITCODE -ne 0) { throw 'S222 focused service discovery failed' }
$missingService = @($serviceNames | Where-Object { $serviceList -notcontains $_ })
if ($missingService.Count -ne 0) { throw "missing S222 service tests: $($missingService -join ', ')" }
go test ./internal/service -run $service -count=10
if ($LASTEXITCODE -ne 0) { throw 'S222 focused service failed' }

$repositoryNames = @(
  'TestDashboardAggregationRepositorySyncGroupUsageRollups',
  'TestGroupUsageSummary',
  'TestUsageCleanupRepositoryDeleteUsageLogsBatch',
  'TestUsageLogRepositoryGetAllGroupUsageSummaryUsesRollupTail'
)
$repository = '^(' + ($repositoryNames -join '|') + ')$'
$repositoryList = @(go test -tags=unit ./internal/repository -list $repository)
if ($LASTEXITCODE -ne 0) { throw 'S222 focused repository discovery failed' }
$missingRepository = @($repositoryNames | Where-Object { $repositoryList -notcontains $_ })
if ($missingRepository.Count -ne 0) { throw "missing S222 repository tests: $($missingRepository -join ', ')" }
go test -tags=unit ./internal/repository -run $repository -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 focused repository failed' }
docker info *> $null
if ($LASTEXITCODE -eq 0) {
  go test -tags=integration ./internal/repository -run '^(TestGroupUsageRollupTrigger|TestGroupUsageSummary)' -count=1
  if ($LASTEXITCODE -ne 0) { throw 'S222 trigger integration failed' }
} else {
  Write-Host 'Docker unavailable: run the mandatory fresh PostgreSQL checklist below; a skipped Go integration suite is not evidence.'
}
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 complete service failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 complete handler failed' }
go test ./internal/repository -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 complete repository failed' }
go test ./internal/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S222 complete server failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S222 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-usage-rollups-s222/frontend
pnpm.cmd exec vitest run src/api/__tests__/admin.groups.usage-summary.spec.ts src/views/admin/__tests__/GroupsView.columnSettings.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S222 frontend focused failed' }
pnpm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S222 frontend typecheck failed' }
pnpm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S222 frontend build failed' }

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-usage-rollups-s222
git diff --check
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S222 conflict or unmerged index found'
}
foreach ($commit in @('cb7b03795','89d826be2','45dcce0e4')) {
  git merge-base --is-ancestor $commit upstream/main
  if ($LASTEXITCODE -ne 0) { throw "missing upstream provenance: $commit" }
}
```

## Output

- Write
  `docs/workflow/worker-results/upstream-v0177-group-usage-rollups-s222-result.md`
  with the required first-line verdict.
- Commit only allowed S222 source/tests/report. Include exact migration fixture,
  commands, timezone cleanup evidence, changed files, risks, and compliance.
- When Docker is unavailable, both Developer and independent QA must use their
  own fresh PostgreSQL database and record evidence for: migrations 222/223
  first apply and second-run idempotency; initial/default state; insert, update,
  delete, cascaded historical delete, and cleanup invalidation; late historical
  insert versus publication serialization; watermark-last publication;
  timezone change rebuild; DST-safe today/yesterday boundaries; rollup plus live
  tail summary; two independent database connections proving startup and
  scheduled advisory-lock peer exclusion plus reacquisition after release; and
  exact database deletion after the run.

## Stop Rules

- Stop if implementation requires dependency/CI/toolchain changes, a shared or
  production DB, unrelated migration/schema, or weakening transaction/trigger
  concurrency semantics.
- Stop if S220 group-pricing UI or S221/user account-modal behavior would be
  overwritten instead of adapted.
- Stop after two failed implementation rounds; do not integrate, push, deploy,
  update containers, or clean worktrees/branches.

## Contract Review

`PASS / topology and acceptance pre-review` (2026-08-16 14:36 +08:00): local
migration slots 222/223 are free; the split upstream group-summary owner maps to
`backend/internal/repository/usage_log_repo.go`; dashboard aggregation,
cleanup, handler, and Groups UI owners match the allowlist. Upstream migration
and repository focused tests use `unit` tags, while trigger/DST/concurrency
tests use `integration`; acceptance now invokes those tags explicitly. Docker
is currently unavailable on the controller host, so an integration-suite skip
cannot pass: Developer and QA must independently execute the documented fresh
PostgreSQL behavior checklist. Final approval still waits for S221 integration
and must recheck the S220 Groups UI plus S221/user account-modal boundaries.

`PASS / Amendment 1` (2026-08-16 15:39 +08:00): the upstream locale files
`frontend/src/i18n/locales/{en,zh}/admin/overview.ts` do not exist in this
checkout. The local Groups page already reads `admin.groups.usageToday` and
`admin.groups.usageTotal` from
`frontend/src/i18n/locales/{en,zh}/admin/groups.ts`; replace the two nonexistent
allowlist paths with those two actual locale owners so the required yesterday
label remains bilingual. No other frontend, source, test, migration,
dependency, database, deployment, or push boundary changes.

`PASS / Amendment 2` (2026-08-16 15:44 +08:00): this checkout does not contain
the upstream generic Redis/Wire `tryAcquireSingletonLeaderLock`, `lockCache`,
`db`, or `instanceID` prerequisite chain. Do not weaken S222 to ungated
multi-replica sync and do not import that unrelated global framework. Instead,
within the existing dashboard service/repository allowlist, expose an optional
PostgreSQL advisory-lock capability backed by a dedicated `*sql.Conn` using
`pg_try_advisory_lock`, deferred `pg_advisory_unlock`, and connection close.
Startup and scheduled group-rollup sync use distinct stable lock IDs; a peer-held
lock or acquisition error skips that sync fail-closed, while repositories without
the optional capability retain unit-test/single-instance ungated behavior. The
task contexts bound runtime, and PostgreSQL releases a lock on process or
connection failure, so no Redis TTL is required. Add focused service coverage
and fresh-PostgreSQL two-connection exclusion/reacquisition evidence. No new
dependency, Wire, shared lock service, provider, deployment, push, or database
authorization is granted.

`PASS / Amendment 3` (2026-08-16 16:03 +08:00): controller review proved the
original focused repository command could exit zero while discovering only one
of four required tests, and the draft omitted a focused regression for advisory
lock acquisition errors plus the standard `TZ` precedence change. Add exact
service/repository `-list` discovery gates, require peer-held and acquisition
error fail-closed service tests, and require
`TestLoadStandardTZEnvironmentTakesPriority` in the config package. The two
already-allowlisted frontend test files must be created and executed; their
absence is not a topology exemption. These are acceptance and test-coverage
corrections only and do not expand product, migration, dependency, database,
deployment, or push scope.

## Controller Review R2

`PASS / advance-to-independent-QA` (2026-08-16 16:34 +08:00): Developer fix
`ea91e490a` and controller-only formatting commit `6ae204733` satisfy the
amended source and acceptance boundaries. The worker report intentionally
remains `### FAIL` because its deleted local dependency directory could not
resolve `vite/client`; controller direct invocation of the preserved S221
toolchain passed both focused Vitest files, `vue-tsc --noEmit`, `vue-tsc -b`,
and Vite build, then removed the exact S222 junction. Focused service discovery
is 9/9, repository discovery is 4/4, config TZ x10 and complete service,
handler, repository, server, and server compile pass. An independent controller
database `sub2api_s222_controller` passed two migration rounds, insert/update/
delete/cascade invalidation, timezone rebuild, 23-hour DST boundaries, rollup
plus live-tail `8/5/3`, startup/scheduled advisory exclusion and reacquisition,
and late-write serialization (`1567ms`, final watermark `2026-08-13`), then was
exactly deleted. Next legal action is a separate `gpt-5.6-terra` QA that creates
and deletes its own database and commits only the QA report.

## Independent QA And Main Integration

`PASS / independent-qa` (2026-08-16 17:06 +08:00): Terra QA commit
`23a6dc75a` contains only
`docs/workflow/qa-reports/upstream-v0177-group-usage-rollups-s222-qa.md`.
It independently reviewed all 26 allowlisted files, reproduced focused and
complete backend/frontend checks, and used then dropped its own
`sub2api_s222_terraqa_20260816_1645` database. The report records two migration
rounds, mutation/cascade/cleanup invalidation, publication-last late-write
serialization, rollup-tail, timezone rebuild, DST, advisory 622101/622102,
and `database_exists_after_drop=f`.

`PASS / main-integration` (2026-08-16 17:21 +08:00): S222 source/fix/style and
QA commits were cherry-picked to main as `6131972c2`, `ec85d1c3f`, `b6ad86460`,
and `f02ac091a`. Main full Go packages and correct-cwd frontend focused tests,
typecheck, and build passed; user account-modal files and `outputs/` remain
outside the S222 commits.
