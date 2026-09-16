# Wire baseline reproducibility repair proposal

Status: user confirmed scope on 2026-09-16. Isolated implementation contract
`docs/workflow/tasks/wire-baseline-repair.md` in
`E:/codex-worktrees/sub2api/wire-baseline-repair` independently approved; execution
and independent current-main QA completed PASS. Final evidence is recorded in
`docs/workflow/integrations/wire-baseline-repair-20260916/`; this proposal is not
itself an acceptance result. Two reviewed existing-hook provider declarations and
a one-line cleanup test fixture amendment were necessary to preserve behavior.

## Verified cause

Checked local HEAD `d69da80b9` and existing source on 2026-09-16.
Composite P1 independent QA proved Ent regeneration has zero drift; Wire fails
on two pre-existing dependencies. The committed generated server already passes
`proxyRepository` into ContentModerationService and `welfareService` into
UsageBillingSettlementService. Build success therefore does not prove the provider
graph can regenerate that wiring.

1. `NewContentModerationService` takes variadic `...ProxyRepository`; Wire requests
   the slice type, although repository providers expose a single ProxyRepository.
2. `ProvideUsageBillingSettlementService` takes the internal `newUserTrialConsumer`
   interface. `*WelfareService` implements its `ConsumeNewUserTrial` method, but
   the provider graph has no explicit binding for that interface.

## Proposed behavior-preserving scope

- Add a fixed-argument moderation provider adapter and register it instead of
  the variadic constructor. Forward every dependency unchanged, including proxy.
  Keep the public constructor and its existing callers unchanged.
- Register an explicit Wire binding from `*WelfareService` to
  `newUserTrialConsumer` within the service package. Do not widen or change
  settlement interfaces, billing logic, trial consumption or lifecycle behavior.
- Add focused adapter/binding regression evidence. Avoid starting workers in
  unit tests; the moderation constructor starts workers when repos are non-nil.
- Regenerate server wiring only in an isolated worktree based on the approved
  committed base. Inspect the entire resulting diff before integration.

Candidate file scope: `backend/internal/service/wire.go`, a dedicated service
provider adapter/test file if needed, and generated `backend/cmd/server/wire_gen.go`.
No schema, migration, database/provider access, container or deployment changes.

## Protected ownership

Main's service/wire.go has an existing Pelican provider addition. Main's generated
server wiring is also dirty. These changes must not be overwritten, staged or
absorbed. Do not run either generator in main. Integrate only reviewed task hunks
after independent QA, preserving the committed-vs-dirty distinction.

## Required gates after scope approval

1. Planner writes exact allowlist and immutable-base contract; independent review.
2. Terra Developer implements only approved provider wiring and regression tests.
3. Fresh Terra QA verifies forwarding, interface implementation, two consecutive
   Wire generation runs (second run no drift), server compilation and backend build.
4. Re-run original Composite service/handler/routes selectors. Check no unrelated
   dependency changes, generation changes or lost provider lifecycle hooks.
5. Any further missing provider or unexpected generated diff returns to Planner;
   no iterative expansion to unrelated repairs. Preserve evidence, do not reset.

This repair would close only the Wire local gate. Real database/admin runtime,
Composite P2/P3, OpenCode and Firefox/provider resource decisions remain separate.
