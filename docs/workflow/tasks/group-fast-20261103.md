---
type: task-contract
status: approved
review_verdict: PASS
task_id: group-fast-20261103
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
spec_ref: docs/workflow/spec.md
last_verified: 2026-11-03
---

# Task Contract

## Task ID
group-fast-20261103

## Role
Controller: Planner, integration and final Evaluator. Independent Contract Reviewer,
Developer and QA follow the repository Agent Matrix model gpt-6.1-sol. Approval
metadata is effective only after an independent PASS binds these exact final bytes.

## Goal
Selectively adapt upstream group-level OpenAI Fast enforcement (PR #6443,
a4fb58e428a0c5d12087d1de2e6d91b56b7db8b2) to the current local architecture.
Pinned upstream main: 4a4a83ee5ec86b19d3a2dd9b6471de5128963bec.
The user authorized "合一下" after the group Fast comparison.

## Context
- Main: F:/mcplugins/sub2api.
- Isolated candidate: E:/codex-worktrees/sub2api/group-fast-20261103.
- Evidence: E:/codex-runtime/sub2api/group-fast-20261103.
- Preserve all current dirty tracked inputs, including existing quota-reset and
  account-stat changes in openai_gateway_service.go. Use their frozen bytes as
  prerequisites; never copy upstream whole files or cherry-pick/merge/rebase.
- Existing pricing is authoritative: priority card when available, existing
  fallback otherwise. The 1.5 long-context output multiplier is unrelated.
- Source migration only: next unused migration number 254, recheck before write.
  Do not execute any migration against a database or update any running service.

## Success Criteria
1. Persist force_openai_fast bool, default false. Expose create/update/admin DTOs,
   trusted domain objects, repository reads/writes, duplication, and API-key
   auth/cache snapshots. Omitted updates preserve the flag; stale caches must
   not silently lose it. Only OpenAI and Composite groups can enable it;
   platform changes normalize incompatible values to false.
2. Add the upstream-style "强制使用 Fast（priority）" toggle to both group create
   and edit forms. Default false; load/save/copy correctly. Explain that global
   Fast/Flex policy can still filter/block and existing WS sessions reconnect.
3. For trusted enabled groups and actual OpenAI upstream accounts, force priority
   even if the client omitted service_tier or supplied default/flex/ultrafast/
   unknown values. Then run the existing global policy normally. Global filter
   removes the final tier; global block remains authoritative. Disabled/untrusted/
   unsupported groups and non-OpenAI upstream accounts retain existing behavior.
4. Cover native Responses, OAuth passthrough, Chat Completions, Messages and WS
   response.create ingress paths, including subsequent WS frames that omit tier.
   Do not modify unrelated WS frame types, replay/billing ownership or routing.
   Group context must remain trusted; do not read a client body/group ID as policy.
5. Leave customer billing logic, all price cards, group multipliers and costs
   unchanged. Forced priority goes through the same established usage/pricing
   flow as client-selected priority. No Free Fast or global force_priority/missing
   rule expansion, and no custom three-state selector.
6. Tests must prove disabled legacy behavior, missing/overridden tier, trusted
   context, Composite-to-OpenAI versus other providers, filter/block precedence,
   create/update/duplicate/cache projection, and both mounted form save flows.
   Run applicable existing gateway, group, cache and billing regressions.
7. Generated ent changes must be limited to the new field. No unrelated generated
   drift or dependency/lockfile changes. Do not weaken existing assertions.

## Allowed Paths
Developer may edit only these product paths in the isolated candidate:
- backend/ent/schema/group.go
- backend/ent/group.go
- backend/ent/group/group.go
- backend/ent/group/where.go
- backend/ent/group_create.go
- backend/ent/group_update.go
- backend/ent/migrate/schema.go
- backend/ent/mutation.go
- backend/ent/runtime/runtime.go
- backend/internal/handler/admin/group_handler.go
- backend/internal/handler/dto/types.go
- backend/internal/handler/dto/mappers.go
- backend/internal/repository/group_repo.go
- backend/internal/repository/api_key_repo.go
- backend/internal/service/group.go
- backend/internal/service/admin_service.go
- backend/internal/service/admin_group_duplicate.go
- backend/internal/service/api_key_auth_cache.go
- backend/internal/service/api_key_auth_cache_impl.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_chat_completions.go
- backend/internal/service/openai_gateway_messages.go
- backend/internal/service/openai_ws_forwarder.go
- backend/internal/service/openai_ws_v2_passthrough_adapter.go
- backend/migrations/254_group_force_openai_fast.sql
- frontend/src/types/index.ts
- frontend/src/views/admin/GroupsView.vue
- frontend/src/views/admin/groupsOpenAIFast.ts
- frontend/src/i18n/locales/zh/admin/groups.ts
- frontend/src/i18n/locales/en/admin/groups.ts
- New focused tests named *group_fast_20261103_test.go under the touched backend
  packages, backend/migrations/group_force_openai_fast_migration_test.go,
  frontend/src/views/admin/__tests__/GroupsView.openaiFast.spec.ts and
  frontend/src/views/admin/__tests__/groupsOpenAIFast.spec.ts.
- E:/codex-runtime/sub2api/group-fast-20261103/developer/** for reports/evidence.

Controller owns this contract, its independent review, bounded additions to
spec/status, evidence helpers, exact integration and final QA report.

## Denied Paths
- All other product files; dependency manifests/locks, pricing/billing sources,
  running configuration, secrets, static build artifacts and knowledge/**.
- Main checkout product sources while Developer or QA is working.
- No add/commit/push/merge/cherry-pick/rebase/reset/clean or deployment.

## Constraints
- Reuse installed dependencies and Go E:/tools/go1.26.10/go/bin/go.exe.
- Independent reviewer must approve the final contract SHA256 before development.
- Developer and QA must be separate invocations; exact candidate sources must be
  frozen and hash-bound for QA, then rechecked before integration.
- Preserve raw original tracked bytes outside approved edits, HEAD, index and
  unrelated untracked files. Stop on conflicting target drift.
- No paid/provider calls, production database access, Docker/service changes.
- Browser automation, if used, must follow dedicated profile/PID/close cleanup.

## Acceptance Commands
- Go focused tests for group/admin/DTO/cache/Fast/WS/OAuth/Chat/Messages and affected
  billing regression names, using -count=1 (both default and unit tags where
  applicable). Record matched test names and require nonzero counts.
- go build ./... from backend using the fixed Go executable.
- Targeted Vitest helper plus mounted GroupsView Fast tests and existing group
  duplicate/model-pricing/messages regressions.
- frontend vue-tsc --noEmit; ESLint on touched TS/Vue; Vite build with output only
  inside evidence. Reuse frontend/node_modules without installing/upgrading.
- git diff --check on the task increment; generated-field review and original
  dirty-input preservation check. Repeat meaningful tests on actual main after
  exact integration. No live migration/provider acceptance is claimed.

## Output
- Developer report: "### PASS: group-fast-20261103", FAIL or BLOCKED, changed paths,
  exact commands/exits/counts, upstream adaptation notes, risks and compliance.
- Independent QA: separate report binds final candidate sources and evidence,
  lists executed checks, failures and uncovered runtime boundaries.
- Final controller report states integrated behavior, validation, preservation
  and no commit/push/deployment. Do not equate builds with live API acceptance.

## Stop Rules
No independent exact-hash contract PASS, model/tool failure, conflicting drift,
required denied-path edit or two failed Developer iterations: stop and report
the precise issue. Do not broaden scope or silently substitute models.

## Budget
Native independent role dispatch. No monetary budget authorization is introduced.
