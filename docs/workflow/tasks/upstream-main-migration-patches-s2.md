# Task Contract

## Task ID
upstream-main-migration-patches-s2

## Role
Codex acts as Planner, Generator, and Final Evaluator for this Sprint. Implement only the migration-sized upstream patch selected here, and stop if conflicts require broader upstream architecture changes.

## Goal
Port upstream `f597c1581 feat(group): 支持自定义 /v1/models 模型列表` onto the current local branch that already contains Sprint 1. This Sprint adds per-group custom `/v1/models` response configuration while preserving local Canvas, tickets, billing/payment, public UI, and support-ticket work.

## Success Criteria
- Admin groups can configure the models list behavior for `/v1/models` without affecting normal chat routing or billing.
- Gateway `/v1/models` responses honor the group-level models list config for API key requests.
- The new group config is persisted through Ent schema, generated Ent code, and a local migration whose filename is renumbered to the next available local migration.
- API key auth/cache invalidation covers the new group field so stale group model-list config is not served.
- Frontend admin group UI can edit and validate the models-list config, with focused tests.
- All selected upstream changes are ported by cherry-pick or equivalent minimal adaptation; skipped conflicts are documented.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-safe-patches-s1`
- Work branch: `codex/upstream-main-migration-patches-s2`
- Upstream source commit: `f597c1581`
- Main worktree `F:/mcplugins/sub2api` has unrelated dirty changes and must not be modified.
- Local migrations currently extend beyond upstream's `143_group_models_list_config.sql`; use the next local migration number instead of overwriting or back-numbering.

## Allowed Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/domain/models_list_config.go`
- `backend/internal/handler/admin/**`
- `backend/internal/handler/dto/**`
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
- `frontend/src/api/admin/groups.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/groupsModelsList*.spec.ts`
- `frontend/src/views/admin/groupsModelsList.ts`
- `frontend/src/views/admin/groupsModelsListCandidates.ts`
- `docs/workflow/tasks/upstream-main-migration-patches-s2.md`
- `docs/workflow/worker-results/upstream-main-migration-patches-s2-result.md`
- `docs/workflow/qa-reports/upstream-main-migration-patches-s2-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `backend/cmd/server/**`
- `backend/internal/handler/auth_dingtalk*`
- `backend/internal/handler/admin/setting_handler_dingtalk*`
- `backend/ent/schema/user_platform_quota.go`
- `backend/ent/userplatformquota*`
- `backend/internal/handler/*user_platform_quota*`
- `backend/internal/handler/quotaview/**`
- Payment, subscription notify, redeem expiry, channel monitor API mode, OpenAI WS/Responses bridge redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Prefer cherry-pick of `f597c1581`, but resolve conflicts by preserving local features and adapting only the selected group-model-list behavior.
- If Ent generation tries to remove local schemas/entities, stop and restore local schemas before continuing.
- Do not overwrite local support-ticket, Canvas, billing/payment, public Model Plaza, or local admin changes.
- Rename upstream migration `143_group_models_list_config.sql` to the next available local migration number, and verify migration ordering.
- Do not include generated frontend build output, `node_modules`, Docker smoke artifacts, or `tmp/**` in commits.

## Candidate Commit
- Primary: `f597c1581 feat(group): 支持自定义 /v1/models 模型列表`

## Explicitly Deferred
- `eba204632` OpenAI OAuth refresh enrichment: wire/cmd-server wiring and wider service behavior; separate Sprint 2b.
- `bf24b6113`, `b60d8bb4c`: admin usage performance/deleted-user history; separate usage Sprint.
- `57d9e15e0`: sync upstream models on account create; separate admin-account Sprint.
- `user_platform_quotas`, DingTalk OAuth, payment/subscription/redeem/channel monitor migrations: separate high-risk migration Sprint.
- OpenAI gateway / WS / Responses bridge redesign and response.failed stream handling: Sprint 3.

## Acceptance Commands
```powershell
git status --short --branch
go generate ./ent
go test ./internal/service ./internal/handler ./internal/repository -run "Group|Models|APIKey|Gateway" -count=1
go test ./internal/server/routes ./cmd/server -count=1
go test ./internal/service ./internal/handler ./internal/repository -count=1
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run lint:check
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/groupsModelsList.spec.ts src/views/admin/__tests__/groupsModelsListCandidates.spec.ts src/views/admin/__tests__/groupsModelsListLayout.spec.ts
```

## Output
- Write `docs/workflow/worker-results/upstream-main-migration-patches-s2-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-migration-patches-s2-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval, implementation, and QA events.

## Stop Rules
- Stop if implementing the selected commit requires touching denied paths.
- Stop if Ent generation removes or rewrites unrelated local schemas/entities.
- Stop if resolving conflicts requires adopting upstream's broad gateway/bridge architecture.
- Stop if tests fail for reasons requiring broader schema/API/config changes than this contract allows.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
