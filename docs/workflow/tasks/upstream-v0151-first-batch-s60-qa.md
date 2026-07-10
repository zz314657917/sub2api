# Task Contract: upstream-v0151-first-batch-s60-qa

## Task ID

`upstream-v0151-first-batch-s60-qa`

## Status

`approved`

## Role

You are the independent P/G/E QA worker. Validate the combined business branch only; do not implement fixes or modify tracked business files.

## Phase And Owners

- Current phase: `build`; this contract is the gate into `qa`.
- Planner and final evaluator: root Codex.
- Generator owners: S60a and S60b Developer Workers, already completed and reviewed.
- QA executor: independent Hermes tester using the approved per-Sprint `gpt-5.4` fallback.

## Review Gate

- Contract meta-review verdict is `### APPROVED: upstream-v0151-first-batch-s60-qa` or `### BLOCKED: upstream-v0151-first-batch-s60-qa`.
- After approval, QA execution uses the separate report verdict protocol under Output: `PASS`, `FAIL`, or `BLOCKED`.

## Goal

Verify that business branch `codex/upstream-v0151-first-batch-s60-business` at `a07c3e669` cleanly combines S60a commit `7f47afa11` and S60b without workflow, knowledge, billing, migration, deployment, or unrelated frontend changes.

## Success Criteria

- The branch starts from fixed baseline `3332c6883e7480f030fcffbccb6dc7ee0a3f69ca` and contains only the two expected business commits.
- S60a focused protocol/admin tests and frontend typecheck pass.
- S60b contract test groups pass, and verbose output proves the explicit Luna, Messages final-originator omission, image account identity, pairing, exact `0.144.1`, and Windows reset tests actually run.
- `git diff --check`, conflict-marker scan, fixed-baseline allowed-path audit, and final clean-status check pass.
- No live upstream credentials or network requests are used.

## Context

- Business worktree: `E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business`
- Workflow worktree: `E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60`
- Developer contracts: `docs/workflow/tasks/upstream-v0150-safe-compat-s60a.md` and `docs/workflow/tasks/upstream-v0151-codex-luna-identity-s60b.md`
- Expected heads: S60a `7f47afa11`, combined `a07c3e669`.

## Allowed Paths

- `E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60/docs/workflow/qa-reports/upstream-v0151-first-batch-s60-qa.md`
- transient ignored test artifacts under the business worktree `.tmp/**` and `frontend/node_modules/**`

## Denied Paths

- every tracked file in the business worktree
- `knowledge/**`, workflow contracts/status/main-log, migrations, schema, billing, pricing, payment, image routing, deployment, and production configuration
- every report path except the single Allowed QA report

## Constraints

- Review and test only. Do not fix failures, stage files, commit, push, install global tools, or use credentials.
- If frontend dependencies are missing, a local install is allowed only for `node_modules`; remove any generated `frontend/package-lock.json` before the final status check.
- Treat a Go command with `[no tests to run]` as failure for required named cases.
- Keep the report concise and include exact commands plus pass/fail evidence.

## Acceptance Commands

Run against the business worktree:

