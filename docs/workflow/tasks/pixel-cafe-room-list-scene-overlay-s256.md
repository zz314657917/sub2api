# Pixel Cafe Room List Scene Overlay S256

## Task ID

pixel-cafe-room-list-scene-overlay-s256

## Role

Planner/Generator/Evaluator: Codex defines and reviews this contract, implements
the approved bounded change in the current dirty primary worktree, and performs
the final evidence review. No sub-agent dispatch is used for this task.

## Goal

Place the only public Pixel Cafe room list directly over the existing lobby
background so the scene and room discovery read as one interface, while keeping
the existing room-details dialog and share purchase behavior unchanged.

## Success Criteria

- The DOM contains exactly one `pixel-cafe-room-list`, and it is a descendant
  of `pixel-cafe-scene`; no room-list copy remains before or after the scene.
- On desktop the list is a translucent right-side overlay with an internal
  vertical scroller. It leaves the top status, left-side front desk, and bottom
  demo badge readable and does not resize the 16:9 scene artwork.
- On mobile the list remains inside the scene and becomes a compact horizontal
  card strip near the bottom of the artwork. The page itself has no horizontal
  overflow.
- Room cards retain code/name, Plus or Pro badge, share progress, buyer count,
  validity, price, and anonymous member avatars.
- Clicking a room card still opens the existing teleported details dialog;
  share quantity, payment, unavailable-state and demo-mode behavior are not
  changed.
- Loading, error, empty and local-demo states remain usable without rendering a
  second list or hiding the scene artwork.
- Desktop `1440x1000` and mobile `390x844` browser acceptance prove containment,
  scrolling geometry, dialog opening and zero document overflow.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: current uncommitted S252-S255 Pixel Cafe work in the primary
  worktree. It is user-owned and must be extended, never reset or cleaned.
- Existing room cards and handlers already implement the required content and
  dialog behavior. This task moves that template once and changes layout CSS;
  it does not create a parallel mobile/desktop component.
- The scene currently owns the right-top status, left-bottom front desk and
  right-bottom local-demo badge, so overlay spacing must explicitly avoid them.

## Allowed Paths

- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue` only if a scene
  stacking hook is strictly required
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts` only
  if `CafeScene.vue` changes
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/pixel-cafe-room-list-scene-overlay-s256.md`
- `knowledge/tasks/current-task.md` only after final verification

## Denied Paths

- Backend, database schema/migrations, Ent, room/share/payment/account state
  machines, settings, admin UI, renderer behavior, dependencies/lockfiles,
  image assets or generation, and `outputs/**`.
- Docker/Compose/container replacement, deployment, shared data, commit, push,
  branch/worktree cleanup, memory writes, broad formatting, and staging.

## Constraints

- Move and reuse the existing room-list markup; do not duplicate it for a
  breakpoint and do not add a second request or room state.
- Preserve the current `openRoom` path and teleported dialog ownership.
- Use semantic buttons and retain visible keyboard focus within the overlay.
- Desktop list scrolling must be internal; mobile horizontal scrolling must be
  confined to the room-card strip with touch-friendly snap behavior.
- The scene overlay may be visually translucent but card text must remain
  readable over the brightest lobby artwork.
- Preserve the S253-S255 Pixi/fallback, walk animation, configurable workstation
  count and reduced-motion behavior.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/features/pixelCafe/components/__tests__/CafeScene.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check -- frontend/src/features/pixelCafe/PixelCafePage.vue frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/pixel-cafe-room-list-scene-overlay-s256.md knowledge/tasks/current-task.md
git ls-files -u
git diff --cached --name-only
```

Browser QA uses Google Chrome 151 from
`C:/Program Files/Google/Chrome/Application/chrome.exe` with task-owned profile
`E:/codex-browser/sub2api-s256/chrome-profile`. Verify demo-mode desktop
`1440x1000` and mobile `390x844`, then close the exact session and prove the
S256 profile, Playwright daemon, task Vite PID and port are gone.

## Output

- Final evaluation is `PASS`, `FAIL`, or `BLOCKED`, with changed files, executed
  commands, browser geometry, residual risks and unchanged external-state
  boundaries.
- Do not commit, push, update the local container, or write shared settings.

## Stop Rules

- Stop rather than changing the public room DTO, purchase state machine or
  dialog contract to make the layout work.
- Stop if containment requires disabling the scene renderer or removing the
  existing accessibility semantics.
- Stop rather than overwriting unrelated dirty changes in either allowed
  business file.

## Budget

- controller_implementation: local Codex
- worker_dispatch: disabled for this task
- qa_mode: runtime plus task-owned browser acceptance
