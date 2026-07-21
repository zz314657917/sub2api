# Default Key Fallback Group S93

## Task ID

`default-key-fallback-group-s93`

## Role

Planner / Generator / Final Evaluator: Codex. No external worker is required;
implementation and QA remain in the primary checkout because the target API-key
and settings files already contain the approved, uncommitted S89-S92 work.

## Goal

Add an administrator-selected fallback group for system-created default API
keys. Persist it inside `studio_bridge_luoye_ai` as
`default_fallback_group`; write it to the new key's base `group_id` while
keeping purpose-specific `default_api_routes` as higher-priority routes; and
provide an explicit administrator action that fills only existing users'
ungrouped default keys without changing their routes or already selected
groups.

## Success Criteria

1. System settings, DTO mappings, and the admin settings UI round-trip
   `default_fallback_group` without a database migration.
2. A system-created default key uses the configured, existing fallback group
   as its base `GroupID`; invalid or missing configuration remains ungrouped.
3. `SkipGroupPermissionCheck` applies consistently to the base group and
   multi-group routes for system-created keys, while ordinary user-created keys
   retain existing permission checks.
4. Purpose-specific multi-group routes remain first choice. Final fallback
   still checks active group context, platform, routing scope, and S91
   `Group.MatchesModel`; incompatible fallback returns the existing stable
   no-route behavior.
5. The explicit backfill updates only the lowest-ID non-deleted key per user when
   that key has `group_id IS NULL`; it preserves routes, pool strategy, quota,
   status, and every already grouped key, and invalidates changed auth-cache
   entries.
6. The settings UI clearly says the fallback applies to all newly registered
   users, exposes an active-group selector, asks for confirmation before
   backfill, and reports the updated count.
7. Focused Go tests, SettingsView/API Vitest, typecheck, production build, and
   diff/conflict gates pass.

## Allowed Paths

- `backend/internal/service/studio_bridge.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/setting_service_studio_bridge_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_initial_test.go`
- `backend/internal/service/api_key_service_fallback_s93_test.go`
- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_routing_s88_test.go`
- `backend/internal/service/api_key_routing_s93_test.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/api_key_repo_integration_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/api_key_fallback_s93_test.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/default-key-fallback-group-s93.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Ent schema/generated code, SQL migrations, billing, subscriptions, payment,
  account scheduling, provider adapters, and channel restriction semantics
- automatic backfill during settings save or application startup
- changing any existing grouped key or any non-default key
- replacing purpose routes with a low-priority ordinary route
- Dockerfiles, compose files, deployment, container/image updates, commits,
  pushes, or history rewrites
- unrelated S89-S91 source, tests, or workflow evidence

## Constraints

- Preserve all current uncommitted work and edit only the relevant hunks.
- The settings field is optional for backwards compatibility and uses the
  existing Studio Bridge JSON storage.
- A missing/deleted fallback group must fail the explicit backfill safely and
  must not prevent creation of an otherwise valid default key.
- Repository backfill selection must match `findDefaultAPIKey`: the lowest
  non-deleted API-key ID per user.
- Auth-cache invalidation must use the concrete changed key values returned by
  the repository operation.

## Acceptance Commands

```powershell
go test ./internal/service -run "APIKeyRoutingS88|APIKeyRoutingS91|APIKeyRoutingS93|DefaultKeyFallback|EnsureInitialKey"
go test ./internal/handler -run "DefaultKeyFallback"
go test ./internal/server/routes -run "^$"
go test -tags=integration ./internal/repository -run "APIKeyRepoSuite/TestBackfillDefaultKeyFallback"
npm.cmd run test -- --run src/views/admin/__tests__/SettingsView.spec.ts
npm.cmd run typecheck
npm.cmd run build
git diff --check
git diff --name-only --diff-filter=U
```

The broader `go test -tags=unit ./internal/service ...` command is a recorded
baseline diagnostic rather than an S93 gate: unrelated unit-tag tests currently
fail package compilation on duplicate `stringPtr`, removed billing fields/old
signatures, and removed Grok runtime-block helpers. S93 therefore includes
default-tag focused tests that compile and execute independently.

## Output

- Source changes and focused tests within the allowed paths.
- Updated workflow status/spec/main log and current-task snapshot.
- Final review in the required order: Findings, Executed Checks,
  Unverified Risks, Recommendation.

## Stop Rules

- Stop if identifying the default key cannot remain consistent with current
  lowest-ID behavior.
- Stop if the backfill would require a schema migration or would overwrite
  route JSON or existing grouped keys.
- Any required acceptance failure is `FAIL` until fixed and rerun; do not claim PASS
  from source inspection alone.

## Evaluator Review

### PASS: default-key-fallback-group-s93 contract

The contract matches the requested behavior, keeps fallback separate from
purpose routing, preserves S91 model checks, makes existing-data changes
explicit, and constrains implementation to the already dirty primary checkout.
