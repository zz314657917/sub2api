### PASS: upstream-v0171-grok-ping-sse-filter-s195

## Findings

- Local Grok HTTP Responses streaming passed `event: ping` frames through the generic OpenAI SSE handler unchanged. The forced Grok WebSocket HTTP bridge discarded the `event:` line but forwarded its JSON `type=ping` payload to the client. Both shapes are invalid for strict Responses event parsers.
- The new Grok-only wrapper rewrites compatible `event: ping` frames to `: ping\n\n` before either consumer scans them. A frame with a different declared JSON event type remains unchanged.
- Non-ping frames stream without frame buffering. Ping candidates that exceed 16 lines or 16 KiB replay unchanged, limiting filter-owned buffering.

## Executed Checks

- `go test ./internal/service -run 'Test(GrokResponsesBillingPingFilter|ForwardGrokResponsesStreamingUsesXAIResponsesAndSnapshots|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge)' -count=1`: passed.
- `go test -v ./internal/service -run '^TestGrokResponsesBillingPingFilter' -count=1`: passed for compatible ping conversion, type-mismatch preservation, line/byte cap pass-through, and non-Grok no-op wrapping.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `gofmt -w internal/service/openai_gateway_grok_sse_filter.go internal/service/openai_gateway_grok_sse_filter_test.go internal/service/openai_gateway_grok.go internal/service/openai_gateway_grok_test.go internal/service/openai_ws_http_bridge.go internal/service/openai_ws_http_bridge_test.go`: passed.
- `git diff --check`, scoped conflict-marker scan, unmerged-index check, and staged allowlist audit: passed; exactly the nine contract-allowed files were staged.

## Unverified Risks

- No real Grok request, CLI client session, account credential, database, container, deployment, primary-worktree modification, push, or merge to `main` was performed.

## Recommendation

Commit this scoped stream-compatibility correction to the isolated integration branch. It provides source-level HTTP/WS regression evidence, not live xAI client certification.
