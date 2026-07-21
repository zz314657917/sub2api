# Task Contract: multi-group-route-fail-closed-s88

## Task ID

`multi-group-route-fail-closed-s88`

## Status

`approved`

## Role

Direct Codex implementation of a fail-closed boundary for model-aware
multi-group API-key routing.

## Goal

Prevent a text, image, video, or embedding request from falling back to an
incompatible API-key default group when no configured multi-group route
matches. Preserve compatible default-group fallback and all single-group API
key behavior.

## Success Criteria

- Model-aware multi-group routing returns no resolved key when the selected
  route set is empty and the default group has a different routing scope.
- A legacy/default group whose enabled route entries all have incompatible
  explicit model rules (`model_patterns`, `image_only`, or `text_only`) cannot
  bypass those rules through default fallback.
- A default group not represented by an enabled route remains a valid fallback
  when its routing scope and platform match the request.
- A matching configured route still wins by existing priority/weight rules.
- Single-group API keys and the pre-body request-resolution pass keep their
  existing behavior.
- Rejected model-aware fallback returns HTTP 403 with stable code
  `NO_MATCHING_GROUP_ROUTE` before subscription/balance enforcement or account
  scheduling.
- Focused service and middleware tests plus default-tag package regressions,
  compile checks, path/conflict checks, and `git diff --check` pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: `96021f068`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Relevant behavior: `backend/internal/service/api_key_routing.go` currently
  returns the original API key whenever route selection returns `nil`, even if
  its default group is incompatible with the parsed request.

## Allowed Paths

- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_routing_s88_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s88_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/multi-group-route-fail-closed-s88.md`
- `docs/workflow/worker-results/multi-group-route-fail-closed-s88-result.md`
- `docs/workflow/qa-reports/multi-group-route-fail-closed-s88-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, repositories, API-key persistence,
  group/channel persistence, billing, pricing, subscription semantics, account
  selection, failover, dependencies, lockfiles, frontend code, and VERSION.
- Channel `restrict_models` semantics, group routing-scope configuration,
  route priority/weight selection, route cooldown policy, endpoint
  registration, protocol-specific gateway error writers, deployment, Docker,
  containers, production configuration, timeline, and global memories.

## Constraints

- Treat multi-group route selection as routing plus a final fail-closed guard;
  do not turn model patterns into a new global model authorization system.
- Do not reject a compatible default group merely because it is not listed as
  an enabled route; the default-group fallback remains a supported feature.
- If the default group is listed in enabled routes with explicit model rules,
  at least one such route must match the parsed request before fallback is
  allowed.
- Do not change `ResolveForRequest` behavior before the request body is parsed.
- Do not push, deploy, update containers, or rewrite history.

## Acceptance Commands

```powershell
Push-Location backend
$found = @(go test ./internal/service ./internal/server/middleware -list '^TestS88')
@(
  'TestS88ModelAwareRouteRejectsImageDefaultForTextRequest',
  'TestS88ModelAwareRouteRejectsExplicitImageOnlyDefaultForTextRequest',
  'TestS88ModelAwareRouteKeepsCompatibleTextDefaultFallback',
  'TestS88ResolveAPIKeyForModelRequestRejectsIncompatibleDefault'
) | ForEach-Object {
  if ($found -notcontains $_) { throw "Missing S88 test: $_" }
}
go test ./internal/service -run '^TestS88ModelAwareRoute' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S88 service tests failed' }
go test ./internal/server/middleware -run '^TestS88ResolveAPIKeyForModelRequest' -count=10
if ($LASTEXITCODE -ne 0) { throw 'S88 middleware tests failed' }
go test ./internal/service ./internal/server/middleware
if ($LASTEXITCODE -ne 0) { throw 'S88 default-tag package regressions failed' }
go test ./internal/handler ./internal/server/routes -run '^$'
if ($LASTEXITCODE -ne 0) { throw 'S88 dependent package compile check failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S88 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S88 has unmerged index entries' }
```

Evaluator additionally audits every changed path against the allowlist, checks
that only model-aware multi-group fallback can return the new error, verifies
compatible fallback and matching-route regressions, and scans real conflict
markers.

## Output

- Worker result: `docs/workflow/worker-results/multi-group-route-fail-closed-s88-result.md`
- QA report: `docs/workflow/qa-reports/multi-group-route-fail-closed-s88-qa.md`
- Workflow status/log entries for contract review, implementation, QA, and
  final verdict.

## Stop Rules

- Stop if the fix requires migration, persistence, frontend, billing, channel
  pricing, account scheduling, or endpoint registration changes.
- Stop if a single-group API key changes behavior.
- Stop if a compatible default text group can no longer act as fallback.
- Stop if a matching route is rejected or route priority/weight behavior
  changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation`
- qa_mode: `fresh focused and package-level Go verification plus evidence-first diff review`
- worktree_root: `current clean primary checkout`
