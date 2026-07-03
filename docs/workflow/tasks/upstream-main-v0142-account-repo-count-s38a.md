---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 01:19 +08:00
---

# Task Contract

## Task ID
upstream-main-v0142-account-repo-count-s38a

## Role
Codex acts as Planner and Final Evaluator. Implementation may be done by Codex directly or a Developer Worker only after this contract is approved. QA may be run by Codex directly after implementation.

## Goal
Port the narrow account repository Count-query pollution fix from upstream `v0.1.142` commit `fd004bdd8`, while explicitly deferring the broader S38 billing/subscription commits that overlap with the current dirty usage-billing and welfare worktree.

## Success Criteria
- Port or locally adapt `fd004bdd8 fix(account-repo): Clone query before Count to prevent state pollution`.
- `accountRepository.ListWithFilters` must not reuse the same Ent query builder for `Count()` and the subsequent list query when interceptors can append predicates.
- Add or update a repository regression assertion proving that when the filtered result set fits on one page, `pagination.Total` matches `len(accounts)`.
- Do not touch billing cache, usage billing, gateway service, subscription revoke flow, frontend, migrations, Ent generated files, deploy files, knowledge files, or unrelated dirty files.
- Explicitly record that `9f5b57fc9` and `03727ac36` remain deferred because they touch dirty billing/subscription/usage/frontend surfaces and need a later clean-tree or dedicated-worktree contract.

## Context
- Repo: `F:/mcplugins/sub2api`
- Base planning Sprint: `upstream-main-v0142-merge-plan-s35`
- Previous completed Sprint: `upstream-main-v0142-openai-codex-gateway-s37`
- Upstream release: `v0.1.142` / `60da9ba17`
- Original S38 candidates: `9f5b57fc9`, `03727ac36`, `fd004bdd8`.
- Current precheck:
  - `backend/internal/repository/account_repo.go` is clean in the current worktree.
  - `backend/internal/repository/account_repo_integration_test.go` is clean in the current worktree.
  - `fd004bdd8` touches only those two account repository files.
  - `9f5b57fc9` touches `usage_billing_repo.go`, `billing_cache_service.go`, `gateway_service.go`, `usage_billing.go`, config, and deploy example paths that currently overlap dirty or broader-risk areas.
  - `03727ac36` touches subscription repository/service, billing cache, admin routes, DTO/types, frontend subscription API/types, and integration tests, so it is too broad for the current dirty tree.

## Allowed Paths
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `docs/workflow/tasks/upstream-main-v0142-account-repo-count-s38a.md`
- `docs/workflow/worker-results/upstream-main-v0142-account-repo-count-s38a-result.md`
- `docs/workflow/qa-reports/upstream-main-v0142-account-repo-count-s38a-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/config/**`
- `backend/internal/repository/usage_billing_repo.go`
- `backend/internal/repository/user_subscription_repo.go`
- `backend/internal/repository/billing_cache.go`
- `backend/internal/repository/welfare_*`
- `backend/internal/service/billing_cache_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/usage_billing.go`
- `backend/internal/service/subscription_service.go`
- `backend/internal/service/user_subscription.go`
- `backend/internal/service/payment_*`
- `backend/internal/service/welfare_*`
- `backend/internal/handler/admin/subscription_handler.go`
- `backend/internal/handler/dto/**`
- `backend/internal/server/routes/admin.go`
- `frontend/**`
- `knowledge/**`
- `deploy/**`
- `assets/**`
- `README*`
- `.github/**`
- `backend/.codex-build/**`
- Any unlisted dirty file.

## Constraints
- Do not merge/rebase `v0.1.142` or `upstream/main`.
- Do not cherry-pick the whole S38 bundle.
- Keep this Sprint repository-only and account-list-only.
- Do not try to resolve the current welfare voucher, usage billing, payment/refund, or subscription dirty tree in this Sprint.
- If the account repo files become dirty before implementation, stop and reassess ownership before editing.
- Do not stage existing dirty files outside allowed paths.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1

cd F:/mcplugins/sub2api
git diff --check -- backend/internal/repository/account_repo.go backend/internal/repository/account_repo_integration_test.go docs/workflow/tasks/upstream-main-v0142-account-repo-count-s38a.md docs/workflow/worker-results/upstream-main-v0142-account-repo-count-s38a-result.md docs/workflow/qa-reports/upstream-main-v0142-account-repo-count-s38a-qa.md docs/workflow/status.md docs/workflow/main-log.md
git diff --cached --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/|backend/internal/config/|backend/internal/repository/usage_billing_repo.go|backend/internal/repository/user_subscription_repo.go|backend/internal/repository/billing_cache.go|backend/internal/repository/welfare_|backend/internal/service/billing_cache_service.go|backend/internal/service/gateway_service.go|backend/internal/service/usage_billing.go|backend/internal/service/subscription_service.go|backend/internal/service/user_subscription.go|backend/internal/service/payment_|backend/internal/service/welfare_|backend/internal/handler/admin/subscription_handler.go|backend/internal/handler/dto/|backend/internal/server/routes/admin.go|frontend/|knowledge/|deploy/|assets/|README|README_|\.github/)" || echo NO_DENIED_PATHS
```

## Output
- Code diff in allowed account repository paths only.
- Worker result: `docs/workflow/worker-results/upstream-main-v0142-account-repo-count-s38a-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-v0142-account-repo-count-s38a-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`

## Stop Rules
- Stop if implementation requires billing cache, usage billing, subscription, frontend, Ent, migration, config, deploy, or generated code changes.
- Stop if the targeted repository tests require a database or environment that is unavailable; report `BLOCKED` with the exact command/error instead of broadening the task.
- Stop if current dirty files outside the allowed account repo paths must be modified for tests to compile.
- Stop if `git status --short` shows pre-existing dirty changes in either allowed account repo file before implementation.

## Budget
- worker_mode: `local-codex-or-deepseek-v4-pro`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0.05`

## Review Result
- Reviewed at: 2026-07-03 01:19 +08:00.
- Verdict: approved.
- Reason: required P/G/E contract fields are present; allowed paths are limited to the clean account repository files and workflow artifacts; denied paths explicitly protect dirty billing, usage, subscription, frontend, migration, Ent, deploy, knowledge, and unrelated files.

## Acceptance Update
- Updated at: 2026-07-03 01:25 +08:00.
- Reason: `account_repo_integration_test.go` is guarded by `//go:build integration`; the original command without `-tags=integration` returned `[no tests to run]`, so the executable acceptance command is the integration-tagged command above.
