---
status: approved
review_verdict: PASS
task_id: channel-monitor-upstream-20261026
worker_model: gpt-6.1-sol
base_commit: d2cd2835e50e39fe44c905e945ca1994cc8abfbb
spec_ref: docs/workflow/tasks/channel-monitor-upstream-20261026.md
---

# Task ID: channel-monitor-upstream-20261026

## Task ID
channel-monitor-upstream-20261026

## Role
Controller Planner / Final Evaluator; independent Contract Reviewer, Developer and QA per current repository Agent Matrix.

## Goal
User approved upstream-first channel monitoring. Adapt the V1 monitoring behavior from upstream ba81c7fd758e425d67f856d601147ef0c5263d3f into the local architecture, replacing the preceding simplified API-key-only extension. Preserve local user availability/capacity pages, account services and unrelated dirty work.

## Success Criteria
- Nine monitor providers: openai/anthropic/gemini/grok/kimi/zhipu/deepseek/minimax support probes; antigravity supports quota only. MiniMax has no locally supported account platform: expose its probe support, never offer or accept an unsupported quota/account combination. Do not advertise OpenCode GO, whose upstream monitor whitelist is incomplete and whose account architecture is absent locally.
- Upstream api_mode chat_completions/responses (Responses only OpenAI), check_mode probe/quota/quota_probe, account binding/explicit unbinding and jitter flow through validation, service, repository, Ent, admin DTO/API and form. Legacy records/default clients retain probe/chat_completions/jitter=0. Pure quota requires a real eligible account, permits empty endpoint/key, uses a quota placeholder model, never calls LLM or ping. quota_probe persists a primary-model snapshot without overwriting probe status when quota fails.
- Reuse upstream quota snapshot normalization, success/error TTL cache, singleflight, error-to-status and thresholds. Adapt dependency interfaces to existing local GetUsage/QueryUsage/QueryBalance(ctx,id) rather than importing account/provider architecture. Preserve original SSRF/credential/timeout and balance protections. Account platform and local capability must match; disallow foreign/missing/unsupported account combinations. No secrets or credentials in quota snapshots/DTO/errors. Admin quota display is included; public/user responses keep quota snapshots, balances and account identity hidden by default, retaining the local public projection.
- Reuse upstream base-path URL joining and Zhipu endpoint-sensitive path selection (official root, Coding Plan base and OpenAI relay), Gemini explicit user role and AQ key compatibility where implemented in the pinned monitor checker, Anthropic block text extraction and Responses reasoning fallback. Preserve SSRF DNS/dial protections and deny query/fragment/userinfo/private endpoints.
- Reuse upstream request-template APIMode behavior, body validation and protected merge keys. Provider/protocol-mismatched templates cannot corrupt monitor requests. Preserve existing snapshot/apply controls and request-race protections.
- Reuse upstream jitter bounds and timer scheduling; prevent concurrent same-monitor runs and preserve cancellation/shutdown. Test min intervals, old defaults, enable/disable and quota mode with no API key.
- Add one new additive migration 253_channel_monitor_upstream.sql for providers, api_mode, check_mode, account_id, jitter and quota history JSON as needed, with legacy-safe defaults. Update schema and all affected generated Ent metadata consistently; do not hand-copy upstream unrelated entity schemas. Archive then remove only the previous task-owned untracked 129_add_channel_monitor_providers.sql; all other existing migrations immutable. Do not execute any migration or database mutations.
- Frontend: upstream provider icons/labels, protocol and check-mode controls, account selection/restore/unbind, endpoint defaults, jitter and admin usage snapshot display. Preserve persisted edit values on load (no watcher key/template reset), guard stale account/list responses, show invalid stored accounts explicitly. User quota placeholder gets a localized label. Keep local capacity pools/availability lists and template request race protections. Chinese and English strings complete; mobile controls wrap without overflow.
- Meaningful tests cover nine-platform capability matrix, default/replace/merge request behavior, paths/Responses reasoning, CRUD/defaults/account bind/unbind/eligibility, quota normalization/cache/error handling, service-to-history/repository persistence, privacy projections, jitter, form saved-state restore and rapid provider/account switching. Port relevant upstream tests; do not simply replace local suites or drop failed assertions.

