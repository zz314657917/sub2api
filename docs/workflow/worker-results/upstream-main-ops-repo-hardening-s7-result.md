### DONE: upstream-main-ops-repo-hardening-s7

## Summary
- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-ops-repo-hardening-s7` and branch `codex/upstream-main-ops-repo-hardening-s7` from baseline `c3625ce46`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-ops-repo-hardening-s7.md`.
- Reviewed and processed all eight approved Ops/repository/account hardening candidates without directly merging `upstream/main`.
- Kept local implementations where current `main` already contains equivalent or stricter behavior.
- Ported the remaining useful test coverage for `account_count` sort attaching both total and active account counts.

## Candidate Results
- `ae6ee23e2`: `APPLIED_EQUIVALENT`. Current baseline already contains the Ops SLA classification helpers and stricter local extensions.
- `271aba1ab`: `APPLIED_EQUIVALENT`. Current baseline already marks API key IP restriction as `OpsClientBusinessLimitedReasonIPRestriction` and excludes it from SLA counting.
- `69305a609`: `APPLIED_EQUIVALENT`. Current baseline already includes the local client/business limit classifier and broader local feature gate/policy/quota coverage.
- `ab6510f1a`: `APPLIED_EQUIVALENT`. Cherry-pick was empty; `announcement.ListActive` already has `Limit(200)` and `group` account-count sorting already uses the optimized ID/count loading path.
- `5465003d0`: `PORTED_WITH_LOCAL_RESOLUTION` as commit `eb786804f`. Kept the current stricter rate-limited account availability test semantics and added the upstream `account_count` sort coverage in `group_repo_sort_integration_test.go`.
- `df2b02e61`: `APPLIED_EQUIVALENT`. Cherry-pick was empty; `groupAccountAvailableSQL`, temporary-limit SQL, `GetByID`, `GetAccountCount`, and `loadAccountCounts` already share the fixed count semantics.
- `49b415e33`: `APPLIED_EQUIVALENT`. Cherry-pick was empty; `refresh_token_reused` is already classified as non-retryable and tested.
- `202aab8e6`: `APPLIED_EQUIVALENT`. Cherry-pick was empty; error-status account updates already set `schedulable=false` and have integration coverage.

## Commits
- `ee3049b16` docs: add ops repo hardening s7 contract
- `eb786804f` cherry-pick of `5465003d0`: added `TestListWithAccountCountSort_AttachesActiveCount`

## Changed Files
- `backend/internal/repository/group_repo_sort_integration_test.go`
- `docs/workflow/tasks/upstream-main-ops-repo-hardening-s7.md`
- `docs/workflow/worker-results/upstream-main-ops-repo-hardening-s7-result.md`
- `docs/workflow/qa-reports/upstream-main-ops-repo-hardening-s7-qa.md`
- `docs/workflow/main-log.md`

## Verification
- `git status --short --branch`
- `git diff --check`
- `git diff --name-status c3625ce46..HEAD`
- `git diff --name-only c3625ce46..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"`
- `go test ./internal/handler ./internal/server/middleware ./internal/service -run "Ops|SLA|IP|Denied|Client|Token|Refresh|Scheduler|Account" -count=1`
- `go test ./internal/repository -run "Announcement|Group|Account|Available|Sort|Count" -count=1`
- `go test ./internal/repository ./internal/service ./internal/handler ./internal/server/middleware -count=1`

## Notes
- No candidate required forbidden paths, Ent schema, SQL migration, frontend changes, public API fields, or route/server wiring.
- The `5465003d0` commit message mentions three test cases, but the current baseline already had the two ListWithFilters coverage cases, including a stricter interpretation where temporarily limited accounts are excluded from current schedulable active count. The only net-new coverage is the account-count sort path.
