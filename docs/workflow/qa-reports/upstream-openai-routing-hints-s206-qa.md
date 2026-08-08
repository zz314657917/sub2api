### PASS: upstream-openai-routing-hints-s206

# Upstream OpenAI OAuth Routing Hints S206 QA

## Findings

- No merge-blocking defect remains in the S206 scope.
- The first complete service run exposed one obsolete image-test assertion for the removed OAuth beta header. The contract was amended only for that upstream-equivalent assertion, and both the focused and complete reruns passed.
- Diff review confirms the routing hint is derived after local body policy/override processing, all case variants are removed before gateway synthesis, and API-key paths cannot retain a caller/account supplied hint.
- WebSocket affinity remains soft: an idle mismatched connection is replaced when capacity permits, while a busy full pool can fall back without making the hint a continuation key. Direct dial generation and prewarm-target checks prevent stale state from re-entering the local pool.
- The `nanoid` change is patch-identical to upstream `8ad0a5ff5`; no other frontend dependency entry changed.

## Executed Checks

```text
go test ./internal/service -run 'TestOpenAICodexRoutingHint|TestOpenAI.*RoutingHint|TestOpenAI.*LegacyBeta|TestOpenAIWS.*Routing|TestOpenAIWS.*Affinity|TestOpenAIWS.*Generation|TestOpenAIWS.*Prewarm|TestOpenAIGatewayServiceForwardImages_OAuthPassesNAndReturnsAllImages' -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/service 1.368s

go test ./internal/service -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/service 61.833s

go test ./cmd/server -run '^$' -count=0
-> PASS: ok github.com/Wei-Shaw/sub2api/cmd/server 0.056s [no tests to run]

gofmt -w <seven changed Go files>
-> PASS; subsequent committed-tree diff check is clean

stable patch-id comparison: upstream 8ad0a5ff5 vs local e6120ec69
-> PASS: 39eab1acf608c09d5492b0615eec3d8250427184

frozen-base ancestry, exact 13-path implementation allowlist, conflict-marker scan,
unmerged-index check, nanoid lockfile occurrences and clean implementation worktree
-> PASS: QA_GATES_PASS changed=13 head=dda605f62b7f
```

Diff precision review:

- Product changes are limited to the local HTTP/WS request builders, routing-hint helper/tests, pool affinity/generation behavior, one upstream-equivalent image assertion, and the exact dependency lockfile patch.
- Workflow changes are limited to the S206 contract, spec, status, log, worker result, QA, and handoff evidence.
- No unrelated refactor, schema/migration, configuration, provider credential, deployment, container, or production path is included.

## Unverified Risks

- `go test -race` for the focused pool tests could not run because this Windows Go environment has `CGO_ENABLED=0` and no `gcc`/`clang` executable. Default-build concurrency regressions passed, but race-detector PASS is not claimed.
- No real OpenAI OAuth HTTP/WebSocket endpoint, proxy, provider credential, network retry, or multi-process traffic was exercised.
- No container, deployment, staging, production, database, Redis, or remote Git publication was performed.
- The security result is limited to applying the audited lockfile patch with upstream-equivalent integrity metadata; no fresh registry audit or frontend install was required by this contract.

## Recommendation

- `PASS / local-regression`. The S206 commit chain is safe to fast-forward into local `main`.
- Do not describe this as provider-verified, deployed, production-verified, or published. Push and deployment require separate authorization.
