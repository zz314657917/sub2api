### DONE: upstream-grok-official-ua-s255

# Worker Result

## Task ID

upstream-grok-official-ua-s255

## Status

done

## Summary

- Added the documented package-local `grokOfficialOAuthUserAgent` constant with
  the upstream-pinned `xai-grok-workspace/0.2.120` value.
- Reused it for Grok OAuth Responses egress, Raw Chat Completions OAuth
  fallback, and the Grok account-connection request. Raw Chat Completions
  keeps the caller User-Agent for Grok API-key accounts.
- Kept test ownership aligned with the contract: Responses request coverage is
  in `openai_gateway_grok_test.go`, Raw Chat coverage and its helper are in
  `openai_gateway_chat_completions_raw_test.go`, and the connection test stays
  in `account_test_service_grok_s125_test.go`.
- Business commit: `21c854c91 fix(grok): use official OAuth user agent`.

## QA Remediation

- Independent QA report `b0c275701` found two stale unit-tag expectations in
  the allowed Grok gateway test owner. Remediation commit
  `a5326be21 test(grok): update OAuth user agent expectations` replaces only
  those expected values with `grokOfficialOAuthUserAgent`.
- Static callchain review confirms the streaming `ForwardAsChatCompletions`
  test enters `forwardAsRawChatCompletions`, whose Grok OAuth branch sets the
  shared constant; `ForwardAsAnthropic` detects `PlatformGrok` and calls
  `buildGrokResponsesRequest`, whose Grok OAuth branch sets the same constant.
- No production behavior, task contract, workflow log, primary worktree, or
  out-of-scope file was changed by this remediation.

## Changed Files

- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_grok_s125_test.go`
- `backend/internal/service/openai_ws_http_bridge_test.go`

## Commands Run

```text
gofmt -w <all seven contract service paths> -> PASS
go test ./internal/service -run "Test(BuildGrokResponsesRequestUsesOfficialCLIUserAgent|ForwardGrok.*|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService_GrokOAuthUsesOfficialCLIUserAgent|ForwardAsRawChatCompletions_Grok(OAuthUsesOfficialCLIUserAgent|APIKeyDoesNotUseOfficialCLIUserAgent))" -count=10 -> PASS (5.615s)
go test ./internal/service -run "Test(BuildGrokResponsesRequest|ForwardGrok|ProxyResponsesWebSocketFromClientForGrokUsesXAIHTTPBridge|AccountTestService.*Grok|ForwardAsRawChatCompletions.*Grok)" -count=1 -> PASS (1.748s)
go test ./internal/service -count=1 -> PASS (65.721s)
go test ./cmd/server -run '^$' -count=1 -> PASS (0.063s)
git diff --check -> PASS
rg conflict-marker check -> PASS (no markers)
git diff --cached --name-only -> PASS (empty before commit)
git ls-files -u -> PASS (empty)
QA remediation default-tag focused x10 -> PASS (0.096s)
QA remediation default-tag scoped focused -> PASS (0.069s)
QA remediation go test ./internal/service -count=1 -> PASS (65.870s)
QA remediation go test ./cmd/server -run '^$' -count=1 -> PASS (0.086s)
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.615s
ok github.com/Wei-Shaw/sub2api/internal/service 65.721s
ok github.com/Wei-Shaw/sub2api/cmd/server 0.063s [no tests to run]
ok github.com/Wei-Shaw/sub2api/internal/service 65.870s
ok github.com/Wei-Shaw/sub2api/cmd/server 0.086s [no tests to run]
```

## Risks

- Supplemental `go test -tags=unit ./internal/service -run ...` cannot build
  because pre-existing unit-tag compilation failures outside this task include
  `stringPtr` redeclaration, stale billing method signatures, and missing
  proxy fields. The required default-tag focused and complete service gates
  passed. The moved Responses/Raw Chat tests remain in their mandated unit-tag
  owners; independent QA should retain the same distinction.
- No real provider, database, container, deployment, push, or mainline action
  was performed.
- Per QA remediation instruction, `-tags=unit` was not retried. The same
  pre-existing, out-of-scope unit baseline failures remain the only missing
  direct execution evidence for the two corrected unit-tag assertions.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Upstream Mapping

- Behavior-level port of `upstream 9fb260439` from
  `upstream/main@e2d9b823f`; local service ownership replaces the obsolete
  `sub2api-grok/1.0` identity without changing the local xAI package.
