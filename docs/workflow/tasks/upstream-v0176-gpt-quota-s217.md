---
task_id: upstream-v0176-gpt-quota-s217
phase: contract-draft
role: Generator
worker_model: gpt-5.6-sol
qa_worker_model: gpt-5.6-terra
---

# Upstream GPT/Codex Quota Correctness S217

## Goal

Behaviorally port the locally missing quota fixes from upstream `v0.1.176`
without merging its divergent history: correct personal subscription expiry,
avoid penalizing an account for an HTML 403, and make successful reset-credit
responses update the admin list without enabling a stale-credit retry.

## Success Criteria

- `358e4a89a`: when the ID token provides a personal plan and `accounts/check`
  resolves a different POID workspace, `subscription_expires_at` is fetched for
  the personal `chatgpt_account_id`, not copied from the workspace entitlement.
  Missing account identifiers preserve the existing compatibility fallback.
- `12abb5470`: a non-structured HTML 403 for an OpenAI account does not increase
  the 403 counter, temporarily unschedule, or permanently disable the account.
  Structured JSON OpenAI 403, plain-text 403, and non-OpenAI platform behavior
  retain their existing penalties.
- `54a2bcfd1` remaining behavior: a reset success applies returned quota/account
  metadata to the row and does not issue an automatic second quota request. If
  the response lacks a fresh quota, the live reset-credit count becomes unknown
  and reset stays disabled; only an explicit user refresh uses the audited POST
  route. Existing S188 detached, recovery-first post-processing remains
  unchanged.
- `99b31067f` and `3d3aee2e` are evidence-only exclusions: rerun their local
  eligibility regressions but do not add the absent upstream threshold framework.
- Every added regression uses a test-local HTTP transport or fixture. No real
  OpenAI/ChatGPT/provider request is permitted.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen local baseline: `main@fbac8466e`.
- Upstream provenance: `358e4a89a`, `12abb5470`, `54a2bcfd1`; all are in
  upstream `v0.1.176` but direct `git apply --check` is not the implementation
  strategy in this divergent checkout.
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, this contract,
  and `docs/workflow/tasks/upstream-v0171-openai-quota-reset-recovery-s188.md`.

## Allowed Paths

- `backend/internal/service/openai_oauth_service.go`
- `backend/internal/service/openai_privacy_service.go`
- `backend/internal/service/openai_subscription_test.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_403_html_test.go`
- `backend/internal/handler/admin/openai_oauth_handler.go`
- `backend/internal/handler/admin/openai_oauth_handler_reset_quota_test.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/service/openai_quota_service.go`
- `backend/internal/service/openai_quota_reset_cache_test.go`
- `frontend/src/api/admin/accounts.ts`
- `frontend/src/types/index.ts`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/views/admin/AccountsView.vue`
- `docs/workflow/worker-results/upstream-v0176-gpt-quota-s217-result.md`
- `docs/workflow/qa-reports/upstream-v0176-gpt-quota-s217-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `outputs/`
- `backend/migrations/**`, `backend/ent/**`, generated wiring, dependencies,
  deployment, container, production configuration, and all unrelated frontend
  views/components.
- Any generic scheduling-threshold settings, database persistence, or frontend
  feature associated with upstream `7c62382d0`.
- Git reset/rebase/force actions, remote push, provider traffic, and production
  data operations.

## Constraints

- Work only in `E:/codex-worktrees/sub2api/s217-gpt-quota` after the contract is
  approved. Keep each of the three behavioral slices independently reviewable.
- Preserve local account recovery, cache key, fallback, and error-response
  compatibility. Do not copy upstream test scaffolding that depends on absent
  local runtime types.
- A positive reset-credit snapshot without expiration details is invalid; expired
  credits must not enable a reset action after rehydration.
- The HTML 403 regression must use the default Go test tag with local minimal
  fakes. Do not rely on the known-broken `-tags=unit` aggregate in this checkout.
- The reset flow must never invoke `GET /quota` or `POST /quota/refresh`
  automatically after a successful reset. The new POST is explicit-user-refresh
  only, and a route contract must prove GET remains read-only while POST is
  registered under the existing admin OpenAI group.
- Do not install or alter tracked dependency manifests. If frontend executables
  are absent, report the exact environment blocker to QA; do not misreport it as
  a product regression.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -run '^(TestFetchChatGPTSubscriptionExpiresAt|TestFetchChatGPTAccountInfo_(ReportsAccountID|WorkspaceEntitlementDoesNotOverridePersonalSubscription)|TestEnrichTokenInfo_(WorkspaceEntitlementDoesNotOverridePersonalSubscription|SameAccountDoesNotRepeatPersonalSubscriptionLookup|MissingAccountIDPreservesCompatibilityFallback)|TestHandleUpstreamError_OpenAIHTML403|TestHandleUpstreamError_OpenAIStructured403|TestHandleUpstreamError_HTML403OnOtherPlatformsUnchanged|TestIsHTMLResponse|TestOpenAIGatewayService_SelectAccountForModelWithExclusions_(AutoPauseBy5hThreshold|AutoPauseBy7dThreshold|AllowsBelow5hThreshold|StaleUsageWindowResetSkipsPause|FreshUsageWindowStillPauses|StaleUsageSnapshotSkipsPause_Issue2994|FreshExhaustedSnapshotStillPauses_Issue2994)|TestOpenAIGatewayService_SelectAccountByPreviousResponseID_QuotaAutoPausedMiss|TestBuildCodexUsageExtraUpdates_FreshAccountUsedPercentNotInverted_Issue2994)$' -count=10
go test ./internal/handler/admin -run '^TestOpenAIOAuthHandlerResetQuota' -count=10
go test ./internal/service -run '^TestOpenAIQuotaServiceCacheResetCreditsSnapshot$' -count=10
go test ./internal/server -run '^TestAdminOpenAIQuotaRefreshRouteContract$' -count=10
go test ./internal/service -count=1
go test ./internal/server -count=1
go test ./cmd/server -run '^$' -count=0

Set-Location F:/mcplugins/sub2api
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts
corepack.cmd pnpm --dir frontend exec vue-tsc --noEmit
corepack.cmd pnpm --dir frontend run build
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-v0176-gpt-quota-s217-result.md`
  with first line `### DONE: upstream-v0176-gpt-quota-s217`, `### BLOCKED: ...`,
  or `### FAILED: ...`. List changed files, per-slice tests, external-call
  boundary, known risks, and knowledge candidates.
- Commit only scoped source/test/report changes in the isolated worktree.

## Stop Rules

- Stop and report `BLOCKED` if a slice needs a schema migration, generic
  threshold feature, dependency manifest change, real provider call, denied
  path, or a compatibility change to reset consumption/recovery semantics.
- Stop before integrating any slice whose focused regression is missing or
  fails. A frontend test-toolchain absence is an environment blocker, not a
  reason to skip backend or static gates.
- Do not integrate, push, deploy, or delete anything. Codex performs independent
  QA, topology review, and local-main integration after PASS evidence.
