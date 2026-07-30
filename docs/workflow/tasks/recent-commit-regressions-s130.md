# Task Contract: recent-commit-regressions-s130

## Task ID

`recent-commit-regressions-s130`

## Role

Generator, with Codex acting as Planner and final Evaluator.

## Goal

Repair the confirmed regressions found while reviewing recent local commits:

1. apply the existing five-attempt same-account capacity retry to every OpenAI gateway failover constructor;
2. prevent partially refunded group-buy orders from being closed as fully refunded;
3. turn a payment arriving after a timed-out round released its seat into the existing group-buy refund flow;
4. remove Grok `tool_choice` when `tools` is absent, `null`, or empty; and
5. defer the leaderboard account-age client guard until the public setting is known.

## Success Criteria

- Exact OpenAI capacity failures set `RetryableOnSameAccount` and `SameAccountRetryLimit` consistently in Responses, Chat Completions, Messages, Images, and applicable raw or stream paths, without changing ordinary pool retry behavior or model selection.
- Only `OrderStatusRefunded` closes a provider-backed group-buy refund. A partial provider refund is quarantined for manual review and never changes the seat to `refunded`.
- A paid callback for a `released` seat caused by a timed-out round changes the seat to `refund_pending`, records an event and refund reason, and returns successfully so the existing refund processor can act on it.
- Grok payload sanitization removes `tool_choice` unless `tools` is a non-empty JSON array.
- A first navigation with public leaderboard settings still unknown is allowed; an explicitly loaded numeric threshold, including `0`, retains current access enforcement.
- Focused Go and frontend regressions, repository compile/typecheck, formatting and diff integrity checks pass.

## Allowed Paths

- `backend/internal/service/openai_gateway_*.go`
- `backend/internal/service/openai_*test.go`
- `backend/internal/service/group_buy.go`
- `backend/internal/service/group_buy_test.go`
- `frontend/src/router/index.ts`
- `frontend/src/router/__tests__/guards.spec.ts`
- `docs/workflow/tasks/recent-commit-regressions-s130.md`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/qa-reports/recent-commit-regressions-s130-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `outputs/**`
- database schema, migration, Ent schema, or generated Ent files
- payment provider integrations and credentials
- deployment, Docker/container, environment, CI, or production configuration
- Git commit, push, branch/worktree cleanup, or remote changes
- unrelated gateway routing, pricing, billing, model catalog, and UI changes

## Constraints

- Preserve all existing user changes, especially untracked `outputs/` and the S129 local-integration documentation.
- Keep the patch minimal and reuse existing state-machine and failover helpers.
- Do not call a real payment provider or upstream API.
- No schema migration or new dependency is permitted.

## Acceptance Commands

- `cd backend; go test ./internal/service -run 'Test(OpenAI|GroupBuy|Grok)' -count=1`
- `cd backend; go test ./... -run '^$'`
- `cd frontend; npm.cmd run test -- --run src/router/__tests__/guards.spec.ts`
- `cd frontend; npm.cmd run typecheck`
- `gofmt -d backend/internal/service/openai_gateway_*.go backend/internal/service/group_buy*.go`
- `git diff --check`; `git diff --name-only`; `git ls-files -u`

## Output

- Minimal source and regression-test diff.
- `docs/workflow/qa-reports/recent-commit-regressions-s130-qa.md` with an evidence-based PASS, FAIL, or BLOCKED verdict.
- Updated workflow status, main log, and current-task handoff.

## Stop Rules

- Stop for user direction if a repair needs a schema migration, payment-provider behavior change, external API call, deployment, or any path outside the allowlist.
- Stop and revise the contract if the existing state model cannot represent the late-payment refund without changing persisted schema.
