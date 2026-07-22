# Task Contract: channel-monitor-timeline-status-colors-s102

- Task ID: `channel-monitor-timeline-status-colors-s102`
- Role: Planner / Generator / Evaluator
- Goal: Make channel-monitor timeline status colors easier to distinguish at a
  glance without changing status semantics or bar geometry.
- Success Criteria:
  - `operational` bars use a clear green color.
  - `degraded` bars use a clear orange color.
  - `failed` and `error` bars use red; empty placeholders remain gray.
  - Existing bar heights, order, count, tooltips, and responsive width behavior
    remain unchanged.
- Allowed Paths:
  - `frontend/src/components/user/monitor/MonitorTimeline.vue`
  - `frontend/src/components/user/monitor/__tests__/MonitorTimeline.spec.ts`
  - `docs/workflow/tasks/channel-monitor-timeline-status-colors-s102.md`
  - `docs/workflow/qa-reports/channel-monitor-timeline-status-colors-s102-qa.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: monitor APIs, backend, status calculation, card/grid sizing,
  unrelated frontend components, deployment, containers, and VERSION.
- Constraints: use existing Tailwind color utilities; keep the change limited
  to the existing `STATUS_COLOR` map and focused assertions.
- Acceptance Commands:
  - `npm.cmd run test:run -- MonitorTimeline`
  - `npm.cmd run typecheck`
  - `git diff --check`, conflict-marker scan, and unmerged-index check.

## Contract Review

`PASS`: This is a presentation-only change in the existing status-color map;
green/orange/red provide stronger visual separation while preserving all
runtime status behavior and responsive layout constraints.
