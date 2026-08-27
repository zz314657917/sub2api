# Upstream Content Moderation Cyber Policy Chain S266-B

## Task ID

upstream-content-moderation-cyber-policy-s266-b

## Role

Planner and Final Evaluator: Codex. Generator: an independent
`gpt-5.6-terra` Developer Worker in the existing clean S266 worktree. QA: a
separate `gpt-5.6-terra` QA Worker after the Controller reviews the worker
diff. The worker implements only this approved contract.

## Goal

Port the upstream OpenAI `cyber_policy` hard-block chain to the local divergent
gateway. An upstream policy block must be passed through unchanged across every
locally supported OpenAI protocol, never retried or failed over, then receive
bounded post-response audit, notification, operations-error, billing and
optional session-block handling. Administrators control the local-only session
block and whether cyber events count toward automatic bans.

## Success Criteria

- Upstream `error.code == "cyber_policy"`, including streamed
  `response.failed`, is terminal across Responses, Chat Completions, Messages,
  compatibility fallbacks and WebSocket turns. It must not select another
  account, retry, append a second error body, or contaminate the next WS turn.
- The actual client-visible response remains upstream compatible. Post-response
  work records the event once: the moderation log is durable before email,
  sensitive values are redacted, operations errors preserve the client-visible
  status semantics, and usage is marked `request_type=cyber` with observed
  upstream tokens only. A zero-token policy block never creates a false charge.
- Cyber auditing is controlled by the Risk Control switch plus its configured
  group/model scope, but not by sampling, observation/pre-block mode or the
  ordinary content-moderation enabled flag. The new `cyber_policy_counts_toward_ban`
  default preserves historical counting; when disabled, the current cyber event
  and historical cyber rows are excluded from automatic-ban counts.
- Optional cyber session blocking is disabled by default. When enabled, only a
  successful policy-marked request causes an exact API-key-plus-session entry
  with a bounded configurable TTL; the next matching session is locally
  rejected before upstream I/O, while a different session on the same key is
  unaffected. Missing session identity, cache errors and disabled storage fail
  open. A WebSocket policy mark is cleared between turns.
- Admin settings, Risk Control, request-type filters/tables, API DTOs and
  Chinese/English labels expose only the intended control and status text.
  Existing privacy rules remain intact: no credentials, proxy URL credentials,
  raw prompts beyond the existing redaction policy, or cross-user session state
  are returned.
