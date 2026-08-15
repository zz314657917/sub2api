---
task_id: upstream-v0177-remote-compaction-v2-s218
phase: contract-approved
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Remote Compaction V2 S218

## Goal

Behaviorally port upstream `9662cff2e`, `a8b9ea22b`, and `8ae6d8f67` onto the
local monolithic OpenAI gateway so native remote compaction v2 stays on the
streaming `/responses` protocol, carries the correct session beta feature, and
is used by the account compact probe. Do not merge the divergent upstream
history or copy its split-file topology.

## Success Criteria

- A bare canonical Responses request with `stream:true` and an input item whose
  `type` is `compaction_trigger` is native remote compaction v2. It remains on
  `/responses`; its stream/store/prompt-cache/reasoning/body fields are not
  normalized into the legacy unary compact schema.
- Explicit `/responses/compact` requests and non-streaming body-signal clients
  keep the existing legacy compact behavior, including compact-only model
  mapping, eligibility, JSON response handling, keepalive, and outcome logging.
- Native v2 selection requires a Responses-capable OpenAI account but does not
  require `openai_compact_supported`, does not honor `force_off` as a native-v2
  block, and does not use `compact_model_mapping`. A Chat-Completions-only API
  key must not silently drop the trigger through the raw-chat fallback.
- Channel restriction and account model resolution follow the actual forwarded
  model after channel mapping. Passthrough/raw-chat/normal forwarding retain
  their existing local model semantics; compact-only mapping applies only to
  the legacy endpoint.
- `x-codex-beta-features` is admitted by the OpenAI request allowlists. Native
  v2 always ensures `remote_compaction_v2` is present for OAuth and API-key
  accounts. Other OAuth Responses requests receive the default session-level
  value only when the client supplied no non-empty value; a client-declared
  value is preserved. HTTP native, HTTP passthrough, and WS handshake behavior
  are consistent. Existing `OpenAI-Beta` routing-hint cleanup is unchanged.
- The admin compact probe uses streaming `/responses` plus one
  `compaction_trigger`, sends `store:false` for OAuth, uses the beta feature,
  requests `text/event-stream`,
  and retains local authentication, identity convergence, proxy, TLS profile,
  429 reconciliation, and invalid-task retry behavior.
- Probe success requires an actual compaction output item from SSE done/added,
  terminal `response.output`, or whole JSON fallback. A 2xx without such an
  item persists unsupported. The probe session identifier is stable per
  account and UUID-shaped.
- All new HTTP/WS/probe tests use local fakes or loopback fixtures. No real
  OpenAI, ChatGPT, provider, production, or shared-runtime request is allowed.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen product base: `main@56d86521bc2220515361fefbcde219fe080c60e4`.
- Worker build base: the contract-approved `main` commit supplied in the
  dispatch. Before source work, the worktree must be clean and its `HEAD` must
  equal that dispatch SHA.
- Upstream: `upstream/main@baeac1f3de21d37b129405f092ef86c24b3f203d`
- Tag: `v0.1.177@073e92d17178a1ccdb0a27017f572f10c9c7ab62`
- Source commits: `9662cff2e`, `a8b9ea22b`, `8ae6d8f67`.
- Direct `git apply --check` fails because upstream split the gateway across
  files that do not exist locally. Implement against the local monolithic
  `openai_gateway_service.go` and scheduler topology.
- S217 is already integrated and unrelated user changes remain only in
  `EditAccountModal.vue`, its test, and `outputs/`.

## Allowed Paths

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_compact_body_signal_test.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/openai_gateway_compact_log_test.go`
- `backend/internal/service/openai_compact_body_signal.go`
- `backend/internal/service/openai_compact_body_signal_test.go`
- `backend/internal/service/openai_compaction_context.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_account_scheduler_compact_test.go`
- `backend/internal/service/openai_channel_restriction_test.go`
- `backend/internal/service/openai_channel_restriction_compaction_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
- `backend/internal/service/upstream_path_guard_test.go`
- `backend/internal/service/openai_oauth_passthrough_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_openai_compact_test.go`
- `backend/internal/service/openai_compact_probe.go`
- `backend/internal/service/openai_compact_probe_test.go`
- `backend/internal/service/openai_agent_identity_compat_test.go`
- `docs/workflow/worker-results/upstream-v0177-remote-compaction-v2-s218-result.md`

## Denied Paths

- `frontend/**`, especially the user-owned account modal files, and `outputs/`.
- `backend/migrations/**`, `backend/ent/**`, dependencies, generated wiring,
  configuration, deployment, containers, VERSION, and unrelated services.
- Upstream `8219dcfc8` / `4d9fedee2` turn-state relay, `fce41e318`
  fingerprint convergence, group daily rollups/migrations 222/223, Grok, and
  every other `v0.1.177` change.
- Remote push, provider calls, shared PostgreSQL/Redis operations, production
  data, release tags, wholesale upstream merge, reset/rebase, or force actions.

## Constraints

- Work only in `E:/codex-worktrees/sub2api/s218-remote-compaction-v2` after
  contract approval. Preserve local failover, billing, profit control, session
  isolation, WS pooling, identity, passthrough, and compact bridge behavior.
- Native-v2 classification requires both `stream:true` and the body trigger on
  a bare accepted Responses alias. Do not classify response subpaths or invalid
  stream values as native v2.
- Do not remove legacy compact support; only stop promoting native v2 to it.
- Header insertion must occur after safe client/header-override assembly so the
  final outgoing header follows the success criteria without broadening other
  request headers.
- Account probe fixtures must return protocol-correct local SSE/JSON. A 2xx
  response with no compaction item must be a negative result, not a success.
- New focused tests must be default-tag discoverable. `[no tests to run]` is a
  failure. Do not install dependencies or use a real provider.
