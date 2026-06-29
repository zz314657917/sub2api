---
task_id: upstream-main-v0139-openai-count-tokens-s26
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-29 22:40 +08:00
---

# Task Contract: upstream v0.1.139 OpenAI count_tokens S26

## Goal

Port upstream `7a38c6621` OpenAI `/v1/messages/count_tokens` bridge so OpenAI groups can serve Anthropic-compatible count_tokens by calling OpenAI `/v1/responses/input_tokens`.

## Success Criteria

- OpenAI groups route `POST /v1/messages/count_tokens` to `OpenAIGatewayHandler.CountTokens` instead of returning local 404.
- The handler validates request body/model, billing eligibility, group messages-dispatch permission, and account selection without recording usage.
- The service converts Anthropic Messages count_tokens input to OpenAI Responses input_tokens body and returns `{"input_tokens": n}`.
- API key accounts honor custom OpenAI `base_url`; OAuth accounts use the public OpenAI platform input_tokens URL and map unsupported auth/404 to Anthropic-format 404.
- Non-OpenAI groups keep existing `Gateway.CountTokens` behavior.
- Targeted route/service tests, `git diff --check`, and staged denied-path audit pass.

## Allowed Paths

- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/service/openai_endpoint_url.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_count_tokens_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `docs/workflow/tasks/upstream-main-v0139-openai-count-tokens-s26.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/upstream-main-v0139-openai-count-tokens-s26-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-openai-count-tokens-s26-qa.md`

## Denied Paths

- Existing dirty `knowledge/*` files, unless separately requested.
- `frontend/*`, public pages, Studio/Canvas, Model Plaza, payment pages, Ops/Keys UI.
- `backend/ent/*`, `backend/migrations/*`, wire generation, VERSION, README, sponsors, CI/deploy.
- OpenAI quota headroom scheduler, API base fetch, keys column settings, subscription/payment display, Ops system log key id, and product UI batches.

## Constraints

- Do not record usage or take normal user streaming/concurrency slots for count_tokens.
- Do not change Anthropic/Gemini/Antigravity count_tokens behavior.
- Keep route changes local to OpenAI group handling; do not introduce upstream helper functions that do not exist locally unless necessary.
- Do not add Ent/migration/wire or frontend changes.
- Do not stage or commit unrelated dirty files.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic" -count=1
go test ./internal/server/routes -run "TestGatewayRoutesOpenAICountTokensPathIsRegistered|TestGatewayRoutesNonOpenAICountTokensPathStillRegistered" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --cached --name-only | rg "^(knowledge/|frontend/|backend/ent/|backend/migrations/|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|deploy/|README|README_|assets/partners/)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in allowed backend and workflow paths only.
- Worker-style result and QA report.
- Final report with commit hash, pushed status, tests run, and remaining dirty file note.

## Stop Rules

- Stop if the bridge requires schema/migration/wire/frontend changes.
- Stop if local account selection or billing dependencies require broad handler refactors.
- Stop if tests require live OpenAI/OAuth upstreams.
