### DONE: upstream-cn-provider-billing-s229-b

# Controller Result

## Task ID

upstream-cn-provider-billing-s229-b

## Status

done

## Summary

- CN provider accounts now remove unpriced client `claude-*` billing candidates while retaining every non-Claude candidate in order.
- Group and Channel pricing remain explicit operator overrides and keep their matching Claude candidates eligible.
- Empty/all-blank candidates now wrap `ErrModelPricingUnavailable`, allowing the existing zero-cost warning and usage-log path to persist the request.
- A Kimi request with only `claude-sonnet-4` now writes a zero-cost usage log in simple mode; no upstream request occurs in this flow.

## Changed Files

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_cn_billing_test.go`
- `docs/workflow/worker-results/upstream-cn-provider-billing-s229-b-result.md`

## Commands Run

```text
backend/go test ./internal/service -run "TestFilterCNProviderBillingModelCandidates|TestCalculateOpenAIRecordUsageCost_EmptyCandidatesIsPricingUnavailable|TestOpenAIGatewayServiceRecordUsage_CNFilteredCandidatesWriteZeroCostLog" -count=10 -> PASS (5.866s)
backend/go test ./internal/service -count=1 -> PASS (65.490s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (5.537s)
backend/gofmt -d ... -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U and backend/git ls-files -u -> PASS (empty)
backend/git merge-base --is-ancestor 10c8b7020 upstream/main -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.866s
ok github.com/Wei-Shaw/sub2api/internal/service 65.490s
ok github.com/Wei-Shaw/sub2api/cmd/server 5.537s [no tests to run]
```

## Risks

- The behavior is validated with local stubs only. The contract excludes real provider, database, container, and deployment traffic.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
