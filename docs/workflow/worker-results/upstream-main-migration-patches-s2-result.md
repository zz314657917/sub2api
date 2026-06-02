### DONE: upstream-main-migration-patches-s2

## Task ID
upstream-main-migration-patches-s2

## Status
done

## Summary
- Ported upstream `f597c1581 feat(group): 支持自定义 /v1/models 模型列表` onto `codex/upstream-main-migration-patches-s2`.
- Added per-group custom `/v1/models` list configuration across Ent schema/codegen, service/domain, API key auth cache snapshots, gateway model list responses, admin group API, and admin group UI.
- Renumbered upstream migration from `143_group_models_list_config.sql` to local `167_group_models_list_config.sql`.
- Preserved local `ModelCatalog` behavior and local modular i18n structure while resolving cherry-pick conflicts.

## Changed Files
- `backend/ent/group.go`
- `backend/ent/group/group.go`
- `backend/ent/group_create.go`
- `backend/ent/group_update.go`
- `backend/ent/migrate/schema.go`
- `backend/ent/mutation.go`
- `backend/ent/runtime/runtime.go`
- `backend/ent/schema/group.go`
- `backend/internal/domain/models_list_config.go`
- `backend/internal/handler/admin/admin_basic_handlers_test.go`
- `backend/internal/handler/admin/admin_service_stub_test.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_models_test.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/group_repo.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/api_key_auth_cache.go`
- `backend/internal/service/api_key_auth_cache_impl.go`
- `backend/internal/service/api_key_auth_cache_version_test.go`
- `backend/internal/service/group.go`
- `backend/internal/service/group_models_list.go`
- `backend/migrations/167_group_models_list_config.sql`
- `frontend/src/api/admin/groups.ts`
- `frontend/src/i18n/locales/en/admin/groups.ts`
- `frontend/src/i18n/locales/zh/admin/groups.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/groupsModelsList.spec.ts`
- `frontend/src/views/admin/__tests__/groupsModelsListCandidates.spec.ts`
- `frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts`
- `frontend/src/views/admin/groupsModelsList.ts`
- `frontend/src/views/admin/groupsModelsListCandidates.ts`

## Commands Run
```text
git status --short --branch -> clean on codex/upstream-main-migration-patches-s2 before implementation and before QA
git cherry-pick f597c1581 -> conflicts resolved in gateway handler/tests, auth cache snapshot, and modular i18n
go generate ./ent -> pass; generated Ent output stable on re-run
git cherry-pick --continue -> success, commit cab7d4bf0
go test ./internal/service ./internal/handler ./internal/repository -run "Group|Models|APIKey|Gateway" -count=1 -> pass
go test ./internal/server/routes ./cmd/server -count=1 -> pass
go test ./internal/service ./internal/handler ./internal/repository -count=1 -> pass
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/groupsModelsList.spec.ts src/views/admin/__tests__/groupsModelsListCandidates.spec.ts src/views/admin/__tests__/groupsModelsListLayout.spec.ts -> pass, 3 files / 12 tests
```

## Conflict Handling
- `backend/internal/handler/gateway_handler.go`: kept local `ModelCatalog` and added upstream custom `/v1/models` helpers.
- `backend/internal/handler/gateway_models_test.go`: kept local model catalog test and added upstream custom models-list tests.
- `backend/internal/service/api_key_auth_cache_impl.go`: kept local multi-group snapshot helper flow, added `ModelsListConfig`, and bumped auth snapshot version to `13`.
- `frontend/src/i18n/locales/en.ts` and `frontend/src/i18n/locales/zh.ts`: kept local modular aggregate files; added new text to `frontend/src/i18n/locales/*/admin/groups.ts`.

## Scope Notes
- No direct merge of `upstream/main`.
- No DingTalk, `user_platform_quotas`, payment/redeem/channel-monitor migration, deploy, README, asset, or bridge redesign changes.
- Local i18n module files were touched instead of upstream single-file i18n blocks because the local branch has already split locale content into modules.

## Risks
- Real deployed migration execution was not run in this Sprint; only SQL ordering and compile/test behavior were verified locally.
- Broader upstream migration candidates remain deferred.

## Knowledge Candidates
- None.

## Contract Compliance
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
- allowed_path_adaptation: local modular i18n files used in place of aggregate `en.ts`/`zh.ts`

## Blocked Reason
- None.
