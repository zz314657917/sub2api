---
type: task-contract
status: approved
review_verdict: PASS
task_id: pelican-layout-choice-20261103
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 582ad5ab092441068c61c4f96eec148d8d022106
spec_ref: docs/workflow/spec.md
last_verified: 2026-11-03
---

# Task Contract

## Task ID
pelican-layout-choice-20261103

## Role
Controller owns planning, snapshots, integration and final evaluation.
Independently dispatch Contract Reviewer, Developer and QA according to
docs/workflow/agent-matrix.md. Approved metadata becomes effective only when an
independent PASS binds these exact final contract bytes.

## Goal
Let the user choose the current horizontal group rails or the previous vertical
group/card presentation on the Pelican tests page, and reduce avoidable preview
scroll work. User requested both presentations because the current page feels slow.

## Context
- Main: F:/mcplugins/sub2api.
- Candidate: E:/codex-worktrees/sub2api/pelican-layout-choice-20261103.
- Evidence: E:/codex-runtime/sub2api/pelican-layout-choice-20261103.
- The view and its test already contain unrelated uncommitted multi-round work.
  Begin with frozen current bytes, not HEAD versions. Preserve all unrelated
  tracked/untracked main files, staging and concurrent group-fast changes.
- Old vertical reference: 6d3f34e9d view, grouped responsive columns with each
  group's models stacked. Reuse current entries, controls and safe preview;
  do not restore the removed history dialog or old one-result-per-model behavior.

## Success Criteria
1. Add compact accessible controls labelled 上下排列 / 横向浏览, with an explicit
   selected state. First visit/invalid/missing storage defaults to vertical.
   Save a validated personal browser preference; refresh/remount restores it.
   Storage get/set failure must not break rendering or switching.
2. Vertical mode uses responsive group columns and vertically stacked result
   cards, retaining all supplied rounds and stable entry keys; horizontal mode
   retains existing group rails and navigation. One shared template/data path
   is preferred. Never mount two hidden copies of the gallery.
3. Switching layout does not call list/settings APIs or change filters,
   pagination, per-model selection, result identity, admin settings or shared
   group priority. The artwork cache remains page-scoped; destroy old preview
   consumers on layout change and correctly observe the new geometry.
4. Preserve search/status/model/group filters, running/failed/fallback states,
   exact-older-round enlarge, replay, automatic refresh, radar, admin settings
   and plans. No history button/dialog reintroduction.
5. Preserve actual-visibility-only iframes, 240px prefetch, maximum three detail
   requests, dedup/cache limits, document-hidden pause, authorization invalidation,
   sandbox/nonce/CSP and abort/cleanup behavior. Coalesce duplicate scroll/resize/
   observer geometry updates to at most one queued animation frame per component;
   cancel queued callbacks on unmount and keep critical visibility/auth clearing
   immediate. Do not add autoplay or a new playback preference.
6. Test defaults/restoration/invalid and blocked storage, selection persistence,
   filter/multi-round/enlarge/replay preservation, no switch-induced list requests
   or duplicate gallery, horizontal navigation and vertical narrow-screen layout.
   Add deterministic coverage for burst coalescing and queued-callback cleanup.
7. Actual browser fixture verification at desktop/mobile, light/dark: both modes,
   refresh persistence, multi-round filtering/enlarge/replay, no whole-page
   overflow, visibility-limited iframe counts, detail concurrency <=3 and console
   errors. Measure geometry updates on a deterministic scroll burst before/after;
   report the limited metric, never promise measured production FPS.

## Allowed Paths
Developer may edit only these files in the isolated candidate:
- frontend/src/views/user/PelicanTestsView.vue
- frontend/src/components/pelican/PelicanLazyArtwork.vue
- frontend/src/components/pelican/__tests__/PelicanTestsView.spec.ts
- frontend/src/components/pelican/__tests__/PelicanLazyArtwork.spec.ts

Controller owns this contract, its independent review, a bounded task section in
docs/workflow/spec.md and docs/workflow/status.md, and the final QA report.
External scripts, baselines, fixtures, screenshots and reports stay in evidence.
QA may create external acceptance scripts/reports, not change product or tests.

## Denied Paths
Everything else, especially backend, API contracts, shared server settings,
dependencies/lockfiles, sandbox/sanitizer, other UI, global rules/memories,
production config and any existing evidence. No commit/push/merge/reset/clean,
container/restart/deployment, real provider/DB requests or new dependencies.

## Constraints
Freeze original HEAD/index and file SHA256; copy current frontend inputs to an
isolated worktree before development. Developer reports changed paths and stops
writes before source freeze. Independent QA tests the frozen bytes and reports
evidence. Integrate only those exact bytes after checking target baseline drift;
preserve non-target raw bytes. No blanket Git or source updates.
Run Windows .cmd tool entries, read-only lint, and explicit external Vite outDir.
Use a task-owned browser session/profile and loopback fixture, record PIDs, close
it and verify owned browser/daemon processes are zero. Leave other sessions alone.

## Acceptance Commands
In candidate/frontend (then fresh main/frontend):
- node_modules/.bin/vitest.cmd run src/components/pelican/__tests__
- node_modules/.bin/vue-tsc.cmd --noEmit
- node_modules/.bin/eslint.cmd src/views/user/PelicanTestsView.vue src/components/pelican/PelicanLazyArtwork.vue src/components/pelican/__tests__/PelicanTestsView.spec.ts src/components/pelican/__tests__/PelicanLazyArtwork.spec.ts
- Developer: node_modules/.bin/vite.cmd build --outDir E:/codex-runtime/sub2api/pelican-layout-choice-20261103/developer-dist --emptyOutDir
- Independent QA: same Vite command with --outDir E:/codex-runtime/sub2api/pelican-layout-choice-20261103/qa-dist
- Fresh main: same Vite command with --outDir E:/codex-runtime/sub2api/pelican-layout-choice-20261103/main-dist
- From the applicable candidate/main repository root: git diff --check -- frontend/src/views/user/PelicanTestsView.vue frontend/src/components/pelican/PelicanLazyArtwork.vue frontend/src/components/pelican/__tests__/PelicanTestsView.spec.ts frontend/src/components/pelican/__tests__/PelicanLazyArtwork.spec.ts
Each stage owns its output directory. A retest uses a new versioned output
directory (for example qa-dist-r2), preserving previous builds and all logs.
Independent QA records its own commands, counts, exit codes and browser probes;
record any baseline failures with unchanged assertions, never waive new failures.

## Output
Developer: external developer-report.md starting with PASS/FAIL/BLOCKED, changed
paths, checks, caveats. QA: external qa-report.md starting
"### PASS/FAIL/BLOCKED: pelican-layout-choice-20261103", frozen hashes, Findings,
Executed Checks, Unverified Risks, Recommendation and browser cleanup evidence.
Controller: docs/workflow/qa-reports/pelican-layout-choice-20261103.md and concise
Chinese delivery noting local-only and fixture-versus-live limits.

## Stop Rules
Stop for target drift, unauthorized scope, missing independent contract PASS,
model unavailability or any introduced regression. Ask Controller to amend the
contract for changed scope; do not silently weaken tests. Do not claim production
speed improvement, deployment or live-backend acceptance from fixture results.
