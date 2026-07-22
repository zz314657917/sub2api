---
phase: done
current_sprint: upstream-frontend-hardcoded-i18n-s103
total_sprints: 103
pending_action: S103 source-only PASS; create the next independent contract before another frontend or upstream patch
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-07-22 17:22 +08:00
---

# S103 Current Sprint

- S103 contract is approved for a partial port of upstream `3401a971a`:
  AppHeader aria-labels, CustomPageView not-found text, and the English/Chinese
  common locale keys. The payment custom-method placeholder is skipped because
  the local PaymentProviderDialog has no matching custom-method UI or field.
- Payment behavior, APIs, backend, billing, and deployment remain frozen; the
  absent payment prerequisite is recorded rather than recreated.
- Contract: `docs/workflow/tasks/upstream-frontend-hardcoded-i18n-s103.md`.
- S103 implementation is complete within the four-file applicable allowlist.
- Final Evaluator: `PASS / source-only`. Focused frontend tests pass `2 files /
  9 tests`, typecheck and production build pass, and static/allowlist gates
  pass. The payment custom-method placeholder remains skipped because its local
  prerequisite is absent; no deployment or container refresh was run.
- QA report: `docs/workflow/qa-reports/upstream-frontend-hardcoded-i18n-s103-qa.md`.

# S102 Current Sprint

- S102 contract is approved. Use green for operational, orange for degraded,
  red for failed/error, and gray for empty timeline placeholders.
- Preserve all existing status calculation, bar heights/order/count, tooltips,
  and responsive width behavior.
- Contract: `docs/workflow/tasks/channel-monitor-timeline-status-colors-s102.md`.
- Implementation uses `bg-emerald-500`, `bg-orange-500`, and `bg-red-500` for
  operational, degraded, and failed/error states respectively; gray empty
  placeholders and all timeline geometry remain unchanged.
- S102 final Evaluator: `PASS / published`. MonitorTimeline tests pass `2/2`,
  typecheck and production build pass, and a 1000px browser screenshot confirms
  green/orange/red distinction. Real backend/auth, deployment, and container
  refresh were not run. Combined S101/S102 commit `fd8719b4c` is published on
  `origin/main`.
- QA report: `docs/workflow/qa-reports/channel-monitor-timeline-status-colors-s102-qa.md`.

# S101 Current Sprint

- S101 contract is approved. The 60 monitor timeline bars must shrink within
  the card content width instead of imposing the previous 298px intrinsic
  minimum and pushing the `NOW` edge past the card.
- Preserve bar count/order, status encoding, tooltips, labels, maintenance
  behavior, card/grid sizing, monitor APIs, and data semantics.
- Contract: `docs/workflow/tasks/channel-monitor-timeline-overflow-s101.md`.
- Implementation removes the fixed 3px minimum from each bar, lets the
  timeline root shrink within the card, and adds a focused 60-bar/order/status
  regression. Focused Vitest passes `1 file / 2 tests`.
- S101 final Evaluator: `PASS / published`. MonitorTimeline plus channel
  capacity regressions pass `2 files / 9 tests`; typecheck and production build
  pass; desktop 1706px and mobile 390px browser screenshots show all 60 bars
  inside each card, with no document horizontal overflow. Real backend/auth,
  deployment, and container refresh were not run. Combined S101/S102 commit
  `fd8719b4c` is published on `origin/main`.
- QA report: `docs/workflow/qa-reports/channel-monitor-timeline-overflow-s101-qa.md`.

# S100 Current Sprint

- S100 contract is approved. Port only upstream `d0fa8c63f`: positive partial
  subscription days round up, exact days remain exact, and expired durations
  remain zero.
- The patch applies cleanly to the current local service with one context-line
  offset. Frontend, billing, persistence, migrations, deployment, and calendar
  day/timezone redesign are explicitly denied.
- Contract: `docs/workflow/tasks/upstream-subscription-days-round-up-s100.md`.
- S100 final Evaluator: `PASS / published`. Test discovery confirmed both the
  deterministic boundary test and progress DTO test; focused and broader
  progress regressions, formatting, diff, conflict, allowlist, and index gates
  pass. Commit `e567ecbb5` is published on `origin/main`. Deployment and
  container refresh were not run.

# S99 Current Sprint

- S99 is complete. Multi-group routing marks missing, expired, or suspended
  subscription groups unavailable on a request-local API Key copy, skips them,
  and continues through existing priority and model matching without deleting
  saved routes. Usage-limit and persistence errors do not trigger failover.
- Default-group fallback now requires current group/subscription eligibility;
  when every enabled route and the default group are unusable, authentication
  returns `NO_MATCHING_GROUP_ROUTE`. `/v1/models` and `/v1/model-catalog` omit
  request-scoped unavailable routes.
- Key updates can preserve existing unavailable base/route bindings but cannot
  add a newly unauthorized group. The editor restores unavailable route names
  from `route_groups`, labels subscription routes as expired, keeps them
  non-selectable, and still allows removal.
- S99 final Evaluator: `PASS / published`. Focused service, middleware,
  handler and S88/S91/S93 routing regressions pass; KeysView is `18/18`,
  typecheck and production build (`1089 modules`) pass, and formatting/diff/
  conflict/unmerged gates pass. Feature commit `34b1844ab` is published on
  `origin/main`. Authenticated browser smoke, real DB/Redis renewal smoke,
  deployment, and container refresh were not run.

# S98 Current Sprint

- S98 is complete. The API-key multi-group route editor keeps its header compact,
  preserves bounded scrolling, hides groups selected by another route, and
  disables Add Route when a row is incomplete or every available group is used.
- S98 final Evaluator: `PASS / published`. KeysView focused Vitest is `18/18`,
  typecheck and production build (`1089 modules`) pass, and diff/conflict/index
  gates pass. The S99 unavailable-route state was verified in the same frontend
  files but remains covered by the separate S99 contract and backend tests.
  The shared feature commit is `34b1844ab` on `origin/main`.

# S97 Current Sprint

- S97 is complete with a behavior-level Redis ACL username port from upstream
  `49200d474`. It is limited to local config, Redis options, setup validation,
  setup wizard fields, deployment examples, and focused tests.
- Direct upstream merge is denied because the local and upstream histories are
  structurally divergent. The implementation stayed within the S97 allowlist;
  focused Go tests, frontend typecheck/build, and path/conflict gates passed.
  Commit `8b70a55a8` is published on `origin/main`. No deployment or container
  refresh was performed.

# S93 Current Sprint

- S93 contract is approved. Administrators can select an active fallback group
  for system-created default API keys under External Access, while existing
  purpose-specific default routes remain higher-priority candidates.
- New default keys persist the fallback as their base `group_id`; system-only
  permission bypass applies to both the base group and purpose routes. Runtime
  fallback still checks active group context, platform, routing scope, and the
  S91 group-owned model rules.
- Explicit backfill updates only each user's lowest-ID non-deleted key when its
  `group_id` is null. Existing groups, secondary keys, routes, and other key
  fields remain unchanged; changed auth-cache entries are invalidated.
- S93 final Evaluator: `PASS / published`. Default-tag service and handler
  tests, PostgreSQL repository integration, SettingsView `23/23` Vitest,
  frontend typecheck, production build (`1089 modules`), diff, conflict-marker,
  and unmerged-index checks pass.
- The aggregate `go test -tags=unit ./internal/service ...` gate remains blocked
  at compile time by pre-existing test drift (`stringPtr`, billing signatures,
  and Grok runtime-block helpers). Default-tag S93 tests cover the implemented
  behavior. Feature commit `01fd0784b` is pushed to `origin/main`; no deployment
  or container refresh was performed.

# S92 Current Sprint

- S92 contract is approved. User API-key multi-group routes now expose only
  group selection, drag order, enabled state, add, and remove. The client
  writes continuous priority plus fixed compatibility defaults (`weight=1`,
  `cooldown_seconds=30`) and drops legacy scope fields on save.
- S91 group-owned model matching, backend rejection/cleanup, and the running
  local container are frozen. No backend, migration, deployment, or container
  change is authorized in S92.
- S92 implementation is complete. User routes now derive strict priority from
  drag order, reject duplicate groups, and normalize compatibility fields while
  omitting legacy model/scope fields.
- S92 final Evaluator: `PASS / source-only`. Focused user KeysView Vitest passed
  `2 files / 16 tests`; typecheck, production build (`1089 modules`), diff,
  and unmerged-index checks passed. No commit, push, deployment, or container
  refresh was performed.

# Workflow Status

- S91 contract is approved. Model matching moves to administrator-maintained
  `groups.model_match_patterns`; API-key users keep only group, priority,
  weight, cooldown, enabled, and existing operational route fields. The
  implementation must guard migration until every effective selectable group
  has a rule, then clear legacy route rules atomically and invalidate auth
  snapshots.
- S91 implementation is complete in the primary checkout. Existing S89/S90
  uncommitted frontend changes remain protected and are not part of this
  sprint's scope.
- S91 final Evaluator: `PASS / source-only`. Group-owned model rules, legacy
  API-key write rejection and cleanup, model-aware priority/weight routing,
  `/v1/models` parity, cache snapshot versioning, and guarded migration are
  implemented within the approved paths.
- S91 evidence: focused Go routing/model/API rejection/cache checks, handler and
  repository tests, Ent compile checks, and migration dry-run/block/cleanup
  tests pass. The S91 frontend Vitest set passes `3 files / 20 tests`; frontend
  typecheck and production build (`1088 modules`) pass; `git diff --check` and
  unmerged-index checks pass.
- Known baseline failures remain outside S91: full `internal/service` still
  fails the existing `group_peak_rate` timezone/multiplier cases, and the
  `admin/user` frontend aggregate still has the existing 5 failing files / 18
  tests (`ImageCreatorView`, `ChatStudioView`, `ChatImageStudioView`,
  `CanvasView`, `support-popup-settings`). No S91 test is among those failures.
- No commit, push, deployment, or container refresh was performed.

- S90 contract is approved. The API-key editor will hide the complete account-
  pool strategy field only when public settings explicitly return
  `account_share_enabled: false`; missing/loading settings remain visible for
  backwards compatibility, and hidden state must not change submitted values.
- S90 implementation is complete in the primary checkout. The field now uses
  the existing public-settings response, and focused pre-QA Vitest passed 2
  files / 18 tests; final QA gates are complete.
- S90 final Evaluator: `PASS`. Focused KeysView Vitest passed 2 files / 18
  tests, typecheck and production build (`1088 modules`) passed, and diff plus
  unmerged-index gates are clean. A hidden `private_only` edit value remains
  unchanged in the submitted update payload. No commit, push, deployment, or
  container update occurred.

- S89 final Evaluator: `PASS`. The API-key dialog now uses the existing
  extra-wide width; key settings occupy the desktop left column and the
  multi-group route editor occupies the right column from the first row. Long
  route lists have independent viewport-bounded scrolling, while sub-`lg`
  layouts remain single-column.
- S89 evidence: layout Vitest `3/3`, combined key/routing regressions `15/15`,
  frontend typecheck, production build (`1088 modules`), and diff checks pass.
  Authenticated screenshot smoke was unavailable because the in-app browser
  had no login state and Chrome control was unavailable. Model matching,
  backend, persistence, group administration, deployment, and containers are
  unchanged; no commit or push occurred.

- S88 draft contract scopes a model-aware multi-group fallback guard: an
  incompatible default group must not receive a request after all configured
  route candidates fail. Compatible default fallback and single-group behavior
  remain frozen as regressions; no push, deployment, or container update is
  authorized.
- S88 contract review: `PASS`. The approved change is limited to final
  model-aware fallback compatibility checks and a stable 403 rejection;
  pre-body routing, compatible defaults, single-group keys, persistence,
  billing, account scheduling, frontend, deployment, and containers are
  frozen.
