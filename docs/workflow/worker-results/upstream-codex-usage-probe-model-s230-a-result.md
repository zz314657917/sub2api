### DONE: upstream-codex-usage-probe-model-s230-a

# Controller Result

## Task ID

upstream-codex-usage-probe-model-s230-a

## Status

done

## Summary

- Added `CodexUsageProbeModel = "codex-auto-review"` for OAuth Codex quota probes.
- `probeOpenAICodexSnapshot` now uses the dedicated model.
- The ordinary `DefaultTestModel` remains unchanged for normal OpenAI account tests.

## Changed Files

- `backend/internal/pkg/openai/constants.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/openai_codex_usage_probe_model_test.go`
- `docs/workflow/worker-results/upstream-codex-usage-probe-model-s230-a-result.md`

## Commands Run

```text
backend/go test ./internal/service -run "TestCodexUsageProbeModel|TestOpenAICodexVersionConsistency" -count=10 -> PASS (8.544s)
backend/go test ./internal/service -count=1 -> PASS (79.802s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (14.171s)
backend/gofmt -d internal/pkg/openai/constants.go internal/service/account_usage_service.go internal/service/openai_codex_usage_probe_model_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 16e4f7ecc upstream/main -> PASS
```

## Risks

- Validation uses local service tests only; real provider, Redis, database, container,
  deployment, and push operations are excluded by contract.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no
