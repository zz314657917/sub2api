# Task Contract: streaming-route-cooldown-s208

## Task ID

`streaming-route-cooldown-s208`

## Status

`approved`

## Role

Direct Codex implementation of a request-local stream-error cooldown signal.

## Goal

Keep the existing multi-group route cooldown behavior effective when a Gateway
or OpenAI handler can only emit a terminal SSE error after response headers
have committed as HTTP `200`.

## Success Criteria

- A terminal handler error classified by the existing cooldown policy (`429`,
  `529`, or `5xx`) marks the current request for API-key route cooldown even
  when the stream was already started.
- API-key middleware applies the selected route's configured cooldown to that
  marker before considering the writer's final HTTP status, so later requests
  skip the cooled group and select the next eligible configured group.
- The Gateway and OpenAI handlers retain their existing protocol-specific SSE
  terminal error frames and do not replay or redirect a started stream.
- An ordinary successful request still clears an existing cooldown; non-stream
  HTTP status handling remains unchanged.
- Focused handler and middleware regressions, dependent package compilation,
  formatting, allowlist, diff, conflict, and unmerged-index checks pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: `main@f50183f83`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Relevant behavior: `applyAPIKeyRouteCooldownAfterRequest` currently derives
  failure only from `c.Writer.Status()`. Both `handleStreamingAwareError`
  implementations emit SSE after a started stream while the writer stays `200`.

## Allowed Paths

- `backend/internal/server/middleware/middleware.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_stream_cooldown_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/stream_error_event_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/streaming-route-cooldown-s208.md`
- `docs/workflow/worker-results/streaming-route-cooldown-s208-result.md`
- `docs/workflow/qa-reports/streaming-route-cooldown-s208-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, repositories, Redis/cache key
  formats, route persistence, route priority/weight algorithms, billing,
  pricing, account selection, retry/failover, dependencies, lockfiles,
  frontend, configuration, Docker, deployment, VERSION, production data,
  push, and global or repository knowledge files.

## Constraints

- Keep the cooldown classification identical to `shouldCooldownAPIKeyRoute`.
- Use request-local state only; do not mutate cached API-key or group objects.
- Do not retry, replay, or redirect a stream after any bytes have been sent.
- Do not turn arbitrary SSE errors into cooldowns; only mark statuses covered by
  the existing HTTP cooldown rule.
- Preserve normal successful-request cooldown clearing and existing non-stream
  HTTP behavior.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler -run 'Test.*Streaming.*RouteCooldown|Test.*StreamingAwareError.*429MarksRouteCooldown' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S208 handler regression failed' }
go test ./internal/server/middleware -run 'Test.*RouteCooldown.*Stream|Test.*RouteCooldown.*HTTP200' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S208 middleware regression failed' }
go test ./internal/handler ./internal/server/middleware
if ($LASTEXITCODE -ne 0) { throw 'S208 package regressions failed' }
go test ./internal/server/routes -run '^$'
if ($LASTEXITCODE -ne 0) { throw 'S208 dependent compile check failed' }
Pop-Location

gofmt -d backend/internal/server/middleware/middleware.go backend/internal/server/middleware/api_key_auth.go backend/internal/server/middleware/api_key_auth_test.go backend/internal/handler/gateway_handler.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/stream_error_event_test.go
git diff --check
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S208 has unmerged index entries' }
```

## Output

- Direct implementation result: `docs/workflow/worker-results/streaming-route-cooldown-s208-result.md`
- QA report: `docs/workflow/qa-reports/streaming-route-cooldown-s208-qa.md`
- Workflow status/log entries for build, QA, and final verdict.

## Stop Rules

- Stop if the fix requires a changed route selection policy, route persistence,
  Redis/cache protocol, billing, account scheduling, schema, configuration, or
  deployment.
- Stop if a started stream would need replay to meet the requested behavior.
- Stop if a non-cooldown client error starts marking a route as unavailable.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `focused runtime and package-level Go verification plus diff review`
- worktree_root: `current primary checkout; preserve user-owned outputs/`

## Contract Review

`PASS`: The contract carries the existing status classification across the
already-committed stream boundary without changing route policy or retrying an
unsafe request. Its allowlist contains the only two handler implementations,
the shared middleware boundary, tests, and required workflow evidence.

## Contract Amendment

`PASS`: The middleware regression is a new default-tag test file, rather than
the legacy unit-tag file, because the repository's unrelated unit-tag baseline
is not a valid S208 acceptance dependency. The product scope and all denied
paths are unchanged.
