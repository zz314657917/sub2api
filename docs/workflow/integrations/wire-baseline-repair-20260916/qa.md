### PASS: wire-current-integration

# Current-main Wire repair independent QA

## Scope and inventory

- QA worktree: `E:/codex-worktrees/sub2api/wire-current-integration`; immutable
  committed baseline: `ce316421cf4384a1cdcdd3e7de33f39b5bc37a7a`.
- Read before execution: the approved `wire-current-integration` contract and
  its independent `PASS` review, the complete approved
  `wire-baseline-repair` contract, including its planner amendment, complete
  amendment review, and the latest `### DONE: wire-baseline-repair` Developer
  result.
- Pre-generation inventory consisted of the controller-installed seven
  approved source/test paths plus the pre-existing, out-of-scope
  `docs/workflow/status.md` modification.  The seven-path cross-tree SHA-256
  comparison matched byte-for-byte for five added adapter/test files.  The two
  existing `wire.go` files differ only because this newer base already contains
  Pelican wiring: `ProvidePelicanTestService` registration and the
  `PelicanTestHandler` parameter/assignment.  The repair hunks themselves are
  present: the content-moderation adapter/bind, OpenAI ops setter provider, and
  Cafe optional quota-reset provider.  No source/test file was edited by QA.
- First generation changed only the contract-allowed
  `backend/cmd/server/wire_gen.go`; post-generation inventory contains that
  file, the seven controller-installed files, and the untouched out-of-scope
  workflow status file.  No unmerged index entries exist.

## Executed local gates

All commands below ran from this QA tree's `backend` directory and use only
local package tests/fakes.  No DB, Redis, external provider, service,
application, container, deployment, commit, reset, or clean action occurred.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/service -list 'WireProvider|CompositeRoute'` | 0 | Listed all four Composite selectors and three `TestWireProvider*` selectors. |
| `go test ./internal/service -run 'WireProvider|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/handler -run WireProvider -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run 'WireProvider|CafeRoom.*Reset|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/service -run PelicanReview -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run KeyRouteFirstResponse -count=1` | 0 | Passed. |
| first `go generate ./cmd/server` | 0 | SHA-256 `AFFBBE80810E25279DDE012DAE27A7AC76A407AB5B932E7BB633BE059BB19772`. |
| second `go generate ./cmd/server` | 0 | Same SHA-256; byte-identical to first generation. |
| `go test ./cmd/server -run '^$' -count=1` | 1 | Blocked by stale direct `provideCleanup` call in `cmd/server/wire_gen_test.go:89`; see finding below. |
| `gofmt -l` over all eight contract Go paths | 0 | No output. |
| root `git diff --check` | 0 | Passed; only pre-existing workflow CRLF warning was emitted. |
| root `git ls-files -u` | 0 | No unmerged entries. |

`go build ./...` was not run after the mandatory server-package compilation
gate failed.  The contract stop rule requires stopping on a failing gate; QA
has no authority to repair or bypass it.

## Generated-diff audit

The complete first `wire_gen.go` diff is 50 lines changed (23 additions,
27 deletions).  It contains the expected repair transformations:

- `NewContentModerationService(..., proxyRepository)` becomes
  `ProvideContentModerationService` with the same eight arguments, including
  exactly one proxy repository.
- The Cafe construction becomes `admin.ProvideCafeRoomHandler` with the three
  original constructor dependencies plus `adminService`; the provider retains
  the optional `CafeQuotaResetService` setter hook.
- The OpenAI handler provider receives the existing coordinator and `opsService`;
  the provider retains `SetOpsService`, so the old generated setter is not lost.

Remaining hunks are Wire order/name changes for existing OAuth/cache, CN quota
and balance, settlement, and current-base Pelican variables.  The generated
call still supplies the same ordered cleanup dependencies; settlement and CN
balance-check construction remain before `provideCleanup`, and the Pelican
service remains both constructed and passed to cleanup.  No additional missing
hook or generated path drift was found.  This structural audit does not cure
the separate compile-gate failure below.

## Blocking finding

`cmd/server/wire_gen_test.go` directly invokes `provideCleanup` but has no
argument for its existing `*service.PelicanTestService` parameter.  The test
therefore supplies 39 arguments where the current function requires 40:

```text
cmd/server/wire_gen_test.go:89:3: not enough arguments in call to provideCleanup
have (..., *service.UsageBillingSettlementService)
want (..., *service.PelicanTestService, ..., *service.UsageBillingSettlementService)
```

This test file is tracked on the current baseline (last commit affecting it:
`6116a22d0e93119c8229034d918e3585582af180`) and is neither one of the seven
controller-transplanted repair paths nor an allowed QA output.  Editing it
would violate the contract.  The required server compile gate therefore cannot
pass without a Planner-approved owner and scope amendment.

## Result and limits

`BLOCKED`: reproducible current-base generation, focused repair regression,
and parallel Pelican/first-response local smoke pass, but the required server
compile gate fails on an out-of-scope stale test call.  The next legal action is
for Planner/owner to authorize and repair that test (or otherwise amend the
contract), then use a fresh QA run to repeat the failed compile, build, and
remaining final gates.  Real DB/admin/provider runtime and Composite P2/P3
remain unverified and out of scope.

---

## Retest after approved current-base fixture amendment — PASS

The preceding `BLOCKED` history is retained as required.  The current contract
and its fixture-amendment review are `PASS`, and
`worker-results/wire-current-fixture-result.md` is `### DONE`.  The only new
Developer diff is the independently reviewed positional fixture insertion:
`nil, // pelicanTests` immediately after `scheduledTestRunner` and before
`backupSvc` in `cmd/server/wire_gen_test.go`.  QA did not edit it or any
business/test/generated file.

