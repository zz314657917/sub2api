### DONE: upstream-v0176-grok-pricing-s215

Commit: pending

Changed files:

- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_grok_fallback_test.go`

Implemented a self-contained `grok-4.5` static fallback card after dynamic
PricingService lookup, including strict `> 200000` session-context pricing.
Known `grok`, `grok-latest`, `grok-4.3`, and Build aliases retain local rates.
The provider-prefix normalizer removes `xai/`, `x-ai/`, and `grok/` before
matching. Unknown numeric/versioned and Build/Composer text aliases use 4.5;
media, voice, realtime, and search families remain fail-closed.

Executed checks:

- `go test -list '^TestGetModelPricing_(Grok45FallbackCard|UnknownGrokTextFallback)$' ./internal/service`
  discovered both required tests.
- `go test ./internal/service -run '^TestGetModelPricing_(Grok45FallbackCard|UnknownGrokTextFallback)$' -count=10`
  passed.
- `go test ./internal/service -count=1` passed in 61.300s.
- `go test ./internal/server -count=1` passed.
- `go test ./cmd/server -run '^$' -count=0` passed.
- `gofmt -d internal/service/billing_service.go internal/service/billing_service_grok_fallback_test.go`
  produced no output.
- `git diff --check`, unmerged-index check, and conflict-marker scan passed.

Note: the contract's discovery verification uses a PowerShell array directly
with `-notmatch`, which produces a false missing-test result despite `go test
-list` printing both names. I joined the captured output into a string before
the same two required-name checks; discovery and all focused runs passed.

Risks: static 4.5 fallback is intentionally used only when dynamic pricing
does not resolve the model. No provider, Realtime, group pricing, migration, or
frontend path was modified.
