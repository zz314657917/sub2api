# Task Contract: upstream-main-compat-s77

## Task ID

`upstream-main-compat-s77`

## Status

`approved`

## Role

Generator coordinator for three independent behavior-level ports from the live
upstream `main` (`09c6c6d74`): WebSocket ingress reliability, platform-aware
Grok image intent, and responsive table scrolling. Workers preserve the local
monolith and product behavior; this is not a merge of upstream history.

## Goal

Port only the approved low-risk slices from upstream commits `74e29670`,
`4f641208`, `716fcc6f`, `410ea849`, and `858f3e4a` into local
`82df0cb41`. Keep this worktree isolated and leave Ent, migrations, billing,
scheduler, payment, deployment, versioning, and knowledge files untouched.

## Success Criteria

- `gateway.openai_ws.client_first_message_timeout_seconds` exists with a
  default of `30`, rejects zero/negative values, and replaces the hard-coded
  first-client-message deadline.
- Cancelled/timed-out client WebSocket reads cannot leave blocked reader
  goroutines; inter-turn idle closure and normal downstream completion remain
  safe. Malformed upstream WS-v2 JSON is rejected before client output.
- Image intent is platform-aware: Grok passive `image_gen` namespace
  declarations do not trigger image permission, routing, or billing; explicit
  `image_generation`, an image model, or an explicit image tool choice still
  does. Existing OpenAI behavior remains unchanged across Responses, WS, and
  Chat Completions.
- Mobile/tablet `TablePageLayout` keeps `.table-wrapper`
  `overflow-x-auto`; local theme and desktop styling remain intact, with a
  component test guarding the rule.
- Only allowlisted implementation and test paths change; no migration, Ent,
  billing, payment, scheduler, deployment, VERSION, container, or
  `knowledge/**` path is touched.
