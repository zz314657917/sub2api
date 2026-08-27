---

## Group Model Match Auth Enforcement Addendum (S272)

- Every grouped API key must enter the existing model-aware resolver after a
  gateway handler parses the requested model. A middleware fast path must not
  bypass the effective group's administrator-owned `model_match_patterns`.
- A mismatch returns the existing HTTP 403 `NO_MATCHING_GROUP_ROUTE`; a match,
  an ungrouped key, multi-group routing, and pinned-account no-fallback behavior
  retain their existing service semantics.
- The repair is limited to the middleware guard and focused regression tests.
  Schema, persistence, group rule syntax, routing priority/weight/cooldown,
  billing, subscriptions, provider traffic, frontend, deployment, containers,
  shared data, commit, push and `outputs/**` are excluded. Contract:
  `docs/workflow/tasks/group-model-match-auth-enforcement-s272.md`.

## API Key Adaptive Route Breaker Addendum (S271)

- Keep the existing immediate `API Key + group` transient cooldown as the
  first-user protection. Add a multi-instance shared breaker scoped to group,
  routing type and normalized exact requested model so one broken model cannot
  disable unrelated traffic in the group.
- Three consecutive transient upstream failures open the shared breaker.
  Recovery is half-open with one probe and bounded adaptive cooldowns of 30
  seconds, 2 minutes, 10 minutes and 30 minutes; any successful response
  resets the streak, while ordinary business/client 4xx responses only release
  a probe and never count as group health failures.
- A sub-threshold failure streak expires after 30 minutes without another
  failure. Shared escalation accepts explicit upstream-transient markers and
  final 502/503/504/529 only; ambiguous bare 429/500 responses retain the
  existing per-Key cooldown but do not poison shared group health.
- Redis state changes are atomic and use Redis time. Cache failure is fail-open;
  pinned accounts retain strict no-fallback, and all skipped routes must not
  silently fall back to the excluded default group.
- Same-request cross-group body replay, admin UI, schemas, billing, quota,
  subscriptions, provider traffic, containers, shared data, commit, push and
  `outputs/**` remain out of scope. Contract:
  `docs/workflow/tasks/api-key-adaptive-route-breaker-s271.md`.

## Pixel Cafe Purchase Information And Round Controls Addendum (S270)

- Public purchase details expose the existing per-share managed-Key total,
  5H/1D/7D limits and display label. Zero means unlimited; a 30-day total is
  not represented as a monthly-reset quota.
- Existing `sort_order` becomes the administrator-facing priority and the
  default public/admin ordering source (`ASC`, then room ID). The legacy
  `featured` field remains wire-compatible but no longer outranks priority.
- An enabled room with no live round can be opened. An `open` round can be
  paused only while atomically proven empty of locked/paid seats and paid
  memberships; pausing closes it as `cancelled`, performs no refund/order
  mutation, and permits a later fresh round. Other live states are display-only
  in this action column.
- Schema/migrations, payment/refund execution, billing, providers, shared data,
  containers, deployment, commit, push and `outputs/**` remain out of scope.
  Contract: `docs/workflow/tasks/pixel-cafe-purchase-controls-s270.md`.

## Upstream OpenCode Go Reset Duration Addendum (S262)

- Adapt only the parser portion of upstream `a6b11ccce`: an OpenCode Go
  `GoUsageLimitError` with a valid `Resets in ...` message supplies the
  temporary account-rate-limit expiry already consumed by the local generic
  429 pipeline.
- Parse bounded compound durations after the marker and return nil for unknown,
  malformed, zero, negative, or overflowing input so current fallback behavior
  remains unchanged. Do not import upstream runtime blocker APIs absent from
  this topology.
- Scheduler/account state, rate-limit persistence, provider calls, schemas,
  APIs, frontend, dependencies, containers, shared data, staging, push, and
  `outputs/` are out of scope. Contract:
  `docs/workflow/tasks/upstream-opencode-go-reset-s262.md`.


## Upstream Composite Billing Fallback Addendum (S261)

- Adapt upstream `ba88cc239` in the local `GatewayService` billing owner: a
  composite public alias selected by `BillingModelSource` uses the concrete
  forwarded model unless an administrator configured an explicit price for the
  alias.
- Preserve local explicit group-model pricing as well as channel pricing; both
  are administrator-owned overrides. Apply the composite guard before general
  price resolvability so family-name aliases cannot inherit an unrelated
  built-in fallback price.
- Outside composite groups, only a completely unresolvable selected billing
  model may fall back to a resolvable concrete forwarded candidate. Keep
  mappings, price tables, usage attribution, image/video behavior, schemas,
  APIs, frontend, providers, containers, shared data, staging, push, and
  `outputs/` out of scope. Contract:
  `docs/workflow/tasks/upstream-composite-billing-fallback-s261.md`.


## Pixel Cafe My-Room Usage Progress Addendum (S260)

- Active “我的包间” cards show two accessible progress bars only after an
  administrator has assigned an account and activation produced a valid
  managed Key: `账号 7D 剩余` uses a safe remaining percentage derived from the
  assigned OpenAI account's official cached 7D snapshot, while `我的限额`
  uses the member Key's existing 7D used/limit window. The 5H projection
  remains available to the private DTO but is temporarily hidden from this
  compact card.
- The private my-room DTO adds safe activation/expiration timestamps and
  derived future 5H/7D reset timestamps. Raw window starts, credentials, Key
  material and full account identity remain excluded.
- The compact card renders only room name, assigned account name, remaining
  validity, `账号 7D 剩余` and `我的限额`. It omits room code, shares/days, status badge,
  platform, masked email, Key metadata, total quota, exact expiry and reset
  copy. Unlimited, unopened, expired and missing-Key states use short safe
  fallback copy.
- No schema, migration, billing write, enforcement, scheduler, admin, public
  lobby, container, shared-data, commit or push behavior changes. Contract:
  `docs/workflow/tasks/pixel-cafe-my-room-usage-s260.md`.

## Pixel Cafe Room List Scene Overlay Addendum (S256)

- The public room list is rendered once inside the existing Pixel Cafe lobby
  scene instead of in a separate block outside the background.
- Desktop uses a translucent right-side list with internal vertical scrolling;
  mobile uses a compact bottom horizontal card strip contained by the scene.
- Existing room metadata, member avatars, details dialog, share purchase,
  loading/error/empty/demo behavior, Pixi/fallback rendering and reduced-motion
  behavior remain unchanged.
- Backend, state machines, admin/settings, images, dependencies, containers,
  shared data, commit, push, unrelated dirty files and `outputs/` remain outside
  this change. Contract:
  `docs/workflow/tasks/pixel-cafe-room-list-scene-overlay-s256.md`.

## Pixel Cafe Variable Workstation Count Addendum (S255)

- The shared lobby layout supports 1 through 50 contiguous numbered computer
  workstations instead of exactly ten; ten remains the missing-setting default.
- The administrator can change the count before saving. Shrinking removes only
  the highest IDs, growing preserves existing coordinates and appends
  deterministic editable positions, and reset keeps the selected count.
- The existing layout array remains the sole persisted source of truth. Public
  Pixi/fallback rendering and seated-avatar capacity follow its validated
  length; no schema, migration, second setting key, or private projection is
  added.
- Current S252-S254 dirty changes, room/share/account behavior, dependencies,
  images, containers, shared data, commit, push, and `outputs/` remain outside
  this change. Contract:
  `docs/workflow/tasks/pixel-cafe-variable-workstation-count-s255.md`.

## Pixel Cafe Workstation Layout Editor Addendum (S254)

- Administrators can drag the ten numbered computer workstations directly on a
  lobby preview, reset the unsaved draft, cancel it, or save one server-backed
  layout shared by all users. Browser-local storage is not authoritative.
- The backend accepts only ten unique IDs covering `1..10` with bounded finite
  coordinates, and public settings expose only those IDs and coordinates.
- The lobby bitmap, editor, Pixi renderer, fallback, and seated-avatar mapping
  share a `960x540` 16:9 design space and cover transform. Missing or malformed
  persisted data falls back to the built-in layout.
- Room/share/account lifecycle, database schema, dependencies, image assets,
  containers, deployment, commit, push, unrelated dirty files, and `outputs/`
  remain outside this change. Contract:
  `docs/workflow/tasks/pixel-cafe-workstation-layout-s254.md`.

## Pixel Cafe Share Fulfillment Addendum (S252)

- Pixel Cafe V2 sells 1-10 configurable shares rather than fixed seats. A Room
  Plan supports only normalized ChatGPT `plus` or `pro`, has a configurable
  distinct-buyer cap (default four), and lets an existing participant top up
  while the Round remains open.
- New Rooms/Rounds have no pre-bound Account. Paid-full changes the Round to
  `awaiting_account`; an administrator then assigns a matching active OpenAI
  Account. Successful activation creates one managed Key per `(round,user)`
  membership with per-share limits multiplied by paid shares, and starts the
  validity window at activation.
- Unfulfilled paid-full Rounds enter idempotent refund handling after 24 hours.
  Public APIs expose aggregate shares/buyers and pseudonymous avatars only;
  Account details remain private until the requesting member is active.
- The existing scene, custom Room/page copy, generic group-buy lifecycle, legacy
  Cafe bindings, current user dirty work, containers, deployment, and push stay
  outside the behavior change. Contract:
  `docs/workflow/tasks/pixel-cafe-share-fulfillment-s252.md`.
repo: sub2api
project_type: web
qa_mode: runtime
last_verified: 2026-08-16 02:41 +08:00
---

## Upstream Chat File Input Addendum (S246)

- Adapt `4d4a0be1a`/`6244090c1` so Chat Completions `type:"file"` content
  parts map to Responses `type:"input_file"`, forwarding `filename`,
  `file_data`, and `file_id`. Parts with neither payload field remain skipped.
- Preserve text/image ordering and conversion, empty-content fallback, custom
  tools, output conversion, and S239 streamed empty tool-name omission. Keep
  the local DTO/converter topology rather than replaying divergent history.
- Scope is exactly three `apicompat` owners plus workflow evidence. File upload/
  download products, validation/MIME/size policy, gateway/security-audit,
  frontend, schema/migrations, dependencies, provider traffic, containers,
  deployment, push, current Pixel Cafe dirty paths, and `outputs/` are excluded.
- Acceptance requires focused x10, complete `apicompat`, service/server compile,
  format/diff, exact scope, ancestry, conflict/index, protected-primary gates,
  and independent Terra QA.

## Upstream OpenAI Sticky System Prefix Addendum (S245)

- Adapt `e45490a36` from upstream merge `2ddda6735` to the local direct
  `gjson` content-seed scanner. Chat messages contribute only their leading,
  contiguous system/developer prefix; any later dynamic system/developer
  message after conversation history starts must not change sticky identity.
- Preserve the first user message, model, tools/functions, instructions,
  canonical JSON, Responses input handling, explicit session hints, scheduler,
  cache keys, TTLs, and hash format. Do not import upstream's unrelated
  single-scan refactor.
- Scope is exactly the content-session-seed owner, its focused tests, and
  workflow evidence. Other backend/frontend paths, schema/migrations,
  dependencies, providers, containers, deployment, push, current Pixel Cafe
  dirty paths, and `outputs/` are excluded.
- Acceptance requires focused x10, the complete seed regression set, complete
  service package, server compile, formatting/diff, exact scope, provenance,
  conflict/index, protected-primary gates, and independent Terra QA.
- `219368ec6` is deferred rather than partially ported: its apparent two-file
  Composite video gate depends on upstream Composite Resolver and Grok media
  handlers that are absent from the local asynchronous OpenAI video topology.

## Upstream Token Refresh Lock Addendum (S244)

- Remove the proactive-refresh shortcut that accepts an unchanged access
  token, refresh token, expiry timestamp, and user identity as peer-refreshed
  state merely because the expiry remains beyond a two-minute buffer.
- Preserve actual peer adoption when the rotating refresh token changes and the
  existing failed-access-token reconciliation proves a replacement. Keep
  same-document sharing, refresh-token race recovery, cross-user isolation,
  logout protection, storage keys, timeouts, and API contracts unchanged.
- Scope is exactly `tokenRefresh.ts`, its focused Vitest owner, and workflow
  evidence. Backend, other frontend paths, dependencies, configuration,
  browser automation, deployment, containers, production state, push, current
  Pixel Cafe dirty paths, and `outputs/` are excluded.
- Acceptance requires the boundary-jitter regression, focused x10, frontend
  typecheck/build, exact scope, provenance, diff/index/conflict, and protected
  primary-worktree gates under independent Terra QA.

## Upstream v0.1.177 Codex Turn-State (S219)

- Relay `x-codex-turn-state` explicitly across native OpenAI HTTP streaming,
  JSON, SSE-to-JSON, and passthrough response paths without widening the global
  response-header filter.
- Track the account that minted a state only after the downstream response
  headers are actually committed, keyed by positive API-key ID plus the
  original client session. Strip only known cross-account client echoes in the
  normal and passthrough request builders; never inject native HTTP state.
- Adapt `8219dcfc8` and `4d9fedee2` plus only the two turn-state guard hook ideas
  from `fce41e318`. Fingerprint defaults/convergence, frontend, migrations,
  dependencies, provider traffic, containers, deployment and push are excluded.
  Contract: `docs/workflow/tasks/upstream-v0177-turn-state-s219.md`.
- Independent Terra QA and post-integration main regressions passed. Local main
  contains implementation commits `2335470c0` and `f347aa460`, worker evidence
  `590921da2`, and QA report `c3e000df0`. Provenance is recorded only after the
  first successful stream flush or a committed non-streaming writer; nil/empty
  upstream headers clear stale state and native HTTP remains stripping-only.

## Upstream v0.1.177 Remaining Candidate Decision

- `e29b93a1f` is behaviorally covered by the local Grok unknown-text fallback,
  which already excludes image/video/audio/voice/search/media families.
- `e215c98c2` is behaviorally covered because saved account auto-refresh state
  is restored at module initialization before `onMounted` starts the timer.
- `fd82dfd52` cannot be ported independently: local billing has neither the
  upstream group long-context toggle nor the OpenAI account veto it corrects.
  The local default Grok ladder remains active, but the upstream configurable
  gate feature is prerequisite-absent.
- The rest of `fce41e318` requires the missing Codex fingerprint convergence
  subsystem and touches the user-owned account modal. `cb7b03795` and follow-up
  fixes require migrations 222/223 and explicit database-impact authorization.
  The VERSION-only commit remains excluded from this selectively diverged
  product line.

## Upstream v0.1.177 Remote Compaction V2 (S218)

- Adapt the final behavior of upstream `9662cff2e`, `a8b9ea22b`, and
  `8ae6d8f67` to the local monolithic OpenAI gateway. Native streaming
  `compaction_trigger` requests stay on `/responses`; legacy compact remains a
  separate compatibility path.
- Native v2 requires Responses capability without legacy compact eligibility or
  compact-only mapping. HTTP/WS requests carry the correct session-level
  `x-codex-beta-features`, while client-declared feature sets are preserved.
- The account compact probe uses local-fixture native v2 and requires a real
  compaction output item. Turn-state, fingerprint convergence, group rollup
  migrations, frontend, dependencies, provider traffic, containers, push, and
  deployment are excluded. Contract:
  `docs/workflow/tasks/upstream-v0177-remote-compaction-v2-s218.md`.
- Independent Terra QA and post-integration main regressions passed. Local main
  contains implementation commits `2058b69c9` and `32c55f9fe` plus QA report
  commit `d6c7435bd`; no provider, migration, container, deployment or push
  operation was performed.

## Upstream GPT/Codex Quota Correctness (S217)

- Preserve personal subscription expiry when workspace entitlement differs,
  skip OpenAI account penalties for HTML 403 bodies, and make reset-credit UI
  state explicit without an automatic second quota request.
- Add an audited POST quota refresh route while retaining read-only GET. Keep
  existing S188 recovery-first semantics and local quota-threshold equivalents.
- Contract and QA:
  `docs/workflow/tasks/upstream-v0176-gpt-quota-s217.md` and
  `docs/workflow/qa-reports/upstream-v0176-gpt-quota-s217-qa.md`.

## Standard Group Time-Window Rate (S211)

