### PASS: upstream-ws-binary-turn-pricing-s241

## Scope

- Ported only the binary client WebSocket `response.create` policy path from
  upstream `9f24a5530` into the local passthrough adapter.
- Binary request frames now use the existing image-tool normalization, model
  resolution, fast-policy evaluation, and `BeforeRequest` audit hook.
- Binary non-`response.create` control frames remain byte-transparent; binary
  `session.update` still refreshes the captured session model for a later
  request that omits `model`.
- The upstream channel time-pricing product, migration 225, frontend/admin
  changes, and the independent turn-start pricing hunk were not ported. The
  local checkout lacks the owning profit-control prerequisites
  (`20ad5ec50`/`dec47e8fa`).

## Changed Paths

- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`

## Verification

- `go test ./internal/service -run 'TestPolicyEnforcingFrameConn.*|TestOpenAIFastPolicy.*|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_Passthrough' -count=1`: PASS.
- The same focused selector with `-count=10`: PASS.
- `go test ./internal/service`: PASS (65.888s).
- `go test ./internal/service/openai_ws_v2 -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=1`: PASS (`[no tests to run]`, package compiled).
- `gofmt` on all three changed Go files: PASS.
- `git diff --check`: PASS.

## Main Integration

- Business commit `96f2aa75d` was cherry-picked from the isolated candidate;
  the contract and result report are tracked in the follow-up evidence commits.
- Main fresh focused x10, `openai_ws_v2`, and `cmd/server` compile-only checks
  all passed after integration.
- No push was performed. `origin/main` remains `01ad28d5b`.
- Unrelated user changes in `backend/internal/service/api_key_service.go`,
  `backend/internal/service/group_buy.go`,
  `backend/internal/service/group_buy_test.go`,
  `backend/internal/service/api_key_room_policy_test.go`, and `outputs/`
  remained unstaged and outside both commits.

## Provenance And Scope Gates

- Source commit: `9f24a5530`; upstream tip checked as `upstream/main@d45135d87`.
- The source commit is an ancestor of the checked upstream tip.
- Business diff is limited to the adapter and the two focused regression
  owners listed above; no migration, frontend, dependency, provider, or
  pricing-product path is included.
- No real provider traffic, Redis/PostgreSQL, container, deployment, remote
  ref, or push operation was performed.

## Residual Risks

- The first ingress frame API still carries payload bytes without its original
  WebSocket message type; this pre-existing boundary is outside this slice.
- Tests use local fake WebSocket connections and do not prove a live provider
  accepts binary JSON frames.
- Turn-start pricing remains intentionally unimplemented until its local
  profit-control prerequisites are separately evaluated.
