---
task_id: upstream-v0177-turn-state-s219
phase: contract-draft
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Codex Turn-State S219

## Goal

Behaviorally port the HTTP `x-codex-turn-state` relay and cross-account echo
protection from upstream `8219dcfc8` plus test fix `4d9fedee2` into the local
monolithic OpenAI gateway. Include only the two turn-state request-guard hook
ideas from `fce41e318`; do not port its fingerprint convergence behavior,
frontend defaults, or unrelated files.

## Success Criteria

- Native OpenAI HTTP streaming, non-streaming JSON, and SSE-to-JSON responses
  explicitly relay `x-codex-turn-state` without widening the generic response
  header allowlist. The passthrough equivalents do the same.
- An upstream response without turn-state clears a value staged by an earlier
  failover attempt. A discarded pre-output attempt must not become the recorded
  provenance for the client session.
- Provenance is keyed by a positive downstream API-key ID plus the original
  client session header. `session-id` has priority over `session_id`; a missing
  or non-positive API-key ID, missing session, nil context, nil account, or
  non-positive account ID is not tracked.
- Streaming provenance is recorded only after the first successful downstream
  flush that commits response headers. In this checkout a keepalive flush also
  commits headers and therefore counts; a pre-output failover with no successful
  flush does not count.
- Provenance expires using the existing OpenAI WS session sticky TTL and is
  opportunistically swept. Expired or malformed entries are removed and do not
  cause a client header to be stripped.
- Both normal and passthrough HTTP request builders strip a client-echoed
  turn-state only when the same API-key/session seed is known to have received
  it from a different OpenAI account. Same-account, unknown, expired, untracked,
  and empty echoes remain unchanged.
- The guard runs after client header whitelisting and before account header
  overrides. It is stripping-only: it never injects turn-state and does not
  change the existing Claude compatibility bridge, HTTP/WS session isolation,
  WS handshake replay, routing hints, beta features, billing, failover, or
  account selection.
- All tests are default-tag discoverable and use only in-process fakes,
  `httptest`, or loopback fixtures. No provider or shared runtime is used.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen product base: `main@d8940bff5caf42c020a548d46bf8b4926400ba54`.
- Worker build base: the contract-approved `main` commit supplied at dispatch;
  the clean worktree `HEAD` must equal that SHA before source work.
- Upstream: `upstream/main@baeac1f3de21d37b129405f092ef86c24b3f203d`.
- Tag: `v0.1.177@073e92d17178a1ccdb0a27017f572f10c9c7ab62`.
- Primary source commits: `8219dcfc8`, `4d9fedee2`.
- Narrow dependency source: only the normal/passthrough turn-state guard call
  placement from `fce41e318`; all fingerprint code and frontend changes remain
  denied.
- Direct apply fails because this checkout has no upstream split
  `openai_gateway_response_handling.go`, `openai_gateway_forward.go`, or
  `openai_gateway_passthrough.go`; the corresponding logic remains in
  `openai_gateway_service.go`.
- The local checkout also lacks upstream `openai_codex_fingerprint.go`, so the
  original client session extraction needed for the provenance seed must live
  in the bounded turn-state helper without enabling fingerprint convergence.

## Allowed Paths

- `backend/internal/service/openai_codex_turn_state.go`
- `backend/internal/service/openai_codex_turn_state_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_tool_namespace_normalization_s92_test.go`
- `docs/workflow/worker-results/upstream-v0177-turn-state-s219-result.md`

## Denied Paths

- `frontend/**`, especially the user-owned account modal files, and `outputs/`.
- `backend/internal/service/openai_codex_fingerprint*`, fingerprint mode/default
  behavior, client metadata convergence, and every frontend hunk from
  `fce41e318`.
- `backend/migrations/**`, `backend/ent/**`, dependencies, generated wiring,
  configuration, deployment, containers, VERSION, and unrelated services.
- Group daily rollups/migrations 222/223, provider calls, shared PostgreSQL or
  Redis, production data, release tags, wholesale upstream merge, remote push,
  reset/rebase, force actions, and every other `v0.1.177` change.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, QA reports,
  `knowledge/**`, and global memories; these remain Planner/Evaluator-owned.

## Constraints

- Work only in `E:/codex-worktrees/sub2api/s219-turn-state` after contract
  approval. Preserve the monolithic gateway topology and keep edits minimal.
- Do not add turn-state to the generic response header filter. Relay and clear
  it explicitly at the existing OpenAI response commit points only.
- Do not record provenance merely because an upstream 2xx header exists. For
  streaming paths, record only after the downstream writer successfully flushes
  and commits that header; ensure the note happens at most once per response.
- When a new attempt has no turn-state, delete any stale staged header before
  output. A pre-output failover must leave no provenance from the abandoned
  account.
- The provenance seed must use the original inbound client headers, before the
  existing API-key session isolation rewrites outbound values. It must require
  both `apiKeyID > 0` and a non-empty original session ID.
- Run the outbound guard after request-header whitelisting but before account
  header overrides. Admin-configured account overrides are not client echoes and
  must retain their existing final precedence.
- Keep the existing Claude compatibility bridge and WS state store independent;
  do not reuse either as the native HTTP provenance map and do not inject a
  stored value into native Codex HTTP requests.
