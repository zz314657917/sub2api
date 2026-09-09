# S295 Composite Gateway Selection Replan

## Status

`DEFERRED / planner evidence complete` as of 2026-09-08. This is a replan,
not an implementation contract and does not authorize a schema migration,
admin API, route dispatch, provider traffic, database write, deployment or push.

## Why The Previous Contract Cannot Run

The former S295 contract was based on `483de927c`, which is not an ancestor of
the inspected local `HEAD` `42ab6da53`. The requested files are therefore not
missing from the current topology because of an accidental deletion: they belong
to an unintegrated Composite feature chain.

That chain has dependencies outside the former service-only allowlist:

- `07619d491`: Ent `CompositeModelRoute` schema and migration.
- `6ba96841e`: repository, route model and resolver.
- `483de927c`: target platform and upstream-model context propagation.
- `3b925e158`: Gateway selection integration.
- Further upstream owners: Composite admin handler/routes and endpoint dispatch.

The current `GatewayService` passes `group.Platform` directly to both account
selection paths. A `PlatformComposite` group has no current resolver that turns
its public model into a concrete account platform, so restoring only the old
selection hunk would either be uncompilable or make routing semantics implicit.
The preserved behavior must remain fail-closed until an explicit route or a
non-ambiguous ownership decision exists.

## Proposed Delivery Sequence

### S295-P0: Contract And Data-Model Decision

- Confirm that this branch may add the Composite route schema and a forward-only
  migration, and decide whether route management belongs in the existing group
  admin APIs.
- Freeze current group/model-allowlist behavior and all protected dirty paths.
- Define deletion behavior, supported endpoint names, route precedence and
  fallback rules. Unknown or ambiguous public models must return no decision.
- No business implementation in this phase.

### S295-P1: Persistence And Route Administration

- Add the Ent schema, generated ownership only where the repository already
  expects generated code, migration, repository, group cleanup behavior and
  authenticated admin CRUD/preview API.
- Include migration rollback/compatibility guidance but do not run it against a
  shared database.
- Acceptance requires isolated PostgreSQL migration and CRUD integration tests.

### S295-P2: Resolver And Account Ownership

- Add pure route matching and model-platform detection with tests for exact vs
  prefix, endpoint specificity, priority, ambiguous aliases, and unknown models.
- Resolve exact account `model_mapping` ownership through the existing model-list
  cache; cache invalidation must be group-scoped.
- The resolver may produce a decision but must not yet rewrite a request body.

### S295-P3: Gateway Selection And Dispatch

- Inject the resolver without changing `NewGatewayService` positional callers.
- Adapt direct and load-aware account selection to use the resolved concrete
  platform and upstream model while preserving pinned accounts, group fallback,
  session affinity, channel restrictions and rate-limit selection.
- Propagate the decision through supported Gateway endpoint owners, then prove
  public model, concrete upstream model and usage/billing attribution stay
  consistent. Split endpoint families into separate contracts if their owners
  overlap or their request-shape handling differs.

### S295-P4: Runtime Acceptance

- On dedicated PostgreSQL and Redis resources, apply the migration, create a
  route via the authenticated test API, and verify two concrete providers plus
  an unknown/ambiguous-model rejection.
- Use synthetic upstreams and test accounts. Provider credentials, shared data,
  container replacement, deployment and push remain outside this phase.

## Current Decision

Do not revive the old S295 contract and do not hand-port its service files.
The first implementation phase requires an explicit authorization for the
database/schema and admin-API scope in S295-P1. Until then, S295 remains
deferred; the current S293 and S296/S297 runtime gates stay independent.

## Evidence

- Local branch: `42ab6da53`; `git merge-base --is-ancestor 483de927 HEAD`
  returns nonzero.
- Upstream feature chain is present under local `upstream/main`, whose current
  tip is `b7dba6267`.
- Current selection owners:
  `backend/internal/service/gateway_service.go` methods
  `SelectAccountForModelWithExclusions`, `SelectAccountWithLoadAwareness` and
  `resolvePlatform`.
- Existing recovery evidence:
  `docs/workflow/qa-reports/plan-recovery-s293-s295-20260908.md`.
