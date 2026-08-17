# Task Contract: upstream-cn-providers-s226

## Task ID

`upstream-cn-providers-s226`

## Status

`contract-approved`

## Role

Planner, Terra Developer Worker, independent Terra QA Worker, and Final
Evaluator. Implementation is split into ordered batches `S226-A` through
`S226-D`; `S226-E` is integration-only QA. No worker may start until the user
explicitly authorizes the next batch and the Controller records its exact base.

## Goal

Behaviorally port the locally reachable parts of upstream
`901a0439f1575a45c29150e04d2ccc3ed87f4948`, including the security and stream
correctness fixes in `4b667ccd45747162ac58545cdbb6f6d88737bf04` and the
format-only follow-up `e7285453829930e2889432dcd18d7a0c2ba18481`, so Kimi,
Zhipu, and DeepSeek are first-class API-key account platforms with protocol-aware
routing, quota/balance visibility, and bounded failure handling.

Adapt behavior to the local monolithic gateway topology. Do not merge or
cherry-pick the upstream chain wholesale.

## Batch Contract

| Batch | Required commit boundary | Scope | Gate before next batch |
| --- | --- | --- | --- |
| `S226-A` | one implementation commit plus one worker report | platform/account foundation | focused tests x10, full service, server compile, exact allowlist |
| `S226-B` | one implementation commit plus one worker report | quota/balance probes and admin API | B4 zero-egress security tests, focused tests x10, full affected packages, Wire compile |
| `S226-C` | one implementation commit plus one worker report | multi-protocol gateway, stream pump, reactive cooldown/failover | B3 timeout tests, protocol matrix, focused tests x10, full service/handler/server compile |
| `S226-D` | one implementation commit plus one worker report | frontend account controls and status cells | focused Vitest, typecheck, build, protected modal patch collision probe |
| `S226-E` | QA evidence only; no product code | cross-batch integration and independent QA | full backend/frontend regression, UI screenshots, provenance/scope/dirty-worktree gates |

Each implementation batch starts from the previous batch's Controller-approved
commit. A later batch must not repair an earlier batch silently; a failed gate
returns to the owner of the failing batch.

## Upstream Review Blockers

| Item | S226 disposition |
| --- | --- |
| B1: accidental root `docker-compose.yml` | Excluded. Do not create, restore, or modify it. |
| B2: migration 224 and `user_platform_quotas` expansion | Local N/A. The local checkout has no table, Ent schema, repository, service, or UI. Do not renumber it to 226 and do not import prerequisite `6b39b344d`/`f7f5e3383`. |
| B3: four Anthropic-native read loops can hang forever | Required in `S226-C`. Use `gateway.stream_data_interval_timeout`, close the response body on idle, and retain accumulated usage semantics. |
| B4: quota/balance probes bypass URL policy | Required in `S226-B`. Validate the final target before constructing/sending the credentialed request; rejection must produce zero outbound requests. |

## Local Topology Decisions

- The local checkout does not contain upstream split owners
  `gateway_anthropic_passthrough.go`, `gateway_upstream_response.go`,
  `openai_gateway_cc_pipeline.go`, `openai_gateway_forward.go`,
  `openai_gateway_passthrough.go`, `openai_gateway_request_body.go`,
  `openai_gateway_scheduling.go`, or `openai_ws_forwarder_payload.go`.
  Their target deltas must be adapted into the allowlisted local owners
  `gateway_service.go`, `openai_gateway_service.go`,
  `openai_gateway_chat_completions_raw.go`, and `openai_ws_forwarder.go`.
- Upstream configurable account scheduling thresholds depend on
  `7c62382d0`, a locally absent 55-file/3542-line product slice. S226 does not
  import that subsystem and does not modify frontend settings for it. Coding
  Plan quota snapshots remain available to the UI and to the reactive 429
  reset-point cooldown; configurable pre-exhaustion thresholds are a separate
  follow-up candidate.
- Upstream `cn_providers_test.go` mixes foundation, probe, and gateway concerns.
  Split its locally relevant cases into the batch-specific test files below.
- The local account-modal patch is user-owned. It is a required compatibility
  baseline, not an implementation change to absorb or discard.

## Success Criteria

### S226-A: platform and account foundation

