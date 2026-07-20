### PASS: upstream-main-compat-s86-generator

## Changed Files

- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_proxy_quality_test.go`
- `frontend/src/views/admin/ProxiesView.vue`

## Implemented Behavior

- Added the canonical `GET https://api.x.ai/v1/models` Grok quality target.
- Treats HTTP 401 as reachable by the existing generic quality-target runner.
- Added exact target-definition and 401 pass regressions.
- Displays backend target `grok` as `Grok` in the existing result table.
- Left scoring, timeouts, other targets, persistence, and table styling unchanged.

## Commands Run

- Focused proxy-quality Go tests: PASS.
- Frontend `npm.cmd run typecheck`: PASS.
- Frontend `npm.cmd run build`: PASS, 1080 modules transformed.
- `gofmt` and `git diff --check`: PASS.

## Risks / Deferred Checks

- No live outbound proxy probe was run; `httptest` proves request method and 401
  classification without contacting xAI.
- No authenticated browser smoke was run; typecheck/build and direct switch
  review cover the one-line display-name change.
- No push, deployment, container update, or primary-worktree change was made.
