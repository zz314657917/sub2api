---
phase: contract-approved
current_sprint: upstream-main-ws-cap-s302
total_sprints: 300
pending_action: S302 contract approved; create isolated worktree and implement, then independent QA.
project_type: fullstack
qa_mode: runtime
approval_required: true
last_verified: 2026-09-12
---

# Current Sprint: cf3577a3c Behavior Port (2026-09-12)

- Reimplemented upstream `cf3577a3c5367afcda9e02bd08ba1f077074c7f4` on
  baseline `c1833f66c5ca5777ae0d11c0645d5e433ba1502e` using the existing local
  monolithic gateway; no wholesale split-file import or history rewrite.
- Coverage: 41 upstream paths, 39 implemented/equivalent and 2 not applicable
  because the local alpha/search route does not exist. Local Astra pro,
  allowed_tools, continuation IDs, account controls and frontend are preserved.
- Independent Terra contract review and final QA: PASS. Final build, actual
  production-source focused service tests, handler tests, response headers,
  formatting and scoped diff checks passed. Controller also reran the original
  item-ID regression tests. Full tagged unit fixtures have baseline compile
  drift; this is not a claim that the full tagged suite passes.
- Review caught and fixed whitespace-sensitive null schema normalization and
  cross-turn WebSocket alias collisions; published alias maps are copy-on-write.
- Evidence: `docs/workflow/tasks/cf3577a3c-coverage.md` and
  `docs/workflow/qa-reports/cf3577a3c-behavior-port-qa.md`.
- Real provider, database, container and deployment are unverified/out of scope.

# Previous Sprint: Upstream v0.2.4 Image 2.5 OAuth S299 (2026-09-10)

- User approved continuing integration and explicitly authorized
  `gpt-5.6-sol` when the default Terra route is unavailable. Contract and
  independent review are approved for a behavior-level adaptation of upstream
  `7ccc8a6f5`; no merge, rebase or cherry-pick is allowed.
- Official OpenAI documentation currently lists
  `gpt-image-2.5-flare` and `gpt-image-2.5-sunburst` for Image API and the
  Responses image-generation tool. Official per-million-token rates match the
  upstream patch: text input/cached `$5/$1.25`, image input/cached `$8/$2`, and
  image output `$30`.
- Port: model catalogs and whitelist, configurable OAuth Responses driver with
  `gpt-5.6-luna` default, rejected-driver error attribution, OAuth test picker,
  official fallback/static prices and compose environment pass-through.
- Equivalent locally: image input-token parsing and billing fields already
  exist and must not be duplicated. Skip: README-only release prose and any
  unrelated image, billing or deployment changes.
- Contract: `docs/workflow/tasks/upstream-v024-image25-oauth-s299.md`.
  Review: `docs/workflow/contract-reviews/upstream-v024-image25-oauth-s299-review.md` (`PASS`).
- Generator completed the approved local adaptation. Initial independent QA
  found that the local generic rate-limit path did not cool a requested image
  model for the Codex plan-gated 400 response. The Images owner was amended to
  classify that exact response without changing the shared rate-limit service.
- Final independent Sol QA is `PASS`: a rejected Responses driver returns the
  actionable upstream error with zero model/temp cooldown writes, while a
  rejected requested image model writes exactly one approximately 30-minute
  model cooldown and returns `UpstreamFailoverError`. Focused backend tests,
  `go build ./...`, whitelist Vitest 15/15, frontend typecheck, four Compose
  config parses, JSON parse, gofmt, exact diff and conflict checks all passed.
- QA: `docs/workflow/qa-reports/upstream-v024-image25-oauth-s299-qa.md`.
  Real provider availability, container startup, deployment and push remain
  out of scope and unverified.

# Current Sprint: Upstream v0.2.4 Payment Fulfillment Isolation S300 (2026-09-11)

- Contract and review: `docs/workflow/tasks/upstream-v024-payment-fulfillment-isolation-s300.md` and `docs/workflow/contract-reviews/upstream-v024-payment-fulfillment-isolation-s300-review.md` (`PASS`).
- Adapted upstream `7a70de401` without merge/rebase/cherry-pick. Public redeem rate limiting remains enforced; trusted payment/admin fulfillment bypasses only the public failure counter. Payment retries now fail closed on lookup errors and validate existing code ownership/type/amount/status before use.
- Independent Sol QA `PASS`: focused service/admin tests, `go build ./...`, gofmt, exact diff and conflict checks passed. Real payment/database/deployment remain unverified.
- QA: `docs/workflow/qa-reports/upstream-v024-payment-fulfillment-isolation-s300-qa.md`.

