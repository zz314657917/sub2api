### DONE: upstream-main-runtime-safety-s8

## Summary

- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-runtime-safety-s8` and branch `codex/upstream-main-runtime-safety-s8` from baseline `fed704641`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-runtime-safety-s8.md`.
- Ported all five approved runtime safety candidates without directly merging `upstream/main`.
- Kept the implementation inside the approved backend and workflow paths.
- Adapted the upstream content moderation auto-ban tests to the local `applyFlaggedSideEffects` helper shape while preserving the same behavior coverage.

## Candidate Results

- `1b6a15b48`: `CHERRY_PICKED` as `d7a07723a`. Adds DB pool lifetime/idle-time clamping and tests.
- `c40a74d98`: `CHERRY_PICKED_WITH_LOCAL_TEST_RESOLUTION` as `db145ea56`. Adds admin auto-ban exemption; adapted tests to local helper names and test repo behavior.
- `3571b082f`: `CHERRY_PICKED` as `a2212dc76`. Adds shared OpenAI-compatible `stream` boolean validation and handler tests.
- `b6c0706e3`: `CHERRY_PICKED` as `dc38e0213`. Syncs scheduler snapshots after overloaded/temp-unschedulable/model-rate-limit clears and adds repository coverage.
- `7513b7ea6`: `CHERRY_PICKED` as `8ff9810e0`. Extracts OpenAI HTTP response IDs and binds them to selected accounts for sticky previous-response routing.

## Deferred / Skipped

- `0a521f09f`: `APPLIED_EQUIVALENT`. Current baseline already closes an open Gemini `tool_use` block before text in messages streaming.
- `362f9e77b`: `DEFERRED`. Leader lock touches broader server/repository/service wiring and should be a dedicated Sprint.
- `69b465451`: `DEFERRED`. Requires migration, which is forbidden in this Sprint.
- `aea2950b1`: `DEFERRED`. Includes frontend changes, forbidden in this Sprint.
- `650981f2e`: `SKIPPED`. lint/gofmt-only and tied to upstream context outside this Sprint.

## Commits

- `fc5736ae2` docs: add runtime safety s8 contract
- `d7a07723a` fix(db-pool): enforce connection lifetime floors
- `db145ea56` fix(risk-control): exempt admins from moderation auto-ban
- `a2212dc76` fix: validate stream field type
- `dc38e0213` fix: sync scheduler snapshots after account state clears
- `8ff9810e0` Bind OpenAI HTTP response IDs to selected accounts

## Changed Files

- `backend/internal/repository/db_pool.go`
- `backend/internal/repository/db_pool_test.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `backend/internal/service/content_moderation.go`
- `backend/internal/service/content_moderation_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_stream_validation.go`
- `backend/internal/handler/openai_stream_validation_test.go`
- `docs/workflow/tasks/upstream-main-runtime-safety-s8.md`
- `docs/workflow/worker-results/upstream-main-runtime-safety-s8-result.md`
- `docs/workflow/qa-reports/upstream-main-runtime-safety-s8-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check`
- denied path audit against `main..HEAD`
- `go test ./internal/repository -run "DBPool|Pool|Connection|Lifetime|SetOverloaded|TempUnschedulable|ClearModelRateLimits|Scheduler" -count=1`
- `go test ./internal/service -run "ContentModeration|AutoBan|Admin|OpenAI|ResponseID|BindHTTP" -count=1`
- `go test ./internal/handler -run "Stream|OpenAI|Gateway|ChatCompletions|Responses" -count=1`
- `go test ./internal/repository ./internal/service ./internal/handler -count=1`

## Notes

- No candidate required forbidden paths, Ent schema, SQL migration, frontend changes, public API fields, or route/server wiring.
- The content moderation test repo now counts flagged logs by user/window, which matches the service behavior needed by the new auto-ban threshold tests.
