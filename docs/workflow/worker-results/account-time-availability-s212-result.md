### DONE: account-time-availability-s212

## Scope

- Added optional same-day account availability settings in `accounts.extra`:
  `account_availability_enabled`, `account_availability_start`, and
  `account_availability_end`.
- Added write validation and a fail-open runtime predicate. The predicate uses
  the first request timestamp and filters cached, private-pool, sticky, pinned,
  Gateway, OpenAI, Gemini-compatible, and WebSocket candidates without writing
  account status or bindings.
- Added the shared account-dialog time-window control. Valid start/end values
  remain stored when the toggle is disabled; enabled windows require a valid
  same-day `[start, end)` interval.

## Evidence

- Focused S212 service tests passed with `-count=10`.
- Complete service, middleware, handler/admin handler packages and server
  compilation passed.
- Component Vitest, frontend lint, typecheck, and production build passed.
- `gofmt -d`, `git diff --check`, and `git ls-files -u` were clean.

## Contract Amendment

- The Antigravity model-scoped eligibility gate now delegates to
  `IsSchedulableWithContext`, so it applies the already-captured request start
  time instead of recomputing current availability.
- `TestAccountAvailabilityExcludesAntigravityModelCandidateOutsideWindow`
  proves the candidate is rejected before the start boundary and accepted at
  the boundary. The focused service suite (`-count=10`) and complete service
  package passed after this amendment.

## Review Fix Resolution

- `UserAccountService.Create` and `Update` now validate account availability
  before persistence. New writes accept exact `HH:MM`; runtime parsing remains
  compatible with historical single-digit-hour values and invalid history
  still fails open.
- OAuth/direct creation and edit submission validate the enabled window before
  continuing. Valid disabled windows are retained and compatible legacy values
  are normalized for editing.
- The expanded focused Go suite passed with `-count=10`, including the direct
  legacy-runtime regression. Complete service/middleware/handler/server checks,
  three frontend suites (`42` tests), lint, typecheck, build, formatting,
  allowlist, diff, and index gates passed under independent Terra QA.

## Browser Note

The task-owned Edge/Playwright profile started and was closed cleanly. Desktop
and 390px snapshots reached the local application, but its Vite-only proxy had
no backend and returned `500` for public-settings requests, so auth redirected
to login before the account dialogs could be inspected. Evidence is outside the
repository at `E:/codex-artifacts/sub2api/account-time-availability-s212`.

## Constraints

No migration, schema/API expansion, API Key rebinding, status mutation,
container/deployment change, push, production traffic, or `outputs/` change was
performed.
