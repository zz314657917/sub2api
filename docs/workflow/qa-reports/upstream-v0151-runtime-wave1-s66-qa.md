### PASS: upstream-v0151-runtime-wave1-s66

# QA Report

## Scope

- Integration head: `696d23875`.
- Audit baseline: `2549e0b3a`.
- Contracts reviewed: `upstream-runtime-hotfixes-s66a`, `upstream-anthropic-grok-usage-s66b`, and `upstream-remote-compact-reliability-s66c`.
- Worker results reviewed: all three corresponding result reports have a `DONE` verdict.

## Findings

- No implementation bug or behavioral regression was found in the integrated wave.
- The full `internal/service` package is not green, but its failures are confined to the existing `group_peak_rate_test.go` peak-window/timezone assertions (`TestPeakMultiplierAt_*`, `TestPeakMultiplier_GatewayBillingSequence`, and `TestPeakMultiplier_SnapshotRoundTrip`). The integration diff does not touch that test or the peak-rate implementation, while all compact and reasoning-effort targeted tests pass. This is classified as pre-existing suite drift, not a failure of S66.
- The path audit found 23 changed paths from `2549e0b3a` to `696d23875`: 21 are explicitly allowed by the three contracts; the remaining two are workflow-only integration metadata (`docs/workflow/status.md` and `docs/workflow/main-log.md`). There are zero business-code paths outside the union of Allowed Paths and no Denied Path changes.
- The broad conflict-marker scan had one `=======` match inside an existing Antigravity raw diagnostic string. A boundary-marker scan for `<<<<<<< ` and `>>>>>>> ` found zero merge-conflict markers.

## Executed Checks

- `go test ./internal/pkg/apicompat -run "Test.*CacheCreation|Test.*Anthropic.*Usage|Test.*Responses.*Anthropic" -count=1` - PASS.
- `go test ./internal/service -run "TestExtractOpenAIReasoningEffortFromBody|Test.*Grok.*ReasoningEffort|TestForwardGrokResponsesStreaming" -count=1 -v` - PASS; the runnable shared extractor covers nested `reasoning.effort` and flat `reasoning_effort`. Grok forwarding tests are behind the repository's `unit` build tag and were not selected by the default build.
- `go test ./internal/handler -run "TestOpsCaptureWriter|Test.*Compact|TestOpenAIGatewayHandler" -count=1` - PASS.
- `go test ./internal/service/openai_ws_v2 -run "Test.*Disconnect" -count=1 -v` - PASS, including the disconnect-message coverage.
- `go test ./internal/service -run "Test.*Compact.*SSE|Test.*Compact.*Keepalive|Test.*Compact.*Output|Test.*RemoteCompact" -count=1` - PASS.
- `go test ./internal/service -run "TestRemoteCompact" -count=20` - PASS.
- `go test ./internal/handler -count=1` - PASS.
- `go test ./internal/service -count=1` - FAIL only in the existing group peak-rate tests described under Findings; no S66-targeted test failed.
- `go test -race ./internal/service -run "Test.*Compact.*Keepalive" -count=1` - BLOCKED by `CGO_ENABLED=0` (`-race requires cgo`).
- Repeated with `CGO_ENABLED=1` - BLOCKED because `gcc` is absent from `PATH`. No tool was installed.
- `git diff --check 2549e0b3a..HEAD` - PASS.
- Conflict-boundary marker scan across `backend` and `docs/workflow` - PASS, zero markers.
- Allowed/Denied audit from `2549e0b3a..HEAD` - PASS: 21 contract-allowed paths, two workflow metadata paths, zero business paths outside contract ownership.

## Unverified Risks

- Go race instrumentation was not executable on this Windows host. The synchronization paths passed focused tests and `TestRemoteCompact -count=20`, but that is not equivalent to a race-detector PASS.
- The `unit`-tagged Claude refresher and Grok forwarding tests cannot be counted as fresh evaluator PASS evidence because the repository's broader `-tags unit` service suite has known compile drift outside these contracts. Default-build service compilation and the runnable shared extractor passed.
- No real upstream compact request or physical Windows WebSocket reset was exercised; coverage is code-level and synthetic runtime testing.

## Recommendation

- PASS for merging the S66 wave. Preserve the race-detector limitation and existing group peak-rate suite drift as follow-up test-infrastructure work; neither is evidence of an S66 implementation defect.
