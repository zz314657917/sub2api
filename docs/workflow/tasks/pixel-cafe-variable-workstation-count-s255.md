# Pixel Cafe Variable Workstation Count S255

## Task ID

pixel-cafe-variable-workstation-count-s255

## Role

Planner/Generator/Evaluator: Codex defines and reviews this contract, implements
the approved bounded change in the current dirty primary worktree, and performs
the final evidence review. No sub-agent dispatch is used for this task.

## Goal

Let an administrator choose how many Pixel Cafe computer workstations appear in
the shared lobby layout instead of forcing exactly ten, while preserving the
existing saved coordinates and the S254 drag editor.

## Success Criteria

- The lobby-layout editor exposes an explicit workstation-count control.
- The accepted range is 1 through 50; the built-in/default count remains 10.
- Increasing the count preserves all existing positions and appends contiguous
  numbered workstations at deterministic editable positions.
- Decreasing the count preserves IDs `1..N` and removes only the highest IDs;
  no saved change occurs until Save is pressed.
- Reset restores deterministic positions for the currently selected count.
- The backend accepts a bounded array whose length is 1..50 and whose unique
  IDs cover exactly `1..length`; coordinate and payload limits remain intact.
- Existing valid ten-workstation settings load without conversion or migration.
- Missing, empty, or malformed stored values still fall back to the ten-slot
  built-in layout.
- Public Pixi and static fallback render exactly the saved workstation count.
  Seated-avatar capacity follows that count, with at most six additional
  walking avatars; repeated workstation sprites remain presentation-only.
- Desktop and 390x844 mobile public layouts have no horizontal overflow.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: the current uncommitted S252/S253/S254 primary worktree. Those
  changes are user-owned and must be extended, never reset or cleaned.
- Existing storage is Setting key `pixel_cafe_workstation_layout`; array length
  is the authoritative count, so no second count key is introduced.
- Current design space remains `960x540` with visible coordinate bounds
  `x=48..912`, `y=72..520`.

## Allowed Paths

- `backend/internal/service/pixel_cafe_workstation_layout.go`
- `backend/internal/service/pixel_cafe_workstation_layout_test.go`
- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/__tests__/createCafeRenderer.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`
- `frontend/src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts`
- `frontend/src/views/admin/pixelCafe/components/CafeWorkstationLayoutEditor.vue`
- `frontend/src/views/admin/pixelCafe/components/__tests__/CafeWorkstationLayoutEditor.spec.ts`
- `frontend/src/i18n/locales/zh/admin/pixelCafe.ts`
- `frontend/src/i18n/locales/en/admin/pixelCafe.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/pixel-cafe-variable-workstation-count-s255.md`
- `knowledge/tasks/current-task.md` only after final verification

## Denied Paths

- Database schema/migrations, Ent, dependencies/lockfiles, image generation,
  room/share/payment/account state machines, providers, and `outputs/**`.
- Docker/Compose/container replacement, deployment, shared or production data,
  commit, push, branch cleanup, memory writes, and broad formatting/staging.

## Constraints

- Persist the existing array shape only: `{ "id": number, "x": number,
  "y": number }`; do not add a count field or new setting key.
- IDs must be sorted and contiguous after normalization: `1..length`.
- The maximum request body remains 4 KiB; 50 normalized items must fit within
  that boundary.
- Auto-placement is deterministic and bounded, but it is only an initial draft;
  the administrator remains responsible for visual placement before Save.
- Workstation-count edits remain local to the dialog until Save. Cancel cannot
  mutate the server value.
- Preserve S253 walk-frame direction, speed, reduced-motion and wall-safe routes.
- Public settings expose only `id/x/y`, as before.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestPixelCafeWorkstationLayout" -count=1
go test ./internal/handler/admin -run "TestCafeRoomHandlerReadsAndUpdatesWorkstationLayout" -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/features/pixelCafe/renderer/__tests__/createCafeRenderer.spec.ts src/features/pixelCafe/components/__tests__/CafeScene.spec.ts src/views/admin/pixelCafe/components/__tests__/CafeWorkstationLayoutEditor.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
git ls-files -u
git diff --cached --name-only
```

Browser QA uses a task-owned independent Edge profile. Verify the admin count
control at 1, 10, and 50; save/readback persistence may use intercepted APIs so
no shared setting is written. Verify public desktop `1440x1000` and mobile
`390x844`, then close the exact browser session and prove the task profile and
Playwright daemon are gone.

## Output

- Final evaluation is `PASS`, `FAIL`, or `BLOCKED`, with changed files, executed
  commands, evidence, residual risks, and unchanged external-state boundaries.
- Do not commit, push, update the local container, or write shared settings.

## Stop Rules

- Stop if dynamic count requires a schema/migration or exposes a second source
  of truth outside the existing layout array.
- Stop rather than changing room/share/account behavior or overwriting unrelated
  dirty work.
- If 50 normalized workstations cannot fit the existing 4 KiB contract, revise
  the contract before widening the payload limit.

## Budget

- controller_implementation: local Codex
- worker_dispatch: disabled for this task
- qa_mode: runtime plus task-owned browser acceptance
