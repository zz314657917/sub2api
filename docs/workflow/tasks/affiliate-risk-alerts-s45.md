---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 17:20 +08:00
---

# Task Contract

## Task ID
affiliate-risk-alerts-s45

## Role
Generator worker implements only after this contract is reviewed and approved by Codex. Codex remains Final Evaluator.

## Goal
Add a minimum viable affiliate-abuse risk scanner that scores suspicious invite/reward clusters, writes reusable `ops_alert_events`, and freezes affiliate reward monetization for high-risk clusters without automatically banning users, disabling API keys, or clawing back historical rewards. The scan interval must be configurable from admin settings, defaulting to `20m`.

## Success Criteria
- A backend scanner interval is configurable from admin settings, defaults to `20m`, and scans the most recent `12h` of affiliate registrations, login IPs, API usage IPs, and `api_call_reward` ledger activity.
- The scanner interval setting is validated with a safe range, recommended `5` to `1440` minutes, and invalid values fall back to the default `20m` rather than crashing the scanner.
- Scanner uses risk scoring rather than single-rule bans.
- Scanner detects at least:
  - same inviter has `>= 3` invitees in the scan window,
  - inviter and invitee share login IP,
  - inviter/invitees share IPv6 `/64`,
  - invitee register IPs are dispersed but later login IP or IPv6 `/64` aggregates,
  - invitee triggers first API reward within `30m` after registration,
  - multiple invitee email local parts look batch-generated,
  - invitee relation was revoked/disabled but `api_call_reward` already exists.
- Scanner assigns severity:
  - `score >= 90`: `P1`
  - `score >= 70`: `P2`
  - `score >= 50`: `P3`
  - `< 50`: no alert/freeze
- `P3` writes an ops alert only.
- `P2/P1` writes an ops alert and freezes affiliate reward monetization for the inviter/relevant cluster.
- Freeze blocks:
  - manual first API call reward claim,
  - transfer of affiliate quota to balance.
- Freeze does not:
  - remove existing inviter binding,
  - deduct existing ledger rewards,
  - ban users,
  - disable API keys,
  - block normal API usage.
- Alert title/description clearly shows why it fired, for example: `疑似刷邀请奖励：3388637010@qq.com 12小时内关联5个小号，3个共享登录IPv6`.
- Events are deduplicated so the same inviter/risk cluster does not create duplicate active alerts every scan.
- If ops email alert config is enabled, risk alerts follow the existing ops email notification path.
- Admin settings expose the scan interval so operators can adjust it without code changes; no dedicated risk-management page is required in this Sprint.
- Migration adds the minimum risk-scan indexes:
  - `idx_users_created_at` on `users(created_at)`,
  - `idx_user_affiliate_ledger_action_created_at` on `user_affiliate_ledger(action, created_at)`,
  - `idx_usage_logs_ip_created_at` on `usage_logs(ip_address, created_at) WHERE ip_address <> ''`.

## Context
- Repo: `F:/mcplugins/sub2api`
- Existing data sources:
  - `users.register_ip`
  - `users.last_login_ip`
  - `usage_logs.ip_address`
  - `user_affiliates.inviter_id`
  - `user_affiliate_ledger.action = 'api_call_reward'`
- Existing ops alert stack:
  - `backend/internal/service/ops_alert_evaluator_service.go`
  - `backend/internal/service/ops_alerts.go`
  - `backend/internal/repository/ops_repo_alerts.go`
  - `backend/migrations/033_ops_monitoring_vnext.sql`
- Existing affiliate reward entry points:
  - `backend/internal/service/affiliate_service.go`
  - `backend/internal/repository/affiliate_repo.go`
  - user handler endpoints for `/aff/transfer` and manual API reward claim.
