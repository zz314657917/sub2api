# Task Contract: standard-group-time-rate-s211

## Task ID

`standard-group-time-rate-s211`

## Status

`approved / fix amendment`

## Role

Direct Codex implementation, independent QA worker review, and final Codex
evaluation. The Generator must follow this contract and may not expand scope.

## Goal

Allow standard (non-subscription) groups to use the existing daily time-window
rate factor. The factor multiplies the already-resolved token rate, uses the
request start time, and returns to factor `1.0` outside the window so the final
rate returns to the original effective rate without disabling the group.

## Success Criteria

- Enabled standard and subscription groups apply `peak_rate_multiplier` within
  the server-timezone interval `[peak_start, peak_end)` and apply factor `1.0`
  outside it.
- Standard groups require an enabled factor strictly greater than zero;
  subscription groups retain the existing zero-factor compatibility.
- Disabled configurations retain valid same-day window values and a valid
  factor for later re-enablement; invalid values are normalized safely.
- Gateway, OpenAI, Gemini, image, embedding, chat, Responses, WebSocket, and
  video usage paths carry one captured request-start timestamp through retry,
  failover, worker-pool, and asynchronous billing boundaries.
- Token billing, balance deduction, and cost-based API-key quota use the
  multiplied rate. Per-request image/video pricing remains independent.
- Existing `peak_rate_*` storage and API fields are reused; API-key auth
  snapshots, billing declarations, and available-channel payloads preserve and
  apply the standard-group configuration without API-key recreation.
- Admin create/edit forms expose the feature for both group types, preserve
  values across type switches, show server timezone/formula guidance, and
  enforce type-specific factor rules.
- Focused and package-level Go tests, frontend component tests, typecheck,
  production build, lint, browser evidence, formatting, scope, and Git
  integrity gates pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen base: `main@79426771b97cd382500711718fadf89cebf0b982`
- Workflow fact source: `docs/workflow/status.md`
- Existing rate implementation: `backend/internal/service/group.go`
- Existing admin UI: `frontend/src/views/admin/GroupsView.vue`
- Existing untracked `outputs/` belongs to the user and must remain untouched.

## Allowed Paths

- `backend/internal/service/group.go`
- `backend/internal/service/group_peak_rate_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_record_usage_test.go`
- `backend/internal/service/gateway_peak_per_request_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_videos.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_peak_rate_test.go`
- `backend/internal/service/admin_service_group_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/gateway_key_billing_test.go`
- `backend/internal/handler/request_started_at_test.go`
- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/handler/openai_videos.go`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.peakRate.spec.ts`
- `frontend/src/utils/peak-rate.ts`
- `frontend/src/i18n/locales/zh/admin/groups.ts`
- `frontend/src/i18n/locales/en/admin/groups.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/standard-group-time-rate-s211.md`
- `docs/workflow/worker-results/standard-group-time-rate-s211-result.md`
- `docs/workflow/qa-reports/standard-group-time-rate-s211-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- Every path not listed above, including Ent/schema, migrations, generated
  files, repositories, configuration, dependencies, lockfiles, Docker,
  deployment, VERSION, remote refs, and `outputs/`.
- Provider calls, shared PostgreSQL/Redis access, container changes, remote
  push, deployment, production traffic, and API-key rebinding/recreation.

## Constraints

- Preserve the existing `peak_rate_*` wire and persistence names; user-facing
  wording may say time-window/group time billing.
- Compute `final token multiplier = existing effective multiplier * time
  factor`; do not replace the group/user/membership multiplier stack.
- Treat the end boundary as outside the window. Do not add cross-midnight or
  multiple-window support.
- Use one request-start timestamp captured before forwarding. Do not calculate
  the factor from usage completion or persistence time.
- A zero `RequestStartedAt` may fall back to `timezone.Now()` only for internal
  compatibility; production call sites must populate it.
- Keep per-request image and video pricing independent from the time factor.
- Keep changes minimal and do not reformat unrelated code.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run '^(TestPeakMultiplier|TestValidatePeakRateConfig|TestNormalizePeakRateConfig|TestAdminService_.*PeakRate|TestGatewayServiceRecordUsage_.*PeakRate|TestOpenAIGatewayServiceRecordUsage_.*PeakRate)' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S211 focused service regressions failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S211 service package regression failed' }
go test ./internal/handler ./internal/handler/admin -count=1
if ($LASTEXITCODE -ne 0) { throw 'S211 handler package regression failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S211 server compile check failed' }
Pop-Location

Push-Location frontend
npm.cmd run test:run -- src/views/admin/__tests__/GroupsView.peakRate.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S211 GroupsView component tests failed' }
npm.cmd run lint:check
if ($LASTEXITCODE -ne 0) { throw 'S211 frontend lint failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S211 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S211 frontend build failed' }
Pop-Location

gofmt -d backend/internal/service/group.go backend/internal/service/group_peak_rate_test.go backend/internal/service/gateway_service.go backend/internal/service/gateway_record_usage_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_record_usage_test.go backend/internal/service/openai_videos.go backend/internal/service/admin_service.go backend/internal/service/admin_service_peak_rate_test.go backend/internal/service/admin_service_group_test.go backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/gemini_v1beta_handler.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_images.go backend/internal/handler/openai_embeddings.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/gateway_key_billing_test.go
git diff --check
git diff --name-only
git ls-files -u
git status --short
```

Browser acceptance uses a task-owned Playwright profile at desktop and 390px
width, stores evidence outside the repository under
`E:/codex-artifacts/sub2api/standard-group-time-rate-s211`, and closes only the
task-owned session/profile/daemon processes.

## Output

- Direct implementation result:
  `docs/workflow/worker-results/standard-group-time-rate-s211-result.md`.
- Independent QA report:
  `docs/workflow/qa-reports/standard-group-time-rate-s211-qa.md`.
- Workflow status/log entries for contract review, build, QA, and final verdict.

## Stop Rules

- Stop if implementation requires schema/migration, new external API fields,
  persistence redesign, routing changes, or API-key recreation.
- Stop if any production billing path still uses completion/persistence time
  after the request-start field is introduced.
- Stop if standard groups can enable a zero or negative factor, or if existing
  subscription zero-factor behavior is broken.
- Stop if per-request image/video pricing receives the time factor.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_qa_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees/sub2api/standard-group-time-rate-s211`

## Contract Review

`PASS / contract-approved`: the existing four stored/API fields, shared
`computePeakAwareMultipliers` helper, API-key group snapshot, and admin form
already provide the necessary seams. The bounded change needs no schema,
repository, routing, dependency, container, or deployment expansion. Capturing
one timestamp at each current handler entry closes the cross-boundary billing
ambiguity without changing retry/failover behavior.

## Contract Amendment

`PASS / user-authorized`: the user explicitly replaced the unavailable
`deepseek-v4-pro` QA worker with an independent `gpt-5.6-terra` Codex agent
acceptance. Product scope, allowed paths, acceptance commands, and stop rules
are unchanged; only the independent QA executor/model changed.

`PASS / review-fix`: the pre-commit multi-agent review found that the generic
Gateway still passed the peak-adjusted multiplier to non-image channel
`per_request` pricing, and that handler-local `time.Now()` calls could diverge
from the middleware-captured request start. The bounded fix must preserve the
original effective multiplier for channel `per_request`/`image` modes, make all
usage-producing handlers reuse `ctxkey.RequestStartedAt` with a compatibility
fallback, and add focused regressions. This amendment does not change storage,
API fields, routing, media pricing, or the denied deployment scope.
