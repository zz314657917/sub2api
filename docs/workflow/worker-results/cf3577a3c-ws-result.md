### DONE: cf3577a3c-behavior-port

# Worker Result

## Task ID
cf3577a3c-behavior-port (WS/bridge owner)

## Status
done — only the assigned WebSocket, v2 passthrough, Chat/Messages bridge scope is complete. This is not the final repository QA verdict.

## Summary
- Ported Responses WS ingress compatibility normalization and OAuth reserved-tool aliasing into the local monolithic forwarder; every later turn merges its reverse map, avoiding an earlier alias being lost before delayed output arrives.
- Per-session normalized-name owners now register every `response.create` declaration, including a client-native `python__sub2api` name that has no reverse entry. A later conflicting `python`/`Python`/native declaration is rejected with policy violation in either order; a no-tools later turn retains the original binding. Published owner/reverse maps use copy-on-write before the gin context update.
- Restored client tool names on ingress stream, buffered JSON, HTTP bridge, v2 passthrough, Chat Completions, and Anthropic Messages output paths. v2 passthrough also restores the caller model after local mapped-model policy handling. A local `httptest` plus real `coderws.Accept`/`Dial` frame-I/O regression verifies that the adapter invokes both restorers for two output turns.
- Preserved local Fast Policy, per-turn model handling, image policy, existing legacy function-ID cleanup, and WS HTTP-bridge selection.

## Changed Files
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/cfport_websocket_test.go`

## Commands Run
```text
go test ./internal/service -run '^$' -count=1 -> PASS
go test ./internal/service -run '^TestCFPortWebSocket' -count=1 -v -> PASS (6 tests, including real local coderws frame I/O and cross-turn owner collisions)
go test <all real internal/service GoFiles> cfport_request_test.go cfport_gateway_test.go cfport_websocket_test.go -count=1 -v -> PASS
git diff --check -> PASS at last scoped check
```

## Test Output
```text
TestCFPortWebSocketReservedToolAliasRoundTripAcrossTurns -> PASS
TestCFPortWebSocketFrameAdapterRestoresModelAndToolAcrossTurns -> PASS (two real local WS frames; model/name restore, large number and ordinary content preserved)
TestCFPortWebSocketReservedToolAliasRejectsCaseCollision -> PASS
TestCFPortWebSocketReservedToolOwnersRejectCrossTurnCollisions -> PASS (python/native alias/Python conflicts in both relevant turn orders; no-tools continuation remains valid)
TestCFPortWebSocketBufferedAliasRoundTrip -> PASS
TestCFPortWebSocketPassthroughTurnModelRestoration -> PASS
```

## Risks
- Focused tests execute real production parsers, transforms, buffered terminal handling and local WebSocket frame I/O, but do not establish a live provider WebSocket connection, database, container or deployment result.
- Final Evaluator must rerun the exact aggregate gate after all parallel owner edits settle.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: partial (this worker's assigned owner scope only; other owners are coordinated separately)
- stop_rules_triggered: no