- Preserve S266-A thresholds/proxy/runtime cache/matched-keyword behavior and
  all current local OpenAI routing, billing, rate-limit, media, scheduling and
  fast-policy behavior not necessary for this policy chain.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266`
- S266-A baseline commits: `c2cd7a0a1`, `bec523227`.
- Frozen local base: `e5b62a9b911f7b3f95d7188c2f7b2aac9cb4aed3`.
- Frozen upstream ref: `efb46db0a960fdad94502b1c3a982a0051cf5245`.
- Behavior source: `b62b573f7` (full cyber policy chain), corrected only by
  `6564d376e` (Risk Control group/model audit scope).
- Do not replay the later `acce29af2` / `b2b2adcf8` transcript compatibility
  rewrite. Its `openAIRequestPayloadView` prerequisite is absent locally and
  it belongs to a later cross-protocol refactor; this contract retains the
  self-contained exact-session policy behavior from the feature source.
- Primary worktree user changes, `frontend/pnpm-lock.yaml`, and `outputs/**`
  remain protected and must not be copied, staged, cleaned or overwritten.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/content_moderation_handler.go`
- `backend/internal/handler/admin/content_moderation_handler_test.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_partial_payload_test.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/openai_gateway_cyber_test.go`
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/ops_error_logger_cyber_test.go`
- `backend/internal/repository/content_moderation_repo.go`
- `backend/internal/repository/content_moderation_repo_test.go`
- `backend/internal/repository/gateway_cache.go`
- `backend/internal/repository/gateway_cache_cyber_test.go`
- `backend/internal/repository/ops_error_where_test.go`
- `backend/internal/repository/ops_repo.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/service/content_moderation.go`
- `backend/internal/service/content_moderation_cyber_test.go`
- `backend/internal/service/content_moderation_email.go`
- `backend/internal/service/content_moderation_test.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/notification_email_service.go`
- `backend/internal/service/notification_email_service_test.go`
- `backend/internal/service/openai_cyber_policy.go`
- `backend/internal/service/openai_cyber_policy_test.go`
- `backend/internal/service/openai_cyber_session_block.go`
- `backend/internal/service/openai_cyber_session_block_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_compat_cyber_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_test.go`
- `backend/internal/service/ops_user_error.go`
- `backend/internal/service/ops_user_error_cyber_test.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/usage_log.go`
- `backend/internal/service/usage_log_cyber_test.go`
- `frontend/src/api/admin/riskControl.ts`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/components/admin/usage/UsageFilters.vue`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/user/UserErrorRequestsTable.vue`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/types/index.ts`
- `frontend/src/utils/usageRequestType.ts`
- `frontend/src/views/admin/RiskControlView.vue`
- `frontend/src/views/admin/__tests__/RiskControlView.spec.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/UsageView.vue`
- `frontend/src/views/user/UsageView.vue`
- `docs/workflow/worker-results/upstream-content-moderation-cyber-policy-s266-b-result.md`
- `docs/workflow/qa-reports/upstream-content-moderation-cyber-policy-s266-b-qa.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, dependencies, lockfiles,
  deployment/container/production configuration, shared data, provider/SMTP/
  Redis/PostgreSQL calls, staging, push and `outputs/**`.
- All Pixel Cafe sources/assets/tests and any file not explicitly allowed.
- New protocol-normalization/transcript architecture, notably
  `openAIRequestPayloadView`, later Grok compatibility work and unrelated
  OpenAI/Codex routing changes.
- Controller-owned workflow status/spec/log and `knowledge/**`.

## Constraints

- Hand-adapt behavior; do not merge or cherry-pick upstream history.
- No schema generation or migration is authorized. Stop if a database enum,
  schema or migration becomes required rather than using the existing local
  request-type storage.
- Session entries must be scoped to a positive API-key ID and a stable explicit
  session signal; never derive a new broad user/IP identity to block requests.
  Cache access must be bounded and fail open.
- A policy block is a response/side-effect invariant, not a generic error-path
  refactor. Do not alter non-cyber failover, billing, moderation, or policy
  outcomes.
- Use mocks, `httptest`, miniredis/local fixtures, static API checks and local
  builds only. No live external service, shared database, container, deploy or
  push operation.
- Commit product code/tests separately from worker and QA evidence. Do not
  merge to local `main`.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266/backend
go test ./internal/service -list 'Cyber|ContentModeration|OpenAI.*Policy'
go test ./internal/handler -list 'Cyber|OpenAI'
go test ./internal/handler/admin -list 'ContentModeration|Cyber|Settings'
go test ./internal/repository -list 'Cyber|ContentModeration|OpsError'
go test ./internal/service -run 'Cyber|ContentModeration|OpenAI.*Policy' -count=10
go test ./internal/handler -run 'Cyber|OpenAI' -count=10
go test ./internal/handler/admin -run 'ContentModeration|Cyber|Settings' -count=10
go test ./internal/repository -run 'Cyber|ContentModeration|OpsError' -count=10
go test ./cmd/server -run '^$' -count=1

Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266/frontend
node node_modules/vitest/vitest.mjs run src/features/prompt-audit/__tests__/integrationSurface.spec.ts src/views/admin/__tests__/RiskControlView.spec.ts
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build

Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266
$taskGoFiles = git diff --name-only HEAD -- backend | Where-Object { $_ -like '*.go' }
if ($taskGoFiles) { gofmt -d $taskGoFiles }
git diff --check
git ls-files -u
git diff --cached --name-only
```

## Output

- Write `docs/workflow/worker-results/upstream-content-moderation-cyber-policy-s266-b-result.md`.
- First line must be `### DONE: upstream-content-moderation-cyber-policy-s266-b`,
  `### BLOCKED: upstream-content-moderation-cyber-policy-s266-b`, or
  `### FAILED: upstream-content-moderation-cyber-policy-s266-b`.
- The independent QA Worker may create only
  `docs/workflow/qa-reports/upstream-content-moderation-cyber-policy-s266-b-qa.md`.
  Its first line must be `### PASS: upstream-content-moderation-cyber-policy-s266-b`,
  `### FAIL: upstream-content-moderation-cyber-policy-s266-b`, or
  `### BLOCKED: upstream-content-moderation-cyber-policy-s266-b`.

## Stop Rules

- Stop for a path outside the allowlist, migration/schema requirement, live
  external-state requirement, missing explicit-session-safe representation,
  empty focused test discovery, or a source behavior that requires the denied
  late transcript refactor.
- Stop after two failed Developer attempts. Do not broaden the chain or weaken
  the no-failover/no-cross-session invariants.

## Budget

- developer_worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees/sub2api`
- no explicit token or USD budget requested
