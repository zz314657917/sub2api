### PASS: upstream-openai-fast-flex-user-scope-s71

## Findings

- No blocking implementation issue or behavioral regression was found at integration head `f1997dcf9841a7ac80ac59f1d3f15be46dc965ac`.
- The evaluator reads only the trusted positive `ctxkey.APIKeyUserID`. Existing API-key authentication writes that value to the request context; client-controlled body and `x-sub2api-user-id` values are neither used for matching nor forwarded as identity headers.
- User-scoped rules are evaluated as a group before global rules. Both groups retain configuration order and first-match behavior, and user, account scope, service tier, and model matching continue to intersect.
- Invalid non-empty UI state does not widen a rule to global scope: the added `0` remains in the submitted `user_ids` payload and backend validation rejects zero, negative, and duplicate IDs. `user_ids` is omitted only after the operator explicitly removes every ID.
- The previously reported i18n namespace finding is closed. The five keys `userIds`, `userIdsHint`, `userIdPlaceholder`, `addUserId`, and `removeUserId` exist under both English and Chinese `openaiFastPolicy`, are absent from both `betaPolicy` objects, and the regression test imports the real locale modules directly.
- The `64d2b0b7c..f1997dcf9` audit contains exactly the 14 contract-allowed paths, with no denied path, missing expected path, conflict marker, or whitespace error.

## Executed Checks

- Ran the complete S71 Acceptance Commands unchanged. Exact discovery gates found service `4/4`, middleware `1/1`, DTO `1/1`, and frontend `1/1` tests.
- Required service matrix, validation/round-trip, parsed WebSocket relay, passthrough WebSocket relay, managed/API-key passthrough/OAuth passthrough HTTP forwarding, and DTO tests all passed.
- The HTTP fake upstream captured matching and non-matching authenticated users in all three transport/account modes. The real WS proxy relay captured both first and model-less follow-up `response.create` frames in parsed and passthrough modes; spoofed body/header identity did not affect policy behavior.
- `go test ./internal/service -run "Test.*OpenAIFastPolicy|Test.*PassthroughBilling" -count=1` passed.
- The targeted `SettingsView.spec.ts` test passed `1/1` after exact Vitest discovery. Its assertions load and clone IDs, preserve an invalid non-empty submission, submit edited IDs exactly, require explicit clear for global scope, and directly import the en/zh locale objects.
- `corepack.cmd pnpm --dir frontend run typecheck` passed.
- The QA worktree initially had no local frontend dependencies. The Acceptance Commands created a temporary `frontend/node_modules` junction to `F:/mcplugins/sub2api/frontend/node_modules`; `finally` removed it and `Test-Path frontend/node_modules` returned `False` afterward.
- The contract clean-worktree and dynamic path/diff gates passed. Because `codex/upstream-v0151-followups-s71-s73` already pointed at `f1997dcf9`, an additional hard-coded `64d2b0b7c..HEAD` audit was run: `changed=14`, `allowed=14`, `unexpected=0`, `missing=0`, `conflict_markers=0`, and `git diff --check` passed.
- Manual diff review confirmed DTO/service/frontend round-trip, trusted context extraction, user-before-global precedence, validation, and the 14-file allowed-path boundary.

## Unverified Risks

- No request was sent to a live OpenAI Priority/Flex endpoint. HTTP and WebSocket behavior is covered by deterministic local fake upstreams and the real in-process proxy relay.
- No unrelated full backend/frontend suite or race test was required or run; validation follows the approved S71 focused contract and adjacent Fast/Flex/passthrough regressions.
- Frontend output retained the pre-existing non-failing unresolved `router-link` warning and stale `caniuse-lite` notice.

## Recommendation

- PASS. S71 is suitable for continued integration with S72 and S73. This verdict does not authorize merging to `main`, pushing, deployment, or container replacement.
