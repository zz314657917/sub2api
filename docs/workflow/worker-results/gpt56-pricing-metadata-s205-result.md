### DONE: gpt56-pricing-metadata-s205

# Worker Result

## Task ID
gpt56-pricing-metadata-s205

## Status
`done`

## Summary
- Cherry-picked the reviewed 15-line GPT-5.6 pricing metadata patch as `78ad91f56`, retaining its upstream source footer.
- Preserved Batch/Flex cache-write and long-context metadata in the dynamic pricing parser without adding a Batch billing path or changing tier selection.
- Extended the real bundled pricing fixture regression for Sol, Terra, and Luna to assert standard, priority, cache, Batch/Flex cache-write, and long-context values directly from `parsePricingData`.

## Changed Files
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `docs/workflow/worker-results/gpt56-pricing-metadata-s205-result.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
```text
go test ./internal/service -run 'TestParsePricingData|TestDefaultPricingIncludesGpt56|Test.*GPT56|Test.*GPT5' -count=1 -> PASS
go test ./internal/handler -run 'TestOpenAIResponsesWebSocket|TestOpenAIWS' -count=1 -> PASS
go test ./internal/service -run 'TestDefaultPricingGPT56FeedsBillingTierAndLongContextMatrix' -count=1 -> PASS
gofmt -d backend/internal/service/pricing_service.go backend/internal/service/pricing_service_test.go -> PASS, no diff
git diff --check -> PASS
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 0.329s
ok github.com/Wei-Shaw/sub2api/internal/handler 1.095s
ok github.com/Wei-Shaw/sub2api/internal/service 0.073s
```

## Risks
- The new Batch/Flex cache-write values are retained metadata only; no request classification or billing selection consumes them, by contract.
- No deployment, container refresh, real provider request, or production pricing refresh was performed.
- An initial focused-test attempt exposed an adjacent test assignment edited at the wrong occurrence; it was restored before the passing rerun.

## Knowledge Candidates
- Dynamic pricing fields must be represented in both `LiteLLMRawEntry` and `LiteLLMModelPricing`, then explicitly copied by `parsePricingData`; JSON-only updates otherwise silently lose metadata.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
