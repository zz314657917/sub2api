### DONE: sub2api-canvas-run-control

## changed files

- `backend/internal/repository/canvas_repo_test.go`
- `backend/internal/service/canvas_service_test.go`
- `frontend/src/api/__tests__/canvas.spec.ts`
- `frontend/src/api/canvas.ts`
- `docs/workflow/worker-results/sub2api-canvas-run-control-result.md`

## commands run

- `gofmt -w backend/internal/service/canvas_service_test.go backend/internal/repository/canvas_repo_test.go`
- `go test ./internal/service ./internal/handler ./internal/repository -run "Canvas" -count=1` from `backend/`
- `npm.cmd run test:run -- canvas` from `frontend/` (first attempt failed because `node_modules` was absent and `vitest` was not installed)
- `corepack.cmd pnpm install --frozen-lockfile` from `frontend/`
- `npm.cmd run test:run -- canvas` from `frontend/`
- `git diff --check`
- `git status --short --untracked-files=all`

## key output

- Go: `ok github.com/Wei-Shaw/sub2api/internal/service`, `ok github.com/Wei-Shaw/sub2api/internal/handler`, `ok github.com/Wei-Shaw/sub2api/internal/repository`.
- Frontend after installing locked dependencies: `src/api/__tests__/canvas.spec.ts (4 tests)` passed and `src/views/user/__tests__/CanvasView.spec.ts (6 tests)` passed; `Test Files 2 passed (2)`, `Tests 10 passed (10)`.
- `corepack.cmd pnpm install --frozen-lockfile` completed with pnpm v10.33.4. It reported ignored build scripts for `esbuild` and `vue-demi`; the targeted Vitest command still passed.
- `git diff --check` produced no output.

## risks

- Canceling a Canvas run still only updates the Canvas run itself. It does not cascade into ImageCreator tasks, matching the contract and current product spec.
- The repository test locks the cancel SQL shape with `sqlmock`; if the query is refactored without behavior changes, the test may need a small expectation update.

## contract compliance

- Added `cancelCanvasRun(id: string): Promise<CanvasRun>` in `frontend/src/api/canvas.ts`.
- `cancelCanvasRun` calls `POST /user/canvas-runs/:id/cancel` and reuses the existing backend run mapper.
- Added `CanvasRun.canceled_at` to the frontend type and mapped `BackendCanvasRun.canceled_at`.
- Kept existing status normalization: `pending` maps to `queued`; `canceled` remains `canceled`.
- Added frontend API tests for cancel helper and `canceled_at` mapping.
- Added service tests for current-user isolation, pending/running cancel, canceled idempotency, and succeeded/failed terminal conflict behavior.
- Added repository test verifying cancel uses `user_id` and only pending/running statuses in the SQL update.
- Did not modify `frontend/src/views/user/CanvasView.vue`, `knowledge/**`, `docs/workflow/status.md`, `docs/workflow/main-log.md`, migrations, production config, RBAC, or ImageCreator task execution logic.

## knowledge_candidates

- None.
