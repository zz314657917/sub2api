# Task Contract: upstream-gpt56-max-effort-s67a

## Task ID

`upstream-gpt56-max-effort-s67a`

## Status

`approved`

## Role

You are the Generator worker for GPT-5.6 reasoning-effort compatibility. Implement only the listed paths.

## Goal

Adapt upstream GPT-5.6 `max` reasoning effort and model-candidate usage metadata fixes across the local Responses, raw Chat Completions, HTTP/WS bridge, and WS passthrough paths. The messages fallback file is owned by S67b and must not be modified here.

## Success Criteria

- Explicit `max` is preserved for GPT-5.6 Sol/Terra/Luna, including aliases, mapped models, and suffix variants.
- Non-GPT-5.6 models retain the existing `max -> xhigh` normalization.
- OpenAI OAuth `/responses/compact` downgrades GPT-5.6 `max` to `xhigh`; ordinary Responses and API-key compact requests preserve it.
- Effort metadata derives from ordered model candidates so suffix-bearing requested models are not lost after upstream normalization.
- HTTP, WS v2, WS HTTP bridge, and passthrough usage metadata use the mapped and requested model candidates correctly.
- Targeted GPT-5.6 max and reasoning candidate tests pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream references: `80b3d4c1f`, `c3ae5fc3c`, `b9b013a08`.
- Reuse the local `isOpenAIGPT56Model`/alias helper rather than introducing a duplicate declaration.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/service/openai_gpt56_max_test.go`
- `backend/internal/service/openai_reasoning_effort_candidates_test.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `backend/internal/service/openai_model_alias.go`
- `backend/internal/service/openai_model_alias_test.go`
- `docs/workflow/worker-results/upstream-gpt56-max-effort-s67a-result.md`

## Denied Paths

- `backend/internal/service/openai_gateway_messages.go` and `backend/internal/service/openai_gateway_responses_chat_fallback*.go` are owned by S67b.
- All other paths, including billing/pricing, migrations, frontend, payment, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Do not change charging formulas, model prices, routing decisions, or request retry behavior.
- Do not duplicate `isOpenAIGPT56Model`.
- Preserve existing non-GPT effort semantics.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|TestExtractOpenAIReasoningEffortFromBody|Test.*WSPassthrough.*Effort" -count=1
go test ./internal/service -run "TestOpenAIGatewayServiceForward.*GPT56|TestNormalizeOpenAICodexCompactReasoningEffort" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-gpt56-max-effort-s67a-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes on the assigned branch and return the commit hash.

## Stop Rules

- Stop if implementation requires billing/pricing changes, duplicate model helpers, or any path outside Allowed Paths.
- Stop if preserving `max` would alter non-GPT-5.6 behavior.
- Do not repair unrelated unit-suite drift.
