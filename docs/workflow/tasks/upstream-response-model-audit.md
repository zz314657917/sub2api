---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-response-model-audit
worker_model: gpt-5.6-terra
base_commit: ebd3db125c52c16496465c7697b4b110f94e4774
spec_ref: docs/workflow/tasks/upstream-response-model-audit-spec.md
---

# Task ID / Role
upstream-response-model-audit. Independent Terra Developer builds after contract approval; separate independent Terra QA runs acceptance; controller reviews and integrates.

## Goal
Implement the authorized response-model audit specification in the existing local monolithic gateway/repository topology. Refer to upstream db0bff82c + 6e34fb09c + c46d07ca0 behavior only. Do not wholesale merge, rebase or cherry-pick.

## Success Criteria
- Observe response-declared models before client-facing model rewriting, using OpenAI response.model/model, Anthropic message.model/model and Gemini modelVersion (including wrapped payloads). Support actual HTTP nonstream, SSE, WS/WS bridge, protocol conversion and retry paths in this checkout.
- First declaration retained unless terminal declaration exists; terminal wins. Track conflicting declarations diagnostically. Isolate each forwarding attempt and WS turn. No inferred model when declaration absent; malformed/non-string values ignored; cap raw observed model at 200 Unicode code points. Optimize model-free delta handling.
- Compare against model actually sent after mapping, fallback requested if absent, using case-insensitive matching and only known Grok 4.5/4.6 latest/build aliases. No modification of payload, billing model, tier, amount, routing or account state from observation.
- Exact Grok alias groups are {grok-4.5, grok-4.5-latest, grok-4.5-build} and {grok-4.6, grok-4.6-latest, grok-4.6-build}; do not collapse other names. Gemini has no universal terminal model-bearing event: each valid modelVersion is observed as terminal, so the latest valid declaration wins. Responses terminal events are response.completed/done/failed/incomplete/cancelled/canceled. Anthropic keeps its first valid declaration.
- WS correlation follows existing relay turn/response-ID ownership. Observe only payloads already attributed to that active turn; explicitly mismatched response IDs never contaminate its observer. ID-less events may use the single active turn, never an ended turn or ambiguous turn. Starting each new attempt/turn resets observation, including after partial output followed by retry. Failed-attempt observations are never copied into a later attempt; if existing logic persists that failed attempt's own usage, its own declaration can be recorded without changing the decision to persist. Test new turns, foreign IDs, ID-less/no-active events, failed-terminal then retry, and terminal override. Gemini reads modelVersion/response.modelVersion/response.response.modelVersion; HTTP/Anthropic nonterminal conflicts retain the first declaration diagnostically.
- Persist nullable upstream_response_model and upstream_model_mismatch in every usage write/read path (including best-effort/batch paths), expose only on AdminUsageLog DTO. Existing/historical records remain NULL. No user-facing DTO field leak.
- Administrator usage list adds raw response label, mismatch/likely-variant badge and tooltip with requested/sent/response. Missing model shows no mismatch assertion. Preserve existing mapped-model display, precision, filters and exports.
- Propagate mismatch filter consistently through list/pagination/count/statistics/charts/export and nullable-aware cache keys. true/false mean IS TRUE/IS FALSE, not unknown. Snapshot/cache/fast-total optimizations must not silently ignore filter.
- Add 246 nullable columns and 247 partial concurrent index, with failed-index recovery in migration runner; no changes to prior SQL. 244/245 belong to active main-worktree work and are forbidden.
- Effective tests on real production functions, dedicated PostgreSQL, frontend and isolated browser; no paid provider access.
- Require production-path paired cases (same route/model/request/tokens with equal vs different observed response model), named TestUpstreamResponseModelHTTPNonInterference, TestUpstreamResponseModelSSENonInterference and TestUpstreamResponseModelWSNonInterference. Assert identical outbound request payload and existing downstream rewriting behavior; identical requested/upstream/billing model, tier/mode, computed usage amounts, selected account, retry/failover counters and account-state mutation calls. Only audit fields may differ. Use captured production Forward/recording results and stub repository/transport, not a parallel toy implementation. Existing baseline decisions need not be redesigned.

