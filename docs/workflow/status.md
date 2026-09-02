---
phase: done
current_sprint: upstream-v0200-group-pricing-layout-s290
total_sprints: 290
pending_action: S290 product commit 7cacdbab1 and its workflow evidence are locally accepted. The user authorized mainline integration and push; preserve all unrelated dirty paths/outputs while completing the approved branch merge.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-09-02 18:29 +08:00
---

# Upstream v0.2.0 Group Pricing Layout S290

- `CONTRACT PASS / scope`: manually adapt upstream `1a33dc8cc` to the local
  six-field pricing layout. The only code owners are the two shared pricing
  components, `GroupsView`, and one focused layout test.
- `CONTRACT PASS / exclusions`: backend, schemas, migrations, APIs,
  billing, lockfile, Pixel Cafe, protected apicompat/admin-service changes and
  `outputs/**` are denied. No upstream merge or cherry-pick is permitted.
- `PLANNER EVIDENCE`: `1a33dc8cc` is not already equivalent locally; its direct
  patch fails because the local `IntervalRow` topology diverged, while the
  underlying responsive-layout defect remains reproducible in source.
- `DEFERRED`: `3510aa22b` needs absent group reasoning-policy foundations,
  migrations and Ent; `05ea883e2` needs absent schema state; Claude Fable 5.1
  is a 44-file feature. These do not enter S290.
- Contract: `docs/workflow/tasks/upstream-v0200-group-pricing-layout-s290.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0200-group-pricing-layout-s290-review.md` (`PASS`).
- `BUILD DONE / scope`: four allowed frontend files implement the local responsive layout;
  focused Vitest, typecheck, production build, diff and protected hash all pass.
- `QA CONTRACT REVISION / no business change`: local Chrome smoke reached both create and
  edit group dialogs and verified the six actual default Token-price controls at desktop
  and mobile widths without horizontal overflow. The former requirement to show a Token
  interval row in those dialogs is unreachable because both callers intentionally pass
  `hide-token-intervals=true`; retain the shared source sentinel and move its enabled-route
  browser smoke to a separate follow-up. The revised contract requires independent review
  before final QA; no code, save action, push or deployment is authorized.
- `INDEPENDENT QA PASS / revised scope`: focused Vitest (2/2), typecheck, production build,
  four-file diff/conflict and protected-hash gates pass. Task-owned Chrome evidence covers
  create/edit pricing at 1440x900 and 390x844 with six reachable default Token-price controls,
  no document/dialog overflow, cancelled forms and exact session/profile/PID/Vite cleanup.
  Enabled channel-pricing IntervalRow browser smoke is explicitly deferred; Final Evaluator
  owns the remaining local integration decision.
- `FINAL EVALUATOR PASS / local acceptance`: code, test and browser evidence agree with the
  revised contract. The four allowed frontend files were committed locally as `7cacdbab1`;
  no save action, container change, deployment or push occurred. The enabled channel-pricing
  IntervalRow browser smoke remains a separate deferred task.

## Previous Sprint: S289

- `PASS / local-integration`: Codex bootstrap normalizing behavior from upstream
  `1be69e56` and `421a83282` was committed as `6050139a3`; handler-focused x10,
  full handler, server compilation and `go build ./...` passed in independent QA.

## Previous Sprint: S287

- `PASS / local-integration`: Anthropic fallback-field sanitization from upstream
  `200b1406d` was committed as `e6845b4ea`; focused x10, complete service,
  server compilation, `go build ./...` and frontend checks passed in independent QA.

## Previous Sprint: S281--S285

- `PASS / local-integration`: the approved cooldown and 429 queue is present as
  `c886cdcac`, `f48b4b77f`, `bb3d3bca6`, `65bf61f5a` and `b686353c3`.

## Previous Sprint: S281

- `CONTRACT PASS / scope`: adapt `897faea33` without renaming the local
  `ResetQuotaUsed` interface. Its single SQL update clears account-level
  `rate_limited_at` and `rate_limit_reset_at`, checks affected rows, preserves
  all other runtime/quota state, then enqueues the existing account-changed
  event and refreshes the scheduler snapshot.
