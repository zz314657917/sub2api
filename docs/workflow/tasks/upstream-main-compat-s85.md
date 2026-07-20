# Task Contract: upstream-main-compat-s85

## Task ID

`upstream-main-compat-s85`

## Status

`approved`

## Role

Direct Codex behavior-level port of upstream `a2acbf553b9e5e51f`.
Preserve the local failover state topology; do not merge upstream history.

## Goal

Avoid forcing cache billing during same-account retries when a sticky/bound
session is present. Force cache billing must still occur for an actual account
switch or when the upstream error explicitly requests it.

## Success Criteria

- A retryable same-account error below the local `maxSameAccountRetries`
  threshold does not set `ForceCacheBilling` solely because `hasBoundSession`
  is true.
- Once same-account retries are exhausted and the handler switches accounts,
  `ForceCacheBilling` becomes true for a bound session.
- An explicit `failoverErr.ForceCacheBilling` remains effective during a
  same-account retry.
- Non-retryable bound-session and existing no-bound-session semantics remain
  unchanged.
- Focused and broader handler tests, gofmt, exact path/conflict/diff gates, and
  protected primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s85`
- Branch: `codex/upstream-main-compat-s85`
- Baseline: `631a8f5a89b56972fe6cab4066181ff611d79c0d`
- Upstream snapshot: `b8e844f4ee130ac069a7c5713c2413233186b83f`
- Upstream source: `a2acbf553b9e5e51f`
- Local topology: `HandleFailoverError` takes no retry-limit argument and uses
  the local `maxSameAccountRetries` constant; upstream tests must be adapted.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `backend/internal/handler/failover_loop.go`
- `backend/internal/handler/failover_loop_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s85.md`
- `docs/workflow/worker-results/upstream-main-compat-s85-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s85-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Other handlers/services, account scheduling, billing calculation, repositories,
  persistence, migrations, Ent, generated code, frontend, dependencies,
  deployment, Compose, VERSION, and lockfiles.
- Changes to retry counts/delays, account selection, temporary unscheduling,
  failover error classification, `UpstreamFailoverError`, or cache billing
  calculation outside the `ForceCacheBilling` decision.
- `knowledge/**`, global memories, handoff/timeline files, and all other
  upstream candidates.

## Constraints

- Work only in the isolated S85 worktree and preserve the dirty primary checkout.
- Compute the local `sameAccountRetry` state before cache-billing selection;
  preserve explicit `ForceCacheBilling` regardless of retry state.
- Keep retry timing/count and account-switch behavior unchanged.
- Adapt tests to the local five-argument `HandleFailoverError` signature and
  existing helpers; do not import unrelated upstream assumptions.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S85 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s85'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build-s85'
New-Item -ItemType Directory -Force -Path $env:GOCACHE,$env:GOTMPDIR | Out-Null
go test ./internal/handler -run '^TestHandleFailoverError_CacheBilling$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S85 focused handler test failed' }
go test ./internal/handler -run '^TestHandleFailoverError_' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S85 broader handler tests failed' }
Pop-Location

gofmt -w backend/internal/handler/failover_loop.go backend/internal/handler/failover_loop_test.go
if ($LASTEXITCODE -ne 0) { throw 'S85 gofmt failed' }
git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S85 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S85 has unmerged index entries' }
```

Evaluator additionally reviews the boolean state transition line-by-line,
confirms retry count/delay and switch paths have no diff, audits all paths,
scans real conflict markers, and rechecks all three protected hashes.

### Pre-commit Tracking Gate

```powershell
git add -u -- backend/internal/handler/failover_loop.go backend/internal/handler/failover_loop_test.go docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -f -- docs/workflow/tasks/upstream-main-compat-s85.md docs/workflow/worker-results/upstream-main-compat-s85-result.md docs/workflow/qa-reports/upstream-main-compat-s85-qa.md
$expected = @(
  'backend/internal/handler/failover_loop.go',
  'backend/internal/handler/failover_loop_test.go',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s85.md',
  'docs/workflow/worker-results/upstream-main-compat-s85-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s85-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) { throw "S85 staged path set differs from allowlist: $($pathDelta | Out-String)" }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S85 has unmerged index entries' }
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) { throw "S85 contains conflict markers: $($conflictMarkers -join ', ')" }
if ($LASTEXITCODE -ne 1) { throw 'S85 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S85 cached diff check failed' }
```

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s85-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s85-qa.md`
- Workflow status/log entries for contract review, implementation, QA, verdict.

## Stop Rules

- Stop if retry counts/delays, account switching, unscheduling, error
  classification, billing calculation, or paths outside the allowlist change.
- Stop if same-account retries still force cache billing solely from the bound
  session, explicit force billing is lost, tests are undiscovered, conflict
  markers appear, or a protected hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a two-file handler compatibility batch`
- qa_mode: `fresh focused/broader Go tests plus evidence-first diff review`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
