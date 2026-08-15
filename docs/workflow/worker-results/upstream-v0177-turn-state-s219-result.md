### DONE: upstream-v0177-turn-state-s219

## Scope

- Behaviorally ported the HTTP `x-codex-turn-state` relay from upstream
  `8219dcfc8` and its test correction `4d9fedee2` into the local monolithic
  OpenAI gateway.
- Added only the normal and passthrough outbound guard placements from
  `fce41e318`. No fingerprint convergence/default behavior, frontend,
  migration, dependency, provider, deployment, or WS protocol changes were
  included.
- HTTP provenance requires a positive API-key ID plus the original inbound
  session (`session-id` preferred over `session_id`). It expires with the
  existing OpenAI WS sticky-session TTL and is opportunistically swept.

## Changed Files

- `backend/internal/service/openai_codex_turn_state.go`
- `backend/internal/service/openai_codex_turn_state_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_tool_namespace_normalization_s92_test.go`

## Verification

- Default-tag discovery for all seven S219 regressions: PASS.
- `go test ./internal/service -run '^(TestOpenAICodexTurnStateSeedRequiresAPIKeyAndSession|TestOpenAICodexTurnStateRelayGuardAndExpiry|TestOpenAIHTTPBuildersGuardCrossAccountTurnState|TestOpenAIStreamingTurnStateRecordsOnlyOnCommit|TestOpenAINonStreamingTurnStateRelaysJSONAndSSE|TestOpenAIPassthroughTurnStateRelayAndGuard|TestWriteOpenAIPassthroughResponseHeadersTurnState)$' -count=10`: PASS.
- Contract failover/Claude/WS compatibility pattern: PASS.
- `go test ./internal/service -count=1`: completed after its task-owned test process exited.
- `go test ./internal/handler -count=1`: completed after its task-owned test process exited.
- `go test ./internal/server -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=0`: PASS.
- All tests used in-process fakes or loopback fixtures; no provider, database,
  Redis, container, deployment, or push action occurred.
- `gofmt`, `git diff --check`, conflict/index checks, and upstream ancestry
  checks for `8219dcfc8`, `4d9fedee2`, and `fce41e318`: PASS.

## Risk

- The relay deliberately tracks only HTTP responses that actually commit a
  downstream response. Native HTTP never injects stored turn state; it strips
  only a known foreign-account client echo. The separate Claude compatibility
  bridge and WS store/replay paths are unchanged.

## R1 Follow-up

- Fixed `writeOpenAIPassthroughResponseHeaders` so a nil source still clears a
  stale turn-state header when a destination exists.
- Split non-streaming response handling into stage, downstream write, then
  provenance note. Normal JSON, normal SSE-to-JSON, passthrough JSON, and
  passthrough SSE-to-JSON now record only after `Writer.Written()` is true.
- Added default-tag coverage for malformed/expired/unknown provenance, sweep,
  nil/zero accounts, nil response headers, commit ordering, and exactly-once
  normal/passthrough streaming provenance.