# Previous Sprint: Upstream v0.2.4 Selective Reliability S298

- Final local-integration verdict `PASS`; committed as `a6993d902`. Sol QA
  passed focused service regression, Go build, UsageView 31/31, frontend
  typecheck, exact diff, gofmt and conflict checks. Real provider/deployment
  remained out of scope.

# Historical Detail: Upstream v0.2.4 Selective Reliability S298 (2026-09-10)

- User approved continuing selective upstream integration after the refreshed
  `upstream/main` audit (`98d86915b`, v0.2.4). This sprint adapts only two
  behaviorally independent fixes: paginating the user Usage API-key filter
  beyond the first 100 records (`dc6b318c3`) and recording diagnostics when an
  already-started OpenAI Responses stream emits `response.failed` (`6aabbdf54`).
- Contract: `docs/workflow/tasks/upstream-v024-selective-reliability-s298.md`.
  Contract review: `docs/workflow/contract-reviews/upstream-v024-selective-reliability-s298-review.md` (`PASS`).
- Build completed by controller takeover after the required Terra Generator
  returned `403 No available group route` before consuming tokens or touching
  code. The four allowed code/test files contain only the approved two
  behaviors. Controller checks passed: focused service regression, UsageView
  Vitest 31/31, `go build ./...`, frontend typecheck and exact-path diff check.
- User authorized `gpt-5.6-sol` as the alternate independent QA model after
  Terra returned 403 twice. Sol QA passed the contract acceptance commands,
  exact four-file diff and owner review. Final Evaluator verdict is `PASS` for
  local integration; real provider, container, deployment and push remain out
  of scope and unverified.
- Existing S297 remains `BLOCKED` only for absent runtime/API evidence. Its
  status is not a product-code failure and does not authorize rewriting or
  claiming completion of S296/S297. S293 R4 and S295 also remain independently
  blocked/deferred as recorded below.
- Image 2.5, payment-fulfillment isolation, and backup/migration locking are
  intentionally separate future contracts because they touch model catalogs,
  money flow, or database concurrency boundaries.

# Current Sprint: S297 Acceptance Recovery (2026-09-08)

- Current inspected HEAD: `42ab6da534c2a26c4d30176f9228f76e51bae839`.
- S296 backend/frontend commits `16453c445` / `e9120ee04` and S297
  aliases/pricing/manifest/frontend commits `88e18f157`, `5e57a2722`,
  `afd05ad49`, `22c83a31b` exist after earlier reverted implementations.
  Independent Terra QA passed focused backend tests (S296: 25, S297: 28),
  frontend tests (S296: 25, S297: 28), server compilation, backend build,
  frontend typecheck and production build. Reports are
  `qa-reports/upstream-v0200-codex-ultrafast-s296-qa.md` and
  `qa-reports/upstream-v0200-gpt6-astra-s297-qa.md`.
  Code checks pass; overall acceptance remains `BLOCKED` without runtime/API
  evidence. Tests include the pre-existing dirty S297 test assertion and do not
  establish a clean committed-tree result. S296's stale format command was
  corrected without changing implementation scope; independent amendment review,
  the exact 15-path format check and the omitted handler/admin gate now pass.
  The handler package compiled but contained no matching tests.
- S295 is `DEFERRED / replan complete`: `483de927c` and its Composite files
  belong to an unintegrated feature chain, not a deletion from this branch.
  The chain requires an Ent schema/migration, repository, admin API, resolver,
  context propagation and gateway dispatch before selection can be safe. The
  old service-only contract remains `BLOCKED`; see
  `docs/workflow/plans/upstream-v0200-composite-gateway-selection-s295-replan.md`.
- S293 R4 is still `BLOCKED`: no dedicated `PROMPT_AUDIT_TEST_POSTGRES_DSN`
  is set; Docker Linux engine pipe is unavailable; no local 5432/6379 listener
  was found. Ollama now lists `qwen3guard-gen:0.6b-q4km`, but model availability
  does not prove Guard output correctness or application integration. No
  database, provider, container or browser action was performed.
- Cached `origin/main...HEAD` is `0 51`; no remote refresh or push performed.
  Existing business/lockfile changes and untracked artifacts remain protected.

