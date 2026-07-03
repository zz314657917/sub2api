---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 10:40 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-antigravity-reasoning-params-s41

## Role
Codex acts as Planner and Final Evaluator. Implementation may be done by Codex directly only after this contract is approved. QA may be run by Codex directly after implementation.

## Goal
Port upstream `v0.1.143` commit `f5b296127 fix: Handle invalid arguments correctly for Gemini reasoning models`: Antigravity Gemini reasoning models must not receive unsupported generation parameters or forced empty `toolConfig` when there are no tools.

## Success Criteria
- Mark the upstream-listed Gemini reasoning models in the local Antigravity model table.
- Add a local helper to identify Gemini reasoning models by requested/mapped model ID.
- For Gemini reasoning models with no tools, omit the forced `toolConfig`.
- For Gemini reasoning models, omit `stopSequences`, `temperature`, `topP`, and `topK` from `generationConfig`.
- Keep non-reasoning Gemini models unchanged, including default `toolConfig`, default stop sequences, and pass-through of temperature/topP/topK.
- Add focused unit tests covering reasoning and non-reasoning behavior.
- Do not touch gateway, billing, usage, frontend, Ent, migrations, deploy, knowledge, or unrelated dirty files.

## Context
- Repo: `F:/mcplugins/sub2api`
- Previous completed Sprint: `upstream-main-v0143-claude-oauth-payload-s40`
- Latest release refreshed on 2026-07-03: GitHub releases list `Sub2API 0.1.143` as latest, tag `v0.1.143` / commit `9caa3c9c5`.
- Candidate commit: `f5b296127 fix: Handle invalid arguments correctly for Gemini reasoning models`.
- Current precheck:
  - `backend/internal/pkg/antigravity/claude_types.go` is clean.
  - `backend/internal/pkg/antigravity/request_transformer.go` is clean.
  - `backend/internal/pkg/antigravity/request_transformer_test.go` is clean.
  - Upstream `f5b296127` touches only `claude_types.go` and `request_transformer.go`; this contract also allows local focused tests in `request_transformer_test.go`.
- Deferred `v0.1.143` candidates include:
  - `e236bff1e` user model stats by requested model, because usage area overlaps higher-risk local changes.
  - `d0b8760eb` OpenAI plan type from inactive workspaces, because it changes broader OpenAI subscription behavior.
  - `df59b8b96` OpenAI subscription expiration persistence, because it is frontend product-state work and should be separate.
  - `a5638a4e5` Codex session import identity, because it touches backend handler/frontend API flow and needs a separate contract.
  - `c797159bf` compact skip, because it touches current dirty `openai_gateway_service.go`.
  - `a5781fe31` keepalive and `7869b7fe3` Anthropic Bearer auth, because they touch current dirty `gateway_service.go` and broader account/gateway surfaces.
  - `c4128580f` count_tokens latest scope, because it rewrites a larger count_tokens surface and dependency set.

## Allowed Paths
- `backend/internal/pkg/antigravity/claude_types.go`
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md`
- `docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md`
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
- Keep this Sprint Antigravity-only.
- Preserve non-reasoning Gemini behavior exactly unless a focused test proves the unchanged path.
- If implementation requires touching gateway/frontend/billing/usage/dependency files, stop and split a new Sprint.
- Do not stage existing dirty files outside allowed paths.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_ReasoningModelOmitsInvalidArgs|TestBuildGenerationConfig_ReasoningModelOmitsUnsupportedParams|TestTransformClaudeToGeminiWithOptions_PreservesWebSearchAlongsideFunctions|TestTransformClaudeToGeminiWithOptions_MessageRoles" -count=1

cd F:/mcplugins/sub2api
git diff --check -- backend/internal/pkg/antigravity/claude_types.go backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md docs/workflow/status.md docs/workflow/main-log.md
git diff --cached --name-only | rg "^(backend/internal/service/gateway_service.go|backend/internal/service/openai_gateway_service.go|backend/internal/service/openai_gateway_count_tokens.go|backend/internal/service/openai_gateway_count_tokens_test.go|backend/go.mod|backend/go.sum|backend/ent/|backend/migrations/|backend/cmd/server/|backend/internal/service/billing_|backend/internal/service/usage_|backend/internal/repository/usage_|backend/internal/repository/welfare_|backend/internal/service/welfare_|backend/internal/payment/|frontend/|deploy/|knowledge/|assets/|README|README_|\.github/)" || echo NO_DENIED_PATHS
```

## Output
- Backend Antigravity code diff in allowed paths only.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`

## Stop Rules
- Stop if the target Antigravity files become dirty before implementation.
- Stop if tests require gateway, frontend, billing, Ent, migration, deploy, or dependency changes.
- Stop if reasoning-model detection cannot be kept local to Antigravity model metadata.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0.03`

## Review Result
- Reviewed at: 2026-07-03 10:40 +08:00.
- Verdict: approved.
- Reason: required P/G/E contract fields are present; allowed paths are limited to clean Antigravity files plus workflow artifacts; denied paths explicitly protect current gateway, billing, usage, frontend, Ent, migration, deploy, knowledge, and unrelated dirty files.
