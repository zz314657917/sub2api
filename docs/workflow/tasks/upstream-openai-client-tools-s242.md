# Upstream OpenAI Client Tools S242

## Task ID

`upstream-openai-client-tools-s242`

## Phase

`contract-approved`

## Role

The Developer Worker implements only this contract in an isolated worktree.
Codex remains Controller and Final Evaluator. The Agent Matrix's Terra route
has already returned a pre-inference 403 in this session; the user-authorized
named replacement for both Developer and independent QA is native
`gpt-5.6-sol`. This is an explicit replacement, not a silent fallback, and
both roles must still produce independent evidence.

## Goal

Selectively adapt the upstream OpenAI client-tool fixes from `44ef88f65`,
`7e579cb28`, and `b94e484e2` onto the local monolithic gateway. API-key
Responses requests must not lose `custom` tool declarations when the upstream
only accepts function tools. The WS-to-HTTP bridge must restore custom-tool
events to the client and retain the lowering mapping across follow-up turns
that omit a new `tools` declaration.

## Frozen Base And Provenance

- Local base: `main@baa6541acb1ef909b85d6d3cdb4817d2da0564c9`.
- Upstream source commits: `44ef88f65`, `7e579cb28`, `b94e484e2`.
- Upstream tip for ancestry: `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- All three source commits must be ancestors of the checked upstream tip.
- The local topology is intentionally different: upstream's
  `openai_gateway_passthrough.go` is owned by local
  `openai_gateway_service.go`; upstream's client-tool protocol helper is not
  present and may be added only under the allowlist below.

## Success Criteria

1. API-key OpenAI `/v1/responses` requests with `custom` tools are lowered to
   valid function declarations before the upstream request, while ordinary
   function tools and namespace-only requests remain unchanged.
2. Non-streaming and SSE passthrough responses restore `function_call` items
   to the original `custom_tool_call` shape, including string `input` and
   matching `call_id`/arguments semantics.
3. WS-to-HTTP bridge turns apply the same lowering/restoration. A follow-up
   turn that omits `tools` inherits the prior mapping and lowered declarations;
   an explicitly present `tools` field replaces the inherited state.
4. Mapping state is request-scoped and cleared before a new HTTP forwarding
   attempt, so account failover/retry cannot restore stale tools. OAuth,
   compact-only paths, non-OpenAI accounts, and existing Grok behavior retain
   their current contracts.
5. Focused tests cover API-key non-streaming/SSE restoration, bridge first and
   follow-up turns, explicit-tool override, malformed/trailing JSON rejection,
   and no-op behavior for ordinary or namespace-only tools.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`
- `backend/internal/service/openai_gateway_responses_client_tools_test.go`
- `backend/internal/service/openai_gateway_grok_tool_protocol.go`
- `backend/internal/pkg/apicompat/responses_client_tools.go`
- `backend/internal/pkg/apicompat/responses_client_tools_test.go`
- `docs/workflow/results/upstream-openai-client-tools-s242-result.md`

The two protocol/helper paths absent from the frozen local base may be newly
created. Keep the local owner names and existing `Responses*` types; do not
copy an unrelated upstream gateway split wholesale.

## Denied Paths And Constraints

- All user-owned dirty paths under `backend/internal/service/api_key_*`,
  `backend/internal/service/group_buy*`, the untracked room-policy test, and
  `outputs/` are protected and must remain byte-for-byte unchanged.
- No migrations, schema, repositories, billing/pricing, frontend, config,
  dependencies, generated files, security-policy widening, or remote refs.
- Do not include the upstream WS binary-policy or turn-start pricing changes
  (`9f24a5530`, `dec47e8fa`, `20ad5ec50`), OAuth identity work, or unrelated
  Grok/product changes.
- Do not make real provider, Redis/PostgreSQL, container, deployment, or push
  operations.
- Preserve existing response namespace normalization, SSE framing, replay
  input, usage accounting, and failover semantics outside the client-tool
  mapping boundary.

## Acceptance Commands

Run from `backend/` in the isolated worktree:

```powershell
go test ./internal/pkg/apicompat -run "TestResponsesClientTool|TestAdaptResponsesClientTools" -count=10
go test ./internal/service -run "TestOpenAIPassthroughAPIKey|TestOpenAIWSHTTPBridge.*ClientTool|TestProxyOpenAIWSHTTPBridgeTurnAPIKey.*ClientTools" -count=10
go test ./internal/pkg/apicompat
go test ./internal/service
go test ./cmd/server -run '^$' -count=1
gofmt -w <changed Go files>
git diff --check
```

The Controller and QA must additionally verify exact allowlist, no denied
paths, no conflict markers, an empty unmerged index, source ancestry, patch
scope, and protected-main status/patch hashes. A selector that discovers no
tests is not acceptance evidence.

## Output

The Developer result must be one commit plus
`docs/workflow/results/upstream-openai-client-tools-s242-result.md`, whose
first line is `### DONE: upstream-openai-client-tools-s242` (or an explicit
`### BLOCKED`/`### FAILED`). The independent QA report must begin with
`### PASS: upstream-openai-client-tools-s242`,
`### FAIL: ...`, or `### BLOCKED: ...`.

## Stop Rules

- Stop and return `BLOCKED` if the behavior requires a schema/product
  prerequisite, changes OAuth/Grok contracts, or cannot be isolated from the
  denied paths.
- Stop if the focused selector is not discoverable, full service or server
  compile fails due to this slice, or protected main changes.
- A worker must not silently replace the required QA model or widen the
  allowlist. Any further model availability problem is reported as `BLOCKED`;
  `gpt-5.6-sol` is the only replacement authorized for this Sprint.

## Evaluator Review

`PASS / contract-approved` (2026-08-22): the three source commits are narrow
OpenAI client-tool compatibility fixes. The local monolithic owner and helper
paths above are sufficient for behavior-level adaptation; no upstream schema,
pricing, migration, provider, or frontend prerequisite is required. The
focused commands exercise both HTTP and WS-HTTP bridge boundaries, while the
protected-main checks preserve the existing dirty worktree. Developer
implementation may start in an isolated E-drive worktree only.
