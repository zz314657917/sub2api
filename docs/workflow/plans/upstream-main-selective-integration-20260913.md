# Upstream Main Selective Integration Plan (2026-09-13)

## Status

`PLANNER DRAFT / awaiting contract review`

This plan records a selective, behavior-level integration sequence. It does
not authorize a merge, rebase, cherry-pick, push, deployment, container
replacement, database write, or changes to the protected dirty worktree.

## Current Evidence

- Local `main` and `origin/main`: `2d0144ffc`.
- `upstream/main`: `bdb42e22f`.
- Merge-base: `18790386a76f12ae5721e557dc652c346ca699d5`.
- The histories are materially divergent (approximately 1554 local commits
  versus 3543 upstream commits and about 2891 changed files), so whole-history
  integration is out of scope.
- Untracked `.tmp-tutorial-live.html`, `pelican-bicycle.html`, `outputs/`, and
  `sub2api` remain protected.

## Delivery Order

### Sprint S302: OpenAI WS global connection cap (`956f4672e`)

Goal: make `mode_router_v2` account connection capacity obey the configured
global hard cap while preserving account-level limits and existing routing
semantics.

Planner checks before contract approval:

- Compare the upstream cap calculation with the local
  `effectiveMaxConnsByAccount` owner and its callers.
- Define precedence for global cap, account concurrency, zero/unlimited values,
  and pool/account modes.
- Add only the focused service/gateway owner and regression tests to the
  allowlist; preserve the local WebSocket aliases, fallback, and scheduler
  behavior.

Acceptance:

- Focused Go tests cover below-cap, at-cap, above-cap, zero/unlimited, and
  multi-account cases.
- `go test` for the owning packages, `go build ./...`, `gofmt`, exact allowlist,
  `git diff --check`, and unmerged-index checks pass.
- Real provider traffic and deployment remain unverified and out of scope.

Stop rules: stop on a required schema/config migration, a change to retry or
scheduler semantics, or any protected-path modification.

### Sprint S303: JWT transient lookup errors (`781a02aea`)

Goal: distinguish `ErrUserNotFound` from temporary/internal user lookup errors
so a database outage does not become a false `401` or clear a valid session.

Planner checks before contract approval:

- Identify the local middleware, token-refresh client path, and error contract.
- Preserve genuine missing-user invalidation and existing auth status codes.
- Decide whether frontend refresh behavior needs a separate owner contract;
  do not mix it with S302.

Acceptance:

- Backend tests cover found user, not found, transient database error, and
  unrelated internal error.
- Frontend tests cover refresh behavior when the server returns invalid-session
  versus transient failure, if the local client owns that behavior.
- Focused Go/Vitest tests, backend build, frontend typecheck/build when
  touched, formatting, exact allowlist, diff, and conflict checks pass.
- Live database and authenticated browser runtime remain separately recorded.

Stop rules: stop if the fix requires changing token format, session storage,
database schema, or global error handling outside the auth owners.

### Sprint U03: Low-risk compatibility queue

Evaluate independently after S302/S303:

- `213de0797`: `Accept-Encoding` header casing. Small transport-only fix.
- `5968fd0ed`: MiniMax monitoring provider allowlist. First prove the local
  MiniMax adapter and monitor registry are complete.
- `188e3a9f9`: reject malformed proxy-list responses. Requires parser and
  caller regression tests.
- `4e5d67df3`: Claude Code `max_tokens=1` probe compatibility.

Each item gets its own mini-contract unless the exact local owner and tests are
identical. No queue item may be bundled with payment, billing, model catalog,
schema, migration, or deployment changes.

## Deferred Candidates

Defer `9eb120dd4` (Firefox/Cloudflare fingerprint), `8ea4dc56f` (Gemini
`finishReason` attribution), `d8326fccf` (Claude mid-conversation output
config), `c0d511937` and `7ccc8a6f5` (Image 2.5/OAuth), and OpenCode commits
such as `242907854` and `981279c99`. Their local topology or runtime/provider
dependencies are too broad for this plan's small-batch contracts.

## Shared Boundaries

- Use a clean task-owned worktree for every implementation Sprint; do not
  modify the dirty primary worktree.
- Port behavior manually. Do not cherry-pick individual upstream commits when
  their surrounding topology differs, and do not merge `upstream/main`.
- Protect the existing dirty paths and all untracked artifacts listed above.
- Before implementation, create `docs/workflow/tasks/<task-id>.md` with
  `status: approved`, `review_verdict: PASS`, base commit, worker model,
  allowed/denied paths, acceptance commands, output, and stop rules.
- Developer and independent QA workers use `gpt-5.6-terra`; final decision is
  made by the Planner/Evaluator owner. If that model route is unavailable,
  record `BLOCKED` rather than silently substituting another model.
- No push, deployment, container update, shared database operation, or real
  provider claim is implied by a local code-level PASS.

## Next Legal Action

Review this plan, then draft and independently review the S302 contract. S303
must remain a separate contract even if S302 finishes cleanly.

