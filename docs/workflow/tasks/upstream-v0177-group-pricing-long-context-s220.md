---
task_id: upstream-v0177-group-pricing-long-context-s220
phase: contract-draft
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Group Pricing And Long Context S220

## Task ID

upstream-v0177-group-pricing-long-context-s220

## Role

You are the independent `gpt-5.6-terra` Generator worker. Execute only this
approved contract, do not make architecture decisions, and do not expand scope.

## Goal

Behaviorally port the complete upstream prerequisite and fix chain
`92dcfb5eb -> a0ac5e024 -> f63d168ae -> e9fb5983c` and
`f3d949107 -> b830bc14d -> fd82dfd52` into the locally diverged product. Add
the OpenAI account-level veto and its audit trail, per-group model pricing, and
a group long-context pricing switch. Ensure Grok long-context pricing follows
the group switch without being disabled by the OpenAI-only account setting.

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
  the account switch allow it. Missing/malformed account state defaults safely
  to disabled, create/import/API writes normalize the value, and account/shadow
  synchronization preserves the upstream ownership rules. Grok and other non-OpenAI paths
  do not inherit the OpenAI account veto and follow only the group switch.
- Usage logs persist and expose whether long-context billing was actually
  applied. Migrations 174/175 add that audit field and backfill/guard the OpenAI
  account flag using the repository migration conventions.
- Admin group API/DTO/repository/service and the Groups UI can read and update
  the new fields. Existing image/video/group pricing controls remain usable.
- Migrations `174_add_usage_log_long_context_billing.sql`,
  `175_default_openai_long_context_billing.sql`, and
  `221_group_model_pricing.sql` are additive, idempotent under the
  repository migration conventions, defaults/backfills long-context pricing to
  the approved account/group defaults, and are validated only against
  task-owned disposable fixtures.
- Ent generated code is regenerated from the schema; no hand-edited generated
  drift remains.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen product base: `main@57ba8dc89e3922637616bb3c7dece95c2b1684c0`.
- Upstream: `upstream/main@baeac1f3de21d37b129405f092ef86c24b3f203d`.
- Account-veto prerequisite chain:
  `92dcfb5eb`, `a0ac5e024`, `f63d168ae`, and `e9fb5983c`.
- Group-pricing source chain: `f3d9491071d0dc8093c1c10de37b9ad78007b52f`,
  `b830bc14d655524357360df1e4301b9cf81fb1fc`, and
  `fd82dfd52d31babdceb2d20e0ef1126e508d0f8d`.
