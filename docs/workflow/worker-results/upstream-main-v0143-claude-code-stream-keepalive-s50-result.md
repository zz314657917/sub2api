### DONE: upstream-main-v0143-claude-code-stream-keepalive-s50

## Summary

- Ported upstream `a5781fe31` into the isolated S50 worktree.
- Added Claude Code no-op delta keepalive helpers and version gating at `claude-cli/2.1.193`.
- Replaced periodic keepalive tickers with idle timers in both Anthropic API-key passthrough streaming and regular Anthropic streaming paths.
- Added regression coverage for idle keepalive ping, affected Claude Code text/tool-use no-op deltas, older Claude Code ping behavior, and existing streaming usage parsing.

## Changed Files

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_service_streaming_test.go`
- `docs/workflow/tasks/upstream-main-v0143-claude-code-stream-keepalive-s50.md`
- `docs/workflow/worker-results/upstream-main-v0143-claude-code-stream-keepalive-s50-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-claude-code-stream-keepalive-s50-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run

```powershell
gofmt -w backend/internal/service/gateway_service.go backend/internal/service/gateway_service_streaming_test.go
go test ./internal/service -run "TestGatewayService_StreamingKeepalive|TestGatewayService_StreamingReusesScannerBufferAndStillParsesUsage|TestDetachUpstreamContextIgnoresClientCancel" -count=1
git diff --check -- backend/internal/service/gateway_service.go backend/internal/service/gateway_service_streaming_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-claude-code-stream-keepalive-s50.md
```

## Test Output

- `internal/service`: PASS.
- `git diff --check`: PASS.

## Risks

- The regression tests use short `time.Sleep(1100 * time.Millisecond)` intervals to force keepalive behavior; they add several seconds to this targeted test run.
- This Sprint intentionally does not include upstream `7869b7fe3` Anthropic API Key Bearer auth. Anthropic-compatible upstreams that require Bearer auth still need a separate Sprint.

## Knowledge Candidates

- None.
