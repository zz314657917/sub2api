# Task Contract: upstream-v0150-safe-compat-s60a

## Task ID

`upstream-v0150-safe-compat-s60a`

## Status

`approved`

## Role

You are the P/G/E Generator worker for the narrow protocol and admin-filter half of S60. Execute only this contract and do not make architecture decisions or expand scope.

## Goal

Port two independent upstream compatibility fixes onto local `main@3332c6883` while preserving the fork's existing architecture:

- `ad8afc8a2`: preserve `parallel_tool_calls`, including explicit `false`, across Chat Completions and Responses request conversion.
- `dda8f7873` plus `ea9f40b63`: parse admin user-breakdown `request_type` with `service.ParseUsageRequestType` and type the frontend parameter as `UsageRequestType`.

## Success Criteria

- Chat Completions -> Responses and Responses -> Chat Completions both preserve a non-nil `*bool` value for `parallel_tool_calls`; JSON output retains explicit `false`.
- `GET /api/v1/admin/dashboard/user-breakdown?request_type=ws_v2|stream|sync` reaches the corresponding enum filter; invalid values return HTTP 400 instead of being silently ignored.
- `UserBreakdownParams.request_type` uses the existing frontend `UsageRequestType` type rather than `number`.
- Focused backend tests, frontend typecheck, formatting, diff checks, conflict-marker scan, and denied-path audit pass.

## Context

- Repo baseline: `F:/mcplugins/sub2api@3332c6883`
- Integration branch: `codex/upstream-v0151-first-batch-s60`
- Read first: `docs/workflow/spec.md`, this contract, and the three upstream commit diffs.
- This worker must adapt patches to local files; it must not merge an upstream tag or branch.
- The primary worktree contains unrelated S58/S59 frontend changes and is untouchable.

## Allowed Paths

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/handler/admin/dashboard_handler.go`
- `backend/internal/handler/admin/dashboard_handler_user_breakdown_test.go`
- `frontend/src/api/admin/dashboard.ts`
- `docs/workflow/worker-results/upstream-v0150-safe-compat-s60a-result.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- payment, subscription, billing, welfare, Studio Bridge, image, Grok, and deployment paths
- `frontend/src/views/**`, `frontend/src/router/**`, and all frontend theme/home files
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, `knowledge/**`
- `C:/Users/Administrator/.codex/memories/**`
- every path not listed under Allowed Paths

## Constraints

- Keep the implementation behaviorally equivalent to the selected upstream commits; do not cherry-pick unrelated test or refactor collateral.
- Preserve explicit `false` with pointer semantics; do not replace it with a plain `bool` plus `omitempty`.
- Reuse `service.ParseUsageRequestType`; do not introduce a second request-type parser.
- Keep the existing numeric `strconv` import if other handler methods still need it.
- Do not modify generated code, schema, migrations, configuration, VERSION, dependencies, docs outside the worker result, or the primary worktree.
- Use scoped staging only; never run `git add .`.

## Acceptance Commands

Run from the worker worktree:

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
New-Item -ItemType Directory -Force '.tmp/go-build' | Out-Null
$env:GOTMPDIR = (Resolve-Path '.tmp/go-build').Path
Push-Location backend
go test ./internal/pkg/apicompat -run 'ParallelToolCalls|ChatCompletionsToResponses|ResponsesToChatCompletionsRequest' -count=1
go test ./internal/handler/admin -run 'TestGetUserBreakdown' -count=1
Pop-Location
Push-Location frontend
npm.cmd run typecheck
Pop-Location
$baseline = '3332c6883e7480f030fcffbccb6dc7ee0a3f69ca'
$goFiles = @(
  git diff --name-only --diff-filter=ACM "$baseline...HEAD"
  git diff --name-only --diff-filter=ACM
  git ls-files --others --exclude-standard
) | Sort-Object -Unique | Where-Object { $_ -like '*.go' -and (Test-Path -LiteralPath $_) }
if ($goFiles.Count -gt 0) { & gofmt -w @goFiles }
git diff --check
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend frontend
$plannerPaths = @(
  'docs/workflow/agent-matrix.md', 'docs/workflow/spec.md', 'docs/workflow/status.md', 'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-v0150-safe-compat-s60a.md',
  'docs/workflow/tasks/upstream-v0151-codex-luna-identity-s60b.md'
)
$allowedPaths = @(
  'backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go',
  'backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go',
  'backend/internal/pkg/apicompat/chatcompletions_responses_test.go',
  'backend/internal/pkg/apicompat/chatcompletions_to_responses.go',
  'backend/internal/pkg/apicompat/types.go',
  'backend/internal/handler/admin/dashboard_handler.go',
  'backend/internal/handler/admin/dashboard_handler_user_breakdown_test.go',
  'frontend/src/api/admin/dashboard.ts',
  'docs/workflow/worker-results/upstream-v0150-safe-compat-s60a-result.md'
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

The fixed-baseline audit must throw on every path outside the worker Allowed Paths and shared Planner artifacts.

## Output

- Commit the scoped implementation on the assigned worker branch.
- Write `docs/workflow/worker-results/upstream-v0150-safe-compat-s60a-result.md` using the worker-result template.
- The report first line must be `### DONE: upstream-v0150-safe-compat-s60a`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Report changed files, commands run, key output, risks, contract compliance, and `knowledge_candidates`.

## Stop Rules

- Stop if any required behavior needs a Denied Path.
- Stop if the local architecture makes either fix non-independent or requires an API/schema change.
- Stop on unexplained focused-test regressions; do not weaken or delete existing assertions.
- Stop if another agent has modified an Allowed Path in this worktree.

## Budget

- worker_model: `gpt-5.4`
- qa_worker_model: `gpt-5.4`
- original_worker_model: `deepseek-v4-pro`
- fallback_reason: the prescribed model returned HTTP 404 before any tool call or code change; after the user instructed Codex to continue, Hermes `custom:ai-3zapi` + `gpt-5.4` passed a live handshake and was approved for this Sprint only.
- worktree_root: `E:/codex-worktrees/sub2api`
