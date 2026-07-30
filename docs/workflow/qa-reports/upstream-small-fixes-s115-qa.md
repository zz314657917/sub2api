### PASS: upstream-small-fixes-s115

## Findings

- Blank or whitespace-only pricing remote URLs now skip scheduler startup while
  non-empty invalid URLs still use the existing validation path.
- Grok pool-mode accounts no longer receive a durable/runtime cooldown for 5xx;
  non-pool accounts retain the two-minute cooldown.
- The new Grok regression runs under the default Go test build because the
  existing `//go:build unit` file has unrelated compile drift in this branch.

## Executed Checks

- `go test ./internal/service -run "TestPricingSchedulerBlankRemoteURLDoesNotStart|TestPricingNonEmptyInvalidRemoteURLStillReturnsValidationError|TestHandleGrokAccountUpstreamError5xxRespectsPoolMode" -count=1`: PASS.
- `gofmt -d internal/service/pricing_service.go internal/service/pricing_service_test.go internal/service/openai_gateway_grok.go internal/service/openai_gateway_grok_s115_test.go`: PASS.
- `git diff --check`: PASS.
- `git diff --name-only --diff-filter=U`: PASS, no unmerged paths.
- Exact path audit: business changes are limited to pricing scheduler files and
  the Grok source/new S115 regression file.

## Unverified Risks

- `go test -tags=unit ./internal/service` remains blocked by pre-existing
  `stringPtr` duplication, stale billing helper signatures, and missing
  runtime-block helpers in existing unit tests; those files are outside S115.
- No live Grok upstream request, deployment, or container refresh was run.

## Recommendation

`PASS / source-only`; integrate the pricing and Grok commits locally, with the
unit-tag aggregate drift kept as a separate maintenance task.
