---
task_id: upstream-v0177-group-pricing-long-context-s220
phase: contract-draft
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Group Pricing And Long Context S220

## Goal

Behaviorally port the complete upstream prerequisite and fix chain
`f3d949107 -> b830bc14d -> fd82dfd52` into the locally diverged product. Add
per-group model pricing and a group long-context pricing switch, preserve the
OpenAI account-level veto only for OpenAI accounts, and ensure Grok long-context
pricing follows the group switch without being disabled by an unrelated OpenAI
account setting.

## Success Criteria

- Groups persist optional `model_pricing` and
  `long_context_pricing_enabled`; the long-context field defaults to true for
  new and existing groups so an upgrade cannot silently flatten active pricing
  ladders.
- Model pricing resolves in the order Group -> Channel -> built-in pricing.
  Invalid group pricing is logged and falls back safely; unrelated local GPT,
  media, profit-control, peak-rate, and channel behavior remains intact.
- Disabling group long-context pricing selects the lowest applicable token
  tier without disabling continuous media-unit pricing. Enabling it uses the
  preset/channel ladder rather than administrator-authored token intervals.
- OpenAI usage applies long-context pricing only when both the group switch and
  the existing OpenAI account switch allow it. Grok and other non-OpenAI paths
  do not inherit the OpenAI account veto and follow only the group switch.
- Admin group API/DTO/repository/service and the Groups UI can read and update
  the new fields. Existing image/video/group pricing controls remain usable.
- Migration `221_group_model_pricing.sql` is additive, idempotent under the
  repository migration conventions, defaults/backfills long-context pricing to
  true, and is validated only against task-owned disposable fixtures.
- Ent generated code is regenerated from the schema; no hand-edited generated
  drift remains.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen product base: `main@57ba8dc89e3922637616bb3c7dece95c2b1684c0`.
- Upstream: `upstream/main@baeac1f3de21d37b129405f092ef86c24b3f203d`.
- Source chain: `f3d9491071d0dc8093c1c10de37b9ad78007b52f`,
  `b830bc14d655524357360df1e4301b9cf81fb1fc`, and
  `fd82dfd52d31babdceb2d20e0ef1126e508d0f8d`.
- The local migration sequence currently ends at product migration 203; the
  upstream-numbered additive migration 221 does not collide with an existing
  file.
- The user explicitly authorized continuing this prerequisite chain and the
  related database-impact work on 2026-08-16. This authorization covers source
  and migration files plus disposable tests, not a shared or production DB.

## Allowed Paths

