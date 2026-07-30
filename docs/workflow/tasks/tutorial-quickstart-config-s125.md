---
task_id: tutorial-quickstart-config-s125
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Make the public `/tutorial` quick-start guide administratively configurable so
operators can change platform connection information, especially Base URLs,
without editing or redeploying frontend source. Keep the guide usable before
any configuration is saved.

## Success Criteria

- Administrators can read, save, and reset the quick-start configuration from
  the existing Tutorial Management page.
- The public guide reads the saved configuration. Changing a platform Base URL
  changes its information card, generated CLI snippets, and cURL example.
- The configuration stores only public, plain-text guide content: platform
  labels, endpoint/auth/protocol/model hints, page copy, and troubleshooting
  entries. It must not support HTML or executable content.
- Missing, malformed, or unavailable persisted configuration falls back to the
  built-in guide; public visitors never receive an administrator-only payload.
- The implementation uses the existing settings repository and does not add a
  schema migration or change deployment/container configuration.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Existing public quick-start UI is an uncommitted, user-requested change in
  `frontend/src/views/public/TutorialView.vue`; preserve and build on it.
- Reuse the DB-backed JSON configuration pattern used by the model-market
  catalog, but keep this feature scoped to tutorial data.

## Allowed Paths

- `backend/internal/service/domain_constants.go`
- `backend/internal/service/quickstart_tutorial.go`
- `backend/internal/service/quickstart_tutorial_test.go`
- `backend/internal/handler/setting_handler.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/server/routes/auth.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/api/tutorials.ts`
- `frontend/src/api/admin/tutorials.ts`
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/tutorialQuickstart.ts`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `frontend/src/views/admin/TutorialPagesView.vue`
- `frontend/src/api/__tests__/admin.tutorials.spec.ts`
- `docs/workflow/**`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/group-buy/**`
- `knowledge/**` except `knowledge/tasks/current-task.md`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`

## Constraints

- Keep the routes behind existing administrator authentication for writes and
  expose only a dedicated public read endpoint for the guide.
- Validate configuration shape, URL schemes, item counts, string lengths, and
  stable platform identifiers before persistence. Do not permit arbitrary HTML
  or external links from configuration.
- Preserve CMS tutorial article CRUD and `/tutorial?view=library` behavior.
- Do not deploy, update containers, push, create a commit, reset, clean, or
  revert any existing user changes.
- The primary worktree is dirty. Inspect diffs carefully and restrict all
  verification to the allowed paths.

## Acceptance Commands

```powershell
cd backend
go test ./internal/service -run "QuickstartTutorial" -count=1
go test ./internal/handler -run "QuickstartTutorial" -count=1
go test ./internal/handler/admin -run "QuickstartTutorial" -count=1
go test ./... -run "^$"
cd ../frontend
corepack.cmd pnpm exec vitest run src/views/public/__tests__/TutorialView.spec.ts src/api/__tests__/admin.tutorials.spec.ts
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build
cd ..
git diff --check
```

## Output

- A focused implementation and `docs/workflow/qa-reports/tutorial-quickstart-config-s125-qa.md`.
- No worker is invoked for this bounded task; Codex performs implementation,
  QA, and final evaluation directly.

## Stop Rules

- Stop if satisfying the configuration requires a database migration,
  deployment/container change, generic CMS rewrite, or modification of a
  denied path.
- Stop if a safe public fallback cannot be maintained when persisted data is
  unavailable or invalid.
- Stop if an existing user modification overlaps a required change and cannot
  be preserved.