## Allowed Paths
- backend/ent/schema/usage_log.go
- backend/ent/migrate/schema.go
- backend/ent/mutation.go
- backend/ent/runtime/runtime.go
- backend/ent/usagelog.go
- backend/ent/usagelog/usagelog.go
- backend/ent/usagelog/where.go
- backend/ent/usagelog_create.go
- backend/ent/usagelog_update.go
- backend/migrations/246_add_usage_log_upstream_response_model.sql
- backend/migrations/247_add_usage_log_upstream_model_mismatch_index_notx.sql
- backend/internal/service/usage_log.go
- backend/internal/repository/usage_log_repo.go
- backend/internal/pkg/usagestats/usage_log_types.go
- backend/internal/repository/migrations_runner.go
- backend/internal/service/upstream_response_model.go
- backend/internal/service/gateway_service.go
- backend/internal/service/antigravity_gateway_service.go
- backend/internal/service/gemini_messages_compat_service.go
- backend/internal/service/gemini_chat_completions_compat_service.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_chat_completions.go
- backend/internal/service/openai_gateway_chat_completions_raw.go
- backend/internal/service/openai_gateway_messages.go
- backend/internal/service/openai_gateway_grok.go
- backend/internal/service/openai_gateway_anthropic_native_pump.go
- backend/internal/service/openai_gateway_chat_completions_anthropic_native.go
- backend/internal/service/openai_gateway_messages_anthropic_native.go
- backend/internal/service/openai_gateway_responses_anthropic_native.go
- backend/internal/service/openai_gateway_responses_chat_fallback.go
- backend/internal/service/gateway_forward_as_chat_completions.go
- backend/internal/service/gateway_forward_as_responses.go
- backend/internal/service/openai_ws_forwarder.go
- backend/internal/service/openai_ws_http_bridge.go
- backend/internal/service/openai_ws_v2_passthrough_adapter.go
- backend/internal/service/dashboard_service.go
- backend/internal/service/openai_ws_v2/passthrough_relay.go
- backend/internal/handler/admin/usage_handler.go
- backend/internal/handler/admin/dashboard_handler.go
- backend/internal/handler/admin/dashboard_query_cache.go
- backend/internal/handler/admin/dashboard_snapshot_v2_handler.go
- backend/internal/handler/dto/types.go
- backend/internal/handler/dto/mappers.go
- frontend/src/api/admin/usage.ts
- frontend/src/api/admin/dashboard.ts
- frontend/src/types/index.ts
- frontend/src/components/admin/usage/UsageFilters.vue
- frontend/src/components/admin/usage/UsageTable.vue
- frontend/src/views/admin/UsageView.vue
- frontend/src/i18n/locales/zh/usage.ts
- frontend/src/i18n/locales/en/usage.ts
- frontend/src/i18n/locales/zh/admin/usage.ts
- frontend/src/i18n/locales/en/admin/usage.ts
- backend/internal/service/upstream_response_model_test.go
- backend/internal/service/upstream_response_model_paths_test.go
- backend/internal/service/openai_gateway_service_test.go
- backend/internal/service/antigravity_gateway_service_test.go
- backend/internal/service/openai_ws_v2/passthrough_relay_internal_test.go
- backend/internal/service/openai_ws_http_bridge_test.go
- backend/internal/service/openai_ws_forwarder_success_test.go
- backend/internal/repository/upstream_response_model_audit_test.go
- backend/internal/repository/upstream_response_model_postgres_test.go
- backend/internal/repository/migrations_runner_notx_test.go
- backend/internal/repository/usage_log_repo_request_type_test.go
- backend/internal/handler/dto/mappers_usage_test.go
- backend/internal/handler/admin/upstream_response_model_audit_test.go
- backend/internal/handler/admin/dashboard_handler_request_type_test.go
- backend/internal/handler/admin/dashboard_handler_cache_test.go
- backend/internal/handler/usage_handler_request_type_test.go
- frontend/src/components/admin/usage/__tests__/UsageFilters.spec.ts
- frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts
- frontend/src/views/admin/__tests__/UsageView.spec.ts
- docs/workflow/worker-results/upstream-response-model-audit-result.md

