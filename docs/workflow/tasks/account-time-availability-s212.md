# Task Contract: account-time-availability-s212

## Task ID

`account-time-availability-s212`

## Status

`approved / fix amendment`

## Role

Codex plans and evaluates the change. The Generator implements only the
approved contract. Independent QA reviews the final diff and runtime evidence.

## Goal

Give every account an optional same-day daily availability window. When the
window is enabled, the account is eligible for new-request scheduling only in
the server-timezone interval `[start, end)`. The window dynamically excludes a
candidate; it never changes persisted account state, API Key/group bindings, or
an already-started upstream request.

## Success Criteria

- Existing `accounts.extra` stores `account_availability_enabled`,
  `account_availability_start`, and `account_availability_end`; no migration or
  new top-level API field is introduced.
- Valid windows use `HH:MM`, are same-day, and require `start < end`.
  Configuration writers reject non-boolean enable values, missing/invalid time
  fields, and equal/reversed/cross-midnight windows. A disabled valid window is
  retained for later re-enablement.
- Runtime invalid or absent legacy configuration fails open to existing account
  behavior. An enabled valid window excludes only outside-time candidates.
- `IsSchedulableAt` and context-aware scheduling use the request's first start
  timestamp. The middleware creates it once; automatic retry, failover,
  pinned/sticky, snapshot, private-pool, Gateway, OpenAI, Gemini-compatible,
  and WebSocket selections retain it. Zero/missing internal contexts use
  current server time.
- Manual inactive/error state, `schedulable=false`, expiry, rate limit,
  overload, temporary pause, quota, model/capability, group, pinned-account and
  private-pool rules remain enforced. An account outside the window must not
  be selected through a sticky, pinned, or Antigravity model-scoped shortcut.
- Create and edit account dialogs show an account time-availability toggle,
  server-timezone and interval guidance, native time inputs, and client-side
  validation. Turning the toggle off retains valid start/end values.
- No API Key needs recreation. Account group bindings, persisted `status`,
  persisted `schedulable`, usage billing, group time pricing, and per-request
  media pricing are unchanged.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen working tree: existing uncommitted S211 scope must be preserved.
- Workflow fact source: `docs/workflow/status.md`
- Account model/scheduling: `backend/internal/service/account.go`, Gateway,
  OpenAI and Gemini-compatible services.
- Client request middleware: `backend/internal/server/middleware/client_request_id.go`.
- Account dialogs: `frontend/src/components/account/CreateAccountModal.vue` and
  `frontend/src/components/account/EditAccountModal.vue`.
- The user-owned `outputs/` remains untouched.

## Allowed Paths

