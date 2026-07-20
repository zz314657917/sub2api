# Task Contract: upstream-main-compat-s83

## Task ID

`upstream-main-compat-s83`

## Status

`approved`

## Role

Direct Codex behavior-level port of upstream `33796030222a16ec40a5291e0c29deafe5e2babd`.
The local subscription views and formatting helper remain authoritative; no
upstream history is merged.

## Goal

Show subscription expiry timestamps to the minute in the administrator and
user subscription views, while preserving existing locale handling, invalid
date behavior, expiration labels, and remaining-days calculations.

## Success Criteria

- Add `formatDateTimeToMinute` beside the existing local date helpers.
- Valid dates render year, month, day, hour, and minute without seconds using
  the existing locale-aware formatter.
- Invalid/null dates continue returning an empty string.
- Admin subscription expiry cells/details and the local `UserSubscriptionsPanel`
  expiry display use the minute formatter; status and days-remaining logic is
  unchanged.
- Focused Vitest, frontend typecheck/build, exact path/conflict/diff gates, and
  protected primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s83`
- Branch: `codex/upstream-main-compat-s83`
- Baseline: `631a8f5a89b56972fe6cab4066181ff611d79c0d`
- Upstream snapshot: `b8e844f4ee130ac069a7c5713c2413233186b83f`
- Upstream source: `33796030222a16ec40a5291e0c29deafe5e2babd`
- Local gap: the admin subscription view and `UserSubscriptionsPanel` still
  call `formatDateOnly`; the shared formatter has no minute-specific helper.
- The primary checkout has unrelated uncommitted UsageView work and a separate
  `usage-model-reasoning-effort-s82` workflow draft. S83 must not absorb or
  overwrite that work; workflow-file merge conflicts are deferred until the
  primary S82 work is reconciled.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `frontend/src/utils/format.ts`
- `frontend/src/utils/__tests__/formatDateTimeToMinute.spec.ts`
- `frontend/src/views/admin/SubscriptionsView.vue`
- `frontend/src/components/user/UserSubscriptionsPanel.vue`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s83.md`
- `docs/workflow/worker-results/upstream-main-compat-s83-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s83-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Backend code, APIs, DTOs, persistence, migrations, billing, payment,
  scheduler, deployment, Compose, VERSION, dependencies, lockfiles, and
  generated files.
- Usage tables/views, model-display helpers, public pages, account forms, or
  unrelated frontend components.
- `knowledge/**`, global memories, handoff/timeline files, and all other
  upstream candidates.
- Any change to expiry status, `getDaysRemaining`, subscription API payloads,
  timezone policy, or date-only displays outside the two subscription views.

## Constraints

- Work only in the isolated S83 worktree and preserve the dirty primary checkout.
- Reuse the existing `formatDate` implementation; do not duplicate Intl logic.
- Preserve the existing local locale selection and empty-string invalid-date
  behavior.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S83 automatically.

## Acceptance Commands

```powershell
Push-Location frontend
if (-not (Test-Path node_modules)) {
  New-Item -ItemType Junction -Path node_modules -Target F:/mcplugins/sub2api/frontend/node_modules | Out-Null
}
npm.cmd run test:run -- src/utils/__tests__/formatDateTimeToMinute.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S83 focused formatter test failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S83 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S83 frontend build failed' }
Pop-Location

if ((Select-String -Path frontend/src/views/admin/SubscriptionsView.vue -Pattern 'formatDateOnly' -Quiet)) {
  throw 'S83 admin subscription expiry still uses formatDateOnly'
}
if ((Select-String -Path frontend/src/components/user/UserSubscriptionsPanel.vue -Pattern 'formatDateOnly' -Quiet)) {
  throw 'S83 user subscription panel expiry still uses formatDateOnly'
}
if (-not (Select-String -Path frontend/src/utils/format.ts -Pattern 'formatDateTimeToMinute' -Quiet)) {
  throw 'S83 minute formatter is missing'
}

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S83 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S83 has unmerged index entries' }
```

Evaluator additionally reviews every business diff line, confirms status and
remaining-days code is unchanged, audits all paths, scans real conflict
markers, and rechecks all three protected hashes.

### Pre-commit Tracking Gate

```powershell
git add -u -- frontend/src/utils/format.ts frontend/src/views/admin/SubscriptionsView.vue frontend/src/components/user/UserSubscriptionsPanel.vue docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -- frontend/src/utils/__tests__/formatDateTimeToMinute.spec.ts
git add -f -- docs/workflow/tasks/upstream-main-compat-s83.md docs/workflow/worker-results/upstream-main-compat-s83-result.md docs/workflow/qa-reports/upstream-main-compat-s83-qa.md
$expected = @(
  'frontend/src/utils/format.ts',
  'frontend/src/utils/__tests__/formatDateTimeToMinute.spec.ts',
  'frontend/src/views/admin/SubscriptionsView.vue',
  'frontend/src/components/user/UserSubscriptionsPanel.vue',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s83.md',
  'docs/workflow/worker-results/upstream-main-compat-s83-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s83-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) { throw "S83 staged path set differs from allowlist: $($pathDelta | Out-String)" }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S83 has unmerged index entries' }
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) { throw "S83 contains conflict markers: $($conflictMarkers -join ', ')" }
if ($LASTEXITCODE -ne 1) { throw 'S83 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S83 cached diff check failed' }
```

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s83-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s83-qa.md`
- Workflow status/log entries for contract review, implementation, QA, verdict.

## Stop Rules

- Stop if implementation requires backend/API/persistence changes or touches
  UsageView/model-display files.
- Stop if invalid dates, status labels, remaining-day calculations, timezone
  selection, or subscription payloads change.
- Stop if any named test is undiscovered, any path leaves the exact ten-item
  allowlist, an unmerged/conflict marker appears, or a protected hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a four-file frontend batch`
- qa_mode: `fresh focused Vitest plus frontend compile and evidence-first review`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
