### PASS: upstream-v0177-group-usage-rollups-s222

# QA Report

## Task ID

`upstream-v0177-group-usage-rollups-s222`

## Verdict

`PASS`

## Scope And Boundary

- Independently reviewed `ba9415446a702fa677c2db59ff7b35878e1f0cd2..6ae2047338d7f944cb602aba66a96fb499f05a9b` in the isolated S222 worktree.
- The range contains 26 files and all 26 are in the approved allowlist; denied-path count is zero. No source, worker report, workflow status/main-log/task, knowledge, dependency, CI, migration outside 222/223, user dirty file, output, container, deployment, or push was changed by QA.
- The Developer report's frontend environment `FAIL` was not used as the verdict. QA independently used the approved S221 tool binaries through a temporary S222 junction; the S222 junction and build output were removed afterward.

## Acceptance Evidence

### Backend

- Exact service discovery: `go test ./internal/service -list <9-name-regex>` discovered `9/9`; the exact suite passed with `go test ./internal/service -run <9-name-regex> -count=10`.
- Exact repository discovery: `go test -tags=unit ./internal/repository -list <4-name-regex>` discovered `4/4`; the exact suite passed with `go test -tags=unit ./internal/repository -run <4-name-regex> -count=1`.
- Migration unit: `go test -tags=unit ./migrations -run '^TestMigration22(2|3)' -count=1` passed.
- Config timezone: `go test ./internal/config -run '^TestLoadStandardTZEnvironmentTakesPriority$' -count=10` passed.
- Complete packages/compile: `go test ./internal/service -count=1` (61.966s), `go test ./internal/handler -count=1` (27.356s), `go test ./internal/repository -count=1` (1.586s), `go test ./internal/server -count=1` (0.083s), and `go test ./cmd/server -run '^$' -count=0` passed.

### Frontend

- Direct approved S221 tools ran from the S222 frontend: `vitest.CMD run src/api/__tests__/admin.groups.usage-summary.spec.ts src/views/admin/__tests__/GroupsView.columnSettings.spec.ts` passed 2 files/2 tests.
- `vue-tsc.CMD --noEmit` passed.
- `vue-tsc.CMD -b` passed.
- `vite.CMD build --outDir E:/codex-tmp/sub2api-s222-terra-qa-vite-output --emptyOutDir` passed (`1873 modules transformed`, production output generated).
- Final cleanup checks: `S222 frontend/node_modules=False`, approved S221 `frontend/node_modules=True`, and QA build output absent. `frontend/pnpm-lock.yaml` git object hash is unchanged from the approval base (`f3dd93d599840b15970af95c44905f2cb103102f`).

### Fresh PostgreSQL Runtime Checklist

- Fixture: fresh task-owned `sub2api_s222_terraqa_20260816_1645` on `127.0.0.1:55432`, local `postgres/postgres`, using `E:/codex-tmp/sub2api-s220-postgres/runtime/pgsql/bin/psql.exe`. It was not `sub2api_s220` or `sub2api_s222_controller`.
- Migrations 222 then 223 applied successfully twice. Initial/default state read back as one row with `1970-01-01`, `Asia/Shanghai`, and epoch retained bound; second apply read back `state_rows=1|triggers=3`.
- With the configured PostgreSQL session timezone `America/New_York`, insert/update/delete/cascade/user-delete and batch cleanup each invalidated a published `2026-03-09` watermark to `2026-03-08`; cleanup ended with `remaining_logs=0`.
- Publication serialization: a transaction locked `usage_group_rollup_state`, rebuilt buckets, emitted `publication_buckets_rebuilt_before_watermark`, slept, then wrote the watermark last. A concurrent late historical insert took `2544ms`, completed after commit, and invalidated the watermark to `2026-08-13`.
- Rollup plus live tail: candidate summary SQL returned `1|total=7.0000000000|today=4.0000000000|yesterday=3.0000000000`.
- Timezone change rebuild: switching state from `America/New_York` to `Asia/Shanghai` rebuilt buckets and published `timezone_rebuild=Asia/Shanghai|watermark=2026-08-16|buckets=2026-08-14:3.0000000000,2026-08-15:4.0000000000`.
- DST today/yesterday: PostgreSQL returned `elapsed=23:00:00` for New York 2026-03-08 to 2026-03-09 and the DST summary returned total `7`, today `4`, yesterday `3`.
- Advisory locks: independent sessions returned `advisory_key=622101|peer=f|reacquire=t|holder_session_ended=true` and the same result for `622102`.
- Cleanup: active sessions before drop were `0`; exact `DROP DATABASE sub2api_s222_terraqa_20260816_1645` succeeded and the final existence query returned `database_exists_after_drop=f`.
- Docker-tagged integration was not used as evidence; the required fresh PostgreSQL fixture supplied the runtime evidence instead.

## Static Gates

- `git diff --check ba9415446..6ae204733`: PASS.
- Conflict/unmerged checks: `git diff --name-only --diff-filter=U` and `git ls-files -u` both returned zero entries.
- Worktree status at final QA check: clean.
- Upstream provenance: `cb7b03795`, `89d826be2`, and `45dcce0e4` each passed `git merge-base --is-ancestor <commit> upstream/main`.

## Findings

未发现明确问题。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

Not applicable; all required gates passed.

## Knowledge Promotion

`none`