- Allow standard and subscription groups to apply the existing daily
  `peak_rate_*` factor to token billing. The factor multiplies the already
  resolved effective rate and returns to `1.0` outside `[start, end)`, so the
  final rate returns to the original group/user/membership result without
  disabling the group.
- Standard groups require an enabled factor greater than zero; subscription
  groups retain zero-factor compatibility. Windows use the configured server
  timezone, remain same-day only, and are evaluated from request start time
  across retry, failover, asynchronous recording, and WebSocket turns.
- Reuse the existing schema and API fields. Per-request image/video billing,
  routing, group status, API-key bindings, migrations, dependencies,
  containers, deployment, push, and production traffic are excluded. Contract:
  `docs/workflow/tasks/standard-group-time-rate-s211.md`.

## Upstream Streaming and Audit Fixes (S210)

- Adapt the compact keepalive behavior from `2f109e74c`: a Responses stream whose keepalive committed `200` headers but emitted no semantic SSE payload must receive one protocol-correct `response.failed` terminal event when forwarding fails. A stream that already emitted semantic output must not receive a duplicate terminal error.
- Adapt the WebSocket audit behavior from `c418fd522`: only an identical `(stage, turn, payload hash)` audit result with `DecisionAllow` may be reused. A new turn, changed payload, block, unavailable, invalid, or flag decision must re-evaluate normally.
- Scope is the OpenAI gateway error boundary, audit helper, focused handler regressions, and workflow evidence. Route cooldown policy, billing, persistence, schema/migrations, configuration, frontend, dependencies, containers, provider calls, push, deployment, and production traffic remain excluded. Contract: `docs/workflow/tasks/upstream-streaming-audit-s210.md`.

## Streaming Route Cooldown (S208)

- Preserve multi-group route cooldown semantics when an already-started Gateway or OpenAI SSE stream emits a terminal `429`, `529`, or `5xx` error. The handler records the existing cooldown classification in request-local Gin state; API-key middleware consumes it after the handler returns, even though the HTTP writer is already `200`.
- A started stream is never replayed or redirected. The current client still receives its protocol-compatible terminal SSE error; only a later request skips the cooled group and follows the configured priority/weight route selection.
- Scope is two handler call sites, the API-key cooldown boundary, and focused regression tests. Route configuration, route selection policy, Redis/cache protocol, billing, account scheduling, schema/migrations, frontend, deployment, containers, provider calls, push, and production traffic are excluded. Contract: `docs/workflow/tasks/streaming-route-cooldown-s208.md`.

## Upstream OpenAI OAuth Routing Hints (S206)

- Port upstream `915cc7e7b`, `815035fcc`, and `de349187d` into the local monolithic OpenAI HTTP/WS topology: remove the legacy OAuth beta-header injection, generate a gateway-owned final-model routing hint, and keep WebSocket routing affinity advisory rather than continuation-breaking.
- Apply upstream `8ad0a5ff5` exactly to update the audited `nanoid` lockfile entry from `3.3.16` to `3.3.17`.
- Preserve local identity, Fast/Flex policy, billing, session isolation, fixed-account routing and retry behavior. Schema, migrations, configuration, deployment, containers, provider calls, production traffic, push and unrelated `v0.1.172` candidates are excluded. Contract: `docs/workflow/tasks/upstream-openai-routing-hints-s206.md`.

## GPT-5.6 Pricing Metadata Integration (S205)

- Apply the reviewed upstream GPT-5.6 Luna/Terra pricing metadata commit onto the latest local `main`, retaining
  the upstream source record without merging unrelated upstream history.
- Restore the dynamic pricing parser's three long-context fields and retain the new cache-write Batch/Flex
  metadata. Existing standard/priority/flex billing behavior remains unchanged; no Batch billing path is added.
- Acceptance requires real-fixture parser assertions for Sol/Terra/Luna, focused billing and WebSocket
  regressions, formatting and Git integrity checks. No push, deployment, container, database, frontend, or real
  provider operation is allowed. Contract: `docs/workflow/tasks/gpt56-pricing-metadata-s205.md`.

## Upstream OpenAI Account+Model Transient Breaker (S203)

- Port the behavior of upstream `40b8f04a6` and its sparse-traffic correction `7d38e6712`: retryable OpenAI
  failures must create an in-process, bounded account+model streak; a success clears only that pair; selection
  must skip a pair during its short/long cooldown while leaving the same account eligible for other models.
- The retained streak must expire only after a 30-minute inactivity TTL, not the previous one-minute reset window.
  This avoids low-traffic deployments permanently reselecting a consistently failing account before every
  eventual failover.
- Scope is gateway/handler/service propagation and focused regressions only. Configuration, dependencies,
  persistent state, Ent/schema/migrations, frontend, provider credentials, containers, deployment and production
  traffic are excluded. Contract: `docs/workflow/tasks/upstream-openai-account-model-transient-s203.md`.

## Pixel Cafe Phase 30 Addendum: configurable presentation chrome (S176)

- Remove the user-facing "今日使用用户" lobby card from the Pixel Cafe page and stop its
  page-local polling. Keep the existing anonymous lobby service and persistence behavior intact;
  this is a presentation-only removal.
- Add three settings through the existing public/admin settings chain: a title, an optional
  description, and a header-visibility toggle. Existing installations with no new keys retain the
  current title "像素网吧", description "把每个模型分组变成一间可订阅的数字包间。", and a visible
  header. A disabled header hides the eyebrow, title, and description together.
- The administrator Settings page owns the controls. Values must use normal Vue text interpolation;
  no HTML rendering, schema/migration, room/order/payment, lobby API, or provider changes are in scope.
- Acceptance requires focused settings/public-injection and PixelCafePage regressions, frontend
  typecheck/build, an actual local browser screenshot after guarded container promotion, and Git
  integrity checks. No remote deployment, push, image/volume pruning, or production setting update.

### S176 QA Result

- `BLOCKED / browser-tool`: focused Go/Vitest regressions, typecheck, production build, guarded local
  image promotion, `/health`, PostgreSQL/Redis health, and public-settings HTTP fields passed. The
  required browser screenshot could not be captured because Playwright Chrome exited, isolated
  Chromium installation timed out, the in-app browser blocked localhost, and the Chrome extension
  transport closed before tab creation. No browser or production setting was changed.

## Pixel Cafe Phase 28 Addendum: registered gateway forwarding and usage attribution (S174)

- A disposable PostgreSQL integration test must exercise the actual `RegisterGatewayRoutes`
  and API-key middleware through one active managed Cafe Key, then forward a non-streaming
  Anthropic API Key request to a loopback-only synthetic terminal using the real gateway and
  HTTP upstream path.
- The selected Account, upstream request authentication, and durable `usage_logs.account_id`
  must all equal the Room Binding's fixed Account. An expired Binding or unschedulable pinned
  Account must fail closed before upstream or usage mutation.
- No real provider, merchant, payment sandbox, existing Key, shared database/Redis, deployment,
  container update or production write is permitted. Contract:
  `docs/workflow/tasks/pixel-cafe-phase28-gateway-usage-s174.md`.

### S174 QA Result

- `PASS / runtime-isolated`: fresh PostgreSQL plus actual registered gateway routes, API-key middleware,
  pinned Account selection, repository HTTP transport and a loopback-only Anthropic terminal proved that
  the request authentication and durable `usage_logs.account_id` both equal the Room Binding Account.
  Expired Binding and disabled Account fail closed before upstream/usage mutation; the core test passed
  once and across three fresh reruns. QA:
  `docs/workflow/qa-reports/pixel-cafe-phase28-gateway-usage-s174-qa.md`.
- This evidence is implementation-level, not a real provider, payment sandbox, capacity, staging,
  deployment, rollback or production gate. A generic best-effort usage batch-state decode warning is
  deliberately tracked separately and does not change the persisted Cafe attribution result.

## Pixel Cafe Phase 29 Addendum: best-effort usage batch state (S175)

- S174 observed a type mismatch in the generic best-effort usage batch state payload: the synthetic
  SQL `input_idx` reaches `json_build_object` as a string while Go decodes it as an integer. The
  existing synchronous fallback preserves durable usage, but it turns the normal batch path into an
  avoidable warning and synchronous write.
- S175 may correct only this synthetic input-index SQL type and add a regression. It must retain all
  usage business fields, idempotency, billing, schema and gateway/provider behavior. Contract:
  `docs/workflow/tasks/pixel-cafe-phase29-usage-batch-state-s175.md`.

### S175 QA Result

- `PASS / runtime-isolated`: the state-only best-effort query now casts its synthetic `input_idx`
  parameter as PostgreSQL `integer`; the focused repository regression and adjacent route tests passed.
  The registered Gateway usage integration passed once and across three fresh PostgreSQL reruns with
  pinned Account authentication, durable `usage_logs.account_id`, Binding/Account fail-closed checks,
  and no `best-effort batch state decode failed` warning. Formatting, diff and unmerged-index checks
  passed; Docker's existing nine-container stack was unchanged and disposable testcontainers were
  cleaned up. QA: `docs/workflow/qa-reports/pixel-cafe-phase29-usage-batch-state-s175-qa.md`.
- This closes the generic batch-state type regression only. Real provider/payment, performance,
  staging, deployment, rollback and production readiness remain outside this runtime-isolated gate.

## Pixel Cafe Phase 27 Addendum: isolated JWT, gateway and Redis cross-instance smoke

### Goal

- Combine the existing Cafe JWT My Rooms boundary, managed-Key fixed-account auth boundary, and cache invalidation behavior in fresh PostgreSQL/Redis Testcontainers without any external provider call.

### Scope Boundary

- Use real Cafe user-route registration/JWT, real API-key auth plus gateway auth preflight, two APIKeyService instances, real Redis L2/Pub/Sub, and synthetic temporary entitlement facts only.
- The terminal route returns locally after auth. It neither invokes `RegisterGatewayRoutes` provider handlers nor makes any provider, merchant or payment-sandbox request.
- Exclude all production source, schema/migration/generated code, existing Key/Binding, shared containers/data, deployment, performance, staging, rollback and production validation.

### Acceptance Boundary

- An authenticated temporary user can access only its own My Rooms facts, while unauthenticated traffic is rejected before Cafe data.
- A temporary active managed Key reaches the local preflight with its one fixed Account pin. Cross-instance Redis invalidation must cause inactive rejection, valid re-enable recovery, and Binding-expiry fail-closed behavior through real middleware requests.
- Disposable PostgreSQL, Redis, client and subscription resources are bounded and terminated; Docker before/after evidence records no impact to the existing local stack.

### S173 QA Result

- `PASS / runtime-isolated`: the sole new integration test passed once and across three consecutive fresh PostgreSQL/Redis runs. It exercised real JWT My Rooms ownership, API-key/gateway auth preflight, L1/L2/Pub/Sub invalidation, inactive/re-enable and Binding-expiry fail-closed behavior with only synthetic facts and a local terminal response.
- Provider traffic, payment sandbox, performance, staging, deployment, rollback and production validation remain excluded.

## Pixel Cafe Phase 12 Addendum: last-Seat PostgreSQL transaction verification

### Goal

- Verify the existing Cafe last-Seat order transaction under concurrent PostgreSQL row locking, using a temporary database and without invoking payment providers.

### Scope Boundary

- Invoke only `CafeRoomOrderService.lockSeatAndCreateOrder` with distinct active users against one open Cafe Round containing one Seat, then verify the committed Seat/Order/Round relations.
- This Sprint does not validate a provider request, payment callback, paid-full activation, refund/expiry/migration lifecycle, managed Key enablement, HTTP/JWT, Redis, gateway routing, configuration, deployment, or production behavior.

### Acceptance Boundary

- A fresh Testcontainers PostgreSQL instance is populated by current Ent schema creation and destroyed by test cleanup.
- At least 16 concurrent transaction calls leave exactly one locked Seat, one pending Order and one reserved Round capacity; every losing call returns the specific Cafe seat-unavailable conflict rather than an infrastructure error.

## Pixel Cafe Phase 11 Addendum: isolated Testcontainers concurrency and Redis verification

### Goal

- Verify the existing Room repository locking and anonymous Lobby Redis persistence against fresh, Testcontainers-managed PostgreSQL and Redis instances, without connecting to a configured application environment.

### Scope Boundary

- Cover only two repository races: enabled Room assignment to one Account, and open Round creation for one Room. Cover the public Lobby recorder/snapshot projection using the harness's namespaced Redis client.
- This Sprint adds no user HTTP/JWT route, payment/provider invocation, managed-Key enablement, gateway/auth-cache routing, Account health check, schema/migration, feature behavior, configuration, deployment, or production-container action.
- A last-Seat order/payment race and full activation/expiry/refund/migration lifecycle require later service-level Testcontainers contracts; they are not inferred from these checks.

### Acceptance Boundary

- The existing integration TestMain creates and cleans isolated PostgreSQL 18.1 and Redis 8.4 containers, applies migrations, and namespaces all Redis facts.
- PostgreSQL evidence proves one committed result for each asserted Room/Round race; Redis evidence proves anonymous bounded projection and 72-hour TTLs after real persistence.
- Existing local application containers, credentials, data, and configuration are not read or written.

## Pixel Cafe Phase 10 Addendum: Pixi scene foundation

### Goal

- Replace the CSS-only Pixel Cafe room-grid scene with a route-lazy PixiJS enhancement that uses a local static pixel scene/layout, anonymous Lobby avatar anchors, bounded Room hotspots, DOM controls, and a deterministic fallback.

### Scope Boundary

- Render only existing public Cafe overview/Room/Lobby facts. Canvas has no private data, API requests, payment state, Key/account fields, live-presence meaning, or business-only controls.
- PixiJS v8 is isolated in `features/pixelCafe/renderer/`; it is dynamically imported, its canvas/ticker/listeners are disposed on unmount, and a failed renderer preserves a keyboard-accessible DOM navigator.
- Use a self-authored local scene asset and layout manifest. Do not import third-party artwork, remote assets, a tilemap engine, WebSocket, free controls, or a second rendering library.
- Do not modify backend, API/types, schema/migration, Redis, Key enablement, auth/gateway/provider/payment behavior, configuration, deployment, or `knowledge/**`.

### Acceptance Boundary

- Focused renderer/page tests prove lazy initialization, bounded redacted data, hotspot selection, accessible DOM navigation, reduced-motion behavior, fallback, and destruction.
- Frontend lint, typecheck, production build, lockfile review, integrity checks, plus mocked browser desktop/mobile screenshots and teardown checks pass.

### S156 QA Result

- `PASS / source-level + mocked-browser`: focused tests, lint, typecheck, an 1844-module production build, and local synthetic-response browser evidence passed. Canvas hotspot selection, DOM keyboard selection, reduced-motion, forced renderer fallback, 390px overflow containment, and route teardown were checked.
- This result does not validate real JWT/API behavior, PostgreSQL, Redis, payment/provider, managed-Key routing, gateway usage, performance, deployment, or production readiness.

## Pixel Cafe Phase 9 Addendum: today usage lobby activity

### Goal

- Implement the Pixel Cafe "今日使用用户" lobby facts from successful, persisted usage logs: a Redis day-scoped user ZSET and request counter, a strictly anonymous user API projection, overview aggregation, and visible-tab 60-second frontend refresh with a harmless Redis failure fallback.

### Scope Boundary

- Count only a usage log that was actually inserted; idempotent conflicts, failed upstream calls, browser page visits, and unauthenticated requests must not increment the lobby.
- Redis holds internal user IDs only. Public responses expose date, timezone, counts, and bounded HMAC-derived avatar seeds with coarse activity buckets; they never expose user, API-key, Account, request, IP, email, timestamp, credential, or source-key fields.
- Redis failure must neither fail usage logging/billing nor Cafe room discovery. The activity endpoint and overview return an explicit unavailable/empty lobby projection instead.
- Do not add Pixi, canvas assets, WebSocket, live-presence semantics, API Key enablement, Account routing, payment, schema/migration, deployment, or production configuration changes.

### Acceptance Boundary

