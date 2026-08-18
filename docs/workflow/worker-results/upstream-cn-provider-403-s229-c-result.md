### DONE: upstream-cn-provider-403-s229-c

# Controller Result

## Task ID

upstream-cn-provider-403-s229-c

## Status

done

## Summary

- CN Kimi/Zhipu/DeepSeek accounts now reuse the existing OpenAI 403 handler.
- HTML 403 responses remain request/path failures and do not increment counters or mutate account schedulability.
- Structured CN 403 responses use the existing first-hit temporary cooldown and threshold permanent-disable behavior.
- No OpenAI, Anthropic, Antigravity, billing, stream, or disconnect behavior was changed outside the dispatch branch.

## Changed Files

- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_cn_403_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-403-s229-c-result.md`

## Commands Run

```text
backend/go test ./internal/service -run "TestHandleUpstreamError_CNProviderHTML403SkipsAccountPenalty|TestHandleUpstreamError_CNProviderStructured403TempUnschedulable|TestHandleUpstreamError_CNProviderStructured403ThresholdDisables" -count=10 -> PASS (0.501s)
backend/go test ./internal/service -count=1 -> PASS (65.690s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (5.517s)
backend/gofmt -d internal/service/ratelimit_service.go internal/service/ratelimit_service_cn_403_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U and backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 10c8b7020 upstream/main -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 0.501s
ok github.com/Wei-Shaw/sub2api/internal/service 65.690s
ok github.com/Wei-Shaw/sub2api/cmd/server 5.517s [no tests to run]
```

## Risks

- Validation uses local repository/counter stubs only; real provider, Redis, database, container, deployment, and push operations are excluded by contract.
- An initial root-level gofmt path probe used the wrong worktree directory; it failed before execution, was corrected to run from `backend/`, and the final format gate passed.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
