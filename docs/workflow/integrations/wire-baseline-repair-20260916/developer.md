### DONE: wire-baseline-repair

## Initial BLOCKED history

## Scope and changed paths

- Worktree: `E:/codex-worktrees/sub2api/wire-baseline-repair`
- Immutable base: `d69da80b99ac88a246fc7502414c2215a31dde18`
- Business changes made only in the approved allowlist:
  - `backend/internal/service/wire.go`
  - `backend/internal/service/wire_provider_adapters.go`
  - `backend/internal/service/wire_provider_adapters_test.go`
  - `backend/cmd/server/wire_gen.go` (first `go generate` output only)
- The pre-existing dirty baseline files `docs/workflow/main-log.md` and
  `docs/workflow/status.md` were not changed by this worker.

## Implemented repair

- Added the fixed-argument `ProvideContentModerationService` adapter. It forwards
  all original constructor inputs and exactly one `ProxyRepository` to the
  unchanged variadic `NewContentModerationService` constructor.
- Registered the adapter in `ProviderSet` and added
  `wire.Bind(new(newUserTrialConsumer), new(*WelfareService))`.
- Added `TestWireProvider*` tests for full adapter forwarding, the two safe
  one-sided non-nil worker-guard inputs, nil proxy preservation, and compile-time
  `*WelfareService` satisfaction of `newUserTrialConsumer`. The tests never call
  settlement construction and never make both moderation worker guard inputs
  non-nil.

## Executed commands

| Command | Exit code | Result |
| --- | ---: | --- |
| `go test ./internal/service -list 'WireProvider|CompositeRoute'` | 0 | Listed both `TestWireProviderContentModerationAdapterForwardsDependencies` and `TestWireProviderWelfareServiceImplementsNewUserTrialConsumer`, plus existing Composite selectors. |
| `go test ./internal/service -run 'WireProvider|CompositeRoute' -count=1` | 0 | Passed; rerun after the worker-guard test addition also passed (0). |
| `go test ./internal/handler/admin -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run CompositeRoute -count=1` | 0 | Passed. |
| `gofmt -w internal/service/wire.go internal/service/wire_provider_adapters.go internal/service/wire_provider_adapters_test.go` | 0 | Applied only to the three allowed source/test files. |
| `go generate ./cmd/server` (first run) | 0 | Wrote `cmd/server/wire_gen.go`; first SHA-256: `89CFFE7002ECD0594700188E1365CE30DB0CCA90D984F187270DD26A0B880477`. |

## Full first generated diff analysis and blocker

The first generator diff includes the expected call replacement:

```go
service.NewContentModerationService(..., proxyRepository)
// became
service.ProvideContentModerationService(..., proxyRepository)
```

It also moves existing construction/call sites for CN provider quota and balance,
Gemini/OpenAI OAuth dependencies, usage/identity/TLS caches, and the settlement
service. Those hunks appear to be generator ordering changes, but were not
accepted as sufficient proof because the same generated diff also removes two
unrelated behavioral hooks:

```go
if quotaReset, ok := adminService.(service.CafeQuotaResetService); ok {
	 cafeRoomHandler.SetQuotaResetService(quotaReset)
}

openAIGatewayHandler.SetOpsService(opsService)
```

These removals are outside this task's two verified Wire dependencies and cannot
be repaired, manually restored, or accepted as equivalent ordering under the
approved contract. Per the stop rule for unrelated generation changes or
behavioral drift, generation was not run a second time and no manual edit was
made to `wire_gen.go`.

## Not run after blocker

- Second `go generate ./cmd/server` and second SHA-256 comparison.
- `go test ./cmd/server -run '^$' -count=1`.
- `go build ./...`.
- Final `gofmt -l`, `git diff --check`, `git ls-files -u`, final status and full
  diff acceptance gate.

## Required next action

Planner/Evaluator must review the unrelated generated hook removals before any
further generation, manual generated-file handling, or QA. This worker releases
the task owner without starting services, providers, DB/Redis, network activity,
commit, reset, clean, or changes outside the allowlist.

