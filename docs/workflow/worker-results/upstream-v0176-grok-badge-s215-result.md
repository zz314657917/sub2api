### DONE: s215-grok-badge

## Scope

- Base: `9b08c8126435cc1839dd71f3351fb0b07d73d2a9`
- Changed only the S215 Slice B allowlist: Grok account tier resolution, Grok
  snapshot refresh-key comparison, utility tests, and a mounted AccountsView
  regression test.

## Implementation

- `getAccountPlanType` preserves non-Grok `credentials.plan_type` behavior and
  applies the approved Grok credential, billing, canonical usage, legacy,
  extra, and parent-plan precedence.
- `buildGrokUsageRefreshKey` deterministically normalizes object key order,
  preserves array order, and excludes legacy snapshot churn when a canonical
  billing or usage tier is available.
- AccountsView retains `share_display_tier` override and replaces rows when
  canonical Grok snapshot data changes.

## Verification

- Discovery: `npx.cmd vitest list src/utils/__tests__/accountUsageRefresh.spec.ts src/views/admin/__tests__/AccountsView.grokBadge.spec.ts` found both required names:
  `buildGrokUsageRefreshKey > uses canonical tier before legacy snapshot` and
  `admin AccountsView Grok badge > replaces a Grok row after canonical snapshot changes`.
- Repeated focused Vitest: three consecutive runs passed, each `2 files / 8 tests`.
- `npx.cmd vue-tsc --noEmit` passed.
- `npx.cmd eslint src/utils/accountUsageRefresh.ts src/utils/__tests__/accountUsageRefresh.spec.ts src/views/admin/AccountsView.vue src/views/admin/__tests__/AccountsView.grokBadge.spec.ts` passed.
- `git diff --check`, unmerged-index check, allowed-path diff check, and conflict-marker check passed.

## Risks

- No live provider or browser runtime validation was performed; this slice is
  limited to the admin list's existing incremental API response path.
- The worktree had no local frontend dependencies. A local junction to the
  primary checkout's existing `frontend/node_modules` was created solely for
  validation and is untracked/ignored; no dependency or lockfile changed.

## Knowledge Candidates

- None.