- `PASS / topology`: local `account_repo.go` keeps monthly/share-display quota
  extensions and has many interface test doubles; behavior-level adaptation is
  safer than importing the upstream method rename and unrelated test edits.
- `PASS / protection`: existing apicompat, Pixel Cafe, frontend lockfile,
  workflow/user state and `outputs/**` changes remain outside S281.
- Contract: `docs/workflow/tasks/upstream-v0185-quota-reset-cooldown-s281.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0185-quota-reset-cooldown-s281-review.md`.

## Previous Sprint: S280

- `CONTRACT PASS / scope`: adapt upstream `b1e60ba45` so the Anthropic
  compatibility Chat Completions and Responses forwarders retain pool-mode
  same-account retry after failover statuses without retrying an account that
  local rate-limit handling disabled.
- `PASS / topology`: both local owners are clean and map directly to the
  upstream behavior, but the patch must be applied manually because the wider
  branch topology diverges. Existing account retry-status and rate-limit APIs
  are reused without modification.
- `PASS / protection`: six existing dirty files and `outputs/**` are denied.
  Their controller SHA-256 manifest and aggregate dirty diff hash
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c` are frozen in the contract.
- Contract: `docs/workflow/tasks/upstream-v0185-gateway-pool-retry-s280.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0185-gateway-pool-retry-s280-review.md`.
- `RESOLVED / worker-routing`: collaboration Terra dispatch and the repository
  `Invoke-PgeWorker.ps1` Terra dispatch both failed before execution with
  `HTTP 403 NO_MATCHING_GROUP_ROUTE`; worker output recorded zero tokens and
  zero cost. No S280 business file changed. The task-owned empty worktree was
  removed after a clean-status check; user then authorized controller takeover.
- `DONE / controller-takeover`: user selected option 3 and explicitly
  authorized the final evaluator to implement S280 directly while preserving
  independent QA. The two failover blocks and one focused test now satisfy the
  approved contract; focused x10, complete service and server compile pass.
- Worker/result evidence:
  `docs/workflow/worker-results/upstream-v0185-gateway-pool-retry-s280-result.md`.
- `PASS / independent-qa`: Terra QA report is PASS; focused x10, complete
  service, server compile, gofmt, diff/conflict and protected-hash checks all
  passed.
- `PASS / local-integration`: exact five-file S280 scope committed as
  `5d4810801`; no push, provider, database, container or deployment action.
- QA evidence: `docs/workflow/qa-reports/upstream-v0185-gateway-pool-retry-s280-qa.md`.

## Previous Sprint: S279

- `DONE / scope`: upstream `9f1effd71` is adapted to the local topology.
  Handler JSON limit fields preserve omitted/null/number tri-state behavior;
  ordinary groups update only provided limits, while `room_managed` groups
  continue to force all group-level limits to unlimited.
- `PASS / dirty-owner-preservation`: controller and independent QA compared
  `admin_service.go` and `group_handler.go` against repository-external SHA-256
  snapshots. The service differs only in the target `UpdateGroup` limit block;
  the pre-existing Pixel Cafe quota-reset hunk remains unstaged and unchanged.
- `PASS / verification`: focused handler/service suites passed x10, complete
  handler/service contract commands and the staged-snapshot service run
  passed, and server compile, gofmt, diff/conflict and exact-scope checks pass.
- `PASS / local-integration`: only the handler sentinel, two focused tests,
  exact service hunk and workflow evidence were committed as `df4f4f511`.
  The approved S276--S279 queue is complete. No provider, database, container,
  deployment, shared-data or push action occurred.
- Contract: `docs/workflow/tasks/upstream-v0184-group-limit-partial-s279.md`.
- Worker: `docs/workflow/worker-results/upstream-v0184-group-limit-partial-s279-result.md`.
- QA: `docs/workflow/qa-reports/upstream-v0184-group-limit-partial-s279-qa.md`.

- Earlier workflow status was archived by pge-compact at 20260901T043512271Z.
