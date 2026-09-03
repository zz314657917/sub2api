### DONE: upstream-v0200-ops-proxy-attribution-s291a

## Changed files

- `backend/internal/service/ops_upstream_context.go`
- `backend/internal/service/ops_service.go`
- `backend/internal/service/ops_queue_sanitize_test.go`
- `docs/workflow/status.md`
- `knowledge/tasks/current-task.md`

## Implementation

- Added credential-free proxy ID/name snapshots and direct/unknown normalization.
- Added legacy JSON detail-read normalization without database rewrites.
- Added newest-16 rich payload window, 256-event cap and 512 KiB serialized queue bound with dropped-attempt count.
- Preserved local request-body redaction and existing event sanitization.

## Commands run

- `go test ./internal/service -run 'Test(NormalizeOps|BoundOps|SanitizeOps|AppendOpsUpstreamError|SafeUpstreamURL)' -count=1` PASS
- `go test ./internal/service -run 'TestOps|TestPrepareOpsRequestBodyForQueue' -count=1` PASS
- `go test ./internal/service` PASS
- `go build ./...` PASS
- `git diff --check` PASS
- `git diff --name-only --diff-filter=U` PASS (empty)

## Risks / unverified

- Gateway/provider call-site attribution is intentionally not included; it is deferred to S291-B/C.
- Full repository test suite and runtime provider traffic were not part of this batch.

## knowledge_candidates

- none
