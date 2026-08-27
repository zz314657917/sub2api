# Upstream Content Moderation Core Parity S266-A

## Task ID

upstream-content-moderation-core-s266-a

## Role

Planner and Final Evaluator: Codex. Generator: independent
`gpt-5.6-terra` Developer Worker in the clean S266 worktree. QA: an
independent `gpt-5.6-terra` QA Worker after Controller diff review. The Worker
implements only this contract and does not make architecture or release
decisions.

## Goal

Port the remaining self-contained upstream Risk Control/content-moderation
core into the local fork without merging divergent history. Administrators
must be able to configure and observe the same final moderation behavior while
the hot request path remains bounded, proxy-safe, and backward compatible.

## Success Criteria

- `GET/PUT /api/v1/admin/risk-control/config` round-trips normalized per-category
  thresholds. The Risk Control UI exposes all supported categories, current
  values, reset/default behavior, validation, and save/readback without
  confusing test-result thresholds with persisted configuration.
- Runtime status includes the final upstream pre-block counters and key-load
  details: active, checked, allowed, blocked, errors, average latency, available
  keys, total calls, and per-key active/total/success/error/latency data.
  Concurrent updates are race-safe and config refreshes do not hit settings or
  rebuild keyword state for every request.
- Keyword matching uses the upstream compiled matcher/runtime cache behavior,
  preserves case-insensitive substring semantics and configured keyword order,
  supports the existing 10,000-keyword bound, and refreshes immediately after
  config save. The final upstream fail-open behavior is preserved when the
  moderation backend fails.
- Keyword-block logs persist and return `matched_keyword`; local migration
  `237_content_moderation_matched_keyword.sql` is idempotent and repository/UI
  tests prove round-trip and display. Existing rows safely read an empty value.
- Risk Control supports an optional proxy selected from existing admin proxy
  records. `null` update means unchanged, `0` means direct, and a positive ID
  means validated proxy. Test and live moderation use the same selection;
  resolution/client failures never silently fall back direct, and logs/errors
  never expose proxy credentials.
