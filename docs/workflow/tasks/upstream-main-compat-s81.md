# Task Contract: upstream-main-compat-s81

## Task ID

`upstream-main-compat-s81`

## Status

`approved`

## Role

Direct Codex behavior-level port of one upstream subscription fix. The local
service topology is retained; no upstream history is merged.

## Goal

Port upstream `1db10dc5599d96f6613efa48ce161b0acb1f41bd` from snapshot
`d4b9797ff72024960a035cf22fdd8f213e149169` onto local baseline
`f62f8bbce2e354a59de29970c171e233a3c3f06e`.

When an administrator assigns a subscription to a user/group that already has
an expired reusable row, renew that row instead of returning a successful but
still-expired result. Preserve the separate `AssignOrExtendSubscription`
purchase/redeem semantics.

## Success Criteria

- `AssignSubscription` and `BulkAssignSubscription` reuse the existing row ID
  when its status is `expired`, or when its term is past expiry and its status
  is not `suspended`.
- Renewal starts at the current time, uses normalized requested validity days,
  respects `MaxExpiresAt`, sets status active, resets daily/weekly/monthly
  windows to `startOfDay(new StartsAt)`, and zeros all three usage counters.
- Existing `AssignedAt` and `AssignedBy` provenance remain unchanged.
- Admin renewal does not append duplicate notes when trimmed existing/input
  notes match; different notes are appended once through the existing helper.
- A suspended subscription is never reactivated or term-reset by admin
  assignment, whether its expiry is future or past.
- A non-expired active semantic match remains unchanged; a non-expired active
  semantic mismatch continues returning `SUBSCRIPTION_ASSIGN_CONFLICT`.
- Bulk renewal reports the existing row as `reused`, not `created`.
- `AssignOrExtendSubscription` remains unchanged and continues appending equal
  notes for a new purchase/redeem event.
- Focused default-tag tests, exact discovery, path/conflict/diff gates, and
  protected primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s81`
- Branch: `codex/upstream-main-compat-s81`
- Baseline: `f62f8bbce2e354a59de29970c171e233a3c3f06e`
- Upstream snapshot: `d4b9797ff72024960a035cf22fdd8f213e149169`
- Existing local behavior: `AssignOrExtendSubscription` already renews expired
  purchases; admin `AssignSubscription` currently returns an expired row as a
  successful reuse without renewing it.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `backend/internal/service/subscription_service.go`
- `backend/internal/service/subscription_assign_idempotency_test.go`
- `backend/internal/service/user_subscription_daily_quota_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s81.md`
- `docs/workflow/worker-results/upstream-main-compat-s81-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s81-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Subscription handlers, DTOs, repositories, Ent, migrations, generated code,
  billing/payment/welfare/redeem logic, caches outside the existing service
  invalidation calls, frontend, and API response shapes.
- `AssignOrExtendSubscription` implementation, `renewedSubscriptionTerm`, and
  `updateExistingSubscriptionTerm`; S81 reuses them without changing semantics.
- Deploy/Docker, VERSION, dependency manifests, lockfiles, README/docs outside
  workflow evidence, `knowledge/**`, global memories, handoff/timeline files.
- WS documentation (`8b75dd557`), special-password hardening, and all other
  upstream candidates.

## Constraints

- Work only in the isolated S81 worktree; preserve the dirty primary checkout.
- Add the renewal branch before semantic-conflict detection so expired rows can
  accept a new validity period and notes, while suspended rows still fall
  through to existing reuse/conflict semantics.
- Adapt cache invalidation to the local explicit L1/billing-cache pattern; do
  not import upstream's payment-deferred cache helper or refactor other callers.
- Use the existing transaction/update helpers for term reset. Do not create a
  second subscription row or change assignment provenance.
