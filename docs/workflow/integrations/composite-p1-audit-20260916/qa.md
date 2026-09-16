### BLOCKED: composite-p1-audit

# Composite P1 independent local QA

## Scope and baseline

- Audit worktree: `E:/codex-worktrees/sub2api/composite-p1-qa`.
- Contract review is `approved` / `PASS`; this report is a fresh QA result, not
  a reuse of the prior QA report.
- Pre-generation inventory found two pre-existing workflow baseline changes:
  `M docs/workflow/main-log.md` and `M docs/workflow/status.md`.  Neither was
  edited or absorbed by this task.  There were no untracked files or index
  conflicts.  `git diff --check` exited 0 (with only the existing-file CRLF
  warnings) and `git ls-files -u` exited 0 with no output.
- No database, Redis, provider, migration, container, service, deployment,
  commit, push, reset, or manual business/test/generated-code edit was made.

## Original success-criteria evidence map

| Criterion | Static/source evidence | Executed evidence and boundary |
| --- | --- | --- |
| Ent persistence and forward migration | `ent/schema/composite_model_route.go` defines the Composite route fields/default priority and soft-delete mixin; `migrations/241_composite_model_routes.sql` defines the table, partial active uniqueness index, active indexes and the eight concrete target platforms. | Migration was inspected only, never executed. PostgreSQL constraints/indexes remain runtime-unverified. |
| Composite-only CRUD and ownership | `internal/service/composite_route_admin_service.go` calls `requireCompositeGroup` for each operation and checks route membership before update/delete. `composite_route_admin_service_test.go` covers non-composite rejection and cross-group route rejection. | `go test ./internal/service -run CompositeRoute -count=1` exited 0. This is unit evidence, not database persistence proof. |
| Group deletion transaction/route soft deletion | `internal/repository/group_repo.go:719-803` creates an Ent transaction, soft-deletes `composite_model_routes` at line 795, then soft-deletes the group before commit. | Source inspected only; no database transaction was run. |
| Ordered enabled preview and fail-closed candidate match | `composite_model_route_repo.go:19-28` orders by `priority`, then `id`, and filters `enabled`; `composite_route_admin_service.go` filters static matches without resolver/account access. Service test proves enabled/static filtering and an unknown model returns an empty candidate list. | Service focused test exited 0. It does not prove a live HTTP/database response. |
| Unknown input distinction | An unknown model is syntactically valid and produces `candidates: []` through `compositeRouteMatches`, matching the contract's no-guess/no-candidate behavior. An unknown endpoint is rejected by `normalizeCompositeRoutePreviewRequest` / `isCompositeRouteEndpoint` as `INVALID_COMPOSITE_ROUTE`; handler error mapping makes this a 400 invalid-input path, not a no-candidate lookup. | This is source/unit-level evidence only; no authenticated HTTP server was started. |
| Input validation | Public model, match type, endpoint and concrete target platform are validated in `validateCompositeRouteInput`; migration also checks match type/endpoint/target platform. Priority has no contract-defined range: schema is signed `Int`, SQL is `INTEGER NOT NULL DEFAULT 100` with no priority CHECK, and implementation normalizes only `0` to `100`. | The focused service tests cover default normalization, but do not establish an intended negative-priority policy. Therefore negative priority acceptance is not a demonstrated P1 product defect; the intended range is an open specification clarification. |
| Admin API/auth/audit registration | `RegisterAdminRoutes` applies existing `adminAuth` then `auditLog` to the `/admin` group; `registerGroupRoutes` registers all five Composite endpoints beneath it. Route test checks route presence. | `go test ./internal/server/routes -run CompositeRoute -count=1` exited 0. Static registration is not authenticated/audited runtime proof. |
| `composite` group validation and handler binding | Create/update request bindings include `composite`; handler-focused tests cover request mapping/default enabled and invalid IDs. | `go test ./internal/handler/admin -run CompositeRoute -count=1` exited 0. |

## Executed local gates

All commands below ran from `E:/codex-worktrees/sub2api/composite-p1-qa/backend`, except Git inventory commands which ran at the worktree root.

- `gofmt -l <original Go allowlist>`: exit 0, no output.
- `go test ./internal/service -list CompositeRoute`: exit 0; selected `TestCompositeRouteAdminRejectsNonCompositeGroups`, `TestCompositeRouteAdminCreateNormalizesDefaults`, `TestCompositeRouteAdminPreviewFiltersStaticCandidates`, `TestCompositeRouteAdminRejectsCrossGroupRouteAndPropagatesLookupErrors`.
- `go test ./internal/handler/admin -list CompositeRoute`: exit 0; selected `TestGroupHandlerCompositeRouteDefaultsEnabledAndMapsPreview`, `TestGroupHandlerCompositeRouteRejectsInvalidIDs`, `TestGroupHandlerCompositeRouteAcceptsCompositeGroupPlatform`.
- `go test ./internal/server/routes -list CompositeRoute`: exit 0; selected `TestCompositeRouteAdminRoutesRegisteredUnderAdminGroups`.
- `go test ./internal/service -run CompositeRoute -count=1`: exit 0.
- `go test ./internal/handler/admin -run CompositeRoute -count=1`: exit 0.
- `go test ./internal/server/routes -run CompositeRoute -count=1`: exit 0.
- `go test ./cmd/server -run '^$' -count=1`: exit 0 (`[no tests to run]`; this is the requested compilation gate, not a test PASS claim).
- `go build ./...`: exit 0, before generation.

## Generator reproducibility and stop result

1. Pre-generation inventory was the two workflow baseline paths above.  The Ent generated-file allowlist was taken exactly from the original P1 contract.
2. `go generate ./ent`: exit 0.  Post-command `git status --short`, tracked-diff inventory and untracked inventory still showed only the same two baseline workflow paths.  Thus Ent generation created no drift, including no allowlist-external path.
3. Because the Ent audit was clean, `go generate ./cmd/server` was attempted.  It exited 1 at `backend/cmd/server/wire.go:34:1`, `inject initializeApplication`:
   - no provider for `[]github.com/Wei-Shaw/sub2api/internal/service.ProxyRepository`, needed through `ContentModerationService` and `ContentModerationHandler`;
   - no provider for `github.com/Wei-Shaw/sub2api/internal/service.newUserTrialConsumer`, needed through `UsageBillingSettlementService`.
4. The failure left no additional tracked or untracked path: post-failure inventory still contained only `docs/workflow/main-log.md` and `docs/workflow/status.md`; `git diff --check` and `git ls-files -u` both exited 0.  No repair, reset, retry, or post-generation rebuild was performed.

## Verdict and remaining gates

`BLOCKED`: all focused local test/build gates and the Ent reproducibility audit passed, but the contract requires the `cmd/server` generator gate and it fails on the two pre-existing non-P1 Wire providers above. This QA task may not fix those dependencies.

This is not release approval. No migration, PostgreSQL transaction/constraint behavior, authenticated/admin-audit HTTP path, provider interaction, resolver/account ownership, request context, Gateway selection or dispatch was run. P2/P3 remain outside this audit.
