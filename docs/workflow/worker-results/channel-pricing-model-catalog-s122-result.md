# DONE: channel-pricing-model-catalog-s122

## Changed Files

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_hotpath_optimization_test.go`
- S122 workflow and current-task evidence.

## Result

- `GatewayService.GetAvailableModels` now merges account model mappings with the active channel's concrete `SupportedModels()` for the current group and platform.
- A model configured only in channel pricing, such as `kimi-k3`, is returned by both the model list and model catalog as long as the platform still has a schedulable account.
- Empty models, wildcard pricing patterns, other-platform models, inactive/missing channels, and groups without schedulable accounts do not gain catalog entries.
- Existing sorting, deduplication, default fallback, route filtering, channel restriction, billing, and short-cache behavior remain unchanged.

## Verification

- Focused `TestGetAvailableModels*` regressions at `count=10`: PASS.
- Full `internal/handler` package: PASS.
- chatgpt2api Sub2API Canvas model bridge regressions: PASS.
- Aggregate `internal/service` run reproduced only the five pre-existing `TestPeakMultiplier*` timezone/order groups; isolated `TestPeakMultiplier*` run: PASS.
- `gofmt`, diff, allowlist, conflict-marker, and unmerged-index checks: PASS.
