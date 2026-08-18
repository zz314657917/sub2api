---
task_id: upstream-ops-small-s227
phase: done
base: 3ccb86afc42c21752a4890a02101fc9547d8978f
---

# Task Contract

## Role

Codex Controller/Generator；本批次不调用 worker，先在隔离 worktree 实现，再用新的 QA worktree 独立复核。

## Goal

选择性移植上游 `e8ff2017c` 与 `943f09d35` 的两个独立 Ops 前端修复：错误分布图例显示分类名称，内存容量按合理的 MB/GB 单位显示。保持本地 S226、用户 dirty/untracked 内容和远端状态不变。

## Success Criteria

- 错误分布图例同时显示分类名称和计数。
- Ops 内存容量小于 1 GiB 显示 MB，达到 1 GiB 显示 GB；无效值显示 `-`。
- 相关 focused Vitest、前端 typecheck 和 scoped diff 检查通过。
- 仅允许路径发生变化，主工作区用户 patch IDs、未跟踪教程文件和 `outputs/` 保持不变。

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`
- Upstream: `upstream/main@8869775ed385fe985e05d4e9f414c9062b64af5a`
- Upstream commits: `e8ff2017c`, `943f09d35`

## Allowed Paths

- `frontend/src/views/admin/ops/components/OpsErrorDistributionChart.vue`
- `frontend/src/views/admin/ops/components/OpsDashboardHeader.vue`
- `frontend/src/views/admin/ops/utils/opsFormatters.ts`
- `frontend/src/views/admin/ops/utils/__tests__/opsFormatters.spec.ts`
- `docs/workflow/worker-results/upstream-ops-small-s227-result.md`
- `docs/workflow/qa-reports/upstream-ops-small-s227-qa.md`

## Denied Paths

- `backend/**`
- `frontend/src/components/account/**`
- `frontend/src/views/public/**`
- `knowledge/**`
- `backend/migrations/**`
- `outputs/**`
- `Dockerfile`, `backend/Dockerfile`, `deploy/Dockerfile`
- `package.json`, `pnpm-lock.yaml`, `backend/go.mod`
- push、部署、容器、共享/生产数据库和真实 provider 调用

## Constraints

- 保持最小改动，不整包 cherry-pick 或 merge 上游分支。
- 不覆盖或回滚主工作区已有 dirty/untracked 内容。
- 不新增依赖，不修改 manifest 或 lockfile。
- 先在 `E:/codex-worktrees/sub2api/upstream-ops-small-s227` 实现；QA 使用独立 worktree。

## Acceptance Commands

```powershell
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/utils/__tests__/opsFormatters.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
git diff --check -- frontend/src/views/admin/ops/components/OpsErrorDistributionChart.vue frontend/src/views/admin/ops/components/OpsDashboardHeader.vue frontend/src/views/admin/ops/utils/opsFormatters.ts frontend/src/views/admin/ops/utils/__tests__/opsFormatters.spec.ts
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor e8ff2017c061dd559b4f9ac0b7a5ada72573118a upstream/main
git merge-base --is-ancestor 943f09d357f53afaf1caf5cf11fa32c1fa60fdc9 upstream/main
```

## Output

- Controller implementation commit on the isolated branch.
- `docs/workflow/worker-results/upstream-ops-small-s227-result.md`, first line `### DONE: upstream-ops-small-s227`.
- Independent QA report `docs/workflow/qa-reports/upstream-ops-small-s227-qa.md`, first line `### PASS: upstream-ops-small-s227` or an explicit FAIL/BLOCKED verdict.

## Stop Rules

- Stop on any denied-path change, dependency/lockfile change, conflict, unresolved index entry, or protected-main change.
- Stop if focused tests are undiscoverable or typecheck cannot run; report the verification blocker instead of claiming PASS.
- Do not expand this batch into CN provider, fingerprint, Docker, migration, deployment, or push work.
