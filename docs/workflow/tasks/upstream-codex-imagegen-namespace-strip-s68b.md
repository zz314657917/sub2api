# Task Contract: upstream-codex-imagegen-namespace-strip-s68b

## Task ID

`upstream-codex-imagegen-namespace-strip-s68b`

## Status

`draft`

## Role

You are the Generator worker adapting the Codex `image_gen` namespace strip fix to the local monolithic gateway.

## Goal

Adapt upstream `d3a1835ed` on top of the completed S68a policy: when an OpenAI account's Codex explicit image tool policy is `strip`, remove flat `image_generation` and Codex `image_gen` namespace declarations consistently from managed HTTP, HTTP passthrough, Responses Lite `additional_tools`, tool choices, and WebSocket ingress/raw payloads.

## Success Criteria

- Image intent detection recognizes flat `image_generation`, top-level namespace `{type:"namespace",name:"image_gen"}`, Responses Lite `input[].additional_tools`, and equivalent `tool_choice` shapes.
- Strip mode removes flat and namespace image declarations from top-level `tools` and `input[].additional_tools`.
- Empty `additional_tools` carriers are removed after their last image declaration is stripped; carriers with non-image tools remain.
- Matching image tool choices are removed, including namespace choices by `name` or `namespace` and nested `tool` objects.
- Ordinary functions, non-image namespaces, message/input items, `tool_choice:"auto"`, and custom functions merely named `imagegen` remain unchanged.
- Managed HTTP, HTTP passthrough, API-key passthrough, WS ingress, and Spark paths use consistent stripping without changing default `allow` behavior.
- Passthrough uses the stripped body for actual upstream forwarding, not only for accounting or intent inspection.
- S67 custom/tool-search/namespace MCP fallback behavior remains unchanged and its focused tests pass.
- Stripping is idempotent.

## Context

- Upstream implementation: `d3a1835ed fix(image): strip Codex image_gen namespace declarations`.
- Required policy/UI prerequisite S68a is complete and QA PASS at `d4b04b96b`.
- Upstream split files `openai_gateway_forward.go`, `openai_ws_forwarder_ingress.go`, and `openai_ws_forwarder_v2.go` map to local `openai_gateway_service.go` and `openai_ws_forwarder.go`.
- S67 modified those local monoliths; adapt against current HEAD and do not replay stale whole-file patches.

## Allowed Paths

- `backend/internal/service/image_generation_intent.go`
- `backend/internal/service/image_generation_intent_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `docs/workflow/worker-results/upstream-codex-imagegen-namespace-strip-s68b-result.md`

## Denied Paths

- `backend/internal/service/codex_image_generation_bridge.go` and its tests; S68a policy semantics are frozen.
- `backend/internal/pkg/apicompat/**`, message/fallback service files, and S67 stream lifecycle files.
- Frontend, API/types/DTOs, migrations, billing/pricing/accounting, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Keep exact image declaration matching: namespace name must be `image_gen`; do not infer image intent from arbitrary child function names.
- Preserve default `allow` and ordinary non-Codex behavior.
- Reuse the S68a account policy; do not introduce another flag or configuration surface.
- Adapt to local monolithic files without importing upstream file-splitting refactors.
- Keep passthrough routing, retries, billing, and response handling unchanged apart from forwarding the stripped request body.
- Do not change standalone image endpoints or image prices/accounting.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_AccountPolicyStrips|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey" -count=1
go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice" -count=1
go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1
go test ./internal/service -run "^$" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-imagegen-namespace-strip-s68b-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if the S68a policy prerequisite or its PASS report is missing.
- Stop if implementation needs frontend, API/DTO, `codex_image_generation_bridge.go`, apicompat, fallback, billing, or accounting changes.
- Stop if a non-image namespace or custom `imagegen` function would be removed.
- Stop if default `allow` behavior changes or passthrough requires routing/retry changes beyond the request body.
- Stop if the local monolithic mapping is ambiguous; report the exact ownership conflict instead of importing broad upstream refactors.
- Do not repair unrelated full-suite drift.
