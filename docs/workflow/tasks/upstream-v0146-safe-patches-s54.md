# Task Contract: upstream-v0146-safe-patches-s54

## Task ID

`upstream-v0146-safe-patches-s54`

## Role

Generator / Codex direct integration.

## Goal

Port the first safe subset from upstream `v0.1.146`:

- `089a7b7fa` / `fa70a7217`: API Key list exposes real-time API Key concurrency counts.
- `2fb212b7` / `75e308949` / `ddd63a840`: normalize inbound `/responses/compact` endpoint accounting.
- `1c0ccb477` / `cb151e36e`: OAuth account tests send required Codex CLI headers and respect custom User-Agent.

## Success Criteria

- User API Key list responses include `current_concurrency` without breaking existing multi-group route fields.
- API Key request lifecycle tracks and releases stats-only API Key concurrency slots for normal HTTP and OpenAI WebSocket turn paths.
- `/responses/compact` inbound endpoint normalization matches upstream selected commits.
- OAuth account test requests include the selected Codex header fixes.
- Changes are committed on `codex/upstream-v0146-s54-safe-patches` and validated with targeted backend/frontend checks.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/dto/api_key_mapper_last_used_test.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/endpoint.go`
- `backend/internal/handler/endpoint_test.go`
- `backend/internal/handler/gateway_helper.go`
- `backend/internal/handler/gateway_helper_hotpath_test.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/stream_error_event.go`
- `backend/internal/repository/concurrency_cache.go`
- `backend/internal/repository/concurrency_cache_integration_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/api_key.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_delete_test.go`
- `backend/internal/service/concurrency_service.go`
- `backend/internal/service/concurrency_service_test.go`
- `backend/internal/service/wire.go`
- `backend/internal/server/api_contract_test.go`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/en/keys.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/i18n/locales/zh/keys.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/__tests__/KeysView.spec.ts`
- `docs/workflow/tasks/upstream-v0146-safe-patches-s54.md`
- `docs/workflow/worker-results/upstream-v0146-safe-patches-s54-result.md`
- `docs/workflow/qa-reports/upstream-v0146-safe-patches-s54-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `.github/**`
- `README*`
- unrelated payment, welfare, public-page, auth-theme, or visual redesign files
- any upstream `v0.1.146` feature not listed in the Goal

## Constraints

- Do not merge `upstream/main` or tag `v0.1.146` directly.
- Preserve local Studio Bridge / multi-group route / account-pool strategy behavior.
- For API Key concurrency, adapt upstream logic to the local `ProvideAPIKeyService(..., settingService)` signature instead of replacing it.
- Keep all work in the isolated worktree; do not touch the dirty main worktree.
- Do not use `git add .`.

## Acceptance Commands

Run from repo root unless noted:

- `go test ./internal/service -run "TestAPIKey.*Concurrency|TestConcurrencyService.*APIKey|TestAccountTestService|TestOpenAI.*Endpoint|Test.*Responses.*Endpoint" -count=1` from `backend`
- `go test ./internal/handler -run "Test.*Endpoint|Test.*Gateway.*Concurrency|Test.*Responses.*Compact" -count=1` from `backend`
- `go test ./internal/server -run "Test.*APIContract" -count=1` from `backend`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/KeysView.spec.ts"`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
- `git diff --check`
- `rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .`
- denied-path audit over `git diff --name-only main..HEAD`

## Output

- One or more scoped commits containing the selected upstream subset.
- Worker result and QA report under `docs/workflow/`.
- Clear final summary of validation and remaining deferred upstream items.

## Stop Rules

- Stop if changes require migrations, Ent regeneration, deploy config, README, payment/welfare product paths, or broad scheduler/Redis cleanup outside the selected scope.
- Stop if conflict resolution would overwrite local product behavior rather than adapt around it.
- Stop if targeted validation exposes unrelated baseline failures that cannot be separated from S54 behavior.