- `backend/internal/pkg/ctxkey/ctxkey.go`
- `backend/internal/server/middleware/client_request_id.go`
- `backend/internal/server/middleware/client_request_id_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_time_availability_test.go`
- `backend/internal/service/antigravity_quota_scope.go`
- `backend/internal/service/account_pool_strategy.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_account_time_availability_test.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/scheduler_snapshot_service.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_account_time_availability_test.go`
- `backend/internal/service/user_account_service.go`
- `backend/internal/service/user_account_service_account_time_availability_test.go`
- `frontend/src/components/account/AccountTimeAvailabilityWindow.vue`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/AccountTimeAvailabilityWindow.spec.ts`
- `frontend/src/components/account/__tests__/CreateAccountModal.timeAvailability.spec.ts`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/utils/peak-rate.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/account-time-availability-s212.md`
- `docs/workflow/worker-results/account-time-availability-s212-result.md`
- `docs/workflow/qa-reports/account-time-availability-s212-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- Every path not listed above, including Ent/schema, migrations, generated
  files, repositories, configuration, dependencies, lockfiles, Docker,
  deployment, remote refs, and `outputs/`.
- Provider calls, shared PostgreSQL/Redis access, container changes, remote
  push, deployment, production traffic, account status mutation, API Key
  rebinding/recreation, or changes to group time-rate behavior.

## Constraints

- Use the server's local time and left-closed, right-open `[start, end)`
  interval. Do not add cross-midnight windows, multiple windows, timers, or
  background jobs.
- Treat account availability as a runtime predicate. Do not write its current
  result back to `Status` or `Schedulable`; caches may still coarse-filter by
  their existing persisted fields and must be dynamically filtered afterward.
- Capture `RequestStartedAt` once at request middleware entry. Do not overwrite
  it during handler forwarding, retry, sticky lookup, failover, or async work.
- Invalid persisted legacy values must not take an account offline. New writes
  must return a validation error before persistence.
- Keep changes minimal. Do not refactor unrelated scheduler rules or alter
  billing, media pricing, group factor, response schemas, API Key bindings, or
  UI outside the account dialogs.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run '^(TestAccountAvailability|TestAdminService_.*AccountAvailability|TestUserAccountService.*Availability|TestGateway.*AccountAvailability|TestOpenAI.*AccountAvailability|TestGemini.*AccountAvailability)' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S212 focused service regressions failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S212 service package regression failed' }
go test ./internal/server/middleware -run '^TestClientRequestID' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S212 middleware regression failed' }
go test ./internal/handler ./internal/handler/admin -count=1
if ($LASTEXITCODE -ne 0) { throw 'S212 handler package regression failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S212 server compile check failed' }
Pop-Location

Push-Location frontend
npm.cmd run test:run -- src/components/account/__tests__/AccountTimeAvailabilityWindow.spec.ts src/components/account/__tests__/CreateAccountModal.timeAvailability.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S212 account availability component tests failed' }
npm.cmd run lint:check
if ($LASTEXITCODE -ne 0) { throw 'S212 frontend lint failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S212 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S212 frontend build failed' }
Pop-Location

gofmt -d backend/internal/pkg/ctxkey/ctxkey.go backend/internal/server/middleware/client_request_id.go backend/internal/server/middleware/client_request_id_test.go backend/internal/service/account.go backend/internal/service/account_time_availability_test.go backend/internal/service/antigravity_quota_scope.go backend/internal/service/account_pool_strategy.go backend/internal/service/gateway_service.go backend/internal/service/gateway_account_time_availability_test.go backend/internal/service/openai_account_scheduler.go backend/internal/service/openai_account_scheduler_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_ws_forwarder.go backend/internal/service/gemini_messages_compat_service.go backend/internal/service/scheduler_snapshot_service.go backend/internal/service/admin_service.go backend/internal/service/admin_service_account_time_availability_test.go backend/internal/service/user_account_service.go backend/internal/service/user_account_service_account_time_availability_test.go
git diff --check
git diff --name-only
git ls-files -u
git status --short
```

Browser acceptance uses a task-owned Playwright profile at desktop and 390px
width, stores evidence outside the repository under
`E:/codex-artifacts/sub2api/account-time-availability-s212`, and closes only
the task-owned session/profile/daemon processes.

## Output

- Generator result: `docs/workflow/worker-results/account-time-availability-s212-result.md`.
- Independent QA: `docs/workflow/qa-reports/account-time-availability-s212-qa.md`.
- Workflow status/log entries for contract review, build, QA, and final verdict.

## Stop Rules

- Stop if the behavior requires schema/migration, a new public API field, a
  timer/background job, persisted status mutation, API Key rebinding, or a
  routing-policy redesign.
- Stop if a known candidate shortcut can bypass the enabled availability window.
- Stop if a request retry/failover recomputes the window from a later current
  time after the original request timestamp was captured.
- Stop if invalid legacy data blocks an account or a new invalid configuration
  can reach persistence.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_qa_budget_usd: `0.10`

## Contract Review

`PASS / contract-approved`: `Account.Extra` already traverses account create,
update, response, scheduler cache, and snapshot boundaries. Existing Gateway,
OpenAI, Gemini-compatible, private-pool, pinned, and sticky selectors expose
post-retrieval runtime checks where an availability predicate can be applied.
`ClientRequestID` covers the gateway routes and can capture one context value
without changing request payloads. A small shared account-dialog control avoids
two divergent time-window validators. No schema, external API, account-state,
binding, billing, provider, or deployment change is required.

## Contract Amendment

`PASS / Planner-Evaluator`: `Account.IsSchedulableForModelWithContext` is the
Antigravity model-scoped candidate gate. It must call the context-aware base
predicate so an enabled availability window cannot be bypassed after the
model-level rate-limit check. This amendment allowlists only
`antigravity_quota_scope.go` and requires the existing account-availability
regression file to prove outside-window rejection and start-boundary
acceptance. Product scope, storage, API surface, and all denied paths remain
unchanged.

`PASS / review-fix`: the pre-commit multi-agent review found two uncovered
writers and two client-side boundary gaps. `UserAccountService.Create` and
`Update` must reject invalid availability values before persistence, new writes
must accept only exact `HH:MM` while legacy runtime parsing remains fail-open,
and account dialogs must validate before OAuth/direct submission while
retaining a valid disabled window. The frontend must normalize compatible
single-digit-hour legacy values for editing. Focused service and parent-dialog
payload tests are required; schema, top-level API fields, account state, routing,
and deployment scope remain unchanged.
