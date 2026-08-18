### DONE: upstream-cn-group-entry-s228

## Changed Files

- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/admin/group_handler_platform_test.go`
- `frontend/src/types/index.ts`
- `frontend/src/components/common/GroupBadge.vue`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/ChannelsView.vue`
- `frontend/src/i18n/locales/en/admin/groups.ts`
- `frontend/src/i18n/locales/zh/admin/groups.ts`
- `frontend/src/components/account/CNProviderBalanceCell.vue`
- `frontend/src/components/account/__tests__/CNProviderBalanceCell.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`

Business commits:

- `df43f3876 feat(groups): allow CN provider groups`
- `26a5dec9d fix(accounts): localize CN balance and expired state`

## Manual Adaptation

- `7cdca9e49`: local group handler and GroupsView owners receive Kimi/Zhipu/DeepSeek support. The local checkout did not have the upstream CompositeRouteRequest owner, so no composite route surface was created or widened.
- `c38c5beef`: empty CN balance state now uses `admin.accounts.cnProviders.balance`, with English and Chinese keys.
- `cb7841d85`: account status adds `expired` in English and Chinese.
- The GroupPlatform extension required `ChannelsView` to fill its exhaustive fallback-suggestion record; all three values are empty and remain outside `platformOrder`.

## Commands

- `go test -tags=unit ./internal/handler/admin -run "TestGroupPlatformBinding" -count=10` — PASS.
- `go test ./internal/handler/admin -run "^$" -count=1` — PASS.
- `corepack.cmd pnpm --dir frontend install --frozen-lockfile` — PASS; worktree-local dependencies restored with no manifest or lockfile change.
- `corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/CNProviderBalanceCell.spec.ts src/views/admin/__tests__/GroupsView.columnSettings.spec.ts src/views/admin/__tests__/GroupsView.modelPricing.spec.ts` — PASS, 3 files / 7 tests.
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__` — PASS, 28 files / 117 tests.
- `corepack.cmd pnpm --dir frontend run typecheck` — PASS.
- `git diff --check`, allowlist, conflict/index, and provenance checks — PASS.

## Risks

- Vitest emitted existing Browserslist stale-data and router-mock warnings; all tests passed.
- No browser, provider, database, migration, deployment, container, or remote operation was used.

## Knowledge Candidates

- None.
