### DONE: upstream-openai-empty-capabilities-s238-a

- Baseline: `main@7bfeae6a8`; upstream source reviewed: `40c26f343`.
- Implementation commit: `bd86e3464` (`fix(openai): treat empty capabilities as unset`).
- Empty `[]any`, `[]string`, `map[string]any`, and `map[string]bool` now return `found=false`; missing/nil remains unchanged.
- Non-empty all-false maps and malformed non-container values remain restrictive (`found=true`).
- Focused tests cover all four empty container types plus restrictive false-map and malformed-value regressions.

Acceptance evidence from `backend/`:

- `go test ./internal/service -run 'TestAccountSupportsOpenAIEndpointCapability' -count=10`: PASS.
- `go test ./internal/service -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=1`: PASS (`[no tests to run]`).
- `gofmt -l internal/service/account.go internal/service/openai_images_test.go`: PASS (no output).

Repository checks:

- `git diff --check`: PASS.
- Product/test diff allowlist: PASS; business commit contains exactly `backend/internal/service/account.go` and `backend/internal/service/openai_images_test.go`.
- Upstream ancestry (`40c26f343` reachable from `upstream/main`) and frozen baseline ancestry: PASS.
- Conflict-marker scan over allowed product/test files: PASS.
- No provider, DB, container, deployment, remote write, push, or force operation performed.

Risk: acceptance is source-level/local Go validation only; no real provider or shared-state runtime was used, per contract.
