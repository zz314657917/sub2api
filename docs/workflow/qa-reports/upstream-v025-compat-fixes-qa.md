### PASS: upstream-v025-compat-fixes

# QA Report

## Task ID
`upstream-v025-compat-fixes`

## Verdict
`PASS (首轮 FAIL 修复后独立复验)`

## Contract Checked

- `docs/workflow/tasks/upstream-v025-compat-fixes.md` (`status: approved`, `review_verdict: PASS`)
- `docs/workflow/contract-reviews/upstream-v025-compat-fixes-review.md` (`### PASS`)
- Developer result: `docs/workflow/worker-results/upstream-v025-compat-fixes-result.md` (`### DONE`)
- Final implementation range: `a022d9a3457457463ed06dd025370ad57936e668..c589b7689`

## Evidence

- Diff reviewed: `yes`. Four commits (`ac5fce6f6`, `745a1021f`, `8fb44b12e`, `c589b7689`) change exactly 17 allowlisted `backend/**` paths. `git diff --check` and `gofmt -d` over every changed Go file passed. The pre-existing dirty `docs/workflow/status.md` is outside that committed range; no denied production path is in the range.
- Dedicated Redis production-cache acceptance: `UPSTREAM_V025_REDIS_ADDR=127.0.0.1:60229 go test ./internal/service -run '^(TestUpstreamV025RedisRuntime|TestHandleGeminiStreamingResponse_.*|TestOAuthInputInternalMetadata.*|TestBuildOpenAICompactSSEPayload.*|TestWriteOpenAICompactSSEFailureMessage.*|TestBuildOpenAIWSHTTPBridgeErrorEvent.*|TestOpenAIWSHTTPBridgeRelaysSSEFramesAsWebSocketMessages)$' -count=1 -v` — `PASS`.
  - Actual task Redis test reached PING/set/get/delete through `repository.NewGeminiTokenCache` and `service.NewCompositeTokenCacheInvalidator`; it passed account isolation, current plus legacy invalidation, and no-project invalidation.
  - Gemini owner tests passed LF/CRLF separator, non-data line, keepalive, normal completion, client-disconnect, and context-cancel cases.
  - OAuth input metadata tests passed map/raw HTTP/WebSocket OAuth removal, API-key preservation, nested/text/nonobject preservation, and the production `Forward` path with in-process upstream recorder.
  - Compact, handler, and WS bridge selected tests passed, including upstream WS sequence `41/42/43` relay preservation.
- `go test ./internal/handler -run '^(TestOpenAIHandleStreamingAwareError_ResponsesStreamingEmitsResponseFailed|TestGatewayHandleStreamingAwareError_ResponsesStreamingEmitsResponseFailed)$' -count=1 -v` — `PASS`.
- `go test ./internal/pkg/apicompat -count=1 -v` — `PASS`, including `TestWire_SequenceNumberPresentAtZero` with explicit map-key presence checks.
- `go build ./...` — `PASS`.
- Immutable-baseline limitations from `E:/codex-runtime/pge/sub2api/upstream-v025-compat-fixes/baseline-notes.txt`: tagged service fixtures do not compile because of existing fixture/API drift; the broad untagged regex has the pre-existing `TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested` expected/actual hash mismatch. These were not rerun as no-tests-to-run substitutes and are not attributed to this range.

## Initial Findings (Resolved)

1. **Resolved:** compact sequence tests formerly used `gjson.Int() == 0`, which could not distinguish omitted fields. Commit `9608870be` now uses `json.Unmarshal` into an optional `*int`, requires non-nil field presence, and checks exact compact `0/1/2` plus standalone failure `0`.
2. **Resolved within the approved scope:** `TestBuildOpenAIWSHTTPBridgeErrorEventIncludesZeroSequenceNumber` strictly decodes normal output and uses Go AST to locate the fallback literal in the actual production `buildOpenAIWSHTTPBridgeErrorEvent` function, then strictly decodes it into an optional `*int` and requires numeric zero. The map contains only fixed strings/integers, so its `json.Marshal` error branch is unreachable in normal Go execution; no production seam was added solely to force that branch.

## Bug Owner Recommendation
`none`

## Root Cause
`none`

## Retest Evidence

1. Reviewed final range `a022d9a3457457463ed06dd025370ad57936e668..9608870be`: five commits and exactly the 17 contract-allowlisted `backend/**` files. No denied production path is in the final range.
2. Re-ran with `UPSTREAM_V025_REDIS_ADDR=127.0.0.1:60229`:
```text
go test ./internal/service -run '^(TestUpstreamV025RedisRuntime|TestHandleGeminiStreamingResponse_.*|TestOAuthInputInternalMetadata.*|TestBuildOpenAICompactSSEPayload.*|TestWriteOpenAICompactSSEFailureMessage.*|TestBuildOpenAIWSHTTPBridgeErrorEvent.*|TestOpenAIWSHTTPBridgeRelaysSSEFramesAsWebSocketMessages)$' -count=1 -v
PASS
```
3. Re-ran both required handler tests, `go test ./internal/pkg/apicompat -count=1`, and `go build ./...`: all `PASS`.
4. `git diff --check a022d9a3457457463ed06dd025370ad57936e668..HEAD` and `gofmt -d` on every changed final-range Go file: `PASS`.

## Limits

- No real provider, shared Redis, database migration, container deployment, or push was exercised. HTTP forwarding used the production path with an in-process recorder; provider behavior remains unverified.
- The initial QA gate failure concerned missing executable assertions, not a demonstrated production serialization defect; the final retest resolved it.

## Knowledge Promotion
`none`
