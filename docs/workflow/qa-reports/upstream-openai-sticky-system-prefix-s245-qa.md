### PASS: upstream-openai-sticky-system-prefix-s245

# Independent QA Report

## Findings

- Business candidate `b45f9ac389e7208eace784c280cd040016eba5cb` changes exactly
  `backend/internal/service/openai_content_session_seed.go` and
  `backend/internal/service/openai_content_session_seed_test.go`; Developer
  evidence `4d373dac62b201f06c78248d88912cf1df184955` changes only its approved
  worker-result report.
- The local direct-`gjson` scanner now closes `systemPrefixOpen` at the first
  non-`system`/`developer` role. This preserves the leading contiguous prefix,
  continues capturing the first user message, and ignores a later dynamic
  system/developer message without importing upstream's unrelated single-scan
  refactor.
- Upstream source `e45490a36` and merge `2ddda6735` are both ancestors of
  `upstream/main`; each first-parent product scope is the same two S245 owner
  paths. Candidate range `git diff --check` passed with no conflict markers,
  staged entries, or unmerged index entries.

## Commands Run

From `backend/` in the isolated QA worktree:

- `go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_ChatCompletions_(IgnoresLaterSystemMessages|UsesLeadingSystemDeveloperPrefix)$' -count=10`
- `go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_' -count=1`
- `go test ./internal/service -count=1`
- `go test ./cmd/server -run '^$' -count=1`
- `gofmt -l internal/service/openai_content_session_seed.go internal/service/openai_content_session_seed_test.go`

From the QA worktree and read-only primary worktree:

- Candidate business/evidence exact-scope review, source/merge ancestry and
  first-parent scope, frozen-range `git diff --check`, staged/unmerged index,
  and conflict-marker checks.
- Exact eleven-path primary Pixel Cafe patch-ID/path check plus primary staged,
  unmerged-index, and `outputs/` preservation checks.

## Key Output

- Focused leading-prefix regressions: `ok github.com/Wei-Shaw/sub2api/internal/service 0.070s`, exit `0` at `-count=10`.
- Complete seed suite: `ok github.com/Wei-Shaw/sub2api/internal/service 0.068s`, exit `0`.
- Complete service suite: `ok github.com/Wei-Shaw/sub2api/internal/service 64.597s`, exit `0`.
- Server compile: `ok github.com/Wei-Shaw/sub2api/cmd/server 5.680s [no tests to run]`, exit `0`; `gofmt -l` output was empty.
- Primary scoped user patch ID remains
  `370ac77de0e2f530ab652b99fb3eb35e809f4c84` across exactly eleven contract
  paths. Primary staged and unmerged indexes are empty, and `outputs/` retains
  exactly two untracked files.

## Risks

- Browser automation, provider traffic, shared/production data, dependencies,
  deployment, containers, and push were not exercised because they are denied
  by the S245 contract.
- The complete service command takes about 65 seconds. QA captured its real
  exit code after the command outlived the terminal's 30-second output window;
  it passed and was not treated as a timeout or skipped test.

## Contract Compliance

- QA ran only in `E:/codex-worktrees/sub2api/upstream-openai-sticky-system-prefix-s245-qa`
  and changes only this allowed QA report.
- No business-file edit, primary-worktree write, dependency update, provider
  call, browser automation, remote write, container/deployment operation, or
  push occurred.

## knowledge_candidates

- none
