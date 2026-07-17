# Task Contract: upstream-main-compat-s78

## Task ID

`upstream-main-compat-s78`

## Status

`approved`

## Role

Generator coordinator for a narrowly scoped frontend compatibility batch from
the live upstream `main`. The batch is behavior-level and must preserve the
local frontend theme and payment behavior; it is not an upstream history
merge.

## Goal

Port the low-risk UI/runtime slices from upstream commits `08e994ad8` and
`980a46ffb` onto local baseline `6113b0f5e`: lazy-load the Stripe SDK through
its side-effect-free entrypoint and fill the missing OpenAI account auth
labels. `174ea22ee` is explicitly deferred because this local checkout does
not yet have the upstream Grok Codex configuration template it modifies.

## Success Criteria

- Stripe payment views dynamically import `@stripe/stripe-js/pure`; Stripe is
  isolated from the generic vendor chunk without changing payment API or
  checkout behavior.
- English and Chinese OpenAI account forms resolve the Mobile RT and AT labels
  without fallback strings.
- Focused frontend tests, typecheck, production build, allowlist audit, and
  `git diff --check` pass; no backend, migration, billing, scheduler, VERSION,
  deployment, lockfile, or `knowledge/**` path changes occur.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-main-compat-s78`
- Baseline: `6113b0f5ebdd454869f4a21b4ceb9d332034c3db`
- Upstream snapshot: `bc50c9d01336dc2af3225b31a13a8d6c8f213b1f`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, and this contract.

## Allowed Paths

- `frontend/src/components/payment/StripePaymentInline.vue`
- `frontend/src/views/user/StripePaymentView.vue`
- `frontend/src/views/user/StripePopupView.vue`
- `frontend/src/views/user/__tests__/StripePaymentView.spec.ts`
- `frontend/src/views/user/__tests__/stripeLazyLoading.spec.ts`
- `frontend/vite.config.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-compat-s78.md`
- `docs/workflow/worker-results/upstream-main-compat-s78-*-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s78-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `backend/**`, `backend/ent/**`, `backend/migrations/**`, generated code, and Wire output.
- `frontend/src/api/payment.ts`, backend payment API/handler/service code,
  billing, pricing, settlement, scheduler, account schema, or authentication logic.
- `deploy/**`, Docker/container files, `VERSION`, README, lockfiles, and dependency installation.
- `knowledge/**` and global memory files.
- Any frontend path not listed in Allowed Paths.
- Upstream commits outside `08e994ad8` and `980a46ffb`.

## Constraints

- Work only in this isolated worktree; do not touch the dirty primary checkout.
- Adapt to the local component structure and theme; do not replace local UI wholesale.
- Keep Stripe changes limited to loading/chunk boundaries, with no checkout or API semantics changes.
- Build-generated `backend/internal/web/dist/**` output is ignored/generated and
  is not part of the S78 change boundary.
- Preserve all existing generated config fields and test expectations; do not
  introduce a new Grok Codex configuration template in this slice.
- Do not push, deploy, update containers, or merge this S78 branch automatically.

## Acceptance Commands

```powershell
Push-Location frontend
npm.cmd run test:run -- src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/stripeLazyLoading.spec.ts
if ($LASTEXITCODE -ne 0) { throw "S78 focused Vitest failed" }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw "S78 typecheck failed" }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw "S78 production build failed" }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw "S78 diff check failed" }
```

Evaluator additionally audits changed paths against this allowlist and checks
that no conflict markers or unmerged index entries remain.

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s78-*-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s78-qa.md`
- Main workflow log entry for contract review, implementation, QA, and final verdict.

## Stop Rules

- Stop if a change requires backend, payment business logic, dependency/lockfile changes, or a denied path.
- Stop if the local UI would need a broad rewrite to accommodate an upstream hunk.
- Stop after two failed attempts on the same slice; return to Planner for re-splitting.

## Budget

- worker_mode: `Codex direct implementation; no external worker required for this small UI batch`
- qa_worker_mode: `Codex focused evaluator`
- worktree_root: `E:/codex-worktrees`
