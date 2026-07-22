# Task Contract: channel-monitor-timeline-overflow-s101

- Task ID: `channel-monitor-timeline-overflow-s101`
- Role: Planner / Generator / Evaluator
- Goal: Keep all 60 channel-monitor timeline bars and their labels inside the
  card content width at desktop and mobile sizes.
- Success Criteria:
  - The 60 timeline bars can shrink below their previous fixed 3px minimum so
    their combined width and gaps do not enlarge the timeline flex item.
  - All 60 bars remain rendered in oldest-to-newest order; no bar is clipped or
    removed to hide the overflow.
  - Timeline status colors, heights, tooltips, maintenance state, and refresh
    labels retain their existing behavior.
  - Desktop and mobile visual checks show no horizontal overflow from the
    timeline or `NOW` label.
- Allowed Paths:
  - `frontend/src/components/user/monitor/MonitorTimeline.vue`
  - `frontend/src/components/user/monitor/__tests__/MonitorTimeline.spec.ts`
  - `docs/workflow/tasks/channel-monitor-timeline-overflow-s101.md`
  - `docs/workflow/qa-reports/channel-monitor-timeline-overflow-s101-qa.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: channel-monitor APIs, backend, monitor data semantics, card/grid
  sizing, unrelated frontend components, migrations, deployment, containers,
  and VERSION.
- Constraints:
  - Fix the intrinsic minimum-width cause instead of clipping the timeline.
  - Preserve the 60-bar count and existing 2px gap unless visual verification
    proves that another in-scope spacing adjustment is necessary.
  - Keep the implementation limited to existing Vue and Tailwind patterns; do
    not add a dependency or JavaScript resize observer.
- Acceptance Commands:
  - `npm.cmd run test:run -- MonitorTimeline`
  - `npm.cmd run typecheck`
  - `npm.cmd run build`
  - Desktop and mobile browser screenshots of the channel-status monitor cards.
  - `git diff --check`, conflict-marker scan, exact allowlist audit, and
    unmerged-index check.
- Output: Scoped component diff, focused regression test, visual evidence, QA
  report, and final `PASS`, `FAIL`, or `BLOCKED` evidence.
- Stop Rules: Stop if the fix requires API/data changes, card/grid redesign,
  hiding bars with overflow clipping, or any path outside the allowlist.

## Contract Review

`PASS`: The screenshot and component source establish a deterministic intrinsic
width defect: 60 bars at a 3px minimum plus 59 2px gaps require 298px inside a
roughly 263px card content area. Allowing the existing flex items to shrink
removes the cause without changing monitor behavior or surrounding layout.
