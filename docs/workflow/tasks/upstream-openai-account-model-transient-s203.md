# Task Contract

- Task ID: `upstream-openai-account-model-transient-s203`
- Role: Generator and Evaluator (direct Codex implementation; no worker is delegated).
- Goal: Behaviorally adapt upstream `40b8f04a6` and `7d38e6712` so retryable OpenAI failures are cooled down per account+model pair, including sparse traffic where failures are separated by more than one minute.

## Success Criteria

- A bounded, case-normalized in-process account+model state records retryable failures. The second consecutive failure blocks only that pair for 10 seconds, the third and later failures for 45 seconds; a successful request clears the pair.
- A failure streak is retained for 30 minutes of inactivity. Five-minute gaps must still reach the short and long cooldowns; a clock rollback or a stale entry resets safely.
- OpenAI account selection skips a currently blocked account+model pair without making the same account unavailable for another requested model.
- Relevant HTTP, image, passthrough, WebSocket and failover paths report the requested model on retryable failure and report success after a terminal success. Client cancellation and non-retryable/user errors must not poison the transient state.
- Focused state, selection/failover and existing OpenAI service/handler regressions pass. No persistent scheduler state, configuration, dependency, schema/migration, frontend, container, deployment or production behavior is changed.

## Context

- Repo: `F:/mcplugins/sub2api`
- Upstream source: `40b8f04a6 fix(openai): scope transient cooldowns by model`; `7d38e6712 fix(openai): keep transient failure streak from resetting on sparse traffic`.
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, `backend/internal/service/openai_account_scheduler.go`, `backend/internal/service/openai_proxy_stream_circuit.go`.

## Allowed Paths

- `backend/internal/service/openai_account_model_transient.go`
- `backend/internal/service/openai_account_model_transient_test.go`
- `backend/internal/service/openai_account_scheduler*.go`
- `backend/internal/service/openai_gateway*.go`
- `backend/internal/service/openai_images*.go`
- `backend/internal/service/openai_upstream_transport_error*.go`
- `backend/internal/service/openai_ws_*.go`
- `backend/internal/handler/openai_*.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-openai-account-model-transient-s203.md`
- `docs/workflow/qa-reports/upstream-openai-account-model-transient-s203-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, `backend/go.mod`, `backend/go.sum`, `deploy/**`, `Dockerfile*`, `frontend/**`, `outputs/**`
- Any provider credential, shared database/Redis, container, deployment, production or remote Git action.

## Constraints

- Port behavior, not the upstream file layout. Preserve existing local identity, proxy stream circuit, agent-identity and fixed-account routing behavior.
- The transient state is process-local and bounded; it must not write Account state, scheduler cache, Redis or database data.
- Do not treat client cancellation, validation failure, or ordinary upstream user errors as retryable transient failures.
- Keep the existing S180 browser QA separate; do not alter its files or claims.

## Acceptance Commands

```powershell
go test ./internal/service -run 'TestOpenAIModelTransient|TestOpenAI.*Transient|TestOpenAISelectAccount' -count=1
go test ./internal/service -count=1
go test ./internal/handler -run 'TestOpenAI|Test.*Failover' -count=1
go test ./cmd/server -run '^$' -count=0
gofmt -w <changed Go files>
git diff --check
```

## Output

- Write `docs/workflow/qa-reports/upstream-openai-account-model-transient-s203-qa.md` with an evidence-backed `PASS`, `FAIL` or `BLOCKED` verdict.
- Record P/G/E transitions in `docs/workflow/main-log.md`; update task handoff only after final evidence is available.
- Commit only this contract's allowed files to local `main`; do not push.

## Stop Rules

- Stop if the port requires migration, config, dependency, frontend, persistent scheduler-state or production changes.
- Stop if a local path cannot carry the requested model without changing an unrelated protocol contract.
- Stop if focused failures show existing fixed-account, identity, proxy circuit or agent-identity behavior would regress.