- S88 final Evaluator: `PASS`. Incompatible default scope/platform and
  explicit model rules now fail closed with HTTP 403 code
  `NO_MATCHING_GROUP_ROUTE`; compatible fallback, matching routes,
  single-group keys, and pre-body routing pass focused regressions.
- S88 fresh evidence: 9 tests discovered, service/middleware S88 tests passed
  at `count=10`, existing request/model routing regressions and dependent
  compile checks passed. The aggregate service package retains only the known
  `group_peak_rate` global-timezone failures reproduced at clean baseline
  `96021f068`; isolated peak tests pass. No push, deployment, or container
  update occurred.
- S88 scoped commit was authorized after final PASS and prepared with an exact
  11-path staged allowlist. Source push, deployment, and container update remain
  unauthorized.

- S87 amended contract selects three low-risk `v0.1.162` behavior ports:
  API-key IP-list partial updates, OpenAI-compatible quota errors, and
  Available Channels scrolling. The plan-currency display commit `a05b87321`
  is deferred because its required local schema/API prerequisite is absent.
  S85 is already present as `24ade9b71` and is a regression gate, not a new
  implementation slice.
- S87 worktree is `E:/codex-worktrees/sub2api/upstream-main-compat-s87` from
  clean local baseline `3418267b3`; no business code, push, deployment, or
  container operation has occurred. Quota scope is limited to the three
  registered Responses roots with exact subpath matching. Backend focused QA,
  Available Channels Vitest, frontend typecheck, and production build are green.
  The frontend dependency tree was restored in the isolated worktree with
  pnpm `10.33.4` under `--frozen-lockfile`; manifests and lockfiles remain
  unchanged.

- S87 final Evaluator: `PASS`. The three approved v0.1.162 behavior ports pass
  focused backend/frontend checks, S85 failover regression, exact test
  discovery, allowlist, conflict, unmerged-index, and diff gates. The
  plan-currency candidate `a05b87321` remains deferred. No push, deployment,
  container update, or knowledge-file change occurred.

- S87 was fast-forwarded into local `main` as `628c990d7` after final PASS.
  The clean S87 worktree and fully merged local branch were removed with
  non-force Git operations. Local `main` is ahead of `origin/main`; no push,
  deployment, or container update has occurred.

- S82-S86 integration final Evaluator: `PASS`. Local Usage S82 plus upstream
  compatibility S82, S83, S84, and stacked S85-S86 are combined in four merge
  commits without a business conflict.
- Combined verification passed frontend Vitest `7 files / 55 tests`, frontend
  typecheck, production build (`1088 modules`), broader Anthropic/proxy-quality
  service regressions, and broader `TestHandleFailoverError_` handler
  regressions.
- Final integration audit passed all five source-ancestor checks, 22 business-
  blob comparisons, 18 expected artifact checks, the exact 43-path audit,
  unmerged/conflict scans, and `git diff --check`.
- Refreshed `origin/main` remains `37e0b493c` and is an ancestor of the
  integration head. Refreshed `upstream/main` is `db4295d646`; this release is
  a selective behavior port and does not merge the unreviewed upstream tail.
- Source publish completed under the user's explicit authorization: local
  `main`, `origin/main`, and `git ls-remote` all matched release head
  `7f5e02030` after push. No deployment or container update was performed.
- Local cleanup completed after a fresh clean-worktree and remote-ancestor
  audit: S76-S86 plus the S82-S86 integration worktree/branches were removed
  with non-force Git commands; only the primary `main` worktree remains.
- S82 draft contract scopes a conditional reasoning-effort suffix to the
  requested model in user/admin usage-record tables. Backend, exports,
  dashboards, column defaults, deployment, and `knowledge/**` remain denied.
- S82 contract passed review: `UsageLog.reasoning_effort` already reaches both
  target tables; the six frontend business/test paths and acceptance commands
  are sufficient, with no backend or migration work required.
- S82 implementation is complete: shared model display helper, user usage row/
  detail, admin usage row/chain, and focused tests pass. Typecheck, production
  build, diff, path, conflict, and pre-existing knowledge-diff gates are in
  final QA.
- S82 final Evaluator: `PASS`. Focused Vitest 33/33, frontend typecheck,
  production build, six-path allowlist, conflict, diff, evidence, and
  pre-existing knowledge-diff gates passed. No browser smoke, deployment, or
  container update was performed.
- S81 final Evaluator: `PASS`. Baseline, exact discovery, focused/count=10, broader subscription regression, line-level diff, helper freeze, path/conflict/diff, and protected-hash gates passed. Live database/cache integration remains unverified.
- S81 implementation is complete. Exact 11-test discovery/run, focused `count=10`, and broader subscription assignment/window regressions passed; evidence-first diff and final path gates are in progress.
- S81 contract passed evidence-first review: expired/active/suspended branches, admin note deduplication, assignment provenance, frozen purchase semantics, exact discovery, and nine-path gates are explicit and executable.
- S81 draft contract ports upstream `1db10dc55`: admin assignment renews an expired reusable subscription row and resets its term/windows/usage, while suspended subscriptions and `AssignOrExtendSubscription` behavior remain unchanged.
- S80 was fast-forwarded into local `main` as `f62f8bbce`; it was not pushed or deployed, no container was touched, and protected knowledge hashes remained unchanged.
- S80 final Evaluator: `PASS`. Six password-mode renders, six shell syntax checks, three baseline/current topology comparisons, path/diff/conflict gates, and protected hashes passed without contacting Docker daemon. Runtime Redis `CONFIG GET` remains explicitly deferred.
- S80 implementation is complete: six empty/non-empty-password command renders and three baseline/current topology comparisons passed without contacting Docker daemon. Independent diff review and fresh static Compose evaluation are in progress.
- S80 contract independently passed after adding hard gates for unmerged index entries, cached conflict markers, and an exact nine-path staged allowlist. Build is authorized only inside the isolated S80 worktree.
- S80 draft contract ports upstream `be74deae7` to the three built-in Redis Compose topologies. It changes only shell line continuations; container/volume operations, runtime smoke, auth redesign, standalone/external Redis, docs, code, migrations, and `knowledge/**` remain denied.
- S79 was fast-forwarded into local `main` as `366e590b3` after its final PASS. It was not pushed or deployed, no container was updated, and all three protected primary-checkout knowledge hashes remained unchanged.
- S79 final Evaluator: `PASS`. Fresh default-tag service/handler discovery and runs, locale Vitest 2/2, typecheck, build, path/conflict/diff audits, legacy-test preservation, and protected knowledge hashes all passed. The branch remains isolated; no push, deployment, container update, or local-main merge is authorized.
- S79 implementation is complete in the isolated worktree. Default-tag Go discovery and focused service/handler tests, locale Vitest 2/2, typecheck, and production build passed; independent code/path review and final repository gates are in progress.
- S79 contract independently passed after adding exact test-discovery gates, shared handler body-propagation proof for all three call sites, force-tracked workflow evidence, protected knowledge hashes, and explicit stop rules. Build is authorized only inside the S79 worktree.
- S79 Planner amendment: local `gateway_request_test.go` is unit-tagged, so the parser regression moved to allowlisted default-tag `gateway_request_s79_test.go`; this is a test-topology correction only.
- S79 Planner amendment 2: local `channel_monitor_checker_body_test.go` is also unit-tagged, so the monitor regression likewise moved to default-tag `channel_monitor_checker_s79_test.go`; implementation scope is unchanged.

- S79 approved contract `docs/workflow/tasks/upstream-main-compat-s79.md` is limited to four low-risk `v0.1.161` behavior ports: Antigravity paid-tier preservation, Anthropic monitor text-block extraction, Claude Code trailing `[1m]` normalization, and unit-neutral subscription-plan validity copy. Deployment/Compose, protocol state machines, routing, subscription renewal, migrations, billing, security, VERSION, and `knowledge/**` remain denied.
- S78 was freshly revalidated on 2026-07-20 (Vitest 5/5, typecheck, production build, diff check) and merged into local `main` as `e50c51274`; it was not pushed or deployed and no container was updated.

- S78 contract `docs/workflow/tasks/upstream-main-compat-s78.md` passed independent contract review. The approved scope is Stripe lazy loading/chunk isolation and missing OpenAI Mobile RT/AT labels; the Grok Codex template commit `174ea22ee` is explicitly deferred because the local UI has no matching template.
- S78 implementation is complete in isolated branch `codex/upstream-main-compat-s78` from merged S77 baseline `6113b0f5e`; focused Vitest, typecheck, production build, allowlist, and diff checks passed. No push, deployment, container update, backend, migration, billing, or knowledge change was performed.

- S77 contract `docs/workflow/tasks/upstream-main-compat-s77.md` passed independent contract review at 2026-07-16 14:00 +08:00. The approved scope is limited to WS reliability, platform-aware Grok image intent, and TablePageLayout horizontal scrolling; migration, Ent, billing, payment, scheduler, deployment, VERSION, and `knowledge/**` remain denied.
- The user explicitly authorized Codex workers to replace the unavailable `deepseek-v4-pro` worker on 2026-07-16. S77 may proceed to build with that model exception recorded; no deployment or container update is authorized.
- S77 implementation commits are `0ce58e7d6` (Grok worker), `864a760e3` (WS worker), and `279b9f18d` (root WS/Grok platform call-sites plus TablePageLayout). An independent evaluator found and scoped a passthrough malformed-JSON P1; the adapter now validates every upstream text frame and has before/after-output regression coverage. Post-fix focused QA is in progress. The full default-tag service package still has unrelated `group_peak_rate` timezone/peak assertion drift.
- S77 post-fix final evaluator: `PASS`. No Ent, migration, billing, scheduler, payment, deployment, VERSION, container, or `knowledge/**` path is authorized or changed; no push, deployment, or container update was performed.

- Current phase: `done`; S76 final evaluator PASS. Fast/Flex email search selection, Grok Composer reasoning sanitization, and platform-aware no-account diagnostics are implemented in the isolated branch.
- S76 was merged into local `main` as `bff599ee1` with validated implementation parent `e9ddb900f`; the branch has not been pushed or deployed.
- S76 evidence: `docs/workflow/worker-results/upstream-v0152-low-risk-compat-s76-result.md` and `docs/workflow/qa-reports/upstream-v0152-low-risk-compat-s76-qa.md`.
- Fresh verification passed backend service 2/2, handler 2/2, frontend selector/i18n 6/6, Settings integration 1/1, typecheck, production build, 18-path allowlist audit, and `git diff --check`.
- No authenticated browser or live Grok upstream smoke was run. The existing full `unit`-tag service suite remains non-compiling for unrelated legacy test drift; S76 uses exact-discovery default-tag tests.
- No deployment, container update, migration, billing, account-type, prompt-cache, or `knowledge/**` change was performed.
- Current phase: `done`; S71-S73 were integrated in order, merged with the latest local main, and passed fresh post-merge regression.
- Release merge `ccac358e4` has parents `f6ee836d4` (latest main) and `d101ac2d2` (S71-S73 integration/workflow closeout). The final evaluated business head before integration closeout was `5e4b08af9`.
- S71 adds trusted user-scoped Fast/Flex policy across managed HTTP, API-key/OAuth passthrough HTTP, parsed WS, passthrough WS, DTO/settings, and the admin UI. Independent review found one i18n namespace bug; fix commit `24c22c9a9` and retest closed it before integration.
- S72 adds the strict bare `gpt-5.6` Sol alias/catalog, billing candidates, frontend whitelist/preset, and OpenCode `max` variants without collapsing unknown suffixes or explicit Terra/Luna models.
- S73 adds aliased legacy `request_type` fallback only to `GetUserBreakdownStats`, preserving RequestType+Stream AND semantics, seven-column scans, ordering/LIMIT, leaderboard behavior, and ordinary inclusion of leaderboard-excluded users.
- Post-merge verification passed backend exact discovery `S71=4 / S72=6 / S73=2`, Fast/Flex HTTP/WS/DTO/service regressions, legacy request-type and leaderboard regressions, S74 ticket regressions, and repository compile-only checks.
- Post-merge frontend verification passed `6 files / 59 tests`, `vue-tsc --noEmit`, and the production build. Existing router-link, stale `caniuse-lite`, Vite import/chunk-size, and Node `DEP0190` warnings remained non-blocking.
- Final merge audit found no business-path overlap between the two parents, verified all 39 business blobs against their owning parent, confirmed all ten S71-S73 artifacts, and found no conflict markers or unmerged index entries.
- Final evidence: `docs/workflow/qa-reports/upstream-v0151-followups-s71-s73-qa.md`, the three Sprint QA reports, and the post-merge release audit. Remote push has not yet occurred at this snapshot; no deployment or container update was performed.
- All three contracts kept payment concurrency `fc66a30ff`, Cyber migration, deployment, and the main-worktree `knowledge/05-current-focus.md` change outside scope.
- Cleanup audit proved all six Generator/QA branches patch-equivalent (`git cherry` only `-` entries); their clean worktrees and local branches were removed. The integration and release worktrees remain only until remote publication is verified.

