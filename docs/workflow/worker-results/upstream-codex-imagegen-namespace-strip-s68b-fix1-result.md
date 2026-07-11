### DONE: upstream-codex-imagegen-namespace-strip-s68b-fix1

# Worker Result

## Task ID

`upstream-codex-imagegen-namespace-strip-s68b-fix1`

## Status

`done`

## Summary

- Closed the local `OpenAIWSIngressModePassthrough` gap by resolving Codex-client/account strip policy once per session and applying the shared raw strip helper to the first frame and every later text frame.
- First-frame stripping now happens before model capture, fast-policy evaluation, usage metadata initialization, prompt-cache extraction, and actual upstream write.
- Later text-frame stripping happens after the existing non-text early return and before hooks, session-model fallback, fast-policy evaluation, usage metadata updates, and actual relay.
- Added a two-turn passthrough session test whose four captured upstream writes prove both `response.create` frames are stripped while an intervening binary frame and non-image `response.cancel` remain unchanged.
- Added default-allow Codex and non-Codex passthrough preservation assertions, invalid raw JSON behavior, and OAuth HTTP passthrough `lastBody` evidence.

## Changed Files

- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `docs/workflow/worker-results/upstream-codex-imagegen-namespace-strip-s68b-fix1-result.md`

## Commands Run

```text
go test ./internal/service -run "TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Passthrough.*Image.*Strip|TestStripOpenAIImageGenerationToolsFromRawPayload|TestOpenAIGatewayServiceForward_.*OAuth.*Passthrough.*Image.*Strip" -count=1 -> PASS
go test ./internal/service -run "Test.*WSPassthrough.*Effort|TestPassthroughBilling_|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_Passthrough" -count=1 -> PASS
go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_AccountPolicyStrips|Test.*Passthrough.*Image.*Strip|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey" -count=1 -> PASS
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice" -count=1 -> PASS
go test ./internal/service -run "^$" -count=1 -> PASS package compile
go test ./internal/service -run "^TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughImageNamespaceStripAcrossTurns$" -count=3 -> PASS
go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1 -> PASS
go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1 -> PASS
git diff --check -> PASS
allowed-path audit -> PASS
denied-path audit -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.799s
ok github.com/Wei-Shaw/sub2api/internal/service 0.321s
ok github.com/Wei-Shaw/sub2api/internal/service 0.395s
ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.072s
ok github.com/Wei-Shaw/sub2api/internal/service 0.069s [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/service 0.688s
ok github.com/Wei-Shaw/sub2api/internal/service 0.154s
ok github.com/Wei-Shaw/sub2api/internal/service 0.126s
```

## Evidence

- `TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughImageNamespaceStripAcrossTurns` inspects four `openAIWSCaptureConn.writes`: stripped first turn, unchanged binary frame, unchanged `response.cancel`, and stripped second turn. Its `BeforeRequest` hook also receives the stripped second-turn payload.
- `TestOpenAIGatewayServiceForward_OAuthPassthroughImageNamespaceStripUsesForwardedBody` inspects `httpUpstreamRecorder.lastBody`, proving OAuth passthrough forwards stripped bytes while retaining instructions, message input, non-image namespace, and custom `imagegen` function.
- `TestStripOpenAIImageGenerationToolsFromRawPayload/invalid_JSON_returns_original_payload_and_error` proves invalid JSON returns original bytes, `changed=false`, and an error.

## Risks

- No live OpenAI/Codex upstream was used. Actual relay behavior is covered by the in-process WebSocket upstream capture and HTTP upstream recorder.
- The passthrough test fixture delays its second synthetic terminal event by 200 ms so client turn two enters the relay before the fake upstream responds; the focused test passed three consecutive runs.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
