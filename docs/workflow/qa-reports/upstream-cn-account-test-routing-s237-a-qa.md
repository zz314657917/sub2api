### PASS: upstream-cn-account-test-routing-s237-a

# QA Report

## Task ID

`upstream-cn-account-test-routing-s237-a`

## Verdict

`PASS`

## Changed Files

The candidate under QA is commit `53e80223c379c1c7d4e8059fa4b76dca7a600d6a`.
Its product/test diff contains exactly the contract allowlist:

- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_cn_protocol.go`
- `backend/internal/service/account_test_service_cn_protocol_test.go`

The QA worktree had no working-tree or index changes relative to the candidate
commit. This report is the only new file written by QA.

## Executed Checks

```text
go test ./internal/service -list 'TestAccountTestService_(CN|DeepSeek)' -> PASS
  four focused CN/DeepSeek tests listed
go test ./internal/service -run 'TestAccountTestService_(CN|DeepSeek)' -count=10 -> PASS (10.694s)
go test ./internal/service -count=1 -> PASS (65.574s)
go test ./cmd/server -run '^$' -count=1 -> PASS (5.609s, no tests to run)
gofmt -l internal/service/account_test_service.go internal/service/account_test_service_cn_protocol.go internal/service/account_test_service_cn_protocol_test.go -> PASS (no output)
git diff --check -> PASS
conflict-marker scan over the three allowed files -> PASS (no markers)
git diff --name-only HEAD -> PASS (empty)
git diff --cached --name-only -> PASS (empty index)
exact candidate allowlist audit -> PASS (3 changed, 0 unexpected)
fake upstream audit -> PASS (httptest context plus in-memory DoWithTLS recorder; no provider traffic)
```

## Contract Compliance

- CN routing is selected before generic platform branches and covers Chat
  Completions, native Anthropic, and explicit DeepSeek Responses.
- Focused tests assert provider URL selection, mapped model/auth behavior,
  native Anthropic URL without the beta query, OpenAI-shaped Anthropic URL
  rejection before outbound, and stale Responses capability metadata bypass.
- The selected upstream source commits `ac6208de1`, `a749673de`, `01a008394`,
  `f75c4161f`, `85051616f`, and `b3092145d` are all ancestors of
  `upstream/main@67380eafd`; the candidate is a local adaptation with parent
  `4e59289ec` and does not merge divergent upstream history.
- No denied product paths, Adaptive behavior, migrations, dependencies,
  frontend changes, deployment, database, or remote operations were used.
- The main worktree dirty/untracked snapshot observed at QA start was still
  unchanged at completion; those user-owned paths were not edited.

## Risks and Unverified Items

- Tests use in-memory fake upstreams and do not validate real provider behavior,
  shared databases, containers, deployment, or production traffic, as required
  by the contract.
- Full service regression and server compile passed, but no browser or live
  provider acceptance was attempted.
- The shared main worktree remains dirty by design; this report makes no
  judgment on unrelated paths.

## knowledge_candidates

None. The verified result is task-scoped QA evidence and does not establish a
new durable repository rule.

## Recommendation

`PASS`: accept candidate `53e80223c` for Controller review and local
integration, subject to the separate protected-main and final review gates.
