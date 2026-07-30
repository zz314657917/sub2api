---
task_id: upstream-postcss-security-s114
repo: F:/mcplugins/sub2api
phase: done
owner: codex
source: upstream a5aae5db9
---

## Task ID

`upstream-postcss-security-s114`

## Role

Planner/Generator/Evaluator by Codex; no worker delegation is needed for this
narrow dependency-security patch.

## Goal

Port upstream `a5aae5db9` so every frontend dependency path resolves PostCSS
to at least `8.5.18`, covering the source-map vulnerabilities reported against
the local lockfile's `postcss@8.5.6`.

## Success Criteria

- `frontend/package.json` adds the `postcss@<8.5.18` override and raises the
  direct PostCSS requirement to `>=8.5.18`.
- `frontend/pnpm-lock.yaml` resolves PostCSS to `8.5.23` and follows only the
  expected `nanoid` patch update and peer/snapshot references.
- No frontend source, backend, database, deployment, container, or knowledge
  files change in this sprint.
- The audit result contains no PostCSS advisory; the repository-wide audit may
  retain the pre-existing `xlsx` high findings whose exception dates are
  already expired and must be reported as a baseline risk.
- Frontend typecheck, build, diff, and exact-path checks pass.

## Allowed Paths

- `frontend/package.json`
- `frontend/pnpm-lock.yaml`
- `docs/workflow/tasks/upstream-postcss-security-s114.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `frontend/src/**`
- `backend/**`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `frontend/src/views/admin/group-buy/**`
- The separate `E:/codex-worktrees/sub2api/group-buy-lifecycle-refund-hardening-s110` worktree
- Any path not listed under Allowed Paths

## Constraints

- Use the final upstream security commit only; do not apply the superseded
  `b5979050f`, `67bb446b5`, or `41456b69d` PostCSS attempts.
- Preserve lockfile version 9 and the existing pnpm dependency topology.
- Do not install or upgrade unrelated dependencies.
- Do not commit, push, deploy, or refresh containers from the primary worktree.

## Acceptance Commands

Run from `F:/mcplugins/sub2api/frontend`:

```powershell
corepack.cmd pnpm audit --prod --audit-level=high
corepack.cmd pnpm exec vue-tsc --noEmit
corepack.cmd pnpm run build
```

Run from `F:/mcplugins/sub2api`:

```powershell
git diff --check
git diff --name-only -- frontend/package.json frontend/pnpm-lock.yaml docs/workflow/tasks/upstream-postcss-security-s114.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output

- Changed paths, audit/typecheck/build output, and final PASS/FAIL conclusion.
- Unverified risk: no production deployment or live browser smoke.

## Stop Rules

- Stop if the lockfile requires unrelated dependency churn.
- Stop if audit still reports a high-severity PostCSS finding.
- Stop if any denied path changes.

## Budget

`worker_mode: codex-direct`
`qa_mode: runtime`
