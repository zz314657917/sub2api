---
task_id: upstream-v0169-proxy-stream-circuit-integration
status: contract-approved
role: Generator
qa_mode: runtime
source_commits:
  - a4b3f1890
base_commit: 580ecea3c
---

# Task Contract

## Goal

Adapt the S169 OpenAI Responses SSE proxy-stream circuit to current main. Prefer
healthy proxy-backed accounts after repeated incomplete streams, clear state on
successful terminals, and fail open only when quarantine alone removes ordinary
pool capacity.

## Success Criteria

- Circuit settings, bounded in-process state, failure aggregation, TTL expiry,
  success clearing, disabled mode, and fail-open behavior are covered by tests.
- Non-pinned OpenAI scheduler paths prefer healthy proxy accounts and fail open
  only after quarantine is the sole availability blocker.
- Current Pixel Cafe pinned-account paths remain immutable: a pinned account is
  never replaced by another account, and proxy quarantine does not create a
  fallback path. A focused regression makes that deliberate behavior explicit.
- Incomplete server-side SSE streams record the proxy failure; terminal streams
  clear it; client disconnects retain their existing behavior.

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/service/openai_proxy_stream_circuit.go`
- `backend/internal/service/openai_proxy_stream_circuit_test.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-v0169-proxy-stream-circuit-integration.md`
- `docs/workflow/qa-reports/upstream-v0169-proxy-stream-circuit-integration-qa.md`

## Denied Paths

- Deploy, Docker, Compose, CI, Go modules, migrations, frontend, Passkey, Prompt Audit,
  Pixel Cafe business/auth behavior, account-binding model, database, remote, push, and primary-worktree paths.

## Constraints

- Use `a4b3f1890` as the behavioral source, but manually adapt the scheduler around
  current pinned-account routing. Do not overwrite `pinnedAccountIDFromContext` or
  `selectPinnedOpenAIAccount` behavior.
- No new external dependency, persistence, cache, migration, or deployment configuration.
- Stop on any behavior that could route a pinned key to an alternative account.

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/config -run 'Test(LoadDefaultOpenAIWSConfig|ValidateConfig_OpenAIWSRules)' -count=1
go test ./internal/service -run 'Test(OpenAIProxyStreamCircuit|OpenAIGatewayService.*ProxyStream|DefaultOpenAIAccountScheduler.*Proxy|OpenAIGatewayService.*Pinned)' -count=1
go test ./... -run '^$'
go build ./...
Set-Location ..
git diff --check
git ls-files -u
```

## Contract Review

`PASS / contract-approved`: the upstream circuit/config/gateway patches apply to
`main@580ecea3c`; the scheduler conflict is limited to the newer pinned-account
branch. The approved adaptation excludes quarantined accounts only from ordinary
selection, preserves pinned account identity, and retains upstream fail-open
semantics for ordinary pools.
