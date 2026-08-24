### PASS: upstream-chat-file-input-s246

# Independent QA Report

## Findings

- Business candidate `6f22dbae4` changes exactly the three allowed
  `apicompat` owners; Developer evidence `b340ebb9d` changes only
  `docs/workflow/worker-results/upstream-chat-file-input-s246-result.md`.
- Chat `type:"file"` maps to Responses `type:"input_file"` only when
  `file_data` or `file_id` is present. The conversion preserves filename,
  file_data, file_id, and text-part order; an empty file payload is skipped.
- `ChatFunctionCall.Name` remains `json:"name,omitempty"` at
  `types.go:610`. No local S239 streamed empty-name regression was found.
- Source `4d4a0be1a` and merge `6244090c1` are both ancestors of
  `upstream/main`; candidate diff/index/conflict gates pass.

## Commands Run

From `backend/` in the isolated QA worktree:

- `go test ./internal/pkg/apicompat -run '^TestChatCompletionsToResponses_(FilePartFileData|FilePartFileID|EmptyFilePartSkipped)$' -count=10`
- `go test ./internal/pkg/apicompat -count=1`
- `go test ./internal/service -run '^$' -count=1`
- `go test ./cmd/server -run '^$' -count=1`
- `gofmt -l internal/pkg/apicompat/types.go internal/pkg/apicompat/chatcompletions_to_responses.go internal/pkg/apicompat/chatcompletions_responses_test.go`

From the QA and read-only primary worktrees:

- Candidate scope/diff, S239 tag, source/merge ancestry, `git diff --check`,
  staged/unmerged-index, and conflict-marker checks.
- Exact 22-path protected-primary patch-ID and five protected untracked-file
  SHA-256 checks, including both `outputs/` files.

## Key Output

- Focused three-test selector x10: `ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.795s`.
- Complete `apicompat`: `ok github.com/Wei-Shaw/sub2api/internal/pkg/apicompat 0.040s`.
- Service compile: `ok github.com/Wei-Shaw/sub2api/internal/service 0.069s [no tests to run]`, exit `0`.
- Server compile: `ok github.com/Wei-Shaw/sub2api/cmd/server 1.618s [no tests to run]`, exit `0`; gofmt output was empty.
- Candidate diff check passed; staged and unmerged indexes were empty and no
  conflict marker was found.
- The 22 protected tracked user paths have stable patch ID
  `941b1edf9df9e465a6100007edfc4a6715e38b5e`. All three product untracked
  files and both outputs files match the contract's SHA-256 values; primary
  staged/unmerged indexes remain empty.

## Risks

- This bounded adaptation does not exercise file uploads/downloads, provider
  traffic, MIME or size policy, browser automation, database/container work,
  deployment, or push; all are outside the contract.
- The first combined command crossed the terminal's 30-second output boundary
  during service compilation. QA waited for that exact Go process to exit, then
  reran the same service/server compile commands to record explicit exit-0
  evidence.

## Contract Compliance

- QA used only `E:/codex-worktrees/sub2api/upstream-chat-file-input-s246-qa`
  for writes and modifies only this permitted QA report.
- No business file, dependency, other workflow/knowledge path, primary
  worktree, provider, database, container, browser, remote, or push operation
  was modified or invoked.

## knowledge_candidates

- none
