# Task Contract: upstream-main-compat-s86

## Task ID

`upstream-main-compat-s86`

## Status

`approved`

## Role

Direct Codex behavior-level port of upstream `b8c640d218471321d1bbaf1f1ba1bb5a5b50154a`.
Preserve the local proxy quality-check topology; do not merge upstream history.

## Goal

Include xAI/Grok reachability in administrator proxy quality checks and show the
target with the existing `Grok` product label in the result table.

## Success Criteria

- Proxy quality checks issue `GET https://api.x.ai/v1/models` through the
  checked proxy.
- An HTTP 401 response from the Grok target counts as reachable/pass, matching
  the existing unauthenticated OpenAI probe semantics.
- Existing OpenAI, Anthropic, Gemini, scoring, timeout, challenge, persistence,
  and error-classification behavior remains unchanged.
- The proxy-quality result table renders target `grok` as `Grok`.
- Focused backend tests, frontend typecheck/build, exact path/conflict/diff
  gates, and protected primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-main-compat-s86`
- Branch: `codex/upstream-main-compat-s86`
- Baseline: `24ade9b7166fc90114baa6d97f162f360b77dbc7`
- Upstream snapshot: `db4295d646bf6ea69f2b079e0f855647b06a0200`
- Upstream source: `b8c640d218471321d1bbaf1f1ba1bb5a5b50154a`
- The upstream commit also adds `whitespace-nowrap` table styling. That visual
  cleanup is unrelated to Grok reachability and is intentionally excluded.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_proxy_quality_test.go`
- `frontend/src/views/admin/ProxiesView.vue`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s86.md`
- `docs/workflow/worker-results/upstream-main-compat-s86-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s86-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Proxy CRUD, proxy scoring/grades, exit-IP probing, timeouts, challenge
  detection, persistence, batch operations, account routing, billing, scheduler,
  repositories, migrations, Ent, generated code, dependencies, deployment,
  Compose, VERSION, and lockfiles.
- UI layout/class changes, new columns, i18n keys, or frontend behavior outside
  the target display-name switch.
- `knowledge/**`, global memories, handoff/timeline files, and all other
  upstream candidates.

## Constraints

- Work only in the isolated S86 worktree and preserve the dirty primary checkout.
- Add exactly one Grok target with `GET`, the canonical xAI models endpoint, and
  401 as its only allowed status.
- Reuse `runProxyQualityTarget`; do not add special-case transport logic.
- Do not port upstream whitespace/style changes.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S86 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s86'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build-s86'
New-Item -ItemType Directory -Force -Path $env:GOCACHE,$env:GOTMPDIR | Out-Null
go test ./internal/service -run '^TestProxyQualityTargets_IncludesGrok$|^TestRunProxyQualityTarget_GrokUnauthorizedPasses$|^TestRunProxyQualityTarget_' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S86 backend proxy-quality tests failed' }
Pop-Location

Push-Location frontend
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S86 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S86 frontend build failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S86 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S86 has unmerged index entries' }
```

Evaluator additionally reviews the exact target URL/method/status, confirms
the target-name switch only gained `grok`, audits all paths, scans real conflict
markers, and rechecks all three protected hashes.

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s86-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s86-qa.md`
- Workflow status/log entries for contract review, implementation, QA, verdict.

## Stop Rules

- Stop if proxy scoring, transport timeouts, other targets, persistence, UI
  layout, or paths outside the allowlist change.
- Stop if Grok uses an authenticated probe, accepts statuses beyond 401, tests
  are undiscovered, conflict markers appear, or a protected hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a three-file compatibility batch`
- qa_mode: `focused Go tests plus frontend compile/build and evidence-first review`
- worktree_root: `E:/codex-worktrees`
