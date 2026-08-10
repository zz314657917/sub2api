# Task Contract: upstream-api-key-validation-s209

## Task ID

`upstream-api-key-validation-s209`

## Status

`approved`

## Role

Direct Codex implementation and evaluator review; no worker is delegated.

## Goal

Behaviorally adapt upstream `f5c108c83` so API Key quota, rate-limit, and
create-expiry inputs fail closed before persistence or idempotency execution,
while preserving the local divergent handler/service topology.

## Success Criteria

- Create and Update reject negative, `NaN`, and infinite `quota`,
  `rate_limit_5h`, `rate_limit_1d`, and `rate_limit_7d` values.
- Create rejects non-nil `expires_in_days` values less than or equal to zero;
  nil and positive values remain valid.
- Handler validation runs after JSON binding but before Create enters the
  idempotency boundary and before Update calls the service.
- Service validation repeats the boundary for non-HTTP/internal callers and
  returns structured HTTP `400` errors with stable limit/expiry reason codes.
- Zero and large finite numeric values remain valid. Update expiration parsing,
  clearing, and valid timestamp behavior remain unchanged.
- Focused default-tag tests, complete handler/service regressions, server
  compilation, formatting, allowlist, provenance, conflict, and index gates pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-api-key-validation-s209`
- Frozen base: `bb4b74e73df85962fa5fd75c3d73a9545b63c61d`
- Upstream source: `f5c108c836a849dd6b982e46238c42c6db611899`
- Upstream provenance: `f5c108c83` is an ancestor of the fetched
  `upstream/main`.
- Main-worktree S208 changes are uncommitted and excluded from this isolated
  branch. S209 must not modify or merge the primary checkout.

## Allowed Paths

- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/api_key_handler_validation_s209_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_validation_s209_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-api-key-validation-s209.md`
- `docs/workflow/worker-results/upstream-api-key-validation-s209-result.md`
- `docs/workflow/qa-reports/upstream-api-key-validation-s209-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Every path not listed above, including `backend/ent/**`,
  `backend/migrations/**`, repositories, generated code, dependencies,
  lockfiles, frontend, configuration, Docker, deployment, VERSION, global or
  repository knowledge files, primary-worktree S208 files, and `outputs/`.
- Cherry-picking or merging `f5c108c83`, remote push, provider calls, shared
  database/Redis access, container updates, deployment, and production traffic.

## Constraints

- Port behavior manually; do not replace the local API Key handler/service or
  use upstream unit-tag tests unchanged.
- Keep both handler and service validation. HTTP-only validation is not enough
  because internal callers invoke `APIKeyService` directly.
- Keep omitted pointer semantics and `0 = unlimited` behavior intact.
- Do not change idempotency keys, request hashing, API response envelopes,
  persistence, billing, group routing, Cafe managed-Key rules, or expiration
  update semantics.
- No push, deployment, container update, schema/migration, shared resource, or
  production operation.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler -run '^TestS209' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S209 handler validation regressions failed' }
go test ./internal/service -run '^TestS209' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S209 service validation regressions failed' }
go test ./internal/handler ./internal/service
if ($LASTEXITCODE -ne 0) { throw 'S209 package regressions failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S209 server compile check failed' }
Pop-Location

gofmt -d backend/internal/handler/api_key_handler.go backend/internal/handler/api_key_handler_validation_s209_test.go backend/internal/service/api_key_service.go backend/internal/service/api_key_service_validation_s209_test.go
git diff --check
$changed = @(git diff --name-only bb4b74e73df85962fa5fd75c3d73a9545b63c61d...HEAD)
$allowed = @(
  'backend/internal/handler/api_key_handler.go',
  'backend/internal/handler/api_key_handler_validation_s209_test.go',
  'backend/internal/service/api_key_service.go',
  'backend/internal/service/api_key_service_validation_s209_test.go',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/tasks/upstream-api-key-validation-s209.md',
  'docs/workflow/worker-results/upstream-api-key-validation-s209-result.md',
  'docs/workflow/qa-reports/upstream-api-key-validation-s209-qa.md',
  'docs/workflow/main-log.md'
)
$unexpected = @($changed | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -ne 0) { throw "S209 unexpected paths: $($unexpected -join ', ')" }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S209 has unmerged index entries' }
rg -n '^(<<<<<<<|=======|>>>>>>>)' backend/internal/handler/api_key_handler.go backend/internal/handler/api_key_handler_validation_s209_test.go backend/internal/service/api_key_service.go backend/internal/service/api_key_service_validation_s209_test.go docs/workflow
if ($LASTEXITCODE -eq 0) { throw 'S209 conflict marker detected' }
git merge-base --is-ancestor f5c108c836a849dd6b982e46238c42c6db611899 upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S209 upstream provenance failed' }
```

## Output

- Direct implementation result:
  `docs/workflow/worker-results/upstream-api-key-validation-s209-result.md`
- QA report: `docs/workflow/qa-reports/upstream-api-key-validation-s209-qa.md`
- Workflow status/log entries for contract approval, build, QA, and final verdict.
- Commit only allowed files to the isolated S209 branch. Do not integrate into
  local `main` while S208 remains uncommitted.

## Stop Rules

- Stop if the change requires repository, schema, migration, generated-code,
  frontend, configuration, dependency, or deployment edits.
- Stop if valid zero/unlimited inputs or existing expiration-clear semantics
  would be rejected.
- Stop if validation cannot occur before idempotency/repository side effects.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `focused default-tag and complete package verification plus diff review`
- worktree_root: `E:/codex-worktrees/sub2api/upstream-api-key-validation-s209`

## Contract Review

`PASS / contract-approved`: The four-file product scope mirrors the upstream
behavior without importing its unit-tag assumptions. Both HTTP and service
boundaries are covered, acceptance commands are executable from the frozen
worktree, and primary-worktree S208 changes are explicitly excluded.