- Current phase: `done`.
- Historical S75: `admin-usage-column-menu-layer-s75`.
- Final evaluator: PASS. `UsageView` adds `z-[221]` to the `UsageFilters` root only while the column-settings menu is open, putting it just above the `DataTable` sticky-header maximum of 220 and restoring its normal `z-30` layer when closed.
- Independent QA: PASS. The S75 targeted Vitest suite (2 tests), typecheck, production build, compiled-CSS inspection, and target diff check passed; implementation remains limited to `UsageView.vue` and its test.
- Runtime caveat: no authenticated administrator browser smoke was run because Docker, PostgreSQL, and Redis are unavailable; direct Compose in this worktree would reuse existing persisted data. No deployment or container update was performed.
- Historical S74: admin ticket list/detail return only the joined user summary; the detail opens the existing read-only user-information modal, and recent usage requests 30 scrollable records. Targeted Go tests, 11 frontend tests, typecheck, production build, and target diff check passed.
- Historical S70: the administrator-controlled leaderboard exclusion was persisted, invalidated leaderboard caches, and excluded the user from rank, champion, badge, reward, model, and trend paths. S65-S70 were merged into `main` as `d6ff6a158` and pushed to `origin/main`; no deployment or container update was performed.
- Post-release cleanup removed the fully merged local S69/S70 branches and remote S70/group-buy branches; local and origin now retain only `main` for this repository.

- 当前阶段：`done`。
- 历史 S75：`admin-usage-column-menu-layer-s75`。
- 最终 Evaluator：PASS。`UsageView` 仅在列设置菜单打开时给 `UsageFilters` 根卡片追加 `z-[221]`，恰好高于 `DataTable` 固定表头最高层级 220；关闭后恢复原有 `z-30`。
- 独立 QA：PASS。S75 定向 Vitest（2 个测试）、typecheck、production build、编译 CSS 检查和目标 diff check 均通过；实现限定在 `UsageView.vue` 及其测试。
- 运行态缺口：Docker、PostgreSQL、Redis 不可用，因此未做带真实管理员会话的浏览器 smoke；直接在本工作树启动 Compose 会复用已有持久化数据。本轮未部署、未更新容器。
- 历史 S74：管理端工单列表/详情现已返回最小用户摘要；详情可打开既有只读用户信息弹窗；最近使用请求 30 条并在固定滚动区展示。定向 Go 测试、前端 11 个用例、类型检查、生产构建和目标 diff check 均通过。
- 历史 Sprint：`upstream-gpt56-priority-cache-billing-s69`。
- S69 独立 QA 报告 `docs/workflow/qa-reports/upstream-gpt56-priority-cache-billing-s69-qa.md` 首行 PASS；主控最终裁决为 DONE。GPT-5.6 Priority cache-write 专价、override、272k、Flex、breakdown 和 RecordUsage 均已闭环。S69 已随 S65-S70 集成历史合入并推送到 `main`；尚未部署。
- S69 Generator 实现已集成为 `d5a1aef0b`，worker result 收口为 `07399e50d`；独立代码审查 PASS，8 个 changed paths 全部在 contract allowlist；随后 fresh worktree 独立 QA 也已 PASS。
- S69 contract 经两项独立 Evaluator 复核通过：计费语义 PASS；验收命令在补齐 native exit-code、clean worktree 和精确测试枚举门禁后 PASS。下一合法动作是从批准 HEAD 派发单一 billing Generator worker。
- S69 draft contract 已生成：只补 GPT-5.6 Priority cache-write 价格从静态/动态/override 到结算与 RecordUsage 的完整链；裸别名、用户级 Fast/Flex、usage breakdown、Cyber migration、支付与前端明确冻结。
- 三组 S69 Planner 审计确认：协议/Ops/Grok 候选均已等价覆盖或应跳过；真正最高优先级缺口是 GPT-5.6 Priority cache-write 仍按 Standard 价格少计费。
- 独立组合 QA 报告 `docs/workflow/qa-reports/upstream-codex-imagegen-namespace-strip-s68b-qa.md` 判定 PASS；主控最终裁决为 DONE。S68 未合入 `main`、未推送、未部署。
- S66-S68 的 15 个临时 worktree/branch 已在确认工作树干净且 `git cherry HEAD <branch>` 无 unique commit 后清理；当前只保留集成工作树。
- S68b 初始实现与 fix1 已集成为 `7593079a9`、`19066c93d`；组合 diff 经主控关键验收、独立 fix1 复核和独立组合 QA PASS。
- S68b 初始 worker commit `36fe56441` 经主控关键验收 PASS，但独立 review 曾判定 FAIL：本地 `OpenAIWSIngressModePassthrough` 在 parsed ingress 前提前返回，首帧和后续帧均绕过图片声明 strip；该 finding 已由 fix1 关闭后统一集成。
- Fix1 contract 已由原 Generator 完成：WS passthrough 首帧/后续帧、invalid raw JSON 与 OAuth HTTP passthrough actual-body 证据均已补齐，独立复核确认此前 P1/P2 finding 全部关闭。
- S68b backend worker 已在独立 worktree `E:/codex-worktrees/sub2api/upstream-codex-imagegen-namespace-strip-s68b` 完成初始实现与 fix1；两笔实现均已进入当前集成分支。
- S68b draft contract 已生成：基于 S68a policy PASS，只扩展 namespace、Responses Lite、passthrough/raw 和 WS stripping；策略 UI 与 apicompat 明确冻结。
- S68b contract 已通过主控审查，并补充 HTTP passthrough 实际 forwarded-body 的独立验收；API-key/OAuth passthrough 与完整 WS ingress 证据均已由 fix1 和组合 QA闭环。
- S68a 历史阶段：`done`。
- S68a Sprint：`upstream-codex-image-tool-strip-policy-s68a`。
- Planner 已确认 `d3a1835ed` 依赖本地缺失的 `f385cdceb` 账号策略，且现有 UI 没有安全的任意 `extra` 编辑入口；S68a 因此拆成 backend policy 与专用 UI 两个互斥 draft contract。
- Backend 与 UI worker 均已完成并经独立审查；UI 的 nested precedence 与 alias normalization 两项 finding 已由 follow-up 提交闭环。组合分支集成到 `2f68fdc90`，fresh backend service、组件 `30/30` 和 typecheck 均 PASS。
- 当前唯一合法动作是在独立 worktree 对 S68a 执行组合 QA；namespace、Responses Lite `additional_tools` 和 passthrough 仍属于后续 S68b，不能在 QA 中扩实现。
- 独立 QA 已按 `docs/workflow/qa-reports/upstream-codex-image-tool-strip-policy-s68a-qa.md` 判定 PASS；S68a 当前裁决为 DONE，可作为 S68b 的策略与维护入口前置。
- S67 历史阶段：`done`。
- S67 Sprint：`upstream-v0151-protocol-wave2-s67`。
- 已批准三个互斥 contract：`upstream-gpt56-max-effort-s67a`、`upstream-codex-mcp-tool-bridge-s67b`、`upstream-compact-ops-stream-log-s67c`。
- `openai_gateway_messages.go` 由 S67b 独占并同时完成 message fallback effort-candidate 调整；S67a 不得修改该文件。
- S67 第一轮集成后再单独启动 Codex `image_gen` namespace strip，避免与 GPT 核心网关 ownership 重叠。
- 三个 S67 Worker 均已从共同基线 `09bfc7e9b` 完成实现、合规 result 和独立代码审查；主分支按 Ops logging、GPT effort、MCP bridge 顺序集成到 `fbe5dd123`。
- 首轮独立 QA 已按 `docs/workflow/qa-reports/upstream-v0151-protocol-wave2-s67-qa.md` 判定 FAIL：S67b tool-only 流的 streamed/final output index 不一致，reasoning-plus-tool 缺完整 reasoning item lifecycle。
- 已批准 fix contract：`docs/workflow/tasks/upstream-codex-mcp-tool-bridge-s67b-fix1.md`。只适配祖先 `f10bca815` 的 response-stream state/index/lifecycle 子集，不引入其 request-direction 大重构。
- Fix1 已集成为 `d1c858392`：动态 output index、reasoning/message 完整生命周期和 streamed-vs-terminal 索引断言均已进入组合分支。当前唯一合法动作是由原 QA owner 在新 worktree 复测原失败场景与 S67 组合回归。
- Retest1 已按 `docs/workflow/qa-reports/upstream-v0151-protocol-wave2-s67-retest1.md` 判定 PASS：初次 tool-only index 与 reasoning lifecycle 两项阻断均已闭环，完整 apicompat、S67 service 组合、Ops 定向、完整 handler 和路径审计通过。
- S67 当前裁决为 DONE。下一步只能先处理 `d3a1835ed` 缺失的 `f385cdceb` 账号级图片工具策略前置；不得直接启动 namespace strip 实现。
- S66 历史阶段：`done`。
- 当前 Sprint：`upstream-v0151-runtime-wave1-s66`。
- 已批准三个互斥实现 contract：`upstream-runtime-hotfixes-s66a`、`upstream-anthropic-grok-usage-s66b`、`upstream-remote-compact-reliability-s66c`。
- S66 第一波只处理 setup-token/ops/Windows WS 小修、Anthropic/Grok 用量兼容和 remote compact 可靠性；GPT-5.6 max、image_gen、MCP tool bridge 与支付并发加固不在本波范围。
- 用户已明确授权多智能体；每个 Generator 使用 `E:/codex-worktrees/sub2api/<task-id>` 独立 worktree，主 Codex 负责 diff 审核、集成、统一 QA 和最终裁决。
- 三个 Worker 均从共同基线 `2549e0b3a` 完成独立分支实现和 worker result，主线程已按 ownership 审查后集成。
- S66 第一波已集成：`d826463f1`（Anthropic cache creation + Grok effort）、`7e4420c04`（setup-token/ops/Windows WS）、`696d23875`（remote compact reliability）。
- 独立 QA：`docs/workflow/qa-reports/upstream-v0151-runtime-wave1-s66-qa.md` 首行 PASS；定向回归、compact `count=20`、完整 handler、diff/path audit 全部通过。
- 已知非阻断项：完整 service 包仍有既有 `group_peak_rate` 时区断言漂移；race 因 `CGO_ENABLED=0` 且主机无 `gcc` 未执行；未做真实上游 compact 或物理 Windows reset 联调。
- 下一波候选：GPT-5.6 `max`/effort candidates、Codex `image_gen` namespace strip、MCP/custom/tool_search bridge；支付并发大补丁继续保持独立审计边界。
- S65 历史阶段：`done`。
- 当前 Sprint：`upstream-latency-health-column-s65`。
- 当前 contract：`docs/workflow/tasks/upstream-latency-health-column-s65.md`，只移植上游合并延迟健康列，不带排行、筛选、tab 或整页布局重构。
- S65 已完成：管理端和用户端用量明细改为合并延迟健康列，首字/总耗时使用上下渐变色条和四档阈值；4 files / 31 tests、主题 2 files / 24 tests、typecheck、build 与 `git diff --check` PASS，证据见 `docs/workflow/qa-reports/upstream-latency-health-column-s65-qa.md`。
- 当前 contract：`docs/workflow/tasks/tutorial-toc-content-adjacency-s64.md`，只把桌面本页目录从视口右侧移到正文右侧邻接位置。
- S64 已完成：`2560px` 正文到 TOC 间距 `32px`，`1280px` 间距 `19px`；正文保持 `800px`，移动目录/TOC 无回归。
- S64 验证：3 files / 21 tests、typecheck、`git diff --check` 与 `2560x900` / `1280x720` / `390x844` 截图验收 PASS，证据见 `docs/workflow/qa-reports/tutorial-toc-content-adjacency-s64-qa.md`。
- 当前 contract：`docs/workflow/tasks/public-docs-layout-toolbar-s63.md`，只调整教程详情三栏和模型工具栏视觉层级。
- S63 已完成：教程详情为左目录/中正文/右 TOC 三个同级列，宽屏正文 `800px` 且文章无整卡背景；模型搜索独占一行，分类与数量统计同一条扁平行。
- S63 验证：4 files / 26 tests、typecheck、production build、`git diff --check` 全部 PASS；`2560x900`、`1280x720`、`390x844` 浏览器几何与截图验收 PASS，证据见 `docs/workflow/qa-reports/public-docs-layout-toolbar-s63-qa.md`。
- 已批准 contract：`docs/workflow/tasks/tutorial-reading-flow-s60.md`、`tutorial-markdown-interactions-s61.md`、`model-discovery-ux-s62.md`。
- 多智能体偏差记录：当前 runtime 未暴露 Agent Matrix 指定的 `deepseek-v4-pro`；本轮按用户明确授权使用可用的继承模型 worker，Final Evaluator 仍由主 Codex 执行。
- S60-S62 实现与 Evaluator 退修已完成；主线程合并验证为 5 files / 29 tests、`vue-tsc --noEmit`、`git diff --check` 全部通过。
- 首轮独立 QA 唯一 FAIL 为 S60 contract 漏列共享 `public-pages.spec.ts`；Planner 已补批仅同步 S60 文案/行为断言的窄范围，业务实现无需回滚。
- 独立 QA 重跑 PASS：5 files / 29 tests、typecheck、production build、`git diff --check` 全部通过；`1280x720`、`1024x768`、`390x844` 浏览器门禁通过，证据见 `docs/workflow/qa-reports/public-learning-model-discovery-s60-s62-qa.md`。
- 当前 contract：`docs/workflow/tasks/home-reveal-background-cover-s59.md`，只允许修正 reveal 背景缩放/定位和对应测试。
- S59 已完成：hero 背景改为容器级 `cover`，桌面保留低对比度全幅背景与局部 reveal，触屏、窄屏和 reduced-motion 使用静态背景；定向 Vitest 2 files / 29 tests、生产构建、三个视口 Chrome 验收与 `git diff --check` 均通过，QA 证据见 `docs/workflow/qa-reports/home-reveal-background-cover-s59-qa.md`。
- 当前 contract：`docs/workflow/tasks/home-conversion-accessibility-s58.md`，已完成 Planner/Evaluator 边界复核；允许进入前端实现，不允许修改数据库协议数据、后端或部署配置。
- S58 已完成：定向 Vitest 5 files / 43 tests、生产构建、桌面/移动 Chrome 验收与 `git diff --check` 均通过；QA 证据见 `docs/workflow/qa-reports/home-conversion-accessibility-s58-qa.md`。
- 当前目标：把 2026-07-08 的暖白/陶土/黑灰全前端统一、首页内嵌登录注册、排行榜增强、共享账号渠道状态可见性和首充福利 only 语义，收口成新的默认产品入口与验收基线。
- 当前结论：`640b9341d`、`7a457f25d`、`71dad20f9`、`ebc477720`、`eaf8dba78` 已把默认续做入口从 S53 推送收尾，前移到用户可见前端/支付/共享账号设置面；旧的 `upstream-main-v0144-safe-patches-s53` 结果仍成立，但已退成稳定背景层。
- 当前默认续做提示：如果用户说“继续”，优先按“首页认证入口 + 暖白控制台/公共页 + leaderboard 新语义 + shared account channel status visibility + first-time recharge bonus”这条链路判断，而不是默认回到 S53 推送或纯上游 patch 语境。
- 历史背景事实：
  - `640b9341d feat(frontend): unify warm public and console UI` 已把公共页、认证页、控制台基础壳、首页和大量后台/用户面板统一到暖白/陶土/黑灰体系。
  - 首页当前默认入口已变为 `/home` 右侧内嵌认证卡片；登录/注册仍复用原业务逻辑，但默认用户路径已不再是独立暗色登录页。
  - `7a457f25d feat: add channel status visibility setting for shared accounts` 已把 shared account 的 channel status 暴露收口为后台可配置的稳定权限边界。
  - `71dad20f9 fix(payment): make recharge package bonus first-time only` 已把充值 bonus 收口为首充 only 语义，后续支付/福利验收必须带上这个前提。
  - `ebc477720` 与 `eaf8dba78` 已把 leaderboard 稳定面继续扩展到 rank movement、new rank 和 cached refresh state。
  - `d722d24a6 merge: upstream v0.1.146 backend safe patch s56` 与 `30d4da899 docs: record upstream s56 backend safe patch validation` 已说明 S56 级别的后端安全补丁和验证记录也已进入背景层，但本轮默认用户面主线优先级更高。
  - `55d0e1ec3 fix(models): support non-v1 OpenAI models URLs` 已进入近期稳定兼容背景，后续模型目录/同步问题排查不应继续假设所有 OpenAI models URL 都固定带 `/v1`。