```powershell
$business = 'E:/codex-worktrees/sub2api/upstream-v0151-first-batch-s60-business'
if (-not (Get-Command rg -ErrorAction SilentlyContinue)) { throw 'required tool rg is unavailable' }
git -C $business status --short --branch
git -C $business merge-base HEAD 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca
git -C $business log --oneline 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca..HEAD
New-Item -ItemType Directory -Force "$business/.tmp/go-build-qa" | Out-Null
$env:GOTMPDIR = (Resolve-Path "$business/.tmp/go-build-qa").Path
Push-Location "$business/backend"
$testCommands = @(
  @('./internal/pkg/apicompat', '-run', 'ParallelToolCalls|ChatCompletionsToResponses|ResponsesToChatCompletionsRequest', '-count=1', '-v'),
  @('./internal/handler/admin', '-run', 'TestGetUserBreakdown', '-count=1', '-v'),
  @('./internal/pkg/openai', '-run', 'Codex|PairCodexClientIdentity', '-count=1', '-v'),
  @('./internal/service', '-run', 'TestCodexVersionConstants_Consistency|TestEnforceCodexIdentityHeaders|Test.*Luna.*Identity|Test.*CompatMessagesBridge.*Originator|TestIsOpenAIWSClientDisconnectError', '-count=1', '-v'),
  @('./internal/service', '-run', 'Test.*(OpenAICodex|CodexIdentity|BuildOpenAIWSHeaders|AccountTest).*', '-count=1', '-v'),
  @('./internal/service', '-run', 'TestOpenAIGatewayServiceForwardAsAnthropicMappedNonCodexOmitsOriginator|TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders|TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestOpenAIBuildUpstreamRequestOAuthOfficialClientOriginatorCompatibility|TestIsOpenAIWSClientDisconnectError|TestCodexVersionConstants_Consistency', '-count=1', '-v')
)
foreach ($testArgs in $testCommands) {
  $testOutput = @(& go test @testArgs 2>&1)
  $testExit = $LASTEXITCODE
  $testOutput
  if ($testExit -ne 0) { throw "go test failed: go test $($testArgs -join ' ')" }
  if (($testOutput -join "`n") -match '\[no tests to run\]') { throw "required tests did not run: go test $($testArgs -join ' ')" }
}
Pop-Location
Push-Location "$business/frontend"
if (-not (Test-Path -LiteralPath 'node_modules/.bin/vue-tsc.cmd')) {
  npm.cmd install --no-package-lock
  if ($LASTEXITCODE -ne 0) { throw 'frontend dependency recovery failed' }
  Remove-Item -LiteralPath 'package-lock.json' -Force -ErrorAction SilentlyContinue
}
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'frontend typecheck failed' }
Pop-Location
git -C $business diff --check 3332c6883e7480f030fcffbccb6dc7ee0a3f69ca..HEAD
$conflictMarkers = @(& rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' "$business/backend" "$business/frontend")
$conflictExit = $LASTEXITCODE
if ($conflictExit -eq 0) { $conflictMarkers; throw 'conflict markers found' }
if ($conflictExit -ne 1) { throw 'conflict marker scan failed to execute' }
$baseline = '3332c6883e7480f030fcffbccb6dc7ee0a3f69ca'
$allowedPaths = @(
  'backend/internal/handler/admin/dashboard_handler.go',
  'backend/internal/handler/admin/dashboard_handler_user_breakdown_test.go',
  'backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go',
  'backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go',
  'backend/internal/pkg/apicompat/chatcompletions_responses_test.go',
  'backend/internal/pkg/apicompat/chatcompletions_to_responses.go',
  'backend/internal/pkg/apicompat/types.go',
  'frontend/src/api/admin/dashboard.ts',
  'backend/internal/pkg/openai/request.go',
  'backend/internal/pkg/openai/request_identity_test.go',
  'backend/internal/service/account_test_service.go',
  'backend/internal/service/account_test_service_openai_compact_test.go',
  'backend/internal/service/account_usage_service.go',
  'backend/internal/service/openai_codex_identity.go',
  'backend/internal/service/openai_codex_identity_test.go',
  'backend/internal/service/openai_codex_version_consistency_test.go',
  'backend/internal/service/openai_gateway_messages.go',
  'backend/internal/service/openai_gateway_service.go',
  'backend/internal/service/openai_gateway_service_test.go',
  'backend/internal/service/openai_oauth_passthrough_test.go',
  'backend/internal/service/openai_ws_forwarder.go',
  'backend/internal/service/openai_ws_forwarder_ingress_test.go',
  'backend/internal/service/openai_ws_forwarder_success_test.go'
)
if ($allowedPaths.Count -ne 23) { throw "contract allowed-path list count is $($allowedPaths.Count), expected 23" }
$changedPaths = @(git -C $business diff --name-only "$baseline...HEAD")
$unexpected = @($changedPaths | Where-Object { $_ -notin $allowedPaths })
$missing = @($allowedPaths | Where-Object { $_ -notin $changedPaths })
if ($unexpected.Count -gt 0) { throw "unexpected paths: $($unexpected -join ', ')" }
if ($missing.Count -gt 0) { throw "expected paths missing: $($missing -join ', ')" }
if ((git -C $business rev-parse HEAD) -ne (git -C $business rev-parse a07c3e669)) { throw 'unexpected business HEAD' }
if ((git -C $business rev-parse HEAD^) -ne (git -C $business rev-parse 7f47afa11)) { throw 'unexpected S60a parent' }
if ((git -C $business rev-parse HEAD^^) -ne $baseline) { throw 'unexpected fixed baseline' }
$finalStatus = @(git -C $business status --short)
$finalStatus
if ($finalStatus.Count -gt 0) { throw 'business worktree is not clean' }
```

The executable `$allowedPaths` audit above is authoritative and must contain exactly the 23 source/test files in commits `7f47afa11` and `a07c3e669`.

## Output

- Write the QA report to `docs/workflow/qa-reports/upstream-v0151-first-batch-s60-qa.md` in the workflow worktree using `C:/Users/Administrator/.codex/templates/qa-report.md`.
- First non-frontmatter line must be `### PASS: upstream-v0151-first-batch-s60-qa`, `### FAIL: ...`, or `### BLOCKED: ...`.
- Include Findings, executed commands, verbose named-test evidence, confirmation that none of the six test groups emitted `[no tests to run]`, allowed/denied-path result, unverified risks, and bug-owner recommendation.

## Stop Rules

- Report `FAIL` on any test failure, no-tests-run false positive, dirty tracked business file, unexpected changed path, or commit mismatch.
- Report `BLOCKED` only for an unavailable required tool or environment that cannot be recovered without changing tracked files.
- Do not repair implementation failures.

## Budget

- qa_worker_model: `gpt-5.4`
- original_qa_worker_model: `deepseek-v4-pro`
- fallback_reason: prescribed model returned HTTP 404; user instructed Codex to continue and Hermes `custom:ai-3zapi` + `gpt-5.4` passed a live handshake.
