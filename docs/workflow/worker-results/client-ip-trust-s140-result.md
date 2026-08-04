### DONE: client-ip-trust-s140

## Changed Behavior

- Adapted the forwarded client-IP trust chain to the local configuration,
  settings, HTTP ingress, API-key middleware, session-binding, and audit seams.
- Raw forwarded headers remain disabled by default. Missing or explicitly empty
  `server.trusted_proxies` keeps Gin's forwarded-client resolution fail-closed.
- Added normalized, case-insensitive, de-duplicated custom header settings with
  a 16-header bound and immutable per-request snapshots. Primary and
  Google/Gemini API-key ACLs use the same security client-IP helper.
- Added administrator settings API/UI fields and Chinese/English locale text.
- Added secure operator guidance to `README.md`, `README_CN.md`, and
  `deploy/config.example.yaml`.

## Files

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_partial_payload_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/pkg/ip/ip.go`
- `backend/internal/pkg/ip/ip_test.go`
- `backend/internal/server/http.go`
- `backend/internal/server/http_ingress_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_google.go`
- `backend/internal/server/middleware/api_key_auth_google_test.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `backend/internal/server/middleware/session_binding.go`
- `backend/internal/server/middleware/session_binding_test.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/setting_service_partial_payload_test.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/wire.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `README.md`
- `README_CN.md`
- `deploy/config.example.yaml`

## Checks

- Focused Go config, IP, HTTP ingress, middleware, admin-handler, and service
  tests passed.
- `go test ./... -run '^$'` passed as a repository compile probe.
- Frontend typecheck, SettingsView Vitest (26/26), eslint, and production build
  (1101 modules) passed.
- `gofmt -d`, `git diff --check`, unmerged-index, conflict-marker, and the
  implementation/documentation allowlist audit passed. Workflow bookkeeping
  edits are retained for the separate publication receipt.

## Risks

- `go test -race` could not run: the environment has no C compiler (`gcc`).
- A wider service test selection retains unrelated pre-existing peak-multiplier
  failures; the focused S140 service selection passed.
- No real reverse-proxy chain, authenticated PostgreSQL settings round trip,
  deployment, container refresh, or production browser smoke was run.