- `OpsService.CreateAlertEvent` currently only writes events; existing evaluator sends email after event creation. This task should either reuse that email helper behavior in a shared way or add a small, tested helper for risk scanner notifications without duplicating large mail logic.
- Existing scheduler services use `Start()`, `Stop()`, `time.Timer`, Redis leader lock, and ops job heartbeat patterns. Follow that style.
- Current main worktree is heavily dirty in payment/welfare/settings/gateway/frontend. This task must be implemented in an isolated clean worktree or after unrelated dirty paths are resolved.
- Existing useful indexes already include `usage_logs(created_at)`, `usage_logs(user_id, created_at)`, `usage_logs(ip_address)`, `user_affiliates(inviter_id)`, `user_affiliates(inviter_id) WHERE inviter_id IS NOT NULL AND affiliate_revoked_at IS NULL`, and the API reward uniqueness index. The new migration should only add the missing scan-specific indexes listed in Success Criteria.

## Risk Scoring
- Same inviter invites `>= 3` accounts within `12h`: `+25`
- Inviter and invitee same `last_login_ip`: `+40`
- Same IPv6 `/64`: `+35`
- Register IP dispersed but login IP or `/64` aggregated: `+25`
- API reward within `30m` after invitee registration: `+20`
- Multiple invitee emails look batch-generated: `+10`
- Invitee relation revoked/disabled but `api_call_reward` exists: `+30`
- Cap or explain duplicate signal accumulation so one underlying IP relationship does not inflate unboundedly across many invitees.

## Required Preflight
- Review `git status --short --branch` before edits.
- Stop if implementing in the current dirty main worktree would touch unrelated dirty payment/welfare/gateway/frontend paths.
- Treat Settings frontend/backend files as allowed only for adding the interval setting; do not mix in unrelated settings changes already present in the dirty tree.
- Confirm the next safe migration number. Current `main` already contains `181_add_group_peak_rate_multiplier.sql`; the welfare voucher branch uses `182_welfare_vouchers.sql`. Do not assume the next number without checking at implementation time.
- Confirm no existing affiliate risk scanner/freeze table was added since this contract was written.