- Existing default-tag tests use wall-clock-relative active terms after S81;
  do not weaken assertions merely to avoid renewal behavior.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S81 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s81'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build-s81'
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
$expectedTests = @(
  'TestAssignSubscriptionReuseWhenSemanticsMatch',
  'TestAssignSubscriptionDoesNotReactivateFutureSuspendedSubscription',
  'TestAssignSubscriptionDoesNotReactivatePastExpirySuspendedSubscription',
  'TestAssignSubscriptionRenewsExpiredSemanticMatch',
  'TestAssignSubscriptionRenewsActiveSubscriptionPastExpiry',
  'TestAssignSubscriptionRenewsExpiredAndAppendsDifferentNotes',
  'TestAssignSubscriptionConflictWhenSemanticsMismatch',
  'TestBulkAssignSubscriptionCreatedReusedAndConflict',
  'TestBulkAssignSubscriptionRenewsExpiredSemanticMatch',
  'TestAssignOrExtendSubscription_ExpiredDailyCardStartsNewOneTimeQuota',
  'TestAssignOrExtendSubscription_ExpiredSubscriptionAppendsMatchingNotes'
)
$listed = @(go test ./internal/service -list '^Test(AssignSubscription|BulkAssignSubscription|AssignOrExtendSubscription_)')
if ($LASTEXITCODE -ne 0) { throw 'S81 test discovery failed' }
foreach ($name in $expectedTests) {
  if ($listed -notcontains $name) { throw "Missing S81 test: $name" }
}
$pattern = '^Test(AssignSubscription(ReuseWhenSemanticsMatch|DoesNotReactivateFutureSuspendedSubscription|DoesNotReactivatePastExpirySuspendedSubscription|RenewsExpiredSemanticMatch|RenewsActiveSubscriptionPastExpiry|RenewsExpiredAndAppendsDifferentNotes|ConflictWhenSemanticsMismatch)|BulkAssignSubscription(CreatedReusedAndConflict|RenewsExpiredSemanticMatch)|AssignOrExtendSubscription_(ExpiredDailyCardStartsNewOneTimeQuota|ExpiredSubscriptionAppendsMatchingNotes))$'
go test ./internal/service -run $pattern -count=1
if ($LASTEXITCODE -ne 0) { throw 'S81 focused service tests failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S81 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S81 has unmerged index entries'
}
```

Evaluator additionally reviews every business diff line, verifies the renewal
branch uses the existing transaction helper and local cache invalidation, checks
that the two helper implementations remain unchanged, audits all paths, scans
real conflict markers, and rechecks all three protected hashes.

### Pre-commit Tracking Gate

```powershell
git add -u -- backend/internal/service/subscription_service.go backend/internal/service/subscription_assign_idempotency_test.go backend/internal/service/user_subscription_daily_quota_test.go docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -f -- docs/workflow/tasks/upstream-main-compat-s81.md docs/workflow/worker-results/upstream-main-compat-s81-result.md docs/workflow/qa-reports/upstream-main-compat-s81-qa.md
git ls-files --error-unmatch docs/workflow/tasks/upstream-main-compat-s81.md docs/workflow/worker-results/upstream-main-compat-s81-result.md docs/workflow/qa-reports/upstream-main-compat-s81-qa.md
if ($LASTEXITCODE -ne 0) { throw 'S81 workflow evidence is not tracked' }
$expected = @(
  'backend/internal/service/subscription_service.go',
  'backend/internal/service/subscription_assign_idempotency_test.go',
  'backend/internal/service/user_subscription_daily_quota_test.go',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s81.md',
  'docs/workflow/worker-results/upstream-main-compat-s81-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s81-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) {
  throw "S81 staged path set differs from allowlist: $($pathDelta | Out-String)"
}
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S81 has unmerged index entries'
}
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) {
  throw "S81 contains conflict markers: $($conflictMarkers -join ', ')"
}
if ($LASTEXITCODE -ne 1) { throw 'S81 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S81 cached diff check failed' }
```

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s81-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s81-qa.md`
- Workflow status/log entries for contract review, implementation, QA, and verdict.

## Stop Rules

- Stop if a suspended subscription is reactivated, reset, or mutated.
- Stop if non-expired semantic conflict behavior or
  `AssignOrExtendSubscription` implementation/notes behavior changes.
- Stop if renewal creates a second row, changes assignment provenance, or
  requires a repository/schema/migration/handler/UI change.
- Stop if any named test is undiscovered, any path leaves the nine-item
  allowlist, an unmerged/conflict marker appears, or a protected hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a three-file service batch`
- qa_mode: `fresh default-tag service tests plus evidence-first diff review`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
