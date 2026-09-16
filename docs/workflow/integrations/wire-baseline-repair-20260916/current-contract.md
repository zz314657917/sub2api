---
status: approved
review_verdict: PASS
task_id: wire-current-integration
worker_model: gpt-5.6-terra
base_commit: ce316421cf4384a1cdcdd3e7de33f39b5bc37a7a
spec_ref: E:/codex-worktrees/sub2api/wire-baseline-repair/docs/workflow/tasks/wire-baseline-repair.md
---

## Task ID
wire-current-integration
## Role
Fresh independent Terra QA of repaired wiring on current committed main.
## Goal
Verify all wire-baseline-repair original and amended requirements on the newer
base that includes Pelican and first-response changes. Developer remains on
d69da80b9; controller transplants only reviewed source hunks (not old wire_gen).
## Success Criteria
- Read approved wire-baseline-repair contract, all review amendments and Developer
  report in E:/codex-worktrees/sub2api/wire-baseline-repair. Independently compare
  all seven source/test changed paths with original developer output and current
  base; preserve all current-main Pelican and first-response implementation.
- Repeat ALL original/amended QA commands and generated-diff gates on this tree.
  Two generations must match byte-for-byte. Inspect complete current-base diff,
  constructor arguments, Start order and cleanup; prove two baseline hooks preserved
  in providers and Pelican Start/handler/cleanup unchanged. New behavior drift stops.
- Additionally run service -run 'PelicanReview' and routes -run 'KeyRouteFirstResponse'
  to smoke the parallel changes. Existing tests may use local fake HTTP servers;
  no external provider, database, application or container startup allowed.
- PASS is local only; real DB/admin/provider and Composite P2/P3 remain unverified.
## Allowed Paths
- backend/cmd/server/wire_gen.go (automatic generation only)
- docs/workflow/qa-reports/wire-current-integration-qa.md
Source changes preinstalled by controller: seven non-generated paths from the
approved repair contract, no hand edits by QA. Document full pre/post inventory.
## Denied Paths
Manual business/test/generated changes, main/other worktrees writes, dependencies,
migrations, DB/Redis/external-provider, deployment, git commit/push/reset/clean.
## Constraints
Use E:/codex-worktrees/sub2api/wire-current-integration only. No tests or generator
until current contract independent PASS and Developer DONE. QA has no fix authority.
## Acceptance Commands
All original/amended wire-baseline-repair commands, with working directory set to
this tree/backend, including list/tests, two generations/hashes, server compile,
build and gofmt -l of all eight Go files; root diff-check, unmerged/inventory review.
Extra: go test ./internal/service -run PelicanReview -count=1
Extra: go test ./internal/server/routes -run KeyRouteFirstResponse -count=1
## Output
docs/workflow/qa-reports/wire-current-integration-qa.md; first line
### PASS/FAIL/BLOCKED: wire-current-integration. List actual selected tests/results,
source equivalence, generated full-diff audit, hashes and runtime limitations.
## Stop Rules
Missing independent review, Developer not DONE, further provider/hook drift,
out-of-scope output or failing gate -> stop, report, preserve evidence; no repair.
Terra unavailable -> BLOCKED, no alternate model.

## Current-base test fixture amendment
Independent QA discovered current-base provideCleanup now takes PelicanTestService
but cmd/server/wire_gen_test.go omits its positional argument. This is visible in
ce316421c before this patch: generated signature has pelicanTests after
scheduledTestRunner while test passes backupSvc immediately after scheduledTestRunner.
Planner authorizes only a Terra Developer one-line test-fixture synchronization
after independent review PASS: insert `nil, // pelicanTests` in that exact slot in
backend/cmd/server/wire_gen_test.go. No assertion, production signature, cleanup
behavior or other test may change. This path is added to Developer's allowlist;
QA still cannot edit it. Developer runs the full cmd/server test package and writes
docs/workflow/worker-results/wire-current-fixture-result.md. No startup of app/DB.
After Developer DONE, original QA repeats failed server compile and full cmd/server
tests, then all contract gates including two generations, build, nine-file gofmt
and diff audit. Preserve initial BLOCKED history. Any additional fixture failure
or production drift stops for Planner; do not weaken gates or silently exempt it.
