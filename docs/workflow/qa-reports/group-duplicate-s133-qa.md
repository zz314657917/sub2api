### PASS: group-duplicate-s133

# S133 Administrator Group Duplication QA

## Findings

- No clear implementation defect or contract drift was found. The new action
  copies the persisted group configuration into an inactive independent group,
  clones eligible account bindings and priorities atomically with the scheduler
  outbox event, and uses an internal operation identity for ambiguous retry
  recovery.
- A test-only cleanup removed the unused `testing` import from
  `backend/internal/repository/gateway_cache_live_test.go`, which unblocked the
  repository integration package.
- The full `unit` API-contract suite remains stale outside S133: its expected
  group/settings payloads omit current fields such as `allow_live` and newer
  settings. The duplicate endpoint itself is covered by the focused admin
  handler regression.
- The rollback integration test now reaches a PostgreSQL instance and passes.

## Executed Checks

- `go generate ./ent`: PASS.
- `go generate ./cmd/server`: PASS.
- `go test ./internal/service -run "Test.*DuplicateGroup" -count=1`: PASS.
- `go test ./internal/handler/admin -run "Test.*Duplicate" -count=1`: PASS.
- `go test ./internal/repository -run "Test.*Group.*Duplicate" -count=1`:
  PASS / compile only; the integration test is tag-gated and no default-tag
  repository test matched.
- `go test -tags=integration ./internal/repository -run '^TestCreateGroupFromSourceRollsBackWhenOutboxInsertFails$' -count=1`:
  PASS.
- `go test ./... -run "^$"` and `go build ./...`: PASS.
- `corepack.cmd pnpm --dir frontend exec vitest run ...`: PASS, 2 files and
  9 tests.
- `corepack.cmd pnpm --dir frontend run typecheck` and `run build`: PASS.
- `gofmt` on every changed Go file, migration schema/index review,
  `git diff --check`, staged diff check, unmerged-index check, and conflict
  marker check: PASS.

## Unverified Risks

- `go test -tags=unit ./internal/server -run "^TestAPIContracts$"` is blocked
  by pre-existing exact-payload drift unrelated to duplication.
- No authenticated browser interaction, production API, provider call,
  deployment, container update, or push ran.

## Contract Compliance

- Success criteria and allowed paths: satisfied. No denied path, deployment,
  dependency, or production configuration change was introduced.
- The internal `duplicate_operation_id` remains absent from public DTOs and
  uses migration `199`, avoiding upstream migration-number conflict.

## Recommendation

- `PASS / source-level + PostgreSQL smoke`. Commit and merge into local `main`
  are appropriate; retain the listed runtime and stale-suite gaps for a
  separately scoped task.