- Repository regression proves only inserted usage records update the date-scoped ZSET/counter with 72-hour TTL, and Redis write failure leaves usage-log persistence successful.
- Service/handler regressions prove deterministic daily opaque seeds, no PII/raw IDs/timestamps, bounded display, correct date/timezone wording, endpoint/overview degradation, and feature-flag behavior.
- Frontend regression proves initial lobby fetch, 60-second polling only while the document is visible, safe unavailable display, and no "online" terminology; typecheck/build remain green.

## Pixel Cafe Phase 0 Addendum: audit and architecture baseline

### Goal

- Import the Pixel Cafe development package into `docs/pixel-cafe/` and audit the current repository before any business implementation.
- Produce the Phase 0 ADR, lifecycle/auth call graphs, source-level file plan, baseline test evidence and explicit stop-at-review decision.

### Scope Boundary

- Documentation and read-only audit only; no backend/frontend/schema/migration/generated/deployment/container changes.
- Preserve the existing S134/S138/S133 handoff and dirty paths; do not fetch, push, commit, deploy or update containers.
- Do not enter Phase 1 until the Phase 0 report is reviewed.

### Acceptance Boundary

- Package copy matches the current download source by file list and SHA-256.
- Current route, GroupBuy lifecycle, UserSubscription/auth-cache dependency, fixed-account insertion layer and provider bypass risks are evidenced from source.
- Backend/frontend baseline commands and diff checks are executed and reported truthfully.
- Allowed-path review shows only `docs/pixel-cafe/**` and this Sprint's workflow artifacts changed.

### Contract

- `docs/workflow/tasks/pixel-cafe-phase0-s139.md`

## S138 Addendum: hide empty user subscription panel

### Goal

- Remove the subscription panel from the user Usage layout when the active-subscription request completes with no rows.

### Scope Boundary

- Keep loading feedback while the request is pending and preserve the existing cards and renewal action for non-empty results.
- Do not change subscription APIs, backend semantics, routing, translations, billing, deployment, containers, or unrelated Usage behavior.

### Acceptance Boundary

- Focused UsageView regression proves the empty panel is absent while usage analytics remain visible; the existing non-empty subscription and renewal regression stays green.
- Frontend typecheck, production build, diff, conflict-marker, unmerged-index, and allowed-path checks pass.

## S131 Addendum: publication receipt for S128-S130

### Goal

- Record the verified `origin/main@ef5881c6b` publication of the selective
  S128 compatibility port, S129 handoff, and S130 regression repair.

### Scope Boundary

- Change only workflow and handoff evidence, then push the resulting
  documentation-only receipt. Do not alter source, dependencies, deployment,
  containers, provider state, or runtime behavior.

### Acceptance Boundary

- Cached paths stay within the S131 allowlist and final local, tracking, and
  remote main refs match after push. `outputs/` remains untracked.

## S130 Addendum: recent-commit regression repair

### Goal

- Correct the confirmed OpenAI capacity retry, group-buy refund lifecycle,
  Grok tool sanitization, and leaderboard cold-start guard regressions without
  changing public route contracts or persistence schema.

### Scope Boundary

- Modify only the existing gateway/service and router seams plus focused
  regressions. Reuse the established refund-pending, needs-review, event, and
  same-account retry mechanisms.
- Preserve the S129 local-integration record and untracked `outputs/`. Exclude
  migrations, payment-provider calls, deployment, containers, remote Git, and
  unrelated source changes.

### Acceptance Boundary

- Source-level tests demonstrate the repaired state transitions and payload
  behavior; compile/typecheck, formatting, diff, conflict-marker, and
  allowlist checks pass. External provider and browser runtime behavior remain
  explicitly unverified.

## S129 Addendum: S128 local-integration record alignment

### Goal

- Make workflow and handoff records accurately reflect that the selective
  `v0.1.168` S128 compatibility port has been committed and merged locally,
  while remaining unpublished.

### Scope Boundary

- Update only workflow evidence and the current-task handoff. Record
  implementation commit `85439ff50`, merge commit `fbf4ea10e`, and
  `origin/main@49af8e1bb` without changing any source or runtime behavior.
- Do not fetch, push, deploy, update containers, or add external smoke claims.

### Acceptance Boundary

- Workflow status, specification, main log, and handoff agree on the local
  commit relationship, `outputs/` remains untracked, and only contract-allowed
  documentation paths appear in the diff.

## S128 Addendum: selective v0.1.168 protocol compatibility ports

### Goal

- Adapt four isolated `v0.1.168` compatibility fixes to local architecture
  without merging the tag: preserve a migrated Claude OAuth system cache
  breakpoint, generate Anthropic-compatible synthetic message IDs/schema,
  preserve `max` reasoning effort for GPT-5.6 Messages bridging, and display
  the Claude Sonnet 5 status alias.

### Scope Boundary

- Port behavior, not upstream file layout or commit history. The local tree
  combines several upstream files into `gateway_service.go` and
  `openai_gateway_messages.go`.
- Preserve existing Claude OAuth cache-control limits, account selection,
  billing, model mapping, response routing, and OpenAI compatibility logic.
- Do not change databases, migrations, Passkey/WebAuthn, Kimi, Model Plaza,
  Prompt Audit, deployment, container configuration, or server versioning.

### Acceptance Boundary

- Focused Go regressions prove each new behavior and prove adjacent legacy
  behavior remains intact. Backend compile, frontend alias regression and
  typecheck, formatting, diff, unmerged-index, and allowed-path gates pass.
- Real OpenAI, Anthropic, Gemini, and Antigravity upstream calls remain out of
  scope; this Sprint supplies source-level protocol compatibility evidence.

## S127 Addendum: OpenAI capacity failures retry five times

### Goal

- Raise only the exact OpenAI `Selected model is at capacity` same-account
  retry limit from three to five attempts before existing account failover.

### Scope Boundary

- Carry an explicit retry limit from the capacity classifier to each OpenAI
  handler failover loop.
- Retain the generic handler fallback and per-account pool-mode retry settings
  for every non-capacity error.
- Do not add configuration, modify model substitution, alter account selection,
  or change persistence, frontend, deployment, or container behavior.

### Acceptance Boundary

- Focused normal, passthrough, and pre-output stream regressions show the
  explicit capacity limit is five and other error types leave it unset.
- Failover-loop coverage proves five retries precede the existing switch path;
  service and handler checks plus compilation and static path gates pass.

## Branch Consolidation Addendum: S110-S126 local integration

### Goal

- Consolidate the already reviewed local and upstream feature chains into one
  non-rewritten `main` history, while preserving each Sprint's contract and QA
  evidence.

### Scope Boundary

- Integrate `origin/main`, S110 group-buy lifecycle/refund hardening, and S120
  leaderboard account-age gating with normal merge commits.
- Keep the existing local S124-S126 and tutorial commits as independent,
  reviewable commits. Preserve the S114, S115, S125, and S126 workflow
  artifacts after proving their code patches are already equivalent in `main`.
- Do not push, deploy, update containers, remove remote branches, apply stashes,
  or force-remove the unresolved historical S121 worktree.

### Acceptance Boundary

- Every merge leaves no unmerged index entries, passes `git diff --check`, and
  receives its corresponding focused Go/Vitest/typecheck verification.
- The final integration passes Go module verification, repository compilation,
  selected S121/S123/S110/S120 regressions, frontend typecheck, and a local
  Redis PING/write/read/delete smoke.

## S126 Addendum: strip local `group_id` from OpenAI upstream payloads

### Goal

- Prevent strict OpenAI-compatible upstreams from rejecting standard API
  requests with `HTTP 400 Unknown parameter: 'group_id'`.

### Scope Boundary

- Treat only the exact top-level JSON field or multipart form field `group_id`
  as sub2api-local metadata and remove it at the OpenAI upstream boundary.
- Cover Responses, Chat Completions, and Images JSON/multipart forwarding while
  preserving API-key-bound group selection and nested application data.
- Do not add request-body group routing, generic unknown-field filtering,
  schema changes, frontend changes, deployment, or container work.

### Acceptance Boundary

- Focused upstream-body regressions cover native Responses, raw Chat
  Completions, Images JSON, and Images multipart. Backend compilation,
  formatting, diff, and allowed-path gates pass.

## S121 Addendum: upstream v0.1.165 administrator operation audit logs

### Goal

- Restore the upstream administrator operation-audit capability missing from
  local `main`, including the redacted audit trail, administrator console,
  trusted client-IP handling, session IP/User-Agent binding, and TOTP step-up
  safeguards.

### Scope Boundary

- Apply the final `v0.1.165` behavior rather than only its initial feature
  commit: no raw bearer/API-key/password/TOTP/refresh-token/cookie/session
  values may enter audit persistence; session data must remain unavailable to
  the audit trail.
- Record only management-plane operations and designated sensitive reads;
  failure to persist an audit record must not alter the original operation.
- Add the audit table at local migration sequence `198`; local migration `180`
  is already occupied by invoice-download tracking and must not be reused.
- Keep the primary dirty worktree untouched until the isolated implementation
  has passed review and all source-level acceptance gates.

### Acceptance Boundary

- Focused middleware, service, and administrator-handler tests cover redaction,
  session binding, step-up verification, API-key restrictions, query/detail and
  clear behavior. Go compilation, frontend typecheck/build, migration ordering,
  formatting, diff and path audits pass.
- PostgreSQL migration execution, authenticated browser interaction, real TOTP,
  deployment, and container update remain explicitly out of scope.

### S121 Implementation Result

- The isolated implementation now includes the final redaction/session fixes,
  fail-closed security dependency handling, a writer flush barrier, and an
  atomic repository transaction for clear-plus-trace. Source-level acceptance
  passed; no primary-worktree integration, push, deployment, or container
  update was performed.

## S117 Addendum: preserve omitted admin settings fields

### Goal

- Adapt upstream `0b5903d45` so a partial `PUT /api/v1/admin/settings`
  payload does not overwrite unrelated value-typed settings with Go zero values.

### Scope Boundary

- Capture the incoming top-level JSON field names before binding the existing
  request DTO. Preserve omitted value-typed setting keys, including the
  `smtp_from_email` JSON alias, while retaining explicit empty/false/zero
  updates and the existing pointer-field merge semantics.
- Refresh in-process setting caches from persisted settings after a partial
  write. Preserve all existing validation and auth-source default behavior.
- Do not modify schema, migrations, frontend, routes, deployment, containers,
  billing, account routing, or unrelated dirty S114-S116 work.

### Acceptance Boundary

- Handler/service regressions prove partial writes preserve unrelated settings,
  explicit zero values clear sent fields, JSON aliases remain writable, and
  full requests retain existing semantics. Focused Go checks, formatting, and
  static diff/path gates pass.

## S118 Addendum: Gemini pool-mode retry eligibility

### Goal

- Adapt upstream `fd7e2039d` so existing Gemini API-key pool failover paths
  preserve configured same-account retry eligibility when the error policy
  otherwise skips handling an upstream error.

### Scope Boundary

- Introduce one shared helper for Gemini pool-mode skipped-policy errors and
  call it from messages HTTP, native messages, and chat-completions forwarding.
- Keep retry eligibility gated by `account.IsPoolModeRetryableStatus`; all
  non-pool accounts, non-failover statuses, error-policy matches, retry counts,
  cooldowns, and account selection remain unchanged.
- Do not modify database, migrations, routes, frontend, scheduler, billing,
  deployment, containers, or S114-S117 work.

### Acceptance Boundary

- Focused tests prove pool 429 and configured 500 paths enter failover with
  the expected same-account flag, while pool unconfigured 500, non-pool, and
  400 cases retain current behavior. Go compile, formatting, and diff/path
  gates pass.

## S119 Addendum: Gemini client-side web-search function preservation

### Goal

- Preserve ordinary Chat Completions function tools named `web_search` when
  forwarding to Gemini, so Hermes-style client-side tools remain function
  declarations instead of becoming Gemini's built-in Google Search tool.

### Scope Boundary

- Classify Gemini built-in search only by explicit server-side tool type:
  `web_search*` or `google_search`.
- Keep every normal `type: function` tool, regardless of its nested function
  name, as a function declaration. Do not change request routing, tool-call
  response handling, persistence, account selection, frontend, deployment,
  or containers.

### Acceptance Boundary

- A forwarded Chat Completions request containing normal functions named
  `web_search` and `read_file` keeps both function declarations and emits no
  Google Search tool. Existing explicit search tool types retain their current
  built-in conversion. Focused Go tests, repository compile, formatting, and
  diff/path gates pass.

## S113 Addendum: user proxy smart input

### Goal

- 在用户“我的代理”和管理员批量导入入口统一支持
  `scheme://host:port:username:password` 冒号分隔格式，同时保留标准
  `scheme://username:password@host:port` URL。

### Scope Boundary

- 只增加前端代理文本解析、用户单条/多行表单创建、管理员批量导入复用和对应测试/提示文案。
- 复用现有 `http`、`https`、`socks5`、`socks5h` 结构化 API；不修改后端、数据库、部署、容器或运行时代理实现。

### Acceptance Boundary

- 解析器覆盖正常、认证、IPv6、额外冒号和无效输入；用户侧多行输入逐条校验并批量创建；前端定向测试、typecheck、生产构建和 diff 门禁通过。

# Workflow Spec

## S108 Addendum: user usage column menu layer

### Goal

- Keep the user usage column settings menu above the sticky table header and
  record rows while it is open.

### Scope Boundary

- Dynamically elevate only the user usage filter card while the existing
  column menu is open, following the established admin usage solution.
- Do not change table-wide z-index values, menu behavior, filtering, export,
  persistence, backend APIs, deployment, or containers.

### Acceptance Boundary

- Focused view regression, frontend typecheck/build, diff/path gates, and a
  desktop browser smoke must confirm that the menu is not obscured.

## S107 Addendum: x/text security dependency update

### Goal

- Adapt upstream `c5971a6fc` so `golang.org/x/text` reaches `v0.39.0` and
  GO-2026-5970 is removed from the backend module graph.

### Scope Boundary

- Upgrade only the eight `golang.org/x/*` modules selected by the upstream
  security commit, preserving the local direct/indirect dependency topology.
- Do not change Go source, generated code, schema, frontend, deployment,
  containers, VERSION, or unrelated dependencies.

### Acceptance Boundary

- Exact module-version and checksum review, `go mod verify`, backend build,
  focused/broad Go regression attempts, vulnerability scan, diff, conflict,
  and unmerged-index gates must pass or record unrelated baselines explicitly.

## S106 Addendum: selective upstream small fixes

### Goal

- Port five isolated upstream behavior fixes for scheduler quota metadata,
  monitor decrypt-failure scheduling, subscription validity-unit display,
  usage multiplier precision, and promo expiry local-time prefill.

### Scope Boundary

- Preserve all local scheduler quota dimensions, including local-only monthly
  quota fields, while continuing to filter unrelated account metadata.
- Treat only API-key decryption failure as terminal for channel-monitor
  scheduling; ordinary failures keep retrying.
- Keep validity labels aligned with backend day/week/month billing semantics,
  retain meaningful multiplier decimals, and use the existing local-time
  formatter for promo edit values.
- Do not change persistence, schema, billing, payment execution, dependencies,
  deployment, containers, VERSION, or unrelated upstream behavior.

### Acceptance Boundary

- Focused and broader unit-tag Go tests, focused Vitest regressions, frontend
  typecheck/build, formatting, exact allowlist, conflict-marker, unmerged-index,
  and `git diff --check` gates must pass.

## S105 Addendum: filter admin accounts by OpenAI plan type

### Goal

- Distinguish OpenAI `Plus`, `Pro`, `K12`, `Team`, `Free`, `Other`, and
  `Unrecognized` accounts in account management and filter them consistently.

### Scope Boundary

- Filter the persisted `credentials.plan_type` in the repository before count
  and pagination; a selected plan category implicitly limits the query to
  OpenAI accounts.
- Propagate the same `plan_type` through list, owner/share list, filtered bulk
  edit, filtered share-status changes, and filtered export.
- Keep `share_display_tier` display-only. Do not rewrite credentials, add
  manual plan editing, change import/OAuth enrichment, add schema/migrations,
  or touch scheduler, gateway, billing, deployment, and containers.

### Acceptance Boundary

- Repository integration tests cover known aliases, K12, other, unrecognized,
  OpenAI scoping, total, and pagination behavior.
- Service/handler tests prove list, filtered bulk, and export propagation;
  frontend tests cover filter snapshots and normalized badge labels.
