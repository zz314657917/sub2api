# Upstream OpenCode Go Reset Duration S262

## Task ID

upstream-opencode-go-reset-s262

## Role

Planner/Generator/Evaluator: Codex defines, reviews, implements, and verifies
this bounded backend correctness port in the protected primary worktree. No
worker dispatch is needed for the self-contained parser change.

## Goal

Adapt the parser portion of upstream `a6b11ccce` so OpenCode Go subscription
`GoUsageLimitError` responses use the reset duration embedded in their message
instead of falling through to the existing short 429 cooldown.

## Success Criteria

- `parseOpenAIRateLimitResetTime` recognizes `GoUsageLimitError` only and
  converts valid `Resets in ...` duration sequences to a future Unix timestamp.
- Compound and common units (`4hr 59min`, days, seconds through weeks) work;
  malformed, zero, negative, overflowing, and unrelated-error messages leave
  the existing fallback behavior unchanged.
- The existing rate-limit pipeline remains the only caller; no scheduler,
  account-runtime blocker, response contract, provider traffic, or persistence
  design is imported from the divergent upstream implementation.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: `main@0c0d2fb86`, clean apart from protected untracked `outputs/**`.
- Upstream provenance: `a6b11ccce` (`fix(openai): honor OpenCode Go usage reset
  durations`). Its runtime-blocker test depends on APIs absent from the local
  topology, but the local generic `RateLimitService` already calls
  `parseOpenAIRateLimitResetTime` before applying normal 429 fallback.

## Allowed Paths

- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_opencode_go_reset_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-opencode-go-reset-s262.md`
- `knowledge/tasks/current-task.md` only after final verification

## Denied Paths

- OpenAI gateway runtime-blocker implementation/tests, account scheduling,
  account state, rate-limit persistence interfaces, provider traffic, response
  contracts, configuration, schemas/migrations, frontend, dependencies,
  containers, deployment, staging/push, shared data, and `outputs/**`.

## Constraints

- Hand-port parser behavior only; do not cherry-pick divergent history.
- Accept duration text only after the case-insensitive `Resets in` marker and
  only for the known `GoUsageLimitError` type.
- Bound parsing and overflow safely; return nil for malformed input so the
  existing short fallback stays authoritative.
- Reuse the file's existing `regexp`, `strconv`, `strings`, and `time`
  facilities; do not add a dependency.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -run 'TestParseOpenAIRateLimitResetTimeOpenCodeGoUsageLimit|TestParseOpenCodeGoUsageLimitResetDuration' -count=1
go test ./internal/service -run '^$'
go test ./cmd/server -run '^$'
gofmt -w internal/service/ratelimit_service.go internal/service/ratelimit_opencode_go_reset_test.go

Set-Location F:/mcplugins/sub2api
git diff --check -- backend/internal/service/ratelimit_service.go backend/internal/service/ratelimit_opencode_go_reset_test.go docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-opencode-go-reset-s262.md knowledge/tasks/current-task.md
git ls-files -u
git diff --cached --name-only
```

The tagged legacy suite is an additional observation only: its existing
compile drift is outside this contract and must not be repaired here.

## Output

- Final evaluation is `PASS`, `FAIL`, or `BLOCKED`, with upstream provenance,
  changed files, actual command output, scope/index review, and residual risk.
- A local implementation commit is allowed only after the acceptance gates
  pass; normal push remains a separate user-authorized action.

## Stop Rules

- Stop if parser behavior requires changing rate-limit persistence, runtime
  blocker, scheduler, account, or response-handler behavior.
- Stop rather than accepting a duration for an unknown error type or malformed
  message.
- Stop rather than touching `outputs/**` or unrelated user work.

## Budget

- controller_implementation: local Codex
- worker_dispatch: disabled for this focused task
- qa_mode: runtime
