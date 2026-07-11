### DONE: upstream-codex-image-tool-strip-policy-s68a-backend

# Worker Result

## Task ID

`upstream-codex-image-tool-strip-policy-s68a-backend`

## Status

`done`

## Summary

- Adapted upstream `f385cdceb` to the local monolithic gateway: OpenAI accounts now resolve `codex_image_generation_explicit_tool_policy` from top-level `extra` first and nested `extra.openai` second.
- The default and unknown policy remain `allow`; `strip`, `remove`, and `drop` normalize to `strip`.
- Managed HTTP and WS ingress strip only flat `image_generation` tools and matching `tool_choice` values for Codex clients when the account policy is `strip`.
- HTTP strip mode disables Codex image bridge injection so removed tools and bridge instructions are not reintroduced.
- Existing Spark behavior remains separate and unchanged; namespace tools, Responses Lite `additional_tools`, passthrough paths, ordinary OpenAI clients, and S67-owned paths were not changed.

## Changed Files

- `backend/internal/service/codex_image_generation_bridge.go`
- `backend/internal/service/codex_image_generation_bridge_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `docs/workflow/worker-results/upstream-codex-image-tool-strip-policy-s68a-backend-result.md`

## Commands Run

```text
GOTMPDIR=E:/codex-worktrees/.gotmp/sub2api-s68a-backend go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|TestStripOpenAIImageGenerationTools|TestOpenAIGatewayServiceForward_AccountPolicyStripsExplicitImageTool|TestStripOpenAIImageGenerationToolFromRawPayload|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Strips.*Image" -count=1 -> PASS
GOTMPDIR=E:/codex-worktrees/.gotmp/sub2api-s68a-backend go test ./internal/service -run "TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool" -count=1 -> PASS
GOTMPDIR=E:/codex-worktrees/.gotmp/sub2api-s68a-backend go test ./internal/service -run "Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1 -> PASS
GOTMPDIR=E:/codex-worktrees/.gotmp/sub2api-s68a-backend go test ./internal/service -run "^$" -count=1 -> PASS package compile
git diff --check -> PASS
allowed-path audit -> PASS
denied-path audit -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.615s
ok github.com/Wei-Shaw/sub2api/internal/service 0.247s
ok github.com/Wei-Shaw/sub2api/internal/service 0.127s
ok github.com/Wei-Shaw/sub2api/internal/service 0.100s [no tests to run]
```

## Risks

- No live OpenAI/Codex upstream was used; managed HTTP and WS behavior is covered by local recorder and in-process WebSocket tests.
- Namespace declarations, Responses Lite `additional_tools`, and passthrough/raw expansion are intentionally deferred to S68b by contract.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