- Add exact platform constants for `kimi`, `zhipu`, and `deepseek` and exact
  credential values for `account_mode=payg|coding` and
  `api_protocol=chat_completions|anthropic|responses`.
- Preserve `chat_completions` as the compatibility default; only DeepSeek may
  select `responses`.
- Resolve official default Base URLs by provider, mode, and protocol while
  preserving a user-supplied `base_url`.
- Provide protocol-aware API-key access, Anthropic auth scheme selection, and
  model-list request construction without enabling gateway routes yet.

### S226-B: quota/balance probes and admin API

- Query Kimi/Zhipu Coding Plan 5h/weekly windows and persist provider-prefixed
  snapshots without erasing a prior snapshot on probe failure.
- Query Kimi and DeepSeek payg balances, including DeepSeek CNY/USD details;
  any currency at or above the configured threshold keeps the account usable.
- Periodically probe eligible accounts with bounded concurrency and workload-
  scaled timeout. Balance recovery may clear only a temporary pause written by
  this CN balance subsystem.
- Expose admin quota/balance reads through dedicated routes and generated Wire
  wiring.
- Apply the same `security.url_allowlist` policy used by gateway forwarding to
  the final quota/balance URL before the API key can leave the process. A
  rejected URL performs zero network I/O.

### S226-C: multi-protocol gateway and failure handling

- Route Kimi/Zhipu/DeepSeek groups only to the exact same platform; do not
  collapse them into OpenAI or Grok pools.
- Support Chat Completions, native Anthropic messages/count-tokens, and
  DeepSeek Responses according to the account protocol. Same-protocol requests
  remain native; cross-protocol requests use existing `apicompat` conversions.
- Normalize DeepSeek Responses to `/responses`, `store=false`, and no
  `previous_response_id`; preserve all non-DeepSeek behavior.
- Preserve usage collection after client disconnect where required, while the
  B3 interval pump bounds all four Anthropic-native read loops and closes the
  upstream body on idle.
- Treat balance insufficiency as recoverable temporary unschedulability, and
  Coding Plan 429 as a cooldown to the earliest valid future quota reset. Do
  not turn these cases into permanent account errors.
- Keep existing OpenAI, Grok, Anthropic, Bedrock, WebSocket, fingerprint,
  billing, and failover behavior intact.

### S226-D: frontend account management and status

- Create/edit flows expose Kimi, Zhipu, and DeepSeek API-key accounts with
  mode/protocol controls, legal combinations, Base URL presets, and watcher
  race protection.
- Credentials persist `api_key`, `base_url`, `account_mode`, and `api_protocol`
  consistently in create and edit flows.
- Account rows display Coding Plan windows for Kimi/Zhipu and balances for
  Kimi/DeepSeek, retain the last snapshot on probe failure, and debounce
  automatic probes for five minutes.
- Platform icon, badge, colors, model whitelist behavior, types, and English/
  Chinese strings are complete.
- Preserve the user modal layout patch exactly as a separate working-tree
  delta. Before implementation, create a task-local baseline commit containing
  that patch only; the baseline commit must never be integrated. The S226-D
  business commit is the delta after that baseline.

### S226-E: integration and independent QA

- Integrate only the approved A-D business commits and evidence commits.
- Run the full affected backend/frontend gates, then inspect real account forms
  and usage cells at desktop and mobile sizes with a task-owned browser profile.
- Prove exact upstream provenance, allowlist compliance, no migration/dependency
  expansion, no unresolved index entries, and preservation of both protected
  patch IDs.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen S226-A implementation base:
  `98daf5b8d9008c9db6753631a62ede9a3ff8ca6d`
- Local branch at contract review: `main`
- Origin at contract review: `a865d8b6eb06048f7cf7e3b983b65cf393197806`
- Upstream at contract review:
  `e330c243a8f142f8963d784916da0093ab7084ee`
- S226-A dispatch worktree:
  `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-a`, branch
  `pge/upstream-cn-providers-s226-a`, exact clean HEAD
  `98daf5b8d9008c9db6753631a62ede9a3ff8ca6d`.
- S226-B dispatch worktree:
  `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-b`, branch
  `pge/upstream-cn-providers-s226-b`, exact clean HEAD and approved base
  `3ed89c9952f09e03861e197d59aad456f3b19b29`.