- 当前已确认事实：
  - S53 contract：`docs/workflow/tasks/upstream-main-v0144-safe-patches-s53.md`。
  - S53 集成分支：`codex/upstream-main-v0144-s53-safe-patches`。
  - S53 隔离 worktree：`E:/codex-worktrees/sub2api/upstream-main-v0144-s53-safe-patches`。
  - S53 候选提交：`e5dc1f597`、`4dd3aee5c`、`6bd248fd1`。
  - S53 merge commit：`dbdeb1ba1 merge: add upstream s53 safe patches`。
  - S53 实际补充一条范围修正提交：`test(openai): scope s53 mapped billing tests`，用于移除 upstream hotpath 测试中依赖本地未 port helper 的非目标测试块。
  - S53 QA 报告：`docs/workflow/qa-reports/upstream-main-v0144-safe-patches-s53-qa.md`，结论 PASS。
  - S53 明确跳过：usage log queue backpressure、group capacity batching、concurrency cleanup、Codex image tool policy、error request UI alignment、Anthropic Fable 7d_oi、deploy migration timeout、Grok UI/README changes。
  - S45-S52 集成分支：`codex/upstream-main-v0143-s45-s52-batch`。
  - S45-S52 基线：`main` / `origin/main` 的 `485eaf801 docs: record affiliate risk merge`。
  - S45-S52 按顺序 `cherry-pick -x` 完成：`544accdd3`、`af6c8fdeb`、`9aa85e59e`、`6888e9da5`、`512f44c13`、`6abeb0796`、`248bf80dc`、`c558b6eda`、`fed128046`、`5ce438fa7`、`b6970cdc6`、`4f9542e34`。
  - S45 redeem 普通兑换现在拒绝 invitation/internal marker code，不再消耗这类 code。
  - S46 Codex import 优先按 `chatgpt_user_id` 等用户身份匹配，shared account id 仅作最后 fallback；import API 请求超时提升到 120s。
  - S47 ops realtime stats 新增内部缩减账号查询路径，并抑制 request canceled 日志噪音。
  - S48 Codex OAuth input 保留 reasoning item 的 `encrypted_content` / `content` / `summary`，剥离 replay-unsafe `rs_*` id，缺失 `summary` 时补空数组。
  - S49 `/responses/compact` 不再注入 Codex image bridge tool、tool_choice 或 bridge instructions。
  - S50 Claude Code `>= claude-cli/2.1.193` streaming idle keepalive 改用空 content delta；旧客户端继续 ping。
  - S51 Anthropic API Key 账号支持 `extra.anthropic_apikey_auth_scheme = "authorization_bearer"`，默认仍使用 `x-api-key`。
  - S52 Antigravity OAuth 401 在有 refresh token 时进入临时不可调度恢复路径；无 refresh token 时仍置 error。
  - S45 contract 已起草：`docs/workflow/tasks/affiliate-risk-alerts-s45.md`。
  - S45 目标是“风险评分 + ops 告警 + 奖励兑现冻结”，不是单条规则封号。
  - S45 扫描周期默认 `20m`，但必须加后台设置允许运营自行调整；扫描窗口固定为最近 `12h`。
  - S45 扫描间隔设置建议限制在 `5-1440` 分钟，非法值回退默认 `20m`，并尽量运行中生效。
  - S45 默认动作：`P3` 只告警；`P2/P1` 告警并冻结邀请奖励兑现。
  - S45 冻结范围只覆盖首次 API 调用奖励 claim 和邀请返佣 quota 转余额；不自动封号、不自动禁用 API key、不回滚历史奖励、不撤销邀请关系。
  - S45 数据源已确认可用：`users.register_ip`、`users.last_login_ip`、`usage_logs.ip_address`、`user_affiliates.inviter_id`、`user_affiliate_ledger.action = 'api_call_reward'`。
  - S45 现有索引可支撑候选用户优先扫描：已有 `usage_logs(user_id, created_at)`、`usage_logs(created_at)`、`usage_logs(ip_address)`、`user_affiliates(inviter_id)` 和 API reward 防重复索引。
  - S45 实现仍需补三个窄索引：`users(created_at)`、`user_affiliate_ledger(action, created_at)`、`usage_logs(ip_address, created_at) WHERE ip_address <> ''`。
  - S45 复用 `ops_alert_events`；但 `OpsService.CreateAlertEvent` 本身只写事件，邮件通知当前在 `OpsAlertEvaluatorService` 内部完成，实现时要抽小 helper 或保留为明确 follow-up，不能复制大段邮件逻辑。
  - S45 必须在干净 worktree 实现；实现前仍要做 dirty-tree preflight，settings 相关文件只允许新增扫描间隔设置，不允许吸收无关脏改。
  - S44 contract 已起草：`docs/workflow/tasks/upstream-main-v0143-group-peak-rate-impl-s44.md`。
  - S44 曾因 dirty-tree preflight 触发 stop rule；后续已在隔离 worktree 完成并合入 `main`。
  - S44 最终使用本地 migration `backend/migrations/181_add_group_peak_rate_multiplier.sql`；福利券合并使用 `backend/migrations/182_welfare_vouchers.sql`，编号未冲突。
  - S43 contract 已起草：`docs/workflow/tasks/upstream-main-v0143-group-peak-rate-plan-s43.md`。
  - S43 contract 已评审批准：planning-only，允许路径只包含 workflow 文档，业务实现必须转入后续 Sprint。
  - S43 目标上游提交是 `915c60b15`、`1034f576d`、`11a3da65c`，三者共同组成订阅分组高峰时段倍率主功能、全链路透传/计费边界修正、配置加固和服务端时区展示。
  - 本地预检确认尚无 `peak_rate_enabled`、`peak_start`、`peak_end`、`peak_rate_multiplier` 字段或链路；当前只有普通 `rate_multiplier`、`image_rate_multiplier` 和用户专属分组倍率/RPM。
  - S43 不是小补丁：它会触达 Ent/migration、admin group handler/DTO、group service、API key auth cache、billing/gateway usage recording、available channels、payment/subscription plan API、admin/user frontend display 和 i18n。
  - S43/S44 规划时主工作树存在 payment、welfare voucher、settings、billing、gateway、frontend payment/i18n、knowledge 脏改；高峰倍率触达路径与这些文件重叠，因此当时必须先收口或隔离。
  - 上游使用 migration `158_add_group_peak_rate_multiplier.sql`，本地 migration 已推进到更高编号且有未提交 migration 工作；实现 Sprint 必须分配本地下一安全编号，不能照搬上游 158。
  - 上游允许 `peak_rate_multiplier=0` 代表高峰免费/折扣策略；实现前需要确认本地产品语义是否接受。
  - S35 合并计划已批准：`docs/workflow/tasks/upstream-main-v0142-merge-plan-s35.md` 标记为 approved，后续实现必须按独立 Sprint 推进。
  - S36 contract 已批准并完成：`docs/workflow/tasks/upstream-main-v0142-payment-refund-s36.md` 标记为 approved，结果和 QA 记录已落到 worker-results / qa-reports。
  - S37 contract 已批准：`docs/workflow/tasks/upstream-main-v0142-openai-codex-gateway-s37.md`，实现必须限制在该 contract 的 allowed paths 内。
  - S37 实际合入：Codex OAuth `reasoning` items 不再整项丢弃，而是保留 `encrypted_content` / `content` / `summary`，仅移除 `rs_*` id，并在缺少 `summary` 时补空数组；`gpt-5.5-pro` 作为独立 Codex 模型名保留，同时沿用 GPT-5.5/GPT-5.4 fallback 价格和长上下文计费策略。
  - S38a contract 已批准：`docs/workflow/tasks/upstream-main-v0142-account-repo-count-s38a.md`，候选只包含 `fd004bdd8 fix(account-repo): Clone query before Count to prevent state pollution`。
  - S38a 预检显示 `backend/internal/repository/account_repo.go` 和 `backend/internal/repository/account_repo_integration_test.go` 当前干净，且 `fd004bdd8` 只触碰这两个文件。
  - S38a 实际合入：`accountRepository.ListWithFilters` 使用 `q.Clone().Count(ctx)`，避免 Ent interceptor 追加 predicate 时污染随后分页列表查询；集成测试新增单页 `pagination.Total == len(accounts)` 断言。
  - S39a contract 已批准：`docs/workflow/tasks/upstream-main-v0142-frontend-api-base-s39a.md`，候选只包含 `2a58a57a7` 中不触碰 dirty `SettingsView.vue` 的 direct frontend request API-base 核心。
  - S39a 预检显示：`2a58a57a7` 除 `frontend/src/views/admin/SettingsView.vue` 外的 touched paths 当前干净；`SettingsView.vue` 已有 19 行本地脏改，本轮 denied。
  - S39a 本地适配：上游聚合 i18n 文件在本地拆分为 `frontend/src/i18n/locales/{en,zh}/admin/settings.ts`，因此 contract allowed/acceptance paths 已按本地结构调整。
  - S39a 实际合入：新增 `frontend/src/api/url.ts`，并让 admin ops WebSocket、API client refresh、setup API、account test streaming、key usage gateway fetch、custom page fetch 和 Stripe popup polling 使用配置 API base / gateway origin。
  - S40 contract 已批准：`docs/workflow/tasks/upstream-main-v0143-claude-oauth-payload-s40.md`，候选只包含 `5bd9368ab fix claude oauth token exchange payload`。
  - S40 预检显示 `backend/internal/repository/claude_oauth_service.go` 与 `backend/internal/repository/claude_oauth_service_test.go` 当前干净，且 `5bd9368ab` 只触碰这两个文件。
  - S40 实际合入：Claude OAuth setup-token code exchange 不再向 token endpoint 发送 `expires_in`；对应测试改为断言 setup token 也 omit `expires_in`。
  - S41 contract 已批准：`docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md`，候选只包含 `f5b296127 fix: Handle invalid arguments correctly for Gemini reasoning models`。
  - S41 预检显示 `backend/internal/pkg/antigravity/claude_types.go`、`backend/internal/pkg/antigravity/request_transformer.go`、`backend/internal/pkg/antigravity/request_transformer_test.go` 当前干净；上游 `f5b296127` 只触碰前两个文件，本地允许补 focused unit tests。
  - S41 目标行为：Gemini reasoning models 不再收到空工具场景下的强制 `toolConfig`，且不再收到 `stopSequences`、`temperature`、`topP`、`topK` 等无效参数；非 reasoning Gemini 行为保持不变。
  - S41 实际合入：`modelDef` 增加 `IsReasoning`，新增 `IsGeminiReasoningModel`；Antigravity 转换器在 Gemini reasoning 模型且无工具时省略 `toolConfig`，并在 generationConfig 中省略 reasoning 模型不支持的 stop/temperature/topP/topK 参数；新增测试覆盖 reasoning 与 non-reasoning 分支。
  - S42 contract 已批准：`docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md`，候选只包含 `e236bff1e fix: aggregate user model stats by requested model`。
  - S42 预检显示 `backend/internal/repository/usage_log_repo.go` 与 `backend/internal/repository/usage_log_repo_request_type_test.go` 当前干净，且上游 `e236bff1e` 只触碰这两个文件。
  - S42 目标行为：单用户模型统计复用本地 `getModelStatsWithFiltersBySource(..., usagestats.ModelSourceRequested)`，按 `requested_model` 聚合并在空值时回落 `model`。
  - S42 实际合入：`GetUserModelStats` 改为复用 requested-model 聚合 helper；新增 sqlmock 测试锁定 `requested_model` fallback 表达式、user/time 参数顺序和扫描结果。
  - 最新 release 已刷新为 `v0.1.143` / tag commit `9caa3c9c5`；`v0.1.143` 后 `upstream/main` 还有 `a5638a4e5` 与 `0b8e5eec3`，均未进入本轮。
  - S39 中 `8c2d9b9a1` 继续 deferred，因为它移除 OpenAI 默认模型列表中的 `gpt-5.3-codex`，属于可见产品策略而非纯兼容小修。
  - S38 full bundle 中 `9f5b57fc9` 继续 deferred，因为它触碰 `usage_billing_repo.go`、`billing_cache_service.go`、`gateway_service.go`、`usage_billing.go`、config 和 deploy 示例等当前脏树/高风险区域。
  - S38 full bundle 中 `03727ac36` 继续 deferred，因为它触碰 subscription repository/service、billing cache、admin routes、DTO/types、frontend subscription API/types 和 integration tests，当前不适合混入窄补丁。
  - S37 明确本地等价：`9491de0a3`、`ae5e980dd`、`65fa72892`、`0a97a5f46`、`2b49d662c`、`011278204`、`e5f7836bf`、`82553c4dc`、`7a38c6621` 已由本地 S21-S30/S23-S26 相关实现和本轮定向测试覆盖。
  - S37 未触碰 Ent、migrations、wire、frontend、payment/refund、welfare voucher、Studio Bridge、user proxy/account ownership、repository、generic `gateway_service.go`、knowledge、deploy、assets 或 README。
  - S37 候选提交是 `9491de0a3`、`ae5e980dd`、`65fa72892`、`0a97a5f46`、`2b49d662c`、`011278204`、`e5f7836bf`、`73de2ea7f`、`b28a22333`、`82553c4dc`、`7a38c6621`；本轮必须逐项判定 `ported / local-equivalent / skipped`。
  - S37 预检显示：chat transport failover、token refresh non-retry、tool args dedupe、Spark image tool strip、Codex image bridge `tool_choice=auto`、quota platform billing、OpenAI count_tokens bridge 大概率已有本地等价；但当前本地 Codex OAuth transform 仍会丢弃 `reasoning` items，和 `73de2ea7f` 的 encrypted reasoning 保留方向不等价，需作为主要待 port 风险点。
  - S36 实际合入：退款状态新增 `REFUND_PENDING`；Stripe/Airwallex/WxPay 增加退款查询；pending refund 支持回滚本地扣减并由 admin query/finalize 收口；匿名 `out_trade_no` public verify 返回最小 DTO；支付金额显示读取订单币种，余额 credit 显示保留本地 credit 语义；支付提供商卡片和弹窗对空 `supported_types` 做防御。
  - S36 明确跳过：`c6f375d3a` 被 `b1403e8b2` superseded，订阅订单金额继续是 plan direct price；`65ad7df4f` 中 `SettingsView.vue` 相关部分因 denied path 跳过。
  - S36 候选提交是 `c6f375d3a`、`b1403e8b2`、`55242ffac`、`65ad7df4f`、`7316d8302`、`93a3bf307`、`930326116`；它们只允许进入支付/退款批次。
  - S36 不允许触碰 Ent、migrations、wire、welfare voucher、user proxy/account ownership、Studio Bridge、knowledge、Docker/deploy、README/assets 或当前未列入 allowed paths 的脏文件。
  - GitHub latest release 已确认为 `v0.1.142` / `60da9ba17`，发布时间为 2026-07-01。
  - 临时 worktree 试算 `git merge --no-commit --no-ff v0.1.142` 会产生大量冲突，冲突集中在 Ent 生成代码、account/proxy schema、gateway、payment、usage 和前端视图。
  - `git apply --check --3way` 显示多个小补丁可单独迁入，但 Grok、OpenAI Spark shadow、Codex detect、Anthropic dateline / Sonnet5 是大功能链路，不能混入小补丁批次。
  - S35 规划时本地主工作树仍有福利券、设置、用户代理、前端 locale/view 和 knowledge 脏改，且 `main` 相对 `origin/main` ahead；后续代码 port 前必须先收口或隔离这些本地变更。
  - 本地 `main` 与 `upstream/main` 严重分叉，直接 merge 会冲突大量 Ent、wire、网关、设置页和前端文件。
  - 本地当前主线包含 Studio Bridge / 落叶AI、支付套餐、模型市场、Canvas、工单和公共页定制；上游小步迁移 Sprint 不允许覆盖这些产品面，产品合并批次则必须单独列出真实触达范围和验证。
  - 上游 `v0.1.137` 低风险候选包括前端依赖安全、token refresh 不可重试、zstd、非 JSON/SSE 错误保留、计费兜底、thinking 协议兼容、Responses sticky hash、Haiku 探测、OpenAI responses tool probe 和 ACL 拒绝信息。
  - `b81694929` 是完整功能链，不是安全/兼容小补丁；适合独立 S17，且不需要 Ent/migration/VERSION。
  - S17 新增的上游 quota/reset 能力只挂在管理员 OpenAI OAuth 账号路径，不改变本地账号 quota reset 语义。
  - 当前 `origin/main..HEAD` 已实际触达 `backend/cmd/server/wire_gen.go`、`backend/internal/repository/studio_bridge_repo.go`、`frontend/src/views/public/ModelPlazaView.vue`、`frontend/src/components/layout/AppHeader.vue`、`frontend/src/views/user/KeysView.vue`、`frontend/src/views/admin/SettingsView.vue`，以及统一 Key、APIMart 图片网关、quota reset 相关文件。
  - 真实 UI smoke、真实 OpenAI OAuth 上游和真实 APIMart 上游仍未在本地完成；当前证据以代码级定向测试、typecheck/build 和审查为主。
  - APIMart task webhook 适合补强视频/长任务可靠结算；Sub2API 仍是 Studio Bridge / 落叶AI余额和扣费真源，`chatgpt2api` 不应绕过 Sub2API 决定扣费。
  - 当前 APIMart 图片异步模型仍通过 `openai_images.go` 内部轮询后同步返回，S18 不改变普通 `/v1/images/generations` 兼容行为。
  - S19 已明确跳过 OpenAI image failover、token refresh retry amplification、OAuth promo signup、scheduler outbox dedup/cleanup、cyber policy、channel monitor jitter、Claude OAuth system prompt blocks 和 migration-heavy 链路。
  - S20 明确跳过 `prefer soonest reset` 调度策略、订阅支付返佣、Claude mimicry 去掉 `cch`、邮箱绑定后缀白名单、CI/deploy/README/sponsor/VERSION 和前端 UI 合并；这些需要独立 Sprint 或产品确认。
  - S20 实际迁入 Gemini schema 清理、OpenAI images `response.incomplete` / no-output 诊断、Vertex Anthropic beta 过滤、Claude Code 任意 `cc_entrypoint=` 识别、GLM reasoning effort 归一、OpenAI chat-only upstream endpoint 记录、promo 过期清空。
  - 本地图片 handler 会把 `OpenAIImagesUpstreamError` 当作已写出的上游错误直接结束；因此 S20 在 `openai_images_responses.go` 内将非内容过滤的 `response.incomplete` 转为 `UpstreamFailoverError`，避免 502 incomplete 被误当用户错误提前返回。
  - S21 必须把 Spark `image_generation` tool strip 放在本地图片权限 gate 前；否则 Codex CLI 默认携带的 tool 会在被剥离前被误判为图片生成意图。
  - S21 明确跳过订阅支付返佣、prefer-soonest-reset 调度、Claude mimicry `cch` 删除、GPT-5.5 instructions fallback、README/sponsor/VERSION 和 SELinux compose 标签。
  - S21 实际合入 Spark `image_generation` tool strip、OpenAI weekly reset 二次确认、usage cache token 明细展示、邮箱绑定后缀白名单校验。
  - S22 候选评估中，usage cache breakdown、Spark image tool strip 已本地等价；支付/订阅/余额预扣、order currency、Antigravity fallback、GPT-5.5 instructions fallback、ops chart UI、Claude terminal template、payment supported-types 继续跳过。
  - S22 实际移植范围限制在后端 OpenAI/Codex 兼容和 auth/token 小修，不触碰前端或本地产品面。
  - S22 实际合入：ChatCompletions->Responses first-chunk tool arguments 不翻倍、Responses passthrough function-call arguments 去重、`refresh_token_invalidated` 非重试、chat-completions transport error 进入 failover、email auth identity create err 不再被 shadow。
  - S23 候选评估中，`0da1fe28e`、`9491de0a3`、`cc7612bdb`、`8a7269f53`、`40c825273`、`e5f7836bf` 都适合组成纯后端兼容小批；`82553c4dc` quota platform billing、`da810c3b4` Keys unlimited reactivation、`7c2fee6c9` fallback pricing log dedup 可后续独立批次；Grok、codex_cli_only 全量指纹加固、model_not_found 404 handler refactor、支付/订阅/Ops/Keys 前端功能继续跳过。
  - S23 实际合入：OpenAI image output text-only 防误计数、图片模型拒绝 400 透传、OpenAI overloaded/slow_down failover、response.failed 脱敏、Responses->Anthropic custom/unknown tool schema 规范化、Codex image bridge `tool_choice=auto`。
  - S24 候选评估中，`82553c4dc`、`7c2fee6c9`、`da810c3b4` 适合组成纯后端运维/计费小批；`b105cc0fd` Codex JSON/developer input 行为需要单独确认本地 transform 语义，暂留 S25；前端 API base、Keys column settings、subscription/payment 显示、Ops system log key id 继续跳过。
  - S24 实际合入：OpenAI async usage billing 捕获并传递 request-time quota platform、fallback pricing warn 按模型去重、quota exhausted key 设置为 unlimited 时重新激活。本地缺少上游完整 `UserPlatformQuotaRepository` 持久层，S24 只保留平台字段到统一 billing command，不引入 migrations/flusher。
  - S25 实际合入：Codex OAuth transform 将 `role:"system"` input item 改写为 `role:"developer"` 并保留在 `input`，同时继续把 system 文本镜像到 `instructions`；这覆盖 Responses JSON mode 需要在 input 中保留 JSON 指令的场景。`7a38c6621` OpenAI count_tokens bridge 范围较大，留作 S26 单独 contract。
  - S26 实际合入：OpenAI group 的 `/v1/messages/count_tokens` 不再路由级 404，新增 handler/service 将 Anthropic count_tokens 请求转换为 OpenAI `/v1/responses/input_tokens` 并返回 `input_tokens`。本地没有上游部分新 helper，已按现有 `ParseGatewayRequest`、billing check、account selection 和 HTTP upstream 调用适配。
  - S27 候选评估中，`27600b1d2c` count_tokens generation field filter、`1d47fd6300` DeepSeek `reasoning_content`、`2c14efeaa0` Images `n` 透传、`888cd8092d` image moderation error、`32ea9cfe1f` API key SSE body fallback、`89dffdd2e1` Anthropic cache token input semantics、`6aec505016`/`be3613593b` OAuth 401 no credentials overwrite、`c10598dfe5` idempotency UTF-8 truncation 都已在本地等价；本轮只迁入仍缺失且范围小的 `709cf6185` / model-aware Codex instructions。
  - S27 实际合入：新增 `openai.CodexBaseInstructionsForModel` 与 GPT-5.1/GPT-5.2/GPT-5.5 Codex base prompt 资源；`applyCodexOAuthTransform` 和 `OpenAIGatewayService.Forward` 在空/空白 `instructions` 时使用模型感知默认 prompt，`gpt-5.5` 走 GPT-5.5 Codex base prompt。
  - S28 候选评估中，`7cbf82ed6` 属于纯后端 OpenAI 错误分类小补丁，适合独立迁入；其中上游 `openai_account_runtime_block_fastpath.go` 在本地不存在对应文件和编译入口，本轮只迁入本地已有网关路径。
  - S28 实际合入：新增 `isOpenAIContextWindowError`，让 HTTP 502/5xx、Responses stream、Chat Completions bridge buffered/stream 的 context-window 超限错误不再构造 `UpstreamFailoverError`；同时保留 `server_is_overloaded` 等 transient 错误的 failover 行为。
  - S29 候选评估中，`b9509e823a` / `ed2aac25a` long-context cache_read/cache_creation 倍率、`8a999f438d` WS terminal first-token 排除、`0a521f09fb` Gemini messages tool_use 关闭、`03ae510c68` ops count_tokens metrics 排除均已本地等价；本轮只迁入仍缺失且范围小的 `e9a2db8e80`。
  - S29 实际合入：OpenAI Responses streaming 在终端 `response.completed` / `response.done` / incomplete/cancel 事件的 `response.output` 为 null 或空且可从流式 delta 重建时，补齐为 SDK 可解析 output 数组；没有可重建内容时补为空数组，避免客户端解析 terminal output 为 null。
  - S30 候选评估中，`ae5e980dd` 与 `dbdbfb112` 都是纯后端 OpenAI chat-completions bridge 小补丁，且本地 `ForwardAsChatCompletions` 仍缺失对应行为，适合合并为一批。
  - S30 实际合入：`ForwardAsChatCompletions` 在入口补齐 `codex_cli_only` 检测和拒绝日志；OAuth 普通 Chat Completions 转 Responses 后调用 `SkipDefaultInstructions`，并显式保留空 `instructions` 字段，避免给 chat bridge 注入默认 Codex prompt。
  - S31 基准刷新后，上游已到 `v0.1.140-1-g89b2d63ef`；`v0.1.140` 尾部主要是前端排序、OAuth completion flow、支付退款 pending、Grok/platform quota migration、sponsor 和 VERSION，均不适合混入本轮小补丁。
  - S31 候选评估中，`82576e0a3` email auth identity create err 和 `65fa72892` chat transport failover 已本地等价；本轮只迁入仍缺失的 `56c62c59c` API Key ACL 拒绝信息补 client IP。
  - S31 实际合入：API Key whitelist/blacklist 拒绝响应返回 `Access denied. Your IP is <client-ip>`，继续通过 `ip.GetTrustedClientIP(c)` / Gin `ClientIP()` 解析客户端 IP；未配置 trusted proxies 时不信任伪造转发头，配置 trusted proxies 时可显示转发后的 client IP。
  - S32 基准刷新后，上游已到 `v0.1.141-1-gdc1bc1545`；`v0.1.141` 尾部主要是 admin/user usage parity、订阅支付金额显示、sponsor/VERSION 和前端/产品面范围，均不适合混入本轮小补丁。
  - S32 实际合入：新增 no-account 错误分类器与模型可用性诊断；Anthropic/Gemini/OpenAI 网关在账号池非空但无账号支持请求模型时返回 `404 model_not_found`，空池、compact unsupported、查询失败、限流、quota pause、runtime block 和 slot/wait 容量问题继续返回 `503`。
  - S32 本地主工作树 handler 定向测试通过；service 定向测试在主工作树被无关 proxy/account 脏改导致的 `ProxyRepository` stub 编译错误挡住，同一 S32 patch 已在临时干净 worktree 上通过 handler/service 定向测试。
  - S33 筛选中，`a611742910` 图片生成上游 context 脱离和 `6acb46c113` 通用网关本地容量错误标记已被本地等价覆盖；`0f8e2d093` admin 账号敏感凭证 redaction 有安全价值但触达 `admin_service.go`、`frontend/src/types/index.ts` 等当前脏文件，留作后续独立批次。
  - S33 实际合入：上游 `147c1879d` / `fix(payment): support plural subscription validity units`。本地前端套餐编辑器会保存 `weeks` / `months`，后端创建支付订单时现在将 `weeks` 折算为 `days * 7`、`months` 折算为 `days * 30`，保持 `days` / 未知单位回退原值。
  - S34 筛选中，`271aba1abe` IP 拒绝 SLA 排除、`930326116` 订阅支付金额显示、`0ae3329613` API Key 名称 XSS 转义、`04deb819b0` EasyPay 查单 `trade_status`、`1e2193c3d2` WebSocket 用量去重、`bf1a2d6dc2` Codex reset window 统计、`c40a74d98` 内容审计管理员自动封禁豁免、`55655b865` Responses→Chat reasoning-only stream、`727ac3f68` / `app_session_terminated` 非重试、`f6e0ebc6b` Anthropic window cooldown、`bf3787de1` Claude Code count_tokens 放行、`20f534078` Responses→Chat usage 明细都已被本地等价覆盖。
  - S34 实际合入：上游 `65559ac58` / `fix(antigravity): merge system role messages`。Antigravity 转换器现在从 `messages` 中提取 `role:"system"` 的 parts，并追加到 Gemini `systemInstruction`；普通 user/assistant contents 不变，assistant 仍映射为 Gemini `model` role。
  - S44 实际合入：上游 `v0.1.143` 订阅分组高峰时段倍率能力，使用本地 migration `181_add_group_peak_rate_multiplier.sql`，并保留本地模块化 i18n。高峰倍率仅作用于 token 计费，token-mode 图片 output tokens 也走 token 高峰倍率；图片/per-request 计费不受高峰倍率影响。
  - S44 复核修复：Create handler 不再先于 service normalization 拒绝标准分组携带高峰字段；Keys 智能多分组摘要现在展示 route group badge 和高峰窗口；前端类型/测试 fixture 补齐 peak 字段；OpenAI/generic gateway 按最终 billing mode 选择日志倍率。
