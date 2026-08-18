### PASS: upstream-cn-group-entry-s228

## Scope

- QA base: `b0a7a6e8b` (`docs(workflow): record S228 controller result`).
- Reviewed the two business commits `df43f3876` and `26a5dec9d`; no product files were changed by QA.
- Allowed-path, dependency, conflict/index, and upstream-provenance gates passed. The three source commits remain ancestors of `upstream/main@8869775ed`.

## Independent Verification

- `corepack.cmd pnpm --dir frontend install --frozen-lockfile` - PASS; no manifest or lockfile change.
- `go test -tags=unit ./internal/handler/admin -run "TestGroupPlatformBinding" -count=10` from `backend` - PASS.
- `go test ./internal/handler/admin -run "^$" -count=1` from `backend` - PASS.
- `corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/CNProviderBalanceCell.spec.ts src/views/admin/__tests__/GroupsView.columnSettings.spec.ts src/views/admin/__tests__/GroupsView.modelPricing.spec.ts` - PASS, 3 files / 7 tests.
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__` - PASS, 28 files / 117 tests.
- `corepack.cmd pnpm --dir frontend run typecheck` - PASS.

## Review

- Create and Update bindings admit only the existing platforms plus Kimi, Zhipu, and DeepSeek; invalid values remain rejected.
- The CN platforms appear in group creation, filtering, display, and labels. `ChannelsView` supplies only empty exhaustive fallbacks and keeps the CN platforms out of `platformOrder`.
- The empty CN balance state uses `admin.accounts.cnProviders.balance`; account and proxy status consumers resolve the new `admin.accounts.status.expired` keys.
- No composite route surface, gateway behavior, pricing, migration, dependency, provider, deployment, container, database, push, or user-owned dirty path was changed.

## Notes

- Vitest reported existing Browserslist stale-data and router-mock warnings; all commands exited successfully.
