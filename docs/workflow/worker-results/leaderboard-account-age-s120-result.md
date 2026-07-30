### DONE: leaderboard-account-age-s120

# Worker Result

## Task ID

`leaderboard-account-age-s120`

## Status

`done`

## Summary

- Added a backend account-age authorization check using the existing user
  repository and `created_at`; accounts are eligible at exactly the configured
  number of complete 24-hour periods.
- Applied the check before leaderboard reads and at the leaderboard reward
  claim service entry point.
- Added a shared frontend timestamp predicate, sidebar feature flag, and route
  guard. Added the administrator system setting, public settings/SSR exposure,
  and default/range normalization. All contract success criteria are met.

## Changed Files

- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/usage_handler_leaderboard_test.go`
- `backend/internal/service/usage_leaderboard_access.go`
- `backend/internal/service/usage_leaderboard_access_test.go`
- `backend/internal/service/usage_leaderboard_reward.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/setting_service_update_test.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/setting_handler.go`
- `backend/internal/handler/dto/settings.go`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/components/layout/__tests__/AppSidebar.spec.ts`
- `frontend/src/router/index.ts`
- `frontend/src/router/meta.d.ts`
- `frontend/src/router/__tests__/guards.spec.ts`
- `frontend/src/utils/leaderboardAccess.ts`
- `frontend/src/utils/__tests__/leaderboardAccess.spec.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/stores/app.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/leaderboard-account-age-s120.md`

## Commands Run

```text
go test ./internal/service ./internal/handler -run 'Leaderboard.*Access|DashboardLeaderboard.*AccountAge|ClaimDashboardLeaderboard.*AccountAge' -count=1 -> PASS
go test ./internal/handler -run 'DashboardLeaderboard|ClaimDashboardLeaderboard' -count=1 -> PASS
go test ./internal/service -run 'Leaderboard' -count=1 -> PASS
go test ./internal/service ./internal/handler -run 'Setting.*Leaderboard|PublicSettings' -count=1 -> PASS
go test ./... -run '^$' -> PASS
corepack.cmd pnpm exec vitest run src/utils/__tests__/leaderboardAccess.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/router/__tests__/guards.spec.ts src/views/admin/__tests__/SettingsView.spec.ts -> PASS (4 files, 85 tests)
corepack.cmd pnpm run typecheck -> PASS
corepack.cmd pnpm run build -> PASS (1092 modules)
git diff --check -> PASS
conflict-marker scan -> PASS
unmerged-index scan -> PASS
allowed/denied path audit -> PASS / NO_DENIED_PATHS
```

## Test Output

```text
Go focused and full compile checks passed.
Vitest: 4 files, 85 tests passed.
Frontend typecheck passed; production build transformed 1092 modules.
```

## Risks

- No authenticated browser session, live PostgreSQL-backed request, deployment,
  container refresh, or production runtime smoke was performed.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
