---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 10:47 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-user-model-stats-requested-s42

## Role
Codex acts as Planner and Final Evaluator. Implementation may be done by Codex directly only after this contract is approved. QA may be run by Codex directly after implementation.

## Goal
Port upstream `v0.1.143` commit `e236bff1e fix: aggregate user model stats by requested model`: user-facing model stats for a single user should aggregate by requested model, falling back to stored model when `requested_model` is blank.

## Success Criteria
- `usageLogRepository.GetUserModelStats` uses the same requested-model source behavior as `GetModelStatsWithFilters`.
- Grouping expression must be `COALESCE(NULLIF(TRIM(requested_model), ''), model)` through the existing local helper path.
- Preserve existing cost, actual cost, account cost, token totals, date range, and user filter semantics.
- Add a focused repository sqlmock test for `GetUserModelStats`.
- Do not touch gateway, billing service, usage billing, frontend, Ent, migrations, deploy, knowledge, or unrelated dirty files.

## Context
- Repo: `F:/mcplugins/sub2api`
- Previous completed Sprint: `upstream-main-v0143-antigravity-reasoning-params-s41`
- Latest release refreshed on 2026-07-03: GitHub releases list `Sub2API 0.1.143` as latest, tag `v0.1.143` / commit `9caa3c9c5`.
- Candidate commit: `e236bff1e fix: aggregate user model stats by requested model`.
- Current precheck:
  - `backend/internal/repository/usage_log_repo.go` is clean.
  - `backend/internal/repository/usage_log_repo_request_type_test.go` is clean.
  - Upstream `e236bff1e` touches only those two files.
  - Local `getModelStatsWithFiltersBySource` already exists and defaults `GetModelStatsWithFilters` to `usagestats.ModelSourceRequested`, so this Sprint can stay a narrow repository refactor.
- Deferred candidates remain:
  - `d0b8760eb` OpenAI plan type from inactive workspaces, because it changes broader OpenAI subscription/account info behavior.
  - `df59b8b96` OpenAI subscription expiration persistence, because it is frontend composable product-state work and should be separate.
  - `a5638a4e5` Codex session import identity, because it touches backend handler plus frontend API and needs a dedicated contract.
  - `c797159bf`, `a5781fe31`, `7869b7fe3`, and `c4128580f`, because they touch dirty gateway/count_tokens surfaces or broader account/gateway behavior.

## Allowed Paths
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md`
- `docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_count_tokens_test.go`
- `backend/internal/service/billing_*`
- `backend/internal/service/usage_*`
- `backend/internal/repository/usage_billing_*`
- `backend/internal/repository/welfare_*`
- `backend/internal/service/welfare_*`
- `backend/internal/payment/**`
- `backend/go.mod`
- `backend/go.sum`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `assets/**`
- `README*`
- `.github/**`
- Any unlisted dirty file.

## Constraints
- Do not merge/rebase `v0.1.143` or `upstream/main`.
- Do not cherry-pick broader release content.
- Keep this Sprint repository-only and `usage_logs` model stats only.
- Do not change billing write paths, usage billing service, leaderboard, admin stats, or frontend display code.
- If implementation requires changing SQL schema, generated Ent, frontend, or service-layer billing semantics, stop and split a new Sprint.
- Do not stage existing dirty files outside allowed paths.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/repository -run "TestUsageLogRepositoryGetUserModelStatsUsesRequestedModel|TestUsageLogRepositoryGetModelStatsWithFiltersRequestTypePriority|TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1

cd F:/mcplugins/sub2api
git diff --check -- backend/internal/repository/usage_log_repo.go backend/internal/repository/usage_log_repo_request_type_test.go docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md docs/workflow/status.md docs/workflow/main-log.md
git diff --cached --name-only | rg "^(backend/internal/service/gateway_service.go|backend/internal/service/openai_gateway_service.go|backend/internal/service/openai_gateway_count_tokens.go|backend/internal/service/openai_gateway_count_tokens_test.go|backend/internal/service/billing_|backend/internal/service/usage_|backend/internal/repository/usage_billing_|backend/internal/repository/welfare_|backend/internal/service/welfare_|backend/internal/payment/|backend/go.mod|backend/go.sum|backend/ent/|backend/migrations/|backend/cmd/server/|frontend/|deploy/|knowledge/|assets/|README|README_|\.github/)" || echo NO_DENIED_PATHS
```

## Output
- Backend repository diff in allowed paths only.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`

## Stop Rules
- Stop if target repository files become dirty before implementation.
- Stop if tests require gateway, frontend, billing service, Ent, migration, deploy, or dependency changes.
- Stop if local helper semantics differ from the requested-model fallback expected by upstream.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0.03`

## Review Result
- Reviewed at: 2026-07-03 10:47 +08:00.
- Verdict: approved.
- Reason: required P/G/E contract fields are present; allowed paths are limited to clean usage-log repository files and workflow artifacts; denied paths protect current gateway, billing, usage billing, frontend, Ent, migration, deploy, knowledge, and unrelated dirty files.
