### DONE: affiliate-risk-alerts-s45

## Summary
- Implemented a minimum viable affiliate risk scanner with 12h scan window, configurable interval setting defaulting to 20 minutes, IPv6 /64 normalization, score-based P1/P2/P3 alerting, and active freeze records for P1/P2 monetization blocks.
- Added `affiliate_risk_freezes` persistence plus scan-specific indexes for `users(created_at)`, `user_affiliate_ledger(action, created_at)`, and `usage_logs(ip_address, created_at) WHERE ip_address <> ''`.
- Wired the scanner into server startup/cleanup and reused `ops_alert_events`; active alerts are deduplicated by risk dimensions.
- Added service-level freeze checks before first API call reward claim and affiliate quota transfer. This does not ban users, disable API keys, revoke bindings, claw back rewards, or block normal API usage.
- Added admin settings field `affiliate_risk_scan_interval_minutes` in backend DTO/service and the existing admin Settings affiliate section.

## Changed Files
- Backend scanner/settings/freeze: `backend/internal/service/affiliate_risk_scanner.go`, `backend/internal/service/affiliate_service.go`, `backend/internal/service/setting_service.go`, `backend/internal/service/domain_constants.go`, `backend/internal/service/settings_view.go`.
- Repository/ops/wire: `backend/internal/repository/affiliate_risk_repo.go`, `backend/internal/repository/affiliate_repo.go`, `backend/internal/repository/ops_repo_alerts.go`, `backend/internal/service/ops_port.go`, `backend/internal/service/wire.go`, `backend/internal/repository/wire.go`, `backend/cmd/server/wire.go`, `backend/cmd/server/wire_gen.go`.
- Migration: `backend/migrations/183_affiliate_risk_freezes.sql`.
- Frontend setting: `frontend/src/api/admin/settings.ts`, `frontend/src/views/admin/SettingsView.vue`, locale files.
- Tests: IP normalization, affiliate risk scoring/interval/freeze tests, affiliate risk repository tests, wire cleanup test update.

## Notes
- Ops alert email for scanner-created events uses the existing ops email notification config and severity threshold. It is intentionally small and does not duplicate the evaluator rule engine.
- The scanner uses Redis leader lock when Redis is configured; lock errors fail closed for that cycle.
- The candidate query is time-windowed and starts from invitees created in the window or invitees with API reward activity in the window.
