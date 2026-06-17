### PASS: upstream-main-openai-quota-reset-s17

## Executed Checks

- `go test -tags=unit ./internal/service -run "TestOpenAIQuota" -count=1` passed.
- `go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1` passed.
- `go test ./internal/service -run "^$" -count=1` passed.
- `go test ./internal/handler/admin -run "^$" -count=1` passed.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"` passed, 2 files / 20 tests.
- `git diff --check` passed with only repository line-ending warnings.
- Denied-path audit returned `NO_DENIED_PATHS`.

## Diff Review

- Backend quota access is isolated in `OpenAIQuotaService` and uses the existing token provider plus privacy client factory rather than creating a new HTTP path.
- Admin handler and routes use `/api/v1/admin/openai/accounts/:id/...`, leaving the local account quota reset endpoint unchanged.
- The service validates OpenAI OAuth account prerequisites before sending upstream requests.
- Frontend quota controls are only rendered in the OpenAI OAuth account usage cell branch and keep existing local usage-window UI intact.
- New frontend copy is placed in the local modular admin account locale files.
- The unit-only export helper is guarded by `//go:build unit` and ordinary package compile checks passed.

## Contract Compliance

- Did not merge or rebase `upstream/main`.
- Did not modify `backend/ent/**`, `backend/migrations/**`, or `backend/cmd/server/VERSION`.
- Did not modify public pages, payment pages, Canvas, Studio Bridge, model market, or production configuration.
- No new third-party dependency was added.

## Residual Risk

- Real upstream ChatGPT WHAM quota and reset behavior still needs a staging or controlled admin account check if production credentials are available.
- Full frontend regression remains out of scope for S17 because known unrelated failures were already documented during S15.

## Recommendation

PASS for the contracted S17 scope. `b81694929` is suitable as an independent merge batch and should not be mixed with broader upstream main integration.
