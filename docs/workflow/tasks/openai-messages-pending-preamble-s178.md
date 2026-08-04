---
task_id: openai-messages-pending-preamble-s178
phase: contract-draft
role: Planner/Generator/Evaluator
qa_mode: runtime
---

# S178 OpenAI Messages Pending-Preamble Disconnect Contract

## Goal

Repair the regression introduced by `9544a268`: a leading `response.created` preamble is correctly
deferred for rate-limit failover, but an EOF before any terminal event no longer probes the downstream
writer and therefore converts a client disconnect into an upstream failover.

## Success Criteria

- A leading rate-limit `response.failed` still fails over before any downstream Anthropic output.
- When a stream contains only buffered `response.created` data and ends without a terminal event, the
  pending preamble is written once only to determine whether the client disconnected.
- A failed pending-preamble write returns the existing missing-terminal error with
  `ClientDisconnect=true`, creates no Ops upstream-error fact, and does not produce an
  `UpstreamFailoverError`.
- A writable downstream retains the existing missing-terminal classification; no normal terminal or
  partial-output behavior changes.

## Allowed Paths

- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_compat_model_test.go`
- `docs/workflow/tasks/openai-messages-pending-preamble-s178.md`
- `docs/workflow/qa-reports/openai-messages-pending-preamble-s178-qa.md`

## Denied Paths

- All other backend/frontend source, generated code, schema/migrations, configuration, deployment,
  Docker, remote Git, production resources, and the primary worktree.

## Constraints

- Preserve `9544a268`'s no-output-before-leading-rate-limit-failover guarantee.
- Reuse the existing SSE conversion, write, disconnect, result, Ops and failover paths; do not add a
  second client-disconnect mechanism or change retry policy.
- Make the smallest additive correction and regression assertion needed to distinguish an EOF from a
  rate-limit terminal event.

## Acceptance Commands

```powershell
cd backend
go test ./internal/service -run 'TestForwardAsAnthropic_MissingTerminalAfterClientDisconnectSkipsOpsAndFailover|TestHandleAnthropicStreamingResponse_RateLimitAfterCreatedReturns429FailoverBeforeOutput' -count=1
go test ./internal/service -run 'TestForwardAsAnthropic_MissingTerminal|TestHandleAnthropicStreamingResponse_RateLimit' -count=1
go test ./cmd/server
gofmt -d internal/service/openai_gateway_messages.go internal/service/openai_compat_model_test.go
git diff --check
git ls-files -u
```

## Stop Rules

- Stop if the correction requires retry-policy, billing, account-selection, transport, schema, or
  frontend changes.
- Stop if the rate-limit no-output assertion can no longer pass together with the disconnect test.

## Evaluator Review

### PASS: openai-messages-pending-preamble-s178 contract-approved

The regression is source-local, reproducible against the integration branch, absent from
`origin/main`, and has an existing failing test plus a directly adjacent rate-limit regression. The
allowed path is sufficient to restore the intended behavior without changing the wider integration
scope.
