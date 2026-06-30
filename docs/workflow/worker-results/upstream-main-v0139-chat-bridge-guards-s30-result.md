### DONE: upstream-main-v0139-chat-bridge-guards-s30

# Worker Result

## Task ID
upstream-main-v0139-chat-bridge-guards-s30

## Status
done

## Summary
- Ported `/v1/chat/completions` `codex_cli_only` enforcement from upstream `ae5e980dd`.
- Ported OAuth chat bridge default-instructions suppression from upstream `dbdbfb112`.
- Added regressions for non-official client rejection before upstream forwarding and normal chat-completions OAuth bridge keeping `instructions` empty.
- Kept Responses-shaped `/v1/chat/completions` behavior unchanged by only skipping default instructions for normal Chat Completions conversion.

## Changed Files
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `docs/workflow/tasks/upstream-main-v0139-chat-bridge-guards-s30.md`

## Commands Run
```text
gofmt -w internal/service/openai_gateway_chat_completions.go internal/service/openai_gateway_chat_completions_test.go -> pass
go test ./internal/service -run "TestForwardAsChatCompletions_EnforcesCodexCLIOnlyRestriction|TestForwardAsChatCompletions_OAuthDoesNotInjectDefaultInstructions|TestForwardAsChatCompletions_APIKeyPropagatesPromptCacheKeyInResponsesBody|TestForwardAsChatCompletions_TransportErrorReturnsFailover" -count=1 -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/service	5.779s
```

## Risks
- Live OpenAI OAuth traffic was not exercised; verification is unit-level around local forwarding decisions.
- The new restriction gate runs before APIKey raw chat-completions fallback as upstream intended.
- Existing unrelated dirty proxy/account/frontend/knowledge files remain outside this sprint.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
