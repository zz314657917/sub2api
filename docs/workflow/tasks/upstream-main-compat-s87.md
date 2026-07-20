# Task Contract: upstream-main-compat-s87

## Task ID

`upstream-main-compat-s87`

## Status

`approved`

## Role

Direct Codex behavior-level port of three low-risk fixes from upstream
`v0.1.162`. Preserve the local API-key routing and table layout topology; do
not merge upstream history.

## Goal

Port these behavior slices from upstream snapshot
`e625ce3b3b3b955b7c3afc93221f7c5f0ae55aa8` onto local baseline
`3418267b324412071b6b8ec1ae378df2e6372f2a`:

- `e50276617`: preserve API-key IP allow/deny lists when an update omits them.
- `604b764cf`: emit OpenAI-compatible `insufficient_quota` errors on the local
  Responses roots while preserving existing protocol-specific behavior.
- `4c73149af`: restore the Available Channels table's vertical scroll chain.

## Success Criteria

- API-key updates distinguish an omitted IP list (`nil`, preserve existing)
  or JSON `null` (preserve existing) from an explicit empty JSON array (clear
  the list). Explicit non-empty lists are still validated and replace the
  existing list; invalid input must fail before repository update.
- Both handler and service update DTOs preserve that three-state contract, and
  the handler passes the pointers through without synthesizing empty slices.
- Exhausted API-key quota returns HTTP 429 with OpenAI error fields
  `type=insufficient_quota`, `code=insufficient_quota`, `param=null` on the
  local Responses endpoint and its existing aliases. Anthropic Messages,
  Gemini native, usage/admin, and other non-Responses paths keep
  `API_KEY_QUOTA_EXHAUSTED` (or their existing protocol-specific behavior).
- Quota classification reuses the local model-aware endpoint guard, then
  applies an exact Responses root/subpath check so paths such as
  `/v1/responsesx` are not reclassified. The local roots are `/v1/responses`,
  `/responses`, and `/backend-api/codex/responses`.
- `AvailableChannelsTable` mounts its root table on `.table-wrapper` and does
  not add an inner `card overflow-hidden` wrapper.
- The subscription-plan currency-symbol fix is explicitly deferred because its
  required upstream currency field/schema/API prerequisite is absent locally.
- Focused Go/Vitest tests, S85 failover billing regressions, frontend typecheck,
  production build, exact path/conflict/diff gates, and protected primary
  checkout hashes pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-main-compat-s87`
- Branch: `codex/upstream-main-compat-s87`
- Baseline: `3418267b324412071b6b8ec1ae378df2e6372f2a`
- Upstream snapshot: `e625ce3b3b3b955b7c3afc93221f7c5f0ae55aa8`
- Upstream release tag: `v0.1.162` at `27f094e09`
- Previous audited upstream baseline: `db4295d646`
- S85 prerequisite is already present as local commit `24ade9b71`; S87 must
  re-run its failover tests but must not edit S85 business paths.

## Allowed Paths

- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/api_key_update_s87_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_update_s87_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s87_test.go`
- `backend/internal/server/middleware/middleware.go`
- `frontend/src/components/channels/AvailableChannelsTable.vue`
- `frontend/src/components/channels/__tests__/AvailableChannelsTable.spec.ts`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s87.md`
- `docs/workflow/worker-results/upstream-main-compat-s87-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s87-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, repositories, billing calculation,
  scheduler, account selection, failover state, payment execution/refunds,
  generated code, dependencies, lockfiles, and VERSION.
- API-key create semantics, multi-group routing, quota arithmetic, rate-limit
  enforcement, subscription enforcement, or protocol error writers outside the
  exhausted API-key quota response.
- Frontend payment flows, currency configuration, column settings, table
  layout components, public pages, branding/assets, and all frontend paths not
  explicitly allowlisted.
- `deploy/**`, Docker/Compose, container operations, production configuration,
  `knowledge/**`, global memories, handoff/timeline files, and all other
  upstream `v0.1.162` candidates.

## Constraints

- Work only in the isolated S87 worktree.
- Adapt upstream behavior to the local multi-group topology; do not cherry-pick
  the upstream commits.
- Preserve API-key create request slice types; only update DTOs are nullable.
- Do not make Anthropic Messages, Gemini native, Chat Completions, Images, or
  other non-Responses endpoints return an OpenAI quota envelope.
- Do not change `isModelAwareBillingEndpoint` behavior or add routes. The
  existing Responses roots are recognized only for quota classification.
- Keep Available Channels styling changes limited to the root wrapper class;
  no table design or column changes.
- New contract/result/QA files match `docs/*` ignore and must be force-added by
  exact path only.
- Do not push, deploy, update containers, or merge S87 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s87'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build-s87'
New-Item -ItemType Directory -Force -Path $env:GOCACHE,$env:GOTMPDIR | Out-Null
$found = @(go test ./internal/handler ./internal/service ./internal/server/middleware -list '^TestS87' 2>$null)
@('TestS87APIKeyUpdateJSONPresence','TestS87APIKeyIPRestrictions','TestS87APIKeyQuotaErrorEnvelope','TestS87APIKeyQuotaErrorPathMatrix') | ForEach-Object {
  if ($found -notcontains $_) { throw "Missing S87 test: $_" }
}
go test ./internal/handler -run '^TestS87APIKeyUpdateJSONPresence$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S87 API-key handler test failed' }
go test ./internal/service -run '^TestS87APIKeyIPRestrictions$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S87 API-key service test failed' }
go test ./internal/server/middleware -run '^TestS87APIKeyQuotaError' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S87 quota middleware tests failed' }
go test ./internal/handler -run '^TestHandleFailoverError_' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S85 failover billing regression failed' }
Pop-Location

Push-Location frontend
npm.cmd run test:run -- src/components/channels/__tests__/AvailableChannelsTable.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S87 focused frontend tests failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S87 frontend typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S87 frontend build failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S87 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) { throw 'S87 has unmerged index entries' }
```

Evaluator additionally audits every changed path against the allowlist, scans
real conflict markers, confirms all named tests are discovered, checks the
Responses/Anthropic/Gemini/non-Responses quota-path matrix, and verifies that
`backend/internal/handler/failover_loop.go` has no S87 diff.

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s87-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s87-qa.md`
- Workflow status/log entries for contract review, implementation, QA, and
  final verdict.

## Stop Rules

- Stop if implementation requires a denied path, migration, dependency update,
  payment execution change, account routing change, or product-level decision.
- Stop if omitted IP arrays still clear existing values, explicit empty arrays
  cannot clear, or invalid non-empty lists bypass validation.
- Stop if Anthropic/Gemini paths receive OpenAI quota envelopes, or if any
  non-quota auth error changes shape.
- Stop if any deferred plan-currency, migration, Ent, or payment path is
  touched.
- Stop if S85 failover regressions fail or an S85 business path changes.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation after unavailable external reviewers`
- qa_mode: `fresh focused Go/Vitest tests plus evidence-first diff review`
- worktree_root: `E:/codex-worktrees`
