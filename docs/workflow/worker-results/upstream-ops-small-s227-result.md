### DONE: upstream-ops-small-s227

## Changed Files

- `frontend/src/views/admin/ops/components/OpsErrorDistributionChart.vue`
- `frontend/src/views/admin/ops/components/OpsDashboardHeader.vue`
- `frontend/src/views/admin/ops/utils/opsFormatters.ts`
- `frontend/src/views/admin/ops/utils/__tests__/opsFormatters.spec.ts`

Business commits preserve upstream provenance:

- `e1e6b7e7c` cherry-picks `943f09d35` with matching patch-id `6eaa4c2dcf0ca88dacd5559d6e62fa0e67c77620`.
- `86d8d597c` cherry-picks `e8ff2017c`.

## Commands

- `corepack.cmd pnpm --dir frontend install --frozen-lockfile` — PASS; worktree-local dependencies restored, no manifest or lockfile changes.
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/utils/__tests__/opsFormatters.spec.ts` — PASS, 1 file / 3 tests.
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops` — PASS, 7 files / 24 tests.
- `corepack.cmd pnpm --dir frontend run typecheck` — PASS.
- `git diff --check main...HEAD` — PASS.
- `git diff --name-only --diff-filter=U` and `git ls-files -u` — PASS; no conflicts or unresolved index entries.
- Upstream ancestry checks for `e8ff2017c` and `943f09d35` — PASS.

## Scope And Risks

- Diff is limited to the four S227 allowlisted frontend files.
- Browserslist emitted its existing stale-data advisory during Vitest; tests still passed.
- No browser session, backend, database, provider, deployment, container, or remote push was used.

## Knowledge Candidates

- None.