### Fresh local results

All commands were rerun independently from this worktree's `backend`
directory; every process completed with its shown exit code (no live or reused
Go session remained).  No application, DB/Redis, external provider, container,
deployment, commit, reset, or clean operation occurred.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/service -list 'WireProvider|CompositeRoute'` | 0 | Listed 4 Composite and 3 WireProvider selectors. |
| `go test ./internal/service -run 'WireProvider|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/handler -run WireProvider -count=1` | 0 | Passed. |
| `go test ./internal/handler/admin -run 'WireProvider|CafeRoom.*Reset|CompositeRoute' -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run CompositeRoute -count=1` | 0 | Passed. |
| `go test ./internal/service -run PelicanReview -count=1` | 0 | Passed. |
| `go test ./internal/server/routes -run KeyRouteFirstResponse -count=1` | 0 | Passed. |
| first `go generate ./cmd/server` | 0 | SHA-256 `AFFBBE80810E25279DDE012DAE27A7AC76A407AB5B932E7BB633BE059BB19772`. |
| second `go generate ./cmd/server` | 0 | Same SHA-256; byte-identical. |
| `go test ./cmd/server -run '^$' -count=1` | 0 | Compiled successfully. |
| `go test ./cmd/server -count=1` | 0 | Passed. |
| `go build ./...` | 0 | Passed. |
| `gofmt -l` over all 9 approved Go paths | 0 | No output. |
| root `git diff --check` | 0 | Passed; only the pre-existing workflow CRLF warning appeared. |
| root `git ls-files -u` | 0 | No unmerged entries. |

### Current-base generated-diff and scope audit

The complete generated diff remains 50 lines (23 additions, 27 deletions) and
is reproducible by the matching hashes.  It preserves all eight arguments of
`ProvideContentModerationService`, including exactly one proxy repository.
`admin.ProvideCafeRoomHandler` retains the original three constructor inputs
and its optional quota-reset hook.  `handler.ProvideOpenAIGatewayHandler`
receives both the existing coordinator and `opsService`, with the prior
`SetOpsService` hook now retained inside the provider.

The other hunks are generator ordering/name movement for existing OAuth/cache,
CN quota/balance, settlement, and current-base Pelican variables.  The Pelican
service is still constructed through `ProvidePelicanTestService`, supplied to
its handler, and passed into `provideCleanup`; its existing lifecycle is not
removed.  Settlement and CN balance-check construction still precede cleanup,
with the same ordered cleanup dependency list.  No extra provider error,
missing setter hook, out-of-scope generated path, or unprovable cleanup drift
was observed.

Final inventory is exactly the generated `wire_gen.go`, the approved fixture
file, seven controller-installed repair source/test paths, and the pre-existing
out-of-scope `docs/workflow/status.md` modification.  The five untracked
adapter/test paths are among the approved seven; QA report tracking is ignored
by this worktree's Git rules.  No unrelated path was edited by QA.

### Final verdict

`PASS` for local current-base integration only.  This establishes reproducible
generation and the listed local fake/package checks; real DB, Redis, admin HTTP,
external provider, service/container runtime, deployment, and Composite P2/P3
remain unverified and out of scope.
