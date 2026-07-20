### PASS: upstream-main-compat-s83

## Findings

- No blocking finding remains. The local implementation ports upstream
  `337960302` while adapting the user-side owner from the upstream view path to
  the existing `UserSubscriptionsPanel` component.
- The shared helper delegates to the existing locale-aware `formatDate`, so
  timezone and locale behavior remain unchanged; only seconds are omitted.
- Admin table/detail and user panel expiry text now show minute precision.
  Expired/today/days-remaining labels and calculations are unchanged.
- No UsageView/model-display, API, backend, billing, or subscription payload
  path changed.

## Executed Checks

- Generator focused Vitest: PASS, 1 file / 2 tests.
- Generator `npm.cmd run typecheck`: PASS.
- Generator `npm.cmd run build`: PASS, 1080 modules transformed.
- Fresh Evaluator focused Vitest: PASS, 1 file / 2 tests.
- Fresh Evaluator `npm.cmd run typecheck`: PASS.
- Fresh Evaluator `npm.cmd run build`: PASS, 1080 modules transformed.
- Local owner/static checks: PASS; no `formatDateOnly` remains in the two
  expiry owners and remaining-days calls remain present.
- Business diff, unmerged-index, conflict-marker, and `git diff --check`
  gates: PASS.

## Unverified Risks

- No authenticated browser smoke was run; the change is a static formatting
  helper and two subscription display call sites covered by compilation and
  focused tests.
- Existing Vite dynamic-import and large-chunk warnings remain non-blocking.
- Primary checkout workflow files and Usage S82 changes are external dirty
  work and were not merged or modified by S83.

## Recommendation

`PASS` — create the scoped S83 commit after the exact ten-path tracking,
conflict-marker, cached-diff, and protected-hash gates pass. Keep the branch
isolated until the primary S82 workflow conflict is reconciled.
