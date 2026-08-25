# Pixel Cafe Workstation Layout Editor S254

## Task ID

pixel-cafe-workstation-layout-s254

## Role

Planner/Generator/Evaluator: Codex defines and reviews this contract, implements
the approved bounded change in the current dirty primary worktree, and performs
the final evidence review. No sub-agent dispatch is used for this task.

## Goal

Let an administrator place the ten Pixel Cafe computer workstations directly on
the lobby image, persist that layout on the server, and render the same layout
for every visitor. Align the Pixi coordinate transform with the lobby image so
saved positions remain visually stable across viewport sizes.

## Success Criteria

- `/admin/pixel-cafe/rooms` exposes a desktop-oriented lobby-layout dialog.
- Each of the ten numbered workstations can be dragged and keyboard-nudged;
  the draft can be reset, cancelled, or saved.
- Saving uses an authenticated administrator endpoint and persists one shared
  layout, not browser-local storage.
- The backend requires exactly ten unique IDs, finite in-bounds coordinates,
  and a bounded JSON payload; invalid layouts do not overwrite the saved value.
- Public settings expose only workstation IDs and coordinates. They expose no
  administrator or account data.
- The public Pixi renderer and its fallback consume the saved layout. Seated
  avatars follow their workstation; walking routes remain inside the lobby.
- Missing or malformed stored settings safely fall back to the built-in layout.
- The background and Pixi overlay use the same 16:9 design space and cover
  transform, so layout positions do not drift solely because the viewport
  aspect ratio changes.
- Existing room/share/account workflows, S253 walk frames, reduced-motion,
  unrelated dirty files, and `outputs/` remain intact.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `AGENTS.md`, `docs/workflow/status.md`,
  `docs/workflow/spec.md`, this contract.
- Current public scene owners:
  `frontend/src/features/pixelCafe/components/CafeScene.vue`,
  `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`, and
  `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`.
- Current administrator workspace:
  `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`.
- The lobby bitmap is `2048x1152` (16:9). Existing overlay constants use a
  different aspect ratio and a contain transform while the image uses cover.

## Allowed Paths

- `backend/internal/service/domain_constants.go`
- `backend/internal/service/pixel_cafe_workstation_layout.go`
- `backend/internal/service/pixel_cafe_workstation_layout_test.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service_public_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/setting_handler.go`
- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `backend/internal/handler/wire.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/cafe_admin_routes_test.go`
- `backend/cmd/server/wire_gen.go`
- `frontend/src/types/index.ts`
- `frontend/src/types/pixelCafe.ts`
- `frontend/src/api/admin/cafeRooms.ts`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/__tests__/createCafeRenderer.spec.ts`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`
- `frontend/src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts`
- `frontend/src/views/admin/pixelCafe/components/CafeWorkstationLayoutEditor.vue`
- `frontend/src/views/admin/pixelCafe/components/__tests__/CafeWorkstationLayoutEditor.spec.ts`
- `frontend/src/i18n/locales/zh/admin/pixelCafe.ts`
- `frontend/src/i18n/locales/en/admin/pixelCafe.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/pixel-cafe-workstation-layout-s254.md`
- `knowledge/tasks/current-task.md` only after final verification

## Denied Paths

- Database schema, migrations, Ent generation, dependencies, lockfiles, image
  generation, room/share/payment/account state machines, and provider behavior.
- Docker, Compose, container replacement, deployment, shared/production data,
  commit, push, branch cleanup, memory writes, and `outputs/**`.
- Any cleanup, reset, broad formatting, or staging of existing user changes.

## Constraints

- Canonical design space is `960x540`, matching the bitmap's 16:9 ratio.
- The persisted shape is a JSON array of exactly ten objects:
  `{ "id": 1..10, "x": number, "y": number }`.
- IDs must be unique and cover `1..10`. Coordinates are rounded to one decimal
  and constrained to safe visible anchor bounds. The backend is authoritative.
- The editor may snap to a small grid but must preserve exact server-normalized
  values after save/readback. Cancel must not mutate saved state.
- Reset changes only the unsaved draft until the administrator presses Save.
- Public users cannot write layouts. Mobile public rendering remains supported;
  precise editing may instruct administrators to use a desktop viewport.
- Use the existing setting repository; do not add a table or migration.
- Preserve all current S252/S253 behavior and dirty-worktree changes.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestPixelCafeWorkstationLayout|TestSettingService_GetPublicSettings_PixelCafe" -count=1
go test ./internal/handler ./internal/handler/admin ./internal/server/routes -run "Test.*Cafe.*Layout|Test.*PixelCafe.*Settings" -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/features/pixelCafe/components/__tests__/CafeScene.spec.ts src/features/pixelCafe/renderer/__tests__/createCafeRenderer.spec.ts src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/admin/pixelCafe/components/__tests__/CafeWorkstationLayoutEditor.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
git ls-files -u
```

Browser QA uses a task-owned independent Edge profile because the installed
Chrome 103 is incompatible with the current Playwright CLI. Verify desktop
drag/save/refresh persistence and `1440x1000` plus `390x844` public rendering,
then close the exact session and prove its profile/daemon processes are gone.

## Output

- Final evaluation is `PASS`, `FAIL`, or `BLOCKED` with changed files, commands,
  evidence, residual risks, and the unchanged external-state boundary.
- No commit, push, container update, or shared database write is performed.

## Stop Rules

- Stop if the existing setting repository cannot provide atomic single-key
  persistence without a migration, if authentication cannot be inherited from
  the admin route group, or if the scene cannot share one coordinate transform.
- Stop before changing room/payment/account schema or behavior.
- Stop rather than overwrite or clean unrelated dirty work.

## Budget

- controller_implementation: local Codex
- worker_dispatch: disabled for this task
- qa_mode: runtime plus task-owned browser acceptance
