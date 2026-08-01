### PASS: openai-overload-retry-s135

# QA Report

## Task ID

`openai-overload-retry-s135`

## Verdict

`PASS / published`

## Contract Checked

- `docs/workflow/tasks/openai-overload-retry-s135.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run <S135 overload/capacity/normal-passthrough-stream regex> -count=1 -> PASS (fresh publication worktree; 5.504s)
go test ./internal/handler -run <S135 retry-limit/delay/loop regex> -count=1 -> PASS (fresh publication worktree; 18.649s)
go test ./internal/handler -run <same-account and overload switch regression regex> -count=1 -> PASS (pre-publication feature worktree; 8.580s)
go test ./... -run '^$' -> PASS (fresh publication worktree; all backend packages compile; 53.3s)
go test -tags=unit ./internal/service -run <fallback constructor regression regex> -count=1 -> BLOCKED by pre-existing unit-test source drift (`stringPtr` redeclaration, stale billing signatures, removed Grok runtime-block helpers)
gofmt -d <all changed Go files> -> PASS (empty output)
git diff --check -> PASS
conflict-marker audit -> PASS (no matches)
unmerged-index audit -> PASS
allowed-path audit -> PASS
constructor policy audit -> PASS (all in-scope OpenAI capacity-aware failover constructor sites)
git fetch origin main -> PASS
git rev-list --left-right --count HEAD...origin/main -> 0 0 before this documentation receipt
git rev-parse HEAD origin/main -> 3ef7f36de for both refs before this documentation receipt
```

- manual/source checks:

```text
server_is_overloaded, slow_down, and the exact overload sentence -> explicit limit=3, base=1s
overload handler attempts -> delay calculation is 1s, 2s, 3s; fourth failure unschedules and switches
Selected model is at capacity -> explicit limit=5 and no overload backoff, preserving the fixed 500ms handler default
normal HTTP, passthrough HTTP, standard pre-output SSE, and passthrough pre-output SSE -> policy metadata is carried before client output
generic overloaded text, extended overload messages, ordinary transient errors, and generic passthrough 429/529 -> no overload-only policy
Responses, Messages, Chat Completions, and Images -> all in-scope failover constructors use the same per-error policy; the existing handler loops use the delay helper with the account retry-limit fallback
```

## Findings

- The initial amendment overreached into Embeddings and Videos. Their handlers
  only re-entered the scheduler, and Embeddings has no session hash, so that did
  not guarantee a retry on the same account.
- The Embeddings/Videos extension was withdrawn. Their current selection loops
  do not pin retries to the original account, so claiming same-account behavior
  there would be misleading.
- Overload matching is restricted to the two structured codes or the complete
  sentence (with an optional final period); generic `overloaded` text and
  extended messages stay on the existing failover behavior.
- Feature commit `84915599b` was pushed to its scoped remote branch. The clean
  publication worktree integrated it on `origin/main@1c1021133` as
  `3ef7f36de`, then fast-forwarded `origin/main` with fetched ref parity.
- Publication did not add schema, configuration, persistence, deployment, or
  container changes.

## Unverified Risks

- No live OpenAI upstream was held in an overload state, so real provider timing,
  rate-limit headers, and account cooldown side effects remain unverified.
- The loop regression uses a microsecond backoff to keep CI fast; production
  values are asserted through the policy and delay-calculation tests rather
  than a six-second wall-clock test.
- Unit-tag fallback tests remain unexecuted because unrelated pre-existing test
  source drift prevents the package from compiling; the production sources are
  covered by the default package tests and full compile probe.
- No deployed service or refreshed container was exercised after publication.

## Recommendation

`PASS / published`: the narrowed S135 source change is published on
`origin/main` as `3ef7f36de`. Live provider behavior, deployment, and container
smoke remain unverified.

## Knowledge Promotion

- `none`
