# Task Contract: usage-user-frontend-s136

## Task ID

`usage-user-frontend-s136`

## Status

`approved`

## Role

Codex owns Planner, implementation, QA execution, and final evaluation for this isolated frontend sprint.

## Goal

Complete the authenticated user's `/usage` experience on top of S135: filtered summary cards, token trend and model/group/endpoint distributions, plus an opt-in redacted failed-request list/detail workflow and the administrator setting needed to enable it.

## Source References

- Local baseline: `codex/usage-full-alignment-s135@d48370f75`.
- Upstream behavior references: `cafc95c3e`, `cfb195c7b`, and `ebbdc7031`.
- Do not cherry-pick these commits wholesale and do not replace locally evolved Usage files with upstream snapshots.

## Success Criteria

- The user Usage page shows summary cards, trend, requested-model, group, and endpoint analytics above the existing detail table.
- Date, API key, group, model, request type, and billing-mode filters are propagated consistently to table, stats, trend, model, group, and endpoint data.
- Day/hour granularity and distribution metric/source controls update charts without changing the existing user table semantics.
- The existing subscriptions panel, usage detail drawer, column persistence, CSV export, image usage, reasoning-effort, billing-mode, and pagination behavior remain intact.
- When `allow_user_view_error_requests` is true, the page exposes an Error Requests tab with user-safe date/API-key/model/category/status filters, stable sorting, pagination, configurable columns, and a redacted detail modal.
- When the setting is false or unavailable, no user error tab or request is exposed. The admin Settings page can persist the opt-in switch.
- User error UI types are a strict whitelist and contain no IP, User-Agent, email, account, upstream endpoint, retry, owner/source, or API-key-prefix fields.

## Allowed Paths

- `frontend/src/api/usage.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/stores/app.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `frontend/src/components/user/UserErrorRequestsTable.vue`
- `frontend/src/components/user/UserErrorDetailModal.vue`
- `frontend/src/components/user/__tests__/UserErrorRequestsTable.spec.ts`
- `frontend/src/components/user/__tests__/UserErrorDetailModal.spec.ts`
- `frontend/src/components/admin/usage/UsageStatsCards.vue`
- `frontend/src/components/admin/usage/__tests__/UsageStatsCards.spec.ts`
- `frontend/src/components/charts/EndpointDistributionChart.vue`
- `frontend/src/components/charts/GroupDistributionChart.vue`
- `frontend/src/components/charts/ModelDistributionChart.vue`
- `frontend/src/components/charts/TokenUsageTrend.vue`
- `frontend/src/components/charts/__tests__/*DistributionChart.spec.ts`
- `frontend/src/utils/errorBadges.ts`
- `frontend/src/utils/errorCategory.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.userErrorRequests.spec.ts`
- `docs/workflow/tasks/usage-user-frontend-s136.md`
- `docs/workflow/qa-reports/usage-user-frontend-s136-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- All backend, schema, migration, generated, deployment, container, and production configuration files.
- `frontend/src/views/admin/UsageView.vue`, `frontend/src/views/admin/ops/**`, administrator error-log APIs, Token ranking, and leaderboard behavior; these belong to S137.
- Primary-worktree user changes, `outputs/**`, global memories, and unrelated knowledge files.

## Constraints

- Adapt to the local modular i18n, settings form, chart components, and evolved Usage detail table.
- Prefer existing components and types; do not duplicate chart or table infrastructure.
- Preserve S135 user ownership and redaction boundaries. Never render or add sensitive admin-only error fields to user types.
- Error UI must be opt-in and fail closed. Do not fetch error records before the user opens the enabled tab.
- Keep all work in `E:/codex-worktrees/sub2api/usage-user-frontend-s136`.
- Do not merge into the dirty primary worktree, push, deploy, update containers, or run a production migration.

## Acceptance Commands

```powershell
corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts src/components/user/__tests__/UserErrorRequestsTable.spec.ts src/components/user/__tests__/UserErrorDetailModal.spec.ts src/components/admin/usage/__tests__/UsageStatsCards.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run build
git diff --check
```

Focused tests must prove analytics filter propagation, enabled/disabled error visibility, lazy error loading, error filtering/sorting/pagination, redacted details, and settings round-trip. A browser smoke should inspect desktop and mobile layout if the local frontend can be started without changing deployment state.

## Output

- S136 frontend implementation in the isolated branch.
- `docs/workflow/qa-reports/usage-user-frontend-s136-qa.md` with a first-line PASS, FAIL, or BLOCKED verdict.
- Updated workflow status, main log, and current task handoff.
- No push, deployment, container update, production migration, or primary-worktree merge.

## Stop Rules

- Stop if implementation requires backend behavior beyond S135, an administrator error-log rewrite, Token-ranking backend work, schema changes, or broad settings architecture changes.
- Stop on any user error field that exceeds the S135 whitelist, any unowned API-key filter path, or any automatic default-on behavior.
- Stop if a locally evolved Usage behavior would be lost by upstream adaptation, or if changed paths exceed this contract.
