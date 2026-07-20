### PASS: usage-model-reasoning-effort-s82-generator

## Changed Files

- `frontend/src/utils/modelDisplay.ts`
- `frontend/src/utils/__tests__/modelDisplay.spec.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`

## Implemented Behavior

- Added `displayModelWithReasoningEffort`, reusing the existing model-label
  sanitization and `formatReasoningEffort` normalization.
- User usage rows and their model detail show `model (Effort)` only when the
  effort is meaningful; empty, `none`, and `minimal` values keep the model
  unchanged.
- Admin requested/upstream rows and mapping chains annotate only the requested
  model, preserving upstream model labels and brand stripping.
- Kept the existing standalone reasoning-effort column and all API/backend
  behavior unchanged.

## Commands Run

- Targeted Vitest: 3 files, 33 tests PASS.
- `npm.cmd run typecheck`: PASS.
- `npm.cmd run build`: PASS; Vite emitted only existing warning classes
  (stale browserslist data, dynamic-import topology, and large chunks).
- `git diff --check`: PASS.

## Risks / Deferred Checks

- No authenticated browser smoke was run; the rendering branches are covered
  by the user/admin component tests.
- No backend, database, Redis, deployment, or container operation was needed
  or performed.
