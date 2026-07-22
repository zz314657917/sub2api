# Task Contract: key-route-editor-polish-s98

- Task ID: `key-route-editor-polish-s98`
- Role: Planner / Generator / Evaluator
- Goal: Keep the API-key multi-group route editor compact when many groups exist, and prevent users from selecting the same group twice.
- Success Criteria:
  - Desktop route header stays on one row with concise, current guidance.
  - The route list keeps its existing bounded internal scrolling.
  - Each route selector shows only unused groups plus that row's current group.
  - Add Route is disabled when a route is incomplete or every available group is already selected.
  - Priority-only payload and backend behavior remain unchanged.
- Allowed Paths:
  - `frontend/src/views/user/KeysView.vue`
  - `frontend/src/views/user/__tests__/KeysView.spec.ts`
  - `frontend/src/views/user/__tests__/KeysView.createQuery.spec.ts`
  - `frontend/src/i18n/locales/en/keys.ts`
  - `frontend/src/i18n/locales/zh/keys.ts`
- Denied Paths: backend, migrations, deployment, containers, billing, group administration, unrelated frontend views.
- Constraints: Preserve drag ordering, enabled state, fixed compatibility defaults, and duplicate-submit validation.
- Acceptance Commands:
  - Focused KeysView Vitest files.
  - Frontend typecheck.
  - Frontend production build.
  - `git diff --check` and unmerged-index check.
  - Desktop and narrow viewport browser smoke.
- Output: Scoped source diff, test evidence, and visual verification result.
- Stop Rules: Stop on backend contract changes, unavailable group identity, or unrelated dirty-file conflicts.

## Contract Review

`PASS`: The change is a frontend-only refinement built on the approved S92 priority-only route editor. The allowed paths and acceptance checks cover the requested behavior without changing routing semantics.
