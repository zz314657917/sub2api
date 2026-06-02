# Task Contract

## Task ID
upstream-main-oauth-401-no-credentials-write-test-s2i

## Role
Codex acts as Planner, Generator, and Final Evaluator for this tiny backend unit-test Sprint. Update only the OAuth 401 no-credentials-write regression tests selected here.

## Goal
Port the safe test subset of upstream `be3613593 test(oauth): update OAuth 401 tests to match new no-write behavior` onto the current upstream-sync branch. Local production code already matches upstream `6aec50501 fix(oauth): don't overwrite credentials JSONB in 401 handler`; the remaining gap is that the `//go:build unit` test file still asserts the removed write-back behavior.

## Success Criteria
- `RateLimitService.HandleUpstreamError` OAuth 401 tests assert `updateCredentialsCalls == 0`.
- The previous `OAuth401UsesCredentialsUpdater` test is converted into a regression test proving the handler does not persist request-start credentials snapshots.
- Missing-refresh-token OAuth 401 behavior remains unchanged.
- No production code, schema, migration, config, public API, frontend, gateway routing, OpenAI WS bridge, or Responses bridge files are changed.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-openai-failover-body-remap-s2h`
- Work branch: `codex/upstream-main-oauth-401-no-credentials-write-test-s2i`
- Upstream source commits:
  - implementation already equivalent: `6aec50501`
  - test update to port: `be3613593`
- Main worktree `F:/mcplugins/sub2api` must not be modified.

## Allowed Paths
- `backend/internal/service/ratelimit_service_401_test.go`
- `docs/workflow/tasks/upstream-main-oauth-401-no-credentials-write-test-s2i.md`
- `docs/workflow/worker-results/upstream-main-oauth-401-no-credentials-write-test-s2i-result.md`
- `docs/workflow/qa-reports/upstream-main-oauth-401-no-credentials-write-test-s2i-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
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
- Do not cherry-pick if it attempts to widen the patch beyond allowed paths.
- Do not change production OAuth 401 behavior; local `ratelimit_service.go` is already equivalent to upstream no-write semantics.
- The target test file has `//go:build unit`; acceptance must include `-tags unit`.
- If tests require production code changes, stop and split a new Sprint.

## Candidate Commit
- Primary: `be3613593 test(oauth): update OAuth 401 tests to match new no-write behavior`

## Explicitly Deferred
- `eba204632` OpenAI OAuth refresh enrichment because it changes service wiring and generated wire paths.
- Migration/schema/frontend/API tasks such as group custom models list are outside this tiny test-only Sprint.
- OpenAI WS bridge/failover and Responses bridge redesign patches remain deferred.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test -tags unit ./internal/service -run "OAuth401InvalidatorError|OAuth401DoesNotOverwriteCredentials|OAuth401NoRefreshToken" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-oauth-401-no-credentials-write-test-s2i-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-oauth-401-no-credentials-write-test-s2i-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval and QA events.
- Update `knowledge/tasks/current-task.md` with the current handoff snapshot after QA.

## Stop Rules
- Stop if implementation requires touching denied paths.
- Stop if the OAuth 401 tests require changing production code to pass.
- Stop if unit-tag test failures point to unrelated production behavior changes outside this contract.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
