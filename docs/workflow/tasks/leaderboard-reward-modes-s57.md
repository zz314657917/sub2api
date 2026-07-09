# Task Contract

## Task ID
leaderboard-reward-modes-s57

## Role
Generator workers implement a scoped product change under Codex review. Backend and frontend workers must keep write ownership disjoint and must not revert unrelated edits.

## Goal
Replace the legacy top-3 weekly leaderboard reward with configurable leaderboard reward modes: disabled, red packet, and lottery. The user-facing leaderboard side panel must always show last week's top 10 token usage; reward UI only appears when a reward mode is enabled.

## Success Criteria
- Admin settings expose a three-state reward mode: `disabled`, `red_packet`, `lottery`.
- Disabled mode returns and renders last week's top 10 token usage only.
- Red packet mode pre-splits 10 packet amounts for a reward period, allows each eligible top-10 user to claim one random unclaimed packet, and prevents duplicate claims.
- Lottery mode draws exactly one winner from last week's top 10 after the configured cron time, grants the configured amount once, and is idempotent under repeated scheduler/lazy settlement attempts.
- Existing balance grant and `redeem_codes` audit semantics are preserved for paid-out rewards.
- The old top-3 reward cards and copy are removed from the user-visible leaderboard panel.

## Context
- Repo: `E:/codex-worktrees/sub2api/leaderboard-reward-modes`
- Existing backend entrypoints:
  - `backend/internal/service/usage_leaderboard_reward.go`
  - `backend/internal/repository/usage_log_repo.go`
  - `backend/internal/pkg/usagestats/usage_log_types.go`
  - `backend/internal/service/setting_service.go`
  - `backend/internal/handler/admin/setting_handler.go`
  - `backend/internal/handler/usage_handler.go`
- Existing frontend entrypoints:
  - `frontend/src/views/user/LeaderboardView.vue`
  - `frontend/src/views/admin/SettingsView.vue`
  - `frontend/src/types/index.ts`
  - `frontend/src/api/usage.ts`
  - `frontend/src/api/admin/settings.ts`
  - `frontend/src/i18n/locales/{zh,en}/leaderboard.ts`
  - relevant admin settings locale files and tests.

## Allowed Paths
- `backend/migrations/186_leaderboard_reward_modes.sql`
- `backend/internal/service/**`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/usage_handler_leaderboard_test.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/server/api_contract_test.go`
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `frontend/src/views/user/LeaderboardView.vue`
- `frontend/src/views/user/__tests__/LeaderboardView.spec.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/api/usage.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/{zh,en}/leaderboard.ts`
- `frontend/src/i18n/locales/{zh,en}/admin/settings.ts`
- `frontend/src/__tests__/leaderboard-theme.spec.ts`
- `docs/workflow/tasks/leaderboard-reward-modes-s57.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/leaderboard-reward-modes-s57-*.md`
- `docs/workflow/qa-reports/leaderboard-reward-modes-s57-qa.md`

## Denied Paths
- `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- Existing dirty files in the primary checkout unless explicitly listed in Allowed Paths.
- Payment, Studio Bridge, channel status, monitor, console theme, Docker/deploy, production config, unrelated Ent schema/codegen.

## Constraints
- Keep the implementation compatible with legacy settings: if the new mode is absent, `leaderboard_daily_reward_enabled=true` should map to an enabled reward mode; otherwise default to disabled.
- Use service timezone for reward windows and lottery cron evaluation.
- Use database uniqueness/transactions for duplicate protection; do not rely only on in-memory state.
- Do not introduce a new Ent schema/codegen path for this task; use raw SQL repository methods consistent with existing leaderboard reward claims.
- Keep frontend layout compact and responsive; no nested cards inside cards.

## Acceptance Commands
```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
go test ./internal/service -run "Test.*Leaderboard.*Reward|Test.*RedPacket|Test.*Lottery" -count=1
go test ./internal/repository -run "Test.*Leaderboard.*Reward|Test.*Leaderboard.*Packet|Test.*Leaderboard.*Lottery" -count=1
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/LeaderboardView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
git diff --check
```

## Output
- Backend worker report: `docs/workflow/worker-results/leaderboard-reward-modes-s57-backend.md`
- Frontend worker report: `docs/workflow/worker-results/leaderboard-reward-modes-s57-frontend.md`
- QA report: `docs/workflow/qa-reports/leaderboard-reward-modes-s57-qa.md`
- Report first line must be `### DONE: leaderboard-reward-modes-s57`, `### BLOCKED: leaderboard-reward-modes-s57`, or `### FAILED: leaderboard-reward-modes-s57`.

## Stop Rules
- Stop if a change requires Ent code generation, production config changes, or touching denied product areas.
- Stop if backend and frontend contracts disagree on API shape.
- Stop if the reward mode cannot be made idempotent without a schema addition.
