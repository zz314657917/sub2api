### DONE: upstream-v025-compat-fixes

## Changed paths

- `backend/internal/service/antigravity_token_provider.go`
- `backend/internal/service/token_cache_invalidator.go`
- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/openai_responses_compatibility.go`
- `backend/internal/service/openai_compact_stream_bridge.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/handler/stream_error_event.go`
- `backend/internal/pkg/apicompat/types.go`
- Contract-allowlisted tests for the four behaviors, including `upstream_v025_compat_test.go` for the required external-package Redis production-cache test.

`docs/workflow/status.md` was already modified before this worker began and was not changed.

## Implemented behavior

1. Antigravity OAuth cache keys now always use `ag:account:<id>`. Invalidation removes that current key and the old nonempty trimmed `ag:<project_id>` key.
2. Gemini streaming skips an upstream blank separator after a rewritten `data:` frame, retaining nonempty event/comment lines. A real `httptest` upstream sends LF and CRLF events, and an idle-pipe owner test observes downstream keepalive output.
3. OAuth normalization removes only `input[]` object top-level `internal_chat_message_metadata_passthrough`. Map transform, raw-body normalization, and WebSocket OAuth paths are covered; API-key WebSocket preserves it. A production `Forward` call with `httpUpstreamRecorder` verifies OAuth deletion and API-key HTTP passthrough. Nested content, stringified arguments, nonobjects, empty/null/missing input, and prompt-to-input compatibility are preserved.
4. Synthesized Responses events always serialize `sequence_number`, including zero. Compact synthesized events strictly decode an optional integer pointer and require it nonnil for exact `0/1/2`; handler failures use zero. WS normal error JSON is strictly decoded the same way, while the production-source fallback literal is located with Go AST and strictly decoded to require integer zero. `ResponsesStreamEvent` wire JSON includes zero; the WS/HTTP bridge regression preserves upstream nonzero sequence values.

## Executed evidence

- `gofmt -w ...` on all changed Go files; `git diff --check` passed.
- `go test ./internal/pkg/apicompat -count=1` — PASS.
- With `UPSTREAM_V025_REDIS_ADDR=127.0.0.1:60229`:
  `go test ./internal/service -run '^(TestHandleGeminiStreamingResponse_EventSeparatorIsExactlyOneBlankLine|TestHandleGeminiStreamingResponse_EmitsKeepaliveWhileUpstreamIsIdle|TestOpenAIWSHTTPBridgeRelaysSSEFramesAsWebSocketMessages|TestOAuthInputInternalMetadata|TestOAuthInputInternalMetadataMapAndInputOnlyCompatibility|TestOAuthInputInternalMetadataHTTPForwardAndAPIKeyBypass|TestUpstreamV025RedisRuntime)$' -count=1 -v` — PASS.
  The Redis case uses `repository.NewGeminiTokenCache` plus `service.NewCompositeTokenCacheInvalidator`, verifies PING, same-project account isolation, current+legacy invalidation, and no-project invalidation.
- `go test ./internal/handler -run '^(TestOpenAIHandleStreamingAwareError_ResponsesStreamingEmitsResponseFailed|TestGatewayHandleStreamingAwareError_ResponsesStreamingEmitsResponseFailed)$' -count=1 -v` — PASS.
- `go build ./...` — PASS.
- `go test ./internal/service -run '^(TestBuildOpenAICompactSSEPayloadNumbersSynthesizedFramesFromZero|TestWriteOpenAICompactSSEFailureMessageIncludesZeroSequenceNumber|TestBuildOpenAIWSHTTPBridgeErrorEventIncludesZeroSequenceNumber)$' -count=1 -v` — PASS. The fallback assertion parses the actual `openai_ws_http_bridge.go` function AST to obtain its production fallback literal; no test-only production failure seam was added.

## Baseline exclusions reproduced

- Contract broad untagged command fails only on pre-existing `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested`: expected `fc_6336c46cf80eccddb18637bafbdab3b1de6137977fd8f941d2ccc9763b32d`, actual `fc_477489beb020a16331ccd2917b0310e7bb315418ee7f92d3781a5e28eb761`.
- `go test -tags unit ./internal/service ...` cannot compile before the selected tests because existing fixtures have duplicate `stringPtr`, obsolete `computeTokenBreakdown`/`calculateCostInternal` arities, `buildCountTokensRequest` arity, and missing `FallbackMode`/`ExpiryWarnDays`. These match `E:/codex-runtime/pge/sub2api/upstream-v025-compat-fixes/baseline-notes.txt`; none were modified.

## Risks / limits

- No real provider, shared Redis, deployment, migration, or push was used.
- The `Forward` HTTP coverage uses the production forwarding path with the existing in-process `httpUpstreamRecorder`; it does not contact a real provider. The raw OAuth helper, WebSocket compatibility normalizer, and map transform are also directly executed.
- Not committed. Independent QA and controller review remain required.