# Previous Sprint: Claude Fable 5.1 S294

- `CONTRACT APPROVED`: upstream `b3f796972`, `32ac921f2` and `34b8bf1a6`
  are adapted behavior-first to the local consolidated gateway/rate-limit
  topology. Model catalogs, OAuth prompt compatibility, Fable 7d_oi
  model-level limiting, passive usage and frontend display are in scope.
- Migration 232, channel `cache_write_1h_price`, the absent account scheduling
  threshold framework, provider smoke, deployment, push and existing dirty
  paths are denied/deferred.
- Contract: `docs/workflow/tasks/upstream-v0200-claude-fable-5-1-s294.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0200-claude-fable-5-1-s294-review.md` (`PASS`).
- `BUILD DONE / FOCUSED PASS`: Fable catalogs, mappings, pricing, OAuth prompt shape, 7d_oi model scope, passive usage and frontend display are implemented; focused backend tests and frontend Vitest 47/47 pass. Antigravity Claude usage aggregation now includes `claude-fable-5-1`.
- `INDEPENDENT QA UPDATE`: the previously recorded `mustJSON` and Prompt
  Audit `owasp_tags` type errors are resolved; `go build ./...`, frontend
  `typecheck` and production build now pass. `go test -tags unit ./...`
  retains the known repository fixture/test drift. No Fable code defect was
  found in focused evidence.
- Worker result: `docs/workflow/worker-results/upstream-v0200-claude-fable-5-1-s294-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-v0200-claude-fable-5-1-s294-qa.md` (`BLOCKED`).
- Local product commits: catalogs/usage `74a5d2c35`; pricing/rate-limit `a92afcde3`.
- Provider smoke, database, container, deployment, push and upstream history integration remain denied/deferred.

# Prompt Audit Policy Matrix S293

- `CONTINUATION / CODE CHECKS PASS, RELEASE BLOCKED` (2026-09-05): strict draft CAS,
  monotonic snapshot installation, parser-based active/candidate comparisons,
  editor lifecycle/version continuity and final history/attribution fixes have
  focused regression and independent review evidence. Frontend 43/43, backend
  focused checks and builds pass. PostgreSQL/Redis/Guard and real browser acceptance
  remain unverified; see `qa-reports/prompt-audit-policy-matrix-s293-remediation-continuation-qa.md`.

- `REMEDIATION CODE COMMITTED`: R1 core fixes, R2 draft-base CAS and R3 editor state
  fixes are locally committed as `3281801e1`. Previous local-acceptance text below is historical
  and must not be read as current release approval. R4 runtime acceptance remains
  pending; real PostgreSQL/Redis/Qwen3Guard evidence is unavailable.

- `HISTORICAL / superseded`: S293-A through S293-D were previously recorded as
  implemented locally. The policy matrix supports scoped matching, monotonic
  escalation, explainable `matched_rule_id`/OWASP metadata, bounded history,
  CAS drafts, preview, publish, rollback, admin API/UI and copy-on-write
  shadow evaluation. Synchronous and asynchronous audit paths share the same
  evaluator; unknown Guard categories fail closed, and publication hot-reloads the atomic snapshot through the
  existing PostgreSQL/Redis path without rebuilding sub2api.
- Evidence: `docs/workflow/worker-results/prompt-audit-policy-matrix-s293b-result.md`,
  `docs/workflow/worker-results/prompt-audit-policy-matrix-s293c-result.md`,
  `docs/workflow/worker-results/prompt-audit-policy-matrix-s293d-result.md`,
  `docs/workflow/qa-reports/prompt-audit-policy-matrix-s293b-qa.md`,
  `docs/workflow/qa-reports/prompt-audit-policy-matrix-s293c-qa.md` and
  `docs/workflow/qa-reports/prompt-audit-policy-matrix-s293d-qa.md`.
- Local checks pass: securityaudit/routes/middleware/migrations tests,
  `go build ./...`, Prompt Audit Vitest 31/31, `vue-tsc --noEmit`, frontend
  production build, `git diff --check` and unmerged-index check.
