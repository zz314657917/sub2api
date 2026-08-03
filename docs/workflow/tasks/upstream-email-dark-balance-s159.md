# Upstream Email Dark Balance S159

## Task ID

`upstream-email-dark-balance-s159`

## Role

Primary Codex performs Planner, direct implementation, and final Evaluator gates in sequence. No worker is used for this isolated upstream port.

## Goal

Port the user-visible intent of upstream commit `841544345`: keep the selected user's email address and current balance readable in the administrator balance modal when dark mode is active.

## Success Criteria

- The email uses a dark-mode foreground with adequate contrast on the existing `dark:bg-dark-700` surface.
- The current-balance line has a readable dark-mode secondary foreground.
- The local custom credit symbol/formatter behavior stays unchanged.
- The new balance-modal console-theme assertion, typecheck, scoped lint, diff, and integrity checks pass; unrelated full-suite baseline failures are recorded separately.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-email-alias-dedup-s157`
- Base: `a346b23d4`
- Upstream behavior: `841544345`
- The upstream patch does not apply directly because local code has a custom credit formatter, but the relevant DOM and background remain equivalent.

## Allowed Paths

- `frontend/src/components/admin/user/UserBalanceModal.vue`
- `frontend/src/__tests__/console-theme.spec.ts`
- `docs/workflow/tasks/upstream-email-dark-balance-s159.md`
- `docs/workflow/qa-reports/upstream-email-dark-balance-s159-qa.md`

## Denied Paths

- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- API, billing, credit formatting, translations, global theme configuration, backend, database, deployment, containers, and the primary worktree `F:/mcplugins/sub2api`

## Constraints

- Preserve the existing component layout and credit formatter.
- Apply only the two foreground contrast classes required by the upstream fix.
- Do not start a development server, change user data, or deploy.

## Acceptance Commands

```powershell
npm.cmd run test:run -- src/__tests__/console-theme.spec.ts
npx.cmd vitest run src/__tests__/console-theme.spec.ts --testNamePattern "keeps the balance modal user identity readable in dark mode"
npx.cmd eslint src/components/admin/user/UserBalanceModal.vue src/__tests__/console-theme.spec.ts
npm.cmd run typecheck
git diff --check
git status --short
```

## Output

- QA report: `docs/workflow/qa-reports/upstream-email-dark-balance-s159-qa.md`
- Final verdict must be `PASS`, `FAIL`, or `BLOCKED`, with executed checks and unverified risks.

## Stop Rules

- Stop if the required contrast change needs a global theme, component API, or credit-formatting change.
- Stop if validation requires a real admin session or production environment.