- 目标验证入口：
  - `docs/workflow/tasks/upstream-main-v0141-antigravity-system-role-s34.md`
  - `docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md`
  - `docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md`
  - `docs/workflow/tasks/upstream-main-v0143-user-model-stats-requested-s42.md`
  - `docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md`
  - `docs/workflow/tasks/upstream-main-v0143-claude-oauth-payload-s40.md`
  - `docs/workflow/worker-results/upstream-main-v0143-claude-oauth-payload-s40-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0143-claude-oauth-payload-s40-qa.md`
  - `docs/workflow/tasks/upstream-main-v0142-frontend-api-base-s39a.md`
  - `docs/workflow/worker-results/upstream-main-v0142-frontend-api-base-s39a-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0142-frontend-api-base-s39a-qa.md`
  - `docs/workflow/tasks/upstream-main-v0142-account-repo-count-s38a.md`
  - `docs/workflow/worker-results/upstream-main-v0142-account-repo-count-s38a-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0142-account-repo-count-s38a-qa.md`
  - `docs/workflow/tasks/upstream-main-v0142-openai-codex-gateway-s37.md`
  - `docs/workflow/worker-results/upstream-main-v0142-openai-codex-gateway-s37-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0142-openai-codex-gateway-s37-qa.md`
  - `docs/workflow/tasks/upstream-main-v0142-payment-refund-s36.md`
  - `docs/workflow/tasks/upstream-main-v0142-merge-plan-s35.md`
  - `docs/workflow/worker-results/upstream-main-v0141-antigravity-system-role-s34-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0141-antigravity-system-role-s34-qa.md`
  - `docs/workflow/tasks/upstream-main-v0143-group-peak-rate-impl-s44.md`
  - `docs/workflow/worker-results/upstream-main-v0143-group-peak-rate-impl-s44-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0143-group-peak-rate-impl-s44-qa.md`
  - `docs/workflow/tasks/upstream-main-v0141-payment-validity-units-s33.md`
  - `docs/workflow/worker-results/upstream-main-v0141-payment-validity-units-s33-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0141-payment-validity-units-s33-qa.md`
  - `docs/workflow/tasks/upstream-main-v0141-model-not-found-s32.md`
  - `docs/workflow/worker-results/upstream-main-v0141-model-not-found-s32-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0141-model-not-found-s32-qa.md`
  - `docs/workflow/tasks/upstream-main-v0140-api-key-acl-denial-s31.md`
  - `docs/workflow/worker-results/upstream-main-v0140-api-key-acl-denial-s31-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0140-api-key-acl-denial-s31-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-chat-bridge-guards-s30.md`
  - `docs/workflow/worker-results/upstream-main-v0139-chat-bridge-guards-s30-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-chat-bridge-guards-s30-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-responses-stream-output-s29.md`
  - `docs/workflow/worker-results/upstream-main-v0139-responses-stream-output-s29-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-responses-stream-output-s29-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-openai-context-window-s28.md`
  - `docs/workflow/worker-results/upstream-main-v0139-openai-context-window-s28-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-openai-context-window-s28-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-codex-model-instructions-s27.md`
  - `docs/workflow/worker-results/upstream-main-v0139-codex-model-instructions-s27-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-codex-model-instructions-s27-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-openai-count-tokens-s26.md`
  - `docs/workflow/worker-results/upstream-main-v0139-openai-count-tokens-s26-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-openai-count-tokens-s26-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-codex-json-input-s25.md`
  - `docs/workflow/worker-results/upstream-main-v0139-codex-json-input-s25-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-codex-json-input-s25-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-backend-compat-s24.md`
  - `docs/workflow/worker-results/upstream-main-v0139-backend-compat-s24-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-backend-compat-s24-qa.md`
  - `docs/workflow/tasks/upstream-main-v0139-backend-compat-s23.md`
  - `docs/workflow/worker-results/upstream-main-v0139-backend-compat-s23-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0139-backend-compat-s23-qa.md`
  - `docs/workflow/tasks/upstream-main-v0138-followup-safe-patches-s22.md`
  - `docs/workflow/worker-results/upstream-main-v0138-followup-safe-patches-s22-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0138-followup-safe-patches-s22-qa.md`
  - `docs/workflow/tasks/upstream-main-v0138-followup-safe-patches-s21.md`
  - `docs/workflow/worker-results/upstream-main-v0138-followup-safe-patches-s21-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0138-followup-safe-patches-s21-qa.md`
  - `docs/workflow/tasks/apimart-task-webhook-s18.md`
  - `docs/workflow/tasks/upstream-main-v0137-postfixes-s19.md`
  - `docs/workflow/tasks/upstream-main-v0138-small-patches-s20.md`
  - `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
  - `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
  - `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
  - `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
  - `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
  - `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
  - `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`
  - `docs/workflow/worker-results/upstream-main-v0137-postfixes-s19-result.md`
  - `docs/workflow/qa-reports/upstream-main-v0137-postfixes-s19-qa.md`
