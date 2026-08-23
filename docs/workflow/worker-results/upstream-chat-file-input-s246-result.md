### DONE: upstream-chat-file-input-s246

## Changed files

- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `docs/workflow/worker-results/upstream-chat-file-input-s246-result.md`

## Commands run

- Read approved S246 workflow state, spec addendum, contract, local DTO/converter owners, and upstream source/merge behavior.
- `go test ./internal/pkg/apicompat -run '^TestChatCompletionsToResponses_(FilePartFileData|FilePartFileID|EmptyFilePartSkipped)$' -count=10`
- `go test ./internal/pkg/apicompat -count=1`
- `go test ./internal/service -run '^$' -count=1`
- `go test ./cmd/server -run '^$' -count=1`
- `gofmt -l` for all three owners, plus contract diff/ancestry/conflict/index/S239/protected-primary gates.

## Key output

- Focused file-part regressions: PASS, `0.663s` at `-count=10`.
- Complete `apicompat`: PASS, `0.042s`.
- Service compile: PASS, `0.070s`; server compile: PASS, `5.428s`.
- `file_data` maps with filename; `file_id` maps; empty file payload skips; converter retains both fields when supplied.
- Both upstream commits are ancestors of `upstream/main`; S239 `ChatFunctionCall.Name` remains `json:"name,omitempty"`; formatting, exact scope, diff, conflict/index and protected-primary gates passed.

## Risks

- No known implementation risk in the approved mapping scope. This deliberately does not add upload/download, validation, MIME/size policy, gateway behavior, or external provider execution. Independent Terra QA remains required.

## Contract compliance

- Kept local DTO/converter topology and all pre-existing compatibility fields; added only file DTO fields and converter mapping.
- Business commit contains exactly the three approved product/test paths; evidence commit contains only this report.
- No primary-worktree, dependency, external-state, browser, or remote operation occurred.

## knowledge_candidates

- none; this is bounded compatibility behavior pending independent QA.
