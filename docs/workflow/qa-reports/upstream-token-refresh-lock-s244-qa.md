### PASS: upstream-token-refresh-lock-s244

# Independent QA Report

## Findings

- Independent QA accepted business commit `0f27ab7e4` and Controller evidence
  commit `a345614d4`. The business commit changes exactly
  `frontend/src/api/tokenRefresh.ts` and
  `frontend/src/api/__tests__/tokenRefresh.spec.ts`; the evidence commit changes
  exactly `docs/workflow/worker-results/upstream-token-refresh-lock-s244-result.md`.
- The business patch ID is `103c149ba901659c14be13616449cf2e25ae3d37`, equal
  to both upstream source `3445485eb` and the first-parent diff of merge
  `5fc977846`. `3445485eb` is an ancestor of `upstream/main`.
- Initial QA correctly stopped because it calculated the protected patch ID over
  an over-broad primary-worktree diff that included Controller-owned workflow
  state. Contract correction `8c3b13fdd` explicitly limits that gate to the
  eleven listed user-owned Pixel Cafe paths. The retest calculated only those
  paths and obtained the required
  `370ac77de0e2f530ab652b99fb3eb35e809f4c84`; Controller-owned workflow files
  were checked separately and were not part of that user patch ID.

## Commands Run

From the isolated QA worktree frontend, using direct existing executables only:

- `& .\node_modules\.bin\vitest.cmd run src/api/__tests__/tokenRefresh.spec.ts`
- Ten additional identical focused Vitest runs, each guarded for a nonzero
  exit code.
- `& .\node_modules\.bin\vue-tsc.cmd --noEmit`
- `& .\node_modules\.bin\vue-tsc.cmd -b`
- `& .\node_modules\.bin\vite.cmd build`

From the QA worktree and read-only primary worktree:

- Exact business/evidence commit scope, stable patch-ID, upstream ancestry,
  frozen-range `git diff --check`, lockfile/workspace, index, unmerged-index,
  and conflict-marker checks.
- Scoped eleven-path protected-primary patch-ID and path-list check; separate
  Controller-owned workflow-path check; staged/unmerged-index and untracked
  `outputs/` count check.

## Key Output

- Focused Vitest passed once plus ten repetitions: every run reported `1` test
  file passed and `7` tests passed.
- `vue-tsc --noEmit`: `VUE_TSC_NOEMIT_EXIT=0`; `vue-tsc -b`:
  `VUE_TSC_BUILD_EXIT=0`; Vite production build transformed `1880` modules,
  completed in `19.81s`, and returned `VITE_BUILD_EXIT=0`.
- Vite emitted only existing dynamic-import/chunk-size warnings; it did not
  report a build error. No pnpm command was run.
- Candidate frozen-range diff check and lockfile diff both exited `0`;
  `frontend/pnpm-workspace.yaml` is absent, staged index and unmerged index are
  empty, and no conflict marker was found in either product/test owner.
- The exact eleven primary user paths remain dirty with required stable patch
  ID `370ac77de0e2f530ab652b99fb3eb35e809f4c84`. The primary staged and
  unmerged indexes are empty; `outputs/` still has exactly two untracked files.

## Risks

- Browser automation, real provider traffic, production data, deployment,
  containers, and push were not exercised; all are denied by this contract.
- Vite warning output notes existing dynamic-import and large-chunk guidance.
  The production build nevertheless completed successfully and QA found no
  S244 scope expansion.

## Contract Compliance

- QA ran only in `E:/codex-worktrees/sub2api/upstream-token-refresh-lock-s244-qa`
  and used the Controller-provided task-local `node_modules` junction through
  direct `vitest.cmd`, `vue-tsc.cmd`, and `vite.cmd` invocation.
- No `pnpm`, install/add/update/exec/run command, dependency operation,
  lockfile/workspace change, business-file change, browser automation, remote
  write, provider call, deployment, container operation, or push occurred.
- This retest modifies only the allowed QA report. The initial FAIL was resolved
  by the Controller's contract-scoping correction, not by any product change.

## knowledge_candidates

- none
