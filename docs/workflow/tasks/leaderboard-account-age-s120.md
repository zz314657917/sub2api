---
task_id: leaderboard-account-age-s120
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Task ID

`leaderboard-account-age-s120`

## Role

Implement the approved S120 access-control slice without expanding leaderboard
product behavior.

## Goal

Allow leaderboard viewing and reward claiming only after an authenticated
account has reached the configured number of complete 24-hour periods since
`created_at`. Hide the frontend entry for younger accounts and reject direct
backend access.

## Success Criteria

- Accounts younger than the effective configured boundary do not see the
  sidebar leaderboard entry.
- Direct `/leaderboard` navigation by an ineligible account redirects to its
  user or administrator dashboard.
- `GET /api/v1/usage/dashboard/leaderboard` and
  `POST /api/v1/usage/dashboard/leaderboard/daily-reward/claim` return HTTP 403
  for ineligible accounts before leaderboard queries or reward settlement.
- An account exactly at the configured boundary is eligible in frontend and
  backend logic.
- The administrator can persist `leaderboard_min_account_age_days`; the
  default is 7, zero is valid, and missing/invalid stored values fall back to
  7.
- The existing public settings response and SSR injection expose the effective
  value to the authenticated frontend.
- Existing eligible-user leaderboard behavior remains unchanged.

## Context

- Repo: `E:/codex-worktrees/sub2api/leaderboard-account-age-s120`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Related files: `backend/internal/handler/usage_handler.go`,
  `backend/internal/service/usage_service.go`, `frontend/src/router/index.ts`,
  `frontend/src/components/layout/AppSidebar.vue`

## Allowed Paths

- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/usage_handler_leaderboard_test.go`
- `backend/internal/service/usage_leaderboard_access.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service_update_test.go`
- `backend/internal/handler/setting_handler.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/dto/public_settings_injection_schema_test.go`
- `backend/internal/service/usage_leaderboard_access_test.go`
- `backend/internal/service/usage_leaderboard_reward.go`
- `backend/internal/handler/admin/setting_handler.go`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`
- `frontend/src/router/index.ts`
- `frontend/src/router/meta.d.ts`
- `frontend/src/router/__tests__/guards.spec.ts`
- `frontend/src/utils/leaderboardAccess.ts`
- `frontend/src/stores/app.ts`
- `frontend/src/types/index.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/utils/__tests__/leaderboardAccess.spec.ts`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/leaderboard-account-age-s120.md`
- `docs/workflow/worker-results/leaderboard-account-age-s120-result.md`
- `docs/workflow/qa-reports/leaderboard-account-age-s120-qa.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/internal/server/routes/**`
- `backend/internal/service/usage_log_repo.go`
- `deploy/**`
- `Dockerfile*`
- `knowledge/**`
- `outputs/**`
- Any path not listed under Allowed Paths.

## Constraints

- Read the effective day count from the persisted setting in the backend and
  public settings in the frontend; use 7 as the fallback in both runtimes.
  Compare absolute timestamps and do not use calendar-week or timezone-
  dependent boundaries.
- Missing, zero, or invalid `created_at` values fail closed.
- Backend authorization is authoritative; frontend hiding is an additional UX
  gate, not the security boundary.
- Reuse the existing `UsageService` user repository and existing response
  error mapping. Do not add constructor dependencies or database queries to
  ranking repositories.
- Do not change leaderboard participation, output DTOs, ranking, cache TTL, or
  reward amounts.
- Work only in this isolated worktree. Do not push, deploy, or update containers.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/leaderboard-account-age-s120/backend
  go test ./internal/service ./internal/handler -run "Leaderboard.*Access|DashboardLeaderboard.*AccountAge|ClaimDashboardLeaderboard.*AccountAge|Setting.*Leaderboard" -count=1
go test ./internal/handler -run "DashboardLeaderboard|ClaimDashboardLeaderboard" -count=1
gofmt -d internal/service/usage_leaderboard_access.go internal/service/usage_leaderboard_access_test.go internal/handler/usage_handler.go internal/handler/usage_handler_leaderboard_test.go
cd E:/codex-worktrees/sub2api/leaderboard-account-age-s120
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/utils/__tests__/leaderboardAccess.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/router/__tests__/guards.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
git diff --name-only HEAD
git diff --name-only --diff-filter=U
```

## Output

- Narrow backend authorization gate, shared frontend eligibility predicate,
  sidebar and route guards, focused regressions, Generator result, and QA report.

## Stop Rules

- Stop if the feature requires schema/migration changes, a new independent
  public settings API, ranking repository changes, or reward calculation
  changes. Extending the existing public settings payload is in scope.
- Stop if `created_at` is not available in the authenticated frontend user or
  existing backend user repository.
- Stop if implementation requires touching any denied or unlisted path.
