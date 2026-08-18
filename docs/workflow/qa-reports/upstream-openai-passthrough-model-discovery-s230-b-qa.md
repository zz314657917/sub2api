### PASS: upstream-openai-passthrough-model-discovery-s230-b

# QA Report

## Task ID

upstream-openai-passthrough-model-discovery-s230-b

## Verdict

PASS

## Contract Checked

- `docs/workflow/tasks/upstream-openai-passthrough-model-discovery-s230-b.md`

## Evidence

- diff reviewed: yes
- allowed paths checked: yes; only the gateway owner, hotpath test file, and Controller result report are present
- denied paths touched: no
- independent QA worktree: `E:/codex-worktrees/sub2api/upstream-openai-passthrough-model-discovery-s230-b-qa`
- commands run:

```text
backend/go test ./internal/service -run "TestGetAvailableModels_OpenAIPassthroughUsesDefaultFallback|TestGetAvailableModels_GlobalListPreservesMappedModelsWithOpenAIPassthrough|TestGetAvailableModels_ErrorAndGlobalListBranches" -count=10 -> PASS (5.976s)
backend/go test ./internal/service -count=1 -> PASS (77.359s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (10.775s)
backend/gofmt -d internal/service/gateway_service.go internal/service/gateway_hotpath_optimization_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 1ea4150bf upstream/main -> PASS
```

## Findings

未发现明确问题。

## Behavioral Coverage

- OpenAI passthrough accounts ignore stale model mappings and use the default fallback.
- Ordinary OpenAI mapped accounts retain their configured whitelist.
- Global discovery preserves mapped models from non-OpenAI accounts.

## Risks

- Validation uses local service tests only; real provider, Redis, database, container,
  deployment, and push operations are excluded by contract.

## Bug Owner Recommendation

original-worker

## Root Cause

none

## Retest Scope

不适用；PASS。

## Knowledge Promotion

none
