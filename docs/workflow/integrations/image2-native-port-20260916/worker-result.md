### DONE: image2-native-port

Controller aggregation of completed Terra development slices. This is build
completion, not independent QA or main integration acceptance.

## Current build evidence

- Native allowlist/dispatch and single 404/405 Responses fallback implemented.
- JSON/SSE public-vs-mapped model, URL/base64 format, root MIME, multipart edit
  uploads/mask and image_edit events tested through production ForwardImages.
- Native PNG/JPEG/WebP dimensions, legacy/unknown/configured Responses driver,
  plan gates, empty/error output and partial output error/read/EOF tested.
- Pre-output transport read failure delegates to existing failover; downstream
  writes/client failure/cancellation never authorize replay. Completed image
  count and measured usage are preserved, without counting previews as images.
- Cached-image usage derives bounded explicit details and propagates in SSE.
  RecordUsage clones size maps and adds metadata; fixed prices, exact JSON,
  balance/subscription settlement and stateful request-ID dedupe are covered.
- Preview-only image/per-request pricing is zero; token pricing uses measured
  usage. Candidate precedence and non-image overrides have regressions.

## Fresh controller verification after all writers stopped

- `go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)|TestOpenAIGatewayServiceForwardImages' -count=1`: PASS (0.194s).
- `go build ./...`: PASS (exit 0).
- Changed tracked/untracked Go files: gofmt clean. `git diff --check` and
  `git ls-files -u` clean. Corrected exact-path gate PASS; contract amendment
  independently reviewed. Ignored task/review/report artifacts explicitly read.

## Changed business paths

- backend/internal/service/openai_images_direct.go
- backend/internal/service/openai_images_native_helpers.go
- backend/internal/service/openai_images_responses.go
- backend/internal/service/openai_images_test.go (Responses fixtures only)
- backend/internal/service/openai_images_native_test.go
- backend/internal/service/openai_images_native_edges_test.go
- backend/internal/service/openai_gateway_service.go (image usage hunks)
- backend/internal/service/openai_gateway_record_usage_test.go

## Remaining gates

Independent Terra QA, final review and local integration. Real provider and
account-test parity remain open Epic requirements outside this mock contract.
No real-provider, database, container, deployment, commit or push performed.

## Historical first attempt (superseded failure, retained for traceability)

## Current implementation evidence

- Added native adapter and image-size helper in the approved service allowlist.
- OAuth dispatch now selects the explicit native model allowlist after account
  mapping, changes the target endpoint, and performs a single 404/405 fallback
  to the established Responses route.
- Native JSON/SSE conversion currently handles public model rewriting, data URL
  conversion and image dimensions where decodable.

## Executed commands

- `go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)' -count=1`
  initially passed after the first compile correction, before the full OAuth
  compatibility subset was rerun.
- `go test ./internal/service -run 'TestOpenAIGatewayServiceForwardImages_OAuth' -count=1`
  currently fails. Seven existing Responses-protocol stream/edit fixtures use
  `gpt-image-2` but return Responses SSE payloads; the newly correct default
  native route rejects them as no native output. Only the fixtures whose purpose
  is specifically Responses conversion may be force-routed in tests; the
  remaining suite needs explicit protocol split and native fake-transport tests.

## Missing before acceptance

- Complete direct fake-transport JSON/SSE/edit/404/405/partial-failure tests.
- Complete `RecordUsage` captured repository and billing/dedupe regression for
  cached image tokens, immutable size maps and exact charging.
- Re-run the full contracted selector, build, formatting and scope gates.

No real-provider, database, container, deployment, commit or push action was performed.

## Fix round: native streaming drain and error signaling (2026-09-16)

- On a client write failure, the native Images SSE adapter now marks the
  downstream disconnected, stops all later writes, and continues draining and
  parsing upstream frames. A later completed frame contributes its actual
  image count and measured usage (including cache detail) without replay.
- A still-connected downstream now receives normalized `event: error` for
  native `error`, `response.failed`, and `response.incomplete` payloads. The
  adapter does not attempt a replay after any downstream write attempt.
- New production `ForwardImages` fake-transport regressions cover failed
  started/partial writes followed by completed usage, plus `error` and
  `response.failed` after a partial frame. The partial-write then read-failure
  regression retains the original client-write error and remains non-retryable.

### Commands executed for this fix

- `go test ./internal/service -run 'TestCodexDirectImages|TestOpenAIGatewayServiceForwardImages' -count=1` — PASS.
- `go build ./...` — PASS.
- Changed-file `gofmt -l`, `git diff --check`, and `git ls-files -u` — clean.

This is Generator fix evidence only. Independent QA has not yet retested or
issued a PASS verdict. Real provider, database, container, deployment, commit,
and push remain unverified/out of scope.
