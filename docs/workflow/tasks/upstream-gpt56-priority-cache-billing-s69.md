# Task Contract: upstream-gpt56-priority-cache-billing-s69

## Task ID

`upstream-gpt56-priority-cache-billing-s69`

## Status

`draft`

## Role

You are the Generator worker adapting only the GPT-5.6 Priority cache-write pricing chain from upstream billing hardening commits.

## Goal

Close the remaining GPT-5.6 underbilling gap: when `service_tier` is `priority`, cache creation tokens for Sol, Terra, and Luna must use the official Priority cache-write price instead of the Standard price, while Standard, Flex, long-context, channel/interval overrides, and explicit-zero behavior remain stable.

## Success Criteria

- `ModelPricing` carries a dedicated Priority cache-creation price, and Priority-tier selection treats that field as a valid dedicated-tier signal.
- `LiteLLMModelPricing` and `LiteLLMRawEntry` carry and parse `cache_creation_input_token_cost_priority`, including a numeric zero without a JSON decode fallback.
- Static GPT-5.6 fallback pricing exposes official cache-write prices per token:
  - Sol: Standard `6.25e-6`, Priority `12.5e-6`.
  - Terra: Standard `3.125e-6`, Priority `6.25e-6`.
  - Luna: Standard `1.25e-6`, Priority `2.5e-6`.
- The embedded pricing JSON contains the same Priority cache-write values for Sol, Terra, and Luna.
- Dynamic LiteLLM pricing maps the Priority cache-write field into `ModelPricing`.
- A non-nil channel or interval `CacheWritePrice` overrides both Standard and Priority cache-write prices with the configured value and marks the override explicit; a configured zero remains zero for both tiers.
- For GPT-5.6 pricing that is not explicitly overridden, missing Standard cache-write price derives from `input * 1.25`, and missing Priority cache-write price derives from `priority input * 1.25`.
- Standard uses the Standard cache-write price; Priority uses the dedicated Priority price; Flex continues applying the existing `0.5` tier multiplier to Standard pricing.
- Session long-context input-side total is `InputTokens + CacheCreationTokens + CacheReadTokens`; only totals `> 272000` trigger the policy, while `== 272000` does not.
- Priority selects the dedicated cache-write price before applying the existing `2.0` long-context input multiplier; Flex uses Standard cache-write pricing and then applies the existing `0.5` tier multiplier.
- Existing cache breakdown 5m/1h semantics and non-GPT-5.6 model pricing remain unchanged; breakdown mode continues reading its 5m/1h fields and does not switch to the new flat Priority field.
- `RecordUsage` with GPT-5.6 cache-creation tokens and `Result.ServiceTier=priority` persists `ServiceTier`, the dedicated Priority `CacheCreationCost`, and matching `TotalCost` / `ActualCost`, proving the real usage-recording path passes the tier into billing.

## Context

- Upstream composite: `de28eba3c fix(openai): harden GPT-5.6 billing and usage`.
- Relevant upstream ancestors: `4a2b10c94`, `383f61d0e`, and `062af81fb`.
- Local commits `492e4cfec` and `3332c6883` already cover usage extraction, explicit-zero cache token parsing, Standard/Flex pricing, long-context behavior, and Sol/Terra/Luna base prices.
- Current local gap: `computeTokenBreakdown` selects Priority input/output/cache-read prices but always sends the Standard cache-creation price to `computeCacheCreationCost`.
- Do not replay the 28-file upstream composite; adapt only this pricing chain to the current unified resolver architecture.

## Allowed Paths

- `backend/internal/service/billing_service.go`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/gpt56_priority_cache_billing_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `docs/workflow/worker-results/upstream-gpt56-priority-cache-billing-s69-result.md`

## Denied Paths

