### PASS: studio-bridge-model-scoped-apimart-billing

## Findings
- No blocking defect remains in the scoped Studio Bridge billing change.
- The first implementation was too broad and disabled APIMart cost settlement for Midjourney/Grok; cross-repo HTTP tests caught it before completion. The final rule rejects only fixed-price `gpt-image-2 + apimart_cost`.
- `gpt-image-2-official` and other existing cost-based models retain the `7 * 1.2` conversion.

## Executed Checks
- `go test ./internal/service -run 'TestStudioBridge(APIMartImageChargeUsesSub2APIMultiplierForOfficialModel|APIMartImageChargeRejectsFixedPriceGPTImage2Model|APIMartImageChargeKeepsOtherCostBasedModels|ImageChargeWithoutAPIMartAmountUnitKeepsAmount)$' -count=1` -> pass.
- `go test ./internal/service -run 'TestStudioBridge' -count=1` -> pass.
- `go test ./internal/service -run 'TestPeakMultiplier' -count=1` -> pass; confirms the unrelated peak-rate baseline succeeds in isolation.
- `git diff --check -- backend/internal/service/studio_bridge.go backend/internal/service/studio_bridge_test.go docs/workflow/main-log.md` -> pass.
- Diff precision review -> only the Studio Bridge model/unit guard, focused tests and workflow evidence belong to this task.

## Unverified Risks
- `go test ./internal/service -count=1` was attempted and failed in existing `group_peak_rate_test.go` global-timezone cases; those cases pass alone and no peak-rate files were changed.
- Historical overcharges and already-reserved surcharge records were not mutated or repaired.
- No real database, production balance or live Studio Bridge settlement was exercised.

## Recommendation
- Deploy the paired `chatgpt2api` executor policy first, then deploy this Sub2API fail-closed validation.
- Audit historical `gpt-image-2` surcharge charges separately before deciding on refunds.