- 已执行验证：
  - `go test -tags=unit ./internal/service -run "TestGetFallbackPricing_FamilyMatching|TestGetModelPricing_DoubaoEmbeddingVisionImageInputRate|TestCalculateCost_DoubaoEmbeddingVisionDifferentialInput|TestHandleNonStreamingResponse|TestHandleStreamingResponse_SSEErrorEvent|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback|TestExtractOpenAIReasoningEffortFromBody" -count=1`
  - `go test ./internal/service -run "TestOpenAIGatewayServiceRecordUsage_MissingPricingRecordsZeroCostUsageLog|TestExtractOpenAIReasoningEffortFromBody|TestIsNonRetryableRefreshError|TestResolveThinkingProtocol|TestThinkingFilters|TestNormalizeChineseLLMThinking|TestApplyThinkingEnabledFallback" -count=1`
  - `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback" -count=1`
  - `go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1`
  - `go test ./internal/pkg/apicompat -count=1`
  - `go test -tags=unit ./internal/service -run "TestParseGatewayRequest_ResponsesInput|TestGenerateSessionHash_ResponsesInputProducesHash|TestDecideResponsesProbeSupportRequiresFunctionCallOn2xx|TestOpenAIResponsesProbePayloadForcesFunctionCall|TestSelectResponsesProbeModelUsesMappedUpstreamModel|TestProbeOpenAIAPIKeyResponsesSupportPersistsToolCapability" -count=1`
  - `go test -tags=unit ./internal/handler -run "TestDetectInterceptType_MaxTokensOneHaiku|TestSendMockInterceptResponse_MaxTokensOneHaiku" -count=1`
  - `go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionUsesGenericMessageForBlacklistDenial" -count=1`
  - `go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback|Test.*GenerateSessionHash|TestParseGatewayRequest" -count=1`
  - `go test -tags=unit ./internal/service -run "TestOpenAIQuota" -count=1`
  - `go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1`
  - `go test ./internal/service -run "^$" -count=1`
  - `go test ./internal/handler/admin -run "^$" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"`
  - `git diff --check`
  - S15-S17 当时的 denied-path audit returned `NO_DENIED_PATHS`；当前 `origin/main..HEAD` 已包含后续产品合并，不再适用该结论。
  - lockfile scan confirmed no `form-data@4.0.5` / `form-data: 4.0.5` remains.
  - `go test ./internal/service -run "Test.*Failover.*Body|Test.*Cached.*Body|Test.*Anthropic.*Window|Test.*Cooldown|TestOpenAI.*Images" -count=1`
  - `go test ./internal/repository -run "Test.*Account.*List|Test.*Refresh.*Candidate|Test.*Temp.*Unscheduled|TestAccountsToService" -count=1`
  - `go test ./internal/server -run "Test.*APIContract" -count=1`
  - `go test -tags=unit ./internal/service -run "TestOpenAIGatewayService_HandleFailoverSideEffects_DoesNotRereadResponseBody|TestOpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|TestHandleUpstreamError_AnthropicWindowLimitPreemptsTempUnschedRule" -count=1`
  - `go test -tags=unit ./internal/repository -run "TestAccountsToService_LargeActiveAccountSetDoesNotExceedPostgresParameterLimit" -count=1`
  - S19 denied-path audit returned `NO_DENIED_PATHS`.
  - `go test ./internal/service -run "TestCleanToolSchema|TestExtractImagesUpstreamError|TestSummarizeNoOutputBody|TestImagesOAuthNonStreaming_CompletedNoImageTriggersSameAccountRetry|TestImagesOAuthNonStreaming_Incomplete|TestVertexBetaFilter|TestFilterVertexBetaTokens|TestClaudeCodeValidator|TestNormalizeGLMOpenAIReasoningEffort|TestForwardAsRawChatCompletions_NormalizesGLMReasoningEffortForUpstream" -count=1`
  - `go test -tags=unit ./internal/service -run "TestNormalizeGLMOpenAIReasoningEffort|TestForwardAsRawChatCompletions_NormalizesGLMReasoningEffortForUpstream" -count=1`
  - `go test ./internal/handler -run "Test.*OpenAI|Test.*ChatCompletions|Test.*Responses|Test.*Messages" -count=1`
  - `git diff --check` 通过；仅提示 `docs/workflow/status.md` 下次 Git 触碰时 LF 会替换为 CRLF。
  - `go test ./internal/service -run "TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey|TestStripCodexSparkImageGenerationToolFromRawPayload|TestAuthServiceBindEmailIdentity_.*Suffix|TestAuthServiceSendEmailIdentityBindCode_.*Suffix" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/admin/usage/__tests__/UsageStatsCards.spec.ts"`
  - `git diff --check`
  - S21 denied-path audit returned `NO_DENIED_PATHS`.
  - `go test ./internal/pkg/apicompat -run "TestStream_ToolCallArgumentsInFirstChunkNotDoubled" -count=1`
  - `go test -tags=unit ./internal/service -run "TestHandleStreamingResponsePassthroughDeduplicatesFunctionCallArguments|TestForwardResponsesChatCompletionsFallbackKeepsFunctionArgumentsSingle|TestForwardAsChatCompletions_TransportErrorReturnsFailover|TestForwardAsRawChatCompletions_TransportErrorReturnsFailover|TestIsNonRetryableRefreshError|TestEnsureEmailAuthIdentityCreateErrorReturnsFalse" -count=1`
  - `git diff --check`
  - S22 denied-path audit over `git status --short` paths returned `NO_DENIED_PATHS`.
  - `go test ./internal/service -run "TestOpenAIImageOutputCounter|TestImagesOAuthNonStreaming_ContentRefusalReturns400NoRetry|TestExtractModelRefusal_EmptyWhenNoText|TestIsOpenAITransientProcessingError|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIStreamingPassthroughResponseFailedAfterOutputSanitizesVerboseResponseForClient|TestOpenAIGatewayServiceForward_CodexImageInjectionRespectsGroupCapability|TestOpenAIGatewayServiceForward_ChannelBridgeOverrideEnablesCodexInjection|TestOpenAIGatewayServiceForward_CodexBridgePreservesExistingToolChoice|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool" -count=1`
  - `go test ./internal/pkg/apicompat -run "TestResponsesToAnthropic_.*Tool|TestResponsesToAnthropic_Custom" -count=1`
  - `git diff --check`
  - S23 raw denied-path audit only命中本轮前已有的 `knowledge/*` 脏改；S23 提交前需以 staged audit 确认未纳入这些文件。
  - `go test -tags=unit ./internal/service -run "TestGetModelPricing_FallbackWarn|TestGetModelPricing_GLM52|TestAPIKeyService_Update_ReactivatesQuotaExhaustedWhenQuotaUnlimited|TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1`
  - `go test ./internal/handler -run "TestOpenAIRecordUsageInputsCarryQuotaPlatform" -count=1`
  - `go test ./internal/service -run "TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform" -count=1`
  - `git diff --check`
  - `go test ./internal/service -run "TestExtractSystemMessagesFromInput|TestApplyCodexOAuthTransform_ExtractsSystemMessages|TestApplyCodexOAuthTransform_JsonObjectKeepsJsonInstructionInInput" -count=1`
  - `go test ./internal/service -run "TestApplyCodexOAuthTransform_.*Instructions|TestExtractSystemMessagesFromInput" -count=1`
  - `git diff --check`
  - `go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic" -count=1`
  - `go test ./internal/server/routes -run "TestGatewayRoutesOpenAICountTokensPathIsRegistered|TestGatewayRoutesNonOpenAICountTokensPathStillRegistered" -count=1`
  - `go test ./internal/service -run "TestOpenAIGatewayService_ForwardCountTokensAsAnthropic|TestBuildOpenAIEndpointURL" -count=1`
  - `go test ./internal/handler -run "TestResolveOpenAIMessagesDispatchMappedModel|TestNewOpenAIModelMappedBodyCache|TestOpenAIGatewayHandler" -count=1`
  - `go test ./internal/server/routes -run "TestGatewayRoutes" -count=1`
  - `git diff --check`
  - `go test ./internal/pkg/openai -run "TestCodexBaseInstructionsForModel" -count=1`
  - `go test ./internal/service -run "TestDefaultCodexSynthInstructionsModelAware|TestApplyCodexOAuthTransform_GPT55SuppliesModelSpecificInstructions|TestApplyCodexOAuthTransform_CodexCLI_SuppliesDefaultWhenEmpty|TestApplyCodexOAuthTransform_NonCodexCLI_PreservesExistingInstructions|TestOpenAIGatewayServiceForwardGPT55InjectsModelSpecificInstructions" -count=1`
  - `git diff --check`
  - `go test ./internal/service -run "TestIsOpenAIContextWindowError|TestShouldFailoverOpenAIUpstreamResponseContextWindow502|TestOpenAIHandleErrorResponse_ContextWindow502KeepsMessageWithoutFailover|TestForwardAsChatCompletions_BufferedContextWindowResponseFailedReturnsErrorWithoutFailover|TestForwardAsChatCompletions_BufferedTransientResponseFailedTriggersFailover|TestForwardAsChatCompletions_StreamContextWindowResponseFailedReturnsErrorWithoutFailover|TestOpenAIStreamingContextWindowResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover" -count=1`
  - `go test ./internal/service -run "TestOpenAIStreamingNormalizesTerminalOutputFromDeltas|TestOpenAIStreamingNormalizesTerminalOutputToEmptyArray|TestOpenAIStreamingPreambleKeepaliveUsesDownstreamIdle|TestOpenAIStreamingPolicyResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient" -count=1`
  - `go test ./internal/service -run "TestForwardAsChatCompletions_EnforcesCodexCLIOnlyRestriction|TestForwardAsChatCompletions_OAuthDoesNotInjectDefaultInstructions|TestForwardAsChatCompletions_APIKeyPropagatesPromptCacheKeyInResponsesBody|TestForwardAsChatCompletions_TransportErrorReturnsFailover" -count=1`
  - `go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial|TestAPIKeyAuthIPRestrictionUsesForwardedClientIPWhenProxyTrusted" -count=1`
  - `go test ./internal/service -run "TestComputeValidityDaysSupportsSingularAndPluralUnits" -count=1`
  - `git diff --check -- backend/internal/service/payment_service.go backend/internal/service/payment_order_result_test.go`
  - `go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_MessageRoles|TestTransformClaudeToGeminiWithOptions_PreservesBillingHeaderSystemBlock" -count=1`
  - `git diff --check -- backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go`
  - `go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert,intercept,sql/execquery,sql/lock --idtype int64 ./schema` in `backend/ent`
  - `go test ./internal/service -run "Test.*Peak.*|Test.*Group.*Peak.*|Test.*Billing.*Peak.*|Test.*Gateway.*Peak.*|Test.*RecordUsage.*Peak.*" -count=1`
  - `go test ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestFulfillPaidOrder.*Redeem|TestPaymentRechargePackage|TestFirstRechargeBonus|TestMonthlyRecharge" -count=1`
  - `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex|TestOpsRealtimeRequestCanceled|TestOpsRealtime|TestGetConcurrencyStats|TestGetAccountAvailability|TestGetRealtimeTrafficSummary" -count=1`
  - `go test ./internal/service -run "TestListAllAccountsForOps|TestOps.*Concurrency|TestOps.*Availability|TestFilterCodexInput|TestApplyCodexOAuthTransform|TestOpenAIGatewayServiceForward_CodexBridge|TestOpenAIGatewayServiceForward_.*Image|TestOpenAIGatewayService_CodexImageGenerationBridge|TestGatewayService_StreamingKeepalive|TestGatewayService_StreamingReusesScannerBufferAndStillParsesUsage|TestDetachUpstreamContextIgnoresClientCancel|TestAccount_GetAnthropicAPIKeyAuthScheme|TestGatewayService_AnthropicAPIKeyPassthrough_BearerAuthScheme|TestGatewayService_AnthropicAPIKeyBearerAuthScheme|TestBuildUpstreamModelsRequestsForAPIKeyAccounts" -count=1`
  - `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListOpsAccountsForStats|TestAccountRepoSuite/TestListWithFilters" -count=1`
  - S52 unit-tag command attempted: `go test -tags=unit ./internal/service -run "TestRateLimitService_HandleUpstreamError_OAuth401SetsTempUnschedulable|TestRateLimitService_HandleUpstreamError_OAuth401NoRefreshTokenSetsError|TestTokenRefreshService_RefreshWithRetry_Antigravity|TestTokenRefreshService_RefreshWithRetry_ClearsTempUnschedulable" -count=1`; result: non-blocking known unrelated service unit compile baseline in billing tests (`ImageOutputPriceExplicit`, old `computeTokenBreakdown` / `calculateCostInternal` signatures).
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `git diff --check`
  - `rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .` no matches.
  - S45-S52 denied-path audit over `git diff --name-only origin/main..HEAD` returned `DENIED_PATH_AUDIT_PASS`.
  - `go test ./internal/service -run "TestIsNonRetryableRefreshError|TestTokenRefreshService_RefreshWithRetry|TestOpenAIGatewayServiceRecordUsage|TestOpenAIGatewayService_.*Mapped|TestOpenAIGatewayService_Forward" -count=1`
  - `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex" -count=1`
  - S53 `git diff --check` passed.
  - S53 conflict marker scan returned no matches.
  - S53 denied-path audit over `git diff --name-only origin/main..HEAD` returned `DENIED_PATH_AUDIT_PASS`.
  - `go test ./internal/handler -run "Test.*AvailableChannel.*Peak.*|Test.*Payment.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - `go test ./internal/handler/admin -run "Test.*Group.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/utils/apiKeyCapabilities.spec.ts"`
  - `git diff --check`
  - `go test ./internal/service -run "Test.*Payment.*|Test.*Refund.*|Test.*Order.*|TestComputeValidityDaysSupportsSingularAndPluralUnits" -count=1`
  - `go test ./internal/handler -run "Test.*Payment.*|Test.*Refund.*" -count=1`
  - `go test ./internal/handler/admin -run "Test.*Payment.*|Test.*Refund.*" -count=1`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/components/payment/__tests__/currency.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts"`
  - `git diff --check -- <S36 allowed paths>`