- Focused Go/Vitest, typecheck, production build, formatting, exact allowlist,
  conflict-marker, unmerged-index, and `git diff --check` gates must pass.

## S104 Addendum: preserve OpenAI plan type across inactive workspaces

### Goal

- Preserve a token-derived OpenAI/K12 plan type when ChatGPT
  `accounts/check` also returns inactive workspace billing metadata.

### Scope Boundary

- Skip deactivated, disabled, deleted, inactive, suspended, and expired
  workspace candidates when selecting fallback account information.
- Apply the `accounts/check` plan type only when the token did not provide one;
  keep subscription-expiry, email, privacy, timeout, and logging behavior.
- Do not change Codex session identity matching, PAT/Agent Identity support,
  persistence, scheduler routing, gateway behavior, frontend, migrations,
  billing, deployment, or containers.

### Acceptance Boundary

- Focused service tests cover expired/deactivated candidate rejection and
  token-plan preservation while retaining empty-plan fallback behavior.
- Formatting, exact allowlist, conflict-marker, unmerged-index, and
  `git diff --check` gates must pass.

## S82 Addendum: clarify OpenAI WS account-mode prerequisite

### 一句话目标

- 明确账号级 OpenAI Responses WS mode 只有在全局 `gateway.openai_ws.mode_router_v2_enabled=true` 时生效，避免管理员配置后仍走 legacy 路由却缺少提示。

### 边界与验收

- 只修改 README、示例配置注释、中英文账号帮助文案与定向 locale 测试；不修改任何运行时配置值或路由代码。
- 本地账号级模式继续限定为 `off / ctx_pool / passthrough`；上游新增的账号级 `http_bridge` 不属于本地能力，不能写入帮助文案。
- 本地 `http_bridge_enabled` 仍只是大首包 HTTP fallback 开关，与账号级模式保持区分。
- 定向 Vitest、typecheck、production build、精确十一项路径审计和 protected-hash gate 必须通过。

## S83 Addendum: minute-level subscription expiry display

### 一句话目标

- 在管理员和用户订阅页面将有效期时间显示到分钟，保留现有日期 locale、失效状态和剩余天数逻辑。

### 边界与验收

- 复用 `formatDate` 新增分钟级 helper，只替换管理员订阅视图和本地 `UserSubscriptionsPanel` 的有效期展示；不修改 API、后端、计费、时区或 UsageView。
- 无效日期继续显示空字符串，状态文案和剩余天数计算保持不变。
- 定向 Vitest、typecheck、production build、十项路径审计和 protected-hash gate 必须通过。

## S84 Addendum: buffered Anthropic JSON content type

### 一句话目标

- 修复 OpenAI-compatible Anthropic buffered 响应被上游 SSE header 污染的问题，确保非流式 JSON 响应声明 `application/json`。

### 边界与验收

- 只在 buffered 转换路径 `c.JSON` 前覆盖 Content-Type；流式 SSE、响应 body、usage、计费和 failover 不变。
- 按本地函数签名重写回归测试，不直接复制上游测试参数。
- 定向 Go 测试、gofmt、八项路径审计、冲突/diff 和 protected-hash gate 必须通过。

## S85 Addendum: avoid cache billing on same-account retry

### 一句话目标

- 同账号重试期间不因 sticky/bound session 单独强制缓存计费；真正切换账号或上游显式要求时仍保持缓存计费。

### 边界与验收

- 只修改 failover 状态中的 `ForceCacheBilling` 判定和对应 handler 测试；不修改重试次数/延时、账号切换、临时封禁、错误分类或计费计算。
- 定向及 broader handler 测试、gofmt、八项路径审计、冲突/diff 和 protected-hash gate 必须通过。

## S81 Addendum: renew expired admin subscription assignments

### 一句话目标

- 管理员重新分配已过期的同分组订阅时，复用原记录并从当前时间开启新周期，而不是返回成功但继续保持过期。

### 边界与验收

- 仅修改 admin assignment 的复用分支；单个/批量分配都复用同一逻辑。
- 新周期重置 starts/expires、active 状态、日/周/月窗口与用量，并保留原 ID 和分配来源。
- suspended 无论是否已过期都不得自动恢复；有效 active 订阅的幂等复用/冲突语义保持。
- 管理员相同备注不重复追加，不同备注追加一次；购买/兑换使用的 `AssignOrExtendSubscription` 仍记录每次事件，即相同备注继续追加。
- 不修改 repository、schema、migration、handler、frontend、billing、payment、redeem 或部署配置。

## S80 Addendum: Redis Compose command continuation

### 一句话目标

- 把上游 `be74deae7` 的 Redis 启动参数续行修复移植到本地三个内置 Redis 的 Compose，确保 RDB、AOF、fsync 与可选密码参数真正传给同一次 `redis-server` 调用。

### 边界与验收

- 只修改 `docker-compose.yml`、`docker-compose.local.yml`、`docker-compose.dev.yml` 的 Redis command；本地/开发文件是与上游主文件同构的行为补齐。
- 保持 Redis image、容器名、healthcheck、`REDISCLI_AUTH`、volume、network、端口和其他服务完全不变。
- `docker-compose.standalone.yml` 使用外部 Redis，不修改。
- 验收只做 Docker Compose 静态渲染、空/非空受控密码命令检查和路径审计；不启动、更新、重启或删除任何容器/volume。
- 任意特殊字符密码的 shell 安全化、真实 Redis `CONFIG GET`、磁盘增长评估和部署均属于后续独立任务。

## S79 Addendum: upstream v0.1.161 low-risk compatibility

### 一句话目标

- 在不整体合并 `v0.1.161` 的前提下，把四个互不扩张的低风险行为移植到本地：Antigravity 付费 tier 保留、Anthropic 监控文本块提取、Claude Code `[1m]` 后缀归一化、套餐有效期动态单位文案。

### 边界与验收

- Antigravity 的异常状态与原因继续保留，但已识别的 `Pro/Ultra` 不再被 `IneligibleTiers` 覆盖为 `Abnormal`。
- Anthropic monitor 只拼接 `content[]` 中的 `type=text` 块；thinking/tool 块不参与 challenge 判断。
- `[1m]` 只在 Anthropic 请求模型末尾按大小写不敏感方式剥离，可处理重复后缀；归一化必须进入实际 handler 转发 body，其他协议和中间后缀保持不变。
- 现有 payment locale key 不改名，只去掉“天/days”的写死文案，继续由 `validity_unit` 表达单位。
- 不修改 deploy/Compose、Responses SSE、Grok media、routing、subscription assignment、migration、billing、security、VERSION、lockfile 或 `knowledge/**`。

## S75 Addendum: admin usage column-menu stacking

### 一句话目标

- 修复管理端使用记录中“列设置”及筛选下拉被固定表头/记录遮挡的问题，让筛选卡片的浮层稳定位于表格之上。

### 边界与验收

- `UsageFilters` 当前为 `z-30`，而 `DataTable` 固定表头最高为 `z-index: 220`；仅在 `showColumnDropdown` 为真时，从 `UsageView` 向筛选组件传入 `z-[221]`。
- 不修改 `DataTable`、`UsageTable`、筛选状态、菜单交互或请求逻辑，也不引入 Teleport。
- 视图测试锁定菜单打开/关闭时筛选卡片的动态层级；跑定向 Vitest、typecheck、production build 与 `git diff --check`。

## S74 Addendum: support ticket user context

### 一句话目标

- 让管理员在工单管理中直接看到用户的用户名和注册邮箱，并可从工单详情以只读方式打开现有用户信息弹窗，查看最近使用、订阅订单和充值/余额流水。

### 边界与验收

- 管理端工单列表和详情只增加 `{ id, username, email }` 用户摘要；用户侧工单接口不能新增该字段，完整 `AdminUser` 只能在管理员主动点击后通过既有 `/admin/users/:id` 接口获取。
- 资料摘要必须由工单查询一次性关联用户表获得，不能在前端按工单逐条请求用户，避免 N+1。
- 管理端必须使用独立的 ticket DTO mapper 填充摘要；原有用户侧 mapper 和序列化结果必须继续不含 `user` 摘要，并由定向测试锁定。
- `UserBalanceHistoryModal` 在工单页以 `hideActions` 只读模式复用；最近使用请求最近 30 条，并以固定高度的纵向滚动区域展示，不应撑高整个弹窗。
- 不修改数据库 schema、支付/余额/订阅业务、鉴权路由或用户端工单界面。

## S45 Addendum: affiliate risk scoring and alert scanner

### 一句话目标

- 做一个最小可用的邀请返佣风控扫描器：扫描间隔默认 `20m` 且可在后台设置调整，扫描最近 `12h`，按风险评分写入 `ops_alert_events`，并在高风险时冻结邀请奖励兑现；第一版不自动封号、不自动禁用 API key、不回滚历史奖励。

### 当前结论

- 这不是单条规则封禁，而是“风险评分 + 告警 + 奖励兑现冻结”。
- 默认扫描窗口从原先讨论的 `24h` 收敛为 `12h`；扫描周期默认 `20m`，但必须加入后台设置，方便运营自行调整。
- 扫描间隔设置建议限制在 `5-1440` 分钟，非法值回退 `20m`，避免过频扫描压数据库。
- 高风险处理只冻结返佣变现路径：
  - 阻止被邀请人首次 API 调用奖励 claim。
  - 阻止邀请返佣 quota 转余额。
  - 不扣回已发 ledger、不移除邀请关系、不封用户、不禁用 API key。

### 主要触达面

- 数据源：`users.register_ip`、`users.last_login_ip`、`usage_logs.ip_address`、`user_affiliates.inviter_id`、`user_affiliate_ledger.action = 'api_call_reward'`。
- 后端服务：新增 affiliate risk scanner、IPv6 `/64` 归一化、风险评分和去重。
- 后台设置：新增扫描间隔设置项，不需要新增独立风控页面。
- 性能索引：补齐 `users(created_at)`、`user_affiliate_ledger(action, created_at)`、`usage_logs(ip_address, created_at) WHERE ip_address <> ''`。
- 持久化：新增风险冻结记录或等价状态，用于拦截 claim/transfer。
- 运维告警：复用 `ops_alert_events`，并尽量复用现有 ops email 通知路径。
- 启动调度：按现有后台服务模式启动，使用 Redis leader lock / heartbeat 风格。

### 评分和分级

- 同一邀请人 `12h` 内邀请 `>=3` 个账号：`+25`。
- 邀请人和被邀请人登录 IP 相同：`+40`。
- IPv6 同 `/64`：`+35`。
- 注册 IP 分散但登录 IP 或 `/64` 聚合：`+25`。
- 注册后 `30m` 内触发 API 奖励：`+20`。
- 多个被邀请账号邮箱像批量生成：`+10`。
- 被邀请关系已撤销/禁用但存在 `api_call_reward`：`+30`。
- `>=50` 为 `P3` 告警，`>=70` 为 `P2` 告警并冻结兑现，`>=90` 为 `P1` 高风险冻结兑现。

### 当前阻塞与风险

- 当前主工作树仍有 payment、welfare voucher、settings、billing、gateway、frontend payment/i18n、knowledge 脏改。
- S45 会触达 affiliate、ops、wire、migration、后台任务启动链路和最小 settings 表单；必须在干净 worktree 或收口脏树后实现。
- `OpsService.CreateAlertEvent` 只负责写事件；现有 email 发送在 alert evaluator 内部，S45 实现时必须避免复制大段邮件逻辑。
- migration 编号不能直接假定；实现前要重新检查 tracked/untracked migrations。
- 当前已有 `usage_logs(user_id, created_at)`、`usage_logs(created_at)`、`user_affiliates(inviter_id)` 等基础索引，但缺少上述风控扫描专用索引；S45 实现必须补窄范围 migration，避免按 IP/时间或 ledger action/time 扫全表。

### 推荐执行计划

1. 评审并批准 `docs/workflow/tasks/affiliate-risk-alerts-s45.md`。
2. 在干净 worktree 开发，先实现 IPv6 `/64` 归一化和评分纯函数测试。
3. 增加 risk repository 查询最近 `12h` 邀请/登录/API 奖励/usage IP 聚合数据。
4. 增加 scan-specific indexes migration：`users(created_at)`、`user_affiliate_ledger(action, created_at)`、`usage_logs(ip_address, created_at) WHERE ip_address <> ''`。
5. 增加 scanner 服务：读取后台配置的扫描间隔，默认 `20m`，Redis leader lock、ops heartbeat、去重写 `ops_alert_events`。
6. 增加风险冻结持久化，并在 `ClaimInviteeAPICallReward` / `TransferAffiliateQuota` 前拦截。
7. 增加后台设置项，允许调整扫描间隔。
8. 跑定向 Go 测试、frontend typecheck、migration 编号检查、`git diff --check` 和 denied-path audit。

### 明确不在 S45 范围内

- 不做前端新页面；只允许在后台设置中增加扫描间隔控制，运维中心先复用现有告警列表。
- 不自动封号。
- 不自动禁用 API key。
- 不扣回历史奖励。
- 不删除或撤销邀请关系。
- 不混入支付、福利券、Studio Bridge、OpenAI image/video 或前端 payment 脏改。

## S43 Addendum: upstream v0.1.143 group peak-rate synthesis plan

### 一句话目标

- 把上游 `v0.1.143` 的订阅分组高峰时段倍率能力拆成一个独立产品级合成批次，先完成边界和验收计划，再决定是否进入 schema/migration + billing/gateway + frontend 实现。

### 当前结论

- 本地尚未合入订阅分组高峰时段倍率能力。
- 本地现有能力包括普通 `rate_multiplier`、`image_rate_multiplier` 和用户专属分组倍率/RPM；未发现 `peak_rate_enabled`、`peak_start`、`peak_end`、`peak_rate_multiplier` 这套字段和链路。
- 上游该能力不是小补丁，不能混入 S38a-S42 这类安全兼容提交。

### 上游来源

- `915c60b15 feat(group): 订阅分组新增可选的高峰时段倍率，以支持智谱等coding plan的高峰时段`
- `1034f576d fix: 高峰倍率全链路透传、计费术语修正与边界处理`
- `11a3da65c fix(group): harden peak-rate config handling and label peak windows with server timezone`

### 主要触达面

- 数据层：`backend/ent/schema/group.go`、Ent 生成代码、group migration。
- 后端接口：admin group create/update/list DTO、available channels、public settings server timezone。
- 服务层：group validation/normalization、API key auth cache/group hydration、gateway/openai gateway usage recording。
- 计费层：token 计费倍率叠加高峰因子，图片按次计费不受高峰因子影响。
- 前端：admin GroupsView、GroupBadge、GroupOptionItem、AvailableChannelsTable、Payment/Subscriptions/Keys 页面和 i18n。

### 当前阻塞与风险

- 当前主工作树仍有 payment、welfare voucher、settings、billing、gateway、frontend payment/i18n、knowledge 脏改。
- 上游高峰倍率触达路径与当前脏文件重叠：`backend/internal/handler/dto/settings.go`、`backend/internal/handler/payment_handler.go`、`backend/internal/service/billing_service.go`、`backend/internal/service/gateway_service.go`、`backend/internal/service/openai_gateway_service.go`、`backend/internal/service/setting_service.go`、`frontend/src/types/index.ts`、`frontend/src/types/payment.ts`。
- 上游 migration 是 `158_add_group_peak_rate_multiplier.sql`，本地 migration 已推进到更高编号且存在未提交 migration 工作；实现时必须使用本地下一安全编号。
- 上游允许 `peak_rate_multiplier=0` 表示高峰免费策略；本地实现前需要确认是否接受该产品语义。

### 推荐执行计划

1. 先收口或隔离当前 payment/welfare/settings/gateway/frontend dirty tree。
2. 在干净 branch/worktree 上开启实现 Sprint，不直接 cherry-pick 三个上游提交。
3. 先做 schema、migration、Ent 生成和 DTO/mappers。
4. 再做 group service 的校验、归一化、server timezone 计算和 API key cache 透传。
5. 然后接入 gateway/openai gateway 计费快照，锁定“高峰只影响 token 倍率，不影响图片按次计费”。
6. 最后接前端 admin/user display 和 i18n，跑后端定向测试、frontend typecheck/Vitest、`git diff --check` 与 staged denied-path audit。

