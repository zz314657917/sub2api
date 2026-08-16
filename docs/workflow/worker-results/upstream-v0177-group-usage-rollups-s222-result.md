### FAIL: upstream-v0177-group-usage-rollups-s222

## R1 Result

R1 corrects rollup lifecycle ownership: ordinary `AggregateRange` no longer
locks, invalidates, or publishes group rollups; `RecomputeRange` owns those
transactional actions. `CleanupUsageLogs` now only invalidates/cleans, while
the scheduled service defer publishes under the advisory lock. The concrete
PostgreSQL repository implements that optional lock; peer-held and acquisition
errors skip synchronization, while unit-only fakes remain direct.

The R1 implementation, backend checks, fresh PostgreSQL checklist, and two
new frontend Vitest tests passed. The report is FAIL solely because the required
frontend typecheck cannot resolve `vite/client` after the mandated deletion of
this worktree's `frontend/node_modules`; no dependency directory or lockfile
was recreated to mask the failure.

## Scope

- Contract baseline: `ba9415446`.
- Amendment 1: `6b57fe8ba` replaces the nonexistent `admin/overview.ts` locale
  owners with the actual `admin/groups.ts` files; no other locale path changed.
- Amendment 2: `e2b7ef4b8` requires PostgreSQL advisory locking for production
  group-rollup startup and scheduled synchronization.
- Only the contract allowlist and Amendment 1 locale substitutions were changed.
  `frontend/pnpm-lock.yaml` was restored to HEAD and generated
  `frontend/pnpm-workspace.yaml` was removed; neither is in the final diff.

## Implementation

- Added forward migrations 222/223 for daily group rollups, state watermark,
  retained lower bound, timezone metadata, and insert/update/delete triggers.
- Replaced the local monolithic group summary scan with closed daily rollups plus
  the live tail. Today and yesterday use the configured server timezone and
  natural-day/DST-safe boundaries. The endpoint and frontend no longer send or
  trust browser timezone values.
- Added watermark invalidation around recompute, task cleanup, batch retention
  cleanup, and partition deletion. Publication rebuilds buckets before updating
  the watermark/state row.
- Startup and scheduled sync paths use distinct advisory lock IDs (`622101` and
  `622102`). The concrete PostgreSQL repository holds the lock on a dedicated
  connection and releases/close it in a deferred callback; peer-held and
  acquisition-error paths do not run synchronization. Unit fakes without the
  optional lock capability retain their pre-existing direct behavior.
- Added Yesterday to the Groups usage display while preserving the existing
  S220 pricing controls.

## Verification

- `go test -tags=unit ./migrations -run '^TestMigration22(2|3)' -count=1`: PASS.
- `go test ./internal/repository -count=1`: PASS.
- Focused service timezone/DST, startup/scheduled advisory-lock, peer-skip, and
  dashboard-early-return sync tests: PASS with `-count=10`.
- `go test ./internal/service ./internal/repository ./internal/handler -run '^$' -count=1`: PASS (compile checks).
- `git diff --check`: PASS.
- R1 focused service suite, including timezone/DST, startup/scheduled advisory
  locking, early return, peer-held, and acquisition-error behavior: PASS with
  `-count=10`.
- R1 repository discovery (4/4): PASS.
- `TestLoadStandardTZEnvironmentTakesPriority`: PASS.
- Task-owned direct frontend tools (a temporary junction to the approved S221
  `node_modules`) ran the two allowlisted Vitest files: PASS (2 files / 2
  tests). The lockfile SHA-256 remained
  `47961DDE09DEF2FBD378C8D7C139C144DE02AEC00148FB5BAAA1D7ECED7AAC2D`.
  Direct `vue-tsc.CMD -b` against the S222 tsconfig failed with TS2688 because
  the intentionally absent local node_modules cannot resolve `vite/client`.
  The S222 node_modules path was then exactly deleted; final check is false,
  while the S221 toolchain path remains true.

## Fresh PostgreSQL Evidence

All database checks used only `sub2api_s222_dev_terra` at `127.0.0.1:55432`
with the local `postgres` administrative role. The database was created fresh
for each fixture and precisely dropped after each run; final checks returned
`database_exists_after_drop=f`.

- Migration 222 then 223 applied successfully; applying both again succeeded
  idempotently. Initial state was `1970-01-01|Asia/Shanghai|1`.
- With session timezone `America/New_York`, a late historical insert at the DST
  boundary invalidated the publication watermark to `2026-03-07`.
- Insert, update, delete, and cascaded user deletion each invalidated a
  published `2026-03-09` watermark to `2026-03-08` for the affected historical
  date (`insert/update/delete/cascade=2026-03-08`).
- Two independent PostgreSQL connections proved advisory exclusion:
  holder confirmed, peer `pg_try_advisory_lock(622101)=f`, and reacquisition
  after release returned `t`.
- R1 fixture repeated 222/223 first and second application, observed initial
  `1970-01-01|Asia/Shanghai|0`, and recorded insert/update/delete/cascade
  cleanup invalidation to `2026-03-08`.
- A transaction holding `usage_group_rollup_state FOR UPDATE` blocked a late
  historical INSERT. After publish-side release it completed and invalidated
  the watermark to `2026-08-13`. The retained tail query produced total `7`,
  today `4`, yesterday `3`. Startup lock `622101` and scheduled lock `622102`
  each independently returned peer `f` and post-release reacquisition `t`.

## Remaining Risk

- Docker is unavailable, so the repository's Docker-tagged integration suite
  was not run. Fresh PostgreSQL manual evidence covers migrations, default
  state, mutation/cascade cleanup, late-write serialization, tail and both
  advisory lock keys; application-level timezone-rebuild and watermark-last
  publication still require a passing integration harness.
- The fresh PostgreSQL fixture exercised migrations, trigger invalidation,
  timezone/DST behavior, advisory-lock exclusion/release, and cleanup. Full
  application-level rollup publication/tail-query concurrency remains for the
  independent QA worker to execute against its own disposable database.