- Replace the stale default-tag
  `TestRemoteCompactBodySignalMarksClientStream` legacy expectation with a
  native-v2 regression: headerless bare `/responses` plus `stream:true` and a
  `compaction_trigger` must preserve the original stream/body/path and must not
  set the legacy compact client-stream marker. This exact existing test file is
  allowlisted only for that bounded semantic correction.
- The Developer writes only the worker result. QA report, workflow status,
  main-log, current-task, and timeline remain Planner/Evaluator-owned.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/s218-remote-compaction-v2/backend

$handlerTests = @(
  'TestNormalizeOpenAIResponsesCompactRequest_RemoteV2StaysOnResponses',
  'TestOpenAIResponsesCompactionRoutingFlags',
  'TestNormalizeOpenAIResponsesCompactRequest_NonRemoteV2BodySignalPromoted',
  'TestRemoteCompactBodySignalPreservesNativeStream'
)
$serviceTests = @(
  'TestOpenAIGatewayService_SelectAccountWithScheduler_NativeCompactionIgnoresLegacyCompactProbe',
  'TestOpenAIGatewayService_SelectAccountWithScheduler_NativeCompactionRequiresResponsesCapability',
  'TestOpenAIGatewayService_SelectAccountWithScheduler_LegacyCompactionKeepsCompactEligibility',
  'TestOpenAIChannelRestriction_NativeCompactionUsesForwardModelWithoutCompactMapping',
  'TestApplyOpenAICodexBetaFeatures',
  'TestBuildOpenAIWSHeaders_CarriesSessionBetaFeatures',
  'TestCreateOpenAICompactProbePayload_NativeV2Shape',
  'TestOpenAICompactProbeFoundCompactionItem',
  'TestBuildOpenAICompactProbeExtraUpdates_2xxWithoutCompactionItemMarksUnsupported',
  'TestAccountTestService_TestAccountConnection_OpenAICompactOAuthSuccessPersistsSupport',
  'TestAccountTestService_TestAccountConnection_OpenAICompactAPIKeyUsesNativeResponsesPath'
)
foreach ($test in $handlerTests) {
  $listed = go test ./internal/handler -list "^$test$"
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "missing handler test: $test" }
}
foreach ($test in $serviceTests) {
  $listed = go test ./internal/service -list "^$test$"
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "missing service test: $test" }
}

$handlerPattern = '^(' + ($handlerTests -join '|') + ')$'
$servicePattern = '^(' + ($serviceTests -join '|') + ')$'
go test ./internal/handler -run $handlerPattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S218 handler regressions failed' }
go test ./internal/service -run $servicePattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S218 service regressions failed' }
go test ./internal/service -run '^(TestOpenAIGatewayService_SelectAccountWithScheduler_Compact|TestOpenAIGatewayService_Forward_Compact|TestOpenAIGatewayService_OAuthPassthrough_Compact|TestNormalizeOpenAICompactRequestBody|TestRemoteCompact)' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S218 legacy compact compatibility failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S218 complete service failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S218 complete handler failed' }
go test ./internal/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S218 complete server failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S218 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/s218-remote-compaction-v2
git diff --check
$buildBase = git merge-base HEAD main
if (-not $buildBase) { throw 'S218 build base cannot be resolved' }
git diff --name-only "$buildBase..HEAD"
git diff --name-only --diff-filter=U
git ls-files -u
$goPaths = @(git diff --name-only "$buildBase..HEAD" | Where-Object { $_ -like '*.go' -and (Test-Path $_) })
if ($goPaths.Count -gt 0) {
  $formatDiff = gofmt -d $goPaths
  if ($LASTEXITCODE -ne 0 -or $formatDiff) { throw 'S218 Go formatting failed' }
}
foreach ($commit in @('9662cff2e','a8b9ea22b','8ae6d8f67')) {
  git merge-base --is-ancestor $commit upstream/main
  if ($LASTEXITCODE -ne 0) { throw "missing upstream provenance: $commit" }
}
```

## Output

- Write `docs/workflow/worker-results/upstream-v0177-remote-compaction-v2-s218-result.md`
  with first line `### DONE: upstream-v0177-remote-compaction-v2-s218`,
  `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit only allowed source, tests, and the worker result in the isolated
  worktree. Include changed files, real commands, local-fixture boundary,
  risks, and contract compliance.

## Stop Rules

- Stop if the change requires turn-state/fingerprint behavior, migration,
  dependency/config changes, frontend, generated files, provider traffic, or a
  redesign of local billing/failover/WS pooling.
- Stop if native v2 can reach raw Chat Completions, legacy compact loses its
  eligibility/model mapping, client-declared beta values are overwritten, or a
  2xx probe without compaction is accepted.
- Stop after two failed implementation rounds and return to Planner. Do not
  integrate, push, deploy, delete branches/worktrees, or update containers.

## Contract Review

`PASS / contract-approved` (2026-08-16 00:18 +08:00): the local monolithic
topology is explicitly allowlisted; native v2 and legacy compact have separate
routing, capability, mapping, and error contracts; all new acceptance tests are
required to be default-tag discoverable, including a dedicated channel
restriction regression; beta-header precedence and the local-fixture probe
boundary are explicit. Developer dispatch is authorized only at the supplied
clean build-base SHA.

`PASS / Amendment 1` (2026-08-16 00:36 +08:00): complete handler regression
exposed one stale default-tag test outside the original allowlist. The existing
test required a headerless streaming compaction trigger to be promoted to the
legacy unary bridge, directly contradicting the approved upstream v2 behavior.
Allow only `openai_gateway_handler_test.go` to rename and correct that one
regression; product semantics, all other allowed/denied paths, and acceptance
gates remain unchanged.
