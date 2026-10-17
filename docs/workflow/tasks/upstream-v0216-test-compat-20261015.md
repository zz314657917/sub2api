---
status: approved
review_verdict: PASS
task_id: upstream-v0216-test-compat-20261015
worker_model: gpt-6.1-sol
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
spec_ref: docs/workflow/tasks/upstream-v0216-test-compat-20261015.md
---

# Task Contract

## Task ID
upstream-v0216-test-compat-20261015

## Role
Independent gpt-6.1-sol Reviewer approves this contract; a separate Developer implements it and a different QA validates the result. Controller reviews exact patches and integrates only accepted bytes.

## Goal
Repair six stale unit-test fixtures against existing local production APIs so the previously blocked gateway batch can execute canonical tagged tests. Test compatibility only; no production logic or assertion weakening.

## Context
- User authorized continuing selective upstream integration. Frontend accepted 15 paths are already in main, uncommitted; preserve them and every unrelated dirty file.
- Clean base service unit compilation fails at duplicate stringPtr, obsolete billing parameters, obsolete third countTokens return, unused context import, and removed proxy fallback/expiry fields. Full diagnostic: E:/codex-runtime/upstream-v0216-20261015/gateway-baseline-unit-all-errors.log.
- Resume the existing detached worktree E:/codex-worktrees/sub2api/upstream-v0216-test-compat-20261015 on the stated base, not the frontend candidate. No commits. This corrects the previous documentation-only worktree path mismatch.
- Allowed corrections: rename test-local stringPtr helper and its calls; remove unused import; pass current computeTokenBreakdown(pricing,tokens,rateMultiplier,serviceTier,applyLongCtx) and calculateCostInternal(model,tokens,rateMultiplier,serviceTier,channelPricing) arguments; bind current two-value buildCountTokensRequest return; remove nonexistent FallbackMode/ExpiryWarnDays fields from proxy test literals while retaining update-boundary assertions.
- Amendment 2026-10-17: the original five corrections expose further pre-existing compilation errors in external package service_test. Evidence: E:/codex-runtime/upstream-v0216-20261015/compat-dev-20261017-164729-01-impacted.log. Add only auth_service_email_bind_test.go: define its missing createEmailBindTestUser helper using the same Ent create fields/defaults already used by adjacent tests; adapt emailBindUserRepoStub.Update to current explicit UserUpdateFields semantics, preserving unselected fields and the email index; add fail-loud AdjustBalance/SetBalance stubs, since email identity binding must not call balance mutation. No production changes or no-op fallback for unexpected balance calls.

## Success Criteria
- All six fixtures compile with -tags unit using real canonical files, no overlays, file hiding, build-tag changes or skipped tests.
- Preserve test names, assertions, business inputs, expected costs and both proxy update paths. Test helper rename must not change production helper or other callers.
- Exact impacted cases execute nonzero with no skips, then relevant existing billing/health/countTokens/proxy and Gemini policy regressions execute. Record pre-existing runtime failures separately, do not remove or hide assertions to force PASS.
- Default-tag service tests compile, production build succeeds, formatting/diff checks pass.
- Preserve all existing email-bind test names, inputs, assertions and concurrency expectations; run every TestAuthServiceBindEmailIdentity and TestAuthServiceSendEmailIdentityBindCode case without skips. If any expectation fails, record it and stop rather than adjust production behavior or weaken the test.
- Independent QA seals the six test file hashes. No main integration before PASS; if main owners differ incompatibly, stop instead of importing production edits.

## Allowed Paths
- backend/internal/service/ops_health_score_test.go
- backend/internal/service/billing_service_test.go
- backend/internal/service/billing_service_unified_test.go
- backend/internal/service/gateway_context_management_test.go
- backend/internal/service/proxy_update_probe_invalidation_test.go
- backend/internal/service/auth_service_email_bind_test.go
- docs/workflow/worker-results/upstream-v0216-test-compat-20261015.md
- docs/workflow/qa-reports/upstream-v0216-test-compat-20261015.md

## Denied Paths
All production files, other tests, dependencies/locks, handler, migrations, configuration, containers, databases/providers, frontend, shared knowledge/memories. Main-tree product files may only be copied by controller after independent acceptance.

## Constraints
No commits/staging/push/merge/rebase/cherry-pick; no changes outside the six test files. Keep independent roles and exact ownership. Logs/artifacts only E:/codex-runtime/upstream-v0216-20261015/. Do not run whole unrelated suites as a substitute for the named cases.

## Acceptance Commands
From isolated backend:
```powershell
go test -p 1 -tags unit ./internal/service -run '^Test(ComputeTokenBreakdown_(ExplicitZeroImagePrice_NoFallback|NonExplicitZeroImagePrice_FallsBackToOutput)|CalculateCostUnified_NilResolver_FallsBackToOldPath|BuildCountTokensRequest_OAuthMimic_DropsInjectedMaxTokens|BothProxyUpdateServicesUseRepositoryUpdateBoundary|ComputeDashboardHealthScore.*|ComputeBusinessHealth|ComputeInfraHealth)$' -count=1 -v
go test -p 1 -tags unit ./internal/service -run 'TestShouldFailoverGeminiUpstreamError|TestGemini.*ErrorPolicy|TestBuildCountTokensRequest|TestCalculateCostUnified|TestGetModelPricing|TestComputeTokenBreakdown|TestBothProxyUpdate' -count=1 -v
go test -p 1 -tags unit ./internal/service -run '^TestAuthService(BindEmailIdentity|SendEmailIdentityBindCode)' -count=1 -v
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61' -count=1
go build -p 1 ./...
git diff --check
```
Require actual nonzero pass counts and no skips for impacted tests and gateway policy cases; discover exact policy names if regex omits any. No compilation overlays or selective file removal. Scoped gofmt only the six paths. Preserve command output and test counts.

## Output
Developer DONE/FAILED/BLOCKED report and separate QA PASS/FAIL/BLOCKED report. QA writes E:/codex-runtime/upstream-v0216-20261015/test-compat-sealed-files.json as exact path-to-SHA256 dictionary for six files, including scope check and unchanged production/dependencies evidence.

## Stop Rules
Stop on unavailable configured model, further compiler errors requiring other files, failed impacted expectations requiring production fixes, unsafe scope expansion, owner drift or denied changes. Controller must separately review any revised scope; do not weaken assertions.
