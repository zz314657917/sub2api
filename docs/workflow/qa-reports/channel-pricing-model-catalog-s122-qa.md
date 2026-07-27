# PASS: channel-pricing-model-catalog-s122

## Findings

- No blocking issue remains in the S122 diff.
- Root cause: the gateway model directory only collected schedulable account `model_mapping` keys, while channel pricing-only models already existed in `Channel.SupportedModels()` but were not consulted by `/v1/models` or `/v1/model-catalog`.
- The fix reuses the channel domain boundary and does not add any `kimi-k3` or provider-specific branch.
- The schedulable-account and platform gates remain authoritative, so a pricing row alone cannot advertise an unroutable platform.

## Executed Checks

- Pre-fix regression: failed as expected; account mapping returned `gpt-5.4`, while pricing-only `kimi-k3` was absent and the pricing-only result was `nil`.
- `go test ./internal/service -run "TestGetAvailableModels" -count=10`: PASS.
- `go test ./internal/handler -count=1`: PASS.
- chatgpt2api `go test ./internal/httpapi -run "TestLuoyeIndependentCanvasModelsSupplementCatalogWithGroupModels|TestCanvasModelsUseSub2APIGatewayForBoundUser|TestCanvasModelsFallbackToSub2APIModelsForBoundUser" -count=1`: PASS.
- `go test ./internal/service ./internal/handler -count=1`: handler PASS; service reproduced only five existing aggregate `TestPeakMultiplier*` groups.
- `go test ./internal/service -run "TestPeakMultiplier" -count=1`: PASS, confirming the aggregate failure remains the known timezone/test-order baseline.
- `gofmt`, exact path audit, conflict-marker scan, unmerged-index check, and `git diff --check`: PASS.

## Unverified Risks

- No live Sub2API database/channel configuration or authenticated Studio Bridge session was exercised.
- Model lists retain the existing short cache, so a newly edited channel can take up to the configured model-list TTL to appear without explicit invalidation.
- No deployment or container refresh was performed.

## Recommendation

PASS / source-ready. Publish the scoped S122 commit, then deploy Sub2API before expecting chatgpt2api to see pricing-only models in a live environment.