## Allowed Paths
- `backend/internal/pkg/ip/ip.go`
- `backend/internal/pkg/ip/ip_test.go`
- `backend/internal/service/affiliate_risk_scanner.go`
- `backend/internal/service/affiliate_risk_scanner_test.go`
- `backend/internal/service/affiliate_service.go`
- `backend/internal/service/affiliate_service_test.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/ops_alert_evaluator_service.go`
- `backend/internal/service/ops_alert_evaluator_service_test.go`
- `backend/internal/service/ops_alerts.go`
- `backend/internal/service/ops_alerts_test.go`
- `backend/internal/service/ops_port.go`
- `backend/internal/service/wire.go`
- `backend/internal/repository/affiliate_repo.go`
- `backend/internal/repository/affiliate_repo_integration_test.go`
- `backend/internal/repository/affiliate_risk_repo.go`
- `backend/internal/repository/affiliate_risk_repo_test.go`
- `backend/internal/repository/ops_repo_alerts.go`
- `backend/internal/repository/ops_repo_alerts_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/admin/setting_handler.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `backend/migrations/*_affiliate_risk_freezes.sql`
- `backend/migrations/*_affiliate_risk_indexes.sql`
- `docs/workflow/worker-results/affiliate-risk-alerts-s45-result.md`
- `docs/workflow/qa-reports/affiliate-risk-alerts-s45-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `frontend/**`, except the explicitly allowed admin settings files above.
- `knowledge/**`
- `deploy/**`
- `assets/**`
- `README*`
- `.github/**`
- `backend/internal/payment/**`
- `backend/internal/server/routes/payment.go`
- `backend/internal/handler/payment_handler.go`
- `backend/internal/handler/admin/payment_handler.go`
- `backend/internal/service/payment_*`
- `backend/internal/service/welfare_*`
- `backend/internal/repository/welfare_*`
- `backend/internal/handler/openai_images.go`
- `backend/internal/repository/openai_video_task_repo.go`
- `backend/internal/service/openai_images_*`
- `backend/internal/service/openai_video*`
- `backend/internal/service/studio_bridge.go`
- `backend/internal/repository/studio_bridge_repo.go`
- `backend/go.mod`
- `backend/go.sum`
- Any unrelated SettingsView/settings DTO/i18n/type changes not required for the scan interval setting.
- Any unlisted dirty path.

## Constraints
- Do not add automatic user banning in this Sprint.
- Do not add automatic API key disabling in this Sprint.
- Do not claw back historical rewards in this Sprint.
- Do not revoke inviter bindings as part of risk scoring; existing same-payment-method revoke logic remains separate.
- Do not block normal API usage.
- Keep risk display backend-only for the minimum viable version; frontend UI work is limited to the admin settings control for scan interval because ops center already lists `ops_alert_events`.
- Scanner should read the interval setting each cycle or otherwise refresh it without requiring a process restart.
- Prefer one focused persistence model for freezes, for example `affiliate_risk_freezes`, with reason, severity, score, source window, active status, and timestamps.
- Dedup active risk events/freeze records by inviter/risk fingerprint and window so a configurable scanner interval does not spam.
- IPv6 normalization must map `2409:8962:e1:391d:7d22:7006:9425:c2f8` to `2409:8962:e1:391d::/64`; IPv4 should normalize to the exact IP string.
- Scanner should degrade safely if Redis lock is unavailable: prefer the existing ops-service stance of fail-closed when distributed lock is enabled.
- If ops monitoring is disabled, scanner should not create ops alerts, but freeze behavior should be explicitly tested as enabled/disabled according to the chosen setting or config.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/pkg/ip -run "Test.*Normalize.*IPv6.*64|Test.*Normalize.*IP" -count=1
go test ./internal/service -run "TestAffiliateRisk.*|Test.*Affiliate.*Freeze.*|Test.*Affiliate.*Risk.*|Test.*Ops.*Alert.*Email|Test.*Affiliate.*Scan.*Interval|Test.*Setting.*Affiliate.*Risk" -count=1
go test ./internal/repository -run "TestAffiliateRisk.*|TestAffiliateRepo.*Freeze.*|TestAffiliateRepo.*Claim.*Risk.*|TestAffiliateRepo.*Transfer.*Risk.*" -count=1
go test ./cmd/server -run "TestWire.*AffiliateRisk|TestWireGenerated" -count=1

cd F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
git diff --check
git diff --cached --name-only | rg "^(knowledge/|deploy/|assets/|README|README_|\.github/|backend/internal/payment/|backend/internal/server/routes/payment.go|backend/internal/handler/payment_handler.go|backend/internal/handler/admin/payment_handler.go|backend/internal/service/payment_|backend/internal/service/welfare_|backend/internal/repository/welfare_|backend/internal/handler/openai_images.go|backend/internal/repository/openai_video_task_repo.go|backend/internal/service/openai_images_|backend/internal/service/openai_video|backend/internal/service/studio_bridge.go|backend/internal/repository/studio_bridge_repo.go|backend/go.mod|backend/go.sum)" || echo NO_DENIED_PATHS
```

## Output
- Business implementation in allowed paths only.
- Worker result: `docs/workflow/worker-results/affiliate-risk-alerts-s45-result.md`
- QA report: `docs/workflow/qa-reports/affiliate-risk-alerts-s45-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`.

## Stop Rules
- Stop before editing if the implementation worktree contains unrelated dirty files in allowed or denied business paths.
- Stop if freeze semantics require changing payment/welfare/account-ban/API-key-disable flows.
- Stop if adding the settings control would require a broad SettingsView refactor; only add the minimal setting field.
- Stop if ops email sending requires a broad rewrite of `OpsAlertEvaluatorService`; extract a small shared helper or leave email as a documented follow-up.
- Stop if scanner queries require full-table scans without usable time-window predicates or indexes.
- Stop if migration numbering conflicts with existing tracked/untracked migrations.
- Stop if adding scan indexes would require rewriting existing usage log partitioning or dashboard aggregation migrations; add narrow indexes only.

## Budget
- worker_mode: `claude-bare-deepseek-v4-pro`
- qa_worker_mode: `claude-bare-deepseek-v4-pro`
- worker_model: `deepseek-v4-pro`
- max_budget_usd: `0.20`
- worktree_root: `E:/codex-worktrees`
