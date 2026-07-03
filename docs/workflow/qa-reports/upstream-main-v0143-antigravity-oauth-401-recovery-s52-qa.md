### PASS: upstream-main-v0143-antigravity-oauth-401-recovery-s52

# QA Report

## Findings
- No blocking findings for the S52 code path.
- A temporary, uncommitted smoke test verified the core runtime behavior:
  - Antigravity OAuth 401 with `refresh_token` sets temporary-unschedulable and does not call SetError.
  - Antigravity OAuth 401 without `refresh_token` calls SetError and does not set temporary-unschedulable.
- Committed unit tests were updated to encode the same expectations, plus Antigravity refresh success clearing temp-unschedulable state.

## Executed Checks
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestS52AntigravityOAuth401" -count=1
```

Result: PASS. The temporary smoke test file was removed after the run and is not part of the final diff.

```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test -tags=unit ./internal/service -run "TestRateLimitService_HandleUpstreamError_OAuth401SetsTempUnschedulable|TestRateLimitService_HandleUpstreamError_OAuth401NoRefreshTokenSetsError|TestTokenRefreshService_RefreshWithRetry_Antigravity|TestTokenRefreshService_RefreshWithRetry_ClearsTempUnschedulable" -count=1
```

Result: BLOCKED by unrelated pre-existing package compile errors:
- `proxyRepoStub` / `proxyRepoStubForAdminList` missing `CountUserOwnedAccountsByProxyID`.
- Legacy billing tests still reference removed `ImageOutputPriceExplicit` and the old `computeTokenBreakdown` signature.

```powershell
git diff --check -- backend/internal/service/ratelimit_service.go backend/internal/service/ratelimit_service_401_test.go backend/internal/service/token_refresh_service_test.go
```

Result: PASS.

## Scope Review
- Touched backend service logic and focused tests only.
- No denied paths were touched in the working diff before staging.

## Unverified Risks
- The package-wide committed `unit` tests cannot be executed until the unrelated service test baseline is repaired.
- No live Antigravity upstream 401 recovery was exercised.

## Recommendation
Ship S52 after final scoped staging, cached diff check, and denied-path audit. Track the unrelated `-tags=unit` package compile baseline separately.
