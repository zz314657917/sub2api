### DONE: upstream-main-v0138-followup-safe-patches-s22

## Summary

Implemented S22 backend-only upstream follow-up patches without merging `upstream/main`.

## Ported

- `29122e305` / `fd0da2570` style fix: Chat Completions -> Responses stream conversion now clears copied first-chunk tool arguments before shared accumulation, preventing `{"cmd":"ls"}{"cmd":"ls"}`.
- `2b49d662c` style fix: Responses passthrough now dedupes exactly repeated JSON `arguments` in function-call done/item/completed payloads.
- `0a97a5f46`: `refresh_token_invalidated` is treated as non-retryable.
- `65fa72892`: both converted and raw `/v1/chat/completions` transport errors now return OpenAI transport failover errors rather than writing terminal 502 responses.
- `82576e0a3`: email auth identity creation errors now assign to the outer `err`, so create failures are logged and not shadowed before reload.

## Equivalent

- `063454ae9` / `4567f6582`: usage cache creation/read stats already existed locally.
- `011278204` / S21 Spark image tool strip: already present locally.

## Skipped

- `9491de0a3` images content refusal passthrough, `cc7612bdb` overloaded codes, and `ae5e980dd` `/v1/chat/completions codex_cli_only` were left as next-batch candidates to keep S22 narrow.
- Payment/subscription/balance/order currency, Antigravity project fallback, GPT-5.5 instructions fallback, ops chart UI, CI/README/sponsor/VERSION, Ent/migrations/wire, frontend/public product surfaces remained out of scope.

## Changed Files

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_passthrough_function_args_test.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/auth_service.go`
- `backend/internal/service/auth_service_identity_shadow_test.go`

## Commands Run

- `go test ./internal/pkg/apicompat -run "TestStream_ToolCallArgumentsInFirstChunkNotDoubled" -count=1`
- `go test -tags=unit ./internal/service -run "TestHandleStreamingResponsePassthroughDeduplicatesFunctionCallArguments|TestForwardResponsesChatCompletionsFallbackKeepsFunctionArgumentsSingle|TestForwardAsChatCompletions_TransportErrorReturnsFailover|TestForwardAsRawChatCompletions_TransportErrorReturnsFailover|TestIsNonRetryableRefreshError|TestEnsureEmailAuthIdentityCreateErrorReturnsFalse" -count=1`
- `git diff --check`
- denied-path audit over `git status --short` paths returned `NO_DENIED_PATHS`

## Risks

- Real upstream GLM/OpenAI passthrough traffic was not exercised; evidence is targeted unit/runtime tests.
- Skipped image refusal / overloaded-code / chat-completions `codex_cli_only` candidates should be evaluated in a later Sprint if still desired.
