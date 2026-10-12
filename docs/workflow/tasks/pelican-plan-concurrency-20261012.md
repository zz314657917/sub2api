---
status: approved
review_verdict: PASS
task_id: pelican-plan-concurrency-20261012
worker_model: gpt-6.1-sol
base_commit: a62b9246397f1caf41f0011db353698164e848f7
spec_ref: docs/workflow/tasks/pelican-plan-concurrency-20261012.md
---

# Task ID: pelican-plan-concurrency-20261012

## Role
Controller Planner/Final Evaluator; independent Contract Reviewer, Developer and QA per Agent Matrix.

## Goal
Allow different Pelican plans to run concurrently instead of the current process-wide single slot.

## Success Criteria
- Bounded per-process capacity of 4 concurrent plan rounds, shared by manual and scheduled starts.
- Different plans can start while other plans are running; same-plan atomic repository Claim and generation guards remain unchanged.
- At capacity, reject manual starts without claiming/reserving quota; scheduled plans remain due and may start on subsequent ticks.
- Keep models within one plan sequential, account slot enforcement, quota, timeout, failure and feature-switch semantics intact.
- Release slots on settings/claim failure, completion and cancellation; Stop safely cancels and waits for all workers, preventing new starts.
- Deterministic channel-based tests prove simultaneous execution, capacity rejection, failure-slot reuse, same-plan claim conflict and shutdown; test manual/scheduled common path.
- No frontend behavior changes required: request busy state ends when accepted, running state is per-plan.

## Allowed Paths
- backend/internal/service/pelican_test_service.go
- backend/internal/service/pelican_test_runner_test.go
- backend/internal/service/pelican_plan_concurrency_test.go
- docs/workflow/worker-results/pelican-plan-concurrency-20261012.md
- docs/workflow/qa-reports/pelican-plan-concurrency-20261012.md

## Denied Paths
All other business files; migrations, payment/store, dependencies and production configuration.

## Constraints
Isolated tree E:/codex-worktrees/sub2api/pelican-plan-concurrency-20261012. Protect shared dirty work. No paid provider calls, live DB mutation, container update, commit or push. Use local fake repository/selector/executor for controlled concurrency evidence.

## Acceptance Commands
- From backend: go test -p 1 ./internal/service -run Pelican -count=1
- Repeat concurrency tests 20 times.
- go build -p 1 ./internal/service
- git diff --check
- If unrelated committed tests fail compilation, record exact blocker and use documented external overlay only for unrelated broken test files; never change them.

## Output
Developer report with changed paths and executed checks; independent QA report begins PASS/FAIL/BLOCKED, includes concurrency execution and shutdown evidence.

## Stop Rules
Stop for scope expansion, unsafe shutdown behavior, unsupported worker model or unexpected owner-file drift. Return actionable finding for Controller.
