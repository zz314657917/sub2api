---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-06-30 14:22 +08:00
---

# Task Contract

## Task ID
upstream-main-v0139-chat-bridge-guards-s30

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small follow-up port. No external worker is used.

## Goal
Port the local-relevant parts of upstream `ae5e980dd` and `dbdbfb112` so OpenAI `/v1/chat/completions` bridge behavior matches current OpenAI/Codex guard expectations.

## Success Criteria
- `/v1/chat/completions` through `ForwardAsChatCompletions` enforces account `codex_cli_only` the same way `/v1/responses` already does.
- Rejected non-official clients are marked as local policy limited and do not reach upstream.
- OAuth chat-completions bridge does not inject default Codex base instructions for normal Chat Completions converted into Responses.
- OAuth Responses-shaped payloads sent to `/v1/chat/completions` still keep the existing default-instructions behavior.
- No frontend, migrations, Ent, wire, deploy, VERSION, README, or `knowledge/*` files are included in the commit.

## Context
- Repo: `F:/mcplugins/sub2api`
- Upstream references:
  - `ae5e980dd12f7f8061ba8644788b156565e704cb fix(gateway): enforce codex_cli_only restriction on /v1/chat/completions`
  - `dbdbfb11225d2f036c34287ccc8f6028fe2289ed fix: avoid default codex instructions for chat bridge`
- Current worktree contains unrelated dirty proxy/account/frontend/knowledge files. They are outside this sprint.

## Allowed Paths
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `docs/workflow/tasks/upstream-main-v0139-chat-bridge-guards-s30.md`
- `docs/workflow/worker-results/upstream-main-v0139-chat-bridge-guards-s30-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-chat-bridge-guards-s30-qa.md`
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
- `backend/internal/service/wire.go`
- Payment, subscription, keys UI, ops UI, Grok routing, risk-control, OAuth email flow, proxy/account ownership work, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep this sprint to OpenAI chat-completions bridge restrictions and instruction behavior only.
- Do not change raw third-party OpenAI-compatible forwarding semantics beyond the shared restriction gate.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, or knowledge files.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestForwardAsChatCompletions_EnforcesCodexCLIOnlyRestriction|TestForwardAsChatCompletions_OAuthDoesNotInjectDefaultInstructions|TestForwardAsChatCompletions_APIKeyPropagatesPromptCacheKeyInResponsesBody|TestForwardAsChatCompletions_TransportErrorReturnsFailover" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/service/openai_gateway_chat_completions.go backend/internal/service/openai_gateway_chat_completions_test.go docs/workflow/tasks/upstream-main-v0139-chat-bridge-guards-s30.md docs/workflow/worker-results/upstream-main-v0139-chat-bridge-guards-s30-result.md docs/workflow/qa-reports/upstream-main-v0139-chat-bridge-guards-s30-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0139-chat-bridge-guards-s30-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0139-chat-bridge-guards-s30-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementing this requires frontend, Ent, migration, wire, route, handler, repository, deploy, README, or `knowledge/*` changes.
- Stop if chat-completions bridge would lose the intentionally preserved Responses-shaped payload behavior.
- Stop if codex_cli_only rejection cannot be proven to avoid upstream forwarding.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`
