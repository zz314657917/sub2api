### DONE: upstream-v0200-ops-proxy-attribution-s291c

## Changed files

- OpenAI HTTP, compatibility, Images, Responses Images, Videos, Grok,
  transport and WebSocket fallback service owners listed in the contract.

## Implementation

- Added event-time proxy snapshots to all local OpenAI production event literals.
- Guarded OpenAI proxy-client construction with both proxy ID and hydrated proxy.
- Kept WebSocket default-client routes as `unknown`; added fallback regression tests.
- Local repository has no upstream compact/first-output/raw-stream helper files,
  so no artificial split-file ports were made.

## Commands run

- `go test ./internal/service -run 'Test(OpenAI|OpsUpstream|GatewayUpstream|Grok)' -count=1` PASS
- `go test ./internal/service` PASS
- `go build ./...` PASS
- `git diff --cached --check` and unmerged-index checks PASS

## Risks / unverified

- Antigravity uses a local monolithic gateway owner and remains separate S291-D.
- Real provider/WS traffic remains unverified.

## knowledge_candidates

- none
