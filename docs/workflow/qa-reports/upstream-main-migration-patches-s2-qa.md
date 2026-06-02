### PASS: upstream-main-migration-patches-s2

# upstream-main-migration-patches-s2 QA Report

## Task ID
upstream-main-migration-patches-s2

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-migration-patches-s2.md`

## Evidence
- diff reviewed: yes
- denied paths touched: no
- migration ordering checked: yes, local file is `backend/migrations/167_group_models_list_config.sql`
- commands run:
```text
git status --short --branch -> clean on codex/upstream-main-migration-patches-s2
go generate ./ent -> pass
git status --short --branch after go generate -> clean
go test ./internal/service ./internal/handler ./internal/repository -run "Group|Models|APIKey|Gateway" -count=1 -> pass
go test ./internal/server/routes ./cmd/server -count=1 -> pass
go test ./internal/service ./internal/handler ./internal/repository -count=1 -> pass
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/groupsModelsList.spec.ts src/views/admin/__tests__/groupsModelsListCandidates.spec.ts src/views/admin/__tests__/groupsModelsListLayout.spec.ts -> pass, 3 files / 12 tests
```
- manual checks:
```text
Cherry-pick source commit -> f597c1581
Implementation commit -> cab7d4bf0
Ent generation re-run produced no working-tree diff
Migration renumbered from upstream 143 to local 167
Local ModelCatalog test retained and custom /v1/models tests added
Modular i18n preserved; text added under frontend/src/i18n/locales/*/admin/groups.ts
```

## Findings
- 未发现当前 Sprint 2 补丁引入的明确阻断问题。
- 后端 focused tests、后端路由/启动包测试、后端核心包测试全部通过。
- 前端 `typecheck`、`lint:check` 和目标 Vitest 全部通过。
- 仅存在一个合同路径适配点：本地 i18n 已模块化，因此实际更新的是 `frontend/src/i18n/locales/en/admin/groups.ts` 和 `frontend/src/i18n/locales/zh/admin/groups.ts`，未触碰 denied paths。

## Bug Owner Recommendation
none

## Root Cause
- none

## Retest Scope
- None.

## Unverified Risks
- 未执行真实数据库迁移 smoke。
- 未执行完整 Docker runtime smoke；本 Sprint 合同未要求 Docker smoke，且 S1 已覆盖 runtime smoke。

## Knowledge Promotion
- none
