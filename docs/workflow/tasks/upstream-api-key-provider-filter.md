---
status: approved
review_verdict: PASS
task_id: upstream-api-key-provider-filter
worker_model: gpt-5.6-terra
base_commit: 0ef64d868b35351c2a878cf022d2b93c447ea296
spec_ref: docs/workflow/tasks/upstream-api-key-provider-filter.md
---

# API key provider filter

## Task ID
upstream-api-key-provider-filter

## Role
Planner/controller approves contract; independent reviewer reviews before implementation; separate gpt-5.6-terra Developer and QA execute implementation and acceptance.

## Goal
Adapt upstream 55a95d4c674f0624fad1c31ee4afd0ec03e2a53f API key provider group filtering into the existing local KeysView and merge into local main. No push or deployment in this task.

## Success Criteria
- Create-key dialog offers Anthropic/OpenAI/domestic/other provider categories classified by configured platform, not group name; disabled empty categories and first available default. Unknown platforms fall back to other; use local GroupPlatform union, not unsupported upstream enum members.
- Creation default-group selector filters by selected category, clears stale selected group on category changes, handles late group loading and dialog reopen, and keeps valid choices. Empty groups handled without error.
- Edit dialog retains all authorized groups with original selection, unaffected by create provider state. Provider selection is UI-only and never enters API payload.
- Preserve local multi-group routing: provider filter applies only to creation default-group selector; route rows may intentionally span platforms, retain selections/order/timeouts and access all authorized groups. Preserve existing first-response timeout, failover and default selection behavior.
- Both Chinese and English labels/hints and accessible radio/fieldset controls. Existing key table/filter/group change dialogs unaffected.
- Independent executable acceptance and controller diff review before commit/merge. Preserve all main dirty files and existing records.

## Allowed Paths
- docs/workflow/agent-matrix.md (controller only; restore missing declared matrix)
- frontend/src/i18n/locales/en/keys.ts
- frontend/src/i18n/locales/zh/keys.ts
- frontend/src/utils/keyGroupProviders.ts
- frontend/src/utils/__tests__/keyGroupProviders.spec.ts
- frontend/src/views/user/KeysView.vue
- frontend/src/views/user/__tests__/KeysView.spec.ts
- docs/workflow/tasks/upstream-api-key-provider-filter.md
- docs/workflow/contract-reviews/upstream-api-key-provider-filter-review.md
- docs/workflow/worker-results/upstream-api-key-provider-filter-result.md
- docs/workflow/qa-reports/upstream-api-key-provider-filter-qa.md
- docs/workflow/status.md (controller only; isolated checkout, restore before integration)
- docs/workflow/main-log.md (controller only; isolated checkout, restore before integration)

## Denied Paths
All other tracked paths; backend, package/lock files, shared main workspace, store, knowledge, global memories, production configuration/data.

## Constraints
Use existing isolated E:/codex-worktrees/sub2api/upstream-api-key-provider-filter checkout. Behavior port, no wholesale cherry-pick. Developer and QA must use gpt-5.6-terra; do not substitute silently. No user/paid provider calls. No new API or dependencies. Runtime artifacts outside repository in E:/codex-runtime/pge/sub2api/upstream-api-key-provider-filter. Do not change main dirty contents, stash or force merge.

## Acceptance Commands
From frontend (install frozen dependencies if missing; preserve manifests):
- node_modules/.bin/vitest.cmd run src/views/user/__tests__/KeysView.spec.ts src/utils/__tests__/keyGroupProviders.spec.ts
- node_modules/.bin/vue-tsc.cmd --noEmit
- node_modules/.bin/eslint.cmd src/views/user/KeysView.vue src/views/user/__tests__/KeysView.spec.ts src/utils/keyGroupProviders.ts src/utils/__tests__/keyGroupProviders.spec.ts src/i18n/locales/en/keys.ts src/i18n/locales/zh/keys.ts
- node_modules/.bin/vite.cmd build --outDir E:/codex-runtime/pge/sub2api/upstream-api-key-provider-filter/dist
- git diff --check; before commit git diff --cached --check.
Tests must exercise real component create/edit submission, provider switching, disabled/empty categories, delayed groups, reopen, unknown platform, misleading names, and preservation of cross-provider multi-group rows and timeouts. No-tests-to-run is not PASS. If baseline compilation fails, reproduce against immutable base; affected production component must still have executable acceptance or block integration.
Independent QA additionally inspect full-style mocked page desktop/mobile rendering using task-isolated browser profile if browser is used; verify labels/layout and cleanup owned browser/Vite processes. Never use user browser profile or real credentials.

## Output
Developer report lists changed files, executed commands, limitations, contract compliance. Independent QA report first line ### PASS/FAIL/BLOCKED: upstream-api-key-provider-filter, with findings, executed evidence and unverified scope. Commit feature and process evidence separately after controller review; safe fast-forward merge only after confirming main HEAD, index and dirty intersection.

## Stop Rules
Block on Terra unavailable, needed allowlist expansion, unrelated changes, failed effective acceptance, ambiguous behavior affecting local routes, main overlap or divergence requiring reassessment. Report concrete reason; do not overwrite or silently lower standards.