## Denied Paths
Anything outside exact allowlist; main worktree; existing migrations; billing/pricing policy; account configuration; deployment files; secrets; knowledge and global memories. Ent generation may modify only listed generated files; report unrelated generator drift, do not silently include it. No upstream split-file imports.

Independently reviewed scope amendment: openai_ws_forwarder_success_test.go is allowed only for UpstreamResponseModel/UpstreamResponseModelConflict assertions in existing WS Forward success tests (four added assertions reviewed). No fixture/import/control-flow changes authorized there. Initial out-of-allowlist edit was paused and independently reviewed before being accepted; see contract review amendment.

## Constraints
Work only E:/codex-worktrees/sub2api/upstream-response-model-audit. Preserve all existing main dirty work, including concurrent Pelican changes. No commits by Developer or QA. No dependency changes. Do not create global marshal hooks or rewrite history. Browser uses a task-only profile/PID; no broad process cleanup. QA artifacts/scripts outside repo at E:/codex-runtime/pge/sub2api/upstream-response-model-audit.
Controller may create task workflow artifacts independently; shared main workflow records are not overwritten.

## Acceptance Commands
From backend, after implementing untagged tests prefixed TestUpstreamResponseModel:
```powershell
go test ./internal/service -run '^TestUpstreamResponseModel' -count=1 -v
go test ./internal/service/openai_ws_v2 -count=1
go test ./internal/repository ./internal/handler/admin ./internal/handler/dto ./internal/handler -run 'Test.*(UpstreamResponseModel|UsageLog|Usage|Dashboard|NonTransactionalMigration)' -count=1 -v
go test -tags unit ./internal/service -run '^TestUpstreamResponseModel' -count=1
go build ./...
```
Controller reproduced baseline tagged service fixture compile errors before edits (duplicate stringPtr, outdated billing arities, countTokens arity, proxy removed fields). They are excluded, not repaired. Untagged admin/DTO/repository targeted baseline and build PASS. usagestats matching zero tests is NOT acceptance.
Tests must exercise forwarded payloads, recording and retry/turn ownership, not solely observer helper.

Dedicated PostgreSQL: controller/QA starts named task container sub2api-response-model-audit-pg labeled codex.task=upstream-response-model-audit, postgres:16-alpine, disposable credentials and random loopback port, no volume. Set UPSTREAM_RESPONSE_MODEL_POSTGRES_DSN to that instance and execute:
```powershell
go test ./internal/repository -run '^TestUpstreamResponseModelPostgres' -count=1 -v
```
Test in untagged upstream_response_model_postgres_test.go uses actual migration runner and production repository. DSN set plus failed connect must FAIL. Validate migration twice, old rows NULL, create/batch/query roundtrip, true/false/NULL, paginated totals/stats consistency and partial-index validity/recovery. No shared DB. Cleanup after exact label verification.

Controller/QA PostgreSQL commands: `docker run -d --name sub2api-response-model-audit-pg --label codex.task=upstream-response-model-audit -e POSTGRES_USER=audit -e POSTGRES_PASSWORD=audit_task_only -e POSTGRES_DB=audit -p 127.0.0.1::5432 postgres:16-alpine`; poll `docker exec sub2api-response-model-audit-pg pg_isready -U audit -d audit` until success; obtain `docker port sub2api-response-model-audit-pg 5432/tcp`, set DSN `postgres://audit:audit_task_only@127.0.0.1:<assigned-port>/audit?sslmode=disable`. Credentials belong only to this disposable empty container. Finally inspect exact name/id/label and `docker rm -f sub2api-response-model-audit-pg`; verify gone.

