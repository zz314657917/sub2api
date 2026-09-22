---
status: approved
review_verdict: PASS
task_id: upstream-v027-ui
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---
## Task ID
upstream-v027-ui
## Role
Terra Developer; independent Terra QA, controller final review.
## Goal
Port missing low-risk UI behavior from 406be7c518 b51f0759d4 be313735dc 017e9e98d1 fe36f4a914 f8e5a0d94f 0838e0e610 98321a054e ee9ac3e45f 14636d2f0e 1a32b91eb2 130ba634ad 6a4938bdfb 0ed735d3ba 21532add46. Scope is the 15 UI items in parent plan; report per-item implemented/equivalent/not-applicable with source evidence.
## Success Criteria
Clear subscription loading, concurrent callers await same payment config; unique dialog title IDs; partial refund warnings compare requested refund not entire order; registration promo visibility initialized from injected config; empty model Tab navigates, populated Tab behavior preserved; reject negative quota; invalid payment amount restores accepted DOM text; TOTP DOM digits follow state and normalized API errors display; numeric pagination jump accepts numbers; clipboard fallback exceptions return failure; partial bulk announcement reads retained; proxy tests deduplicated across single/batch; status changes reset order page. Do not change backend/payment calculations/auth boundaries.
## Allowed Paths
frontend/src/components/admin/channel/__tests__/ModelTagInput.keyboard.spec.ts
frontend/src/components/admin/channel/ModelTagInput.vue
frontend/src/components/admin/payment/__tests__/AdminRefundDialog.balance.spec.ts
frontend/src/components/admin/payment/AdminRefundDialog.vue
frontend/src/components/common/__tests__/BaseDialog.ids.spec.ts
frontend/src/components/common/__tests__/Pagination.jump.spec.ts
frontend/src/components/common/__tests__/ProxySelector.testing.spec.ts
frontend/src/components/common/BaseDialog.vue
frontend/src/components/common/Pagination.vue
frontend/src/components/common/ProxySelector.vue
frontend/src/components/payment/__tests__/AmountInput.spec.ts
frontend/src/components/payment/AmountInput.vue
frontend/src/components/user/profile/__tests__/Totp.errors.spec.ts
frontend/src/components/user/profile/__tests__/TotpSetupModal.inputs.spec.ts
frontend/src/components/user/profile/TotpDisableDialog.vue
frontend/src/components/user/profile/TotpSetupModal.vue
frontend/src/composables/__tests__/useClipboard.spec.ts
frontend/src/composables/useClipboard.ts
frontend/src/stores/__tests__/announcements.markAll.spec.ts
frontend/src/stores/__tests__/payment.config.spec.ts
frontend/src/stores/__tests__/subscriptions.clear.spec.ts
frontend/src/stores/announcements.ts
frontend/src/stores/payment.ts
frontend/src/stores/subscriptions.ts
frontend/src/views/auth/__tests__/RegisterView.spec.ts
frontend/src/views/auth/RegisterView.vue
frontend/src/views/user/__tests__/UserOrdersView.filters.spec.ts
frontend/src/views/user/UserOrdersView.vue
docs/workflow/worker-results/upstream-v027-ui.md
docs/workflow/qa-reports/upstream-v027-ui.md
outputs/upstream-v027-ui/ (task-owned harness/logs/screenshots only)
frontend/v027-ui-smoke.html
frontend/v027-ui-smoke.ts
## Denied Paths
All others including dependencies, lockfiles, main workspace, backend and migrations. No commit/push/deployment. No shared service writes.
## Constraints
Adapt behavior, preserve local APIs and existing tests; no complete upstream file replacement. Read-only shared node_modules junction provided by controller; never install/change dependencies through it. Temporary browser harness allowed in specified files only; remove harness before final delivery. No real payments or real auth/data writes. Minimal production changes.
## Acceptance Commands
From frontend:
```powershell
pnpm.cmd exec vitest run src/components/admin/channel/__tests__/ModelTagInput.keyboard.spec.ts src/components/admin/payment/__tests__/AdminRefundDialog.balance.spec.ts src/components/common/__tests__/BaseDialog.ids.spec.ts src/components/common/__tests__/Pagination.jump.spec.ts src/components/common/__tests__/ProxySelector.testing.spec.ts src/components/payment/__tests__/AmountInput.spec.ts src/components/user/profile/__tests__/Totp.errors.spec.ts src/components/user/profile/__tests__/TotpSetupModal.inputs.spec.ts src/composables/__tests__/useClipboard.spec.ts src/stores/__tests__/announcements.markAll.spec.ts src/stores/__tests__/payment.config.spec.ts src/stores/__tests__/subscriptions.clear.spec.ts src/views/auth/__tests__/RegisterView.spec.ts src/views/user/__tests__/UserOrdersView.filters.spec.ts
pnpm.cmd exec vue-tsc -b
pnpm.cmd exec vite build --outDir ../outputs/upstream-v027-ui/dist
```
From root: git diff HEAD --check; git diff --name-only --diff-filter=U (empty). QA checks all changed paths including untracked tests. If baseline tests fail, identify immutable base version and reproduce using task-owned temporary harness/source, never drop assertions.
## Runtime/browser acceptance
Use real imported components in temporary v027-ui-smoke harness, mocked APIs only. Start pnpm.cmd exec vite --host 127.0.0.1 --port 4328 --strictPort from frontend; record PID. Visit http://127.0.0.1:4328/v027-ui-smoke.html with playwright-cli session v027-ui, explicit profile E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/profile. Verify two dialogs distinct title IDs; pagination jump; invalid amount visible correction; model Tab focus; TOTP retry digits synchronized. Other data/store flows covered by actual mounted component Vitest with mocks. Capture desktop and 390px screenshots for changed common components. Close session using playwright-cli -s=v027-ui close; verify task-owned profile/cliDaemon processes absent, stop only exact task Vite PID. If tooling fails, report BLOCKED browser acceptance; no forced global process cleanup. QA may use CLI run-code browser interactions but not add Playwright test framework.
## Output
First line ### PASS/FAIL/BLOCKED: upstream-v027-ui. Per-item mapping, files, commands/exit codes, runtime screenshots and cleanup proof. Independent QA reruns tests and browser smoke; no remote availability claims.
## Stop Rules
Missing APIs/dependencies or out-of-scope changes require controller review. Terra unavailable => BLOCKED. Preserve useful partial edits, report exact limitations. Never claim complete acceptance without runtime checks.

## Controller amendment 2026-09-22
0838e0e610 is NOT APPLICABLE: baseline has no UserPlatformQuotaModal.vue and no equivalent administrator platform-quota editor found by scoped search. Do not introduce a new feature. Exclude its nonexistent UserPlatformQuotaModal.spec.ts from the explicit Vitest command above; remaining 14 listed specs must execute. Report this omission and source evidence. All other scope/gates unchanged.
