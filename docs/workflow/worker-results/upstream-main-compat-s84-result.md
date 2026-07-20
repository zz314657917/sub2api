### PASS: upstream-main-compat-s84-generator

## Changed Files

- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_compat_model_test.go`

## Implemented Behavior

- Buffered Anthropic-compatible JSON responses now override filtered upstream
  headers with `application/json; charset=utf-8` immediately before `c.JSON`.
- The buffered response body, usage conversion, and result metadata are
  unchanged.
- Streaming `text/event-stream` header handling is untouched.
- Added a local-signature regression using an upstream SSE response and the
  existing response-header filter; it verifies JSON content type, response ID,
  and converted usage values.

## Commands Run

- `gofmt` on both business/test files: PASS.
- Focused Go test: PASS.
- Broader local `Test(ForwardAsAnthropic|HandleAnthropicBufferedStreamingResponse)`
  selection: PASS.
- Static buffered/streaming header ownership check: PASS.
- `git diff --check`, unmerged-index, and conflict-marker scans: PASS.

## Risks / Deferred Checks

- No live upstream Anthropic/OpenAI smoke was run; the regression uses a local
  SSE response and existing header-filter path.
- The first static regex probe had an incorrect function-boundary pattern; it
  was corrected and the line-level header gate then passed.
- Primary Usage S82 changes and workflow files remain external dirty work and
  are not included in S84.