- Signature changes required only to carry `account` into existing response
  helpers may update the allowlisted call sites/tests; no unrelated refactor.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/s219-turn-state/backend

$serviceTests = @(
  'TestOpenAICodexTurnStateSeedRequiresAPIKeyAndSession',
  'TestOpenAICodexTurnStateRelayGuardAndExpiry',
  'TestOpenAIHTTPBuildersGuardCrossAccountTurnState',
  'TestOpenAIStreamingTurnStateRecordsOnlyOnCommit',
  'TestOpenAINonStreamingTurnStateRelaysJSONAndSSE',
  'TestOpenAIPassthroughTurnStateRelayAndGuard',
  'TestWriteOpenAIPassthroughResponseHeadersTurnState'
)
foreach ($test in $serviceTests) {
  $listed = go test ./internal/service -list "^$test$"
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) {
    throw "missing default-tag service test: $test"
  }
}
$servicePattern = '^(' + ($serviceTests -join '|') + ')$'
go test ./internal/service -run $servicePattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S219 focused regressions failed' }

$compatPattern = '^(' + (@(
  'TestOpenAIStreamingReadErrorBeforeOutputReturnsFailover',
  'TestOpenAIStreamingPreambleOnlyMissingTerminalReturnsFailover',
  'TestOpenAIStreamingPassthroughResponseFailedBeforeOutputReturnsFailover',
  'TestForwardAsAnthropic_ReusesOAuthCodexTurnState',
  'TestForwardAsAnthropic_OAuthDigestFallbackReusesTurnStateWithoutExplicitKey',
  'TestOpenAIGatewayService_Forward_WSv2_TurnStateAndMetadataReplayOnReconnect',
  'TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughHeadersUsePromptCacheAndTurnState'
) -join '|') + ')$'
go test ./internal/service -run $compatPattern -count=1
if ($LASTEXITCODE -ne 0) { throw 'S219 failover/Claude/WS compatibility failed' }

go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S219 complete service failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S219 complete handler failed' }
go test ./internal/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S219 complete server failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S219 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/s219-turn-state
git diff --check
$buildBase = git merge-base HEAD main
if (-not $buildBase) { throw 'S219 build base cannot be resolved' }
$changed = @(git diff --name-only "$buildBase..HEAD")
$allowed = @(
  'backend/internal/service/openai_codex_turn_state.go',
  'backend/internal/service/openai_codex_turn_state_test.go',
  'backend/internal/service/openai_gateway_service.go',
  'backend/internal/service/openai_gateway_service_test.go',
  'backend/internal/service/openai_tool_namespace_normalization_s92_test.go',
  'docs/workflow/worker-results/upstream-v0177-turn-state-s219-result.md'
)
$outside = @($changed | Where-Object { $_ -notin $allowed })
if ($outside.Count -ne 0) { throw "S219 outside allowlist: $($outside -join ', ')" }
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S219 conflict or unmerged index found'
}
$goPaths = @($changed | Where-Object { $_ -like '*.go' -and (Test-Path $_) })
if ($goPaths.Count -gt 0) {
  $formatDiff = gofmt -d $goPaths
  if ($LASTEXITCODE -ne 0 -or $formatDiff) { throw 'S219 Go formatting failed' }
}
foreach ($commit in @('8219dcfc8','4d9fedee2','fce41e318')) {
  git merge-base --is-ancestor $commit upstream/main
  if ($LASTEXITCODE -ne 0) { throw "missing upstream provenance: $commit" }
}
```

## Output

- Write only
  `docs/workflow/worker-results/upstream-v0177-turn-state-s219-result.md` with
  first line `### DONE: upstream-v0177-turn-state-s219`, `### BLOCKED: ...`, or
  `### FAILED: ...`.
- Commit only allowed source/tests and the worker result in the isolated
  worktree. Include changed files, real commands, commit-boundary evidence,
  local-fixture boundary, risks, and contract compliance.

## Stop Rules

- Stop if implementation requires fingerprint mode/default changes, frontend,
  schema/migration, dependency/config, WS protocol redesign, Claude bridge
  redesign, provider access, or a new shared persistence layer.
- Stop if turn-state is injected into native HTTP, same-account/unknown echoes
  are stripped, an abandoned attempt records provenance, stale response headers
  can survive failover, or the generic response filter must be widened.
- Stop after two failed implementation rounds and return to Planner. Do not
  integrate, push, deploy, update containers, or clean branches/worktrees.

## Contract Review

`PASS / contract-approved` (2026-08-16 01:53 +08:00): the contract matches the
local monolithic topology and closes the upstream dependency gap without
importing fingerprint behavior. The review specifically requires a positive
API-key/session seed, first successful downstream flush as the streaming
provenance commit point, stale-header clearing across abandoned attempts, both
normal and passthrough builder guards, and independent Claude/WS compatibility
tests. All seven referenced existing compatibility tests are default-tag
discoverable, and `8219dcfc8`, `4d9fedee2`, and `fce41e318` are verified
ancestors of `upstream/main`. Source work is authorized only in the supplied
clean S219 worktree at the contract-approval SHA and only within Allowed Paths.