- Controller main-worktree protection snapshot at S226-A dispatch: account
  modal `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`, tutorial view
  `9e0894bc8af07e9d358f06d367dc976cf3bb3f65`, knowledge
  `2abee47db90ce1d54e1f9ba7d1a3cc2d633c2374`, backend tutorial tests
  `a81fbffbe14121ef62387f28cfee09a6d247ac94`; untracked migrations 226/227
  and `outputs/` remain excluded.
- Direct apply checks fail across Wire, config, gateway, missing quota schema,
  and split-file topology. Manual behavioral adaptation is mandatory.
- Read first: this contract, `docs/workflow/spec.md`,
  `docs/workflow/status.md`, and `knowledge/tasks/current-task.md` from the main
  checkout.

## Allowed Paths

### S226-A business and tests

- `backend/internal/domain/constants.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_service.go`
- `backend/internal/service/anthropic_apikey_auth.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/cn_provider_foundation_test.go`
- `docs/workflow/worker-results/upstream-cn-providers-s226-a-result.md`

### S226-B business and tests

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/admin/cn_provider_handler.go`
- `backend/internal/handler/handler.go`
- `backend/internal/handler/wire.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/admin_cn_provider_test.go`
- `backend/internal/service/cn_provider_balance_check_service.go`
- `backend/internal/service/cn_provider_balance_check_service_test.go`
- `backend/internal/service/cn_provider_balance_service.go`
- `backend/internal/service/cn_provider_probe_url.go`
- `backend/internal/service/cn_provider_probe_url_test.go`
- `backend/internal/service/cn_provider_quota_balance_test.go`
- `backend/internal/service/cn_provider_quota_service.go`
- `backend/internal/service/upstream_billing_probe.go`
- `backend/internal/service/upstream_billing_probe_multiplatform_test.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`
- `docs/workflow/worker-results/upstream-cn-providers-s226-b-result.md`

### S226-C business and tests

- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_apikey_responses_probe.go`
- `backend/internal/service/openai_embeddings.go`
- `backend/internal/service/openai_gateway_anthropic_native_pump.go`
- `backend/internal/service/openai_gateway_anthropic_native_pump_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_count_tokens.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_anthropic_native.go`
- `backend/internal/service/openai_gateway_model_availability.go`
- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/ratelimit_cn_providers.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/cn_provider_gateway_test.go`
- `docs/workflow/worker-results/upstream-cn-providers-s226-c-result.md`

### S226-D business and tests

- `frontend/src/api/admin/cnProviders.ts`
- `frontend/src/api/admin/index.ts`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/account/CNProviderBalanceCell.vue`
- `frontend/src/components/account/CNProviderQuotaCell.vue`
- `frontend/src/components/account/CnBaseUrlPresets.vue`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/ModelWhitelistSelector.vue`
- `frontend/src/components/account/credentialsBuilder.ts`
- `frontend/src/components/common/PlatformIcon.vue`
- `frontend/src/components/common/PlatformTypeBadge.vue`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/types/index.ts`
- `frontend/src/utils/platformColors.ts`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/components/account/__tests__/CNProviderBalanceCell.spec.ts`
- `frontend/src/components/account/__tests__/CNProviderQuotaCell.spec.ts`
- `frontend/src/components/account/__tests__/CreateAccountModal.spec.ts`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts`
- `frontend/src/components/account/__tests__/credentialsBuilder.spec.ts`
- `docs/workflow/worker-results/upstream-cn-providers-s226-d-result.md`

### S226-E evidence only

- `docs/workflow/qa-reports/upstream-cn-providers-s226-qa.md`

Controller-owned workflow files may be updated only at gates:

- `docs/workflow/main-log.md`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-cn-providers-s226.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Every path outside the active batch allowlist.
- All database and quota-product paths, especially
  `backend/ent/schema/user_platform_quota.go`,
  `backend/internal/repository/user_platform_quota_repo_integration_test.go`,
  `backend/internal/handler/admin/user_platform_quota_admin_test.go`,
  `backend/internal/server/api_contract_test.go`,
  `backend/migrations/224_user_platform_quotas_add_cn_providers.sql`, and
  `backend/migrations/user_platform_quota_cn_providers_migration_test.go`.
- Generic scheduling-threshold prerequisites and UI, including
  `backend/internal/service/account_scheduling_threshold_eval.go` and
  `frontend/src/api/admin/settings.ts`.
