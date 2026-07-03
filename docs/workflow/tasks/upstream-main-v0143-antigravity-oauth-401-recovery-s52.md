---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 18:17 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-antigravity-oauth-401-recovery-s52

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `d0a1443a41eb5d0dfc51c09916824ee323119c64` so Antigravity OAuth accounts that hit upstream 401 can enter the same temporary-unschedulable refresh window as other OAuth accounts instead of being permanently marked error, while accounts without `refresh_token` still become error.

## Success Criteria
- `RateLimitService.HandleUpstreamError` treats Antigravity OAuth 401 with a refresh token like other OAuth 401 cases:
  - invalidates token cache,
  - sets temporary unschedulable state,
  - keeps status recoverable for the token refresh worker.
- Antigravity OAuth 401 without `refresh_token` still calls SetError with a clear no-refresh-token reason.
- `TokenRefreshService` successful refresh path clears temporary-unschedulable state for Antigravity OAuth 401 recovery candidates where applicable.
- Existing OpenAI/Anthropic/Gemini OAuth 401 behavior is unchanged.
- No frontend, Ent, migrations, deploy, README, `.github`, knowledge, generated wire, or config default changes are introduced.

## Allowed Paths
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_401_test.go`
- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/token_refresh_service_candidates_test.go`
- `docs/workflow/tasks/upstream-main-v0143-antigravity-oauth-401-recovery-s52.md`
- `docs/workflow/worker-results/upstream-main-v0143-antigravity-oauth-401-recovery-s52-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-antigravity-oauth-401-recovery-s52-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of `upstream/main` or the full release.
- Do not change temp-unschedulable duration config semantics.
- Do not change non-OAuth 401 handling.
- Do not change Antigravity privacy, quota, gateway routing, model mapping, or failover delay behavior.
- If local tests do not have `token_refresh_service_candidates_test.go`, adapt the equivalent assertions to existing token refresh tests rather than creating broad unrelated test scaffolding.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestRateLimitService_HandleUpstreamError_OAuth401SetsTempUnschedulable|TestRateLimitService_HandleUpstreamError_OAuth401NoRefreshTokenSetsError|TestTokenRefreshService_.*TempUnsched|TestTokenRefreshService_.*Candidates|TestTokenRefreshService_RefreshWithRetry_Antigravity" -count=1
cd ..
git diff --check -- backend/internal/service/ratelimit_service.go backend/internal/service/ratelimit_service_401_test.go backend/internal/service/token_refresh_service.go backend/internal/service/token_refresh_service_test.go backend/internal/service/token_refresh_service_candidates_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-antigravity-oauth-401-recovery-s52.md docs/workflow/worker-results/upstream-main-v0143-antigravity-oauth-401-recovery-s52-result.md docs/workflow/qa-reports/upstream-main-v0143-antigravity-oauth-401-recovery-s52-qa.md
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/|deploy/|knowledge/|\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on a S52 branch or the current S52 continuation branch.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-antigravity-oauth-401-recovery-s52-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-antigravity-oauth-401-recovery-s52-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if the fix requires schema, migration, scheduler snapshot contract, or API DTO changes.
- Stop if token refresh tests reveal broad existing failures unrelated to Antigravity OAuth 401 recovery; document the blocker instead of expanding scope.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 18:17 +08:00.
- Verdict: approved.
- Reason: the upstream patch is a narrow backend recovery fix, local code still has the old Antigravity exclusion, and behavior can be verified with focused rate-limit/token-refresh tests.