- S37 验证：
  - `go test ./internal/service -run "TestApplyCodexOAuthTransform_PreservesEncryptedReasoningAndStripsReasoningID|TestFilterCodexInput_PreservesReasoningItemsAndStripsReasoningIDs|TestNormalizeCodexModel_Gpt53|TestResolveOpenAIForwardModel|TestNormalizeOpenAIModelForUpstream|TestUsageBillingModelCandidatesPreserveGPT55ProModel|TestGetModelPricing_OpenAICompactAliasesFallback|TestCalculateCost_OpenAIGPT55ProUsesGPT55PricingPolicy|TestGetFallbackPricing_FamilyMatching|TestShouldAutoInjectPromptCacheKeyForCompat" -count=1`
  - `go test ./internal/service -run "TestOpenAIImageOutputCounter|TestImagesOAuthNonStreaming_ContentRefusalReturns400NoRetry|TestForwardAsChatCompletions_EnforcesCodexCLIOnlyRestriction|TestForwardAsChatCompletions_TransportErrorReturnsFailover|TestForwardAsRawChatCompletions_TransportErrorReturnsFailover|TestIsNonRetryableRefreshError|TestHandleStreamingResponsePassthroughDeduplicatesFunctionCallArguments|TestForwardResponsesChatCompletionsFallbackKeepsFunctionArgumentsSingle|TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey|TestStripCodexSparkImageGenerationToolFromRawPayload|TestOpenAIGatewayServiceForward_CodexImageInjectionRespectsGroupCapability|TestOpenAIGatewayServiceForward_ChannelBridgeOverrideEnablesCodexInjection|TestOpenAIGatewayServiceForward_CodexBridgePreservesExistingToolChoice|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InjectsCodexImageGenerationTool|TestOpenAIGatewayService_ForwardCountTokensAsAnthropic|TestBuildOpenAIEndpointURL|TestOpenAIQuotaPlatform|TestOpenAIGatewayServiceRecordUsage_.*QuotaPlatform|TestCodex.*Reasoning|Test.*Encrypted.*Reasoning|Test.*GPT55.*Pro|Test.*ModelMapping|TestCodexBaseInstructionsForModel|TestFilterCodexInput_PreservesReasoningItemsAndStripsReasoningIDs|TestNormalizeCodexModel_Gpt53|TestResolveOpenAIForwardModel|TestNormalizeOpenAIModelForUpstream|TestUsageBillingModelCandidatesPreserveGPT55ProModel|TestGetModelPricing_OpenAICompactAliasesFallback|TestCalculateCost_OpenAIGPT55ProUsesGPT55PricingPolicy|TestGetFallbackPricing_FamilyMatching|TestShouldAutoInjectPromptCacheKeyForCompat" -count=1`
  - `go test ./internal/handler -run "TestOpenAIRecordUsageInputsCarryQuotaPlatform|TestResolveOpenAIMessagesDispatchMappedModel|TestNewOpenAIModelMappedBodyCache|TestOpenAIGatewayHandler|TestOpenAIGateway.*CountTokens" -count=1`
  - `go test ./internal/server/routes -run "TestGatewayRoutes" -count=1`
  - `git diff --check -- <S37 allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S36 QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0142-payment-refund-s36-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0142-payment-refund-s36-qa.md`。
