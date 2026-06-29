### DONE: upstream-main-v0139-backend-compat-s23

## Summary

Implemented S23 backend-only compatibility patches from post-`v0.1.138` / `v0.1.139` upstream candidates without merging `upstream/main`.

## Ported

- `0da1fe28e`: OpenAI image output accounting now ignores text-only `data` items and empty `image_generation.completed` events.
- `9491de0a3`: OpenAI image model refusal text with no image output now returns a non-retryable 400 `content_policy_violation`.
- `cc7612bdb`: OpenAI `server_is_overloaded` and `slow_down` codes now enter transient failover handling.
- `8a7269f53`: streaming `response.failed` events are sanitized before client delivery, preserving error details while removing verbose response internals.
- `40c825273`: Responses-to-Anthropic custom and unknown tool schemas are normalized to Anthropic-compatible object schemas.
- `e5f7836bf`: Codex image bridge requests set `tool_choice: "auto"` when an image-generation tool is present and no explicit tool choice exists, including WS ingress.

## Skipped

- `82553c4dc` quota platform billing remains a separate candidate.
- `da810c3b4` Keys unlimited reactivation remains a separate frontend/admin candidate.
- `7c2fee6c9` fallback pricing log dedup remains low-priority follow-up.
- Grok, codex_cli_only full fingerprint hardening, model-not-found 404 handler refactor, payment/subscription/Ops/Keys frontend, VERSION/sponsor/README/CI/deploy remained out of scope.

## Changed Files

- `backend/internal/service/image_output_accounting.go`
- `backend/internal/service/image_output_accounting_test.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_incomplete_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_tools_test.go`
- `docs/workflow/tasks/upstream-main-v0139-backend-compat-s23.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run

- `go test ./internal/service -run "TestOpenAIImageOutputCounter|TestImagesOAuthNonStreaming_ContentRefusalReturns400NoRetry|TestExtractModelRefusal_EmptyWhenNoText|TestIsOpenAITransientProcessingError|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIStreamingPassthroughResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIGatewayServiceForward_CodexImageInjectionRespectsGroupCapability|TestOpenAIGatewayServiceForward_ChannelBridgeOverrideEnablesCodexInjection|TestOpenAIGatewayServiceForward_CodexBridgePreservesExistingToolChoice|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool" -count=1`
- `go test ./internal/pkg/apicompat -run "TestResponsesToAnthropic_.*Tool|TestResponsesToAnthropic_Custom" -count=1`
- `git diff --check`

## Risks

- No real OpenAI OAuth, Responses WS, or image upstream traffic was run locally; evidence is targeted code-level/runtime tests.
- Raw working-tree denied-path audit still reports pre-existing dirty `knowledge/*` files from outside S23. Those files are excluded from S23 staging and commit.
