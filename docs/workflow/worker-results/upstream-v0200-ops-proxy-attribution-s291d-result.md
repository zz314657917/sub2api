### DONE: upstream-v0200-ops-proxy-attribution-s291d

## Implementation

- Added event-time proxy attribution to all 18 local Antigravity upstream error events, using `p.account` for pipeline events and `account` for direct owners.
- Preserved retries, failover, rate-limit and transport behavior.

## Commands run

- `go test ./internal/service -run 'Test(Antigravity|OpsUpstream)' -count=1` PASS
- `go test ./internal/service` PASS
- `go build ./...` PASS
- `git diff --check` and unmerged-index check PASS

## Risks / unverified

- Three non-Antigravity missed call sites were identified by the global scan and deliberately deferred to S291-E to preserve this contract boundary.
- Real provider traffic remains unverified.
