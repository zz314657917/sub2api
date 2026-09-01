---
phase: done
current_sprint: upstream-v0184-group-limit-partial-s279
total_sprints: 279
pending_action: S279 is independently QA PASS and ready for exact local integration. Preserve the unstaged Pixel Cafe hunk and all other dirty paths/outputs; no further v0.1.179-v0.1.184 candidate is approved. Do not push unless explicitly requested.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-09-01 12:28 +08:00
---

# Upstream v0.1.184 Group Limit Partial Updates S279

- `DONE / scope`: upstream `9f1effd71` is adapted to the local topology.
  Handler JSON limit fields preserve omitted/null/number tri-state behavior;
  ordinary groups update only provided limits, while `room_managed` groups
  continue to force all group-level limits to unlimited.
- `PASS / dirty-owner-preservation`: controller and independent QA compared
  `admin_service.go` and `group_handler.go` against repository-external SHA-256
  snapshots. The service differs only in the target `UpdateGroup` limit block;
  the pre-existing Pixel Cafe quota-reset hunk remains unstaged and unchanged.
- `PASS / verification`: focused handler/service suites passed x10, complete
  handler/service contract commands and the staged-snapshot service run
  passed, and server compile, gofmt, diff/conflict and exact-scope checks pass.
- `PASS / local-integration`: only the handler sentinel, two focused tests,
  exact service hunk and workflow evidence belong to S279. No provider,
  database, container, deployment, shared-data or push action occurred.
- Contract: `docs/workflow/tasks/upstream-v0184-group-limit-partial-s279.md`.
- Worker: `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md`.
- QA: `docs/workflow/qa-reports/upstream-v0184-group-limit-partial-s279-qa.md`.

- Earlier workflow status was archived by pge-compact at 20260901T043512271Z.