- The absent upstream split-file paths listed under Local Topology Decisions;
  do not create them merely to mirror upstream layout.
- `docker-compose.yml`, Docker/deployment files, VERSION, dependencies,
  manifests, lockfiles, generated Ent code, and unrelated generated code.
- User-owned dirty files outside the D compatibility scope:
  `frontend/src/views/public/TutorialView.vue`, its test,
  `knowledge/00-start-here.md`, `knowledge/05-current-focus.md`, and
  `outputs/**`.
- Real provider calls, real credentials, shared/production database access,
  container changes, deployment, remote push, release tagging, wholesale merge,
  or direct cherry-pick.

## Constraints

- Preserve all local S220-S225 behavior, account sharing, pricing, billing,
  fingerprint, long-context, daily rollup, image, WebSocket, and Grok changes.
- Use the existing `apicompat`, URL validator, HTTP upstream, proxy, repository,
  scheduler, and rate-limit abstractions. No new dependency is authorized.
- Do not log API keys, Authorization values, raw credentials, or unbounded
  upstream bodies.
- URL validation must occur after final endpoint construction and before
  request creation/dispatch. Tests must use fake/httptest transports only.
- B3 interval timeout `0` retains the existing disabled semantics; nonzero uses
  the configured duration. Timeout paths close `resp.Body` and cannot leak a
  reader goroutine indefinitely.
- Keep old receiver helpers callable where local tests depend on them; add a
  package-level shared helper or wrapper instead of mechanically rewriting
  unrelated Bedrock/Anthropic tests.
- Generated Wire output must match handwritten provider changes and must not
  include unrelated generation churn.
- Frontend dependencies may be installed only inside the task worktree without
  changing package manifests or lockfiles. Missing `vitest`, `vue-tsc`, or Vite
  is `BLOCKED`, not a skipped PASS.
- Browser QA must use a task-owned profile/session and exact PID ownership.
  Close the session and prove the task profile and Playwright daemon are gone.
- Protected main-worktree patch IDs must remain:
  - account modal: `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`
  - tutorial view: `9e0894bc8af07e9d358f06d367dc976cf3bb3f65`
  - backend tutorial tests: `a81fbffbe14121ef62387f28cfee09a6d247ac94`
  - knowledge files: `2abee47db90ce1d54e1f9ba7d1a3cc2d633c2374`

## Acceptance Commands

The Controller records the exact approved base for each batch before dispatch.
Every listed focused test must first be discoverable with `go test -list` or
Vitest collection; undiscoverable tests are a failure.

### S226-A

```powershell
Push-Location backend
$tests = @(
  'TestCNProviderAccountDefaultsAndModes',
  'TestGetOpenAIProtocolAPIKey_CNProviders',
  'TestGetAPIProtocol',
  'TestAnthropicProtocolBaseURL',
  'TestGetOpenAIFormatBaseURL_ProtocolAware',
  'TestBuildUpstreamModelsRequest_CNProviders',
  'TestBuildUpstreamModelsRequest_AnthropicProtocol',
  'TestGetAnthropicAPIKeyAuthScheme_CNProvider'
)
foreach ($test in $tests) {
  $listed = go test -list "^$test$" ./internal/service
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "S226-A undiscoverable: $test" }
}
$pattern = '^(' + ($tests -join '|') + ')$'
go test ./internal/service -run $pattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S226-A focused tests failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S226-A service regression failed' }
go test ./cmd/server -run '^$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S226-A server compile failed' }
Pop-Location
```

### S226-B