- The local migration sequence currently ends at product migration 203; the
  upstream-numbered additive migrations 174, 175, and 221 do not collide with
  existing files in this checkout.
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
- `backend/ent/schema/usage_log.go`
- `backend/ent/usagelog.go`
- `backend/ent/usagelog/**`
- `backend/ent/usagelog_create.go`
- `backend/ent/usagelog_update.go`
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_import_test.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/account_handler_long_context_billing_test.go`
- `backend/internal/handler/admin/admin_service_stub_test.go`
- `backend/internal/handler/admin/openai_oauth_handler.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/group_repo.go`
- `backend/internal/repository/openai_long_context_billing_migration_integration_test.go`
- `backend/internal/repository/usage_log_repo_insert.go`
- `backend/internal/repository/usage_log_repo_query.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_long_context_billing_test.go`
- `backend/internal/service/admin_account.go`
- `backend/internal/service/admin_group.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_spark_shadow_test.go`
- `backend/internal/service/billing_search_audio_cost_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/channel.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/crs_sync_long_context_billing_test.go`
- `backend/internal/service/crs_sync_service.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_usage_billing.go`
- `backend/internal/service/group.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/model_pricing_resolver_test.go`
- `backend/internal/service/openai_alpha_search_billing_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_search_surcharge_test.go`
- `backend/internal/service/openai_videos.go`
- `backend/internal/service/openai_videos_test.go`
- `backend/internal/service/usage_log.go`
- `backend/migrations/174_add_usage_log_long_context_billing.sql`
- `backend/migrations/175_default_openai_long_context_billing.sql`
- `backend/migrations/221_group_model_pricing.sql`
- `backend/migrations/group_model_pricing_migration_test.go`
- `backend/migrations/openai_long_context_billing_migration_test.go`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/constants/channel.ts`
- `frontend/src/i18n/locales/en/admin/channels.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/overview.ts`
- `frontend/src/i18n/locales/zh/admin/channels.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/overview.ts`
- `frontend/src/types/index.ts`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.modelPricing.spec.ts`
- `frontend/src/views/admin/__tests__/groupsImagePricing.spec.ts`
- `frontend/src/views/admin/__tests__/groupsVideoModelPricing.spec.ts`
- `docs/workflow/worker-results/upstream-v0177-group-pricing-long-context-s220-result.md`

## Denied Paths

- `frontend/src/components/account/EditAccountModal.vue` and
  `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` remain
  denied here because they contain the user-owned layout patch; their account
  veto controls are deferred to the already planned S221 baseline flow.
- Other account-component paths not explicitly allowlisted above; `outputs/**`.
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
Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-pricing-long-context-s220/backend
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
go test ./migrations -run '^(TestMigration221|TestOpenAILongContextBillingMigration)' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 migration validation failed' }
go test ./internal/repository -run '^TestOpenAILongContextBillingMigration' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S220 account migration integration failed' }
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

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-pricing-long-context-s220/frontend
pnpm.cmd exec vitest run src/components/account/__tests__/CreateAccountModal.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/admin/__tests__/GroupsView.modelPricing.spec.ts src/views/admin/__tests__/groupsImagePricing.spec.ts src/views/admin/__tests__/groupsVideoModelPricing.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend focused regressions failed' }
pnpm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend typecheck failed' }
pnpm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S220 frontend build failed' }

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-group-pricing-long-context-s220
git diff --check
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S220 conflict or unmerged index found'
}
foreach ($commit in @('92dcfb5eb','a0ac5e024','f63d168ae','e9fb5983c','f3d949107','b830bc14d','fd82dfd52')) {
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

`PASS / Amendment 1` (2026-08-16 11:43 +08:00): the independent Developer
correctly stopped because the local OpenAI-specific record-usage implementation
is in `backend/internal/service/openai_gateway_service.go`, while the prior
allowlist named only the distinct generic `gateway_service.go`. Add the
OpenAI-specific service file; its existing record-usage regression file was
already allowed. This is a topology correction required by the original
success criteria, not a feature expansion. All migration, frontend account,
fingerprint, rollup, dependency, provider, deployment, and database boundaries
remain unchanged. Resume the same Terra Developer from its existing worktree.

`PASS / Amendment 2` (2026-08-16 12:24 +08:00): R1 comparison against the
complete upstream group-pricing chain found that this checkout owns async
OpenAI video estimation in `openai_videos.go`, while upstream kept the relevant
video-unit path in its usage module. Add `openai_videos.go` and its existing
test file so group `billing_mode=video` can use resolution tiers and continuous
seconds without bypassing the established task-reservation billing path. This
is required by the original continuous-media success criterion. No other scope
or database boundary changes.

`FAIL / controller-review S220-R2` (2026-08-16 12:48 +08:00): commit
`61473d06f` resolves the R1 Groups UI, video-unit, i18n, and evidence issues,
but the implementation only reads `openai_long_context_billing_enabled` from
account extra. The checkout still lacks the prerequisite account normalization,
default migration, create/import/API behavior, and usage-log audit field from
`92dcfb5eb` plus its final default/validation fixes. That would make the new
veto unreachable through supported local behavior and leave missing values as
an implicit hard veto without the upstream migration contract. Amendment 3
therefore adds the complete account-veto backend/create/audit chain and
migrations 174/175. The user-owned EditAccountModal delta stays denied in S220
and will be adapted with fingerprint controls in S221's temporary baseline.
Return this bounded correction to the same Terra Developer before QA.
