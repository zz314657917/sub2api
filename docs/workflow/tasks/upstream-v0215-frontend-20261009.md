---
status: approved
review_verdict: PASS
task_id: upstream-v0215-frontend-20261009
worker_model: current-session-default
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
spec_ref: docs/workflow/tasks/upstream-v0215-20261009-spec.md
---
# Task Contract
## Task ID
upstream-v0215-frontend-20261009
## Role
Developer; implement only after independent review PASS, disjoint frontend ownership.
## Goal
Adapt independent v0.2.15 frontend numeric/lifecycle fixes, preserving local customized components.
## Success Criteria
- Fix named RPM edits (integers including zero, reject fractional/invalid, blank doesn't overwrite), stale group multiplier requests, initial-only channel model default autofill, scientific notation prices and sub-byte units.
- Replace hard-coded long-context x2 with neutral long-context text; preserve already-correct nullish zero multiplier handling.
- Escape affects only the top dialog; disabling selectors closes them; backup polling cannot restart after unmount; stale cleanup/scheduled-test/template responses cannot overwrite current state.
- Ops settings cannot save before successful load; alert durations enforce whole-minute values; announcement read failures aren't reported as success; concurrent step-up prompts share a pending operation.
- Add mounted/deferred-promise tests for changes. Adapt existing API signatures and visible Chinese/English locale keys, no broad upstream locale replacement. Add only neutral long-context keys to the already-loaded admin/usage.ts modules.
- Run new and related tests, frontend typecheck/build where feasible; report exact baseline failures separately.
## Context
Upstream commits: 7b1aa9314, 8811c125e, 440784914, 5ace0699b, 3a1a37b7b, d5f490e10, e0019da46, 8167d85b3, 00a922c42, 8a495b0a5, 8f636513f, 4551fb0d8, 81bd85ec5, b07b1fb5d, cee76030b, c64145312. Payments, profile pages, TPS and platform catalog deferred.
## Allowed Paths
- `frontend/src/components/admin/group/GroupRPMOverridesModal.vue`
- `frontend/src/components/admin/group/GroupRateMultipliersModal.vue`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/UsageCleanupDialog.vue`
- `frontend/src/components/admin/account/ScheduledTestsPanel.vue`
- `frontend/src/components/admin/monitor/MonitorTemplateManagerDialog.vue`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/Select.vue`
- `frontend/src/components/common/ProxySelector.vue`
- `frontend/src/components/common/AnnouncementBell.vue`
- `frontend/src/views/admin/BackupView.vue`
- `frontend/src/views/admin/ops/components/OpsSettingsDialog.vue`
- `frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue`
- `frontend/src/utils/pricing.ts`
- `frontend/src/utils/format.ts`
- `frontend/src/composables/useStepUp.ts`
- `frontend/src/stores/announcements.ts`
- `frontend/src/i18n/locales/en/admin/usage.ts`
- `frontend/src/i18n/locales/zh/admin/usage.ts`
- `frontend/src/**/__tests__/*v0215*.spec.ts`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `docs/workflow/worker-results/upstream-v0215-frontend-20261009-result.md`
## Denied Paths
- All non-allowlist paths, knowledge, manifests/lockfiles, payments, platform catalog, backend, original shared tree and shared dependencies.
## Constraints
- Node dependencies are a junction to existing shared modules; don't install/delete/change dependencies or caches under shared node_modules. Use task runtime temp/output locations.
- Preserve business behavior and local custom UI, including zero multipliers and GroupID pricing. Minimal changes only, no commit/push/deploy.
- Narrow relevant-test exception: change only the existing long-context marker assertion (and its mock translation labels if required) from the obsolete x2 expectation to the new neutral long-context text; preserve all other tests and copied baseline content.
- Use `node node_modules/vitest/vitest.mjs run` and `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` when .bin scripts aren't available. Use explicit own test files.
## Acceptance Commands
From frontend: new/related `node node_modules/vitest/vitest.mjs run PATHS`; `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit`; production Vite build to E runtime output, no shared/dist overwrite; `git diff --check`.
## Output
Write first-line DONE/FAILED/BLOCKED report for task ID, changed files, command exits/test counts, limitations, risks, knowledge_candidates.
## Stop Rules
Review/approval absent, required out-of-scope path, model unavailable or repeated failures: stop and report. No browser required; don't claim browser verification.
