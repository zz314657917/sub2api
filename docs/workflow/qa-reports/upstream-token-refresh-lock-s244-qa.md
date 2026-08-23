### FAIL: upstream-token-refresh-lock-s244

# Independent QA Report

## Findings

- QA stopped before executing frontend acceptance commands because the required
  primary-worktree protection gate failed. The contract requires exactly 11
  tracked dirty files with patch ID
  `370ac77de0e2f530ab652b99fb3eb35e809f4c84`; read-only inspection instead
  found 12 tracked dirty paths and patch ID
  `97ae91cd822976bf9709be3d5617cdadc21a3708`.
- The extra tracked dirty path is `docs/workflow/main-log.md`. The candidate QA
  worktree itself was clean before this report and is based at controller
  evidence `a345614d4` following business commit `0f27ab7e4`.
- Candidate business scope is exact: `0f27ab7e4` changes only
  `frontend/src/api/tokenRefresh.ts` and
  `frontend/src/api/__tests__/tokenRefresh.spec.ts`; candidate evidence
  `a345614d4` changes only the approved worker result report. Upstream source
  `3445485eb` and merge `5fc977846` first-parent both have the same two product
  paths.

## Commands Run

- `git status --short`, `git log --oneline --decorate -8`, and commit identity
  / scope inspection in the isolated QA worktree.
- `git merge-base --is-ancestor 5183430fb3373683e938227f34b328788991bac6 HEAD`,
  `git diff --check 5183430fb3373683e938227f34b328788991bac6..0f27ab7e4`, and
  first-parent upstream path checks.
- Read-only protected-primary checks:
  `git status --short`, `git diff | git patch-id --stable`,
  `git diff --cached --name-only`, `git ls-files -u`, and an `outputs/`
  untracked-state query from `F:/mcplugins/sub2api`.

## Key Output

- `baseAncestor=0`; candidate range diff check exited `0`.
- Candidate scopes: business = two allowed product/test paths; evidence = one
  allowed worker-result path.
- Primary worktree: `git diff --cached --name-only` and `git ls-files -u` were
  empty; `outputs/` still listed two untracked files. Its tracked dirty
  baseline, however, did not match the contract:
  `97ae91cd822976bf9709be3d5617cdadc21a3708 0000000000000000000000000000000000000000`.

## Risks

- Running direct Vitest, vue-tsc, or Vite acceptance after a protected-primary
  baseline mismatch would violate the contract's stop rule and could incorrectly
  certify a moving primary-worktree state. Those commands were deliberately not
  run; no behavioral PASS is claimed.
- No pnpm command, dependency operation, browser automation, provider call,
  deployment, container, push, or business-file modification was performed.

## Contract Compliance

- QA used the isolated `E:/codex-worktrees/sub2api/upstream-token-refresh-lock-s244-qa`
  worktree and modified only this permitted QA report.
- The protected-primary baseline mismatch is an explicit stop condition in the
  contract; QA stopped immediately on detection. Re-establish the approved
  primary baseline (or amend the contract with a newly recorded baseline) before
  a fresh independent QA reruns the required direct local frontend commands.

## knowledge_candidates

- none
