# Task Contract: upstream-remote-compact-reliability-s66c

## Task ID

`upstream-remote-compact-reliability-s66c`

## Status

`approved`

## Role

You are the Generator worker for the remote-compact reliability lane. This worker exclusively owns the compact core files listed below.

## Goal

Selectively adapt the upstream remote-compact reliability series so SSE-to-JSON reconstruction preserves compaction items and long unary waits remain alive without corrupting failover or concurrent writer behavior.

## Success Criteria

- Raw `response.output_item.done` compaction items are preserved byte-semantically when terminal output is empty or lacks the compaction item; `output_item.added` is only a fallback.
- Streaming body-signal compact requests send downstream SSE comment keepalives after the configured interval.
- Keepalive bytes do not suppress account failover or count as business response output.
- Keepalive and response writes are synchronized; post-commit errors become `response.failed` events rather than invalid JSON/status rewrites.
- Existing path-based compact and body-signal behavior remains compatible.
- Focused compact tests and targeted race coverage pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream references, in order: `2cffe1cf5`, `ae9a01d85`, `000f6dc65`.
- Local tree predates the upstream service-file split. Adapt behavior to the local layout; do not import broad refactors.

## Allowed Paths

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_gateway_passthrough.go`
- `backend/internal/service/openai_gateway_passthrough_test.go`
- `backend/internal/service/openai_compact_sse_keepalive.go`
- `backend/internal/service/openai_compact_sse_keepalive_test.go`
- `backend/internal/service/openai_compact_stream_bridge.go`
- `backend/internal/service/openai_compact_stream_bridge_test.go`
- `docs/workflow/worker-results/upstream-remote-compact-reliability-s66c-result.md`

## Denied Paths

- All paths not listed above.
- Billing/pricing, migrations, frontend, account scheduling policy, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Port behavior, not the upstream file-splitting refactor.
- Preserve local compact routing, endpoint normalization, image accounting, and error logging.
- No changes to charging formulas or retry counts.
- Add comments only where concurrency or wire semantics are non-obvious.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "Test.*Compact.*SSE|Test.*Compact.*Keepalive|Test.*Compact.*Output|Test.*RemoteCompact" -count=1
go test ./internal/handler -run "Test.*Compact|TestOpenAIGatewayHandler" -count=1
go test -race ./internal/service -run "Test.*Compact.*Keepalive" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-remote-compact-reliability-s66c-result.md`.
- First line must be `### DONE: upstream-remote-compact-reliability-s66c`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit the implementation and report on the assigned worktree branch.

## Stop Rules

- Stop if the port requires importing the upstream service split, modifying billing, or changing files outside Allowed Paths.
- Stop if keepalive integration cannot preserve existing failover semantics with focused tests.
- Do not weaken or delete existing compact regression coverage.
