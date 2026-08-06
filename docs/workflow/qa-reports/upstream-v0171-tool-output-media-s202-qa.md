### PASS: upstream-v0171-tool-output-media-s202

## Findings

- The local Responses-to-Chat bridge previously serialized tool output directly, leaving embedded images in `tool` content and lacking a safe ordering boundary for later media messages.
- The scoped port recognizes direct data URLs and structured image parts, rewrites only recognized media to a marker, and emits `image_url` content after the complete matched reply batch.
- The normalizer reconstructs answered tool replies in assistant call order. This prevents a media user message, an interleaved message, an orphan reply, or an unanswered parallel call from violating Chat tool-call adjacency.
- No explicit issue was found in the final source and scope review. Changed paths remain within the S202 contract; the primary worktree's Pixel Cafe changes were not modified.

## Executed Checks

- `gofmt -w backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go backend/internal/pkg/apicompat/chatcompletions_responses_tool_output_media_test.go`: passed.
- `go test ./internal/pkg/apicompat -count=1`: passed.
- `go test ./internal/service -run 'Test.*(Responses|ChatCompletions).*' -count=1`: passed.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- Manual diff review confirmed image extraction, `json.Number` preservation, duplicate last-wins behavior, and reply-batch ordering against upstream `2bf9c6d56` and the S202 contract.
- `gofmt -d`, `git diff --check`, unmerged-index check, conflict-marker scan, scoped status audit and primary-worktree status check: passed.

## Unverified Risks

- No real Chat Completions upstream, image download, model vision request, browser session, deployment or production traffic was exercised.
- Only the recognized Responses image shapes are moved. Unknown tool-output schemas intentionally retain the existing stringification behavior.

## Recommendation

Commit this isolated source-level port to `codex/upstream-v0171-integration-s183`; do not push, deploy or modify the primary worktree.
