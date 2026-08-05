---
task_id: leaderboard-record-banner-s180
phase: contract-approved
role: Planner/Generator/Evaluator
qa_mode: browser
---

# S180 Leaderboard Record Banner Contract

## Goal

Simplify the leaderboard composition by placing the existing dynamic personal record over a newly generated
wide illustration at the bottom of the left ranking panel. Remove the right-side Thursday promotion and the
standalone personal-record card so the right column contains only the weekly ranking/reward information with
more breathing room.

## Success Criteria

- The bottom illustration remains inside the left ranking panel and uses a new project-owned, text-free wide
  raster asset with calm left-side negative space for dynamic record copy and ranking imagery on the right.
- When reward extras are enabled, `你的战绩`, the existing ranked/unranked headline, and the existing progress
  state render as readable HTML text over the illustration. All current no-data, awaiting-draw, distance, top-1,
  and `掌控token的神` branches keep their existing computation and are not baked into the image.
- The right column no longer imports or renders `crazy-thursday-v50.png`, `leaderboard-thursday-banner`, or the
  standalone `leaderboard-record-card`. It retains the weekly top-10, status, progress, lottery/red-packet facts,
  errors and claim action without changing API calls or reward behavior.
- The right card is visually relaxed: its weekly list shows more rows before scrolling on wide desktop, while
  desktop and mobile remain readable without horizontal overflow or text/image collisions.
- Chinese copy is tightened for the banner context and English parity remains present. Decorative images have
  empty alt text; dynamic record content remains accessible text.

## Allowed Paths

- `frontend/src/views/user/LeaderboardView.vue`
- `frontend/src/views/user/__tests__/LeaderboardView.spec.ts`
- `frontend/src/__tests__/leaderboard-theme.spec.ts`
- `frontend/src/i18n/locales/zh/leaderboard.ts`
- `frontend/src/i18n/locales/en/leaderboard.ts`
- `frontend/src/assets/leaderboard/leaderboard-record-banner.png`
- `output/imagegen/leaderboard-record-banner-source.png`
- `output/playwright/leaderboard-s180/**`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/leaderboard-record-banner-s180.md`
- `docs/workflow/qa-reports/leaderboard-record-banner-s180-qa.md`
- `docs/workflow/qa-reports/pixel-cafe-phase30-presentation-settings-s176-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Backend, API contracts, reward settlement or claim behavior, database/migrations, router/access controls,
  unrelated console/public/Pixel Cafe files, Docker/container/deployment configuration, remote Git, production
  settings/data, `knowledge/tasks/timeline.md`, and global memories.

## Constraints

- Reuse `myRecordHeadline`, `myRecordProgress`, and `buildRecordProgress`; do not introduce a second record
  computation or duplicate reward state.
- The generated bitmap must contain no visible words, logos, watermarks, email/account details, or live user data.
  Do not upload the supplied screenshot or any authenticated page data to image generation.
- Preserve existing dirty Pixel Cafe/sidebar/workflow changes. Do not reformat or revert unrelated files.
- Do not delete the old raster assets in this Sprint; only remove their runtime import/reference where applicable.
- No container update, deployment, push, claim action, settings write, or production operation is authorized.

## Acceptance Commands

```powershell
cd frontend
npm.cmd run test:run -- src/views/user/__tests__/LeaderboardView.spec.ts src/__tests__/leaderboard-theme.spec.ts
npm.cmd run typecheck
npm.cmd run build
cd ..
git diff --check
git ls-files -u
```

Manual browser acceptance: render the authenticated local `/leaderboard` route at a wide desktop viewport and
390px mobile. Confirm the record is inside the left banner; no Thursday or standalone record card remains on
the right; weekly ranking/reward facts remain; no horizontal overflow or overlap occurs. Save evidence under
`output/playwright/leaderboard-s180/`.

## Output

- One generated project banner and a constrained leaderboard layout/copy/test change.
- `docs/workflow/qa-reports/leaderboard-record-banner-s180-qa.md` starts with `### PASS: ...`,
  `### FAIL: ...`, or `### BLOCKED: ...`.

## Stop Rules

- Stop for any required API/backend/reward-semantic change, user-data upload, new dependency, remote call beyond
  the text-only image prompt, production/session mutation, or overlap with unrelated dirty leaderboard edits.
- Stop if authenticated browser proof would require entering or exposing credentials; report the browser gate.

## Evaluator Review

### PASS: leaderboard-record-banner-s180

- Existing `myRecordHeadline` and `myRecordProgress` are view-level computed state and can move into the existing
  illustration without touching reward semantics or backend data.
- The right Thursday image and standalone record block are isolated presentation children; removing them leaves
  the weekly ranking/reward panel and its actions intact.
- A text-free generated asset plus HTML overlay preserves dynamic copy, localization and accessibility while the
  allowed frontend tests and authenticated desktop/mobile screenshots make the acceptance boundary executable.
