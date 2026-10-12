---
status: approved
review_verdict: PASS
task_id: pelican-multi-model-20261007
worker_model: gpt-6.1-sol
base_commit: 10e542353d3a6ddb8c26996e6d35cf6122278329
spec_ref: docs/workflow/tasks/pelican-multi-model-20261007.md
---

# Task ID: pelican-multi-model-20261007

## Role
Controller Planner/Final Evaluator; independent gpt-6.1-sol Contract Reviewer, Developer and QA.

## Goal
Implement user-approved multi-model plans and a group-based plaza with vertically stacked model artworks, prominent model/vendor and generation-time metadata.

## Success Criteria
- Add additive model_ids storage (ordered nonempty list, 1–10 unique trimmed IDs). Backfill legacy model_id; retain model_id as first-model compatibility alias. Omitted model_ids accepts legacy model_id; explicit empty/invalid list rejects. Validate every model against existing group eligibility. Shared effort/timeout/interval/limits, existing global prompt unchanged.
- Each claimed round runs selected models sequentially, one account selection/attempt per model; failures/skips do not prevent later models, cancellation stops. Per-model timeout and slot release. Lease and overall deadline cover all selected models plus bounded overhead; generation guards prevent stale writes and concurrent duplicate rounds.
- Atomic daily quota counts each reserved model attempt, including failures; eligibility skips do not count. Exhaustion prevents more upstream calls. estimated_calls_per_run equals model count. Consecutive failure remains round-based: any success resets, attempted round without successes increments once; all skipped does not increment. Preserve existing pause/resume semantics.
- Latest result, success fallback, history count, count retention and protected latest success are scoped by plan/group/model. Model removal retains historical rows, excludes removed models from current gallery; old plans are not merged/deleted. Optional history model_id filters server-side before limiting, preserving omitted legacy behavior and authorization.
- Gallery pagination counts groups, never splits one group across pages. Keep flat items payload if simplest; return all active plan/model entries for paged groups and total groups. Stable ordering newest group first then group ID, stable child ordering. Same-model entries from distinct plans remain distinct with plan identification. Authorization/filtering applied before group pagination; preserve account filter semantics.
- Plans support multi-select, edit roundtrip, group-switch stale selection cleanup, selected count and per-round/daily estimate. Existing single-model plans remain editable.
- One masonry card per group, models vertically stacked, clear separators. Reuse ModelIcon based on model ID and existing clock icon; bold readable model names, generation time plus complete timestamp tooltip, duration clearly separate. Unknown models use fallback; accessible light/dark/390px presentation.
- Preserve natural preview sizing, 580px cap, history 180px floor, full-width responsive masonry, replay, enlargement, focus restoration, safe sandbox and cache isolation. No changes to sanitizer/security rules.

## Allowed Paths
- backend/internal/service/pelican_test_service.go and pelican-specific *_test.go
- backend/internal/repository/pelican_test_repo.go and pelican-specific *_test.go
- backend/internal/handler/pelican_test_handler.go and pelican-specific *_test.go
- One new unused-number backend/migrations/*_pelican_multi_model.sql and its focused test
- frontend/src/api/pelicanTests.ts
- frontend/src/views/user/PelicanTestsView.vue
- frontend/src/components/pelican/PelicanPlansPanel.vue; new task-specific presentation helpers; existing pelican tests
- Task-specific docs/workflow/{tasks,contract-reviews,worker-results,qa-reports}/pelican-multi-model-20261007* and outputs/pelican-multi-model-20261007/**

## Denied Paths
All unrelated dirty files, provider adapters, global settings, payments/store, sanitizer/sandbox implementation, dependency lockfiles, secrets and production data. No shared container replacement, paid provider calls, commit/push without new instruction.

## Constraints
Use isolated worktree at E:/codex-worktrees/sub2api/pelican-multi-model-20261007 from base commit. Shared tree Pelican owners are clean; do not copy unrelated dirty files. No unrelated fixture repairs. Database migration is additive source only; runtime DB tests use task-owned disposable database, never live data. Repository matrix overrides generic skill model defaults. Reviewer may inspect shared sources read-only; Developer starts only after independent PASS and approved frontmatter.

## Acceptance Commands
- Backend go build ./internal/service ./internal/repository ./internal/handler and focused Pelican tests (discover build tags and exact commands; record baseline blockers separately).
- Frontend targeted Pelican vitest, vue-tsc/typecheck and production build with repo scripts; scoped lint/diff check.
- Isolated PostgreSQL migration/CRUD tests, model-separated history/retention, group paging/authorization, quota exhaustion and stale generation/claim concurrency. Fake executor for first model failure followed by success; prove each timeout independent and lease duration sufficient.
- Real browser fixture at desktop and 390px: multi-model create/edit, same group stacked results, separate model histories, model logos/time, long/short heights, replay/enlarge/focus, dark mode and no overflow. Follow repository task-owned profile/PID and cleanup rules. No paid calls needed.

## Output
Developer result with changed files, commands/results, limitations. Independent QA report first line PASS/FAIL/BLOCKED with evidence and contract compliance. Controller integrates only reviewed allowlist after QA.

## Stop Rules
Unavailable prescribed model => BLOCKED, no silent substitution. Stop on unrelated dirty owner, unsafe migration, scope expansion or ambiguous runtime mutation. Do not claim deploy or real provider validation from builds/fixtures.
