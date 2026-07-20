# Task Contract: upstream-main-compat-s84

## Task ID

`upstream-main-compat-s84`

## Status

`approved`

## Role

Direct Codex behavior-level port of upstream
`01062c324429636b36d349036fdf2d08e1dce699`. Preserve the local OpenAI
compatibility service topology; do not merge upstream history.

## Goal

Ensure buffered Anthropic-compatible responses are serialized with
`Content-Type: application/json; charset=utf-8` even when the upstream
Responses stream advertises `text/event-stream`.

## Success Criteria

- `handleAnthropicBufferedStreamingResponse` overrides the filtered upstream
  content type immediately before writing the JSON response.
- The buffered response body and usage conversion remain unchanged.
- Streaming responses continue using `text/event-stream` and are not changed.
- A local-signature regression proves a buffered SSE upstream response returns
  JSON content type and a valid Anthropic response body.
- Focused Go tests, `gofmt`, exact path/conflict/diff gates, and protected
  primary-checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s84`
- Branch: `codex/upstream-main-compat-s84`
- Baseline: `631a8f5a89b56972fe6cab4066181ff611d79c0d`
- Upstream snapshot: `b8e844f4ee130ac069a7c5713c2413233186b83f`
- Upstream source: `01062c324429636b36d349036fdf2d08e1dce699`
- Local topology: the target function takes `resp, c, originalModel,
  billingModel, upstreamModel, startTime`; the upstream test has an extra
  account argument and must be rewritten locally.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_compat_model_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s84.md`
- `docs/workflow/worker-results/upstream-main-compat-s84-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s84-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Other gateway/service files, handlers, routers, API DTOs, billing, failover,
  migrations, Ent, generated code, frontend, dependencies, deployment,
  Compose, VERSION, and lockfiles.
- Streaming response headers, `ResponsesToAnthropic` conversion, usage
  accounting, request routing, upstream retry/failover behavior, and any
  response body schema change.
- `knowledge/**`, global memories, handoff/timeline files, and all other
  upstream candidates.

## Constraints

- Work only in the isolated S84 worktree and preserve the dirty primary checkout.
- Set the JSON content type after `WriteFilteredHeaders`, immediately before
  `c.JSON`, so an upstream SSE header cannot win while local header filtering
  remains intact.
- Adapt the regression to local function arguments and existing test helpers;
  do not import unrelated upstream test assumptions.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S84 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s84'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build-s84'
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
go test ./internal/service -run '^TestHandleAnthropicBufferedStreamingResponse_OverridesUpstreamContentType$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S84 focused service test failed' }
Pop-Location

gofmt -w backend/internal/service/openai_gateway_messages.go backend/internal/service/openai_compat_model_test.go
if ($LASTEXITCODE -ne 0) { throw 'S84 gofmt failed' }
git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S84 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S84 has unmerged index entries' }
```

Evaluator additionally reviews that the header override is buffered-only,
streaming `text/event-stream` code is untouched, body/usage conversion has no
diff, audits all paths, scans real conflict markers, and rechecks all three
protected hashes.

### Pre-commit Tracking Gate

```powershell
git add -u -- backend/internal/service/openai_gateway_messages.go backend/internal/service/openai_compat_model_test.go docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -f -- docs/workflow/tasks/upstream-main-compat-s84.md docs/workflow/worker-results/upstream-main-compat-s84-result.md docs/workflow/qa-reports/upstream-main-compat-s84-qa.md
$expected = @(
  'backend/internal/service/openai_gateway_messages.go',
  'backend/internal/service/openai_compat_model_test.go',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s84.md',
  'docs/workflow/worker-results/upstream-main-compat-s84-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s84-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) { throw "S84 staged path set differs from allowlist: $($pathDelta | Out-String)" }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S84 has unmerged index entries' }
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) { throw "S84 contains conflict markers: $($conflictMarkers -join ', ')" }
if ($LASTEXITCODE -ne 1) { throw 'S84 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S84 cached diff check failed' }
```

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s84-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s84-qa.md`
- Workflow status/log entries for contract review, implementation, QA, verdict.

## Stop Rules

- Stop if the patch changes streaming headers, body conversion, usage,
  failover, routing, or any path outside the eight-item allowlist.
- Stop if the focused test is undiscovered, unmerged/conflict markers appear,
  or a protected hash changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a two-file backend compatibility batch`
- qa_mode: `fresh focused Go test plus evidence-first diff review`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
