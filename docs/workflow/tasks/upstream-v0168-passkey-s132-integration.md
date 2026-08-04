---
task_id: upstream-v0168-passkey-s132-integration
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Port the S132 Passkey feature onto the current mainline with fail-closed
WebAuthn settings, password confirmation for registration/revocation, and a
new non-conflicting credential migration.

## Success Criteria

- WebAuthn configuration missing/invalid, disabled settings, and backend mode
  restrictions all fail closed before a ceremony can proceed.
- Registration and revocation require current-password confirmation; credential
  request bodies are not captured in audit payloads.
- The source migration is adapted to `203_passkey_credentials.sql` because
  local migration number `199` is already occupied; no existing migration is
  altered.
- Backend and frontend settings/auth/profile flows compile and focused tests
  cover the new public contract.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/go.mod`
- `backend/go.sum`
- `backend/internal/config/config.go`
- `backend/internal/config/webauthn_test.go`
- `backend/internal/handler/**`
- `backend/internal/repository/passkey_*.go`
- `backend/internal/repository/wire.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/server/middleware/{audit_log.go,backend_mode_guard.go}`
- `backend/internal/server/routes/{auth.go,user.go}`
- `backend/internal/service/{audit_log.go,domain_constants.go,passkey.go,passkey_test.go,setting_service.go,settings_view.go,wire.go}`
- `backend/migrations/203_passkey_credentials.sql`
- `deploy/config.example.yaml`
- `frontend/src/{api/**,components/user/profile/ProfilePasskeyCard.vue,i18n/locales/**,stores/{app.ts,auth.ts},types/index.ts,views/{admin/SettingsView.vue,auth/LoginView.vue,user/ProfileView.vue}}`
- `docs/workflow/tasks/upstream-v0168-passkey-s132-integration.md`
- `docs/workflow/qa-reports/upstream-v0168-passkey-s132-integration-qa.md`

## Denied Paths

- Any existing migration, including `backend/migrations/199_*`.
- `Dockerfile*`, `docker-compose*.yml`, `outputs/**`, `output/**`, `knowledge/**`.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, and global memories.
- Any real RP, WebAuthn ceremony, browser credential, database/container,
  provider account, deployment, or production state.

## Acceptance Commands

```powershell
go mod verify
go test ./internal/config ./internal/service ./internal/handler ./internal/server/... -run 'Test(Passkey|BindPasskey|WebAuthn|PublicSettings|BackendMode.*Passkey|Audit.*Passkey)' -count=1
go test ./... -run '^$'
go build ./...
corepack.cmd pnpm --dir frontend exec vitest run src/api/__tests__/passkey.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run build
git diff --check
git ls-files -u
```

## Stop Rules

- Stop if current-main adaptation requires a live migration, real ceremony,
  weakened auth boundary, new unapproved route, or changes outside this list.
- Stop before push, remote deletion, Docker operation, deployment, or any
  production write.

## Contract Review

`PASS / contract-approved`: the feature is opt-in at both file configuration
and persisted setting layers. Source checks establish password confirmation,
bounded handler input, audit redaction, and backend-mode gating. The only
schema adaptation is a new migration `203`, chosen because the local branch
already owns `199`; no existing migration is modified.
