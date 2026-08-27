### PASS: upstream-oauth-image-verbatim-s268

## Scope and provenance

- QA worktree: `E:/codex-worktrees/sub2api/upstream-oauth-image-verbatim-s268`.
- Tested implementation: `1853a3bec`.
- Frozen base: `cd42eebf1`; upstream source: `329b92ef045fd24b49b33e719e42facc7b26e1b2`.
- Relative implementation scope is exactly `backend/internal/service/openai_images.go`, `openai_images_responses.go`, `openai_images_test.go`, plus the contract and worker result. No denied path or unmerged index entry was found.

## Commands

- `go test ./internal/service -run '^TestBuildOpenAIImagesResponsesRequest_(StripsInputFidelity|RequiresVerbatimUserPrompt)$' -count=10` -> PASS.
- `go test ./internal/service -run '^$' -count=1` -> PASS (compile-only).
- `go test ./cmd/server -run '^$' -count=1` -> PASS (compile-only).
- `gofmt -d internal/service/openai_images.go internal/service/openai_images_responses.go internal/service/openai_images_test.go` -> no output.
- `git diff --check -- <contract paths>` -> PASS.
- `git ls-files -u` -> empty.

## Behavioral review

The focused regression confirms the OAuth image Responses request strips unsupported input fidelity and emits the stable verbatim-prompt instruction while retaining the original prompt text. The implementation is confined to the existing image request builders and tests; edit inputs, model/tool parameters, streaming, and non-OAuth paths are not broadened. Provider/runtime behavior was not exercised per contract.

## Residual risk

No live provider call or deployment smoke was performed. This leaves external provider acceptance/runtime behavior unverified, while local request construction and package compilation pass.
