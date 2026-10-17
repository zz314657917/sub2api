---
status: approved
review_verdict: PASS
task_id: upstream-v0216-frontend-20261015
worker_model: gpt-6.1-sol
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
spec_ref: docs/workflow/tasks/upstream-v0216-frontend-20261015.md
---

# Task Contract

## Task ID
upstream-v0216-frontend-20261015

## Role
Controller plans and integrates; separate gpt-6.1-sol reviewer, Developer and QA.

## Goal
Recover and validate the provisional six-commit frontend candidate already present in E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015. The candidate was created before the required contract review; this contract does not retroactively claim that gate passed. No further product edits or main-tree integration before independent review approves the forward recovery.

## Context
- User authorized selective local integration of upstream v0.2.16 and continued work, without push, deployment or new commits.
- Upstream tag object c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d peels to 01842576b3d5c1e2545b5a99ad555d36e28ed306.
- Provisional candidate HEAD a19b132ec; six independent cherry-picked commits exist only in the detached candidate. Preserve them as evidence; do not rewrite history or add commits.
- Reference behaviors c4224045d, bd2f53b85, d01628975, 2bfc49477, 66b7ce08e, 5753552b9.
- Main tree dirty UsageProgressBar files are EXCLUDED from this batch.
- GPT-6.1 Sol 500k/credential subscription exemption already exists in main commit 592d41387. Retain and include backend regression in the following gateway batch.

## Success Criteria
- All three payment QR countdowns catch up to elapsed time and stop polling on expiry.
- Multiplier edits reject nonfinite/negative input, preserve local zero-rate overrides, positive decimal/scientific input and clearing semantics. Existing requests.v0215 zero-rate tests remain unchanged and must pass; rejecting zero conflicts with the local contract.
- IME composition is not intercepted by Select navigation keys.
- datetime-local inputs initialize using local time and preserve existing UTC API serialization.
- Account group names can wrap without changing default GroupBadge truncation elsewhere.
- Mandatory AccountGroupsCell.v0216.spec.ts covers full names and wrapping classes, default GroupBadge truncation, and +N popover expansion/closing.
- Gemini authorization failures expose the existing sanitized API error message.
- New upstream tests execute against actual local components; repair fixtures only for demonstrated local contracts, never weaken behavioral assertions.
- Run relevant existing regressions, scoped ESLint, full vue-tsc and production Vite build in isolated candidate. No runtime/payment/provider verification is claimed from these checks.
- Independent QA reviews every changed path and executes the acceptance commands. Main integration only after QA PASS and byte hash recheck of each target against captured pre-integration state. Preserve all unrelated tracked and untracked files.

## Allowed Paths
- frontend/src/components/payment/PaymentQRDialog.vue
- frontend/src/components/payment/PaymentStatusPanel.vue
- frontend/src/views/user/PaymentQRCodeView.vue
- frontend/src/components/payment/__tests__/paymentCountdown.spec.ts
- frontend/src/components/admin/group/GroupRateMultipliersModal.vue
- frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts
- frontend/src/components/common/Select.vue
- frontend/src/components/common/__tests__/Select.ime.spec.ts
- frontend/src/views/admin/ops/components/OpsDashboardHeader.vue
- frontend/src/views/admin/ops/components/__tests__/OpsDashboardHeader.timezone.spec.ts
- frontend/src/components/account/AccountGroupsCell.vue
- frontend/src/components/common/GroupBadge.vue
- frontend/src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts
- frontend/src/composables/useGeminiOAuth.ts
- frontend/src/composables/__tests__/useGeminiOAuth.errors.spec.ts
- docs/workflow/worker-results/upstream-v0216-frontend-20261015.md
- docs/workflow/qa-reports/upstream-v0216-frontend-20261015.md

## Denied Paths
All other product files, backend, migrations, dependencies/lockfiles, main-tree existing dirty files, shared knowledge, global memories, production data/configuration, containers. Review may write only its named review document; workflow phase updates belong to controller.

## Constraints
- No new git commits, staging, push, merge/rebase/cherry-pick or deployment.
- Existing isolated node_modules restored offline from frozen lockfile; no package/lockfile changes. Vite output is candidate backend/internal/web/dist (ignored), per vite.config.ts; do not copy generated output to main.
- Browser verification, if performed, uses exact task profile/process cleanup required by AGENTS.md; otherwise explicitly leave browser verification unverified.
- Preserve previous local design, payments, subscriptions, digital-store isolation and layout.

## Acceptance Commands
From candidate frontend:
```powershell
node node_modules/vitest/vitest.mjs run src/components/payment/__tests__/paymentCountdown.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.v0215.spec.ts src/components/common/__tests__/Select.ime.spec.ts src/views/admin/ops/components/__tests__/OpsDashboardHeader.timezone.spec.ts src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts src/composables/__tests__/useGeminiOAuth.errors.spec.ts
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build
$env:NODE_PATH='E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/node_modules/.pnpm/node_modules'
# Run scoped ESLint with no --fix on every changed .vue/.ts file.
```
Controller/QA additionally run relevant existing component tests (discover exact names), ESLint on changed files with no --fix, git diff --check for uncommitted changes AND base..candidate commits. Every named test must execute with a nonzero count; missing or skipped suites cannot satisfy acceptance. Logs under E:/codex-runtime/upstream-v0216-20261015/.

## Output
Developer report starts `### DONE: upstream-v0216-frontend-20261015` or FAILED/BLOCKED, includes changed paths, executed checks, remaining risks. QA report starts PASS/FAIL/BLOCKED, reports candidate hash manifest, unverified runtime separately.

## Stop Rules
Stop on unavailable configured model, required denied-path edits, owner drift, unsafe fixture weakening or repeated unresolved failure. Do not silently substitute a model or widen scope.
