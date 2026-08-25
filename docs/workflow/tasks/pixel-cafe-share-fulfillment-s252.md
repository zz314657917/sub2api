# Pixel Cafe Share Fulfillment S252

## Task ID

pixel-cafe-share-fulfillment-s252

## Role

Planner/Generator/Evaluator: implement the approved Pixel Cafe share-purchase,
deferred-account fulfillment, membership, activation, refund, public UI, and
administrator workflow in isolated task worktrees based on current local
`main`. Developer and independent QA workers use `gpt-5.6-terra`; Codex remains
the final evaluator and alone applies reviewed patches to the dirty primary
worktree.

## Goal

Replace Pixel Cafe's pre-bound one-user/one-seat model with configurable shares
and buyer caps. A full Round waits for a matching ChatGPT Plus/Pro Account,
activates one managed Key per participant only after assignment, and refunds all
paid purchases if fulfillment has not succeeded within 24 hours.

## Success Criteria

- Room plans support `plus|pro`, 1-10 total shares, configurable buyer/user
  caps, a 1440-minute fulfillment timeout, and per-share managed-Key limits.
- New Rooms and open Rounds do not require an Account. Each Round freezes the
  fulfillment policy needed for later account validation and Key creation.
- A durable `(round_id,user_id)` Cafe membership aggregates multiple purchase
  batches, supports top-ups while open, and receives exactly one managed Key.
- Purchase locking serializes remaining shares, distinct buyer count, per-user
  shares, payment idempotency, and expiration release without overselling.
- Paid-full Rounds transition `open -> awaiting_account`; account assignment
  validates active OpenAI Plus/Pro compatibility and uniqueness, then performs
  Key/Binding/membership activation atomically.
- Activation time, not Room/Round creation or paid-full time, starts entitlement
  validity. Per-user Key limits equal per-share limits multiplied by paid shares.
- Unfulfilled Rounds transition to the existing idempotent refund machinery at
  24 hours and cannot report `refunded` until every paid batch is settled.
- Public Cafe APIs/UI use share and buyer vocabulary, expose one anonymous avatar
  per paid participant, and never expose Account data before activation.
- The administrator Room editor no longer binds Accounts. A pending-fulfillment
  workspace provides tier-filtered server-side account search and assignment.
- Existing generic group-buy behavior, already-active legacy Cafe bindings, the
  current Pixel Cafe scene work, unrelated dirty files, and `outputs/` survive.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `AGENTS.md`, `docs/workflow/status.md`,
  `docs/workflow/spec.md`, this contract.
- User-approved product plan is recorded in the S252 addendum in
  `docs/workflow/spec.md`.
- The primary worktree already contains user-owned Pixel Cafe scene/page/demo,
  Group, Settings, and knowledge edits. Work on overlapping frontend files must
  preserve those behaviors and tests; no reset, checkout, cleaning, or broad
  formatting is allowed.

## Allowed Paths

- `backend/migrations/235_pixel_cafe_share_fulfillment.sql`
- `backend/migrations/pixel_cafe_share_fulfillment_test.go`
- `backend/ent/schema/group_buy_plan.go`
- `backend/ent/schema/group_buy_round.go`
- `backend/ent/schema/group_buy_seat.go`
- `backend/ent/schema/api_key_account_binding.go`
- `backend/ent/schema/cafe_round_membership.go`
- `backend/ent/**` only for deterministic Ent-generated changes caused by the
  approved schema fields/entity; no unrelated generated churn
- `backend/internal/domain/**` only for the immutable Cafe fulfillment snapshot
- `backend/internal/repository/cafe_room_repo.go`
- `backend/internal/repository/*cafe*test.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/api_key_repo_*test.go` only for strict
  membership-binding lookup and legacy-seat compatibility
- `backend/internal/service/cafe_room*.go`
- `backend/internal/service/cafe_*test.go`
- `backend/internal/service/api_key.go`
- `backend/internal/service/api_key_auth_cache_impl.go`
- `backend/internal/service/*api_key*test.go` only for the new managed-source
  strict-binding boundary and legacy compatibility
- `backend/internal/service/group_buy.go` and focused group-buy tests only for
  Cafe payment callback, expiry, and refund coordination
