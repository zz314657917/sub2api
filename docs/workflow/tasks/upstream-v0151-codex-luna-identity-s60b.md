# Task Contract: upstream-v0151-codex-luna-identity-s60b

## Task ID

`upstream-v0151-codex-luna-identity-s60b`

## Status

`approved`

## Role

You are the P/G/E Generator worker for the Codex identity and `gpt-5.6-luna` 404 half of S60. Execute only this contract and do not expand into GPT-5.6 billing, model catalog, image routing, or unrelated gateway behavior.

## Goal

Adapt upstream `657c4f97d`, `8a51119e3`, and the narrow Windows disconnect behavior from `0a5f34a2e` to the local monolithic OpenAI gateway so ChatGPT Codex requests use a supported client version and a matched `originator`/final `User-Agent` identity across normal forwarding, passthrough, WebSocket, quota probe, and account-test paths.

## Success Criteria

- `codexCLIVersion`, `openAICodexProbeVersion`, and every default Codex UA consistently use `0.144.1`.
- A shared helper derives a valid official `originator` from the final outbound Codex UA, preserves a valid paired identity, and falls back to the default `codex_cli_rs` identity when pairing is impossible.
- A present `version` header below `0.144.0` is raised to the current `codexCLIVersion`; absent version headers remain absent unless the existing call path intentionally sets one.
- Identity enforcement runs only on ChatGPT Codex requests that carry `originator`; compatibility paths that intentionally omit it remain unchanged.
- `ForwardAsAnthropic` marks the OpenAI compatibility Messages bridge before upstream request construction, including mapped non-GPT-5/Codex models, and a fake/recorder test proves the final bridge request still omits `originator`.
- Normal HTTP forwarding, passthrough, WebSocket, quota probe, standard account test, and image account test all apply identity pairing after their final UA selection.
- `isOpenAIWSClientDisconnectError` recognizes `An existing connection was forcibly closed by the remote host.` without widening unrelated error classification.
- Tests cover a request whose model is explicitly `gpt-5.6-luna`, final UA version `0.144.1`, paired originator/UA first segment, custom `codex-tui` UA pairing, invalid/third-party UA fallback, low version repair, absent-originator no-op, Windows reset, and representative HTTP/WS/probe/account-test call sites without real network access.
- Focused tests, formatting, diff checks, conflict-marker scan, and denied-path audit pass.

## Context

- Repo baseline: `F:/mcplugins/sub2api@3332c6883`
- Integration branch: `codex/upstream-v0151-first-batch-s60`
- Upstream source commits: `657c4f97d` and `8a51119e3`.
- The local fork keeps several upstream split files inside `openai_gateway_service.go` and `openai_ws_forwarder.go`; adapt behavior to local ownership rather than recreating every upstream file split.
- The primary worktree contains unrelated S58/S59 frontend changes and is untouchable.

## Allowed Paths

