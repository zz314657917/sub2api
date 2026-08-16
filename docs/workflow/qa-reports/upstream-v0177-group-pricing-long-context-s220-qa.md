### PASS: upstream-v0177-group-pricing-long-context-s220

# Independent QA Report

## Scope

- Reviewed implementation through `be3d0026a298eb7668318d385091b0a3931db09e` on
  `pge/upstream-v0177-group-pricing-long-context-s220`.
- Used the authoritative contract on `main@1c23c7b9f`, including Amendments 7
  and 8. The isolated worktree copy ends at Amendment 6 and is not authoritative.
- Reviewed the complete task range from `1aaf92ad8` through the target commit.

## Result

No product defect found. The change stays within the amended allowlist: group
model pricing and typed UI, group long-context default/override behavior,
OpenAI-only account veto, Grok/non-OpenAI exclusion, usage audit persistence
and exposure, migration 220/221, and video resolution/continuous-seconds
pricing were all present in the reviewed range. The denied EditAccountModal
paths, `outputs/`, and lockfiles are absent from the task diff.

## Acceptance Evidence

- `go generate ./ent`: PASS; worktree remained clean afterward.
- Focused service discovery: all five contract test names discovered. The same
  five tests passed with `go test ./internal/service -run <focused> -count=10`.
- `go test ./migrations -run '^(TestMigration220|TestMigration221|TestOpenAILongContextBillingMigration)' -count=1`: PASS.
- Full backend QA execution: `go test ./internal/service -count=1`: exit 0,
  `61.604s`; `go test ./internal/handler -count=1`: exit 0, `27.520s`;
  repository: PASS (`1.708s`); server: PASS (`0.097s`); `cmd/server`
  compilation: PASS (`0.113s`).
- Frontend focused Vitest: 5 files, 24 tests PASS. `pnpm.cmd run typecheck`:
  PASS. Production Vite build: exit 0, `22.79s`; only existing Browserslist
  and chunk-size warnings were emitted.
- Provenance: all seven required upstream commits are ancestors of
  `upstream/main`; `git diff --check` passed; no conflict markers or unmerged
  index entries.

## Disposable PostgreSQL Proof

Docker engine was unavailable. QA created only the fresh database
`sub2api_s220_qa` on the authorized local PostgreSQL 17.5 endpoint
`127.0.0.1:55432`, applied migration 220 twice, then precisely dropped it.
Direct checks passed: legacy missing/string/numeric OpenAI values became
`false`; boolean `true` remained true; non-OpenAI data was unchanged; a new
OpenAI row defaulted to false; an update omitting the field preserved true; a
malformed write returned SQLSTATE `22023`; `usage_logs.long_context_billing_applied`
is `NOT NULL DEFAULT false`; and second execution was idempotent.

## Finding

The original untagged repository acceptance command returned `[no tests to run]`
because the declared integration test has `//go:build integration`. Amendment
8 explicitly rejects that result as evidence. It is replaced here by the
direct disposable PostgreSQL migration proof above. This is a contract test
discoverability issue, not a product failure.

## Cleanup

The temporary database, QA runner, and QA log files were removed. `pnpm` changed
`frontend/pnpm-lock.yaml` mechanically; it was restored exactly to `HEAD`.
