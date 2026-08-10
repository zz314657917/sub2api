### DONE: streaming-route-cooldown-s208

# Direct Generator Result

## Task ID

`streaming-route-cooldown-s208`

## Contract

`docs/workflow/tasks/streaming-route-cooldown-s208.md`

## Changed Files

- `backend/internal/server/middleware/middleware.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_stream_cooldown_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/stream_error_event_test.go`
- S208 workflow evidence files only

## Implementation

- Added a request-local API-key route-cooldown marker whose classification is
  exactly the existing `429` / `529` / `5xx` policy.
- Gateway and OpenAI streaming error handlers set that marker before emitting
  their existing terminal SSE payload.
- API-key middleware now prioritizes the marker over the committed writer
  status when it applies the selected route's configured cooldown.
- Added regressions for both handler markers and a `200`-writer route switch
  from group `14` to group `3` after a stream `429` marker.

## Commands Run

```text
gofmt -w <six changed Go files> -> PASS
go test ./internal/handler -run '^Test(OpenAI|Gateway)HandleStreamingAwareError_429MarksRouteCooldown$' -count=10 -> PASS
go test ./internal/server/middleware -run '^TestAPIKeyRouteCooldownUsesStreamErrorMarkerAtHTTP200$' -count=10 -> PASS
go test ./internal/handler ./internal/server/middleware -> PASS
go test ./internal/server/routes -run '^$' -> PASS
gofmt -d <six changed Go files> -> PASS (no output)
git diff --check -> PASS
git ls-files -u -> PASS (no output)
```

## Risks

- A started stream is intentionally not replayed. The current client receives
  its original protocol-compatible terminal error; route selection changes on
  later requests while the configured cooldown is active.
- No live provider, deployed service, or production traffic was exercised.

## Knowledge Candidates

`none`
