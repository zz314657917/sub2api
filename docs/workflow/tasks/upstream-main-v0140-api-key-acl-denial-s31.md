---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-06-30 15:12 +08:00
---

# Task Contract

## Task ID
upstream-main-v0140-api-key-acl-denial-s31

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small upstream port. No external worker is used.

## Goal
Port the local-relevant part of upstream `56c62c59c` so API key ACL denial responses include the resolved client IP while preserving the existing trusted-proxy boundary.

## Success Criteria
- API key whitelist/blacklist denial returns `ACCESS_DENIED` with `Access denied. Your IP is <client-ip>`.
- Spoofed forwarded headers are not trusted when Gin trusted proxies are not configured.
- Trusted reverse-proxy configuration can still report the forwarded client IP via Gin `ClientIP()`.
- Ops client business-limited marking for IP restriction remains unchanged.
- No frontend, migrations, Ent, wire, deploy, VERSION, README, payment, OAuth, Grok, or `knowledge/*` files are included in the commit.

## Context
- Repo: `F:/mcplugins/sub2api`
- Current local anchor: `v0.1.126-522-g70c89f2ad`
- Current upstream anchor after refresh: `v0.1.140-1-g89b2d63ef`
- Upstream reference:
  - `56c62c59c813d8232c037b2f49c910982b672a24 fix(auth): include client ip in acl denial message`
- Candidate review also confirmed local equivalence for:
  - `82576e0a3 fix(auth): stop swallowing email auth identity create error via shadowed err`
  - `65fa72892 fix(openai): fail over on chat transport errors`
- Latest `v0.1.140` tail commits touch frontend admin tables, OAuth completion flow, payment refund pending semantics, Grok/platform quota migrations, sponsors, and VERSION. They are intentionally out of scope for this sprint.

## Allowed Paths
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `docs/workflow/tasks/upstream-main-v0140-api-key-acl-denial-s31.md`
- `docs/workflow/worker-results/upstream-main-v0140-api-key-acl-denial-s31-result.md`
- `docs/workflow/qa-reports/upstream-main-v0140-api-key-acl-denial-s31-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/server/routes/**`
- `backend/internal/service/**`
- Payment, subscription, keys UI, ops UI, Grok routing, OAuth email flow, proxy/account ownership work, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep this sprint to API key ACL denial response behavior only.
- Preserve existing `ip.GetTrustedClientIP(c)` behavior and Gin trusted-proxy semantics.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, service, handler, or knowledge files.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial|TestAPIKeyAuthIPRestrictionUsesForwardedClientIPWhenProxyTrusted" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/server/middleware/api_key_auth.go backend/internal/server/middleware/api_key_auth_test.go docs/workflow/tasks/upstream-main-v0140-api-key-acl-denial-s31.md docs/workflow/worker-results/upstream-main-v0140-api-key-acl-denial-s31-result.md docs/workflow/qa-reports/upstream-main-v0140-api-key-acl-denial-s31-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0140-api-key-acl-denial-s31-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0140-api-key-acl-denial-s31-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementing this requires frontend, Ent, migration, wire, route, handler, repository, service, deploy, README, VERSION, payment, OAuth, Grok, or `knowledge/*` changes.
- Stop if the patch requires trusting raw forwarded headers outside Gin trusted-proxy configuration.
- Stop if ACL denial no longer marks ops client business limited.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`
