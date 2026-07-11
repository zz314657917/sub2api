### DONE: upstream-remote-compact-reliability-s66c

changed_files:
- backend/internal/handler/openai_gateway_handler.go
- backend/internal/handler/openai_gateway_handler_test.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_compact_sse_keepalive.go
- backend/internal/service/openai_compact_sse_keepalive_test.go
- backend/internal/service/openai_compact_stream_bridge.go
- backend/internal/service/openai_compact_stream_bridge_test.go
- docs/workflow/worker-results/upstream-remote-compact-reliability-s66c-result.md

commands_run:
- go test ./internal/service -run "Test.*Compact.*SSE|Test.*Compact.*Keepalive|Test.*Compact.*Output|Test.*RemoteCompact" -count=1
- go test ./internal/handler -run "Test.*Compact|TestOpenAIGatewayHandler" -count=1
- go test ./internal/service -run "TestRemoteCompact" -count=20
- go test -race ./internal/service -run "Test.*Compact.*Keepalive" -count=1
- CGO_ENABLED=1 go test -race ./internal/service -run "Test.*Compact.*Keepalive" -count=1
- go test ./internal/service ./internal/handler -count=1
- git diff --check

key_output:
- Focused service compact acceptance: PASS.
- Focused handler compact acceptance: PASS.
- Repeated compact concurrency/regression run, count=20: PASS.
- Full handler package: PASS.
- Full service package reached unrelated existing group_peak_rate timezone failures; compact-focused tests remained PASS.
- Race command with the default toolchain stopped because `-race requires cgo`; with `CGO_ENABLED=1` it stopped because `gcc` is not installed.
- git diff --check: PASS.

summary:
- Added the missing body-signal unary-to-SSE bridge prerequisite in the local pre-split layout, without importing upstream service-file refactors.
- Preserved raw response.output_item.done JSON, used compaction-only added events as a fallback, and supplemented non-empty terminal output that omitted the compaction item.
- Added downstream SSE comment keepalives for body-signal streaming compact requests and synchronized heartbeat/final response writes.
- Excluded heartbeat bytes from semantic response-size checks so keepalives do not suppress account failover.
- Converted errors after heartbeat commit into response.failed terminal events and recorded the intended stream error status.
- Forced application/json Accept for compact upstream requests on API-key passthrough and native request paths.

risks:
- Go race instrumentation could not run on this Windows host because the configured Go toolchain has CGO disabled and no gcc compiler is installed. The synchronized paths were exercised by focused tests and 20 repeated concurrent runs, but this is not equivalent to a race-detector PASS.
- The local branch predates the upstream bridge prerequisite commit, so the minimal bridge foundation was included under the contract-approved bridge paths.
- The local ops logger predates the upstream shared OpsStreamError helper; the equivalent request-context marker is implemented in the allowed compact bridge file, but broader ops middleware consumption was not added because that path is denied.

contract_compliance:
- Modified only Allowed Paths.
- Did not import the upstream service split or modify billing, pricing, retry counts, migrations, frontend, deployment, account scheduling policy, knowledge, or global memories.
- Preserved existing path-based compact JSON behavior while body-signal stream clients receive SSE output_item.done and response.completed events.
