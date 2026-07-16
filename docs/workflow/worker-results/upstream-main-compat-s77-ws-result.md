### DONE: upstream-main-compat-s77-ws

## Task ID

upstream-main-compat-s77-ws

## Status

done

## Summary

Implemented the local-monolith adaptation of upstream WS reliability fixes:
configurable first-message timeout with a 30-second default, deterministic
client reader cancellation/timeout close-and-join, per-turn client idle
timeout reuse, and malformed upstream JSON rejection before client output.
The original worker session stopped after implementation and test execution
without finalizing its report, so Codex completed the report and commit after
reviewing the unchanged worker diff.

## Changed Files

- backend/internal/config/config.go
- backend/internal/config/config_test.go
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_gateway_handler_test.go
- backend/internal/service/openai_ws_forwarder.go
- backend/internal/service/openai_ws_forwarder_ingress_session_test.go
- backend/internal/service/openai_ws_http_bridge.go
- backend/internal/service/openai_ws_http_bridge_test.go

## Commands Run

````text`
gofmt -w <8 changed Go files> -> PASS
GOTMPDIR=F:/mcplugins/sub2api/.tmp/go-build go test ./internal/config ./internal/handler ./internal/service ./internal/server/routes -run 'OpenAIWS|ImageGenerationIntent|ImageIntent' -count=1 -> PASS
git diff --check -> PASS
`````

## Test Output

````text`
internal/config: PASS
internal/handler: PASS
internal/service: PASS
internal/server/routes: PASS (no matching tests)
`````

Coverage includes default/custom first-message timeout, timeout and context
cancellation reader join, HTTP WS-v2 malformed JSON before/after output,
ingress malformed JSON failover, and client inter-turn idle closure.

## Risks

- No live upstream WebSocket or authenticated browser smoke was run.
- The local existing `gateway.openai_ws.read_timeout` is reused for client
  inter-turn idle reads because this baseline has no dedicated upstream idle
  configuration field.
- The user-authorized Codex worker fallback replaced unavailable
  `deepseek-v4-pro`; no push, deployment, or container update was performed.

## Knowledge Candidates

- none

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason

- none
