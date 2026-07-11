# Task Contract: upstream-codex-mcp-tool-bridge-s67b

## Task ID

`upstream-codex-mcp-tool-bridge-s67b`

## Status

`approved`

## Role

You are the Generator worker for the Codex Responses-to-Chat/Messages fallback tool bridge.

## Goal

Adapt the complete upstream custom tool, `tool_search`, and namespace flatten/restore sequence to the local apicompat bridge, including collision and `tool_choice` hardening. This worker also owns `openai_gateway_messages.go` and must apply the message-fallback effort-candidate call-site adjustment from `80b3d4c1f`/`c3ae5fc3c` there.

## Success Criteria

- Responses custom tools survive conversion through proxy function declarations and return to their original call/output forms.
- Built-in `tool_search` calls/outputs are supported through fallback conversion.
- Namespace tools flatten deterministically and return-path calls are restored, fixing Codex MCP `unsupported call` failures.
- Built-in/proxy/namespace naming collisions fail explicitly instead of silently routing the wrong tool.
- `tool_choice` references only emitted tools; forced `tool_search` selects its generated proxy function.
- Messages and Responses fallback paths use the same bridge metadata and preserve effort metadata from mapped/requested model candidates.
- Focused apicompat and fallback tests pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream sequence: `75fb3c41c`, `27e29f056`, `794233832`, `f1082bb78`, `a2cdaa641`, `e2b68d1f9`, `90e9d03de`.
- Local tree predates upstream file splitting; port behavior without importing unrelated refactors.

## Allowed Paths

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_custom_tools_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/responses_stream_event_wire.go`
- `backend/internal/pkg/apicompat/responses_stream_event_wire_test.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_usage_test.go`
- `docs/workflow/worker-results/upstream-codex-mcp-tool-bridge-s67b-result.md`

## Denied Paths

- All other paths, including core routing outside the two fallback files, billing/pricing, migrations, frontend, deployment, and production configuration.
- S67a-owned gateway/WS files.
- `knowledge/**` and global memories.

## Constraints

- Apply the upstream tool sequence in order; do not cherry-pick only the final hardening commits.
- Preserve existing ordinary function/web-search conversion behavior.
- Keep fallback-only scope; do not alter direct Responses passthrough.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice|Test.*Responses.*Chat" -count=1
go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-mcp-tool-bridge-s67b-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes on the assigned branch and return the commit hash.

## Stop Rules

- Stop if the sequence requires broad upstream file splitting or direct-route changes outside Allowed Paths.
- Stop on ambiguous collision semantics; do not invent silent fallback behavior.
- Do not repair unrelated unit-suite drift.
