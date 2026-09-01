### DONE: upstream-v0184-group-limit-partial-s279

# Worker Result

## Task ID

`upstream-v0184-group-limit-partial-s279`

## Status

`done`

## Summary

- Handler now distinguishes omitted limit fields (`nil`), explicit JSON `null` (negative unlimited sentinel), and numeric values (including `0`).
- Ordinary group updates apply only provided daily, weekly, or monthly limits; omitted fields keep their existing values and `normalizeLimit` retains the local `<=0` unlimited behavior.
- `room_managed` groups still clear all three group-level limits regardless of whether the input omits or supplies limits.

## Changed Files

- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/admin/group_handler_limit_test.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_group_limit_partial_test.go`
- `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md`

## Commands Run

```text
cd backend && gofmt -w internal/handler/admin/group_handler.go internal/handler/admin/group_handler_limit_test.go internal/service/admin_service_group_limit_partial_test.go -> passed
cd backend && go test ./internal/handler/admin -run '^TestUpdateGroupRequestLimitFieldsTriState$' -count=10 -> passed
cd backend && go test ./internal/service -run '^TestAdminService_UpdateGroup_(LimitFieldsPartialUpdate|RoomManagedLimitInvariant)$' -count=10 -> passed
cd backend && go test ./internal/handler/admin -> passed
cd backend && go test ./internal/service -> passed (exit code 0)
cd backend && go test ./cmd/server -run '^$' -> passed
gofmt -d on all four S279 Go files -> no output
git diff --check -- <four S279 Go files> -> passed
git diff --name-only --diff-filter=U -> no unmerged paths
git diff --no-index -- <admin_service baseline> backend/internal/service/admin_service.go -> only UpdateGroup limit block
git diff --no-index -- <group_handler baseline> backend/internal/handler/admin/group_handler.go -> only null sentinel
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/handler/admin  (tri-state x10)
ok github.com/Wei-Shaw/sub2api/internal/service        (partial update and room-managed invariant x10)
ok github.com/Wei-Shaw/sub2api/internal/handler/admin
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/cmd/server [no tests to run]
```

## Baseline Preservation

- The external `admin_service.go` snapshot SHA-256 matched its approved value before implementation. The final no-index comparison shows exactly one new hunk: the `UpdateGroup` daily/weekly/monthly limit block.
- The pre-existing Pixel Cafe quota-reset hunk, including its import, remains absent from the no-index comparison and was not edited by S279.
- The handler snapshot comparison shows only `0.0` changed to the negative unlimited sentinel. Protected dirty paths and `outputs/` were not modified by this worker.

## Risks

- Default-tag affected packages and server compilation passed. No provider, database, container, deployment, shared-data, commit, or push action was performed.
- Browser/API runtime smoke is intentionally outside this contract; independent QA must review the exact diff and execute its own evidence run.

## Knowledge Candidates

- A dirty monolithic service owner can be protected reliably with a repository-external baseline plus a final `git diff --no-index` target-block check.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`
