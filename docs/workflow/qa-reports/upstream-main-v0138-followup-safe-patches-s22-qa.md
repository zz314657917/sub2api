### PASS: upstream-main-v0138-followup-safe-patches-s22

## Findings

- No contract violations found.
- Changed paths stayed inside S22 allowed backend/workflow paths.
- Denied-path audit including untracked files returned `NO_DENIED_PATHS`.

## Executed Checks

- PASS: `go test ./internal/pkg/apicompat -run "TestStream_ToolCallArgumentsInFirstChunkNotDoubled" -count=1`
- PASS: `go test -tags=unit ./internal/service -run "TestHandleStreamingResponsePassthroughDeduplicatesFunctionCallArguments|TestForwardResponsesChatCompletionsFallbackKeepsFunctionArgumentsSingle|TestForwardAsChatCompletions_TransportErrorReturnsFailover|TestForwardAsRawChatCompletions_TransportErrorReturnsFailover|TestIsNonRetryableRefreshError|TestEnsureEmailAuthIdentityCreateErrorReturnsFalse" -count=1`
- PASS: `git diff --check`
- PASS: denied-path audit over `git status --short` paths returned `NO_DENIED_PATHS`

## Unverified Risks

- No real OpenAI/GLM upstream traffic was run locally.
- Real OAuth token refresh failure text for `refresh_token_invalidated` was not replayed from live upstream; covered by string-classification regression.

## Recommendation

- PASS S22.
- Keep `9491de0a3`, `cc7612bdb`, and `ae5e980dd` as next safe backend candidates; keep payment/subscription/balance/product-surface commits out of this patch lane.
