---
status: approved
review_verdict: PASS
task_id: upstream-baseline-20261003
worker_model: gpt-6.1-sol
base_commit: 9bb59ba3232b5f53fb20114a19424cc1eaadedb1
spec_ref: docs/workflow/tasks/upstream-baseline-20261003.md
---

## Task ID
upstream-baseline-20261003

## Role
Independent Developer; separate Contract Reviewer and QA per project Agent Matrix.

## Goal
Complete the existing GroupID pricing integration so the clean main snapshot compiles and GroupID-only requests retain channel pricing. This is a prerequisite to the user-authorized selective upstream integration.

## Success Criteria
- TokenCostRequest accepts GroupID; explicit ID takes precedence over Group.ID; both absent preserves catalog fallback.
- Resolver-based billing accepts GroupID-only input; existing resolved pricing and legacy long-context precedence remain intact. The minimal resolver condition may be updated to accept `GroupID != nil` when `Resolved` is absent.
- Tests prove actual channel cost for a GroupID-only request, explicit/fallback/absent ID and existing long-context behavior.
- Independent QA and clean snapshot build pass before integration. Shared user file bytes remain protected; only exact reviewed task hunks may be integrated.

## Contract Review

PASS. The clean main snapshot is missing GroupID support referenced by committed gateway code. The implementation is limited to the two existing shared-worktree behaviors plus the GroupID-only resolver condition and focused tests. The reviewer requires explicit-ID precedence, Group.ID fallback, nil fallback, actual channel pricing, resolver GroupID-only coverage, and no absorption of unrelated dirty changes.

## Allowed Paths
- `backend/internal/service/billing_token_cost_request.go`
- `backend/internal/service/billing_token_cost_group_id_test.go`
- `docs/workflow/tasks/upstream-baseline-20261003.md`
- `docs/workflow/contract-reviews/upstream-baseline-20261003-review.md`
- `docs/workflow/worker-results/upstream-baseline-20261003-result.md`
- `docs/workflow/qa-reports/upstream-baseline-20261003-qa.md`

## Denied Paths
- `backend/migrations/**`
- `frontend/**`
- `outputs/**`
- `knowledge/**`
- Shared workspace writes, store/payment/Pelican, credentials, deployments and provider traffic.

## Constraints
Use only E:/codex-worktrees/sub2api/upstream-integration-20261003. No wholesale merges/cherry-picks, reset --hard or stash. No commit/push by worker. Existing shared diff is evidence only, not authorization to absorb unrelated work. Controller owns integration/publication following review.

## Acceptance Commands
Run in backend:
`go test -tags unit ./internal/service -run 'Test.*(TokenCost|GroupID|LegacyLongContext)' -count=1`
`go build ./...`
At root: `git diff --check`; exact-path audit.

## Output
Worker report begins `### DONE: upstream-baseline-20261003` or `### BLOCKED: upstream-baseline-20261003`; list changes, commands/exits, evidence and remaining limits. Independent QA report uses PASS/FAIL/BLOCKED.

## Stop Rules
Stop for any new baseline failure requiring denied files, unexplained concurrent writes or scope expansion. Do not mask missing tests or claim real provider/production runtime acceptance.
