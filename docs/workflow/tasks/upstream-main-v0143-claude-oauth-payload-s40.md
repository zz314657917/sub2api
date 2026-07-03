---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 10:32 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-claude-oauth-payload-s40

## Role
Codex acts as Planner and Final Evaluator. Implementation may be done by Codex directly only after this contract is approved. QA may be run by Codex directly after implementation.

## Goal
Port upstream `v0.1.143` commit `5bd9368ab fix claude oauth token exchange payload`: Claude OAuth setup-token code exchange must not send the non-standard `expires_in` field in the token exchange request body.

## Success Criteria
- Remove `expires_in` from the Claude OAuth token exchange request body for setup-token flows.
- Keep regular Claude OAuth exchange behavior unchanged.
- Update the Claude OAuth repository test so setup-token exchange asserts `expires_in` is omitted.
- Do not touch gateway, billing, usage, frontend, Ent, migrations, deploy, knowledge, or unrelated dirty files.
- Explicitly leave broader `v0.1.143` feature chains deferred.

## Context
- Repo: `F:/mcplugins/sub2api`
- Previous completed Sprint: `upstream-main-v0142-frontend-api-base-s39a`
- Latest release refreshed on 2026-07-03: `v0.1.143` / tag commit `9caa3c9c5`.
- Candidate commit: `5bd9368ab fix claude oauth token exchange payload`.
- Current precheck:
  - `backend/internal/repository/claude_oauth_service.go` is clean.
  - `backend/internal/repository/claude_oauth_service_test.go` is clean.
  - Upstream `5bd9368ab` touches only those two files.
- Deferred `v0.1.143` candidates include:
  - OpenAI WS `http_bridge` ingress mode and setup-token editing.
  - subscription group peak-rate multiplier chain.
  - usage IP geolocation.
  - subscription revoke restore.
  - admin group column settings.
  - OpenAI count_tokens latest scope handling, which rewrites a large count_tokens service/test surface and adds dependencies.
  - Codex image bridge compact skip, because it touches current dirty `openai_gateway_service.go`.
  - Claude Code keepalive and Anthropic Bearer auth, because they touch current dirty `gateway_service.go` and broader frontend/account surfaces.

## Allowed Paths
- `backend/internal/repository/claude_oauth_service.go`
- `backend/internal/repository/claude_oauth_service_test.go`
- `docs/workflow/tasks/upstream-main-v0143-claude-oauth-payload-s40.md`
- `docs/workflow/worker-results/upstream-main-v0143-claude-oauth-payload-s40-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-claude-oauth-payload-s40-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_count_tokens_test.go`
- `backend/go.mod`
- `backend/go.sum`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/service/billing_*`
- `backend/internal/service/usage_*`
- `backend/internal/repository/usage_*`
- `backend/internal/repository/welfare_*`
- `backend/internal/service/welfare_*`
- `backend/internal/payment/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `assets/**`
- `README*`
- `.github/**`
- Any unlisted dirty file.

## Constraints
- Do not merge/rebase `v0.1.143` or `upstream/main`.
- Do not cherry-pick broader release content.
- Keep this Sprint repository-only and Claude OAuth-only.
- Do not alter token refresh semantics, storage semantics, or setup-token lifetime policy outside the outbound token exchange payload.
- If implementation requires touching denied gateway/frontend/billing files, stop and split a new Sprint.
- Do not stage existing dirty files outside allowed paths.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/repository -run "TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1

cd F:/mcplugins/sub2api
git diff --check -- backend/internal/repository/claude_oauth_service.go backend/internal/repository/claude_oauth_service_test.go docs/workflow/tasks/upstream-main-v0143-claude-oauth-payload-s40.md docs/workflow/worker-results/upstream-main-v0143-claude-oauth-payload-s40-result.md docs/workflow/qa-reports/upstream-main-v0143-claude-oauth-payload-s40-qa.md docs/workflow/status.md docs/workflow/main-log.md
git diff --cached --name-only | rg "^(backend/internal/service/gateway_service.go|backend/internal/service/openai_gateway_service.go|backend/internal/service/openai_gateway_count_tokens.go|backend/internal/service/openai_gateway_count_tokens_test.go|backend/go.mod|backend/go.sum|backend/ent/|backend/migrations/|backend/cmd/server/|backend/internal/service/billing_|backend/internal/service/usage_|backend/internal/repository/usage_|backend/internal/repository/welfare_|backend/internal/service/welfare_|backend/internal/payment/|frontend/|deploy/|knowledge/|assets/|README|README_|\.github/)" || echo NO_DENIED_PATHS
```

## Output
- Backend repository code diff in allowed Claude OAuth paths only.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-claude-oauth-payload-s40-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-claude-oauth-payload-s40-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`

## Stop Rules
- Stop if the target files become dirty before implementation.
- Stop if tests require gateway, frontend, billing, Ent, migration, deploy, or dependency changes.
- Stop if the setup-token behavior depends on a hidden server-side lifetime contract not visible in this repository path.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0.03`

## Review Result
- Reviewed at: 2026-07-03 10:32 +08:00.
- Verdict: approved.
- Reason: required P/G/E contract fields are present; allowed paths are limited to the clean Claude OAuth repository files and workflow artifacts; denied paths explicitly protect current gateway, billing, usage, frontend, Ent, migration, deploy, knowledge, and unrelated dirty files.
