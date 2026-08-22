# Upstream WS Binary Policy And Turn Pricing S241

## Task ID

`upstream-ws-binary-turn-pricing-s241`

## Role

Controller implementation and final evaluator. This slice is implemented in an
isolated worktree and reviewed against the frozen local `main` base before
integration.

## Goal

Selectively port the upstream behavior from `9f24a5530` that makes binary
client WebSocket frames follow the same passthrough policy/audit path as text
frames. Do not port the unrelated channel time-pricing product, migration,
admin API, frontend changes, or its turn-pricing hunk: this checkout does not
contain the upstream profit-control prerequisites that own that billing
behavior (`20ad5ec50`/`dec47e8fa`).

## Frozen Base And Source

- Local base: `main@b4010d780`
- Upstream source: `upstream/main@d45135d87`
- Narrow source commit: `9f24a5530`
- Source ancestry must be checked against `upstream/main` before integration.

## Success Criteria

- Binary `response.create` frames enter the existing passthrough filter,
  policy normalization, model resolution, and `BeforeRequest` security-audit
  hook exactly as text frames do.
- Binary non-`response.create` frames continue to relay unchanged.
- The passthrough ingress regression covers both binary request policy/hook
  handling and unchanged binary control-frame forwarding.
- The existing passthrough turn lifecycle remains unchanged; no pricing or
  profit-control behavior is added without its owning local prerequisite.
- Existing text-frame behavior remains unchanged.

## Allowed Paths

- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `docs/workflow/tasks/upstream-ws-binary-turn-pricing-s241.md`
- `docs/workflow/results/upstream-ws-binary-turn-pricing-s241-result.md`

## Denied Paths And Constraints

- All channel time-pricing implementation, migration `225`, repository pricing
  changes, frontend files, generated files, dependencies, and unrelated
  gateway changes are denied.
- Do not modify user-owned content, `knowledge/*`, or `outputs/*`.
- Do not use real provider traffic, Redis/PostgreSQL, containers, deployment,
  remote refs, or push.
- Preserve existing text-frame, failover, security-audit, and passthrough
  lifecycle semantics except for the timestamp and binary-policy behavior
  stated above.

## Acceptance Commands

From `backend/`:

```powershell
go test ./internal/service -run "TestPolicyEnforcingFrameConn.*|TestOpenAIFastPolicy.*|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_Passthrough" -count=1
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w backend/internal/service/openai_ws_v2_passthrough_adapter.go backend/internal/service/openai_fast_policy_ws_test.go backend/internal/service/openai_ws_forwarder_ingress_session_test.go
git diff --check
```

Also perform exact allowlist, conflict/unmerged-index, upstream ancestry,
patch-scope, and protected-main status checks. If a focused selector resolves
to no tests, it is not acceptance evidence and must be replaced by a selector
that discovers real tests.

## Output

Write `docs/workflow/results/upstream-ws-binary-turn-pricing-s241-result.md`
with first line `### PASS: upstream-ws-binary-turn-pricing-s241`,
`### FAIL: ...`, or `### BLOCKED: ...`; include changed paths, commands,
results, provenance, scope, and residual risks.

## Stop Rules

- Stop and return to planning if the change requires the channel time-pricing
  schema/product, profit-control prerequisite, broad WS protocol redesign, or
  denied paths.
- Stop if the narrow source cannot be adapted without changing local security,
  failover, or billing contracts beyond this goal.
- Stop if any focused command is not discoverable, if the full service or server
  compile fails due to this slice, or if the protected main worktree changes.

## Status

`done`
