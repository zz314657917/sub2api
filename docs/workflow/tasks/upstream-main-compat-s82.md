# Task Contract: upstream-main-compat-s82

## Task ID

`upstream-main-compat-s82`

## Status

`approved`

## Role

Direct Codex behavior-level port of upstream documentation commit
`8b75dd5576459a887b87580d8fed76e0354e7d15`. The local WS mode vocabulary
is authoritative; no upstream history or runtime implementation is merged.

## Goal

Clarify that account-level OpenAI Responses WS modes only take effect when
`gateway.openai_ws.mode_router_v2_enabled=true`, while retaining the local
`off / ctx_pool / passthrough` mode set and the distinct local oversized-frame
HTTP bridge behavior.

## Success Criteria

- README documents the global YAML switch and equivalent environment variable.
- README and the example config state that disabling the v2 router ignores
  account-level WS mode selection and retains legacy routing.
- The example config continues listing only locally accepted ingress modes:
  `off`, `ctx_pool`, and `passthrough`.
- English and Chinese account help text both name
  `gateway.openai_ws.mode_router_v2_enabled=true` and the supported account
  modes `ctx_pool` / `passthrough`.
- Locale help text does not claim `http_bridge` is an account-level mode.
- A focused Vitest locks the bilingual prerequisite and local vocabulary.
- Focused Vitest, frontend typecheck/build, exact path/conflict/diff gates, and
  protected primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s82`
- Branch: `codex/upstream-main-compat-s82`
- Baseline: `631a8f5a89b56972fe6cab4066181ff611d79c0d`
- Upstream snapshot: `d4b9797ff72024960a035cf22fdd8f213e149169`
- Upstream source: `8b75dd5576459a887b87580d8fed76e0354e7d15`
- Local divergence: upstream now offers account-level `http_bridge`, but local
  config validation, service constants, frontend utilities, and selector only
  support `off / ctx_pool / passthrough`. Local `http_bridge_enabled` controls
  oversized first-message fallback and is not an account mode.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `README.md`
- `deploy/config.example.yaml`
- `frontend/src/i18n/__tests__/wsModeLocaleDesc.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s82.md`
- `docs/workflow/worker-results/upstream-main-compat-s82-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s82-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Backend config/defaults/validation, WS handlers/services, account metadata,
  repositories, DTOs, Ent, migrations, generated code, billing, scheduler,
  payment, security, and API response shapes.
- Frontend components, WS mode utilities/enums, account form behavior, routing,
  dependencies, manifests, and lockfiles.
- Compose/env/deployment runtime files other than the allowlisted config sample,
  VERSION, containers, `knowledge/**`, global memories, handoff/timeline files.
- Adding account-level `http_bridge`, changing the legacy/v2 router behavior,
  or importing any other upstream candidate.

## Constraints

- Work only in the isolated S82 worktree and preserve the dirty primary checkout.
- Hand-port the prerequisite clarification; do not cherry-pick the upstream
  wording that advertises unsupported account-level `http_bridge`.
- Keep configuration values unchanged. This Sprint changes explanatory comments
  only in `deploy/config.example.yaml`.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S82 automatically.

## Acceptance Commands

```powershell
Push-Location frontend
npm.cmd run test:run -- src/i18n/__tests__/wsModeLocaleDesc.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S82 focused locale test failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S82 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S82 frontend build failed' }
Pop-Location

$readme = Get-Content README.md -Raw
$config = Get-Content deploy/config.example.yaml -Raw
foreach ($text in @($readme, $config)) {
  if ($text -notmatch 'mode_router_v2_enabled') {
    throw 'S82 prerequisite missing from README or example config'
  }
}
if ($config -notmatch 'off\|ctx_pool\|passthrough') {
  throw 'S82 local ingress vocabulary changed or disappeared'
}

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S82 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S82 has unmerged index entries'
}
```

Evaluator additionally reviews every business diff line, confirms the local
runtime accepts only `off / ctx_pool / passthrough`, audits all paths, scans
real conflict markers, and rechecks all three protected hashes.

### Pre-commit Tracking Gate

```powershell
git add -u -- README.md deploy/config.example.yaml frontend/src/i18n/locales/en/admin/accounts.ts frontend/src/i18n/locales/zh/admin/accounts.ts docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -- frontend/src/i18n/__tests__/wsModeLocaleDesc.spec.ts
git add -f -- docs/workflow/tasks/upstream-main-compat-s82.md docs/workflow/worker-results/upstream-main-compat-s82-result.md docs/workflow/qa-reports/upstream-main-compat-s82-qa.md
git ls-files --error-unmatch docs/workflow/tasks/upstream-main-compat-s82.md docs/workflow/worker-results/upstream-main-compat-s82-result.md docs/workflow/qa-reports/upstream-main-compat-s82-qa.md
if ($LASTEXITCODE -ne 0) { throw 'S82 workflow evidence is not tracked' }
$expected = @(
  'README.md',
  'deploy/config.example.yaml',
  'frontend/src/i18n/__tests__/wsModeLocaleDesc.spec.ts',
  'frontend/src/i18n/locales/en/admin/accounts.ts',
  'frontend/src/i18n/locales/zh/admin/accounts.ts',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s82.md',
  'docs/workflow/worker-results/upstream-main-compat-s82-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s82-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) {
  throw "S82 staged path set differs from allowlist: $($pathDelta | Out-String)"
}
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S82 has unmerged index entries'
}
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) {
  throw "S82 contains conflict markers: $($conflictMarkers -join ', ')"
}
if ($LASTEXITCODE -ne 1) { throw 'S82 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S82 cached diff check failed' }
```

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s82-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s82-qa.md`
- Workflow status/log entries for contract review, implementation, QA, verdict.

## Stop Rules

- Stop if implementation requires runtime, enum, form, API, migration, or
  dependency changes.
- Stop if any text advertises account-level `http_bridge`, omits the global
  router prerequisite, or changes an example configuration value.
- Stop if the focused locale spec is undiscovered, any path leaves the exact
  eleven-item allowlist, an unmerged/conflict marker appears, or a protected
  hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a five-file docs/i18n batch`
- qa_mode: `fresh locale Vitest plus frontend compile and evidence-first review`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
