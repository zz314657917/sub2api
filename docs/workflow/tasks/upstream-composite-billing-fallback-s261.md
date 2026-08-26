# Upstream Composite Billing Fallback S261

## Task ID

upstream-composite-billing-fallback-s261

## Role

Planner/Generator/Evaluator: Codex defines, reviews, implements, and verifies
this bounded backend correctness port in the protected primary worktree. No
worker dispatch is needed for this small, self-contained billing slice.

## Goal

Adapt upstream `ba88cc239` so a composite-group public alias without an
explicit local price is billed by the concrete forwarded model, rather than
silently recording zero cost or receiving an unrelated family fallback price.

## Success Criteria

- When `BillingModelSource` selects a public alias in a composite group and no
  explicit group/channel price exists for that alias, billing uses the concrete
  forwarded model.
- Explicit administrator pricing for that alias continues to use the alias.
  Local group-level pricing is included here because it is an existing,
  more-specific administrator pricing layer; channel pricing remains included.
- Outside composite groups, an unresolvable selected billing model falls back
  to a resolvable concrete forwarded candidate; existing resolvable models and
  fully unpriced requests retain their current behavior.
- The usage-log requested/upstream-model attribution, image/per-request
  selection, account statistics, schema, API, channel mappings, and existing
  prices are unchanged.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: `main@51dbb2f61`, clean apart from protected untracked `outputs/**`.
- Upstream provenance: `ba88cc239` (`fix(billing): bill composite alias
  requests by the concrete forwarded model`). Its current owner file was later
  split upstream; the local equivalent remains
  `backend/internal/service/gateway_service.go`.
- The local resolver additionally has `PricingSourceGroup`, which represents an
  explicit administrator price and must be treated like explicit channel
  pricing for this fallback decision.

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_composite_billing_fallback_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-composite-billing-fallback-s261.md`
- `knowledge/tasks/current-task.md` only after final verification

## Denied Paths

- `backend/internal/handler/grok_media.go` (the upstream attribution-only
  follow-up is independent of this billing correction).
- Channel/group configuration, model mapping, resolver architecture, price
  tables, schemas/migrations, API contracts, frontend, dependencies, provider
  calls, containers, deployment, shared data, staging/push, and `outputs/**`.

## Constraints

- Hand-port behavior; do not cherry-pick divergent upstream history.
- Determine the concrete model before applying `BillingModelSource` overrides.
- Reuse the existing resolver and `BillingService.GetModelPricing`; do not add
  another pricing lookup path or alter family-fallback pricing behavior.
- Apply the composite alias guard before the general resolvability fallback so
  an alias such as `all/claude` cannot be mistaken for a valid Sonnet-family
  price.
- Keep edits minimal and do not repair unrelated unit-suite drift in this
  Sprint.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -run 'Test(CompositeBillableModel|BillableModelWithFallback|HasResolvableTokenPricing)$' -count=1
go test ./internal/service -run '^$'
gofmt -w internal/service/gateway_service.go internal/service/gateway_composite_billing_fallback_test.go

Set-Location F:/mcplugins/sub2api
git diff --check -- backend/internal/service/gateway_service.go backend/internal/service/gateway_composite_billing_fallback_test.go docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-composite-billing-fallback-s261.md knowledge/tasks/current-task.md
git ls-files -u
git diff --cached --name-only
```

The tagged legacy suite is an additional observation only: its existing
compile drift is outside this contract and must not be repaired here.

## Output

- Final evaluation is `PASS`, `FAIL`, or `BLOCKED`, with upstream provenance,
  changed files, actual command output, scope/index review, and residual risk.
- A local implementation commit is allowed only after the acceptance gates
  pass; normal push remains a separate user-authorized action.

## Stop Rules

- Stop if preserving explicit group/channel alias pricing requires changing
  resolver, channel, or group write behavior.
- Stop rather than changing stored usage records, account statistics, source
  attribution, or image/video billing semantics to fit this port.
- Stop rather than touching `outputs/**` or unrelated user work.

## Budget

- controller_implementation: local Codex
- worker_dispatch: disabled for this focused task
- qa_mode: runtime
