### DONE: upstream-main-gateway-compat-s4

## Summary
- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-gateway-compat-s4` and branch `codex/upstream-main-gateway-compat-s4` from baseline `34d02457b`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-gateway-compat-s4.md`.
- Ported all five approved gateway/apicompat fixes without directly merging `upstream/main`.
- Added one local compatibility commit to keep the Images error passthrough port buildable on the local baseline without importing broader upstream gateway helper refactors.

## Commits
- `1561fbd08` docs: add gateway compat upstream sprint contract
- `b66cdc3cf` cherry-pick of `381d1d6d6`: OpenAI Images real upstream error passthrough.
- `5807f54c0` cherry-pick of `2e212d18e`: Chat Completions compat failed responses.
- `1508e00ad` cherry-pick of `5bd3d9043`: preserve upstream `response.failed` errors.
- `bef2b360a` equivalent/minimal port of `9b99f6c1f`: DeepSeek reasoning-only visibility.
- `58d7e4f21` local adaptation: Images upstream error helper methods for current baseline.
- `98d8da6e8` cherry-pick of `60867022b`: Responses-to-Anthropic tool pairing.

## Notes
- `5bd3d9043` had a small conflict in `backend/internal/handler/openai_images.go`; resolved by preserving local account-slot release behavior and adding upstream writer-size tracking for already-communicated upstream errors.
- `9b99f6c1f` depended on a broader upstream stream lifecycle state machine not present in this baseline. The port keeps the local simplified bridge and adds only the reasoning-only fallback behavior plus current-state-machine tests. The upstream `chatcompletions_responses_stream_lifecycle_test.go` file was not restored.
- No candidate commit was deferred.

## Changed Files
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_cc_chain_test.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_tool_pairing_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_test.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-gateway-compat-s4.md`
- `docs/workflow/worker-results/upstream-main-gateway-compat-s4-result.md`
- `docs/workflow/qa-reports/upstream-main-gateway-compat-s4-qa.md`

## Verification
- `git status --short --branch`
- `git diff --check`
- `git diff --name-status 34d02457b..HEAD`
- `go test ./internal/pkg/apicompat -run "Responses|ChatCompletions|Anthropic|Tool|DeepSeek|Reasoning" -count=1`
- `go test ./internal/service -run "OpenAIImages|ChatCompletions|Responses|Failed|Tool|DeepSeek" -count=1`
- `go test ./internal/handler -run "OpenAI|Gateway|Images|Failed" -count=1`
- `go test ./internal/pkg/apicompat ./internal/service ./internal/handler -count=1`

## Integration Verification
- Clean integration worktree `E:/codex-worktrees/sub2api/upstream-main-gateway-compat-s4-integration` was created from current `main@d1c10a7b`.
- `codex/upstream-main-gateway-compat-s4` merged without conflicts into integration commit `1a05d7343`.
- Integration reran the same path audit and Go target/regression tests; all passed.
