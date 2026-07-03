### DONE: upstream-main-v0143-user-model-stats-requested-s42

## Summary
- Ported upstream `e236bff1e fix: aggregate user model stats by requested model`.
- `usageLogRepository.GetUserModelStats` now reuses the existing requested-model aggregation helper.
- Added a focused sqlmock test that asserts the query groups by `COALESCE(NULLIF(TRIM(requested_model), ''), model)` and keeps the user/time filters.
- Broader usage billing, frontend, leaderboard, and service-layer behavior were not changed.

## Changed Files
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `gofmt -w backend/internal/repository/usage_log_repo.go backend/internal/repository/usage_log_repo_request_type_test.go`
- `go test ./internal/repository -run "TestUsageLogRepositoryGetUserModelStatsUsesRequestedModel|TestUsageLogRepositoryGetModelStatsWithFiltersRequestTypePriority|TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1`
  - Result: PASS.
- `git diff --check -- backend/internal/repository/usage_log_repo.go backend/internal/repository/usage_log_repo_request_type_test.go docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md docs/workflow/status.md docs/workflow/main-log.md`
  - Result: PASS, with LF/CRLF warnings for workflow docs only.
- staged denied-path audit
  - Result: `NO_DENIED_PATHS` because no files were staged.

## Contract Compliance
- Changed only allowed usage-log repository files and workflow artifacts.
- Did not edit gateway, billing service, usage billing, frontend, Ent, migrations, deploy, knowledge, dependencies, or generated files.
- Did not merge/rebase `v0.1.143` or cherry-pick broader release content.

## Risks
- Full repository tests were not run because S42 is a narrow repository SQL aggregation change and the worktree contains unrelated dirty files.
- OpenAI plan type, subscription expiration, Codex session import identity, gateway compact/keepalive/Bearer, and count_tokens latest scope remain deferred.
