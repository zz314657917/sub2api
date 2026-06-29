---
task_id: upstream-main-v0139-backend-compat-s23
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-29 20:45 +08:00
---

# Task Contract: upstream v0.1.139 backend compatibility S23

## Goal

Port the next low-risk backend protocol and gateway compatibility fixes from `upstream/main` after `v0.1.138`, without wholesale merging upstream or changing local product surfaces.

## Success Criteria

- Text-only `/v1/responses` payloads are not counted as image outputs when a stray `data` array or empty `image_generation.completed` item appears.
- OpenAI image model text refusals with no image output return a non-retryable 400 content-policy error instead of triggering same-account retry/failover.
- OpenAI `server_is_overloaded` / `slow_down` error codes are treated as transient processing errors and can fail over.
- Streaming `response.failed` events sent to clients are sanitized to remove verbose request/response internals while preserving the error.
- Responses-to-Anthropic conversion normalizes `custom` tools into valid Anthropic object schemas.
- Codex image bridge requests add `tool_choice: "auto"` when injecting or preserving an `image_generation` tool, without overriding explicit tool choice or Spark stripping behavior.
- Targeted backend tests, `git diff --check`, and denied-path audit pass.

## Allowed Paths

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
- `docs/workflow/worker-results/upstream-main-v0139-backend-compat-s23-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-backend-compat-s23-qa.md`

## Denied Paths

- `knowledge/*` currently dirty files, unless separately requested.
- Ent schema/generated files, migrations, wire generation, VERSION, README, sponsors, CI/deploy.
- Payment, subscription, balance, order currency, provider supported-types, or affiliate rebate behavior.
- Grok subscription support, codex_cli_only full fingerprint hardening, model-not-found 404 handler refactor, Ops/Keys frontend features.
- Public pages, Studio/Canvas, Model Plaza, payment pages, and unrelated frontend surfaces.

## Constraints

- Do not cherry-pick upstream commits blindly; adapt only the selected behavior to local helpers.
- Keep content-refusal handling limited to no-image OpenAI image responses with model refusal text.
- Keep failed-event sanitization scoped to `response.failed` SSE payloads; do not rewrite ordinary output events.
- Do not override an explicit client `tool_choice`.
- Do not stage or commit unrelated dirty `knowledge/*` files.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestOpenAIImageOutputCounter|TestImagesOAuthNonStreaming_ContentRefusalReturns400NoRetry|TestExtractModelRefusal_EmptyWhenNoText|TestIsOpenAITransientProcessingError|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIStreamingPassthroughResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIGatewayServiceForward_CodexImageInjectionRespectsGroupCapability|TestOpenAIGatewayServiceForward_ChannelBridgeOverrideEnablesCodexInjection|TestOpenAIGatewayServiceForward_CodexBridgePreservesExistingToolChoice|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool" -count=1
go test ./internal/pkg/apicompat -run "TestResponsesToAnthropic_.*Tool|TestResponsesToAnthropic_Custom" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --name-only | rg "^(knowledge/|backend/ent/|backend/migrations/|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|backend/internal/service/payment_|backend/internal/service/subscription_|deploy/|README|README_|assets/partners/|frontend/)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in allowed backend and workflow paths only.
- Worker-style result and QA report.
- Final report with ported/equivalent/skipped candidate accounting.

## Stop Rules

- Stop if any selected behavior needs Ent/migration/wire generation or frontend/product UI changes.
- Stop if image-refusal detection cannot be distinguished from true empty upstream responses.
- Stop if tests show the Codex image bridge change breaks Spark image tool stripping.
