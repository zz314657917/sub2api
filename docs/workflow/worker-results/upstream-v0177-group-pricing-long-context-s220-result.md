### DONE: upstream-v0177-group-pricing-long-context-s220

## Verdict

S220 and its two approved amendments are complete in the isolated worktree.
Group model pricing resolves as Group -> Channel -> built-in; the group
long-context switch uses preset token ladders, OpenAI also observes its
account-level veto, and Grok does not. The admin Groups UI now sends a typed
`model_pricing` array instead of raw JSON. Group video pricing uses resolution
tiers and charges `video_count * duration_seconds` usage units while retaining
the established per-request video tier behavior as its fallback.

## Historical Block (Superseded)

The initial S220 report correctly stopped when the local OpenAI record-usage
path was not allowlisted. Amendment 1 added
`backend/internal/service/openai_gateway_service.go`. Amendment 2 then added
the local async video path and its regression file. Both topology blocks are
resolved; this history is retained only as the audit trail, not as a current
implementation or verification limitation.

## R1 Changed Files

- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/channel.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/group.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_videos.go`
- `backend/internal/service/openai_videos_test.go`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/constants/channel.ts`
- `frontend/src/i18n/locales/en/admin/channels.ts`
- `frontend/src/i18n/locales/zh/admin/channels.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.modelPricing.spec.ts`
- `frontend/src/views/admin/__tests__/groupsVideoModelPricing.spec.ts`
- `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`

## Implementation Notes

- `BillingModeVideo` is a first-class resolver, validation, billing, usage-log,
  and UI mode. Its configured tier price is per video-second and
  `CostInput.UsageUnits` carries the video count times duration seconds.
- `openai_videos.go` preserves the legacy `per_request` exact-duration tier and
  base-resolution per-second fallback. `video` mode instead selects the base
  resolution tier (`480p`, `720p`, or `1080p`) and always applies usage units.
- The form conversion functions are exported from `GroupsView.vue` for real
  behavior tests. They convert per-token values to and from per-MTok display
  values, omit entries with no model, set the current group platform, and clear
  token-mode custom intervals before create or edit persistence.
- The new card prop `hideTokenIntervals` prevents group token entries from
  accepting intervals that the resolver intentionally ignores. Existing image,
  image quality, peak-rate, profit, and group controls remain local controls.

## Verification

```powershell
Set-Location backend
go generate ./ent
go test ./internal/service -run '^(TestCalculateCostUnified_GroupLongContextToggleUsesPresetLadder|TestResolve_GroupPricingOverridesChannel|TestResolve_GroupLongContextUsesPresetNotCustomIntervals|TestOpenAIGatewayServiceRecordUsage_GroupAndAccountLongContextMustBothAllow|TestOpenAIGatewayServiceRecordUsage_GrokLongContextFollowsGroupToggleOnly|TestCalculateCostUnified_VideoUsesDurationUnitsAndResolutionTier|TestOpenAIGatewayServiceEstimateOpenAIVideoCost_GroupVideoPricingUsesResolutionAndDuration)$' -count=10
go test ./internal/service -run '^(TestCalculateCostUnified_VideoUsesDurationUnitsAndResolutionTier|TestOpenAIGatewayServiceEstimateOpenAIVideoCost.*|TestOpenAIGatewayServiceRecordUsage_ChannelVideoBillingUsesBaseTierAsPerSecondPrice)$' -count=10
go test ./migrations -run '^TestMigration221' -count=1
go test ./internal/service -count=1
go test ./internal/repository -count=1
go test ./internal/server -count=1
go test ./internal/handler -run '^$' -count=0
go test ./cmd/server -run '^$' -count=0

Set-Location frontend
.\node_modules\.bin\vitest.cmd run src/views/admin/__tests__/GroupsView.modelPricing.spec.ts src/views/admin/__tests__/groupsImagePricing.spec.ts src/views/admin/__tests__/groupsVideoModelPricing.spec.ts
.\node_modules\.bin\vue-tsc.cmd --noEmit
.\node_modules\.bin\vite.cmd build

Set-Location ..
git diff --check
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" <allowed R1 paths>
git merge-base f3d949107 upstream/main
git merge-base b830bc14d upstream/main
git merge-base fd82dfd52 upstream/main
```

All listed Go focused tests, migration fixture test, repository/server and
compile gates passed. The focused Vitest suite passed 3 files / 6 tests;
`vue-tsc --noEmit` and Vite production build passed. Vite emitted only existing
dynamic-import/chunk-size warnings, and Browserslist reported stale local data.

## Migration Boundary And Risks

- Migration 221 was exercised only through the repository's disposable test
  fixture. No shared or production database was opened or migrated.
- Video providers may expose counts under provider-specific keys. The local
  parser accepts `n`, `num_videos`, and `count`, defaulting safely to one; an
  unknown provider key therefore cannot overcharge.
- The `pnpm` wrapper attempted a local dependency bootstrap and generated an
  untracked workspace file plus a lockfile edit. Both task-owned artifacts were
  removed/restored before this result; dependency declarations and lockfiles are
  unchanged.

## Contract Compliance

- The final diff contains only approved S220 paths, including Amendment 2's
  video service and test. No account modal, fingerprint, rollup, migration
  222/223, dependency, CI, deployment, container, provider, `outputs/`, or
  main-worktree path changed.
- No shared database, production data, push, merge, deployment, container
  action, branch cleanup, or worktree cleanup was performed.
- Generated Ent state comes from `go generate ./ent`; no generated file was
  hand-edited.

## Knowledge Candidates

- None. The group pricing and video-unit behavior is task-local contract
  evidence; no durable `knowledge/` update is warranted before evaluator review.
