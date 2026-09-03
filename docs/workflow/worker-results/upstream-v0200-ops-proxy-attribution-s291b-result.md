### DONE: upstream-v0200-ops-proxy-attribution-s291b

## Changed files

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service.go`

## Implementation

- Added event-time `ProxyID`/`ProxyName` attribution to local Gateway HTTP,
  retry/failover, Bedrock and Gemini compatibility error events.
- Kept response, retry, failover, billing and scheduling behavior unchanged.
- OpenAI/WS/provider split owners remain deferred to S291-C.

## Commands run

- `go test ./internal/service -run 'Test(Gateway|Gemini|OpsUpstream)' -count=1` PASS
- `go build ./...` PASS
- `git diff --cached --check` PASS
- Gateway event-literal completeness scan PASS (no `Platform`-first event remains)

## Risks / unverified

- OpenAI gateway, WebSocket, images, embeddings, Grok and Antigravity split
  call sites are not included in this batch.
- Full repository tests and real provider traffic remain unverified.

## knowledge_candidates

- none