### 明确不在 S43 范围内

- 不修改业务代码。
- 不新增 migration 或 Ent 生成代码。
- 不触碰当前 dirty payment/welfare/settings/gateway/frontend 文件。
- 不把 post-`v0.1.143` 的 `a5638a4e5`、ops realtime stats、redeem invitation fix 混入高峰倍率计划。

## S35 Addendum: upstream v0.1.142 merge plan

### 一句话目标

- 在不整体 merge `upstream/main` / `v0.1.142` 的前提下，把上游 `v0.1.142` 与本地长期 fork 的差异拆成可评审、可验证、可暂停恢复的分批合并计划。

### 当前背景

- GitHub latest release 已确认为 `v0.1.142`，tag 提交为 `60da9ba17`，发布时间为 2026-07-01。
- 本地 `main` 已经长期分叉，直接 `git merge --no-commit v0.1.142` 在临时 worktree 中触发大量冲突，集中在 Ent 生成代码、account/proxy schema、gateway、payment、usage 和前端视图。
- 本地主工作树当前仍有未提交的福利券、设置、用户代理、知识库等改动，且 `main` 相对 `origin/main` ahead；合入上游补丁前必须先收口或隔离这些本地变更。
- 上轮只读筛选已经证明，多个 `v0.1.138..v0.1.142` 小补丁可以在干净本地 `HEAD` 上通过 `git apply --check --3way`，但 Grok、Spark shadow、Codex detect、Anthropic dateline / Sonnet5 属于大功能链路，需要另开 Sprint。

### 合并策略

- 禁止整段 merge/rebase `v0.1.142`。
- 继续采用“小批次 port + contract + 定向测试 + denied-path audit”的方式推进。
- 每一批只允许包含同一业务域的上游补丁；支付、OpenAI/Codex 网关、usage billing、订阅、前端 API base 不能混成一个提交。
- 当前 dirty tree 未收口前，不启动代码迁移；若必须先 port，应使用干净 worktree，并且 contract 必须声明不会触碰主工作树脏改。

### 推荐批次

1. `S36 payment-refund-safe-bundle`
   - 候选提交：`c6f375d3a`、`b1403e8b2`、`55242ffac`、`65ad7df4f`、`7316d8302`、`93a3bf307`、`930326116`。
   - 目标：订阅金额、汇率换算、退款 pending、支付卡片和币种显示的安全修复。
   - 风险：触碰支付前后端与退款流程，必须定向跑 payment service / handler / frontend vitest。

2. `S37 openai-codex-gateway-safe-bundle`
   - 候选提交：`9491de0a3`、`ae5e980dd`、`65fa72892`、`0a97a5f46`、`2b49d662c`、`011278204`、`e5f7836bf`、`73de2ea7f`、`b28a22333`、`82553c4dc`、`7a38c6621`。
   - 目标：OpenAI/Codex transport failover、tool args 去重、Codex image bridge、GPT-5.5 Pro Codex 名称保留、count_tokens bridge 等网关兼容修复。
   - 风险：部分行为与本地既有 S23-S30 迁移重叠，必须逐项判定 `ported / equivalent / skipped`。

3. `S38 billing-subscription-safe-bundle`
   - 候选提交：`9f5b57fc9`、`03727ac36`、`fd004bdd8`。
   - 目标：余额计费防透支、订阅撤销软删除、account query `Count` 污染修复。
   - 风险：`9f5b57fc9` 会触碰 `usage_billing_repo.go`、`billing_cache_service.go`、`gateway_service.go`，与当前福利券/usage billing 脏改重叠；必须等福利券工作树收口后再做，或独立干净 worktree 迁移。

4. `S39 frontend-small-fixes`
   - 候选提交：`2a58a57a7`、`8c2d9b9a1`。
   - 目标：前端 direct requests 使用配置的 API base；是否移除 `gpt-5.3-codex` 默认模型由本地产品策略决定。
   - 风险：触碰前端 i18n、Settings、KeyUsage 等文件；需要前端 typecheck / targeted vitest。

5. 独立大功能 Epic，不混入小补丁批次：
   - Grok subscription / media / OAuth / quota 链路。
   - OpenAI Spark shadow account。
   - `codex_cli_only` engine fingerprint 加固与 app-server 配置。
   - Anthropic OAuth dateline 指纹抹除与 Sonnet5 适配。

### 明确不在范围内

- 不整体 merge `v0.1.142`。
- 不在 S35 修改任何业务代码。
- 不触碰 Ent、migrations、wire、生产配置、Docker/deploy、README/assets，除非后续独立 Sprint contract 明确批准。
- 不把当前未提交的福利券、发票、用户代理、知识库改动纳入上游合并批次。

### 验收标准

- S35 只交付合并计划和下一步 contract 草案，不做代码迁移。
- `docs/workflow/status.md` 进入 `contract-draft`，下一合法动作是评审 S35 plan contract。
- `docs/workflow/tasks/upstream-main-v0142-merge-plan-s35.md` 明确成功标准、允许路径、禁止路径、候选批次、验收命令和 stop rules。
- `git diff --check` 覆盖本轮 workflow 文档。

## S18 Addendum: APIMart task webhook

### 一句话目标

- 在不改变普通图片同步代理兼容行为的前提下，为 APIMart 视频/长任务接入任务完成 webhook，让 Sub2API 能在任务终态时主动完成本地状态落库、结算和失败退款。

### 当前背景

- Sub2API 当前仍是 Studio Bridge / 落叶AI的账号、余额、分组和扣费真源。
- `chatgpt2api` / 落叶创作台负责任务体验，但任务成功扣费、失败退款和使用记录最终应回到 Sub2API 侧闭环。
- APIMart task webhook 只在任务 `completed` / `failed` 等终态后回调；因此它适合补强长任务可靠结算，不适合作为普通同步 OpenAI 图片接口的直接替代。
- 本地视频任务已经有 `openai_video_tasks`、预扣、`/v1/tasks/:task_id` 查询结算和失败退款逻辑，S18 应复用这条链路，而不是新增并行账本。

### 明确不在范围内

- 不把普通 `/v1/images/generations` 同步响应改为异步 `task_id` 返回。
- 不覆盖客户请求里已经带的 `webhook` 字段；客户 webhook fan-out/relay 另开任务。
- 不新增数据库迁移，不写真实公网域名或 secret，不改 Studio Bridge / chatgpt2api 协议。
- 不整体重构 Image Creator 或 APIMart 图片轮询逻辑。

### 验收标准

- webhook 接收端有 secret 校验、body 大小限制、脱敏日志和幂等处理。
- APIMart 视频任务只在配置完整且请求未带客户 webhook 时注入 Sub2API callback URL。
- 成功终态只结算一次；失败终态只退款一次；重复回调不重复扣退。
- 现有 `/v1/tasks/:task_id` 查询结算仍作为兜底可用。
- 定向后端测试和 `git diff --check` 通过，denied-path audit 不触碰图片同步代理、迁移、Ent、公共页、支付页、Canvas、Studio 前端等禁止范围。

## S19 Draft Addendum: upstream v0.1.137 postfixes

### 一句话目标

- 在 S15-S17 已完成的基础上，继续筛出 `v0.1.137` 中不碰迁移、不碰前端、不覆盖本地产品定制的后端小修，作为 S18 之后的候选小步迁移。

### 当前背景

- 上游 `v0.1.137` 的安全/兼容主干已经通过 S15/S16/S17 小步迁入。
- 仍有少量后端补丁有合并价值，但不应和当前 S18 APIMart webhook 实现混在一起。
- 当前 S19 只是 follow-up contract 草案；`docs/workflow/status.md` 仍以 S18 `contract-draft` 为当前合法动作。

### 候选范围

- OpenAI failover 复用原始错误体。
- Anthropic window cooldown 保留。
- Account repository 列表参数限制与 refresh candidates SQL 修复。

### 明确不在范围内

- 不整体 merge/rebase `upstream/main`。
- 不碰 Ent、migrations、VERSION、wire 生成物或前端。
- 不把 OpenAI image failover、token refresh retry amplification、OAuth promo signup、scheduler outbox dedup/cleanup、cyber policy、channel monitor jitter、Claude OAuth system prompt blocks 混入本轮。

### 验收标准

- 定向后端 service / repository / server contract 测试通过。
- `git diff --check` 通过。
- denied-path audit 返回 `NO_DENIED_PATHS`。
- worker/result 和 QA report 说明每个上游候选是 `ported`、`equivalent` 还是 `skipped`。

## 一句话目标

- 在不覆盖本地 Studio Bridge / 支付治理 / Canvas / 公共页等产品定制的前提下，把上游 `v0.1.137` 的低风险安全、兼容、计费兜底和管理员运维能力按独立 Sprint 小步迁入，并为后续继续评估候选 patch 保留清晰边界。

## 当前背景

- 本地 `sub2api` 已不再处于“直接跟上游 merge”阶段；`main` 与 `upstream/main` 分叉较大。
- 仓库当前产品主线仍是 Studio Bridge / 落叶AI生产联调、支付套餐与用户治理。
- 同期存在一条独立工程主线：从上游按 Sprint 迁入低风险 patch，但显式保护本地定制，不碰 Ent/migrations/VERSION，不整体 merge `upstream/main`。
- 2026-06-17 已完成三个连续 Sprint：
  - `upstream-main-v0137-safe-patches-s15`
  - `upstream-main-v0137-small-compat-s16`
  - `upstream-main-openai-quota-reset-s17`

## 当前范围

- 当前 workflow spec 只描述 2026-06-17 这条上游小步合成链路。
- 不覆盖 Studio Bridge / 支付套餐 / 用户 IP / 首充福利这类产品主线的完整需求文档。
- 不把“后续也许会迁”的候选 patch 说成已批准范围。

## 已完成 Sprint 范围

### S15: 安全 / 兼容 / 计费兜底

- 锁定前端 `form-data` 到 `4.0.6`。
- token refresh 增加不可重试错误分类。
- 上游响应支持 zstd。
- 非流式 2xx 非 JSON 与 SSE `event:error` 进入 failover，并保留原始错误体。
- `tool strict` 缺省补 `false`。
- 国产模型 fallback pricing 和图像输入 token 计费补齐。
- DeepSeek `reasoning_effort=max` 归一到 `xhigh`。
- Anthropic thinking block 过滤改为按 mapped upstream model 分流。

### S16: 小兼容补丁

- Responses API sticky hash 在缺少旧字段时以 `input` 兜底。
- Claude Code `max_tokens=1` 的 Haiku 流式探测拦截补齐。
- OpenAI APIKey `/responses` probe 增加工具能力校验。
- API Key ACL 拒绝信息补充实际 client IP。

### S17: OpenAI OAuth 上游 quota/reset

- 新增管理员 OpenAI OAuth 账号上游 WHAM quota 查询入口。
- 新增 rate-limit reset credit 消费入口。
- 后端复用 token provider、privacy client factory 和账号代理解析。
- 前端仅在 OpenAI OAuth usage cell 展示上游 credits 查询/重置控件。

## 明确不在范围内

- 不整体 merge 或 rebase `upstream/main`。
- 不触碰 Ent schema、migrations、VERSION、wire 大链路生成物。
- 不覆盖本地 Studio Bridge、Canvas、支付页、公共页、模型市场、工单或 Chat/Image Studio 定制。
- 以上边界仅适用于 S15-S17 这条上游小步迁移链路；后续统一 API Key / APIMart 图片网关 / 前端导航合并已经触达 `wire_gen.go`、Studio Bridge repo、公共页、模型市场、Keys 和 Settings 等路径，不能再用本节作为当前 `origin/main..HEAD` 的 denied-path 证明。
- 不引入 migration-heavy、compliance gate、cyber policy、渠道监控 jitter、Claude OAuth system prompt blocks 等高风险链路。
- 不把前端全量 Vitest 失败修复混进本轮 Sprint；这应另开前端稳定化任务。

## 当前稳定工程边界

- 当前允许的上游迁移策略是“低风险 patch + 独立 Sprint + 定向验证 + denied-path audit”。
- 每轮 Sprint 都必须显式说明：
  - 为什么该补丁适合独立迁移
  - 为什么不会覆盖本地产品面
  - 需要哪些定向测试证明行为稳定
  - 哪些更大链路明确跳过
- 当前稳定结论不是“本地已接近上游”，而是“本地已有一条可持续的小步迁移方法”。

## 验收标准

- patch 迁入后，定向后端测试通过。
- 涉及前端控件或管理页时，定向 Vitest 通过。
- `git diff --check` 通过。
- 上游小步 Sprint 的 denied-path audit 应返回 `NO_DENIED_PATHS`；若当前批次是产品合并或 UI/网关主线合并，则必须改为列出实际触达路径和对应验证，不能沿用旧审计结论。
- lockfile 扫描无已知需规避版本残留，例如 `form-data@4.0.5`。
- 迁移结果不触碰本轮禁止路径；若后续合并触达曾经的禁止路径，workflow/knowledge 必须明确说明这是新批次范围，而不是继续复用旧 Sprint 证据。
- workflow 文档能说明“为什么这轮可以迁、为什么其他候选仍应跳过”。

## 当前证据入口

