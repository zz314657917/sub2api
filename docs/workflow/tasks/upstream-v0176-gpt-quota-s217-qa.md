---
task_id: upstream-v0176-gpt-quota-s217-qa
phase: qa
role: Evaluator
worker_model: gpt-5.6-terra
implementation_head: a8105bd5baa349cb625253381e01210f5bc74ef2
---

# S217 Independent QA Contract

## Goal

Independently verify the committed S217 behavioral port against
`docs/workflow/tasks/upstream-v0176-gpt-quota-s217.md`. Do not implement or fix
product code. The Developer and QA roles remain independent even though both
use `gpt-5.6-terra`.

## Required Checks

- Confirm the worktree is clean and `HEAD` is exactly `a8105bd5b` or a
  descendant containing only this QA contract before testing.
- Review `e59573c6a..a8105bd5b` for contract compliance, exact allowed paths,
  unresolved conflicts, unmerged index entries, and upstream provenance.
- Run the S217 focused service regressions with `-count=10`, complete
  `internal/service`, focused reset/refresh handler tests, route contract,
  complete `internal/server`, and server compilation.
- Run the two focused frontend Vitest files. A temporary junction from this
  worktree's `frontend/node_modules` to
  `F:/mcplugins/sub2api/frontend/node_modules` is allowed only for the test and
  must be removed before the verdict. Do not install dependencies or change
  manifests/lockfiles.
- If that reused tree lacks `.bin` launchers, directly invoke the existing
  package entrypoints under `.pnpm`: `vitest@*/node_modules/vitest/vitest.mjs`
  and `vue-tsc@*/node_modules/vue-tsc/bin/vue-tsc.js`. Missing `.bin` wrappers
  alone are not a blocker when these approved package entrypoints execute.
- Run `vue-tsc --noEmit` and the build typecheck step. The known baseline on
  both local main and S217 is:
  `src/views/user/AirwallexPaymentView.vue(103,36): TS2307 Cannot find module
  '@airwallex/components-sdk'`. Treat it as an unverified baseline gap only if
  S217 produces the same single failure and no additional diagnostic.
- Confirm every HTTP fixture is local and no provider, production, container,
  deployment, database migration, or remote push action occurred.

## Allowed Write

- `docs/workflow/qa-reports/upstream-v0176-gpt-quota-s217-qa.md`

All product files, workflow status, knowledge files, Git history, dependencies,
user-owned main-worktree changes, and `outputs/` are read-only.

## Output

Write the QA report with first line exactly one of:

- `### PASS: upstream-v0176-gpt-quota-s217`
- `### FAIL: upstream-v0176-gpt-quota-s217`
- `### BLOCKED: upstream-v0176-gpt-quota-s217`

Include Findings, Executed Checks, Unverified Risks, contract compliance, and a
final recommendation. Do not claim PASS from the Developer report alone.

## Stop Rules

- Return `FAIL` for a product defect, out-of-scope product change, unresolved
  conflict, dirty residual, real provider access, or a new S217-specific test,
  type, or build failure.
- Return `FAIL` if the default-tag route-contract command reports
  `[no tests to run]` or the test is not discoverable with `go test -list`.
- Return `BLOCKED` when a required command cannot execute and no approved local
  dependency reuse can make it executable.
- Do not edit product code, fix tests, install packages, integrate, push,
  deploy, or delete branches/worktrees.
