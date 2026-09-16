### DONE: s303-regression-closure

## Scope

- Changed only `backend/internal/server/middleware/s303_auth_lookup_test.go` and `frontend/src/api/__tests__/client.spec.ts`.
- Preserved the controller-owned pre-existing changes in `docs/workflow/main-log.md` and `docs/workflow/status.md`.
- Installed frontend dependencies in this E: worktree with `pnpm.cmd install --frozen-lockfile` using pnpm 11.19.0. It created an untracked `frontend/pnpm-workspace.yaml` build-approval file; that task-generated file was removed and no manifest or lockfile changed.

## Regression Coverage

- Backend adds 2 `TestS303` top-level tests and 17 subcases: JWT lookup has active, direct/wrapped `ErrUserNotFound`, timeout and generic internal errors; Admin runs active, direct/wrapped not-found, timeout, generic internal and nonadmin cases through both HTTP `Authorization` and WebSocket `Sec-WebSocket-Protocol: jwt.<token>` extraction.
- Frontend adds 8 client cases. The focused Vitest invocation includes the existing 7 `tokenRefresh.spec.ts` cases and 21 `client.spec.ts` cases, for 28 total; the 8 added client cases cover network/429/500/503 session retention, refresh 401/403 and malformed-success cleanup, and changed-session precedence over an Axios network refresh failure. Each relevant case replaces `window.location` with a setter-spy mock and `afterEach` unconditionally restores the original descriptor: transient/changed-session branches assert zero assignments, while invalid-session branches assert a `/login` assignment. The changed-session case also asserts the new `token_expires_at` exact value.

## Commands And Results

- PASS: `go test -tags unit ./internal/server/middleware -list 'TestS303'` listed `TestS303JWTAuthLookupOutcomes` and `TestS303AdminAuthLookupOutcomesHTTPAndWebSocket`.
- PASS: `go test -tags unit ./internal/server/middleware -run 'TestS303|TestJWTAuth|TestAdminAuth' -count=1`.
- PASS: `npm.cmd run test:run -- src/api/__tests__/client.spec.ts src/api/__tests__/tokenRefresh.spec.ts` — 2 files, 28 tests.
- PASS: `npm.cmd run typecheck`.
- PASS: `npm.cmd run build`.
- PASS: `go build ./...`.
- PASS: `gofmt -l backend/internal/server/middleware/s303_auth_lookup_test.go` had no output; `git diff --check` and `git ls-files -u` had no findings.

## Notes

- The final Vitest run has no jsdom navigation diagnostic because the redirect effect is asserted through the per-case location setter spy.
- No provider, database, container, authenticated browser/runtime, commit, or production-code validation was performed. Database fault injection and authenticated runtime remain roadmap gaps.
