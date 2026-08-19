### BLOCKED: upstream-v178-ui-polish-s234

## Contract Compliance

- Worktree: `E:/codex-worktrees/sub2api/upstream-v178-ui-polish-s234`
- Frozen base: `main@e850690ce`
- No business implementation is committed because the contract acceptance
  environment is unavailable.
- The `announcements.ts` allowlist amendment was read before the attempted
  announcement adaptation. All attempted frontend source and test changes were
  discarded after the environment blocker, so this report is the only retained
  worktree change.

## Intended Local Adaptations

- Ops error details custom range propagation with `start_time` and `end_time`,
  plus an incomplete-range `1h` fallback and endpoint watchers.
- Localized header role, native dark date-control color scheme, dashboard cache
  creation/read token breakdown, first-announcement empty copy, and neutral
  zero-request SLA display.

## Commands And Results

- `pnpm exec vitest run ...`: BLOCKED before Vitest started.
- `pnpm run typecheck`: BLOCKED before type checking started.
- `pnpm run build`: BLOCKED before build started.
- Each command stopped in pnpm dependency-status validation with
  `ERR_PNPM_IGNORED_BUILDS`: `esbuild@0.21.5` and `vue-demi@0.14.10` require
  approval through `pnpm approve-builds`.
- The worktree had no usable frontend `node_modules`; approving ignored build
  scripts or changing dependency/policy state is outside the contract.
- `git diff --check`, allowlist, conflict/unmerged-index, and source ancestry
  gates cannot establish implementation acceptance because no implementation
  remains. The six designated source commits are recorded in the approved
  contract for the next run.

## Risks

- None of the six UI/Ops behaviors has a runnable focused-test, typecheck, or
  build verdict in this worktree.
- Retry from a task-authorized environment with prebuilt frontend dependencies
  or explicit approval to prepare them, then recreate and verify the scoped
  implementation before any business commit.

## knowledge_candidates

- No durable knowledge candidate. The pnpm ignored-build policy is environment
  state and should not be recorded as a repository rule without a successful
  authorized retry.