- Existing per-turn client read timeouts continue to close an idle inter-turn
  read; no new inter-turn configuration field is introduced in this slice.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-main-compat-s77`
- Baseline: `82df0cb4122b0abd5b9fa3e18322d724ba47899f`
- Live upstream checked before implementation:
  `09c6c6d74050cf49ed2fb864be6c11647798ef53`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, and this
  contract.

## Worker Ownership

### WS reliability worker

Owns config, first-message deadline, client-reader cancellation/recovery,
inter-turn close behavior, malformed upstream event validation, and backend
tests. Adapt the local `openai_ws_forwarder.go` and handler flow; do not copy
upstream's split-file topology.

### Grok image-intent worker

Owns platform-aware intent classification, non-WS Responses/Chat service call
sites, and focused tests. It may modify `gateway_handler_responses.go`,
`gateway_handler_chat_completions.go`, `openai_chat_completions.go`,
`api_key_routing.go`, and `server/routes/gateway.go`. All WS files, including
`openai_gateway_handler.go` and `openai_ws_forwarder.go`, are owned by the WS
phase; after that phase is reviewed, the root integration owner applies the
Grok platform argument at those WS call sites as a separate sequential phase.
Preserve the public legacy `IsImageGenerationIntent` semantics for callers
without a platform; do not implement `2fe7df9b`'s hard rejection.

### Frontend worker

Owns the mobile `.table-wrapper` rule and
`frontend/src/components/layout/__tests__/TablePageLayout.spec.ts` only.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gateway_handler_responses_test.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/service/api_key_routing.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/gateway_test.go`
- `backend/internal/handler/openai_grok_image_intent_gate_test.go`
- `backend/internal/service/image_generation_intent.go`
- `backend/internal/service/image_generation_intent_test.go`
- `backend/internal/service/image_generation_intent_grok_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_ingress_session_test.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`
- `backend/internal/service/openai_ws_client.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter_effort_test.go`
- `frontend/src/components/layout/TablePageLayout.vue`
- `frontend/src/components/layout/__tests__/TablePageLayout.spec.ts`
- `docs/workflow/tasks/upstream-main-compat-s77.md`
- `docs/workflow/worker-results/upstream-main-compat-s77-*-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s77-qa.md`
- `docs/workflow/qa-reports/upstream-main-compat-s77-*-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Any `backend/ent/**`, `backend/migrations/**`, schema/generated code,
  Wire output, or migration numbering.
- Billing, pricing, usage settlement, scheduler policy, account schema/type
  support, prompt caching, payment, subscription, or failover policy outside
  explicit WS malformed-event behavior.
- `deploy/**`, Docker/container files, `VERSION`, release metadata,
  README, lockfiles, dependency installation, and production configuration.
- `knowledge/**` and global `C:/Users/Administrator/.codex/memories/**`.
- Any frontend path other than the two TablePageLayout paths above.
- Any upstream commit outside the five listed candidates, including
  `2fe7df9b`, `3db00d3f`, and `5e4da92d`.

## Constraints

- Work only in this isolated worktree; do not touch the dirty primary
  checkout.
- Do not cherry-pick or merge the upstream release/tag. Adapt behavior to
  local APIs and file topology.
- Keep changes minimal, preserve public contracts, and add focused regression
  tests for every new branch.
- The WS reader fix must close/cancel the underlying connection or equivalent
  read primitive so cancellation is observable; context timeout alone is not
  sufficient if a goroutine can remain blocked.
- Grok passive namespace handling must be platform-specific and must not
  suppress explicit image intent signals.
- Do not push, deploy, update containers, or modify the main worktree.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOTMPDIR = "F:/mcplugins/sub2api/.tmp/go-build"
$configTests = 'TestLoadDefaultOpenAIWSConfig|TestLoadOpenAIWSClientFirstMessageTimeoutFromEnv|TestValidateConfig_OpenAIWSRules'
go test ./internal/config -run $configTests -count=1
if ($LASTEXITCODE -ne 0) { throw "S77 config tests failed" }
$handlerTests = 'TestOpenAIResponsesWebSocket_(FirstMessageTimeoutUsesConfig|OpenAIPassiveImageNamespacePreservesLegacyPermissionGate)|TestOpenAIGatewayHandler(Responses_Grok|Responses_OpenAI|ChatCompletions_OpenAI)'
go test ./internal/handler -run $handlerTests -count=1
if ($LASTEXITCODE -ne 0) { throw "S77 handler tests failed" }
$serviceTests = 'TestResolveOpenAIWSClientFirstMessageTimeout|TestReadOpenAIWSClientMessage_.*|TestOpenAIGatewayService_(Forward_WSv2RejectsMalformedEvent.*|ProxyResponsesWebSocketFromClient_(RejectsMalformedEventBeforeOutput|InterTurnReadTimeoutClosesClient|PassthroughRejectsMalformedEvent.*))|TestIsImageGenerationIntent.*|TestIsImageGenerationIntentMap.*'
go test ./internal/service -run $serviceTests -count=1
if ($LASTEXITCODE -ne 0) { throw "S77 service tests failed" }
$routeTests = 'TestResolveAPIKeyRouteForJSONModel_GrokImageIntent.*'
go test ./internal/server/routes -run $routeTests -count=1
if ($LASTEXITCODE -ne 0) { throw "S77 route tests failed" }
Pop-Location

$frontendModules = Join-Path $PWD "frontend/node_modules"
$sharedModules = "F:/mcplugins/sub2api/frontend/node_modules"
$createdModulesJunction = $false
if (-not (Test-Path -LiteralPath $frontendModules)) {
  if (-not (Test-Path -LiteralPath $sharedModules)) { throw "Shared frontend dependencies are unavailable" }
  New-Item -ItemType Junction -Path $frontendModules -Target $sharedModules | Out-Null
  $createdModulesJunction = $true
}
try {
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/layout/__tests__/TablePageLayout.spec.ts"
  if ($LASTEXITCODE -ne 0) { throw "S77 TablePageLayout test failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
  if ($LASTEXITCODE -ne 0) { throw "S77 frontend typecheck failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
  if ($LASTEXITCODE -ne 0) { throw "S77 frontend build failed" }
} finally {
  if ($createdModulesJunction) { Remove-Item -LiteralPath $frontendModules -Force }
}

git diff --check
if ($LASTEXITCODE -ne 0) { throw "S77 diff check failed" }
```

Evaluator additionally runs exact changed-path allowlist, denied-path and
conflict-marker scans, and scenario checks for default/custom timeout,
reader exit after cancellation, inter-turn idle closure, malformed JSON
failover, Grok passive/explicit intent distinctions, and unchanged OpenAI
behavior.

The repository's authoritative `docs/workflow/agent-matrix.md` is present in
the primary checkout (the ignored workflow file is not copied into this
worktree). Its `deepseek-v4-pro` Developer Worker requirement remains binding;
if that worker model or CLI is unavailable, the worker phase is `BLOCKED` and
requires an explicit user exception rather than a silent model fallback.

## Output

- Worker reports:
  `docs/workflow/worker-results/upstream-main-compat-s77-<slice>-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s77-qa.md`.
- Main workflow log records contract review, worker completion, QA, and final
  evaluator verdict.

## Stop Rules

- Stop and return to Planner if a slice requires Ent, migration, billing,
  payment, scheduler, deployment, new account types, or a public API change.
- Stop if the WS implementation cannot prove blocked readers exit or malformed
  upstream data can still reach the client.
- Stop if Grok classification changes non-Grok behavior or treats passive
  namespace declarations as explicit image requests.
- Stop if a worker modifies a denied path, writes knowledge/memory, or expands
  frontend scope.
- Stop the worker loop after two failed attempts on the same slice and let the
  Evaluator/Codex re-split the contract.

## Budget

- worker_mode: `Codex worker fallback explicitly authorized by the user on 2026-07-16; deepseek-v4-pro unavailability recorded`
- qa_worker_mode: `Codex final evaluator with focused evidence; deepseek-v4-pro unavailability recorded`
- worktree_root: `E:/codex-worktrees`