## Allowed Paths
- backend/internal/domain/channel_monitor_quota.go
- backend/internal/service/channel_monitor*.go (excluding channel_monitor_v2*)
- backend/internal/repository/channel_monitor*.go (excluding channel_monitor_v2*)
- backend/internal/handler/admin/channel_monitor*.go
- backend/internal/handler/dto/channel_monitor.go
- backend/internal/handler/channel_monitor_user_handler_test.go
- backend/internal/handler/channel_monitor_quota_privacy_test.go
- backend/internal/service/wire.go
- backend/cmd/server/wire_gen.go
- backend/ent/schema/channel_monitor.go
- backend/ent/schema/channel_monitor_history.go
- backend/ent/schema/channel_monitor_request_template.go
- backend/ent/** (generated consequences of the three approved schema changes only)
- backend/migrations/253_channel_monitor_upstream.sql
- backend/migrations/channel_monitor_upstream_test.go
- backend/migrations/129_add_channel_monitor_providers.sql (previous task-owned untracked draft: archive/delete only)
- frontend/src/api/admin/channelMonitor.ts
- frontend/src/api/admin/channelMonitorTemplate.ts
- frontend/src/api/channelMonitor.ts (localized quota model label/types only; never expose account snapshots)
- frontend/src/constants/channelMonitor.ts
- frontend/src/composables/useChannelMonitorFormat.ts
- frontend/src/components/admin/monitor/**
- frontend/src/components/user/monitor/ProviderIcon.vue
- frontend/src/components/user/monitor/MonitorCard.vue
- frontend/src/components/user/monitor/MonitorCardGrid.vue
- frontend/src/components/user/monitor/MonitorHero.vue
- frontend/src/components/user/monitor/MonitorTimeline.vue
- frontend/src/components/user/monitor/__tests__/**
- frontend/src/components/common/MonitorQuotaView.vue
- frontend/src/components/common/__tests__/MonitorQuotaView.spec.ts
- frontend/src/i18n/locales/zh/monitorCommon.ts
- frontend/src/i18n/locales/en/monitorCommon.ts
- frontend/src/i18n/locales/zh/admin/channelMonitor.ts
- frontend/src/i18n/locales/en/admin/channelMonitor.ts
- docs/workflow/worker-results/channel-monitor-upstream-20261026.md
- docs/workflow/qa-reports/channel-monitor-upstream-20261026.md

## Denied Paths
All other product files; dependencies/locks, account/provider architecture and account services, V2/passive monitor, settings/security policy/production configuration, other migrations, gateway/billing/payment/store/Pelican, knowledge and global memory. Do not import duplicate-monitor APIs or upstream-only application dependencies.

## Constraints
Isolated candidate E:/codex-worktrees/sub2api/channel-monitor-upstream-20261026 was created at base HEAD plus existing dirty prerequisites (53 overlays in E:/codex-runtime/channel-monitor-upstream-20261026/candidate-overlay.json). External evidence E:/codex-runtime/channel-monitor-upstream-20261026. Main baseline protects 12176 files and HEAD/index via main-before.json. Controller owns workflow changes and final exact integration, Developer never edits main. Separate contract review and QA. No commit/push/deploy/container manipulation/paid-provider call/real DB/migration execution. localhost httptest/sqlmock/in-memory repositories and mock-API browser checks are allowed. No dependency installation or manifest updates.

## Acceptance Commands
Use E:/tools/go1.26.10/go/bin/go.exe with GOROOT=E:/tools/go1.26.10/go, GOTOOLCHAIN=local, GOMAXPROCS=4 (inherited GOROOT points to 1.26.3 and must be overridden per process). Existing candidate frontend/node_modules junction points to existing main modules; never install dependencies. Capture all full outputs externally.
- backend: go test -json -p 1 ./internal/service ./internal/repository ./internal/handler/admin ./internal/handler -run "Monitor|Quota|RunCheck|JoinURL|ExtractOpenAI|ExtractAnthropic" -count=1
- backend: same command with -tags unit; no skipped new unit or persistence tests.
- Enumerate every Test* declaration in added/changed task-owned test files and compare against go-test JSON events. Execute any names not covered by the broad command explicitly in both applicable tag modes. QA must verify this inventory and reject missing execution or skipped new tests. This includes upstream TestQuotaFetcher_*, TestRunCheck_Quota*, endpoint/body/extraction tests and added local integration probes.
- backend: go test -p 1 ./migrations -run "ChannelMonitor" -count=1 (static only, no database).
- backend: go build -p 1 ./...
- frontend: node node_modules/vitest/vitest.mjs run src/components/admin/monitor/__tests__ src/components/user/monitor/__tests__ src/components/common/__tests__/MonitorQuotaView.spec.ts src/views/user/__tests__/ChannelStatusView.capacityPools.spec.ts
- frontend: node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
- frontend: node node_modules/eslint/bin/eslint.js <all changed frontend TS/Vue paths>
- frontend: node node_modules/vite/bin/vite.js build --outDir E:/codex-runtime/channel-monitor-upstream-20261026/dist
- gofmt verification on changed Go files; git diff --check on task delta.
- Independent API/normalizer/privacy/persistence probes and task-owned isolated browser mock-API smoke at desktop/mobile, covering restore/switch/save and quota display. Capture session/profile/PID and screenshots, close browser/server and verify exact task processes are gone.
- Frozen candidate SHA manifest, actual changed paths and diff review; after independent QA PASS recheck original main target bytes before integration, preserve unrelated bytes and HEAD/index, repeat core commands against combined main source. Baseline failures must be independently reproduced, never silently dropped or waived; environment blockers remain explicit.

## Output
Developer result first line ### DONE: channel-monitor-upstream-20261026, ### FAILED or ### BLOCKED. Include exact changed paths and SHA freeze, upstream behavior sources/adaptations, executed commands/results, full log paths, risks and knowledge_candidates. QA independently starts ### PASS/FAIL/BLOCKED: channel-monitor-upstream-20261026; verifies both contract and implementation, generates independent evidence and frozen hashes. Do not infer PASS from Developer reports.

## Stop Rules
Stop on target-byte drift, unapproved dependency/schema/provider/security expansion, unavailable configured model, or baseline blockers that cannot be isolated. Return ownership issues to Controller; don't rewrite original failing oracles. Two implementation failures require Controller reassessment. Never label mocks as real provider/DB/runtime acceptance.
