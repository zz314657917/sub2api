### PASS: recent-commit-regressions-s130

## Findings

- Repaired all 11 OpenAI `shouldFailoverOpenAIUpstreamResponse` constructors so
  an exact model-capacity response uses the explicit five-attempt same-account
  retry limit while retaining the existing pool retry condition for every
  other response.
- Provider `partially_refunded` now enters `needs_review` with its seat held in
  `refund_processing`; it cannot be closed as a completed group-buy refund.
- A payment callback after a released timed-out seat now records a
  `refund_queued` event and changes the seat to `refund_pending` without
  affecting round share counters.
- Grok removes orphaned `tool_choice` for missing, `null`, and empty `tools`.
- The leaderboard client guard waits for a loaded numeric threshold; a loaded
  `0` remains a valid threshold.

## Executed Checks

- File-scoped Go regression command, loading current production sources plus
  `group_buy_test.go` and the S130 regression test: PASS (5.587 s).
- `go build ./...`: PASS.
- `go test ./... -run '^$'`: PASS as a repository package compile probe.
- `npm.cmd run test -- --run src/router/__tests__/guards.spec.ts`: PASS (44/44).
- `npm.cmd run typecheck`: PASS.
- `gofmt -d` on all changed Go files: PASS (no output).
- Static failover audit: all 11 `if s.shouldFailoverOpenAIUpstreamResponse(...)`
  constructors contain `SameAccountRetryLimit`.
- `git diff --check`, unmerged-index, conflict-marker, and S130 allowlist
  audits: PASS.

## Unverified Risks

- The normal `go test ./internal/service -run ...` and unit-tag Grok command
  remain blocked by pre-existing test compilation drift: `stringPtr` is
  redeclared, billing test calls use stale signatures, and the existing Grok
  unit test references removed runtime-block helpers. These files are outside
  S130 and were not modified.
- No real payment provider, upstream model-capacity response, authenticated
  browser cold start, deployment, or container smoke was run.

## Recommendation

`PASS / source-level`: retain the S130 patch for review and commit only under
explicit user authorization. Before release, run the group-buy provider refund
and first-load leaderboard flows against an authorized runtime environment.