- S37 QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0142-openai-codex-gateway-s37-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0142-openai-codex-gateway-s37-qa.md`。
- S38a 验证：
  - 初始 contract 命令 `go test ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters|TestAccountRepoSuite/TestListWithFiltersGroupFilter" -count=1` 返回 `ok ... [no tests to run]`，不作为验收证据。
  - `go test -tags=integration ./internal/repository -run "TestAccountRepoSuite/TestListWithFilters" -count=1`
  - `git diff --check -- <S38a allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S38a QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0142-account-repo-count-s38a-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0142-account-repo-count-s38a-qa.md`。
- S39a 验证：
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/PaymentView.spec.ts"`，2 files / 22 tests passed。
  - `git diff --check -- <S39a allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S39a QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0142-frontend-api-base-s39a-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0142-frontend-api-base-s39a-qa.md`。
  - S40 验证：
  - `go test ./internal/repository -run "TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1`
  - `git diff --check -- <S40 allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S40 QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0143-claude-oauth-payload-s40-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0143-claude-oauth-payload-s40-qa.md`。
- S41 验证：
  - `go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_ReasoningModelOmitsInvalidArgs|TestBuildGenerationConfig_ReasoningModelOmitsUnsupportedParams|TestTransformClaudeToGeminiWithOptions_PreservesWebSearchAlongsideFunctions|TestTransformClaudeToGeminiWithOptions_MessageRoles" -count=1`
  - `git diff --check -- <S41 allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S41 QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md`。
- S42 验证：
  - `go test ./internal/repository -run "TestUsageLogRepositoryGetUserModelStatsUsesRequestedModel|TestUsageLogRepositoryGetModelStatsWithFiltersRequestTypePriority|TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1`
  - `git diff --check -- <S42 allowed paths>`
  - staged denied-path audit returned `NO_DENIED_PATHS` because no files were staged.
- S42 QA 结论：PASS，详见 `docs/workflow/worker-results/upstream-main-v0143-user-model-stats-requested-s42-result.md` 和 `docs/workflow/qa-reports/upstream-main-v0143-user-model-stats-requested-s42-qa.md`。
- S45 已合入 `main`：merge commit `d1bc3aa40 merge: add affiliate risk scanner alerts`，功能分支 head `41e1befc docs: align affiliate risk workflow status`。
- 下一合法动作：进入下一个已批准 Sprint，或按发布流程做 S45 上线前验证。
- 状态推进规则：`contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
