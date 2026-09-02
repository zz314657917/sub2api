### PASS: upstream-v0200-group-pricing-layout-s290

# QA Report

## Task ID

`upstream-v0200-group-pricing-layout-s290`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0200-group-pricing-layout-s290.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`; the S290 implementation remains limited to the
  four approved frontend files. Existing dirty paths outside that set were
  preserved and are not attributed to this Sprint.
- denied paths touched: `no`

## Executed Checks

```text
frontend: pnpm.cmd exec vitest run src/views/admin/__tests__/groupsModelsListLayout.spec.ts -> PASS (1 file, 2 tests)
frontend: pnpm.cmd run typecheck -> PASS (exit 0)
frontend: pnpm.cmd run build -> PASS (exit 0; existing Browserslist/chunk-size warnings only)
root: git diff --check -- <four allowed frontend files> -> PASS
root: git diff --name-only --diff-filter=U -> PASS (no unmerged paths)
root: protected hash command -> PASS (0e467987fd7aec5fc451983bdb8f8216f97ba69c)
browser: task session/profile cleanup -> PASS (no browser session, owned Chrome/Node/cmd/cliDaemon process, or 5174 listener remains)
```

- Source inspection: `PASS`; both dialogs use `wide`; both headers wrap with a
  shrinkable description and non-shrinking add control; the six default prices
  use `pricing-default-grid`; and the shared `IntervalRow` retains its
  `pricing-interval-grid`, fields and emit behavior. The focused test asserts
  both responsive grid markers as well as both dialog/header controls.
- Runtime evidence reviewed: `PASS`; the local QA image health endpoint returned
  200. Task-owned Chrome using
  `E:\codex-runtime\pge\sub2api\s290\browser-smoke-20260902-retest\chrome-profile`
  opened `/admin/groups`, added pricing in both create and edit dialogs, then
  cancelled both forms without saving. At `1440x900` and `390x844`, document and
  dialog horizontal-overflow assertions were false and six default Token-price
  controls were present.
- Visual evidence inspected: `PASS`; create/edit desktop and mobile screenshots
  under
  `E:\codex-runtime\pge\sub2api\s290\browser-smoke-20260902-retest\artifacts`
  show the reachable six-control layouts without overlap or horizontal clipping.
- Cleanup evidence: `PASS`; post-close `playwright-cli list` is empty; read-only
  process inspection found no task profile/session-related Chrome, Node, cmd or
  `cliDaemon` process; no listener remains on port 5174. No credentials were
  read or emitted.

## Unverified Risks

- Enabled channel-pricing `IntervalRow` browser smoke remains deferred to a
  separate follow-up. The revised S290 contract intentionally keeps
  `hide-token-intervals=true` in both GroupsView callers, so this is outside the
  reachable group-dialog browser scope; its source-level responsive-grid
  sentinel remains covered by the focused test.
- Create and edit forms were intentionally cancelled per contract. No persisted
  pricing payload round trip was exercised, which avoids mutating shared data.

## Findings

- No S290 scope, layout, build, browser-cleanup or protected-dirty-path issue
  was found under the revised contract.

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- No S290 retest is required unless one of the four implementation files
  changes. The enabled channel-pricing `IntervalRow` browser smoke is deferred
  work and must not be inferred from this group-dialog QA.

## Knowledge Promotion

`none`
