# Task Contract: expired-subscription-route-skip-s99

- Task ID: `expired-subscription-route-skip-s99`
- Role: Planner / Generator / Evaluator
- Goal: Keep API-key multi-group routing usable when one configured subscription group is no longer eligible by skipping that route and continuing to the next priority without deleting the saved route.
- Success Criteria:
  - Multi-group requests skip subscription routes whose subscription is missing, expired, or suspended, then continue through the existing priority and model-match selection.
  - Daily, weekly, or monthly usage-limit errors and transient persistence errors do not trigger route failover; the selected route keeps the existing error behavior.
  - A configured default group can be used only when its current group and subscription eligibility remain valid; when every configured route is unusable, model-aware and initial routing return `NO_MATCHING_GROUP_ROUTE`.
  - `/v1/models` and `/v1/model-catalog` exclude request-scoped unavailable subscription routes.
  - Existing unavailable route IDs and an unchanged existing base group can be preserved during Key edits, while create and newly added group bindings still require current permission.
  - The Key editor shows existing unavailable routes with their saved group identity when available, labels them as expired/unavailable, keeps them non-selectable, and still allows removal.
  - Single-group Key behavior remains unchanged, and renewing the subscription makes the saved route eligible again without data cleanup.
- Allowed Paths:
  - `backend/internal/service/api_key.go`
  - `backend/internal/service/api_key_routing.go`
  - `backend/internal/service/api_key_service.go`
  - `backend/internal/service/*api_key*test.go`
  - `backend/internal/server/middleware/api_key_auth.go`
  - `backend/internal/server/middleware/*api_key_auth*test.go`
  - `backend/internal/handler/gateway_handler.go`
  - `backend/internal/handler/gateway_models_test.go`
  - `frontend/src/views/user/KeysView.vue`
  - `frontend/src/views/user/__tests__/KeysView.spec.ts`
  - `frontend/src/views/user/__tests__/KeysView.createQuery.spec.ts`
  - `frontend/src/i18n/locales/en/keys.ts`
  - `frontend/src/i18n/locales/zh/keys.ts`
  - `docs/workflow/tasks/expired-subscription-route-skip-s99.md`
  - `docs/workflow/qa-reports/expired-subscription-route-skip-s99-qa.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: Ent schema, migrations, repositories, billing calculation, usage accounting, account scheduling, group administration, deployment, containers, unrelated frontend views, commits, and pushes.
- Constraints:
  - Treat only `ErrSubscriptionNotFound`, `ErrSubscriptionExpired`, and `ErrSubscriptionSuspended` as route-skippable subscription failures.
  - Store unavailable group IDs only on a request-local API Key copy; never mutate an authentication-cache object or shared group snapshot.
  - Keep route configuration persisted so renewal automatically restores eligibility.
  - Preserve existing route priority, group model matching, channel `restrict_models`, and cooldown semantics.
  - Do not add authorization bypasses for group IDs that are not already bound to the Key being updated.
- Acceptance Commands:
  - Focused Go tests for route eligibility, middleware fallback/error behavior, update preservation, and gateway model catalogs.
  - Focused KeysView Vitest files.
  - Frontend typecheck and production build.
  - `gofmt` on changed Go files.
  - `git diff --check`, conflict-marker scan, allowed-path audit, and unmerged-index check.
- Output: Scoped source diff, focused regression tests, S99 QA report, and final `PASS`, `FAIL`, or `BLOCKED` evidence.
- Stop Rules: Stop on a required schema/repository change, inability to distinguish usage limits from eligibility failures, mutation of cached API Key objects, or unrelated dirty-file conflicts.

## Contract Review

`PASS`: The contract isolates temporary subscription ineligibility from billing-limit failures, keeps renewal reversible, covers model-list parity and editor recovery, and explicitly protects cached authentication state. The allowed paths are sufficient without migration, billing, deployment, or container changes.