```powershell
Push-Location backend
$tests = @(
  'TestCNProviderBalanceCheckRunOnceProbesCodingPlanQuota',
  'TestCNProviderBalanceCheckRunOnceWithoutQuotaService',
  'TestAllCNBalancesBelowThreshold',
  'TestCNValidateProbeURL_AllowlistPolicy',
  'TestCNProviderQuotaService_RejectsURLBlockedByPolicy',
  'TestCNProviderBalanceService_RejectsURLBlockedByPolicy',
  'TestCNProviderBalanceService_OfficialHostPassesValidation',
  'TestCNProviderBalanceService_DeepSeekKeepsAccountWhenAnyCurrencyHealthy',
  'TestCNProviderQuotaService_ProbeFailurePreservesSnapshot',
  'TestCNProviderBalanceCheckClearsOnlyOwnedPause',
  'TestParseKimiUsageTiers',
  'TestParseZhipuTokenTiers_UnitClassification',
  'TestParseZhipuTokenTiers_IgnoresNonTokenEntries',
  'TestCNQuotaExtraUpdates',
  'TestCNBalanceURL',
  'TestKimiQuotaURL',
  'TestZhipuQuotaHost'
)
foreach ($test in $tests) {
  $listed = go test -list "^$test$" ./internal/service
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "S226-B undiscoverable: $test" }
}
$pattern = '^(' + ($tests -join '|') + ')$'
go test ./internal/service -run $pattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S226-B focused tests failed' }
go test ./internal/config ./internal/service ./internal/server/routes ./cmd/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S226-B affected regression failed' }
Pop-Location
```

### S226-C

```powershell
Push-Location backend
$tests = @(
  'TestAnthropicNativeLinePump_TimesOutWithoutData',
  'TestAnthropicNativeLinePump_DataResetsTimer',
  'TestCCStreamingFromNativeAnthropic_HangTimesOut',
  'TestCCBufferedFromNativeAnthropic_HangTimesOut',
  'TestResponsesStreamingFromNativeAnthropic_HangTimesOut',
  'TestResponsesBufferedFromNativeAnthropic_HangTimesOut',
  'TestCCStreamingFromNativeAnthropic_HappyPathStillConverts',
  'TestCCBufferedFromNativeAnthropic_HappyPathStillConverts',
  'TestCNProviderGatewayPlatformIsolation',
  'TestCNProviderAnthropicNativeMessagesPassthrough',
  'TestCNProviderChatCompletionsAnthropicConversion',
  'TestCNProviderResponsesAnthropicConversion',
  'TestDeepSeekResponsesURLAndBodyNormalization',
  'TestCNProviderCountTokensRouting',
  'TestCNProviderReactive429AndBalanceRecovery',
  'TestCNProviderProtocolProbeMode'
)
foreach ($test in $tests) {
  $listed = go test -list "^$test$" ./internal/service
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "S226-C undiscoverable: $test" }
}
$pattern = '^(' + ($tests -join '|') + ')$'
go test ./internal/service -run $pattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S226-C focused tests failed' }
go test ./internal/service ./internal/handler ./internal/server/routes -count=1
if ($LASTEXITCODE -ne 0) { throw 'S226-C affected regression failed' }
go test ./cmd/server -run '^$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S226-C server compile failed' }
Pop-Location
```

### S226-D

```powershell
Push-Location frontend
$focused = @(
  'src/components/account/__tests__/AccountUsageCell.spec.ts',
  'src/components/account/__tests__/CNProviderBalanceCell.spec.ts',
  'src/components/account/__tests__/CNProviderQuotaCell.spec.ts',
  'src/components/account/__tests__/CreateAccountModal.spec.ts',
  'src/components/account/__tests__/EditAccountModal.spec.ts',
  'src/components/account/__tests__/ModelWhitelistSelector.spec.ts',
  'src/components/account/__tests__/credentialsBuilder.spec.ts'
)
npm.cmd run test:run -- $focused
if ($LASTEXITCODE -ne 0) { throw 'S226-D focused Vitest failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S226-D typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S226-D build failed' }
Pop-Location
```

### Cross-cutting scope, provenance, and dirty-worktree gates

Workers run the base, diff, conflict, index, and provenance checks in their
clean task worktree. The protected dirty-worktree checks are Controller-only
and run in the main checkout; workers must not recreate, stage, or commit user
patches merely to satisfy those checks.

