# Task Contract

## Task ID
upstream-main-openai-oauth-refresh-enrichment-s2j

## Role
Codex acts as Planner, Generator, and Final Evaluator for this small backend service Sprint. Implement only the safe OpenAI OAuth refresh enrichment subset selected here.

## Goal
Port the bounded service behavior from upstream `eba204632 fix: enrich OpenAI OAuth token refresh` onto the current upstream-sync branch. The local code already has OpenAI account/privacy enrichment helpers, but the existing-access-token refresh path does not run enrichment and production `OpenAIOAuthService` is not wired with `PrivacyClientFactory`, so refreshed OpenAI OAuth credentials can miss subscription expiry/account metadata.

## Success Criteria
- `OpenAIOAuthService.RefreshAccountToken` reuses an existing access token without calling the OAuth refresh endpoint when `refresh_token` is absent.
- That existing-access-token path preserves stored `subscription_expires_at` and runs best-effort enrichment when a privacy client factory is injected.
- Enrichment can fill `subscription_expires_at` from ChatGPT `/backend-api/subscriptions` when accounts/check did not provide entitlement expiry.
- Production DI wires `OpenAIOAuthService` with `PrivacyClientFactory` through a small provider, without Ent/schema/codegen/migration changes.
- No public API, DTO, database schema, migration, frontend, gateway routing, OpenAI WS bridge, Responses bridge, payment, subscription notify, redeem expiry, DingTalk, or channel monitor API mode behavior is changed.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-oauth-401-no-credentials-write-test-s2i`
- Work branch: `codex/upstream-main-openai-oauth-refresh-enrichment-s2j`
- Upstream source commit: `eba204632 fix: enrich OpenAI OAuth token refresh`
- Main worktree `F:/mcplugins/sub2api` must not be modified.

## Allowed Paths
- `backend/internal/service/openai_oauth_service.go`
- `backend/internal/service/openai_privacy_service.go`
- `backend/internal/service/openai_oauth_service_refresh_test.go`
- `backend/internal/service/openai_subscription_test.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire_gen.go`
- `docs/workflow/tasks/upstream-main-openai-oauth-refresh-enrichment-s2j.md`
- `docs/workflow/worker-results/upstream-main-openai-oauth-refresh-enrichment-s2j-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-oauth-refresh-enrichment-s2j-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/ratelimit_service.go`
- `frontend/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS/Responses bridge redesign, OpenAI gateway routing redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Prefer manual port over cherry-pick because the local wiring and enrichment code already diverged from upstream.
- Keep `wire_gen.go` changes mechanically equivalent to `wire.go` provider changes; do not run Ent codegen or migrations.
- The privacy/subscription calls must remain best-effort and non-blocking on failure.
- If implementation requires changing API shape, schema, config, gateway handlers, or larger generated wiring, stop and defer to a later migration/architecture Sprint.

## Candidate Commit
- Primary: `eba204632 fix: enrich OpenAI OAuth token refresh`

## Explicitly Deferred
- OpenAI WS bridge/failover/image tool injection and Responses bridge redesign commits remain deferred.
- OpenAI generated default model strategy (`gpt-5.5`) remains deferred.
- Pricing snapshot changes remain deferred until pricing resource semantics are evaluated.
- Admin create-user balance pointer semantics remain deferred to a separate account/admin Sprint.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test ./internal/service -run "OpenAIOAuthService_RefreshAccountToken_NoRefreshTokenUsesExistingAccessToken|FetchChatGPTSubscriptionExpiresAt|OpenAI.*Refresh|OpenAI.*Privacy|OpenAI.*Subscription|BuildAccountCredentials|RefreshIfNeeded" -count=1
go test ./cmd/server -run TestNonExistent -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-oauth-refresh-enrichment-s2j-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-openai-oauth-refresh-enrichment-s2j-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval and QA events.
- Update `knowledge/tasks/current-task.md` with the current handoff snapshot after QA.

## Stop Rules
- Stop if the patch needs schema, migration, public API, gateway, WS bridge, Responses bridge, or frontend changes.
- Stop if `wire_gen.go` changes become more than the provider substitution needed for `OpenAIOAuthService`.
- Stop if target tests require network access to real OpenAI/ChatGPT services; all tests must use stubs or `httptest`.
- Stop if failures indicate unrelated production behavior outside this contract.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
