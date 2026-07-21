# S92: Codex redundant custom-tool namespace normalization

## Task ID

`codex-tool-namespace-normalization-s92`

## Role

Codex acts sequentially as Planner, Generator, and final Evaluator. No worker is
used for this narrow compatibility port.

## Goal

Port the behavior of `MayukXT/codex@c5b3596b` to the Sub2API response boundary:
when a Responses `custom_tool_call` item has `namespace == name`, remove the
redundant namespace before returning the payload to Codex clients. This avoids
the `unsupported custom tool call: execexec` failure described by
`openai/codex#32435`.

## Success Criteria

- HTTP JSON, HTTP SSE, native WebSocket, and WebSocket-to-HTTP bridge responses
  normalize redundant custom-tool namespaces.
- Root items, `item`, `output`, and `response.output` item locations are covered.
- A distinct namespace remains unchanged.
- Non-custom response items remain unchanged.
- Invalid JSON and payloads without the target fields remain byte-identical.
- Focused service regressions and repository gates pass.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_tool_namespace_normalization_s92_test.go`
- `docs/workflow/tasks/codex-tool-namespace-normalization-s92.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- request-side tool declarations or tool-choice rewriting
- billing, routing, scheduling, persistence, and account selection
- valid namespace values that differ from the tool name

## Constraints

- Match the upstream fork semantics exactly: compare namespace and name without
  trimming or case folding, and remove namespace only when they are equal.
- Do not recursively rewrite arbitrary JSON objects. Inspect only known
  Responses item locations.
- Keep the no-op path allocation-free before JSON parsing by checking for both
  `custom_tool_call` and `namespace` markers.
- Preserve all existing user changes in the primary checkout.

## Acceptance Commands

```powershell
go test ./internal/service -run 'TestS92' -count=1
go test ./internal/service -run 'TestOpenAIGatewayService_ToolCorrection|TestOpenAIWSMessageLikelyContainsToolCalls|TestS92' -count=1
git diff --check
git diff --name-only HEAD --
git diff --name-only --diff-filter=U
```

## Output

- One scoped implementation commit on `codex/codex-tool-namespace-s92`.
- Final findings, executed checks, unverified risks, and recommendation.

## Stop Rules

- Stop if the fix requires request-side namespace rewriting.
- Stop if a valid distinct namespace must be removed to make tests pass.
- Stop if implementation needs Ent, migrations, frontend, deployment, billing,
  routing, scheduling, persistence, or account-selection changes.
