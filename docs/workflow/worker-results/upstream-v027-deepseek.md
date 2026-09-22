### PASS: upstream-v027-deepseek

## Changed Paths

- `backend/internal/pkg/apicompat/responses_tool_output_media.go`
- `backend/internal/pkg/apicompat/responses_tool_output_media_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/deepseek_media_v027_test.go`

## Implementation

Adapted upstream `881ab1b0c` and `acc05620c` to the local monolithic
`normalizeDeepSeekResponsesRequestBody`. Native DeepSeek Responses requests
still force `store=false` and remove `previous_response_id`; when a tool
output contains image media, the helper converts its output to text and emits
the media in a following user message. Developer and system notices encountered
inside a parallel output batch are held until all sibling outputs and the media
message have been emitted. The request decoder uses `UseNumber`, preserving
large integer JSON values while rebuilding only changed DeepSeek payloads.

## Commands And Evidence

Executed from `backend`:

```powershell
gofmt -w internal/pkg/apicompat/responses_tool_output_media.go internal/pkg/apicompat/responses_tool_output_media_test.go internal/service/openai_gateway_service.go internal/service/deepseek_media_v027_test.go
go test ./internal/pkg/apicompat -count=1
go test ./internal/service -run '^TestV027DeepSeek' -count=1
go build ./...
```

All commands passed. `TestV027DeepSeekForwardsLiftedMediaToLocalhost` creates a
task-local `httptest` server, sends the actual `buildUpstreamRequest` HTTP
request to `/responses`, and verifies the received request body contains two
contiguous `function_call_output` items, a following user media message with
both data-image URLs, then the developer/system notices. It also verifies
`store=false`, absent `previous_response_id`, and exact large integer
`9007199254740993` forwarding. The apicompat tests cover the same ordering and
the no-op plain-output path, user-turn boundary preservation, and idempotence:
running the helper again on an already lifted payload returns unchanged.

Root diff checks were run: `git diff HEAD --check` passed and
`git diff --name-only --diff-filter=U` was empty. The current unrelated dirty
`docs/workflow/status.md` was preserved and is outside this task's allowlist.

## Upstream Mapping And Limits

The upstream helper's response-tool media transformation and parallel-batch
ordering are adapted; no upstream files, commits, dependencies, migrations,
accounts, containers, or deployment state were imported or changed. The
localhost fake upstream proves request shaping only. Real DeepSeek provider,
database, container, deployment, and paid-account behavior remain unverified.
