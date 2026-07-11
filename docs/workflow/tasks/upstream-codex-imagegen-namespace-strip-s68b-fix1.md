# Task Contract: upstream-codex-imagegen-namespace-strip-s68b-fix1

## Task ID

`upstream-codex-imagegen-namespace-strip-s68b-fix1`

## Status

`approved`

## Role

You are the Generator worker closing the S68b WebSocket passthrough gap and two evidence gaps found by independent review.

## Goal

Complete the original S68b contract without narrowing it: apply the account-level Codex image declaration `strip` policy to `OpenAIWSIngressModePassthrough` first and follow-up client frames before actual upstream relay, and add missing invalid-raw and OAuth HTTP passthrough evidence.

## Success Criteria

- A Codex client on an OpenAI account with explicit image tool policy `strip` removes flat `image_generation`, `image_gen` namespace declarations, Responses Lite `additional_tools`, and matching image `tool_choice` values from the first WS passthrough `response.create` frame before it reaches the upstream connection.
- The same strip is applied to every later WS passthrough text frame before actual relay; a two-turn capture proves both upstream writes are stripped.
- Non-text frames, non-`response.create` frames without image declarations, non-Codex clients, and the default `allow` policy remain unchanged.
- Existing WS fast-policy ordering, blocking/filtering, model fallback, hooks, usage metadata, billing metadata, and session routing remain unchanged apart from the stripped request payload.
- Invalid JSON passed to `stripOpenAIImageGenerationToolsFromRawPayload` returns the original bytes, `changed=false`, and a non-nil error; valid non-image JSON remains byte-for-byte unchanged.
- OAuth HTTP passthrough with valid non-empty `instructions` proves through the upstream recorder that the actual forwarded body is stripped, while ordinary content and non-image tools remain.
- The original S68b acceptance commands and S67 preservation checks remain green.

## Context

- Initial S68b worker commit: `36fe5644165e06a1728cfd9ede98985835654181`.
- Independent review found that `ProxyResponsesWebSocketFromClient` returns to `proxyResponsesWebSocketV2Passthrough` before parsed ingress stripping.
- The passthrough adapter applies fast policy to the first frame and wraps later client frames, but neither location currently applies the S68b strip policy.
- This fix expands ownership only enough to cover that local adapter path and the two missing evidence cases.

## Allowed Paths

- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `docs/workflow/worker-results/upstream-codex-imagegen-namespace-strip-s68b-fix1-result.md`

## Denied Paths

- All other business and test paths, including `codex_image_generation_bridge.go`, `openai_gateway_service.go`, `openai_codex_transform.go`, `image_generation_intent.go`, and `backend/internal/pkg/apicompat/**`; the initial S68b implementation is frozen during fix1.
- Frontend, API/types/DTOs, migrations, billing/pricing/accounting, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Reuse `stripOpenAIImageGenerationToolsFromRawPayload`; do not create a second parser or policy flag.
- Determine Codex-client and account policy once for the passthrough session, using the same header/config semantics as existing WS code.
- Apply stripping before fast-policy evaluation and before usage metadata is extracted from the forwarded frame, so metadata reflects the actual upstream payload.
- Preserve the passthrough adapter's first-frame and follow-up-frame symmetry.
- Do not change routing, retry, connection ownership, hook ordering, response relay, or billing calculations.
- Do not repair unrelated full-suite drift.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Passthrough.*Image.*Strip|TestStripOpenAIImageGenerationToolsFromRawPayload|TestOpenAIGatewayServiceForward_.*OAuth.*Passthrough.*Image.*Strip" -count=1
go test ./internal/service -run "Test.*WSPassthrough.*Effort|TestPassthroughBilling_|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_Passthrough" -count=1
go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_AccountPolicyStrips|Test.*Passthrough.*Image.*Strip|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey" -count=1
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice" -count=1
go test ./internal/service -run "^$" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-imagegen-namespace-strip-s68b-fix1-result.md` with first line `### DONE: upstream-codex-imagegen-namespace-strip-s68b-fix1`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit only fix1 changes on the existing S68b worker branch and return the commit hash.

## Stop Rules

- Stop if correct first/follow-up relay coverage requires modifying any path outside Allowed Paths.
- Stop if fast-policy block/filter behavior, usage metadata, hooks, or model fallback changes beyond consuming the stripped payload.
- Stop if default `allow`, non-Codex, non-image namespace, custom `imagegen`, or non-text frame behavior changes.
- Stop if either captured upstream WS write still contains an image declaration or matching image tool choice.
- Do not repair unrelated unit-suite drift.
