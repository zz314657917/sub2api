### PASS: openai-model-capacity-retry-five-s127

# QA Report

## Task ID

`openai-model-capacity-retry-five-s127`

## Verdict

`PASS / source-only`

## Contract Checked

- `docs/workflow/tasks/openai-model-capacity-retry-five-s127.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run <S127 normal/passthrough/stream/non-capacity regex> -count=1 -> PASS (5.458s)
go test ./internal/handler -run 'Test(SameAccountRetryLimit|HandleFailoverError_(SameAccountRetry|BasicSwitch|IntegrationScenario))' -count=1 -> PASS (17.491s)
go test ./... -run '^$' -> PASS (all backend packages compile; 69.7s)
gofmt -d <ten changed Go files> -> PASS (empty output)
git diff --check -> PASS
conflict-marker audit -> PASS (no matches)
handler retry-limit audit -> PASS (four OpenAI handler reads use sameAccountRetryLimit)
capacity constructor audit -> PASS (the exact classifier is the sole source of the explicit limit)
allowed-path audit -> PASS
```

- manual/source checks:

```text
normal HTTP capacity -> retryable=true, explicit limit=5
passthrough HTTP capacity -> retryable=true, explicit limit=5
standard and passthrough pre-output response.failed capacity -> retryable=true, explicit limit=5
generic transient 400, stream overload, and passthrough 429/529 -> explicit limit=0 and existing behavior retained
generic handler loop -> five same-account retries; sixth failure unschedules and switches account
Responses, Messages, Chat Completions, and Images -> prefer explicit limit, otherwise preserve account.GetPoolModeRetryCount()
```

## Findings

- 未发现 S127 范围内的明确问题。
- `SameAccountRetryLimit <= 0` preserves the established handler fallback, so
  pool-mode retry configuration and other retryable upstream errors remain
  unchanged.
- A real capacity response from an external OpenAI upstream, deployment,
  container refresh, commit, and push were not performed.

## Unverified Risks

- The 500 ms delay is covered by the existing handler path and the loop test,
  but no live upstream was held at capacity for the full retry sequence.

## Recommendation

`PASS`: retain the isolated S127 change for local integration when the user
authorizes a commit or merge. Do not broaden the five-retry setting to generic
pool retries.

## Knowledge Promotion

- `none`
