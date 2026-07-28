---
task_id: upstream-v0165-audit-log-s121
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Port the complete administrator operation-audit capability present in upstream
`v0.1.165`, including the final credential/body redaction, trusted client-IP
handling, session IP/User-Agent binding, and TOTP step-up protection for
sensitive operations. The administrator console must expose the paginated
audit-log view at `/admin/audit-logs`.

## Success Criteria

- Authenticated management-plane mutations and configured sensitive reads are
  recorded with actor, method, path, request ID, status, latency, client IP,
  masked credentials, and a redacted, bounded request body.
- Audit records are queryable only by administrators through list/detail APIs
  and the Chinese/English administrator console view; audit records cannot be
  deleted individually.
- Clearing all audit records requires an active administrator TOTP verification
  and writes a post-clear trace record. Administrator API keys cannot clear.
- Session IP/User-Agent binding and step-up behavior use the upstream final
  secure behavior without storing raw credentials or session bodies in audit
  records.
- The local migration sequence remains valid: use a new
  `198_audit_logs.sql`, never reuse occupied migration number `180`.
- Existing S115-S119 Live, session-id, Gemini, and partial-setting behavior
  remains covered by compilation and focused regression gates.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-audit-log-s121-clean`
- Upstream baseline: `v0.1.165`; initial feature `0ddd58aaf`, with the
  later redaction/IP/session fixes through `2faa0891e`.
- The primary worktree is dirty and must remain untouched until this worktree
  has a reviewed, validated commit.

## Allowed Paths

- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `backend/internal/handler/admin/**`
- `backend/internal/handler/auth_handler.go`
- `backend/internal/handler/totp_handler.go`
- `backend/internal/handler/handler.go`
- `backend/internal/handler/wire.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/repository/audit_log_repo.go`
- `backend/internal/repository/audit_log_repo_test.go`
- `backend/internal/repository/totp_cache.go`
- `backend/internal/repository/wire.go`
- `backend/internal/server/http.go`
- `backend/internal/server/middleware/**`
- `backend/internal/server/router.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/auth.go`
- `backend/internal/server/routes/auth_rate_limit_test.go`
- `backend/internal/server/routes/payment.go`
- `backend/internal/server/routes/user.go`
- `backend/internal/service/audit_log*.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/auth_service.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/refresh_token_cache.go`
- `backend/internal/service/session_binding.go`
- `backend/internal/service/totp_service.go`
- `backend/internal/service/wire.go`
- `backend/migrations/198_audit_logs.sql`
- `frontend/src/api/admin/audit.ts`
- `frontend/src/api/admin/index.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/api/totp.ts`
- `frontend/src/components/auth/TotpStepUpDialog.vue`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/composables/useStepUp.ts`
- `frontend/src/composables/__tests__/useStepUp.spec.ts`
- `frontend/src/i18n/locales/{en,zh}/admin/audit.ts`
- `frontend/src/i18n/locales/{en,zh}/admin/index.ts`
- `frontend/src/i18n/locales/{en,zh}/admin/settings.ts`
- `frontend/src/i18n/locales/{en,zh}/common.ts`
- `frontend/src/i18n/locales/{en,zh}/nav.ts`
- `frontend/src/router/index.ts`
- `frontend/src/views/admin/{AccountsView,AuditLogView,BackupView,SettingsView}.vue`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/internal/domain/constants.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/migrations/180_*.sql`
- `frontend/src/components/account/**`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/views/admin/group-buy/**`
- `deploy/**`
- `Dockerfile*`
- `knowledge/**`
- `outputs/**`

## Constraints

- Preserve the upstream audit redaction allow/deny rules and do not log raw
  bearer tokens, API keys, passwords, TOTP codes, refresh tokens, cookies, or
  session payloads.
- Keep audit writes asynchronous/best-effort so persistence failure does not
  alter the original management operation response.
- Do not use a local-only JSONL logger; audit data belongs in PostgreSQL and
  must follow the existing migration runner.
- Do not run migrations, deploy, update containers, push, or clean the primary
  worktree in this task.
- Resolve source drift manually; do not merge broad upstream history.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-audit-log-s121-clean/backend
go test ./internal/server/middleware -run "Audit|SessionBinding|StepUp" -count=1
go test ./internal/service -run "Audit|SessionBinding|StepUp|Auth" -count=1
go test ./internal/handler/admin -run "Audit|StepUp|Settings" -count=1
go test ./... -run "^$"
cd ../frontend
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build
cd ..
git diff --check
git diff --name-only HEAD
```

## Output

- Isolated, reviewable audit-log security port; S121 implementation result;
  runtime QA report; and a single local commit only after all source-level
  checks pass.

## Stop Rules

- Stop if importing the feature requires changes to Ent schema/generated code,
  billing, gateway request handling, deployment, or an occupied migration.
- Stop if redaction/session-binding behavior cannot be kept fail-closed.
- Stop before primary-worktree integration if the committed feature overlaps an
  existing dirty file and the overlap cannot be resolved without changing the
  user's work.

## Amendment Log

- 2026-07-28: allow `setting_service.go` and `settings_view.go` for the three
  persisted audit/session/step-up settings, and `audit_log_repo_test.go` for
  transactional clear rollback coverage. These paths remain within the S121
  security/settings boundary and do not alter Ent, deployment, billing, or
  unrelated feature code.