- Usage extraction/apicompat, gateway forwarders, repositories, handlers, DTOs, API key middleware, settings, and billing cache services.
- Unit-tag suites including `billing_service_test.go` and `model_pricing_resolver_test.go`; current unrelated `-tags=unit` compile drift must not be used as S69 evidence or repaired here.
- Bare `gpt-5.6` alias/catalog/OpenCode work, user-scoped Fast/Flex policy, usage breakdown legacy filtering, and Cyber request-type migrations.
- Payment, subscription, balance/quota mutation, frontend, migrations, deployment, and production configuration.
- `fc66a30ff` and all payment-concurrency paths.
- `knowledge/**` and global memories.

## Constraints

- Preserve the current unified `ModelPricingResolver`; do not import upstream file-layout refactors.
- Preserve existing Standard/Flex/long-context and 5m/1h cache breakdown calculations except for selecting the dedicated Priority cache-write price.
- Put all new tier matrix, channel/interval override, explicit-zero, long-context boundary, and 5m/1h preservation assertions in the untagged `gpt56_priority_cache_billing_test.go` or another explicitly allowed untagged test file; unit-tag tests do not count as PASS evidence.
- Channel and interval overrides are authoritative for both Standard and Priority, including zero; do not apply an automatic 2x multiplier to an explicit override.
- Do not add a new database field, setting, API, UI control, model alias, or pricing source.
- Do not change cache token extraction, token bucket separation, usage log schema, payment settlement, or subscription deduction.
- Do not repair unrelated full-suite drift.

## Acceptance Commands

```powershell
Push-Location backend
$requiredPattern = "^(TestGPT56PriorityCacheWriteTierAndLongContextMatrix|TestGPT56PriorityCacheWriteChannelAndIntervalOverrides|TestGPT56PriorityCacheWritePreservesCacheBreakdown|TestOpenAIGatewayServiceRecordUsage_GPT56PriorityPersistsDedicatedCacheWriteCost|TestParsePricingData_ParsesPriorityCacheCreationField|TestDefaultPricingIncludesGpt56PreviewPrices)$"
$listed = @(go test ./internal/service -list $requiredPattern | Where-Object { $_ -match '^Test' })
if ($listed.Count -ne 6) { throw "S69 required tests missing: $($listed -join ', ')" }
go test ./internal/service -run $requiredPattern -count=1
go test ./internal/service -run "^(TestParsePricingData_.*|TestGetModelPricing_.*|TestOpenAIGatewayServiceRecordUsage_GPT56SeparatesCacheWriteForBillingAndStats|TestOpenAIGatewayServiceRecordUsage_ServiceTier.*)$" -count=1
go test ./internal/service/openai_ws_v2 -run "CacheCreation|Usage" -count=1
go test ./internal/service -run "^$" -count=1
Pop-Location
$base = (git merge-base HEAD codex/upstream-latency-health-column).Trim()
$allowed = @(
  "backend/internal/service/billing_service.go",
  "backend/internal/service/pricing_service.go",
  "backend/internal/service/pricing_service_test.go",
  "backend/internal/service/model_pricing_resolver.go",
  "backend/internal/service/gpt56_priority_cache_billing_test.go",
  "backend/internal/service/openai_gateway_record_usage_test.go",
  "backend/resources/model-pricing/model_prices_and_context_window.json",
  "docs/workflow/worker-results/upstream-gpt56-priority-cache-billing-s69-result.md"
)
$changed = @(git diff --name-only "$base..HEAD")
$unexpected = @($changed | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -gt 0) { throw "S69 path audit failed: $($unexpected -join ', ')" }
git diff --check "$base..HEAD"
```

## Output

- Write `docs/workflow/worker-results/upstream-gpt56-priority-cache-billing-s69-result.md` with first line `### DONE: upstream-gpt56-priority-cache-billing-s69`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if the fix requires any Denied Path, schema/migration, payment, subscription, or frontend change.
- Stop if Standard or Flex totals change for the same tokens and prices.
- Stop if an explicit channel/interval cache-write zero becomes a derived official price.
- Stop if the `272000` long-context boundary changes from exclusive to inclusive.
- Stop if cache creation 5m/1h breakdown is made to use the new flat Priority field without an explicit contract decision.
- Stop if the implementation depends on bare `gpt-5.6` alias support.
- Do not repair unrelated full-suite drift.
