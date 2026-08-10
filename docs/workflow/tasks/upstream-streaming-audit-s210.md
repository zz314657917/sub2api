# Task Contract: upstream-streaming-audit-s210

## Task ID

`upstream-streaming-audit-s210`

## Status

`approved`

## Role

Direct Codex implementation and evaluator review; no worker is delegated.

## Goal

Behaviorally adapt upstream `2f109e74c` and `c418fd522` to the local OpenAI
handler/audit topology. A compact keepalive must not suppress a terminal
Responses error before semantic SSE output, while identical allowed audit work
within one WebSocket turn must not execute twice.

## Success Criteria

- If a compact keepalive committed only HTTP headers/comments and upstream
  forwarding fails, `ensureForwardErrorResponse` emits exactly one
  `response.failed` event for `/v1/responses` while retaining HTTP `200`.
- If semantic SSE output was already written, existing committed-response
  behavior remains unchanged and no second terminal event is appended.
- The audit helper reuses only a prior `DecisionAllow` whose stage, WebSocket
  turn, and SHA-256 payload hash all match the current audit request.
- A different turn or payload, plus block, unavailable, invalid, or flag
  results, is never reused. Existing HTTP audit completion caching and
  first/subsequent-turn audit coverage remain intact.
- Focused regressions, complete handler package tests, server compilation,
  formatting, exact allowlist, provenance, conflict, and index gates pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-streaming-audit-s210`
- Frozen base: `main@d567ed89e87564bc45f157faf44e0cfbcfb9c7af`
- Upstream sources: `2f109e74caee1a33248744b05a700a65f03bec5c` and
  `c418fd522f429e80c5606d90393d7da601ca30d5`.
- Both source commits are ancestors of fetched `upstream/main`; direct patch
  application is intentionally excluded because the local handler diverges.

## Allowed Paths

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/security_audit_helper.go`
- `backend/internal/handler/security_audit_helper_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-streaming-audit-s210.md`
- `docs/workflow/worker-results/upstream-streaming-audit-s210-result.md`
- `docs/workflow/qa-reports/upstream-streaming-audit-s210-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths

- Every path not listed above, including API-key middleware, route selection,
  billing, account scheduling, repositories, Redis/cache storage, Ent/schema,
  migrations, generated code, frontend, configuration, dependencies,
  lockfiles, Docker, deployment, VERSION, `outputs/`, and repository/global
  knowledge files.
- Remote push, provider calls, shared PostgreSQL/Redis access, containers,
  deployment, production traffic, and cherry-picking or merging upstream
  history.

## Constraints

- Keep the two fixes independent and use existing compact keepalive and audit
  APIs; do not change SSE framing, retry/failover policy, or WebSocket turn
  sequencing.
- Treat keepalive comments/header commits as non-semantic only when the
  adjusted written size shows no business SSE output.
- Cache no audit decision other than `DecisionAllow`; the cache is request-local
  and stores one exact turn/payload entry only.
- Preserve `0` changes to route/cooldown behavior: S210 must not change it.
- No push, deployment, container update, schema/migration, configuration, or
  external runtime operation.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/handler -run '^(TestOpenAIEnsureForwardErrorResponse_CompactKeepaliveOnlyWritesResponseFailed|TestCachesSecurityAuditCompletionSkipsWebSocketStages|TestRunSecurityAuditDoesNotSkipSubsequentWebSocketTurns|TestRunSecurityAuditDeduplicatesRepeatedPayloadWithinWebSocketTurn|TestRunSecurityAuditDoesNotCache(Failed|Flagged)WebSocketDecision)$' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S210 focused handler regressions failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S210 handler package regression failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S210 server compile check failed' }
Pop-Location

gofmt -d backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_gateway_handler_test.go backend/internal/handler/security_audit_helper.go backend/internal/handler/security_audit_helper_test.go
git diff --check
git merge-base --is-ancestor 2f109e74caee1a33248744b05a700a65f03bec5c upstream/main
git merge-base --is-ancestor c418fd522f429e80c5606d90393d7da601ca30d5 upstream/main
```

## Output

- Direct implementation result:
  `docs/workflow/worker-results/upstream-streaming-audit-s210-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-streaming-audit-s210-qa.md`.
- Workflow status/log entries for contract review, build, QA, and final verdict.

## Stop Rules

- Stop if the fix needs route/middleware, persistence, billing, schema,
  configuration, frontend, dependency, deployment, or external-runtime edits.
- Stop if semantic SSE output could receive a duplicate terminal error.
- Stop if any deny/flag/unavailable/invalid audit decision can be cached or a
  different WebSocket turn can reuse the previous decision.
- Stop after two failed implementation attempts and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `focused runtime and package-level Go verification plus diff review`
- worktree_root: `E:/codex-worktrees/sub2api/upstream-streaming-audit-s210`

## Contract Review

`PASS / contract-approved`: `OpenAICompactKeepaliveAdjustedWrittenSize` returns
`-1` for keepalive-only bytes, and the handler already owns a request-local Gin
context for the WebSocket callbacks. The listed handler/helper/test seams cover
both invariants without expanding into route, billing, persistence, or security
policy boundaries.