## Approved amendment implementation

The Planner amendment and independent review were read after the initial stop;
both are `approved` / `PASS`. The initial generated output was retained and the
following authorized source providers and tests were added without altering
constructor signatures or handler business logic:

- `handler.ProvideOpenAIGatewayHandler` now receives `*service.OpsService`,
  retains the existing coordinator assignment, and calls the existing
  `SetOpsService` setter. Its regression test asserts both exact identities.
- `admin.ProvideCafeRoomHandler` calls the unchanged
  `NewCafeRoomHandlerWithActivation` with its three original dependencies, then
  performs the original optional `service.CafeQuotaResetService` assertion on
  `service.AdminService`. Its tests cover supported, unsupported, and nil admin
  services without invoking reset, DB, provider, or network work.
- `handler.ProviderSet` registers `admin.ProvideCafeRoomHandler` in place of
  the original constructor. The service adapter and `newUserTrialConsumer` bind
  from the first implementation remain unchanged.

## Amendment acceptance commands

| Command | Exit code | Result |
| --- | ---: | --- |
| `go test ./internal/service -list 'WireProvider|CompositeRoute'` | 0 | Listed all existing Composite selectors and three `TestWireProvider*` service tests. |
| `go test ./internal/service -run 'WireProvider|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/handler -run WireProvider -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run 'WireProvider|CafeRoom.*Reset|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run CompositeRoute -count=1` | 0 | Passed. |
| first `go generate ./cmd/server` | 0 | SHA-256 `CF72C3215F97B5877199E01D27C10EC448151B449AD8613319382E870562C5C9`. |
| second `go generate ./cmd/server` | 0 | SHA-256 `CF72C3215F97B5877199E01D27C10EC448151B449AD8613319382E870562C5C9`; generated bytes are identical. |
| `go test ./cmd/server -run '^$' -count=1` | 0 | Compiled successfully; no tests to run. |
| `go build ./...` | 0 | Passed. |
| `gofmt -l` over all eight allowed Go files | 0 | No output; all files formatted. |
| root `git diff --check` | 0 | Passed; only pre-existing workflow CRLF warnings were emitted. |
| root `git ls-files -u` | 0 | No unmerged entries. |

## Amended generated diff and side-effect review

The complete first amendment diff was inspected. Expected source-level changes
are present:

- `NewContentModerationService` becomes `ProvideContentModerationService` with
  the same eight arguments, including exactly one `proxyRepository`.
- `NewCafeRoomHandlerWithActivation(...)` plus its handwritten optional setter
  becomes `admin.ProvideCafeRoomHandler(..., adminService)`. The adapter retains
  that constructor and assertion/setter internally.
- `ProvideOpenAIGatewayHandler(..., coordinator)` plus its handwritten
  `SetOpsService(opsService)` becomes one provider call with the same existing
  arguments followed by `opsService`; the provider performs the same setter.

All remaining generated hunks are Wire ordering moves: Gemini/OpenAI OAuth
dependencies, usage/identity/TLS caches, CN quota/balance/check construction,
and `ProvideUsageBillingSettlementService`. Constructor argument lists are
unchanged. The two `Start()`-calling providers remain called once with identical
arguments: settlement (its durable dependencies are all constructed before its
new call) and CN balance check (its account, balance, quota, and config
dependencies are all constructed before its new call). Both still occur before
`provideCleanup` is built. The cleanup provider call, its full ordered parameter
list, and the cleanup function body are unchanged; it still receives the same
settlement and CN balance-check instances. No extra missing hook, provider error,
constructor semantic change, or unprovable side-effect ordering drift was found.

## Final inventory and remaining risk

- Modified/generated business files are exactly the eight approved Go paths;
  the two existing workflow baseline modifications remain separate and untouched.
- This is local source/generator/build evidence only. No DB, Redis, provider,
  network, service startup, real administrator HTTP, P2/P3 runtime acceptance,
  commit, or deployment was performed.
- Developer owner is released for fresh independent Terra QA.
