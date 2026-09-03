### DONE: upstream-v0200-ops-proxy-attribution-s291e

## Implementation

- Added attribution to the three Gateway/Gemini production events found by the final completeness scan.
- The final scan confirms every direct production `OpsUpstreamErrorEvent` append now includes proxy attribution.

## Commands run

- `go test ./internal/service -run 'Test(Gateway|Gemini|OpsUpstream)' -count=1` PASS
- `go test ./internal/service` PASS
- `go build ./...` PASS
- `git diff --check` PASS
- production event completeness scan PASS
