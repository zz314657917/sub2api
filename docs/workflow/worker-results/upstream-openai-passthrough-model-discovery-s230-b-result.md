### DONE: upstream-openai-passthrough-model-discovery-s230-b

# Controller Result

## Task ID

upstream-openai-passthrough-model-discovery-s230-b

## Status

done

## Summary

- OpenAI passthrough accounts now return the default model-list fallback even when
  stale `model_mapping` data is present.
- Ordinary OpenAI mapped accounts retain their mapped whitelist.
- Global model discovery still preserves mapped models from non-OpenAI accounts.

## Changed Files

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_hotpath_optimization_test.go`
- `docs/workflow/worker-results/upstream-openai-passthrough-model-discovery-s230-b-result.md`

## Commands Run

```text
backend/go test ./internal/service -run "TestGetAvailableModels_OpenAIPassthroughUsesDefaultFallback|TestGetAvailableModels_GlobalListPreservesMappedModelsWithOpenAIPassthrough|TestGetAvailableModels_ErrorAndGlobalListBranches" -count=10 -> PASS (11.850s)
backend/go test ./internal/service -count=1 -> PASS (78.589s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (11.230s)
backend/gofmt -d internal/service/gateway_service.go internal/service/gateway_hotpath_optimization_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 1ea4150bf upstream/main -> PASS
```

## Risks

- Validation uses local service tests only; real provider, Redis, database, container,
  deployment, and push operations are excluded by contract.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
