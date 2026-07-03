### DONE: upstream-main-v0142-account-repo-count-s38a

## Summary
- Ported upstream `fd004bdd8 fix(account-repo): Clone query before Count to prevent state pollution`.
- `accountRepository.ListWithFilters` now calls `q.Clone().Count(ctx)` before applying pagination and ordering to the list query.
- `AccountRepoSuite.TestListWithFilters` now asserts that single-page `pagination.Total` matches `len(accounts)`.
- Deferred `9f5b57fc9` and `03727ac36` remain outside this Sprint because they touch dirty billing, usage, subscription, frontend, config, or deploy surfaces.

## Changed Files
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `docs/workflow/tasks/upstream-main-v0142-account-repo-count-s38a.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `go test ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters|TestAccountRepoSuite/TestListWithFiltersGroupFilter" -count=1`
  - Result: command returned `ok ... [no tests to run]`; not counted as acceptance evidence because integration tests are behind `//go:build integration`.
- `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1`
  - Result: PASS.
- `gofmt -w backend/internal/repository/account_repo.go backend/internal/repository/account_repo_integration_test.go`
- `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1`
  - Result: PASS.

## Contract Compliance
- Allowed account repository files were changed.
- No billing cache, usage billing, subscription, frontend, Ent, migration, config, deploy, knowledge, or generated files were edited by S38a.
- No upstream merge, rebase, or broad cherry-pick was performed.

## Risks
- This Sprint did not port balance overdraft protection (`9f5b57fc9`) or subscription revoke soft-delete (`03727ac36`); both still need later clean-tree or dedicated-worktree contracts.
