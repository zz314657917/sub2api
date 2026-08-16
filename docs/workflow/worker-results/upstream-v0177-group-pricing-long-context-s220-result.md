### DONE: upstream-v0177-group-pricing-long-context-s220

## Verdict

Amendment 1 added the local OpenAI record-usage service to the allowlist and
removed the only topology blocker. The completed adaptation persists group
model pricing and the long-context switch, resolves pricing as Group ->
Channel -> built-in, and passes the OpenAI account veto only for OpenAI
accounts. Grok receives no account veto and follows the group switch.

## Evidence

- `backend/internal/service/openai_gateway_service.go:6837` defines
  `OpenAIGatewayService.RecordUsage`.
- `backend/internal/service/openai_gateway_service.go:7296` defines
  `calculateOpenAIRecordUsageTokenCost`; its `CostInput` call has no group
  object or account-level long-context gate.
- `backend/internal/service/gateway_service.go:9227` defines the distinct
  generic `GatewayService.RecordUsage` path named by the allowlist.
- The exact S220 focal regression names are for `OpenAIGatewayService`, so
  they cannot cover the contracted OpenAI/Grok intersection through the
  permitted generic gateway path.

## Commands Run

```powershell
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor f3d949107 upstream/main
git merge-base --is-ancestor b830bc14d upstream/main
git merge-base --is-ancestor fd82dfd52 upstream/main
git status --short
```

All commands succeeded. The worktree was clean before this result file.
Focused/backend/frontend acceptance commands were not run because no legal
implementation path exists under the approved contract.

## Changed Files

- `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`

## Contract Compliance

- No production, shared, or disposable database was used.
- No migration, Ent schema, frontend, account-modal, fingerprint, rollup,
  dependency, CI, deployment, provider, `outputs/`, or source file changed.
- No generated code was edited or generated.
- No branch/worktree cleanup, integration, push, deployment, or container
  operation was performed.

## Required Evaluator Decision

No further contract expansion is required. The original block above is retained
as historical evidence; Amendment 1 approved the required local call site.

## Amendment 1 Implementation

- Added Ent schema/generated state and migration 221 for group
  `model_pricing` and `long_context_pricing_enabled`. The migration uses
  additive `IF NOT EXISTS` columns, a true default, and an idempotent true
  backfill. Validation only parsed the embedded SQL; no database was opened.
- Added repository JSON serialization with invalid persisted JSON logging and
  safe fallback, admin DTO/request/service plumbing, and frontend create/edit
  payload support for the switch and pricing entries.
- Group token pricing removes administrator-authored token intervals and keeps
  the preset model ladder. Disabling the group switch selects the base token
  tier without changing continuous per-request/media-unit billing.
- `GatewayService` and `OpenAIGatewayService` both pass the authenticated group
  to the resolver. The OpenAI record-usage path passes an account veto only for
  `PlatformOpenAI`; Grok uses the group switch alone.

## Acceptance After Amendment 1

```powershell
go generate ./ent
go test ./internal/service -run '^(TestCalculateCostUnified_GroupLongContextToggleUsesPresetLadder|TestResolve_GroupPricingOverridesChannel|TestResolve_GroupLongContextUsesPresetNotCustomIntervals|TestOpenAIGatewayServiceRecordUsage_GroupAndAccountLongContextMustBothAllow|TestOpenAIGatewayServiceRecordUsage_GrokLongContextFollowsGroupToggleOnly)$' -count=10
go test ./migrations -run '^TestMigration221' -count=1
go test ./internal/service -count=1
go test ./internal/handler -count=1
go test ./internal/repository -count=1
go test ./internal/server -count=1
go test ./cmd/server -run '^$' -count=0
pnpm.cmd exec vitest run src/views/admin/__tests__/GroupsView.modelPricing.spec.ts src/views/admin/__tests__/groupsImagePricing.spec.ts src/views/admin/__tests__/groupsVideoModelPricing.spec.ts
pnpm.cmd run typecheck
pnpm.cmd run build
```

The focused service suite passed ten repetitions; migration, full affected Go
packages, server compilation, focused Vitest, frontend typecheck, and frontend
build completed successfully. The frontend package emitted only its existing
`pnpm` configuration warning.
