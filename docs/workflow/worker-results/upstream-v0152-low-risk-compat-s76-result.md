### DONE: upstream-v0152-low-risk-compat-s76

## Summary

- Replaced raw Fast/Flex user-ID inputs with a debounced email-search selector while preserving exact positive IDs and unresolved historical IDs.
- Added Grok Composer reasoning-field sanitization for the three approved aliases, including provider-prefixed names, while preserving other Grok models.
- Added Grok-aware account-selection log text and aligned Count Tokens with the existing platform-aware 404/503 no-account classifier.

## Changed Areas

- Frontend selector, Settings integration, locale keys, component tests, locale tests, and the existing Settings save-chain regression.
- Grok Responses request sanitization and focused default-tag tests.
- OpenAI-compatible selection diagnostics for Responses, Messages, Chat Completions, Count Tokens, and WebSocket logs.
- S76 workflow contract, status, log, result, and QA evidence.

## Commands Run

- `go test ./internal/service -list <S76 pattern>` discovered exactly 2 tests.
- `go test ./internal/service -run <S76 pattern> -count=1` PASS, 2 tests.
- `go test ./internal/handler -list <S76 pattern>` discovered exactly 2 tests.
- `go test ./internal/handler -run <S76 pattern> -count=1` PASS, 2 tests.
- Selector and locale Vitest PASS, 2 files / 6 tests.
- Settings save-chain Vitest PASS, 1 selected test.
- `corepack.cmd pnpm --dir frontend run typecheck` PASS.
- `corepack.cmd pnpm --dir frontend run build` PASS.
- `git diff --check` PASS.
- Changed-path audit PASS: all 18 pre-report paths were in the approved S76 allowlist.

## Contract Compliance

- No Ent, migration, repository, billing, pricing, account-type, prompt-cache, payment, deployment, VERSION, Docker, or `knowledge/**` path changed.
- The main worktree's `knowledge/05-current-focus.md` modification remained untouched.
- No dependency was installed or updated; the temporary `node_modules` junction was removed after each frontend check.

## Risks

- No authenticated administrator browser smoke or real Grok upstream request was run.
- The local admin user API cannot hydrate soft-deleted users by ID; such IDs remain visible and removable through the `User #ID` fallback.
- `go test -tags=unit` for the whole service package still fails on pre-existing duplicate helpers and stale billing/runtime test contracts. S76 uses focused default-tag tests with exact discovery gates instead.