- The Security Audit/Risk Control navigation remains available in admin simple
  mode when the risk-control feature flag is enabled. Existing Prompt Audit,
  group/model scope, multi-key health, image input, notification templates,
  admin auto-ban exemption, cleanup, and legacy config behavior remain covered.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266`
- Frozen local base: `e5b62a9b911f7b3f95d7188c2f7b2aac9cb4aed3`
- Frozen upstream ref: `efb46db0a960fdad94502b1c3a982a0051cf5245`
- Upstream behavior sources: `23f3d426c`, `1b2d8873b`, `815bc6c9b`,
  `8b37ba882`, `948b63c9c`, `0d7b6ae64`, with final fail-open state from
  `af6928a26`.
- Already covered locally and not to replay: initial `fff4a300c`/`0eca600ff`,
  keyword blocking (`75acdfea3`), model scope (`388968dbd`), latest-turn
  deduplication (`f43a36bae`), admin auto-ban exemption (`db145ea56`), and
  notification-template integration.
- Primary worktree has user-owned changes in workflow files,
  `frontend/pnpm-lock.yaml`, and `outputs/**`; none may be copied, staged,
  cleaned, deleted, or overwritten.

## Allowed Paths

- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/content_moderation_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/repository/content_moderation_repo.go`
- `backend/internal/repository/content_moderation_repo_test.go`
- `backend/internal/service/content_moderation.go`
- `backend/internal/service/content_moderation_test.go`
- `backend/internal/service/content_moderation_keyword_matcher.go`
- `backend/internal/service/content_moderation_keyword_matcher_test.go`
- `backend/internal/service/content_moderation_matched_keyword_test.go`
- `backend/internal/service/content_moderation_proxy_test.go`
- `backend/internal/service/content_moderation_runtime_cache_test.go`
- `backend/internal/handler/admin/content_moderation_handler_test.go`
- `backend/migrations/237_content_moderation_matched_keyword.sql`
- `backend/migrations/content_moderation_matched_keyword_test.go`
- `frontend/src/api/admin/riskControl.ts`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/features/prompt-audit/__tests__/integrationSurface.spec.ts`
- `frontend/src/i18n/locales/en/admin/riskControl.ts`
- `frontend/src/i18n/locales/zh/admin/riskControl.ts`
- `frontend/src/views/admin/RiskControlView.vue`
- `frontend/src/views/admin/__tests__/RiskControlView.spec.ts`
- `docs/workflow/worker-results/upstream-content-moderation-core-s266-a-result.md`
- `docs/workflow/qa-reports/upstream-content-moderation-core-s266-a-qa.md`

## Denied Paths

- All OpenAI cyber-policy/session-block/billing/settings/cache owners reserved
  for S266-B, including any `*cyber*` production file.
- `backend/internal/service/cafe_*`, `backend/internal/repository/cafe_*`,
  `backend/migrations/236_*`, and all Pixel Cafe source/assets/tests.
- `backend/ent/**`, schema generation, dependencies, lockfiles, production or
  deploy configuration, `frontend/pnpm-lock.yaml`, containers, shared data,
  provider traffic, staging, push, and `outputs/**`.
- Any file not explicitly listed in Allowed Paths. Workflow status/spec/log and
  `knowledge/**` remain Controller-owned.

## Constraints

- Work only in the S266 worktree and hand-adapt final behavior to the current
  local topology. Do not cherry-pick or merge the upstream commits wholesale.
- Migration number is fixed at `237`; use only `ADD COLUMN IF NOT EXISTS` for
  the matched keyword and do not run it against the shared/local application
  database.
- Reuse the existing `ProxyRepository`, `Proxy`, and shared `httpclient` pool.
  Proxy lookup/build failure is observable and fail-closed for that audit call,
  but the overall content-moderation backend-error policy remains the current
  final fail-open behavior.
- Preserve public API compatibility for old config JSON and old log rows. Do
  not expose API keys, full proxy URLs, credentials, prompt bodies beyond the
  current policy, or new cross-user state.
- Use mocks, `httptest`, SQL mock/static migration tests, and local builds only.
  No real provider, SMTP, Redis, PostgreSQL, container, deployment, push, or
  production call.
- Commit product code/tests separately from the Worker report. Do not merge to
  local `main`.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266/backend
go test ./internal/service -list 'ContentModeration|KeywordMatcher'
go test ./internal/service -run 'ContentModeration|KeywordMatcher' -count=10
go test ./internal/repository -run 'ContentModeration' -count=10
go test ./internal/handler/admin -run 'ContentModeration|RiskControl' -count=10
go test ./migrations -run 'ContentModerationMatchedKeyword' -count=1
go test ./internal/service ./internal/repository ./internal/handler/admin -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266/frontend
pnpm.cmd exec vitest run src/views/admin/__tests__/RiskControlView.spec.ts src/features/prompt-audit/__tests__/integrationSurface.spec.ts
pnpm.cmd run typecheck
pnpm.cmd run build

Set-Location E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266
$taskGoFiles = git diff --name-only HEAD -- backend | Where-Object { $_ -like '*.go' }
if ($taskGoFiles) { gofmt -d $taskGoFiles }
git diff --check
git ls-files -u
git diff --cached --name-only
```

## Output

- Write
  `docs/workflow/worker-results/upstream-content-moderation-core-s266-a-result.md`
  from the standard Worker template.
- First line must be
  `### DONE: upstream-content-moderation-core-s266-a`,
  `### BLOCKED: upstream-content-moderation-core-s266-a`, or
  `### FAILED: upstream-content-moderation-core-s266-a`.
- Report source commits, changed files, commands and discovered test names,
  migration/proxy/fail-open evidence, risks, knowledge candidates, and exact
  contract compliance. Do not paste unrelated long logs.
- The independent QA Worker may create only
  `docs/workflow/qa-reports/upstream-content-moderation-core-s266-a-qa.md`.
  Its first line must be `### PASS: upstream-content-moderation-core-s266-a`,
  `### FAIL: upstream-content-moderation-core-s266-a`, or
  `### BLOCKED: upstream-content-moderation-core-s266-a` and it must not edit
  product files.

## Stop Rules

- Stop if any success criterion requires an S266-B cyber-policy owner, Ent/schema
  regeneration, dependency/lockfile change, production config, shared database,
  provider call, or a path outside the allowlist.
- Stop if the frozen base/upstream refs changed inside the worktree, migration
  `237` already exists with different semantics, proxy support requires a new
  architecture instead of existing repository/client owners, or focused test
  discovery is empty.
- Stop after two failed Developer attempts. Do not broaden scope or silently
  implement the reverted fail-closed behavior.

## Budget

- developer_worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees/sub2api`
- no explicit token or USD budget requested