- `backend/internal/handler/cafe_handler.go`
- `backend/internal/handler/cafe_handler_test.go`
- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `backend/internal/handler/wire.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/*cafe*test.go`
- `backend/cmd/server/wire_gen.go` only for the deterministic handler constructor
- `frontend/src/types/pixelCafe.ts`
- `frontend/src/api/cafe.ts`
- `frontend/src/api/admin/cafeRooms.ts`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/demoData.ts`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/views/admin/pixelCafe/**`
- `frontend/src/i18n/locales/{zh,en}/admin/pixelCafe.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/pixel-cafe-share-fulfillment-s252.md`
- `docs/workflow/worker-results/pixel-cafe-share-fulfillment-s252-*.md`
- `docs/workflow/qa-reports/pixel-cafe-share-fulfillment-s252-qa.md`

## Denied Paths

- Existing unrelated dirty files under Group, Settings, knowledge, and outputs
- payment-provider implementations, provider configuration, and production data
- generic `/group-buy` UI/API semantics except the minimum shared callback,
  expiry, and refund hooks explicitly allowed above
- gateway routing/failover, ordinary Account scheduling, billing calculation,
  authentication policy, dependencies, lockfiles, and generated frontend assets
- Docker, Compose, container replacement, deployment, shared/production DB,
  provider traffic, commit, push, branch cleanup, and memory/knowledge writes

## Constraints

- Canonical Cafe states are `open`, `awaiting_account`, `activating`, `active`,
  `completed`, `refunding`, `refunded`, `failed`, and `cancelled`.
- These are real persisted `group_buy_rounds.status` values. The migration
  expands only that CHECK constraint; generic group-buy continues using its
  existing `open/activating/active/failed/cancelled` subset. Cafe timeout moves
  the Round to `refunding`; it becomes `refunded` only when every paid purchase
  batch has a succeeded refund record. Provider-pending or manual-review rows
  keep the Round in `refunding`.
- `max_buyers <= total_shares`; `max_shares_per_user <= total_shares`; both are
  snapshotted into each Round. A new buyer consumes a buyer slot while it has a
  valid locked or paid batch; an existing member may top up at the buyer cap.
- Pixel Cafe orders accept `share_count`; one order remains one refundable
  purchase batch. Top-up is allowed only while the Round is `open`.
- Plus accepts only normalized OpenAI `plan_type=plus`; Pro accepts `pro` or
  `ChatGPTPro`. Missing/unknown tier, inactive Account, wrong platform/group, or
  another activating/active Round must fail before state mutation.
- Public member avatars are deterministic pseudonyms and never expose user IDs,
  email, Account name, Account ID, credentials, or per-member purchased shares.
- `My Rooms` may expose a safe non-email Account label plus platform and masked
  email only after activation. Reuse the existing email-mask boundary and close
  the known Account.Name-as-email disclosure while touching this projection.
- Existing `cafe_rooms.account_id`, seat fields, and legacy binding sources stay
  readable for migration compatibility but are not the source of truth for new
  Rounds. Generic group-buy remains unchanged.
- New `cafe_round_memberships` rows contain `round_id`, `user_id`, status,
  `paid_shares`, `reserved_shares`, optional `bound_api_key_id`, activation and
  expiry timestamps, with a unique `(round_id,user_id)` key. Each Cafe purchase
  batch gets nullable `membership_id`; legacy rows remain readable.
- `api_key_account_bindings` adds nullable `membership_id`; `seat_id` becomes
  nullable for compatibility, and a CHECK requires exactly one of seat or
  membership. Existing active-seat uniqueness remains conditional on non-null
  `seat_id`; new active-membership uniqueness is conditional on non-null
  `membership_id`. New Keys use managed source type `cafe_room_membership` and
  `managed_source_id=membership.id`. Runtime resolves membership bindings first
  and uses seat bindings only for legacy rows.
- New Cafe Rounds persist explicit immutable snapshot columns for
  `subscription_tier`, `max_buyers`, `max_shares_per_user`,
  `fulfillment_timeout_minutes`, `validity_days`, `target_group_id`, platform,
  and per-share quota/5H/1D/7D limits. Account assignment and Key creation read
  these snapshots, never the mutable Plan. `fulfillment_deadline_at` is set only
  when paid shares first reach total shares.
- New Rounds also persist `cafe_fulfillment_version=membership_share`; existing
  rows are backfilled as `legacy_seat`. Gateway binding lookup, activation,
  migration, and expiry select their path from this explicit version rather
  than guessing from nullable relations. Existing `cafe_room_seat` Keys and
  bindings are never rewritten.
