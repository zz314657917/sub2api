# Task Contract: upstream-codex-image-tool-strip-policy-s68a-backend

## Task ID

`upstream-codex-image-tool-strip-policy-s68a-backend`

## Status

`approved`

## Role

You are the Generator worker for the backend half of the Codex explicit image tool strip policy prerequisite.

## Goal

Adapt the backend behavior from upstream `f385cdceb`: add an account-level `allow/strip` policy for client-provided flat `image_generation` tools, apply it to Codex managed HTTP and WebSocket ingress paths, and keep the existing default behavior unchanged.

## Success Criteria

- `codex_image_generation_explicit_tool_policy` defaults to `allow`; unknown values also normalize to `allow`.
- `strip`, `remove`, and `drop` normalize to `strip`.
- Account lookup supports top-level `extra` and nested `extra.openai`, with top-level precedence.
- The policy applies only to Codex clients on OpenAI accounts.
- Managed HTTP and WS ingress remove flat `image_generation` tools and a matching `tool_choice` when policy is `strip`.
- Strip mode prevents the Codex image-generation bridge from re-injecting the removed tool.
- Non-image functions, other account `extra` fields, ordinary OpenAI clients, and existing Spark behavior remain unchanged.
- Tests cover policy normalization/precedence, managed HTTP, WS ingress, bridge suppression, and existing Spark behavior.

## Context

- Repo: `F:/mcplugins/sub2api`.
- Upstream prerequisite: `f385cdceb feat: add Codex image tool strip policy`.
- S67 is complete at `f7972b127`; its GPT effort and MCP bridge paths must not regress.
- S68b will later extend stripping to Codex `image_gen` namespace declarations, Responses Lite `additional_tools`, HTTP passthrough, and all raw payload paths.

## Allowed Paths

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

## Denied Paths

- `backend/internal/pkg/apicompat/**` and S67b fallback/message files.
- Frontend, API DTOs, handlers, repositories, migrations, billing/pricing, deployment, and production configuration.
- Namespace/Responses Lite/passthrough expansion owned by S68b.
- `knowledge/**` and global memories.

## Constraints

- Adapt to the local monolithic `openai_gateway_service.go` and `openai_ws_forwarder.go`; do not import upstream file-splitting refactors.
- Default `allow` must be behaviorally identical to the current branch.
- Do not add a database field, DTO, or dedicated API; reuse `Account.Extra`.
- Do not implement namespace or passthrough stripping in this contract.
- Do not change image pricing, accounting, group permissions, or standalone image endpoints.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "Test.*CodexImageGenerationExplicitToolPolicy|TestStripOpenAIImageGenerationTools|TestOpenAIGatewayServiceForward_AccountPolicyStripsExplicitImageTool|TestStripOpenAIImageGenerationToolFromRawPayload|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_.*Strips.*Image" -count=1
go test ./internal/service -run "TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool" -count=1
go test ./internal/service -run "Test.*GPT56.*Max|Test.*ReasoningEffort.*Candidate|Test.*WSPassthrough.*Effort" -count=1
go test ./internal/service -run "^$" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-image-tool-strip-policy-s68a-backend-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if implementation needs a database field, DTO, handler, repository, frontend, or API change.
- Stop if namespace, Responses Lite `additional_tools`, or passthrough stripping is required; those belong to S68b.
- Stop if default `allow` changes existing forwarding behavior.
- Stop if any S67-owned apicompat/fallback path must change.
- Do not repair unrelated full-suite drift.
