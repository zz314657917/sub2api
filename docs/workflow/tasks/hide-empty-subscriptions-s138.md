# Task Contract: hide-empty-subscriptions-s138

## Task ID

`hide-empty-subscriptions-s138`

## Status

`approved`

## Role

Codex owns Planner, implementation, QA, and final evaluation. This is a direct small-fix Sprint with no worker.

## Goal

Hide the subscription panel at the top of the user `/usage` page after the active-subscription request completes with no subscriptions.

## Success Criteria

- Keep the existing loading indicator while the subscription request is pending.
- Remove the entire subscription panel from layout when the request resolves to an empty list.
- Keep active subscription cards and renewal routing unchanged when subscriptions exist.
- Keep the usage analytics and records below the panel visible in both states.

## Allowed Paths

- `frontend/src/components/user/UserSubscriptionsPanel.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `docs/workflow/tasks/hide-empty-subscriptions-s138.md`
- `docs/workflow/qa-reports/hide-empty-subscriptions-s138-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Backend subscription APIs, persistence, schema, migrations, routing, authentication, billing, administrator pages, deployment, containers, and unrelated UI.

## Constraints

- Reuse the component's existing `loading` and `subscriptions` state.
- Do not move subscription fetching into the parent Usage view.
- Do not delete shared empty-state translations or change error-toast behavior.

## Acceptance Commands

```powershell
corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run build
git diff --check
```

## Output

- Minimal conditional-rendering change and focused regression coverage.
- QA report with a first-line PASS, FAIL, or BLOCKED verdict.

## Stop Rules

- Stop if the change requires backend/API contract changes or alters subscription ownership/renewal behavior.
- Stop on changed paths outside the allowlist or if usage analytics disappear in the empty state.
