---
status: approved
review_verdict: PASS
task_id: upstream-main-ws-cap-s302
worker_model: gpt-5.6-terra
base_commit: 2d0144ffc
spec_ref: docs/workflow/plans/upstream-main-selective-integration-20260913.md
---

## Task ID

`upstream-main-ws-cap-s302`

## Role

Developer Worker: adapt upstream `956f4672e` to the local OpenAI WebSocket
pool. Independent QA reviews the resulting diff and runs the acceptance
commands.

## Goal

In `mode_router_v2`, keep non-positive account concurrency unschedulable while
calculating positive effective context-pool capacity through the existing
dynamic factor and global hard cap. Preserve legacy behavior, account type
factors, retained-session reuse, and routing/retry semantics.

## Success Criteria

- Positive mode-router-v2 capacities apply OAuth/API-key factors, minimum-one,
  and the `MaxConnsPerAccount` hard cap.
- Zero or negative account concurrency remains capacity zero in mode-router-v2.
- Nil account and legacy mode preserve local hard-cap behavior.
- Retained sessions reuse connections and cannot exceed effective capacity.
- No unrelated config, UI, scheduler, retry, billing, provider, schema, or
  migration behavior changes.

## Allowed Paths

- `backend/internal/service/openai_ws_pool.go`
- `backend/internal/service/openai_ws_pool_test.go`
- `backend/internal/config/config.go` only for matching comments
- `deploy/config.example.yaml` only for matching comments
- `docs/workflow/**` evidence files

## Denied Paths

- `backend/migrations/**`, `backend/ent/**`, database or schema files
- `frontend/**` and all lockfiles
- provider credentials, runtime data, `outputs/**`, and untracked artifacts
- gateway routing, retry, scheduler, billing, payment, auth, or deployment
  behavior
- merge, rebase, cherry-pick, push, container or database operations

## Constraints

Port behavior manually; do not cherry-pick `956f4672e`. Preserve the dirty
primary worktree and use a clean task worktree. Keep changes minimal and
gofmt-compliant. Do not claim real provider or runtime acceptance from unit
tests.

## Acceptance Commands

```powershell
go test ./internal/service -run 'TestOpenAIWSConnPool_(EffectiveMaxConnsByAccount|AcquireRetainedSessionsUsesScaledCapacity|AcquireRejectsWhenEffectiveMaxConnsIsZero)' -count=1
go build ./...
gofmt -w backend/internal/service/openai_ws_pool.go backend/internal/service/openai_ws_pool_test.go
git diff --check
git diff --name-only -- backend/internal/service/openai_ws_pool.go backend/internal/service/openai_ws_pool_test.go backend/internal/config/config.go deploy/config.example.yaml
git diff --name-only --diff-filter=U
```

## Output

Write `docs/workflow/worker-results/upstream-main-ws-cap-s302-result.md` with a
first-line `### DONE: upstream-main-ws-cap-s302` or `### BLOCKED:` / `### FAILED:`
verdict, changed files, commands, evidence, contract compliance, and remaining
runtime/provider risks.

## Stop Rules

Stop and report `BLOCKED` if implementation requires schema/migration, changes
routing/retry/scheduler semantics, touches denied paths, or the local owner
cannot preserve legacy behavior without an architecture decision.

