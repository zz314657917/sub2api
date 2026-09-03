---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v0200-ops-proxy-attribution-s291a
worker_model: gpt-5.6-terra
base_commit: 43b38cc32fac9b8a478582b0d792f8c06fedbb2a
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-03
---

# S291-A Contract Review Scope

## Task ID

upstream-v0200-ops-proxy-attribution-s291a

## Role

Generator worker or controller implementation for the first independent batch
of the upstream proxy-attribution feature. This batch owns the event contract,
legacy JSON normalization and queued-event bounds only; it does not decide
gateway call-site attribution.

## Goal

Adapt the behavior from upstream commits `e9e3c46cb`, `4c1f920d5` and
`abc07bb07` into the local Ops event and queue owners without merging history.
Every decoded/stored event must have explicit safe proxy attribution semantics,
and queued upstream events must remain bounded while retaining the newest
outcome-deciding attempts.

## Success Criteria

- `OpsUpstreamErrorEvent` carries credential-free `proxy_id`, `proxy_name` and
  `dropped_earlier_attempts` fields with the direct/unknown normalization
  invariants documented by the upstream behavior.
- `ParseOpsUpstreamErrors` and detail-read normalization materialize legacy
  missing attribution as `proxy_id: null, proxy_name: "unknown"` without
  rewriting unrelated JSON keys or stored database rows.
- Queue sanitization keeps rich payloads for the newest 16 attempts, enforces
  at most 256 events and approximately 512 KiB serialized bytes, and stamps the
  oldest retained event with the count of dropped earlier attempts.
- Existing request-body redaction, status/message sanitization and empty-event
  filtering remain intact.
- Focused service tests cover managed/direct/unknown/legacy attribution,
  malformed attribution, body-window behavior, count/byte bounds and newest
  event retention.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`,
  `knowledge/tasks/current-task.md`
- Upstream baseline: `upstream/main@b1748c4ea`
- Local base: `main@43b38cc32`
- Existing dirty paths are user-owned and must remain byte-for-byte unchanged.

## Allowed Paths

- `backend/internal/service/ops_upstream_context.go`
- `backend/internal/service/ops_upstream_context_test.go`
- `backend/internal/service/ops_service.go`
- `backend/internal/service/ops_queue_sanitize_test.go`
- `backend/internal/service/ops_service_batch_test.go`
- `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291a.md`
- `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291a-review.md`
- `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291a-result.md`

## Denied Paths

- Every gateway/provider implementation, including `gateway_service.go`,
  `antigravity_gateway_service.go`, Gemini, Bedrock, OpenAI, Grok and WebSocket
  forwarding owners. Those call sites belong to later S291-B/C contracts.
- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/admin_service.go`
- `frontend/**`
- `outputs/**`
- `go.mod`, `go.sum`, schema, migrations, repositories, billing, deployment,
  containers, shared data and production configuration.
- `knowledge/**` and `C:/Users/Administrator/.codex/memories/**`
- Any path not listed under Allowed Paths.

## Constraints

- Do not merge, rebase or cherry-pick upstream commits.
- Do not infer historical proxy identity from current account state in this
  batch; only normalize event fields already present in JSON/structs.
- Store only proxy ID/name; never persist proxy URLs or credentials.
- Preserve all pre-existing dirty files and the protected dirty diff hash
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`.
- Keep changes minimal and compatible with the local monolithic service
  topology. No unrelated formatting or refactoring.

## Acceptance Commands

```powershell
Set-Location backend
gofmt -w internal/service/ops_upstream_context.go internal/service/ops_service.go internal/service/ops_upstream_context_test.go internal/service/ops_queue_sanitize_test.go internal/service/ops_service_batch_test.go
go test ./internal/service -run 'Test(AppendOpsUpstreamError|ParseOpsUpstreamErrors|.*OpsUpstream.*|.*OpsQueue.*|PrepareOpsRequestBodyForQueue)' -count=1
go test ./internal/service -run 'TestOps|TestPrepareOpsRequestBodyForQueue' -count=1
go test ./internal/service
go test ./...
go build ./...
git diff --check
git diff --name-only --diff-filter=U
```

The repository's known account repository fixture drift (`32` vs `34` columns)
must be reported separately if it remains the only full-suite failure.

## Output

- Write `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291a-result.md`
  using the worker-result template; its first line must be `### DONE: ...`,
  `### BLOCKED: ...` or `### FAILED: ...`.
- Report changed files, commands, results, risks and `knowledge_candidates`.
- Do not write long-term knowledge from the worker.

## Stop Rules

- Stop if any denied path changes, if existing dirty hashes move, or if a
  gateway call-site must be changed to make this batch pass.
- Stop on contract ambiguity, schema/dependency requirements or an owner
  conflict; return the issue to the controller instead of expanding scope.
- Do not claim complete upstream proxy attribution until S291-B and S291-C are
  separately implemented and verified.

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`
