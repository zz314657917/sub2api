# Task Contract: usage-admin-frontend-s137

## Task ID

`usage-admin-frontend-s137`

## Status

`approved`

## Role

Codex owns Planner, implementation, QA execution, and final evaluation for this isolated administrator Usage sprint.

## Goal

Complete the administrator `/admin/usage` surface on top of S135/S136: a three-tab detail workbench for usage records, failed requests, and per-user Token ranking, with the smallest backend contract additions needed for real filtering and ranking semantics.

## Source References

- Local baseline: `usage-user-frontend-s136@4f4d61008`.
- Upstream behavior references: `cfb195c7b`, `ebbdc7031`, `b062b3664`, and `1a3cc2a78`.
- Do not cherry-pick these commits wholesale or replace locally evolved Usage/Ops files with upstream snapshots.

## Success Criteria

- The administrator Usage page exposes three tabs: usage details, error requests, and user Token ranking.
- Existing summary cards, charts, filters, export, cleanup, usage table, route hydration, column persistence, and balance-history drill-down remain intact.
- Error requests load only after their tab is opened, reuse the administrator Ops list/detail boundary, and support shared date/user/API key/account/model/group filters plus error type/category/status, stable server-side sorting, pagination, and independent column persistence.
- The administrator error-list HTTP handler accepts and validates the existing repository filter surface for user ID, API key ID, model, category, `sort_by`, and `sort_order`; no schema, route, permission, or repository architecture change is introduced.
- The user ranking loads only after its tab is opened, reuses the shared date/user/API key/account/model/group/request/billing filters supported by `user-breakdown`, and supports a bounded Top 20/50/100/200 selector.
- `user-breakdown` returns input, output, cache, and total Token aggregates and applies a strict server-side sort whitelist for requests, Token fields, and actual cost.
- Clicking a ranking row scopes the Usage page to that user, hydrates the user label, returns to usage details, and refreshes the shared analytics and table.
- Error and ranking surfaces remain usable at desktop and 390px mobile widths without page-level horizontal overflow.

## Allowed Paths

- `backend/internal/handler/admin/dashboard_handler.go`
- `backend/internal/handler/admin/dashboard_handler_user_breakdown_test.go`
- `backend/internal/handler/admin/ops_handler.go`
- `backend/internal/handler/admin/ops_handler_error_filters_test.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_breakdown_test.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `frontend/src/api/admin/dashboard.ts`
- `frontend/src/api/admin/ops.ts`
- `frontend/src/types/index.ts`
- `frontend/src/utils/errorBadges.ts`
- `frontend/src/components/admin/usage/UsageFilters.vue`
- `frontend/src/components/admin/usage/UserTokenRanking.vue`
- `frontend/src/components/admin/usage/__tests__/UsageFilters.spec.ts`
- `frontend/src/components/admin/usage/__tests__/UserTokenRanking.spec.ts`
- `frontend/src/views/admin/UsageView.vue`
- `frontend/src/views/admin/__tests__/UsageView.spec.ts`
- `frontend/src/views/admin/ops/components/OpsErrorLogTable.vue`
- `frontend/src/views/admin/ops/components/__tests__/OpsErrorLogTable.spec.ts`
- `frontend/src/i18n/locales/en/admin/usage.ts`
- `frontend/src/i18n/locales/zh/admin/usage.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `docs/workflow/tasks/usage-admin-frontend-s137.md`
- `docs/workflow/qa-reports/usage-admin-frontend-s137-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Schema, migrations, generated code, routes, authentication/authorization, billing, deployment, containers, and production configuration.
- User Usage behavior, Settings behavior, Ops dashboard orchestration, retry/resolution semantics, and unrelated administrator pages.
- Primary-worktree user changes, `outputs/**`, global memories, and unrelated knowledge files.

## Constraints

- Adapt to local modular i18n, evolved Usage filters/charts, current Ops detail modal, and responsive `DataTable` behavior.
- Backend work is limited to DTO aggregates, query ordering, and existing HTTP filter plumbing. Do not add endpoints or alter data ownership.
- Error-list sort keys are restricted to `created_at`, `model`, and `status_code`. Ranking sort keys are restricted to `requests`, `input_tokens`, `output_tokens`, `cache_tokens`, `total_tokens`, and `actual_cost`.
- Error and ranking requests must be lazy. Stale responses must not overwrite newer tab/filter requests.
- Preserve independent usage/error column preferences and keep required identity/status/time/action columns visible.
- Keep all work in `E:/codex-worktrees/sub2api/usage-admin-frontend-s137`.
- Do not merge into the dirty primary worktree, push, deploy, update containers, or run a production migration.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler/admin -run 'Test(GetUserBreakdown|OpsErrorHandler)' -count=1
go test ./internal/repository -run 'TestGetUserBreakdownStats' -count=1
Pop-Location
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/UsageView.spec.ts src/components/admin/usage/__tests__/UsageFilters.spec.ts src/components/admin/usage/__tests__/UserTokenRanking.spec.ts src/views/admin/ops/components/__tests__/OpsErrorLogTable.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run build
git diff --check
```

Focused tests must prove handler filter/sort validation, Token aggregation/sort SQL, tab lazy loading, shared filter propagation, error sorting/pagination/columns/detail, ranking sorting/limit/drill-down, stale-response protection, and responsive rendering. A Playwright mock smoke must inspect desktop and 390px mobile layouts.

## Output

- S137 implementation in the isolated branch.
- `docs/workflow/qa-reports/usage-admin-frontend-s137-qa.md` with a first-line PASS, FAIL, or BLOCKED verdict.
- Updated workflow status, main log, and current task handoff.
- One authorized local S137 commit only; no push, deployment, container update, production migration, or primary-worktree merge.

## Stop Rules

- Stop if implementation requires schema/migration changes, new routes, permission changes, Ops retry/resolution changes, or a broad dashboard/filter rewrite.
- Stop if sort values are interpolated without a strict whitelist, if error requests load before tab activation, or if ranking/error filters silently claim unsupported backend semantics.
- Stop if current Usage analytics, export/cleanup, route hydration, user redaction, or primary-worktree changes would be lost.
- Stop on changed paths outside this contract or on unresolved desktop/mobile page overflow.
