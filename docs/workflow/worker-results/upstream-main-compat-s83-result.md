### PASS: upstream-main-compat-s83-generator

## Changed Files

- `frontend/src/utils/format.ts`
- `frontend/src/utils/__tests__/formatDateTimeToMinute.spec.ts`
- `frontend/src/views/admin/SubscriptionsView.vue`
- `frontend/src/components/user/UserSubscriptionsPanel.vue`

## Implemented Behavior

- Added `formatDateTimeToMinute` by reusing the existing locale-aware
  `formatDate` helper with year/month/day/hour/minute and no seconds.
- Admin subscription expiry table/detail displays now show minute precision.
- The local user subscription owner, `UserSubscriptionsPanel`, now shows the
  same minute precision while retaining expired/today/days-remaining labels.
- Invalid dates still return an empty string; no API, status, quota, timezone,
  or remaining-days logic changed.
- Added focused valid-date and invalid-date formatter regressions.

## Commands Run

- Focused Vitest: PASS, 1 file / 2 tests.
- `npm.cmd run typecheck`: PASS.
- `npm.cmd run build`: PASS, 1080 modules transformed.
- Static owner and remaining-days preservation checks: PASS.
- Business diff, `git diff --check`, unmerged-index, and conflict-marker
  scans: PASS.

## Risks / Deferred Checks

- No authenticated browser smoke was run; the change is limited to formatting
  and static subscription display call sites.
- Existing Vite dynamic-import and large-chunk warnings remain non-blocking.
- The primary checkout has an unrelated Usage S82 dirty change and workflow
  conflict; S83 remains isolated until that work is reconciled.
