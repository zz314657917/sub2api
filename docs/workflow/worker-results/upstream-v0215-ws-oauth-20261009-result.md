### DONE: upstream-v0215-ws-oauth-20261009

## Scope / contract compliance
- Workspace only: `E:/codex-worktrees/sub2api/upstream-v0215-20261009`.
- Read approved contract, independent PASS review and batch spec. Adapted the six referenced commits into existing monolithic owners; no whole-file upstream replacement or absent profit-gate architecture.
- No writes to F tree, other Developers' owners, pricing/ledger/schema, dependencies, original fixtures or overlay. No commit/push/deploy/provider/shared DB/Redis/container operation.
- Completion is Developer implementation/check evidence, not the independent QA/final evaluator verdict.

## Changed paths
Production:
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/handler/openai_gateway_handler.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_gateway_service.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_codex_transform.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_oauth_web_search_history.go` (new)
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_forwarder.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_v2/passthrough_relay.go`

New tests:
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/handler/openai_v0215_ws_billing_test.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_v0215_oauth_history_test.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_v0215_ws_http_test.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_v0215_ws_local_test.go`
- `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend/internal/service/openai_ws_v2/passthrough_v0215_image_test.go`

Report: this file.

## Implemented behavior / Findings
- Later turns refresh through existing `APIKeyService.GetByKey`; adoption requires unchanged key ID, group ID/object ID, platform and subscription type. Failure/changed identity uses frozen connection pricing, not the preceding refreshed turn.
- Pricing-only snapshots retain connection user, quota, route/managed binding, group routing and subscription ownership. Copy multiplier/peak/long-context/image price fields and deeply clone model-pricing pointers/intervals. Turn-indexed snapshots are consumed by AfterTurn without discarding still-pending older turns; async closure captures its key before submission. Existing captured trial/release, RequestStartedAt, request IDs/hash, QuotaPlatform and ChannelUsageFields remain unchanged.
- Local passthrough has **no BeforeTurn calls, including later turns**; it calls BeforeRequest on subsequent response.create frames. Refresh therefore also runs at that existing handler hook, with idempotent begin so ctx-pool BeforeRequest/BeforeTurn do not fetch twice. No new concurrency-slot or authorization lifecycle was introduced. First pricing is frozen before ingress, and unseen turns have a connection-snapshot fallback.
- Relay reads Responses/Chat image details and hosted image_gen fallback (usage wins), carries image input in per-turn/session aggregates and both adapter result mappings. Completed response IDs prevent completed/done/retransmitted terminal events from accumulating or invoking billing twice; duplicate wire frames are still forwarded.
- HTTP 400 recovery adds only `thinking_signature_invalid` with all three exact verification/decrypt phrases to the existing encrypted-content retry condition. Existing trim/no-content behavior, bounded retry flag, HTTP/WS dispatch and replay/output boundaries remain in place.
- OAuth history injection declares cached-only `web_search`; Lite extends/creates input.additional_tools before a trailing compaction_trigger. Existing declarations/explicit choice are respected; no-tools callers pin to none. Compact and API-key callers retain existing normalization byte output. Raw helper keeps large JSON integers exact and malformed/empty input unchanged.
- Self-review found no remaining explicit implementation defect within these criteria.

## Executed checks
All Go commands below ran from `E:/codex-worktrees/sub2api/upstream-v0215-20261009/backend`.

1. `go test -tags unit ./internal/service/openai_ws_v2 -count=1 -v`
   - Exit 0; 39 top-level tests PASS, including 2 own TestV0215 tests.
   - `E:/codex-runtime/upstream-v0215-20261009/ws-oauth-relay.log`
2. `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/handler ./internal/service -run V0215 -count=1 -v`
   - Exit 0; 14 top-level tests PASS: 4 own handler, 6 own service, 4 parallel protocol-worker tests (not credited as own coverage). No no-tests-to-run acceptance.
   - Own total across packages: **12 top-level TestV0215 functions actually executed**, plus their subtests.
   - `E:/codex-runtime/upstream-v0215-20261009/ws-oauth-focused.log`
3. `go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/handler ./internal/service -run '^TestV0215(WS|OAuth)' -count=5`
   - Exit 0; own 10 handler/service tests repeated five times, including loopback WS and actual delayed worker.
   - `E:/codex-runtime/upstream-v0215-20261009/ws-oauth-repeat.log`
4. Preservation regression:
   ```powershell
   go test -tags unit -overlay E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json ./internal/service -run 'Test(OpenAITokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated|GatewayTokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated|CalculateImageCost_|GetImageUnitPrice_|DefaultBuild_.*Image|ExtractOpenAIUsage_CapturesImageInputTokens|MergeOpenAIUsage_CarriesImageInputTokens|BuildOpenAIUsageJSON_CarriesImageInputTokens|Resolve_GroupPricingOverridesChannel|OpenAIGatewayServiceRecordUsage_(UsesUserSpecificGroupRate|PeakRateAffectsTokenModeImageOutputTokens|DuplicateUsageLogSkipsBilling|DuplicateBillingKeySkipsBillingWithRepo|BillingUsesDetachedContext|BillingRepoUsesDetachedContext|ImageCacheRead)|OpenAIGatewayService_Forward_HTTPIngressRetries.*InvalidEncryptedContentOnce)' -count=1 -v
   ```
   - Exit 0; 32 existing top-level tests PASS. Covers GroupID-only channel pricing, group override, user multiplier, image input/cached-image/subscription dedup, peak token pricing, detached settlement context and existing encrypted HTTP recovery.
   - `E:/codex-runtime/upstream-v0215-20261009/ws-oauth-preservation.log`
5. `go build -o E:/codex-runtime/upstream-v0215-20261009/ws-oauth-server.exe ./cmd/server`
   - Exit 0; external binary exists, 149214208 bytes. Not started/deployed.
6. `git diff --check`: exit 0. `gofmt -l` on all 12 owned Go paths: empty.
7. SHA256 comparison against `E:/codex-runtime/upstream-v0215-20261009/baseline.json`:
   - 452 baseline files selected under ent/schema, migrations, service pricing/billing/model_pricing/usage_billing/channel/group/welfare/api_key prefixes; **0 mismatches**.
   - Reviewed tracked owner deltas as minimal behavioral patches (191 additions / 14 deletions across six pre-existing production files); separate new helper/test files are explicitly listed above.

## Runtime simulations / boundaries
- Fake auth lookup: matching identities, lookup failure/nil data, missing connection metadata, changed key/group/platform/subscription, routing/quota/user preservation.
- Interleaved turn 2/3/4 begins before older AfterTurn consumption; nested pricing mutation after snapshots; delayed closures retain distinct prices.
- Actual UsageRecordWorkerPool blocked before submission, then a later turn refresh and parent cancellation: queued task retains earlier group price, ledger/client IDs and captured trial restoration.
- Loopback HTTP through real Forward: precise signature recovery, second rejection bounded to two attempts, no encrypted content, missing either phrase, other code/status and existing invalid_encrypted_content behavior. Replay ID stays removed and summary survives.
- Loopback HTTP OAuth normal/passthrough/Lite/compact/API-key payloads. Compact test uses its supported non-streaming shape.
- Real coder/websocket loopback gateway and upstream, two turns for API-key/OAuth/OAuth-Lite: upstream wire declaration, usage priority/hosted fallback, both billed turn adapter values, and duplicate completed/done wire forwarding without a duplicate AfterTurn.
- Relay fake frames also check session totals, cache-write/read preservation, Chat details, top-level usage/fallback, negative/invalid fields and a late duplicate after another completed response.

## Baseline blockers / limitations / risks
- Unoverlaid service tagged suite has pre-existing fixture compile drift. Used controller-provided exact six-file overlay, never changed it or repository legacy tests; no production replacements/deletions. Original errors and full replacement recipe are in:
  - `E:/codex-runtime/upstream-v0215-20261009/baseline-checks.md`
  - `E:/codex-runtime/upstream-v0215-20261009/create-baseline-overlay.py`
  - `E:/codex-runtime/upstream-v0215-20261009/fixture-overlay/overlay.json`
- Do not interpret focused/overlay PASS as a full unoverlaid service-suite PASS or as email-bind fixture runtime verification.
- No live auth cache invalidation, scheduler/DB settlement, provider, shared database/Redis, container, browser or deployment acceptance. Cache/ledger behavior is exercised with local fakes/stubs and unchanged-source protection checks.
- Race detector was not run (`go env CGO_ENABLED` reports 0); repeated tests and deterministic interleaving are not a race-detector claim.
- Duplicate suppression retains one entry per completed explicit response ID for the relay lifetime, increasing memory with completed-turn count. Id-less upstream correlation keeps the existing fallback semantics; no synthetic billing identity is introduced.
- Independent QA and final main-controller integration/conflict checks remain required. F tree was not touched.

## knowledge_candidates
- None promoted. Local passthrough's BeforeRequest-only lifecycle is recorded here for evaluator review, not written to knowledge/memory.
