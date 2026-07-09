### DONE: leaderboard-reward-modes-s57

## Changed files
- `frontend/src/views/user/LeaderboardView.vue`
- `frontend/src/views/user/__tests__/LeaderboardView.spec.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/api/usage.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/zh/leaderboard.ts`
- `frontend/src/i18n/locales/en/leaderboard.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/__tests__/leaderboard-theme.spec.ts`
- `docs/workflow/worker-results/leaderboard-reward-modes-s57-frontend.md`

## Commands run
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend install --frozen-lockfile"`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/LeaderboardView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts"`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/leaderboard-theme.spec.ts"`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
- `git diff --check -- frontend/src/views/user/LeaderboardView.vue frontend/src/views/user/__tests__/LeaderboardView.spec.ts frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts frontend/src/api/usage.ts frontend/src/api/admin/settings.ts frontend/src/types/index.ts frontend/src/i18n/locales/zh/leaderboard.ts frontend/src/i18n/locales/en/leaderboard.ts frontend/src/i18n/locales/zh/admin/settings.ts frontend/src/i18n/locales/en/admin/settings.ts frontend/src/__tests__/leaderboard-theme.spec.ts`

## Notes
- Frontend settings now uses the contract fields `reward_mode`, `red_packet_pool_amount`, `red_packet_min_amount`, `red_packet_max_amount`, `lottery_amount`, and `lottery_cron`.
- User leaderboard reward panel always renders last week top 10 token usage when reward data is present.
- Disabled mode has no claim controls; red packet mode shows pending/claimed amount and claim action; lottery mode shows draw time and result without a claim button.
- Legacy `leaderboard_daily_reward_enabled=true` maps to `red_packet` when `reward_mode` is absent.

## Risks
- Backend API was not modified by this frontend worker. If backend response naming diverges from the contract fields, the UI will fall back only for legacy enabled/disabled mode, not for alternate new-field aliases.
- Vitest emits existing `router-link` stub warnings in `SettingsView.spec.ts`, and Browserslist data is stale; neither failed the executed checks.
