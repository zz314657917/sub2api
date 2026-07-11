### PASS: upstream-codex-imagegen-namespace-strip-s68b

# QA Report

## Scope

- Integration head: `219683890`; audit baseline: `5869f7b08`.
- Contracts reviewed: `docs/workflow/tasks/upstream-codex-imagegen-namespace-strip-s68b.md` and `docs/workflow/tasks/upstream-codex-imagegen-namespace-strip-s68b-fix1.md`.
- Both worker results begin with the required `### DONE:` verdict.
- QA mode: runtime. This report independently exercises the integrated initial implementation and fix1; worker self-reports were not used as PASS evidence.

## Findings

- No implementation bug, behavioral regression, or contract-boundary violation was found.
- Managed HTTP and API-key passthrough inspect `httpUpstreamRecorder.lastBody` and prove that the actual forwarded body is stripped. OAuth passthrough independently proves the same while preserving non-empty `instructions`, ordinary input, a non-image namespace, and a custom function named `imagegen`.
- Parsed WS ingress inspects `openAIWSCaptureConn.writes` and proves namespace, Responses Lite `additional_tools`, and matching `tool_choice` declarations are stripped before upstream forwarding.
- WS passthrough captures four upstream writes across two turns: the first and follow-up `response.create` text frames are stripped, while the intervening binary frame and `response.cancel` remain unchanged. The second-turn `BeforeRequest` hook also receives the stripped payload. The focused two-turn test passed three consecutive runs.
- Exact matching and preservation behavior are covered: flat `image_generation`, `image_gen` namespace declarations, namespace choices by `name` or `namespace`, and nested `tool` choices are removed; message input, non-image namespaces, ordinary functions, custom `imagegen` functions, `tool_choice: "auto"`, default `allow`, and non-Codex clients remain unchanged.
- Raw helper behavior is bounded and idempotent. Invalid JSON returns the original bytes with `changed=false` and a non-nil error; valid non-image JSON remains byte-for-byte unchanged; applying strip a second time reports no change.
- Spark preservation remains green: image declarations are removed, ordinary tools remain, and an image choice becomes `auto` when an ordinary top-level tool remains.
- S67 GPT-5.6 effort/candidate, WS passthrough effort, apicompat custom/tool-search/namespace/tool-choice, and messages fallback/usage regressions all pass.
- Path audit found 15 changed paths from `5869f7b08..HEAD`: ten backend source/test paths are within the original/fix1 Allowed Paths union, and five are workflow task/result/status/log evidence. No bridge, apicompat source, fallback/messages source, frontend, billing/pricing/accounting, migration, or deployment path changed.

## Executed Checks

- Original S68b primary service command - PASS (`internal/service`, `14.446s`).
  `go test ./internal/service -run "TestIsImageGenerationIntent|TestStripOpenAIImageGenerationTools|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_AccountPolicyStrips|Test.*Passthrough.*Image.*Strip|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Image.*Strip|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey" -count=1`
- Original S68b policy/S67 preservation command - PASS (`internal/service`, `0.145s`).
  `go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1`
- Original S68b apicompat preservation command - PASS (`internal/pkg/apicompat`, `0.981s`).
  `go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice" -count=1`
- Original S68b fallback/messages preservation command - PASS (`internal/service`, `0.135s`).
  `go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1`
- Original S68b compile-only command - PASS (`internal/service`, `0.094s`, no tests to run).
  `go test ./internal/service -run "^$" -count=1`
- Fix1 focused strip command - PASS (`internal/service`, `14.729s`).
  `go test ./internal/service -run "TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Passthrough.*Image.*Strip|TestStripOpenAIImageGenerationToolsFromRawPayload|TestOpenAIGatewayServiceForward_.*OAuth.*Passthrough.*Image.*Strip" -count=1`
- Fix1 WS passthrough effort/billing/regression command - PASS (`internal/service`, `0.342s`).
  `go test ./internal/service -run "Test.*WSPassthrough.*Effort|TestPassthroughBilling_|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_Passthrough" -count=1`
- Fix1 repeats of the S68b primary, apicompat, compile-only, and `git diff --check` commands were deduplicated against the fresh executions listed above.
- Required two-turn WS stability command - PASS (`internal/service`, `14.740s`, three consecutive runs).
  `go test ./internal/service -run "^TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughImageNamespaceStripAcrossTurns$" -count=3`
- Additional explicit preservation group - PASS (`internal/service`, `0.087s`): invalid raw JSON, idempotence, default allow, non-Codex HTTP, default-allow WS passthrough, non-image namespace, custom `imagegen`, and `tool_choice:auto`.
- `go test -list` audits confirmed the contract regexes match the intended service and apicompat tests rather than returning a vacuous PASS.
- `git diff --check 5869f7b08..HEAD` - PASS.
- `git diff --check` on the clean QA worktree before report creation - PASS.
- Allowed/Denied path audit - PASS: ten allowed backend source/test paths, five permitted workflow evidence paths, zero Denied Path changes.
- Diff precision review - PASS: every changed business line traces to namespace/additional-tools/tool-choice detection and stripping, forwarding stripped HTTP/WS bytes, or focused regression evidence.

## Unverified Risks

- No request was sent to a live OpenAI/Codex upstream. Actual relay bytes are verified with the in-process HTTP recorder and WebSocket capture connection.
- The full `internal/service` suite was intentionally not run because it has the known unrelated `group_peak_rate` timezone assertion drift and both contracts prohibit repairing unrelated full-suite drift. Focused contract regressions and package compile-only checks passed.
- Race testing and production deployment were not part of either contract.

## Recommendation

- PASS. S68b and fix1 satisfy the approved contracts and can proceed to the final Evaluator closeout. Do not interpret this QA as authorization to merge into `main` or push; those actions remain outside this worker assignment.

## Bug Owner Recommendation

- `none`

## Root Cause

- `none`

## Retest Scope

- None required.

## Knowledge Promotion

- `none`
