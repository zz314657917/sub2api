### PASS: openai-messages-pending-preamble-s178

`9544a268` deferred the synthetic `response.created` preamble so a leading rate-limit event could
fail over before downstream output. The same deferral hid a failed downstream writer when the upstream
ended without a terminal event. `missingTerminalErr` now drains the pending preamble once at EOF:
successful writes keep the established missing-terminal path, and a failed write sets
`ClientDisconnect` before the existing error classification.

Executed checks:

- `go test ./internal/service -run 'TestForwardAsAnthropic_MissingTerminalAfterClientDisconnectSkipsOpsAndFailover|TestHandleAnthropicStreamingResponse_RateLimitAfterCreatedReturns429FailoverBeforeOutput' -count=1`
- `go test ./internal/service -run 'TestForwardAsAnthropic_MissingTerminal|TestHandleAnthropicStreamingResponse_RateLimit' -count=1`
- `go test ./cmd/server`
- `gofmt -d`, `git diff --check`, and `git ls-files -u`

The full Go suite reaches the pre-existing `internal/middleware` RegistrationRiskLimiter Redis-nil
failures. The same package fails unchanged on a detached `origin/main` worktree; it is not attributed
to S178 or the consolidation merge.
