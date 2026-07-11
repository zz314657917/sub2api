# Task Contract: upstream-compact-ops-stream-log-s67c

## Task ID

`upstream-compact-ops-stream-log-s67c`

## Status

`approved`

## Role

You are the Generator worker for the compact in-band Ops observability follow-up.

## Goal

Make HTTP-200 SSE failures marked by `service.MarkOpsStreamError` visible in the existing Ops error log, without duplicating upstream-context errors or changing response behavior.

## Success Criteria

- When wire status is below 400, no upstream-error context exists, and `GetOpsStreamError` is present, one Ops error entry is enqueued.
- Classification uses the intended status for severity/type/business-limited decisions while retaining the actual wire status in the stored status field.
- `OpsSkipPassthroughKey`, monitoring-disabled filters, and existing skip rules remain effective.
- Normal successful 200 responses and 200 responses with upstream-error context are not double logged.
- Focused Ops stream tests and existing Ops writer tests pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream reference implementation: current `upstream/main` `logOpsStreamError` path, introduced before/alongside compact keepalive hardening.
- S66 already provides `service.MarkOpsStreamError` and `GetOpsStreamError` in `openai_compact_stream_bridge.go`.

## Allowed Paths

- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/ops_error_logger_test.go`
- `backend/internal/handler/ops_capture_writer_nil_test.go`
- `docs/workflow/worker-results/upstream-compact-ops-stream-log-s67c-result.md`

## Denied Paths

- All service, gateway, response, database, migration, frontend, deployment, and production configuration paths.
- `knowledge/**` and global memories.

## Constraints

- Reuse existing Ops classification, latency, API key, IP, and enqueue helpers.
- First marker wins; do not modify marker semantics.
- Do not change response writer behavior or wire output.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler -run "TestLogOpsStreamError|TestMarkOpsStreamError|TestOpsCaptureWriter" -count=1
go test ./internal/handler -run "TestOpsErrorLogger" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-compact-ops-stream-log-s67c-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes on the assigned branch and return the commit hash.

## Stop Rules

- Stop if persistence requires schema, repository, or service changes outside Allowed Paths.
- Stop if the patch would duplicate existing upstream-error logging.
- Do not repair unrelated test-suite drift.
