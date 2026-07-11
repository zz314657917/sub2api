### DONE: leaderboard-participation-exclusion-s70

## Changed Files

- Added `users.exclude_from_leaderboard` migration, Ent schema, generated ORM support, and service entity mapping.
- Extended the existing admin user update path and edit dialog with an explicit participation checkbox.
- Applied the exclusion before raw and daily-stats leaderboard ranks, prior-period rank movement, daily champions, badges, model ranking, and leaderboard trend aggregation.
- Added targeted cache invalidation after an administrator changes the flag.
- Added service, frontend, SQL mock, and real database integration coverage.

## Commands Run

- `go generate ./ent`
- `go generate ./cmd/server`
- `go test ./internal/repository -run "TestUsageLogRepositoryGetUserLeaderboard.*|TestUsageLogRepositoryGetLeaderboardDailyChampions|TestUsageLogRepositoryGetUserLeaderboardBadgeLeaders" -count=1`
- `go test ./internal/service -run "TestAdminServiceUpdateUserLeaderboardExclusion|TestUsageServiceInvalidateLeaderboardCaches|Test.*Leaderboard.*Reward" -count=1`
- `go test ./cmd/server -run "^$" -count=1`
- `corepack.cmd pnpm --dir frontend exec vitest run src/components/admin/user/__tests__/UserEditModal.spec.ts`
- `corepack.cmd pnpm --dir frontend run typecheck`
- `go test -tags=integration ./internal/repository -run "TestUsageLogRepoSuite/TestGetUserLeaderboardExcludesMarkedUsers" -count=1`
- `git diff --check`

## Scope Check

- No payment, billing, gateway, API key, account, deployment, or generic usage reporting behavior was changed.
- Existing leaderboard reward settlement code was not modified; it uses the filtered leaderboard result and therefore does not create new eligibility for excluded users.
