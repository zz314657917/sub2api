### DONE: upstream-codex-imagegen-namespace-strip-s68b

# Worker Result

## Task ID

`upstream-codex-imagegen-namespace-strip-s68b`

## Status

`done`

## Summary

- Extended exact image intent detection and stripping to flat `image_generation`, top-level `image_gen` namespaces, Responses Lite `input[].additional_tools`, and matching direct/nested `tool_choice` shapes.
- Added one idempotent map implementation plus a shared raw payload adapter. Empty image-only `additional_tools` carriers are removed; mixed carriers retain their non-image tools.
- Managed HTTP, HTTP/API-key passthrough, parsed WS ingress, and Spark now apply the same namespace-aware stripping. HTTP passthrough replaces both `body` and the actual `originalBody` passed to `forwardOpenAIPassthrough`.
- Preserved non-image namespaces, normal functions, custom functions named `imagegen`, `tool_choice: "auto"`, default `allow`, non-Codex behavior, and S67 custom/tool-search/namespace fallback behavior.
- Preserved the local Spark rule: when stripping leaves ordinary top-level tools, a matching Spark image choice becomes `auto`; image-only Spark requests remove the choice.

## Changed Files

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

## Commands Run

```text
go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_AccountPolicyStrips|Test.*Passthrough.*Image.*Strip|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey" -count=1 -> PASS
go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1 -> PASS
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice" -count=1 -> PASS
go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1 -> PASS
go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestApplyCodexOAuthTransform_.*ImageGeneration" -count=1 -> PASS
go test ./internal/service -run "^$" -count=1 -> PASS package compile
git diff --check -> PASS
allowed-path audit -> PASS
denied-path audit -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.582s
ok github.com/Wei-Shaw/sub2api/internal/service 0.131s
ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.851s
ok github.com/Wei-Shaw/sub2api/internal/service 0.120s
ok github.com/Wei-Shaw/sub2api/internal/service 5.599s
ok github.com/Wei-Shaw/sub2api/internal/service 0.077s [no tests to run]
```

## Forwarded-Body Evidence

- `TestOpenAIGatewayServiceForward_PassthroughImageNamespaceStripForAPIKey` inspects `httpUpstreamRecorder.lastBody`, proving API-key passthrough forwards the stripped body rather than only using it for intent/accounting.
- `TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_ImageNamespaceStripByAccountPolicy` inspects `openAIWSCaptureConn.writes`, proving parsed WS ingress forwards the stripped namespace/additional-tools payload.

## Risks

- No live OpenAI/Codex upstream was used; HTTP forwarding is covered by the local upstream recorder and WS forwarding by the in-process capture connection.
- Local `IngressModePassthrough` enters `openai_ws_v2_passthrough_adapter.go` before `parseClientPayload`. That file is outside this contract's allowed paths, so this worker intentionally does not change that mode. This contract implements the upstream `d3a1835ed` equivalent for parsed `ctx_pool` / `shared` / `dedicated` WS ingress; full WS passthrough-mode stripping requires a follow-up contract owning the adapter.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`, within the approved upstream-equivalent parsed WS ingress ownership; the explicit adapter gap is documented above
- stop_rules_triggered: `no`

## Blocked Reason

- None.
