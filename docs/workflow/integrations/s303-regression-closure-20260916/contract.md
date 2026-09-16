---
status: approved
review_verdict: PASS
task_id: s303-regression-closure
worker_model: gpt-5.6-terra
base_commit: 75c9ab270e266c11e5548be17abad0a891a90a15
spec_ref: docs/workflow/plans/upstream-main-integration-roadmap-20260913.md
---

## Task ID
s303-regression-closure

## Role
Terra Developer; fresh independent Terra QA; controller final integration.

## Goal
Close the missing committed regression evidence for S303. Production changes
already exist in b675e7f3c and ed91196f8. Add real regression tests for existing
middleware/client implementations, not mock reimplementations. Do not change
authentication policy, production code, tokens or storage contracts.

## Success Criteria
1. New backend unit-tag tests exercise actual JWT middleware and AdminAuth HTTP
   and WebSocket token-extraction paths. For each, found active authorized user
   reaches protected handler; direct/wrapped ErrUserNotFound returns 401 with
   USER_NOT_FOUND; timeout and internal errors return 500/INTERNAL_ERROR, never
   reaching protected handler. Nonadmin remains 403; token mismatch/revocation
   and invalid JWT remain rejected by existing tests. No real WebSocket upgrade
   or network provider required: use httptest request headers and subprotocol.
2. Frontend tests exercise actual apiClient interceptor and refresh flow with
   Axios adapter fakes. Refresh network failure (no response), 429, 500 and 503
   retain auth_token, refresh_token, auth_user, token_expires_at plus existing
   sessionStorage values; no auth_expired flag or login redirect is introduced;
   rejection exposes TOKEN_REFRESH_UNAVAILABLE and real status (0 for network).
3. Refresh 401/403 and malformed success clear invalid session and return
   TOKEN_REFRESH_FAILED. Changed-session failure takes priority over transient
   handling and preserves the new user's session (AUTH_SESSION_CHANGED). Retain
   concurrent refresh/one-flight and existing retry tests; no assertion weakening.
4. New tests are actually compiled/selected: backend names prefix TestS303 and
   use //go:build unit, frontend cases in client.spec.ts. Restore mocks, browser
   storage/location and global state after cases; no order-dependent pollution.
5. All focused tests and builds/typecheck pass. Real database fault injection and
   authenticated browser/runtime remain outstanding roadmap gates, not covered
   by fake/local regression. If existing implementation fails, report concrete
   defect and STOP rather than changing denied production owners.

## Allowed Paths
- backend/internal/server/middleware/s303_auth_lookup_test.go
- frontend/src/api/__tests__/client.spec.ts
- docs/workflow/worker-results/s303-regression-closure-result.md
- docs/workflow/qa-reports/s303-regression-closure-qa.md

## Denied Paths
All production code, old JWT tests, auth/permission architecture, dependencies
and lockfiles, schema, databases, providers, containers/deployment, commits/push,
all other worktrees, shared docs/knowledge/memory. No merge/rebase/cherry-pick.

## Constraints
Use apply_patch. Existing auth production behavior is the subject, not editable.
Frontend dependencies may be installed into this E: worktree with frozen lockfile
or use existing local dependency runtime without changing manifests/lockfiles.
Never reuse user Chrome. No browser startup required for this tests-only task.

## Acceptance Commands
backend: `go test -tags unit ./internal/server/middleware -list 'TestS303'`
backend: `go test -tags unit ./internal/server/middleware -run 'TestS303|TestJWTAuth|TestAdminAuth' -count=1`
backend: `go build ./...`
frontend: `npm.cmd run test:run -- src/api/__tests__/client.spec.ts src/api/__tests__/tokenRefresh.spec.ts`
frontend: `npm.cmd run typecheck`
frontend: `npm.cmd run build`
root: `gofmt -l backend/internal/server/middleware/s303_auth_lookup_test.go`
root: `git diff --check`; `git ls-files -u`; enumerate changed tracked/untracked
files and prove only the two allowed test paths change outside controller docs.
Report actual test counts and names; no-tests/skip is not PASS. Environment or
baseline failure must be reported, not bypassed or claimed as passing.

## Output
Worker starts ### DONE/FAILED/BLOCKED: s303-regression-closure. QA starts
### PASS/FAIL/BLOCKED: s303-regression-closure. Include tests selected, commands,
actual results, scope, findings and explicit runtime gaps.

## Stop Rules
Independent contract PASS and approved frontmatter before implementation. Stop
for needed production changes or unavailable Terra. No credentials/real traffic.