- `backend/ent/group.go`
- `backend/ent/group/**`
- `backend/ent/group_create.go`
- `backend/ent/group_update.go`
- `backend/ent/migrate/schema.go`
- `backend/ent/mutation.go`
- `backend/ent/runtime/runtime.go`
- `backend/ent/schema/group.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/group_repo.go`
- `backend/internal/service/admin_group.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/billing_search_audio_cost_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/channel.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_usage_billing.go`
- `backend/internal/service/group.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/model_pricing_resolver_test.go`
- `backend/internal/service/openai_alpha_search_billing_test.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_search_surcharge_test.go`
- `backend/migrations/221_group_model_pricing.sql`
- `backend/migrations/group_model_pricing_migration_test.go`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/constants/channel.ts`
- `frontend/src/i18n/locales/en/admin/channels.ts`
- `frontend/src/i18n/locales/en/admin/overview.ts`
- `frontend/src/i18n/locales/zh/admin/channels.ts`
- `frontend/src/i18n/locales/zh/admin/overview.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.modelPricing.spec.ts`
- `frontend/src/views/admin/__tests__/groupsImagePricing.spec.ts`
- `frontend/src/views/admin/__tests__/groupsVideoModelPricing.spec.ts`
- `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`

## Denied Paths

- `frontend/src/components/account/**`, especially the user-owned
  `EditAccountModal.vue` and its test; `outputs/**`.
- Fingerprint convergence, turn-state redesign, group daily rollups,
  migrations 222/223, VERSION, dependency upgrades, CI/release workflows,
  containers, deployment, provider calls, shared databases, production data,
  and every other upstream change.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, QA reports,
  `knowledge/**`, and global memories; these remain Evaluator-owned.

## Constraints

- Work only in the isolated S220 worktree at the contract-approved SHA.
- Adapt behavior to local pricing and monolithic gateway topology; do not
  cherry-pick the upstream commits wholesale and do not overwrite newer local
  pricing, billing, image/video, GPT-5.6, or profit-control behavior.
- Generated Ent files must come from `go generate ./ent` after the schema edit.
- Do not execute migrations against the user's local/shared database. Migration
  validation may use parsing, in-memory fixtures, or a fresh task-owned
  Testcontainers PostgreSQL instance only.
- Keep dependency and lock files unchanged.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/s220-group-pricing/backend
go generate ./ent
if ($LASTEXITCODE -ne 0) { throw 'S220 ent generation failed' }

$focused = '^(' + (@(
  'TestCalculateCostUnified_GroupLongContextToggleUsesPresetLadder',
  'TestResolve_GroupPricingOverridesChannel',
  'TestResolve_GroupLongContextUsesPresetNotCustomIntervals',
  'TestOpenAIGatewayServiceRecordUsage_GroupAndAccountLongContextMustBothAllow',
  'TestOpenAIGatewayServiceRecordUsage_GrokLongContextFollowsGroupToggleOnly'
) -join '|') + ')$'
go test ./internal/service -run $focused -count=10
if ($LASTEXITCODE -ne 0) { throw 'S220 focused service regressions failed' }
go test ./migrations -run '^TestMigration221' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 migration 221 validation failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 complete service failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 complete handler failed' }
go test ./internal/repository -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 complete repository failed' }
go test ./internal/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 complete server failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S220 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/s220-group-pricing/frontend
pnpm.cmd exec vitest run src/views/admin/__tests__/GroupsView.modelPricing.spec.ts src/views/admin/__tests__/groupsImagePricing.spec.ts src/views/admin/__tests__/groupsVideoModelPricing.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend focused regressions failed' }
pnpm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend typecheck failed' }
pnpm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend build failed' }

Set-Location E:/codex-worktrees/sub2api/s220-group-pricing
git diff --check
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S220 conflict or unmerged index found'
}
foreach ($commit in @('f3d949107','b830bc14d','fd82dfd52')) {
  git merge-base --is-ancestor $commit upstream/main
  if ($LASTEXITCODE -ne 0) { throw "missing upstream provenance: $commit" }
}
```

## Output

- Write only
  `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`
  with first line `### DONE: upstream-v0177-group-pricing-long-context-s220`,
  `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit only allowed implementation/tests and the worker result. Report exact
  changed files, real commands, migration fixture boundary, risks, and contract
  compliance.

## Stop Rules

- Stop if a shared/production DB, dependency upgrade, unrelated schema,
  fingerprint/account-modal change, daily-rollup work, provider access, or
  broader upstream merge is required.
- Stop if preserving local pricing behavior is ambiguous or generated Ent output
  leaves the allowlist; request Evaluator review before expanding scope.
- Stop after two failed implementation rounds. Do not integrate, push, deploy,
  update containers, or clean branches/worktrees.

## Contract Review

`PASS / contract-approved` (2026-08-16 11:23 +08:00): the complete upstream
prerequisite chain and local topology were reviewed. The local checkout keeps
the affected record-usage core in `gateway_service.go`, so the allowlist uses
that file instead of the absent upstream `openai_gateway_usage.go`. Migration
221 does not collide with the current migration set, is embedded and auto-run
by the existing runner, and therefore must prove additive/idempotent default
and backfill behavior in disposable fixtures. All five focused upstream test
names are default-tag discoverable at `fd82dfd52`, and all three source commits
are ancestors of the frozen upstream main. Source work is authorized only in a
clean S220 worktree at the approval commit and within the amended allowlist.
