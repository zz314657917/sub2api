### PASS: upstream-v0171-grok-filter-lifecycle-tests-s196

## Findings

- The S195 Grok-only filter already closes its source through `sync.Once`, returns the source close error after a successful pipe-reader close, and emits complete non-ping frames before upstream EOF.
- S196 adds direct regression evidence for those lifecycle guarantees, partial ping frames at EOF, bare CR framing, and max-line-size error propagation. No production correction was needed.

## Executed Checks

- `go test ./internal/service -run '^TestGrokResponsesBillingPingFilter' -count=1`: passed.
- `go test -v ./internal/service -run '^TestGrokResponsesBillingPingFilter' -count=1`: passed all compatible ping, non-Grok, bounded-buffer, close-once/error, complete-frame, EOF/CR, and oversized-line cases.
- `go test ./internal/service -run 'Test(ForwardGrokResponsesStreamingUsesXAIResponsesAndSnapshots|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge)' -count=1`: passed.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `gofmt -w internal/service/openai_gateway_grok_sse_filter_test.go`, `git diff --check`, scoped conflict-marker scan, unmerged-index check, and staged allowlist audit: passed; exactly the four contract-allowed files were staged.

## Unverified Risks

- No real Grok request, CLI client session, account credential, database, container, deployment, primary-worktree modification, push, or merge to `main` was performed.

## Recommendation

Commit this test-only completion to the isolated integration branch. It strengthens source-level lifecycle evidence for S195 and does not certify live xAI behavior.
