### PASS: upstream-openai-account-model-transient-s203

# Scope

- Behavior-level local adaptation of upstream `40b8f04a6` and `7d38e6712`.
- Adds bounded, process-local OpenAI API-key transient state per account+model pair. No persistent scheduler
  state, configuration, dependency, migration, frontend, container, deployment, provider credential, or remote
  Git change is included.

# Findings

- No contract-scope regression found. The state is case-normalized and bounded to 4096 pairs; a stale entry or
  clock rollback clears state safely. A success clears only the matching account+model pair.
- Selection now rejects a blocked pair for ordinary scheduling, fixed-account routing, and
  `previous_response_id` sticky routing. Retryable `500/502/503/504/520-524` and eligible transient `400`
  failures are recorded only for OpenAI API-key accounts; `429` and `529` are not recorded.
- Responses WebSocket terminal handling preserves a failure streak for `response.failed`; only
  `response.completed` and `response.done` report scheduling success that clears state.

# Executed Checks

```text
go test ./internal/service -run '^TestOpenAIModelTransient' -count=1 -v
PASS (6 tests)

go test ./internal/service -count=1
PASS (61.186s)

go test ./internal/handler -run 'TestOpenAI|Test.*Failover' -count=1
PASS (19.141s)

go test ./cmd/server -run '^$' -count=0
PASS (0.064s; compile probe)

gofmt -w <S203 changed Go files>
PASS

git diff --check
PASS

git ls-files -u
PASS (empty)
```

# Unverified Risks

- No real OpenAI account, upstream HTTP/WebSocket provider, persistent scheduler cache, deployment, or production
  traffic was exercised. The evidence is local regression and compile coverage only.
- The ignored `outputs/` directory remains untracked and excluded from this Sprint.

# Recommendation

Local integration is ready for the bounded S203 commit. Do not infer provider or production behavior from these
local regressions.
