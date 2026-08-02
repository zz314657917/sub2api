# Task Contract: usage-full-alignment-s135

## Task ID

`usage-full-alignment-s135`

## Status

`approved`

## Role

Codex owns the Planner, implementation, QA execution, and final evaluator gates for this bounded backend foundation sprint.

## Goal

Provide the backend contracts required by the full Usage-page alignment without changing frontend files: user-scoped usage analytics, user-visible failed-request list/detail APIs, the opt-in visibility setting, and the PostgreSQL index needed for user error pagination.

## Source References

- Local baseline: `main@1c1021133`
- Upstream reference: `upstream/main@7ceabb3fd`
- Behavior references: `cafc95c3e`, `cfb195c7b`, `fe8952733`, `ebbdc7031`
- Do not cherry-pick these commits wholesale and do not merge `upstream/main`.

## Success Criteria

- User usage stats, trend, model, group, and snapshot endpoints accept the existing user-safe filters with API-key ownership enforced.
- `GET /api/v1/usage/errors` and `GET /api/v1/usage/errors/:id` return only the authenticated user's redacted error requests.
- User error listing forces `user_id`, includes business-limited failures, excludes `count_tokens` noise, supports category/status/model/date/API-key filters, and uses a stable sort whitelist.
- Cross-user detail access returns not-found semantics; sensitive account, upstream, email, and retry fields never appear in the user DTO.
- `allow_user_view_error_requests` is present in admin/public settings, defaults to false, and fails closed when the setting store is unavailable.
- A new non-transactional migration creates an idempotent concurrent partial index on `ops_error_logs(user_id, created_at DESC)` without reusing migration 148.
- Existing administrator error-log queries, usage semantics, billing data, and route ordering remain unchanged.

## Allowed Paths

- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/usage_handler_*_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/setting_handler.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/repository/ops_repo.go`
- `backend/internal/repository/ops_repo_error_where_test.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_*_test.go`
- `backend/internal/service/ops_models.go`
- `backend/internal/service/ops_service.go`
- `backend/internal/service/ops_user_error.go`
- `backend/internal/service/ops_user_error_test.go`
- `backend/internal/service/ops_service_user_error_test.go`
- `backend/internal/service/setting_public.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/usage_service.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/server/routes/user.go`
- `backend/internal/server/routes/user_test.go`
- `backend/internal/server/api_contract_test.go`
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/migrations/200_add_ops_error_logs_user_time_index_notx.sql`
- `docs/workflow/tasks/usage-full-alignment-s135.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- All `frontend/**` files; those belong to S136/S137.
- Existing migration files, schema rewrites, billing, deployment, containers, and production configuration.
- `outputs/**` and any primary-worktree-only user changes.
- Global memories and unrelated knowledge files.

## Constraints

- Adapt against the local service/repository interfaces; do not copy upstream files over local implementations.
- Preserve current admin error-log filtering, recovery/error classification, and usage billing semantics.
- Public user errors are a whitelist view. Never expose account IDs/names, upstream endpoint, API-key prefix, user email, internal owner/source, or retry controls.
- The user service layer must overwrite caller-controlled ownership fields with the authenticated user ID and return 404 for non-owned details.
- The setting is opt-in and fail closed. No default-on behavior.
- Use migration number 200 because 148 is already occupied locally.
- Keep the worktree isolated and do not push, deploy, refresh containers, or modify the primary checkout.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service ./internal/handler ./internal/repository ./internal/server/routes
go test -run '^$' ./...
gofmt -w <changed-go-files>
Pop-Location
git diff --check
```

Focused tests must cover user-error ownership/redaction, filter construction, setting fail-closed behavior, route registration, snapshot filter propagation, and migration validation. Any pre-existing unrelated test drift must be recorded rather than broadened into this sprint.

## Output

- Implemented S135 backend changes in this isolated worktree.
- `docs/workflow/qa-reports/usage-full-alignment-s135-qa.md` with first line `### PASS`, `### FAIL`, or `### BLOCKED`.
- Updated workflow status, main log, and `knowledge/tasks/current-task.md`.
- No commit, push, deployment, or primary-worktree merge in this sprint unless separately authorized.

## Stop Rules

- Stop if the local interfaces require a broad upstream merge, schema rewrite, billing/routing change, or frontend edit.
- Stop if migration 200 conflicts or the existing `ops_error_logs` schema is not compatible with the planned index.
- Stop on any ownership/redaction test failure, out-of-scope path change, or unresolved generated Wire mismatch.
- Stop before integration if the primary worktree remains dirty in overlapping workflow or source paths.
