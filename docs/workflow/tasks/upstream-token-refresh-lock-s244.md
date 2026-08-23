# Upstream Token Refresh Lock S244

## Task ID

`upstream-token-refresh-lock-s244`

## Role

Developer Worker and independent QA Worker both use `gpt-5.6-terra` in
separate executions and evidence paths. Codex is Planner and Final Evaluator.
The implementation must follow this approved contract without widening scope.

## Goal

Selectively port upstream `3445485eb` (merged by `5fc977846`) so a normal
proactive token refresh cannot mistake an unchanged token pair near the
two-minute expiry boundary for a completed peer refresh. After acquiring the
Web Lock, the caller must perform the refresh unless a peer has actually
published recognizable replacement state.

## Frozen Base And Provenance

- Frozen product base: local `main@5183430fb3373683e938227f34b328788991bac6`.
- Upstream source: `3445485ebc21f8912b95397d0d68e32f2e4c154e`.
- Upstream merge: `5fc9778468f90ff69caee7b9a8fe90600ecd74e4`.
- Upstream audit tip: `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- The complete two-file first-parent patch passes `git apply --check` on the
  frozen local product base; ancestry and patch scope must be rechecked before
  integration.

## Success Criteria

- `readPeerRefreshResult` accepts peer state only when the rotating refresh
  token changed, or when the existing failed-access-token reconciliation rule
  proves another request already replaced the failed token.
- An unchanged access token, refresh token, expiry timestamp, and user identity
  must never be treated as a completed peer refresh merely because the expiry
  remains just beyond a fixed buffer.
- With Web Locks available and no real peer update, a proactive refresh near
  the former 120-second boundary calls `/auth/refresh` exactly once and returns
  the new token pair.
- Existing same-document request sharing, rotated peer-token adoption,
  one-time refresh-token race recovery, different-user isolation, and logout
  protection remain unchanged.
- The obsolete boundary buffer constant and its unchanged-token shortcut are
  removed without restructuring the token refresh module.
- Focused tests repeated ten times, complete focused-file Vitest, frontend
  typecheck/build, formatting/diff, exact scope, provenance, conflict/index,
  and protected-main gates pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`,
  `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`, and this contract.
- Product owner: `frontend/src/api/tokenRefresh.ts`.
- Test owner: `frontend/src/api/__tests__/tokenRefresh.spec.ts`.

## Allowed Paths

- `frontend/src/api/tokenRefresh.ts`
- `frontend/src/api/__tests__/tokenRefresh.spec.ts`
- `docs/workflow/worker-results/upstream-token-refresh-lock-s244-result.md`
- `docs/workflow/qa-reports/upstream-token-refresh-lock-s244-qa.md`

## Denied Paths

- All backend files, other frontend files, dependencies, lockfiles, generated
  files, configuration, schema, migrations, deployment, and containers.
- `docs/workflow/status.md`, `docs/workflow/spec.md`,
  `docs/workflow/main-log.md`, this contract, and all `knowledge/**` paths are
  Controller-owned and denied to workers.
- All user-owned dirty and untracked paths in the primary worktree, including
  the current Pixel Cafe files and `outputs/`.
- Remote writes, push, force operations, history rewrites, real provider
  traffic, shared/production data, and browser automation.

## Constraints

- Keep the implementation behaviorally identical to the bounded upstream
  source; do not cherry-pick or merge divergent upstream history.
- Do not redesign refresh coordination, storage keys, timeouts, error handling,
  axios configuration, API response types, or authentication state.
- Preserve the refresh-token-last commit marker and current user-identity
  checks.
- Do not stage, overwrite, or revert unrelated work.

## Acceptance Commands

From `frontend/` in the isolated worktree:

```powershell
& .\node_modules\.bin\vitest.cmd run src/api/__tests__/tokenRefresh.spec.ts
1..10 | ForEach-Object {
  & .\node_modules\.bin\vitest.cmd run src/api/__tests__/tokenRefresh.spec.ts
  if ($LASTEXITCODE -ne 0) { throw "tokenRefresh focused iteration $_ failed" }
}
& .\node_modules\.bin\vue-tsc.cmd --noEmit
if ($LASTEXITCODE -ne 0) { throw "frontend typecheck failed" }
& .\node_modules\.bin\vue-tsc.cmd -b
if ($LASTEXITCODE -ne 0) { throw "frontend build typecheck failed" }
& .\node_modules\.bin\vite.cmd build
if ($LASTEXITCODE -ne 0) { throw "frontend Vite build failed" }
```

These direct local executables are the exact bodies of the repository's
`test`, `typecheck`, and `build` scripts. They replace `pnpm exec/run` for this
worktree because pnpm 11.19.0 automatically synchronizes the existing lockfile
before executing even a focused test. Do not run install/add/update or accept
any dependency, lockfile, or workspace metadata change.

From the worktree root:

```powershell
# Capture this before the Developer changes any file.
$dispatchBase = git rev-parse HEAD

# After the business commit, use the Controller-recorded commit identities.
git diff --check "$dispatchBase..$businessCommit"
git diff --name-only "$dispatchBase..$businessCommit"
git diff-tree --no-commit-id --name-only -r $businessCommit
git diff-tree --no-commit-id --name-only -r $evidenceCommit
git diff --exit-code -- frontend/pnpm-lock.yaml
if (Test-Path 'frontend/pnpm-workspace.yaml') { throw "unexpected pnpm workspace file" }
git diff --cached --name-only
git ls-files -u
git merge-base --is-ancestor 3445485ebc21f8912b95397d0d68e32f2e4c154e upstream/main
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' frontend/src/api/tokenRefresh.ts frontend/src/api/__tests__/tokenRefresh.spec.ts
```

`$businessCommit` and `$evidenceCommit` are the exact identities reviewed by
the Controller after the two required commits exist. Also verify the exact
allowlist for both commits, empty unmerged index, source/merge first-parent
scope, no user-dirty overlap, and preservation of the primary worktree's
tracked dirty patch ID plus untracked `outputs/` state.

## Output

- Developer produces one business commit containing only the two product/test
  paths and one separate evidence commit containing only
  `docs/workflow/worker-results/upstream-token-refresh-lock-s244-result.md`.
- The Developer report first line must be exactly
  `### DONE: upstream-token-refresh-lock-s244`,
  `### BLOCKED: upstream-token-refresh-lock-s244`, or
  `### FAILED: upstream-token-refresh-lock-s244`.
- Independent QA may modify only
  `docs/workflow/qa-reports/upstream-token-refresh-lock-s244-qa.md`; its first
  line must be exactly `### PASS: upstream-token-refresh-lock-s244`,
  `### FAIL: upstream-token-refresh-lock-s244`, or
  `### BLOCKED: upstream-token-refresh-lock-s244`.
- Reports list changed files, commands run, key output, risks, contract
  compliance, and `knowledge_candidates` without long unrelated logs.

## Stop Rules

- Stop if `gpt-5.6-terra` is unavailable; do not silently replace the model.
- Stop if implementation requires any path outside the allowlist, dependency
  change, authentication redesign, backend change, browser automation, or real
  external state.
- Stop if the focused selector discovers no tests, typecheck/build cannot run,
  or a failure is caused by an owner outside this contract.
- Stop on any unexpected primary-worktree protected-path change.

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `claude-bare-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- developer_max_budget_usd: `0.10`
- qa_max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Status

`contract-approved`

## Worker Output

Same requirements as `Output`; this compatibility heading is retained for the
worker dispatcher.