- `backend/internal/pkg/openai/request.go`
- `backend/internal/pkg/openai/request_test.go`
- `backend/internal/pkg/openai/request_identity_test.go`
- `backend/internal/service/openai_codex_identity.go`
- `backend/internal/service/openai_codex_identity_test.go`
- `backend/internal/service/openai_codex_version_consistency_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_oauth_passthrough_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/account_usage_service_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_openai_compact_test.go`
- `docs/workflow/worker-results/upstream-v0151-codex-luna-identity-s60b-result.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- GPT-5.6 pricing, billing, usage-log, model-catalog, image-generation, Grok, payment, subscription, welfare, and deployment paths
- all frontend files
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- every path not listed under Allowed Paths

## Constraints

- Do not blindly cherry-pick either upstream commit; preserve local gateway structure and existing local headers not involved in Codex identity.
- Identity pairing must use the final outbound UA after account custom UA, fingerprint, ForceCodexCLI, and existing fallback logic.
- The Messages compatibility marker must be set before request construction; post-build header deletion alone is not sufficient evidence.
- Do not fabricate a live upstream test or require credentials. Test request construction and headers with local fakes/recorders.
- Do not change model mappings, prices, billing, compact behavior, image policy, settings schema, VERSION, dependencies, or public API contracts.
- Keep the helper narrow and shared; do not duplicate header-repair logic across call sites.
- Use scoped staging only; never run `git add .`.

## Acceptance Commands

Run from the worker worktree:

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
New-Item -ItemType Directory -Force '.tmp/go-build' | Out-Null
$env:GOTMPDIR = (Resolve-Path '.tmp/go-build').Path
Push-Location backend
go test ./internal/pkg/openai -run 'Codex|PairCodexClientIdentity' -count=1
go test ./internal/service -run 'TestCodexVersionConstants_Consistency|TestEnforceCodexIdentityHeaders|Test.*Luna.*Identity|Test.*CompatMessagesBridge.*Originator|TestIsOpenAIWSClientDisconnectError' -count=1
go test ./internal/service -run 'Test.*(OpenAICodex|CodexIdentity|BuildOpenAIWSHeaders|AccountTest).*' -count=1
Pop-Location
$baseline = '3332c6883e7480f030fcffbccb6dc7ee0a3f69ca'
$goFiles = @(
  git diff --name-only --diff-filter=ACM "$baseline...HEAD"
  git diff --name-only --diff-filter=ACM
  git ls-files --others --exclude-standard
) | Sort-Object -Unique | Where-Object { $_ -like '*.go' -and (Test-Path -LiteralPath $_) }
if ($goFiles.Count -gt 0) { & gofmt -w @goFiles }
git diff --check
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend
$plannerPaths = @(
  'docs/workflow/agent-matrix.md', 'docs/workflow/spec.md', 'docs/workflow/status.md', 'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-v0150-safe-compat-s60a.md',
  'docs/workflow/tasks/upstream-v0151-codex-luna-identity-s60b.md'
)
$allowedPaths = @(
  'backend/internal/pkg/openai/request.go', 'backend/internal/pkg/openai/request_test.go',
  'backend/internal/pkg/openai/request_identity_test.go',
  'backend/internal/service/openai_codex_identity.go', 'backend/internal/service/openai_codex_identity_test.go',
  'backend/internal/service/openai_codex_version_consistency_test.go',
  'backend/internal/service/openai_gateway_service.go', 'backend/internal/service/openai_gateway_service_test.go',
  'backend/internal/service/openai_gateway_messages.go', 'backend/internal/service/openai_oauth_passthrough_test.go',
  'backend/internal/service/openai_ws_forwarder.go', 'backend/internal/service/openai_ws_forwarder_ingress_test.go',
  'backend/internal/service/openai_ws_forwarder_success_test.go',
  'backend/internal/service/account_usage_service.go', 'backend/internal/service/account_usage_service_test.go',
  'backend/internal/service/account_test_service.go',
  'backend/internal/service/account_test_service_openai_compact_test.go',
  'docs/workflow/worker-results/upstream-v0151-codex-luna-identity-s60b-result.md'
) + $plannerPaths
$changedPaths = @(
  git diff --name-only "$baseline...HEAD"
  git diff --name-only
  git ls-files --others --exclude-standard
) | Sort-Object -Unique
$unexpected = @($changedPaths | Where-Object { $_ -notin $allowedPaths })
if ($unexpected.Count -gt 0) { throw "Denied path(s): $($unexpected -join ', ')" }
git status --short
```

The fixed-baseline audit must throw on every path outside the worker Allowed Paths and shared Planner artifacts. Required new test files must be created only when they carry named assertions from Success Criteria.

## Output

- Commit the scoped implementation on the assigned worker branch.
- Write `docs/workflow/worker-results/upstream-v0151-codex-luna-identity-s60b-result.md` using the worker-result template.
- The report first line must be `### DONE: upstream-v0151-codex-luna-identity-s60b`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Report changed files, commands run, key output, risks, contract compliance, and `knowledge_candidates`.

## Stop Rules

- Stop if completing identity pairing requires pricing, billing, image-routing, schema, migration, frontend, or production configuration changes.
- Stop if the final-UA ownership cannot be identified for a required call path without an architecture refactor.
- Stop on unexplained focused-test regressions; do not weaken or delete existing assertions.
- Stop if another agent has modified an Allowed Path in this worktree.

## Budget

- worker_model: `gpt-5.4`
- qa_worker_model: `gpt-5.4`
- original_worker_model: `deepseek-v4-pro`
- fallback_reason: the prescribed model returned HTTP 404 before any tool call or code change; after the user instructed Codex to continue, Hermes `custom:ai-3zapi` + `gpt-5.4` passed a live handshake and was approved for this Sprint only.
- worktree_root: `E:/codex-worktrees/sub2api`
