# Composite P1 independent audit checkpoint

## Final verdict

BLOCKED for the complete local P1 gate. Independent Terra QA closed the missing
independent test audit and proved Ent generation reproducible, but Wire generation
still fails. No product-code or generated-code integration is part of this checkpoint.

## Evidence

- Audit base: `8497ecec71985d4903678c056639be2c7a46c93f`.
- Isolated worktree: `E:/codex-worktrees/sub2api/composite-p1-qa`.
- Independent contract review PASS; fresh `gpt-5.6-terra` QA BLOCKED.
- Focused service/handler/routes tests: 4/3/1 selected, all commands exit 0.
- Server compile gate and `go build ./...`: exit 0.
- `go generate ./ent`: exit 0, no tracked or untracked generation drift.
- `go generate ./cmd/server`: exit 1, missing `[]service.ProxyRepository`
  for ContentModerationService and `service.newUserTrialConsumer` for
  UsageBillingSettlementService. No output drift; no repair or reset performed.
- Formatting, diff whitespace and unmerged-index gates passed.
- Audit baseline contained pre-existing workflow main-log/status modifications;
  these are not business changes and are not included in this snapshot.

## Remaining decisions

Wire remediation needs a separate approved contract: the provider set and generated
server wiring are shared dirty owners in main. Candidate design is an explicit
provider adapter for the optional proxy argument and an explicit binding/provider
for the trial-consumer dependency. Preserve constructor callers and runtime behavior;
do not modify moderation, settlement or billing semantics merely to satisfy Wire.
No such implementation is authorized by the QA-only contract.

The priority range remains unspecified: current source accepts nonzero signed
integers and defaults zero to 100. This is a specification clarification, not a
demonstrated product regression.

Real PostgreSQL migration/transaction constraints, authenticated admin/audit HTTP,
provider requests and Composite P2/P3 selection/dispatch remain unverified/out of
scope. OpenCode and Firefox resource decisions remain as in the roadmap.

## Coordination log

- 2026-09-16: controller resumed approved audit and dispatched fresh Terra QA.
- 2026-09-16: QA completed; controller reviewed the criterion mapping, exact Wire
  errors, Ent no-drift result and protected-worktree boundary; local verdict BLOCKED.
- Only these integration documents and the roadmap update belong to this checkpoint.
  Main's shared status/current-task and other dirty work remain untouched.
