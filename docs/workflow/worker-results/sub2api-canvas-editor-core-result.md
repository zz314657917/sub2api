### DONE: sub2api-canvas-editor-core

## changed files
- `frontend/src/views/user/CanvasView.vue`
- `frontend/src/views/user/__tests__/CanvasView.spec.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `docs/workflow/worker-results/sub2api-canvas-editor-core-result.md`

## commands run
- `corepack pnpm install --frozen-lockfile`
- `npm.cmd run test:run -- CanvasView canvas`
- `npm.cmd run lint:check`
- `git diff --check`

## key test output
- `npm.cmd run test:run -- CanvasView canvas`
  - `src/api/__tests__/canvas.spec.ts`: 3 tests passed
  - `src/views/user/__tests__/CanvasView.spec.ts`: 9 tests passed
  - `Test Files 2 passed (2), Tests 12 passed (12)`
- `npm.cmd run lint:check`
  - exited with code 0
- `git diff --check`
  - exited with code 0

## risks
- Canvas interactions are mouse-based per contract. Touch/pointer-specific gestures were not added.
- `corepack pnpm install --frozen-lockfile` was required because the worktree had no `node_modules` and `vitest` was initially missing; no lockfile changes were produced.

## contract compliance
- Implemented node drag in `CanvasView.vue`; saved document payload now carries updated `nodes[].x/y`.
- Implemented edge creation from selected source node to clicked target node, duplicate-edge suppression, selected-edge deletion, and node deletion cleanup for related edges.
- Implemented canvas viewport zoom controls, wheel zoom, stage panning, fit-view, and viewport normalization so `document.viewport` is included in save payload.
- Preserved existing node config editing, run queue button, ImageCreator task polling, and result rendering behavior; existing tests still pass.
- Added tests for node drag save, create/delete edge behavior, and viewport save.
- Did not modify backend, `frontend/src/api/canvas.ts`, `frontend/src/api/imageCreator.ts`, templates, advanced image editing, knowledge files, or workflow status/log files.

## knowledge_candidates
- None.
