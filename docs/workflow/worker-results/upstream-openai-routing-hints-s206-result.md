### DONE: upstream-openai-routing-hints-s206

# Worker Result

## Task ID
upstream-openai-routing-hints-s206

## Status
`done`

## Summary
- Cherry-picked upstream `8ad0a5ff5` as `e6120ec69`; the `nanoid@3.3.17` lockfile patch retains the same stable patch-id.
- Adapted upstream `915cc7e7b`, `815035fcc`, and `de349187d` to the local monolithic OpenAI HTTP/WebSocket topology.
- OAuth requests now synthesize a gateway-owned hint from the final model and canonical `priority`/`flex` tier, while API-key paths remove caller/account supplied variants.
- Removed the legacy OAuth `responses=experimental` behavior, added WebSocket first/later-turn hint refresh, soft pool affinity, idle mismatch replacement, generation guards, and stale prewarm-target rejection.

## Changed Files
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_images_test.go`
- `backend/internal/service/openai_routing_hint.go`
- `backend/internal/service/openai_routing_hint_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_pool.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `frontend/pnpm-lock.yaml`
- S206 workflow contract/status/log/result files.

## Commands Run
```text
go test ./internal/service -run 'TestOpenAICodexRoutingHint|TestOpenAI.*RoutingHint|TestOpenAI.*LegacyBeta|TestOpenAIWS.*Routing|TestOpenAIWS.*Affinity|TestOpenAIWS.*Generation|TestOpenAIWS.*Prewarm|TestOpenAIGatewayServiceForwardImages_OAuthPassesNAndReturnsAllImages' -count=1 -> PASS
go test ./internal/service -count=1 -> PASS
go test ./cmd/server -run '^$' -count=0 -> PASS
gofmt -w <changed Go files> -> PASS
git diff --check 3cec8bb904bd880d5b2ef56daee85292e8cfc95a -> PASS
stable patch-id comparison for 8ad0a5ff5/e6120ec69 -> PASS
changed-path, conflict-marker and unmerged-index checks -> PASS
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 7.720s
ok github.com/Wei-Shaw/sub2api/cmd/server 10.108s [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/service 61.833s
```

## Risks
- No real OpenAI OAuth provider, WebSocket endpoint, proxy, container, deployment, staging, or production traffic was exercised.
- Routing hints are intentionally advisory for pooled WebSocket continuation; a busy pool can reuse a compatible connection dialed with a different hint.
- The upstream pool has a later topology-wait channel architecture that this historical local branch does not contain. S206 ports only the routing-affinity, generation, and prewarm behavior applicable to the local pool.
- The first full service run failed only because `openai_images_test.go` still expected the removed legacy beta header. The contract allowlist was narrowly amended and the assertion was changed exactly as upstream `915cc7e7b`; the targeted and complete reruns passed.

## Knowledge Candidates
- On deeply diverged upstream WebSocket ports, routing hints should remain separate from continuation compatibility: prefer/replace idle mismatches, but allow busy-capacity fallback.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
