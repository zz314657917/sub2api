### PASS: upstream-v0146-safe-patches-s54

Changed files:
- Backend endpoint normalization: `endpoint.go`, `endpoint_test.go`, `stream_error_event.go`.
- OAuth account test headers: `account_test_service.go`.
- API Key concurrency: service/cache/handler DTO, gateway helper, wire, API contract, user Keys view/types/i18n/tests.

Implementation notes:
- Ported `/responses/compact` endpoint normalization while preserving local Midjourney and Tasks endpoint behavior.
- Added API Key stats-only concurrency tracking through `ConcurrencyService`, Redis concurrency cache, API key list/get enrichment, OpenAI gateway user-slot release wrapping, and WebSocket acquisition paths.
- Adapted `ProvideAPIKeyService` to preserve local Studio Bridge route settings reader and additionally set `ConcurrencyService`.
- Kept local split i18n structure by adding `keys.currentConcurrency` in `frontend/src/i18n/locales/{en,zh}/keys.ts`.
- Added a local follow-up commit to remove an upstream-only helper call from OAuth account test headers; local code already sets `chatgpt-account-id` directly.

Commands run:
- `go test ./internal/service -run "TestAPIKey.*Concurrency|TestConcurrencyService.*APIKey|TestAccountTestService|TestOpenAI.*Endpoint|Test.*Responses.*Endpoint" -count=1` PASS.
- `go test ./internal/handler -run "Test.*Endpoint|Test.*Gateway.*Concurrency|Test.*Responses.*Compact" -count=1` PASS.
- `go test ./internal/server -run "Test.*APIContract" -count=1` completed with `[no tests to run]` because `api_contract_test.go` is `//go:build unit`.
- `go test -tags=unit ./internal/server -run "TestAPIContracts" -count=1` attempted; compile failed on pre-existing stub drift in `stubAccountRepo` / `stubProxyRepo` missing methods required by `NewAdminService`.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"` PASS using a local junction to the existing main-worktree `frontend/node_modules`.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/KeysView.spec.ts"` PASS after adapting the spec to local fixed-column UI.
- `git diff --check` PASS.
- `rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .` PASS, no conflict markers.

Risks / follow-up:
- `backend/internal/server/api_contract_test.go` remains affected by unrelated unit-test stub drift; the S54 API Key contract field is present in the file, but the full unit-tag contract suite cannot compile until those existing stubs are updated.
- Frontend validation reused the existing `F:/mcplugins/sub2api/frontend/node_modules` through a worktree junction because this isolated worktree did not have dependencies installed.