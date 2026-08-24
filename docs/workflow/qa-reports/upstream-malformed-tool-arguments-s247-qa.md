### PASS: upstream-malformed-tool-arguments-s247

# Independent QA Report

## Findings

- Business `a8ce875c2` changes exactly the five amended local owners: three
  `apicompat` bridge/test owners, fallback service code, and the default-tag
  `openai_gateway_responses_chat_fallback_s247_test.go`. Evidence `aab18c3cf`
  changes only the Controller result report.
- The original unit-tag fallback test and intermediate upstream
  `openai_gateway_cc_pipeline.go` have no business-commit diff. The final
  merge first-parent has the same four product owners plus the unit-tag test;
  the default-tag S247 test is the contract-approved one-for-one local
  substitution.
- All source commits `e2d9ce0ca`, `fbc9ee626`, and final merge `fd6cd474d`
  are ancestors of `upstream/main`; `git log fd6cd474d..upstream/main` showed
  no later touch to the five upstream product owners.
- Static review retained custom-tool, tool-search, namespace, empty-arguments,
  and stream lifecycle compatibility code. Complete package regression passed;
  no S242/S243 regression was found.

## Commands Run

From `backend/` in the isolated QA worktree:

- `go test ./internal/pkg/apicompat -list '^(TestResponsesInputToChatMessages_SkipsInvalidHistoricalFunctionCall|TestResponsesInputToChatMessages_SkipsInvalidEmptyCallIDOutput|TestChatCompletionsResponseToResponses_SkipsInvalidFunctionArguments|TestStream_InvalidToolArgumentsAreRejectedBeforeFinalize|TestStream_ValidToolCallAtOutputLimitKeepsIncompleteResponse)$'`
- Same five-test `apicompat` selector with `-count=10`.
- `go test ./internal/service -list '^TestStreamChatCompletionsAsResponses_RejectsInvalidToolArgumentsAtOutputLimit$'`
- Same default-tag service selector with `-count=10`.
- `go test ./internal/pkg/apicompat -count=1`
- `go test ./internal/service -count=1`
- `go test ./cmd/server -run '^$' -count=1`
- Contract `gofmt -l` command for all five local owners.

From the QA and read-only primary worktrees:

- Candidate business/evidence scope, denied-owner deltas, final merge scope,
  source/merge ancestry, no-later-upstream-owner-touch, diff/index/conflict,
  S242/S243 marker, and protected-primary checks.

## Key Output

- apicompat discovery listed all five required tests; focused x10 passed:
  `ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 2.529s`.
- Default-tag service discovery listed
  `TestStreamChatCompletionsAsResponses_RejectsInvalidToolArgumentsAtOutputLimit`;
  focused x10 passed in `0.079s`.
- Complete apicompat passed in `0.042s`; complete service passed in `64.657s`
  with captured `FULL_SERVICE_EXIT=0`; server compile passed with exit `0` and
  gofmt output was empty.
- Candidate `git diff --check` passed; staged and unmerged indexes are empty;
  no conflict marker exists in any amended owner.
- The protected primary 22-path patch ID is
  `941b1edf9df9e465a6100007edfc4a6715e38b5e`. All three product and two
  `outputs/` untracked SHA-256 values match the amended contract; primary
  staged/unmerged indexes remain empty.

## Risks

- The unrelated legacy `-tags=unit` service suite still has known stale compile
  errors. It is intentionally outside amended S247 acceptance; QA did not run
  it or modify build tags/its old test owner.
- Provider, database, container, browser, deployment, push, upload/download,
  and production/shared-state operations were not exercised because they are
  forbidden by contract.

## Contract Compliance

- QA wrote only this report in
  `E:/codex-worktrees/sub2api/upstream-malformed-tool-arguments-s247-qa`.
- No business, dependency, primary-worktree, other workflow/knowledge,
  provider, database, container, browser, remote, or push operation occurred.

## knowledge_candidates

- none