- `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
- `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
- `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
- `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
- `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
- `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`

## 下一步候选

- 若继续做上游迁移，需另开 Sprint 单独评估候选 patch。
- 当前可评估但未批准的方向包括：
  - OpenAI image failover
  - Anthropic window cooldown
  - account list parameter batching
  - token refresh retry amplification / outbox dedup
- 若继续做前端全量测试收口，应另开“前端稳定化”任务，不与上游 patch Sprint 混合。

# S76 Addendum: upstream v0.1.152 low-risk compatibility

## Goal

Selectively port three independently testable `v0.1.152` improvements without importing the release's migration, billing, account-model, or prompt-cache chains.

## Approved Scope

- Replace Fast/Flex raw user-ID editing with email search and exact-ID preservation.
- Strip unsupported reasoning fields only for known Grok Composer aliases.
- Diagnose OpenAI-compatible no-account failures using the API key group's actual platform.

## Explicitly Deferred

- xAI API-key account support and Grok free OAuth prompt caching.
- Codex alpha/search routing, per-call billing, migration `174`, and Ent generated changes.
- Grok quota persistence/capacity UI, CLI proxy routing, deployment, VERSION, and release-wide merge.

## Acceptance Boundary

- Targeted backend service/handler tests, selector/i18n Vitest, frontend typecheck, and path/diff audit must pass.
- No existing user ID may be silently dropped or converted into a global Fast/Flex rule.
- Non-Composer Grok requests and existing account-selection/failover loops must remain unchanged; Count Tokens may align with the shared 404/503 no-account classifier.

# S82 Addendum: usage-record model reasoning effort

## Goal

Show a normalized reasoning-effort suffix immediately after the requested model
name in user and administrator usage-record tables when a meaningful effort is
recorded.

## Scope Boundary

- Reuse existing model-label sanitization and reasoning-effort formatting.
- Keep the standalone reasoning-effort column unchanged.
- For admin mapping chains, annotate only the requested model (the first step).
- Do not change backend capture, API types, exports, filters, dashboards,
  billing, persistence, deployment, or `knowledge/**`.

# S86 Addendum: Grok proxy quality target

## Goal

Add xAI/Grok reachability to the existing administrator proxy quality check.

## Scope Boundary

- Probe `GET https://api.x.ai/v1/models` through the selected proxy and treat
  HTTP 401 as reachable.
- Display the result target as `Grok` in the existing table.
- Do not change scoring, timeouts, other targets, persistence, table layout,
  deployment, containers, or `knowledge/**`.

# S87 Addendum: upstream v0.1.162 low-risk compatibility

## Goal

Selectively port three independently testable `v0.1.162` fixes without merging
upstream history or importing its security, media, identity, audit, branding,
migration, or VERSION changes.

## Scope Boundary

- Preserve API-key IP lists on omitted partial-update fields while retaining
  explicit empty-array clearing and validation.
- Return standard OpenAI `insufficient_quota` errors on the local Responses
  roots without changing existing protocol-specific error formats.
- Restore Available Channels scrolling through the existing `.table-wrapper`
  layout contract.
- Do not port `a05b87321` in isolation: its upstream currency field/schema/API
  prerequisite is absent locally and requires a separate migration-heavy plan.
- Treat existing S85 commit `24ade9b71` as a required regression gate; do not
  modify failover/billing behavior in S87.

## Explicitly Deferred

- S3 ephemeral encryption-key protection, Grok chained video proxying, Codex
  model manifests, Agent Identity team isolation, Prompt Audit fail-closed,
  batch-image i18n, dark-theme bundles, branding SVG, VERSION changes, and the
  subscription-plan currency display chain (`a05b87321` plus its prerequisite).
- Ent, migrations, billing calculation, scheduler, payment execution, account
  routing, deployment, containers, and `knowledge/**`.

# S88 Addendum: model-aware multi-group fallback isolation

## Goal

Prevent parsed model-aware requests from falling back to an incompatible API
key default group after all multi-group route candidates are rejected.

## Scope Boundary

- Preserve compatible default-group fallback and all single-group behavior.
- Reject only when the default group's routing scope/platform is incompatible,
  or its enabled explicit route rules do not match the request.
- Return a stable 403 route error before billing or account scheduling.
- Do not change persistence, routing configuration, channel restrictions,
  billing, account scheduling, frontend behavior, deployment, or containers.

# S89 Addendum: API-key editor split layout

## Goal

Make the API-key create/edit dialog wider and keep multi-group route editing in
a dedicated right column on desktop, with independent scrolling for long route
lists.

## Scope Boundary

- Use the existing extra-wide dialog width and responsive Tailwind layout.
- Keep basic, quota, and advanced settings in the left column; keep the route
  editor in the right column at desktop widths and return to one column below
  `lg`.
- Preserve all route fields, payloads, validation, and backend behavior.
- Do not migrate model matching into group administration in S89; that requires
  a separate persistence and routing contract.

# S90 Addendum: account-pool strategy feature visibility

## Goal

Hide the API-key editor's account-pool strategy control when the system has
explicitly disabled account sharing.

## Scope Boundary

- Read the existing `account_share_enabled` value from the public settings
  already loaded by `KeysView`.
- Hide the whole label/select/hint block only for explicit `false`.
- Preserve visible behavior when the setting is missing or still loading.
- Preserve form state and create/update payloads while the control is hidden.
- Do not change backend behavior, admin settings, translations, deployment, or
  containers.

# S91 Addendum: group model-match centralization

## Goal

Make model eligibility an administrator-owned group property. API-key users
select groups and tune priority/weight; they no longer author per-route model
patterns. Route selection and `/v1/models` must use the same group rule set,
while channel `restrict_models` remains the final channel-level hard limit.

## Scope Boundary

- Add independent `groups.model_match_patterns` JSONB storage and expose it
  through group service/admin DTOs and the group editor.
- Normalize rules by trimming, lower-casing, removing duplicates, and requiring
  at least one rule; `*` is the explicit match-all rule.
- Reject ordinary-user writes containing legacy route `model_patterns`; retain
  only the minimum legacy read/clear path required for the switch migration.
- Filter multi-group route candidates by group rules before priority/weight
  selection and use the same filter for `/v1/models` aggregation.
- Add a guarded migration/preflight: list unconfigured effective groups and
  legacy rules/conflicts, refuse partial switching, then transactionally clear
  legacy API-key route rules and invalidate auth-cache snapshots.
- Do not reuse `model_routing` or `models_list_config`; do not change channel
  restriction semantics, billing, deployment, containers, or unrelated pages.

## Acceptance Boundary

- Focused Go tests cover normalization, wildcard/case matching, priority,
  weight, no-match error, API rejection, cache version, models aggregation,
  and migration preflight/cleanup.
- Group editor and API-key editor Vitest regressions, frontend typecheck,
  production build, migration dry-run/preflight, `git diff --check`, and
  unmerged-index checks pass.
- No selectable active group may remain ruleless at switch time, and no legacy
  `model_patterns` may remain after cleanup.

# S92 Addendum: user route priority-only editor

## Goal

Reduce API-key multi-group route configuration to group selection, drag order,
enabled state, add, and remove. The administrator-owned group model rules from
S91 determine model eligibility; ordinary users should not configure routing
weights, cooldowns, scope flags, presets, or model patterns.

## Scope Boundary

- Derive `priority` from the displayed route order as `index + 1` after every
  reorder, add, remove, and enable action.
- Load legacy routes sorted by positive priority, stable by source order, then
  renumber them continuously before display and save.
- Emit `weight=1`, `cooldown_seconds=30`, and current `enabled` for API
  compatibility; omit `model_patterns`, `image_only`, and `text_only`.
- Reject duplicate route groups in the editor and preserve account-pool
  strategy visibility and hidden-state payload behavior from S90.
- Do not change backend contracts, group rules, channel restrictions, billing,
  admin UI, translations, deployment, containers, or unrelated views.

## Acceptance Boundary

- Focused user KeysView tests prove order-derived priorities, legacy route
  normalization, fixed compatibility defaults, dropped legacy scope/model
  fields, duplicate rejection, and unchanged account-pool behavior.
- Frontend typecheck, production build, `git diff --check`, and unmerged-index
  checks pass.

# S93 Addendum: default API-key fallback group

## Goal

Let administrators choose one active group as the base fallback for every
system-created default API key, without replacing the purpose-specific default
routes used for chat, image, video, or other configured model patterns.

## Scope Boundary

- Store the optional group ID as `studio_bridge_luoye_ai.default_fallback_group`
  and expose it under System Settings -> External Access.
- New system-created default keys write that group to base `group_id` and retain
  configured `default_api_routes` as higher-priority multi-group routes.
- Bypass user group ownership checks only for system-created defaults; ordinary
  user-created API keys retain existing permission checks.
- Validate fallback compatibility at request time against active group context,
  platform, routing scope, and administrator-owned model rules.
- Provide an explicit, confirmed backfill action that updates only each user's
  lowest-ID non-deleted key when its base group is null. Preserve every other
  key field and invalidate only the changed auth-cache entries.
- Do not add a schema migration, run automatic backfill, change billing/account
  scheduling, deploy, update containers, commit, or push.

## Acceptance Boundary

- Default-tag service/handler tests cover creation, fallback routing, settings
  validation, backfill errors, success, and cache invalidation.
- PostgreSQL integration proves the guarded lowest-ID update and preservation of
  grouped defaults, secondary keys, routes, and soft-deleted predecessors.
- SettingsView Vitest covers settings round-trip, explicit confirmation,
  backfill invocation, and success count; typecheck and production build pass.
- `git diff --check`, conflict-marker scan, and unmerged-index checks pass.

# S97 Addendum: Redis ACL username compatibility

## Goal

Port the upstream Redis ACL username support as a local behavior slice. An
optional username must flow through runtime Redis clients, setup-time
connection tests, the setup wizard, and deployment examples without changing
the existing default-user behavior.

## Scope Boundary

- Add only the `redis.username` / `REDIS_USERNAME` field and its setup-wizard
  representation.
- Trim setup input, reject values longer than 128 characters, and pass the
  validated value to `redis.Options.Username`.
- Keep password, DB, TLS, host, port, Redis URL behavior, migrations, generated
  code, containers, deployment, and unrelated UI unchanged.
- Adapt the behavior manually to the local topology; do not merge upstream
  history.

## Acceptance Boundary

- Config, Redis option, and setup config tests prove empty and non-empty
  username behavior plus the length guard.
- English/Chinese setup labels, placeholders, form state, and review output
  remain type-safe; frontend typecheck and production build pass.
- README, config examples, `.env.example`, and all built-in Compose files
  forward `REDIS_USERNAME`.
- Allowlist, conflict-marker, unmerged-index, and `git diff --check` gates pass.

# S111 Addendum: upstream v0.1.164 isolated small fixes

## Goal

Port four low-risk `v0.1.164` behaviors that are missing from the local branch:
Grok CC Switch import, Grok HTTP 402 cooldown, day-aware model-rate-limit
display, and concrete GPT-5.6 Sol default ordering.

## Scope Boundary

- Adapt only upstream commits `a3a1575e9`, `ca0d3314c`, `48d58d72f`, and
  `dd5956be5` to the local URL normalization, Grok policy, and console theme.
- Preserve all existing non-Grok CC Switch behavior, model-status semantics,
  OpenAI aliases, and non-402 Grok error handling.
- Do not include schema, composite groups, Ollama Cloud usage, proxy stream
  quarantine, payment, billing, dependencies, deployment, or containers.
- Do not modify the separate group-buy S110 branch or worktree.

## Acceptance Boundary

- Focused Go tests prove 402 cooldown persistence/runtime blocking and
  concrete GPT-5.6 Sol default ordering through the package and admin API.
- Focused Vitest proves exact Grok CC Switch endpoints and day-aware status
  output without regressing existing account status behavior.
- Frontend typecheck/build, targeted lint, Go formatting, exact allowlist,
  diff, conflict-marker, and unmerged-index gates pass.

# S112 Addendum: OpenAI OAuth passthrough input normalization

## Goal

Port upstream `851436c55` and `3e26dfa5b` so the local OpenAI OAuth
passthrough path always sends a list-shaped `input` to the ChatGPT Codex
endpoint.

## Scope Boundary

- Normalize only the top-level `input` field in
  `normalizeOpenAIPassthroughOAuthBody`:
  non-blank strings become one user message, whitespace-only strings become an
  empty array, and single JSON objects become one-element arrays.
- Preserve existing arrays, absent input, compact stream/store behavior,
  unsupported-field removal, and all non-OAuth request paths.
- Do not modify routing, billing, auth policy, schema, migrations, frontend,
  deployment, containers, VERSION, group-buy changes, or the separate S110
  worktree.

## Acceptance Boundary

- Focused normalization tests cover string, whitespace-only string, object,
  and already-array input.
- Existing OAuth passthrough stream and compact forwarding tests continue to
  pass.
- `gofmt`, exact path audit, and `git diff --check` pass.

# S124 Addendum: v0.1.166 configuration and usage-query compatibility

## Goal

Port three bounded v0.1.166 behaviors missing from local `main`: honor an
explicit `CONFIG_FILE` path, allow administrators to list usage records by an
exact `request_id`, and resolve the user label when the administrator usage
view is opened with a `user_id` route query.

## Scope Boundary

- Reuse one config-source helper from both full config loading and bootstrap
  server-address loading. A non-blank `CONFIG_FILE` is the selected file;
  otherwise preserve the existing `DATA_DIR`, container, local, and system
  path search order. Environment-variable overrides remain intact.
- Add only an exact `request_id` condition to administrator usage list queries.
  Do not add schema, migration, export, dashboard-statistics, or fuzzy-search
  behavior.
- When `/admin/usage?user_id=<id>` is opened, fetch that user once and display
  the email in the existing filter. A concurrent user search must not be
  overwritten by the route lookup.
- Do not port v0.1.166 Caddy, affiliate-input, pricing, Grok, protocol,
  WebSocket, Antigravity, payment, or panel-rate-limit changes in S124.

## Acceptance Boundary

- Focused config, admin handler, and usage repository tests prove explicit
  file loading, existing fallback paths, trimmed exact request ID filtering,
  and unchanged missing-filter behavior.
- Focused UsageFilters and UsageView Vitest cover user-search revision safety
  and route-query label hydration. Frontend typecheck and production build
  must pass.
- `gofmt`, `git diff --check`, allowed-path audit, conflict-marker scan, and
  full Go compilation must pass. No deploy, container update, push, or primary
  dirty-worktree changes are in scope.

# S159 Addendum: Pixel Cafe payment callback and paid-full activation

## Goal

验证支付通知在真实 PostgreSQL 下的 Cafe 专属状态链：金额不匹配拒绝且不改变锁定订单，合法通知只完成一次订单与 Seat 付款，并在满员后创建保持 `disabled` 的受管 Key、strict Binding 和 active Round；相同通知重放必须幂等。

## Boundary

- 仅调用已验签之后的 `PaymentService.HandlePaymentNotification`，并复用真实 `GroupBuyService` 与 `CafeRoomActivationService`；不把 provider 签名、HTTP/JWT、真实支付商户、退款对账、Key enablement、网关路由或部署混入本 Sprint。
- 使用隔离 Testcontainers PostgreSQL 与现有生成 Ent schema；共享本机 `sub2api-*` 容器保持不变。

## S159 QA Result

- `PASS / runtime-isolated`: the invalid-amount guard, one successful paid-full activation and exact callback replay passed against fresh PostgreSQL. Provider signature verification, provider-pending refunds, real HTTP/JWT/gateway usage, Key enablement and deployment remain separate contracts.

# S160 Addendum: Pixel Cafe provider-pending refund reconciliation

## Goal

在隔离 PostgreSQL 中验证失败 Cafe Round 的 `pending_provider` refund：普通 GroupBuy lifecycle 必须隔离，Cafe lifecycle 对 query `pending` 保持可重试，对 query `success` 才完成 Order、Seat 和 GroupBuy refund，并在最终重放时不重复写入。

## Boundary

- 复用真实 `PaymentService.QueryAndFinalizeRefund` 和 Cafe lifecycle，仅使用进程内 query-provider fake；不接真实商户、webhook/HTTP、Key enablement、JWT/gateway、schema/migration 或部署。

## S160 QA Result

- `PASS / runtime-isolated`: ordinary GroupBuy reconciliation skipped the Cafe refund; query `pending` preserved retryable state and query `success` finalized Order/Seat/refund once against fresh PostgreSQL. Initial external refund submission, signature verification/HTTP, real merchant behavior, Key enablement and deployment remain separate gates.

# S161 Addendum: Pixel Cafe initial provider-refund submission

## Goal

In isolated PostgreSQL, verify the first Cafe provider-refund submission through the real `processSeatRefund -> PrepareRefund -> ExecuteRefund` chain. A local fake's `Refund()` returns `pending`, a replay must not submit again, and the existing Cafe-only query reconciler may finalize only after a later local `success` response.

## Boundary

- Reuse the real Cafe lifecycle, GroupBuy state machine, Payment refund services and a fresh Testcontainers PostgreSQL schema. The provider fake and its temporary instance exist only in the test process/database.
- Do not use an external merchant, credential, real payment provider, HTTP/webhook, Key enablement, JWT/gateway, migration/schema update, configuration, existing local container, deployment, or production write.

## Acceptance Boundary

- The initial pending response creates exactly one pending Order/refund/Seat/audit fact and one provider refund request with preserved request metadata. The pending replay is inert for provider submission and durable facts.
- A later query success completes Order/refund/Seat/event once; terminal replay submits or queries nothing further. This is isolated fake-provider evidence, not payment sandbox or merchant validation.

## S161 QA Result

- `PASS / runtime-isolated`: fresh PostgreSQL evidence covered the real initial provider-refund submission, durable `pending` state, no-resubmit replay, later Cafe-only query success, and terminal no-op replay. Provider signature/HTTP, real merchant or sandbox behavior, failure/partial paths, Key enablement, JWT/gateway, deployment and production validation remain separate gates.

# S162 Addendum: Pixel Cafe verified webhook handler boundary

## Goal

Use a real Gin Stripe webhook route and real PaymentService with local verifier and GroupBuy doubles to prove verifier failure is fail-closed, verified callback dispatches exactly once, and HTTP replay remains idempotent before the already-tested S159 Cafe activation service boundary.

## Boundary

- Use an in-memory Ent schema, registry provider double and GroupBuy call recorder only. Preserve raw body/header forwarding and payment order/audit facts through the actual handler and service code.
- Exclude provider signature algorithms, merchant credentials, external callbacks, actual Cafe activation, Key enablement, schema/migration, shared containers, deployment and production writes.

## Acceptance Boundary

- The local verifier records exact body/header and permits one designated signature. Verified request becomes completed and is dispatched once; exact replay remains HTTP-successful without another GroupBuy dispatch/audit.
- Forged signature returns 400 and preserves an independent pending order. This verifies handler sequencing, not real provider cryptography or merchant acceptance.

## S162 QA Result

- `PASS / handler-isolated`: a real Gin Stripe webhook route sent exact body/header to a local verifier; valid callback completed and dispatched once, replay was inert, and forged signature failed before mutation. The `unit` tag handler suite remains blocked by an unrelated missing `strconv` import, so it is not counted as passing evidence. Real provider cryptography, merchant callbacks, multi-instance selection, sandbox, Key enablement, JWT/gateway, deployment and production validation remain separate gates.

# S163 Addendum: Stripe signed order-bound webhook selection

## Goal

Extract the `metadata.orderId` written on Stripe PaymentIntent creation before webhook provider lookup, so verified callbacks select the PaymentOrder's pinned Stripe instance in a multi-instance configuration.

## Boundary

- Use real Gin handler, PaymentService, encrypted temporary in-memory provider instances, and Stripe SDK local signature helpers. Assert only local handler/order/dispatch facts.
- Exclude external Stripe API access, merchant configuration, route/schema changes, Key enablement, Cafe activation, shared containers, deployment and production writes.

## Acceptance Boundary

- A callback signed with the bound second instance's synthetic secret must complete and dispatch exactly once; the same event signed with the other instance's secret must return 400 before mutation. Replays are inert.
- This proves local SDK cryptography plus order-bound candidate selection, not live Stripe webhook delivery or merchant endpoint configuration.

## S163 QA Result

- `PASS / handler-isolated`: Stripe `metadata.orderId` now selects the bound temporary provider instance before verification. The real Stripe SDK accepted the bound second instance's synthetic local signature, rejected the first instance's signature for a second-bound order, and retained replay idempotency. Live Stripe merchant behavior, other providers, sandbox, Key enablement, JWT/gateway, deployment and production validation remain separate gates.

# S133 Addendum: administrator group duplication

## Goal

Port upstream `9fc006546` so administrators can create a reviewable inactive
copy of an existing group without rebuilding omitted configuration from the
list response.

## Scope Boundary

- Copy the current persisted group configuration and eligible account binding
  priorities atomically; do not copy group identity, timestamps, active status,
  or user/API-key ownership.
- Use the existing administrator idempotency boundary plus one internal
  nullable group operation identity to recover an already committed copy after
  an ambiguous response failure.
- Expose only the duplicate action and localized result state in the existing
  administrator group list. Do not change unrelated group editing, routing,
  pricing, or account-priority workflows.
- Adapt the schema change to migration `199`; do not reuse upstream migration
  number `181` or touch production data.

## Acceptance Boundary

- Focused service, handler, repository, API-client, and group-view tests cover
  configuration cloning, unique names, inactive status, exact binding
  priorities, OAuth-only filtering, and retry recovery.
- Ent/Wire generation, repository compile/build, frontend typecheck/build,
  migration ordering/content, formatting, allowlist, conflict-marker,
  unmerged-index, and `git diff --check` gates pass.
- No authenticated production browser, production migration, push, deployment,
  container refresh, or real provider/upstream call is performed.

## S133 QA Result

- `PASS / source-level`: the focused service and handler regressions, frontend
  API/view regressions, Ent/Wire generation, full Go compile/build probes,
  frontend typecheck/build, migration-content review, and Git integrity checks
  passed in the isolated worktree.
- Existing `unit` API-contract assertions and the `integration` repository
  package are stale outside S133, so no claim is made for those broad suites or
  for a real PostgreSQL migration/runtime transaction.
# Pixel Cafe Phase 18: Airwallex order-bound webhook evidence (S164)

- Validate the existing Airwallex `merchant_order_id` candidate extraction at the real Gin handler boundary with two encrypted temporary instances and distinct local HMAC secrets.
- A valid callback must select the bound instance, verify before mutation, dispatch GroupBuy once, and remain idempotent on replay; a signature from the other enabled instance must fail before mutation.
- This is local handler-isolated evidence only. It excludes Airwallex API access, merchant credentials, endpoint registration, Key enablement, shared containers, deployment, and all production writes.

## Pixel Cafe Phase 19: WeChat Pay encrypted webhook candidates (S165)

- Validate the existing all-enabled-candidate WeChat Pay callback path with locally generated RSA signatures and AES-256-GCM transaction resources. The second candidate must verify/decrypt and dispatch exactly once after the first candidate fails.
- The wrong-instance signature over second-key ciphertext must return 400 before mutation; identical body/header replay must be idempotent.
- This is offline handler evidence only. It excludes WeChat API/merchant access, real certificates, Key enablement, shared containers, deployment, and production writes.

## Pixel Cafe Phase 20: EasyPay webhook pinned instance selection (S166)

- Validate the existing EasyPay GET callback parser and order-bound provider selection at the real Gin handler boundary with two temporary encrypted instances and distinct local MD5 keys.
- A valid second-instance callback must select the bound instance, verify before mutation, dispatch GroupBuy once, and remain idempotent on replay; a signature from the other enabled instance must fail before mutation.
- This is local handler-isolated evidence only. It excludes EasyPay API access, merchant credentials, endpoint registration, Key enablement, shared containers, deployment, and all production writes.

## S166 QA Result

- `PASS / handler-isolated`: a real Gin GET route, two temporary encrypted EasyPay instances, the real canonical MD5 verifier and PaymentService proved valid bound-instance fulfillment, exact replay idempotency, wrong-key rejection and malformed/blank parser handling. No production correction was required.
- Whitespace-tolerant `out_trade_no` remains intentionally unimplemented: handler-only trimming cannot normalize the later provider notification value used by PaymentService. It requires a separately owned cross-layer contract.

## Pixel Cafe Phase 21: Alipay webhook pinned instance selection (S167)

- Validate the existing Alipay POST callback parser and order-bound provider selection at the real Gin handler boundary with two temporary encrypted instances and distinct locally generated RSA key pairs.
- A valid second-instance RSA2 callback must select the bound instance, verify before mutation, dispatch GroupBuy once, and remain idempotent on replay; a signature from the other enabled instance must fail before mutation.
- This is local handler-isolated evidence only. It excludes Alipay API access, merchant credentials, endpoint registration, Key enablement, shared containers, deployment, and all production writes.

## S167 QA Result

- `PASS / handler-isolated`: a real Gin POST route, two temporary encrypted Alipay instances, the real smartwalle RSA2 verifier and PaymentService proved valid bound-instance fulfillment, exact replay idempotency, wrong-key rejection and malformed/blank parser handling. No production correction was required.
- The SDK provider tests require the existing `unit` build tag; the accepted tagged command covered configuration, merchant metadata and amount parsing. This is not merchant, endpoint, sandbox or production callback evidence.

## Pixel Cafe Phase 22: Mocked browser user-flow evidence (S168)

- Exercise the feature-enabled authenticated `/group-buy` page in an actual browser using only synthetic public settings, Room/Lobby/my-room payloads, and a locally intercepted order response. Prove Room discovery, zone change, accessible Room selection, empty-Seat selection, agreement gating, local payment-waiting UI, and 390px no-overflow behavior.
- A fresh anonymous visit must redirect to login, and a false `pixel_cafe_enabled` setting must preserve the legacy Group Buy page. Screenshot evidence must remain under `output/playwright/pixel-cafe-s168/`.
- This is mocked-browser UI evidence only. It excludes real JWT/API, database/Redis, payment/provider, Key activation/routing, performance, deployment, rollback, and production validation.

## S168 QA Result

- `PASS / mocked-browser`: local `/api/v1/**` fixtures exercised the actual Vue/Router feature-enabled Cafe page, Claude-zone Room discovery, keyboard Room selection, empty-Seat selection, agreement gate, local payment-waiting transition, anonymous redirect, disabled-feature legacy fallback, and a 390px no-overflow layout. The final clean session had zero console errors; the focused page suite passed `10/10` and frontend typecheck passed.
- The browser assertion captured request construction only. No real JWT/API, persisted order or Seat lock, payment/provider confirmation, managed-Key activation/routing, database/Redis, performance, deployment, rollback, or production validation occurred.

## Pixel Cafe Phase 23: authenticated HTTP ownership smoke (S169)

- Exercise the registered production `/api/v1/cafe` routes through real Gin JWT authentication and fresh Testcontainers PostgreSQL data. Verify that each authenticated user receives only their own My Rooms membership while Room discovery remains redacted, and that an unauthenticated request is rejected before Cafe data is read.
- Verify feature-disabled failure and the mandatory agreement guard before any payment provider or Seat mutation. This contract intentionally stops before successful payment/order creation because that requires a separate payment sandbox and provider configuration scope.
- No shared-container data, real user session, managed-Key enablement, payment/provider call, schema/migration, deployment, rollback, staging, or production operation is permitted.

## S169 QA Result

- `PASS / runtime-isolated`: an integration-tagged test used actual `RegisterUserRoutes`, real JWT signing/validation and UserRepository, real Cafe public queries, and fresh Testcontainers PostgreSQL. Each persisted user saw only its own My Rooms membership; public projections excluded deliberate private metadata; absent auth, disabled feature, and pre-payment agreement failures were enforced at the registered HTTP route.
- The temporary test database received the existing `user_avatars` table shape needed by the actual repository, not a migration-file or shared-database operation. Successful order creation/payment, Key activation/routing, external provider behavior, shared Redis, performance, deployment, rollback, and production remain unverified.

## Pixel Cafe Phase 24: authenticated server-side order creation (S170)

- Exercise the registered authenticated Cafe order route with fresh Testcontainers PostgreSQL, real JWT/UserRepository/idempotency persistence, real payment configuration/instance selection, real Cafe Seat locking and a temporary enabled EasyPay `popup` provider instance. The real provider may construct only its hosted `submit.php` URL from synthetic non-routable values; it must not issue any HTTP request or contact a merchant.
- A successful request must persist one pending group-buy PaymentOrder and one locked Seat/reservation. A byte-identical request with the same `Idempotency-Key` must replay the stored response without a second Order, Seat, event, audit, or durable idempotency record. A separately configured provider-construction failure must persist `failed`, release the Seat, and clear reservations.
- This is isolated route/order evidence only. It excludes payment confirmation/callback, real provider credentials/endpoints, Key activation/routing, Redis, refund/lifecycle, schema/migration, shared containers, performance, deployment, rollback, staging, and production writes.

## S170 QA Result

- `PASS / runtime-isolated`: the integration-tagged registered route used real JWT middleware, `UserRepository`, SQL-backed idempotency, payment selection, `PaymentService`, `GroupBuyService`, and `CafeRoomOrderService` against fresh Testcontainers PostgreSQL. The successful request persisted one pending group-buy PaymentOrder, locked Seat/reservation, lock event and creation audit; same-key replay returned the same semantic success envelope with the replay header and did not duplicate those facts.
- A separately configured temporary EasyPay popup instance reached real provider construction, failed on missing required configuration without network access, and left its Order `failed`, its Seat `released`, and the Round reservation counters at zero. The fixture adds only the migration 057 timestamp defaults in its disposable database; no production code, migration, schema, provider, shared container, Key state, Redis, payment callback, deployment, or production data was changed.

## Pixel Cafe Phase 25: PostgreSQL concurrency convergence (S171)

- Extend S158/S159 evidence without changing those tests: release 100 distinct final-Seat requests together under bounded database concurrency, require exactly one durable winner and 99 explicit `CAFE_SEAT_UNAVAILABLE` losers, then release 100 paid-full activation retries together and require one coherent active Round.
- The activation result must have exactly one disabled managed Key and one active strict Binding per Seat, one activation event, stable activation timing/token facts, and no ordinary subscription. Infrastructure errors cannot count as expected losers.
- This is fresh Testcontainers PostgreSQL evidence only. It excludes payment/provider/callback, Key enablement or gateway use, Redis, frontend, schema/migration, shared containers, performance benchmarking, deployment, and production readiness.

## S171 QA Result

- `PASS / runtime-isolated`: one fresh PostgreSQL run classified 100 final-Seat requests as one durable winner and 99 explicit unavailable losers, and 100 paid-full activation calls converged to one active four-Seat Round with exactly four disabled managed Keys, four active strict Bindings and one activation event.
- Three additional consecutive fresh runs also passed. No product correction was needed. This closes the bounded concurrency correctness gap but does not prove performance, provider/payment, enabled-Key gateway usage, Redis, deployment or production readiness.

# Upstream Billing Rate Sync Addendum (S204)

## Goal

Port the upstream Sub2API declared-rate contract and account probing chain, then add a per-account opt-in that automatically maintains the local account cost multiplier from a validated upstream base rate.

## Product Boundary

- A downstream Sub2API deployment exposes only billing declaration facts available to the authenticated API Key. It does not expose provider credentials, account cost, proxy data, or unrelated administrative state.
- Account probing is explicit and bounded. It supports manual probe, batch probe, periodic due selection, status/backoff snapshots, proxy-aware HTTP access, and identity-change invalidation.
- Automatic synchronization is disabled by default. When enabled, only the declared base rate excluding peak-time uplift may write to `accounts.rate_multiplier`; effective peak-time rate remains display/probe metadata.
- The write-back shares the snapshot CAS boundary so an old probe cannot clobber concurrent account edits. Successful write-back records `synced_rate_multiplier` and a structured log fact.
- Synchronization owns the account rate while active: single and bulk manual edits fail closed, while disabling sync and setting a manual rate atomically is allowed.
- This Sprint does not add upstream-cost profitability scheduling, margin gates, schema/migration, provider account admission for media, deployment, or production validation.

## Acceptance Boundary

- Backend unit and integration-shaped repository tests cover auth, declaration calculation, probe eligibility, URL/proxy safety, persistence/backoff, CAS, value governance, concurrent edit protection, and manual-edit conflicts.
- Frontend component tests cover probe actions/status, precision display, sync/probe coupling, managed-rate UI, and bulk conflict handling.
- Wire generation, complete service regression, server compile, frontend typecheck/build, scoped diff and Git integrity checks must pass before local-main integration.
# Leaderboard Record Banner Addendum (S180)

## Goal

Move the existing dynamic personal record into a new text-free wide banner at the bottom of the left ranking
panel. Remove the right-side Thursday promotion and standalone record card so the right column focuses on the
weekly top-10 and reward facts with more vertical space.

## Boundary

- Frontend presentation, localization, focused tests, generated raster and browser evidence only. Reuse current
  record/reward computed state and APIs; no backend, settlement, claim, access, database, container or production
  change is included.
- The banner is decorative art with no embedded copy, logo or user data. Dynamic personal status remains HTML
  text so every ranking and waiting state stays localized and accessible.

# Upstream v0.1.173 Selective Fixes Addendum (S207)

## Goal

Behaviorally port five bounded upstream fixes: Gemini actual-output image billing, Gemini pool-mode 429 account-state
protection, Web Search missing-setting and dialog-scroll handling, Grok OAuth nil-client protection, and the locally
applicable Antigravity fallback-model correction from the upstream response-observation optimization.

## Boundary

- Preserve the local long-diverged service topology and existing product customizations; never merge the release as a whole.
- Treat image billing and account-level 429 state as financial/availability boundaries with explicit failover and policy tests.
- `6e34fb09c` may correct the model actually sent after Antigravity fallback, but S207 excludes its absent `db0bff82c`
  prerequisite, database fields, migrations, usage audit UI, and response-observer persistence chain.
- No provider call, schema/migration, shared resource, deployment, container update, remote push, or production validation.

# Upstream API Key Validation Addendum (S209)

## Goal

Reject invalid API Key quota, rate-limit, and create-expiry inputs at both the
HTTP and service boundaries, adapting upstream `f5c108c83` without merging its
divergent file history.

## Boundary

- `quota`, `rate_limit_5h`, `rate_limit_1d`, and `rate_limit_7d` must be finite
  and non-negative. Zero remains the existing unlimited value.
- Create-only `expires_in_days` must be nil or greater than zero. Update's
  valid RFC3339 expiration and empty-string clear behavior remain unchanged.
- Handler rejection occurs before Create idempotency execution or Update
  service invocation; service validation independently protects internal calls.
- No persistence, schema/migration, routing, billing, Cafe managed-Key,
  dependency, frontend, configuration, container, push, deployment, shared
  resource, or production behavior changes are included.

# Account Time Availability Addendum (S212)

## Goal

Allow an administrator to opt an individual account into one same-day daily
availability window. The window affects only new-request candidate selection;
it does not mutate the account's real status, schedulable flag, API Key
bindings, or group membership.

## Boundary

- Store `account_availability_enabled`, `account_availability_start`, and
  `account_availability_end` in the existing account `extra` JSON. No schema
  migration or new public top-level API field is needed.
- Use server-local time and a left-closed, right-open `[start, end)` interval.
  Cross-midnight and multi-window schedules are deliberately unsupported.
- Capture one request-start timestamp at middleware entry. Gateway, OpenAI,
  Gemini, sticky-session, and automatic failover selection must use that same
  timestamp; internal callers without it fall back to current server time.
- Disabled or absent configuration preserves current behavior. A manual
  inactive/error state, expiry, rate-limit, quota, overload, temporary pause,
  capability restriction, and account-pool policy keep their existing priority.
- The administrator create/edit account forms expose the window and preserve a
  valid configured window when the toggle is switched off.

## Acceptance Boundary

- Focused service tests prove valid configuration, invalid write rejection,
  start/end boundaries, disabled behavior, manual-state precedence, request
  start-time stability, sticky/pinned exclusion, and no-candidate behavior.
- Gateway, OpenAI, Gemini-compatible, snapshot, and private-pool selector
  paths exclude accounts outside the window without changing persisted state.
- Middleware tests prove the request timestamp is written exactly once.
  Frontend component tests prove both account dialogs persist valid window data
  and reject incomplete or reversed windows.
- Focused/full Go tests, handler/server compilation, frontend Vitest/lint/
  typecheck/build, a task-owned Playwright desktop and 390px check, formatting,
  `git diff --check`, allowlist, and unmerged-index gates must pass.

# Upstream v0.1.176 Correctness Addendum (S213)

## Goal

Behaviorally port four bounded correctness fixes from the upstream `v0.1.176`
window: avoid sticky false OpenAI Responses probe results, align channel pricing
conflict validation with cache normalization, invalidate channel caches after a
persisted group-platform change, and elect one scheduled-backup instance.

## Boundary

- Preserve the local divergent service, pricing, group, and backup topology;
  never merge the release as a whole.
- Treat the four fixes as independent slices. One failed slice does not block
  separately verified slices from local integration.
- Cache invalidation must follow successful persistence and retain existing
  API-key auth invalidation. Scheduled-backup locking must reuse existing local
  PostgreSQL advisory locks, fail closed on acquisition errors, and leave manual
  backup/restore behavior unchanged.
- Grok 4.6/JWT tier/x_search/media, unknown-model pricing, group per-model
  pricing, long-context pricing controls, schema/migrations, frontend, provider
  calls, containers, deployment, push, and production traffic are excluded.

## Acceptance Boundary

- Focused repeated tests cover every new verdict and invalidation/lock branch;
  complete affected service/server packages and server compilation pass.
- Generated wiring, formatting, exact allowlist, upstream provenance, conflict,
  unmerged-index, and preservation of the user-owned frontend edits and
  `outputs/` must pass before local-main integration.

# OAuth Pending Account Takeover Addendum (S214)

## Goal

Close the confirmed OAuth pending-exchange account-takeover path from upstream
security fix `02e50cc22` without merging unrelated upstream history.

## Security Boundary

- A non-terminal pending OAuth session, including
  `choose_account_action_required`, may describe a possible existing account
  but must never bind an OAuth identity, mutate that account through profile
  adoption, consume the pending session, or issue tokens.
- `invitation_required` is the sole non-terminal exception for adoption
  decision persistence: it may retain the existing decision-only write before
  returning its payload, but it still must never bind an identity, mutate a
  profile, consume the session, or issue tokens.
- A terminal login session that already satisfies
  `pendingOAuthCompletionCanIssueTokenPair` may retain its existing identity
  binding, profile-adoption, session-consumption, and token behavior.
- An authenticated `bind_current_user` intent may retain its existing explicit
  identity-binding behavior because its target comes from the authenticated
  bind flow rather than an attacker-submitted email.

## Boundary

- Preserve the existing `invitation_required` decision-only return branch, then
  enforce the invariant at `ExchangePendingOAuthCompletion` before all other
  adoption decision persistence or identity binding, using the existing
  `canIssueTokenPair` result and normalized `bind_current_user` intent.
- Add a realistic handler regression that recreates the choice-state attack
  and proves no identity binding, victim profile mutation, token issuance, or
  session consumption occurs.
- No route, request/response schema, database schema, migration, dependency,
  frontend, OAuth provider callback, deployment, container, production, push,
  or S213 behavior change is included.

## Acceptance Boundary

- The new exploit regression must be shown to fail against the vulnerable
  implementation before the guard is added, then pass repeatedly after it.
- Existing terminal login and `bind_current_user` tests must pass repeatedly to
  prove legitimate identity binding remains intact.
- The complete handler package, server compilation, formatting, exact
  allowlist, conflict-marker, unmerged-index, and upstream provenance gates
  must pass before local-main integration.

# Upstream Grok Correctness Addendum (S215)

## Goal

Behaviorally port the two locally reachable Grok correctness fixes from the
upstream `v0.1.176` window: a real fallback price for unregistered Grok text
models, and correct Grok snapshot use in account badges and incremental list
refresh.

## Boundary

- Add a local static Grok 4.5 fallback only after dynamic pricing misses, with
  an explicit media exclusion so per-unit image/video/audio IDs cannot inherit
  token pricing.
- The account UI treats `grok_usage_snapshot` as canonical because that is the
  locally persisted backend writer key; `grok_quota_snapshot` remains legacy
  fallback only. Refresh-key serialization is stable under object-key order.
- Realtime audio billing is excluded because the entire local Voice/Realtime
  path is absent. Group long-context migration/default behavior is excluded
  because group model-pricing fields, schema and migration prerequisites are
  absent.
- No schema/migration, route, provider call, dependency, deployment, container,
  push, production traffic, user-owned modal edit, or `outputs/` operation is
  included.

## Acceptance Boundary

- Each slice has default-tag discovered regressions, repeated focused tests and
  independent Terra QA. Complete affected backend package/server compilation or
  frontend typecheck/lint must pass as applicable.
- Final integration requires scope/topology/provenance review, format/diff,
  conflict-marker, unmerged-index gates, and preservation of user-owned files.

# Upstream v0.1.177 Group Pricing And Long Context Addendum (S220)

## Goal

Port the complete `f3d949107 -> b830bc14d -> fd82dfd52` chain so groups can
override model pricing and disable long-context ladders without allowing an
OpenAI account-only setting to disable Grok billing tiers.

## Boundary

- Migration 221 adds group model-pricing JSON and a long-context switch that is
  true for new and existing groups. Migration source and disposable validation
  are authorized; shared and production database execution are excluded.
- Resolution is Group -> Channel -> built-in. OpenAI requires both group and
  account switches; non-OpenAI platforms use the group switch only.
- Preserve local GPT-5.6, media, profit-control, peak-rate, billing, and UI
  customizations. No fingerprint, daily rollup, dependency, provider,
  deployment, container, push, user account-modal, or `outputs/` change.

## Acceptance Boundary

- Focused repeated billing/resolver/OpenAI/Grok regressions, migration 221
  validation, generated Ent consistency, complete affected backend packages,
  frontend group-pricing tests/typecheck/build, format/diff, allowlist,
  conflict/index, and upstream provenance must pass.

# Upstream v0.1.177 Codex Fingerprint Convergence Addendum (S221)

## Goal

Port the opt-in fingerprint convergence behavior from `c0ab3a00e` and the
remaining non-turn-state parts of `fce41e318`, including raw passthrough client
metadata rewriting and admin controls.

## Boundary

- Missing/explicit `off` preserves client identifiers; `device`, `session`, and
  `full` are explicit administrator opt-ins. Existing S219 turn-state behavior
  remains independent.
- The user-owned `EditAccountModal.vue` and test patch are a required baseline,
  not disposable conflicts. The final result must preserve that patch while
  adding the fingerprint control and default behavior.
- No schema/migration, provider, dependency, daily rollup, deployment,
  container, push, or unrelated account behavior.

## Acceptance Boundary

- Backend raw/map metadata parity, header/body consistency, failover clearing,
  mode defaults, normal/passthrough routing, focused frontend modal tests,
  complete affected regressions, and exact preservation of the user patch must
  pass under independent Terra QA.

# Upstream v0.1.177 Group Usage Daily Rollups Addendum (S222)

## Goal

Port `cb7b03795` plus corrections `89d826be2` and `45dcce0e4` so group usage
summary reads persistent closed-day rollups and a live tail instead of scanning
the full usage-log history.

## Boundary

- Migrations 222/223, configured-timezone date boundaries, startup backfill,
  periodic synchronization, publication watermark, cleanup invalidation, today,
  yesterday, and total cost are in scope.
- Include only the timezone-test corrections from the follow-up commits. Exclude
  their Go/Node version, dependency lock, CI, release, and security-workflow
  upgrades.
- Database validation uses only fresh task-owned fixtures. Shared/production DB,
  provider, deployment, container, push, and unrelated dashboard changes remain
  excluded.

## Acceptance Boundary

- Migration trigger/watermark/timezone/DST, historical mutation/cleanup,
  startup and scheduled synchronization, API/UI yesterday display, complete
  affected backend/frontend regressions, format/diff, allowlist, conflict/index,
  and upstream provenance must pass under independent Terra QA.

# Upstream GPT/Codex Quota Correctness Addendum (S217)

## Goal

Behaviorally port the locally missing GPT/Codex quota correctness behavior from
upstream `v0.1.176`: personal subscription expiry must not inherit a workspace
entitlement, HTML 403 responses must not punish OpenAI accounts, and reset-credit
actions must leave the account list and credit state consistent without inducing
a second non-refundable reset.

## Boundary

- Port `358e4a89a`, `12abb5470`, and only the remaining client/API behavior of
  `54a2bcfd1`; adapt all patches to the local topology and tests instead of
  cherry-picking them.
- Preserve the existing S188 ordering: successfully consumed reset credit first
  recovers account state with detached bounded post-processing, then optional
  quota/cache refresh work may warn but cannot turn success into failure.
- `99b31067f` and `3d3aee2e` are already behaviorally covered by the existing
  OpenAI eligibility path. The absent upstream cross-platform threshold feature
  is explicitly out of scope.
- No schema/migration, dependency, provider traffic, production data, container,
  deployment, push, user-owned account-modal edit, or `outputs/` operation is
  included.

## Acceptance Boundary

- Tests prove accounts/check can no longer pair a personal plan with a different
  workspace expiry; the fallback personal subscription lookup remains test-local.
- Tests prove HTML 403 neither increments account penalty state nor changes
  schedulability, while structured OpenAI 403 and non-OpenAI behavior remain
  unchanged.
- Tests prove reset response metadata updates the visible account and credit
  state without an automatic second quota request; a missing post-reset quota
  cannot offer stale positive credits or turn a consumed reset into an apparent
  retryable failure. An API contract proves the snapshot-persisting refresh is
  an audited POST while the existing GET remains read-only.
- Focused repeated backend and frontend regressions, complete affected backend
  packages, server compilation, typecheck/build when dependencies are available,
  format/diff, allowlist, conflict-marker, unmerged-index, provenance, and
  preservation of user-owned files must pass before local-main integration.

# Image Model Tutorials Addendum (S223)

## Goal

Publish local Chinese API calling guides in Tutorial Management for the nine
image models shown by the user, using the existing tutorial CMS and the actual
`ai.3zapi.top` gateway behavior.

## Boundary

- Add one migration with nine published tutorial pages and one focused migration
  content test. Existing tutorial CRUD, Vue layout, fallback content, gateway
  routing, providers, dependencies, containers, deployment, and user data stay
  unchanged.
- Use the external reference only for information hierarchy and parameter
  research. Remove its brand, host, key-management copy, support links, and
  verbatim prose.
- GPT, Gemini, and Midjourney document the local final OpenAI-compatible image
  response. Seedream documents submit plus `/v1/tasks/{task_id}` polling because
  it does not use the local internal image-task polling branch.
- Source migration only: do not execute against shared or production databases.

## Acceptance Boundary

- Static tests prove exact nine-model coverage, published/category state,
  local-host examples, collision-safe inserts, and absence of reference branding.
- Migration package tests, server compile, UTF-8/content review, scoped diff,
  conflict/index checks, and desktop/mobile tutorial rendering must pass under
  independent Terra QA before final PASS.

# Upstream Billing Quantization Addendum (S224)

## Goal

Quantize every `UsageBillingCommand` monetary amount to PostgreSQL
`NUMERIC(20,8)` before persistence while preserving the raw-value request
fingerprint used for idempotency.

## Boundary

- Quantization runs after fingerprint generation and covers the local prepaid
  balance field in addition to all upstream monetary fields.
- Reuse the existing decimal dependency. SQL, migrations, repositories, cost
  calculation, routing, frontend, provider calls, containers, deployment, push,
  and production operations are excluded.

## Acceptance Boundary

- Default-tag tests prove rounding boundaries, exact balance/quota reconciliation,
  all monetary fields, raw fingerprint ordering, explicit fingerprints,
  nonfinite values, and negative values.
- Complete service/repository regression, formatting, exact allowlist,
  provenance, and dirty-worktree protection require independent Terra QA.

# Upstream Fingerprint User-Agent Validation Addendum (S225)

## Goal

Behaviorally port upstream `fe2c265c9` so malformed, local-build, or implausible
Claude CLI User-Agent values cannot become an account's persistent fingerprint,
and existing poisoned cache entries recover automatically.

## Boundary

- Validate both first creation and version upgrade. A valid client User-Agent
  heals a poisoned cache; otherwise use the existing local default. Both healing
  paths preserve the cached `ClientID` and do not replace the cache interface.
- Keep the exact local `claude-cli/2.1.92` and Stainless defaults unchanged.
  `claude.CLICurrentVersion` is used only to derive a reasonable Claude CLI
  major-version ceiling. Non-Claude products receive syntax validation without
  the Claude-specific major ceiling.
- No Redis/TTL, gateway, request body, frontend, dependency, schema/migration,
  provider, container, deployment, push, production, user-owned dirty file, or
  `outputs/` change is included.

## Acceptance Boundary

- Default-tag focused tests cover syntax and length rejection, local/dev/build
  suffixes, sentinel major versions, valid non-Claude products, creation,
  upgrade, healthy-cache no-op, both poisoned-cache healing branches, default
  fallback, exact local defaults, and `ClientID` preservation.
- Complete service regression, server compilation, formatting, exact allowlist,
  provenance, conflict/index checks, and dirty-worktree protection must pass
  under independent Terra QA before local-main integration.

# Upstream CN Providers Addendum (S226)

## Goal

Behaviorally port the locally reachable Kimi, Zhipu, and DeepSeek first-class
support from `901a0439f`, including the B3 stream timeout and B4 probe URL
security fixes in `4b667ccd4`, without merging the divergent upstream history.

## Boundary

- S226-A establishes platform constants, account modes, protocols, default Base
  URLs, credential access, authentication, and model-list behavior without
  enabling gateway routes.
- S226-B adds Coding Plan quota and payg balance probes, periodic detection,
  URL-policy enforcement, admin endpoints, and Wire integration.
- S226-C adds exact-platform scheduling, Chat Completions/native Anthropic/
  DeepSeek Responses forwarding, count-tokens behavior, bounded stream pumping,
  reactive 429/reset handling, and recoverable balance pauses.
- S226-D adds account create/edit controls, presets, status cells, platform
  visuals, types, and bilingual strings while preserving the user modal patch.
- S226-E is integration and independent QA only; it adds no product behavior.
- Exclude the accidental root Docker Compose file, the absent
  `user_platform_quotas` product and migration, the absent generic scheduling-
  threshold product, dependencies, deployment, provider calls, shared data,
  push, and all unrelated local changes.

## Acceptance Boundary

- Each A-D batch is one independently compiling implementation commit with
  focused repeated tests and Controller approval before the next batch starts.
- B4 tests prove a rejected final probe URL sends no request and cannot expose
  an API key. B3 tests prove all four native stream loops time out, close their
  bodies, and preserve accumulated usage semantics.
- S226-E reruns complete affected backend/frontend suites, typecheck/build,
  desktop/mobile UI inspection with task-owned browser cleanup, exact allowlist
  and provenance checks, and both user patch-ID checks under independent Terra
  QA before any local-main integration.
