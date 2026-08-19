# Upstream v0.1.178 UI/Ops Polish S234

## Task ID

`upstream-v178-ui-polish-s234`

## Role

Controller/Generator: Codex. Independent QA is required from a separate
worktree before main integration.

## Goal

Port six small, locally applicable UI fixes from the `v0.1.178` release without
importing the surrounding release history: custom Ops error-detail time ranges,
localized header roles, dark native controls, dashboard cache-token breakdown,
announcement empty-state copy, and neutral empty-window SLA presentation.

## Upstream Sources

- `5e72deb7d` Ops error details custom time range
- `3bff4b64b` localized user role label
- `7d796f111` dark native form controls
- `a6d868f27` dashboard cache-token breakdown
- `35e8ba2a3` announcement empty-state copy
- `0d5e3ca9b` neutral SLA card for empty windows

All six commits are ancestors of `v0.1.178`; they are behaviorally adapted to
the current local frontend owners rather than cherry-picking the whole tag.

## Success Criteria

- Ops error details use `start_time`/`end_time` for a valid custom range and
  retain a safe one-hour fallback when the range is incomplete; parent custom
  range changes trigger reloads.
- The app header renders the localized role label through existing i18n keys.
- Native date controls follow the selected light/dark color scheme.
- User dashboard today/total token breakdowns include cache creation plus cache
  read tokens.
- Empty announcements say to create the first announcement in both locales.
- SLA is neutral when the selected window contains no SLA requests, including
  non-red exception styling.

## Frozen Base

`main@e850690ce` (existing user dirty/untracked work is protected).

## Allowed Paths

- `frontend/src/views/admin/ops/OpsDashboard.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDetailsModal.vue`
- `frontend/src/views/admin/ops/components/OpsDashboardHeader.vue`
- `frontend/src/components/layout/AppHeader.vue`
- `frontend/src/components/common/DateRangePicker.vue`
- `frontend/src/style.css`
- `frontend/src/components/user/dashboard/UserDashboardStats.vue`
- `frontend/src/views/admin/AnnouncementsView.vue`
- `frontend/src/i18n/locales/en/admin/resources.ts`
- `frontend/src/i18n/locales/zh/admin/resources.ts`
- `frontend/src/components/layout/__tests__/AppHeader.spec.ts`
- `frontend/src/components/common/__tests__/DateRangePicker.spec.ts`
- `frontend/src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts`
- `frontend/src/views/admin/__tests__/AnnouncementsView.spec.ts`
- `frontend/src/views/admin/ops/__tests__/OpsSmallFixes.spec.ts`
- `docs/workflow/results/upstream-v178-ui-polish-s234-result.md`

## Denied Paths

All backend files, migrations, dependencies, generated files, unrelated
frontend files, `EditAccountModal.vue` and its test, `knowledge/*`, `outputs/*`,
deployment/container files, provider traffic, databases, and remote refs.

## Constraints

- Preserve the existing frontend architecture, i18n key conventions, and the
  user-owned dirty/untracked state.
- Do not import channel-monitor, CN-provider, Codex, email, or model-sync
  features from the release tag.
- Keep custom-range query behavior explicit: valid custom values use start/end
  parameters; incomplete values must not send `time_range=custom`.

## Acceptance Commands

From `frontend/`:

```powershell
pnpm exec vitest run src/components/layout/__tests__/AppHeader.spec.ts src/components/common/__tests__/DateRangePicker.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/admin/__tests__/AnnouncementsView.spec.ts src/views/admin/ops/__tests__/OpsSmallFixes.spec.ts
pnpm run typecheck
pnpm run build
```

Also run `git diff --check`, exact allowlist, conflict/unmerged-index checks,
upstream ancestry checks for all six source commits, and protected-main status
checks before integration.

## Output

One business implementation commit plus
`docs/workflow/results/upstream-v178-ui-polish-s234-result.md` whose first line
is `### PASS`, `### FAIL`, or `### BLOCKED`.

## Stop Rules

- Stop if any source behavior requires backend/API/schema changes outside the
  listed owners.
- Stop if a denied path or user-owned dirty/untracked path changes.
- Stop if frontend dependencies or the required test/build commands are not
  available; report the exact blocker rather than declaring product failure.

## Status

`contract-approved`