```powershell
$base = $env:S226_BATCH_BASE
if (-not $base) { throw 'S226_BATCH_BASE is required' }
git rev-parse --verify "$base^{commit}"
if ($LASTEXITCODE -ne 0) { throw 'S226 batch base is invalid' }
git diff --check "$base...HEAD"
if ($LASTEXITCODE -ne 0) { throw 'S226 diff check failed' }
if (git diff --name-only --diff-filter=U) { throw 'S226 has conflicts' }
if (git ls-files -u) { throw 'S226 index has unresolved entries' }

git merge-base --is-ancestor 901a0439f1575a45c29150e04d2ccc3ed87f4948 upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S226 feature provenance failed' }
git merge-base --is-ancestor 4b667ccd45747162ac58545cdbb6f6d88737bf04 upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S226 review-fix provenance failed' }
git merge-base --is-ancestor e7285453829930e2889432dcd18d7a0c2ba18481 upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S226 format provenance failed' }

$accountPatch = git diff -- frontend/src/components/account/EditAccountModal.vue frontend/src/components/account/__tests__/EditAccountModal.spec.ts
$accountID = ($accountPatch | git patch-id --stable).Split(' ')[0]
if ($accountID -ne '5d316e5b6935fdc5dbf825f940feaf231d79ac0f') { throw 'account modal patch changed' }
$tutorialPatch = git diff -- frontend/src/views/public/TutorialView.vue frontend/src/views/public/__tests__/TutorialView.spec.ts
$tutorialID = ($tutorialPatch | git patch-id --stable).Split(' ')[0]
if ($tutorialID -ne '9e0894bc8af07e9d358f06d367dc976cf3bb3f65') { throw 'tutorial view patch changed' }
$knowledgePatch = git diff -- knowledge/00-start-here.md knowledge/05-current-focus.md
$knowledgeID = ($knowledgePatch | git patch-id --stable).Split(' ')[0]
if ($knowledgeID -ne '2abee47db90ce1d54e1f9ba7d1a3cc2d633c2374') { throw 'knowledge patch changed' }
$outputState = git status --porcelain=v1 --untracked-files=all -- outputs/
if (-not ($outputState | Select-String '^\?\?')) { throw 'outputs state changed unexpectedly' }
```

S226-E additionally runs all A-D focused gates, complete backend service/
handler/server packages, frontend focused tests/typecheck/build, and real UI
inspection. No real provider or shared database is part of acceptance.

## Output

- Developer reports:
  - `docs/workflow/worker-results/upstream-cn-providers-s226-a-result.md`
  - `docs/workflow/worker-results/upstream-cn-providers-s226-b-result.md`
  - `docs/workflow/worker-results/upstream-cn-providers-s226-c-result.md`
  - `docs/workflow/worker-results/upstream-cn-providers-s226-d-result.md`
- Independent QA report:
  `docs/workflow/qa-reports/upstream-cn-providers-s226-qa.md`.
- Each Developer report first line is `### DONE`, `### BLOCKED`, or
  `### FAILED` with its batch ID. The QA report first line is
  `### PASS: upstream-cn-providers-s226`, `### FAIL: ...`, or `### BLOCKED: ...`.
- Reports list changed files, exact base/head, commands, outputs, risks,
  contract compliance, and `knowledge_candidates` without long unrelated logs.
- No push or deployment. Main integration occurs only after S226-E PASS.

## Stop Rules

- Stop if a batch needs a denied path, migration, new dependency, generic
  scheduling-threshold prerequisite, production/shared resource, or real API
  credential.
- Stop if B4 validation happens after request dispatch, any rejected URL emits
  network traffic, or an API key can reach a policy-denied host.
- Stop if any B3 native read loop can block beyond the configured interval, if
  timeout fails to close the body, or if accumulated usage semantics regress.
- Stop if Kimi/Zhipu/DeepSeek platform selection can cross-match another
  provider, or if DeepSeek-only Responses becomes selectable for Kimi/Zhipu.
- Stop if a user patch baseline is committed as product code, either protected
  patch ID changes, or `outputs/` is added/removed/modified.
- Stop if focused tests are undiscoverable, frontend tools are unavailable,
  browser ownership/cleanup cannot be proven, Terra is unavailable, or the
  active batch base is not the exact prior approved commit.

## Budget

- developer_worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees/sub2api`
- worktree sequence: one clean implementation worktree per active batch; do not
  touch `E:/codex-worktrees/sub2api/tutorial-nav-20260817`

## Contract Review

`PASS / contract-approved`: the upstream chain is decomposed into four
independently compilable implementation boundaries and one integration QA gate;
B1/B2 are explicitly excluded, B3/B4 are mandatory in their owning batches,
the absent quota and scheduling-threshold products cannot expand scope silently,
and every upstream split-file delta has a named local owner. The contract
protects all three user patch IDs, `outputs/`, remote/deployment boundaries, and the
unrelated detached worktree. This verdict approves the contract only; no S226
business behavior has been implemented or runtime-verified.