- Local implementation commit: `3281801e1`; no push or deployment was performed.
- Migration `backend/migrations/239_prompt_audit_policy_explanations.sql` is
  added but not executed. Live PostgreSQL/Redis multi-instance convergence,
  authenticated browser lifecycle smoke and local Qwen3Guard `0.6b` provider
  smoke remain unverified. Runtime probes found only the local
  `qwen3-embedding:0.6b` Ollama model; PostgreSQL/Redis and Docker were
  unavailable. Task-owned Playwright smoke using local mock API verified draft
  preview, publish, rollback confirmation/toast and no horizontal overflow at
  `1440x900` and `390x844`; screenshots are under `output/playwright/`.
  The browser session, profile/daemon and Vite process were cleaned up. No
  database, provider, deployment or push action was performed.

# Prompt Audit Qwen3Guard Rules S292

- `FINAL EVALUATOR PASS / local acceptance`: optional `prompt_audit_config.rules`
  is persisted and hot-reloaded with the existing version/Redis path; monotonic
  validation prevents unsafe weakening. Both synchronous and asynchronous
  audit paths apply the same rules. Focused securityaudit, handler/admin,
  service tests and `go build ./...` pass. No frontend, migration, provider,
  container, deployment or push action occurred.

# Upstream v0.2.0 Ops Proxy Attribution S291

- `FINAL EVALUATOR PASS / local acceptance`: independent Terra QA completed
  `upstream-v0200-ops-proxy-attribution-s291-qa.md`. Its focused Gateway/Gemini/OpsUpstream
  test, complete service suite, `go build ./...`, diff/conflict checks and S291 allowlist
  audit passed. A scan of 100 non-test production event literals found `ProxyID` and
  `ProxyName` on every direct event append; the sole non-event scanner match is the JSON
  parser declaration. The protected six-file dirty diff hash remains
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`.
- `LOCAL STATE`: S291-A through S291-E remain the five local commits `2734fbbcc`,
  `fd203d8bd`, `4a692587a`, `9a03e6735` and `983b0585a`. No push, provider, database,
  container or deployment action occurred. Concurrent user dirty paths, including
  `usage_log_repo*`, remain excluded.

# Upstream v0.2.0 Group Pricing Layout S290

# Upstream v0.2.0 Ops Proxy Attribution S291-A

# Upstream v0.2.0 Ops Proxy Attribution S291-B

# Upstream v0.2.0 Ops Proxy Attribution S291-C

- `CONTRACT APPROVED`: OpenAI/Grok/WS production error call-site attribution.
- Contract: `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291c.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291c-review.md` (`PASS`).
- `BUILD DONE`: local OpenAI/Grok/WS call sites now snapshot proxy attribution;
  focused and complete service tests plus build pass. Antigravity remains separate.

- `CONTRACT APPROVED`: local Gateway/Gemini HTTP error call-site attribution.
- Contract: `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291b.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291b-review.md` (`PASS`).
- `BUILD DONE`: Gateway single-file HTTP and Gemini compatibility call sites
  now stamp event-time proxy attribution; focused tests and build pass.

- `CONTRACT APPROVED`: adapt the core event contract, legacy JSON normalization
  and queued-event bounds from upstream `e9e3c46cb`, `4c1f920d5` and
  `abc07bb07`.
- Gateway/provider call sites, protected dirty paths, frontend, dependencies,
  schema, migrations, billing, deployment, containers and `outputs/**` are
  denied. No merge, rebase or cherry-pick is permitted.
- Contract: `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291a.md`.
- Review: `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291a-review.md` (`PASS`).
- `BUILD DONE`: core Ops event attribution and queue bounds implemented in the
  approved allowlist; focused tests, full service tests, `go build ./...`,
  diff-check and unmerged-index checks pass. Gateway/provider call sites remain
  deferred to S291-B/C.

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
- `S266 LINEAGE MERGE / equivalence`: `12e52216e` merges the completed
  content-moderation evidence branch. `git cherry` confirms its product commits
  `c2cd7a0a1` and `eeed2369f` already exist on main, so the merge adds only the
  missing task/result/QA provenance and cannot replay or revert product code.
  Fresh S266-focused backend and frontend checks pass with the S290 layout tests.
- `PUBLISHED / cleanup`: `origin/main` advanced from `6050139a3` to
  `5b95e68dd`. The clean S266 PGE worktrees and branches plus merged S280/S281
  branches were removed. Dirty Pixel Cafe, tutorial, S279 staging, candidate-scan
  and S274 worktrees remain protected. Git deregistered the former S266 parent
  worktree, but its task-owned dependency directory remains on disk because the
  host refused recursive deletion; it has no Git registration or active process.

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