From frontend (install frozen dependencies in isolated worktree if needed, no lockfile edits):
```powershell
pnpm.cmd exec vitest run src/components/admin/usage/__tests__/UsageFilters.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/admin/__tests__/UsageView.spec.ts
pnpm.cmd run typecheck
pnpm.cmd run build
```
Run ESLint without --fix over changed TS/Vue files. Browser QA on actual admin usage page with mock API and full styles at desktop/mobile, independent task profile; test mismatched screenshot example, date variant, missing declaration, mapping chain, filter and export. Record artifact paths, profile and PIDs; close session and verify owned processes gone.

On this host pnpm 11 exec/run may auto-install and reject already-installed locked dependency build scripts. Equivalent direct local entry points are explicitly allowed: `./node_modules/.bin/vitest.cmd run ...`, `./node_modules/.bin/vue-tsc.cmd --noEmit`, `./node_modules/.bin/vue-tsc.cmd -b` then `./node_modules/.bin/vite.cmd build`, and `./node_modules/.bin/eslint.cmd ...`. No dependency/config change is needed; baseline direct Vitest passed 27 tests.

Browser execution recipe (QA owns task-only runtime artifacts): start `node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 62187 --strictPort` in frontend after ensuring port is free; record launched PID. Use `npx.cmd --yes --package @playwright/cli playwright-cli -s=response-model-audit open http://127.0.0.1:62187` then CLI run-code to register `page.context().route('**/api/**', ...)` mocks BEFORE navigating to `/admin/usage`. Seed task context localStorage auth_token=task-mock, auth_user={id:1,role:admin,email:audit@example.test}, token_expires_at future. Fulfill auth/me, public settings/setup, admin usage/list/stats/dashboard/filter endpoints with matching fixture API envelopes from existing UsageView tests; unknown API requests fail closed (no live backend fallthrough). Fixtures include screenshot alias mismatch, exact match, date variant, mapping chain and NULL. Observe fresh snapshot, use actual UI filters and export, assert captured requests/CSV match rows; screenshots at 1440x900 and 390x844. QA records exact fixture script and execution transcript outside repository. `playwright-cli -s=response-model-audit close` and exact owned Vite/browser PID cleanup must pass before QA completion. Browser profile is CLI task-owned, never user default; no real login/provider/API mutation.

Baseline failures must be reproduced, documented and isolated; no skipped/no-tests-to-run success. Missing executable evidence for affected production behavior blocks acceptance. Run gofmt, scoped diff, conflict-marker and staged diff checks before controller commits.

## Output
Developer report first line ### DONE/FAILED/BLOCKED: upstream-response-model-audit, changed files, actual commands/results, protocol coverage matrix, risks and exclusions.
QA report docs/workflow/qa-reports/upstream-response-model-audit-qa.md first line ### PASS/FAIL/BLOCKED: upstream-response-model-audit. Independent QA must not rely only on Developer claims.

## Stop Rules
Missing Terra, required scope expansion, baseline preventing effective validation, unintended billing/security changes, unavailable isolated PostgreSQL/browser, unrelated modifications => stop and report BLOCKED. Controller must review scope amendments before implementation continues. Contract review PASS required before build.

## Integration
Controller groups production changes and evidence into commits. Recheck main HEAD, migration numbers, index, changed-path intersection and new-file collisions. If safe, fast-forward local main; if advanced absorb local main in isolation and rerun affected verification. Never stash/overwrite main dirt. No push/deployment. Revert commits for recovery, no history rewriting.

Local export clarification: existing UsageView export is XLSX, so preserve XLSX format and verify downloaded workbook headers/rows for requested/sent/response/mismatch, including blank for unknown. References above to CSV mean the existing tabular export acceptance, not a requirement to add another format. Export filter propagation alone is insufficient without actual audit columns.

Pre-build migration collision gate: controller checked main migration filenames on 2026-09-17 before development; 246/247 are free, 244/245 are concurrent Pelican migrations. Developer rechecks `Get-ChildItem F:/mcplugins/sub2api/backend/migrations -File | Where-Object { $_.Name -match '^(246|247)[_.-]' }` immediately before first edit and records zero results. Any collision blocks pending controller amendment. Repeat before integration. Invalid partial-index recovery must be tested explicitly, not inferred from clean creation.