- Administrator fulfillment API is fixed as:
  - `GET /api/v1/admin/cafe/rounds/pending?page=&page_size=&search=` returns
    paginated Round ID/status, Room ID/code/name, tier, paid/total shares,
    joined/max buyers, paid-full time and fulfillment deadline.
  - `GET /api/v1/admin/cafe/rounds/:id/account-options?page=&page_size=&search=`
    returns only minimal `id,name,platform,status,plan_type,email_masked` options
    already filtered by the Round snapshot and active-binding uniqueness.
  - `POST /api/v1/admin/cafe/rounds/:id/assign-account` accepts only
    `{ "account_id": number }` and returns the refreshed pending/active Round
    projection. It performs assignment and activation as one transaction.
  - Stable failures are `CAFE_ROUND_NOT_AWAITING_ACCOUNT` (409),
    `CAFE_ACCOUNT_TIER_MISMATCH` (400), `CAFE_ACCOUNT_ALREADY_IN_USE` (409),
    `CAFE_FULFILLMENT_DEADLINE_EXPIRED` (409), and the existing
    `CAFE_ACTIVATION_FAILED` (409). No credentials enter any response.
- Use additive, idempotent PostgreSQL migration SQL and regenerate Ent from the
  approved schema. Do not run a migration against the shared local database.
- Preserve the current scene renderer, local demo mode, custom Room name, and
  configurable page description while replacing fake plan labels and seat UI.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert,intercept,sql/execquery,sql/lock --idtype int64 ./ent/schema
go test ./migrations -run "TestPixelCafeShareFulfillment" -count=1
go test ./internal/service -run "TestCafe(Room|Public|Round|Membership|Fulfillment|Activation|Order|Expiry)" -count=1
go test ./internal/handler ./internal/handler/admin -run "TestCafe" -count=1
go test ./internal/server/routes -run "TestCafe" -count=1
go test ./internal/service ./internal/handler ./internal/handler/admin -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts src/views/admin/pixelCafe/components/__tests__/CafeRoomAccountPicker.spec.ts"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run build"
git diff --check
git ls-files -u
```

Browser QA uses a task-owned Playwright profile and checks desktop `1440x1000`
and mobile `390x844`, share purchase/modal states, no horizontal overflow, and
exact profile/daemon cleanup. It must not reuse or terminate user Chrome.

## Output

- Developer reports use the worker-result template and first-line verdict.
- Independent QA writes
  `docs/workflow/qa-reports/pixel-cafe-share-fulfillment-s252-qa.md` with a
  first-line PASS/FAIL/BLOCKED verdict.
- Reports list changed files, executed commands, test discovery, migration and
  privacy evidence, remaining risks, dirty-worktree preservation, and
  `knowledge_candidates` without writing long-term knowledge.

## Stop Rules

- Stop if implementation requires provider-specific upstream provisioning,
  shared database mutation, production configuration, or payment-provider code.
- Stop if Ent generation changes unrelated entities or cannot be separated from
  concurrent schema work.
- Stop if preserving the existing dirty Pixel Cafe scene/page behavior is not
  possible without resetting or absorbing unrelated changes.
- Stop if exact one-Key-per-membership activation cannot be transactional with
  the existing API-key repository boundary; return to Controller for redesign.
- Stop after two worker failures or any denied-path modification.

## Budget

- worker_mode: `codex-agent-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10` per bounded worker
- worktree_root: `E:/codex-worktrees`

## Controller Takeover Amendment

- The backend `gpt-5.6-terra` Developer failed twice after producing only the
  schema/migration and share-order skeleton. The low-cost backend worker loop is
  closed under the Stop Rules; its partial patch is not an integration
  candidate by itself.
- Controller may retain only reviewed in-scope parts and must complete three
  separately verifiable slices in the same isolated backend worktree:
  1. legacy-compatible Membership/share ordering and payment accounting;
  2. atomic pending-account assignment, activation, and strict Key binding;
  3. public/My Room projections plus timeout refund lifecycle.
- Every original Success Criterion, Allowed/Denied Path, acceptance command,
  no-commit/no-push/no-container/no-shared-DB boundary, and independent QA gate
  remains unchanged. This amendment narrows execution order; it does not reduce
  product or test scope.
