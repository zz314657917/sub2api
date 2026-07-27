---
task_id: upstream-v0165-chatgpt-live-s116
status: done
owner: Codex
qa_mode: runtime
---

# Task Contract

## Goal

Add the ChatGPT Live realtime gateway for compatible OpenAI OAuth accounts,
including SDP negotiation, WebSocket sideband forwarding, Redis-backed call
lifecycle, concurrency leases, usage accounting, and the Codex-compatible
route aliases used by clients.

## Success Criteria

- Group administrators can enable/disable Live through a persisted
  `allow_live` capability with safe default `false` and matching API/UI types.
- `POST /v1/live` creates a call after auth, capability, concurrency, and SDP
  validation; `GET /v1/live/:call_id` returns the negotiated call metadata.
- `/backend-api/codex/realtime/calls` aliases are wired for ChatGPT/Codex
  clients without changing existing `/v1/responses` behavior.
- Redis stores call state and an expiring concurrency lease; disconnect,
  expiry, and lost-lease paths release or terminate calls deterministically.
- Live sideband WebSocket forwarding preserves request metadata and supports
  the local OpenAI OAuth account transport.
- Completed/failed Live calls create usage records with `request_type=live`,
  including the session identifier when supplied.
- Focused handler/service/repository tests, Ent generation checks, migration
  checks, frontend typecheck/build, and runtime smoke pass.

## Scope Boundary

- Adapt upstream `e6eb23eaa`, `988d4b577`, `ec23716ee`, `db6fbdbf2`, and
  `7acc13a29` to the local topology.
- Do not change unrelated gateway protocols, billing rates, account selection,
  deployment, containers, or existing user-facing usage semantics.
- Do not enable Live by default for existing groups or bypass existing auth,
  group, subscription, or concurrency checks.

## Allowed Paths

- backend/ent/schema/group.go
- backend/ent/** (generated output required by schema change)
- backend/internal/config/config.go
- backend/internal/handler/admin/group_handler.go
- backend/internal/handler/openai_live.go
- backend/internal/handler/openai_live_test.go
- backend/internal/handler/dto/types.go
- backend/internal/handler/dto/mappers.go
- backend/internal/repository/api_key_repo.go
- backend/internal/repository/concurrency_cache.go
- backend/internal/repository/concurrency_cache_integration_test.go
- backend/internal/repository/concurrency_cache_live_test.go
- backend/internal/repository/gateway_cache.go
- backend/internal/repository/gateway_cache_live_test.go
- backend/internal/repository/group_repo.go
- backend/internal/service/account.go
- backend/internal/service/admin_group.go
- backend/internal/service/admin_group_duplicate.go
- backend/internal/service/group.go
- backend/internal/service/openai_live.go
- backend/internal/service/openai_live_lifecycle_test.go
- backend/internal/service/openai_live_test.go
- backend/internal/service/openai_live_types.go
- backend/internal/service/usage_log.go
- backend/internal/server/routes/gateway.go
- backend/migrations/196_allow_live_usage_request_type.sql
- backend/migrations/197_add_group_allow_live.sql
- deploy/config.example.yaml
- frontend/src/components/admin/usage/UsageFilters.vue
- frontend/src/components/admin/usage/UsageTable.vue
- frontend/src/i18n/locales/en/admin/overview.ts
- frontend/src/i18n/locales/en/dashboard.ts
- frontend/src/i18n/locales/zh/admin/overview.ts
- frontend/src/i18n/locales/zh/dashboard.ts
- frontend/src/types/index.ts
- frontend/src/utils/errorBadges.ts
- frontend/src/utils/usageRequestType.ts
- frontend/src/views/admin/GroupsView.vue
- frontend/src/views/admin/UsageView.vue
- frontend/src/views/user/UsageView.vue
- docs/workflow/tasks/upstream-v0165-chatgpt-live-s116.md

## Denied Paths

- backend/migrations/195_add_usage_log_session_id.sql
- backend/internal/handler/** except the listed Live/admin/DTO files
- backend/internal/service/** except the listed Live/account/group/usage files
- knowledge/**
- outputs/**
- Dockerfile*
- unrelated frontend pages and components

## Constraints

- Live is opt-in and must fail closed when capability/configuration is absent.
- Reuse the local Redis client, OAuth token provider, proxy handling, usage
  logger, and concurrency conventions; do not introduce a second runtime.
- Keep generated Ent files reproducible from the schema change.
- Preserve existing dirty S114/group-buy/knowledge/output changes.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go generate ./ent
go test ./internal/handler ./internal/service ./internal/repository ./internal/server/routes -run "Test.*Live|Test.*Concurrency.*Live|Test.*GatewayCache.*Live" -count=1
go test ./internal/server -run "Test.*Live|Test.*Realtime" -count=1
cd F:/mcplugins/sub2api/frontend
npm.cmd run typecheck
npm.cmd run build
cd F:/mcplugins/sub2api
git diff --check
```

## Output

- Source/generated code, migrations, focused tests, and a QA report with
  runtime smoke evidence and explicit unverified upstream/network risks.

## Stop Rules

- Stop and split if Live requires a new authentication model, public billing
  contract, or a migration that cannot coexist with local migration history.
- Stop if generated Ent output or existing concurrency semantics cannot be
  preserved without broad unrelated regeneration.
