### PASS: upstream-v027-openai

## Scope And Diff Review

- Independently reviewed the five implementation/test files allowed by the contract:
  `openai_gateway_service.go`, `openai_gateway_chat_completions_raw.go`,
  `openai_chat_roles.go`, `openai_codex_models_service.go`, and
  `openai_compat_v027_test.go`.
- The adjacent DeepSeek work in `openai_gateway_service.go` was not changed or
  assessed as part of this task.
- The task implementation is limited to the approved OpenAI behavior port.
  `git diff HEAD --check` passed and the unmerged-path check was empty.

## Behavioral Review

- Strict Chat role conversion runs only after account selection and target URL
  resolution. It applies to API-key DeepSeek, Kimi, and Zhipu accounts, or to
  an API-key compatible account whose parsed hostname exactly matches one of
  the six strict provider hosts. URL userinfo, path text, model names, and
  lookalike hostnames cannot activate it. OAuth remains unchanged.
- `json.RawMessage` retains unknown message/root fields and numeric literals;
  only `developer` role values become `system`. The original input body remains
  available for a retry against a later account.
- Manifest validation retains case-sensitive `models` lookup and rejects a
  missing/null/wrong-type/malformed envelope. The outer JSON decode validates
  the complete raw array once, avoiding redundant array decoding while
  accepting an empty array.
- HTTP response affinity writes with `context.WithoutCancel` when a request
  context exists, then wraps it with the established Redis timeout. Values
  survive cancellation, the durable operation has a deadline, and nil context
  safely falls back to `context.Background()`.

## Commands And Evidence

```powershell
cd E:/codex-worktrees/sub2api/upstream-v027-port/backend
go test ./internal/service -run '^TestV027OpenAI' -count=1
go test ./internal/service -run '^(TestFetchCodexModelsManifest|TestOpenAIGatewayService_BindHTTPResponseAccount|TestForwardAsRawChatCompletions_(GrokOAuthUsesOfficialCLIUserAgent|GrokAPIKeyDoesNotUseOfficialCLIUserAgent|ForcesStreamUsageUpstreamAndPassesUsageDownstream|TransportErrorReturnsFailover))$' -count=1
go build ./...
```

All commands passed. The V027 suite includes a task-owned `httptest` raw Chat
route that verifies strict role forwarding, original retry-body preservation,
and exact large-number retention. It also covers host/API-key scope, OAuth
preservation, invalid strict payloads, manifest edge cases, cancellation,
value propagation, deadline bounds, and nil guards.

## Limits

Evidence is limited to local tests and compilation. No real provider, Redis,
database, container, deployment, commit, or push was used.
