### PASS: cf3577a3c-observability

# cf3577a3c observability worker result

## Changed paths

- `backend/internal/handler/openai_chat_completions.go`: records payload-free request-body read diagnostics before the existing stable client error.
- `backend/internal/handler/openai_gateway_handler.go`: does the same for Responses reads, normalizes remote-compaction triggers before protocol selection, and marks exhausted WebSocket failover as an Ops stream error before closing.
- `backend/internal/handler/ops_error_logger.go`: captures bounded terminal SSE frames (including split writes), parses `response.failed`/`error`, infers their effective status, favors context request ID over response headers, and does not convert recovered upstream attempts into request errors.
- `backend/internal/handler/request_body_read_log.go`: new bounded categorization helper; never logs request content.
- `backend/internal/service/openai_compact_body_signal.go`: moves a compaction trigger to the final input item; JSON decoding uses `UseNumber` to retain large numeric literals.
- `backend/internal/util/responseheaders/responseheaders.go`: permits `x-reasoning-included` passthrough.
- New focused tests: `handler/cfport_observability_test.go`, `util/responseheaders/cfport_headers_test.go`.

## Executed checks

- `go test <actual handler GoFiles> internal/handler/cfport_observability_test.go -count=1 -v` — PASS (4 tests).
- `go test ./internal/util/responseheaders -count=1 -v` — PASS (3 tests).
- `go build ./...` — PASS.
- `gofmt` on all owned production/test files — PASS.
- `git diff --check` — PASS.
- Final cleanup: removed the unreachable former recovered-attempt block from `ops_error_logger.go`; reran the actual-handler focused test command — PASS (4 tests).

## Scope and risks

- No commit, push, provider/database/container/deployment action occurred.
- Local `UpstreamFailoverError` lacks upstream structured `Reason` and account-auth `Stage`; WebSocket Ops marking therefore preserves local type/message/status capability without importing absent upstream service contracts.
- The request-owner should retain/add the direct compaction regression in `cfport_request_test.go` for final ordering, deduplication, and `>2^53` literal preservation. This worker requested that coordination.
- Other changed service files and workflow status belong to concurrent owners and were not modified or assessed here.
