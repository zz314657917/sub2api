### PASS: upstream-gpt56-priority-cache-billing-s69

# QA Report

## Scope

- Approved contract: `docs/workflow/tasks/upstream-gpt56-priority-cache-billing-s69.md`.
- QA head: `53eb79bd0`; implementation audit range: `2b3b5514e..07399e50d`.
- Worker result begins with the required `### DONE: upstream-gpt56-priority-cache-billing-s69` verdict.
- QA mode: runtime. Worker self-reported results were read for scope only and were not used as PASS evidence.
- Agent Matrix deviation: the specified `deepseek-v4-pro` worker returned model 404. With the user's explicit multi-agent authorization, this independent QA used the currently available collaboration agent as a fallback; the approved contract, fresh QA worktree, and final Evaluator ownership were unchanged.

## Findings

- No implementation bug, behavioral regression, or contract-boundary violation was found.
- Standard cache creation uses the Standard cache-write price. Priority selects the dedicated Priority cache-write price before long-context multiplication. Flex continues using the Standard price and applies the existing `0.5` service-tier multiplier afterward.
- The session input-side total remains `InputTokens + CacheCreationTokens + CacheReadTokens`. Exactly `272000` tokens do not trigger long-context pricing; totals above `272000` apply the existing `2.0` input-side multiplier, including cache creation.
- Flat channel and interval `CacheWritePrice` overrides set both Standard and Priority cache-write prices and mark the value explicit. A configured zero remains zero in both tiers and is not replaced by the GPT-5.6 derived price.
- Missing non-explicit GPT-5.6 Standard and Priority cache-write prices derive independently from their corresponding input prices times `1.25`.
- Cache breakdown pricing remains isolated from the new flat Priority field: when 5m/1h breakdown is enabled, `computeCacheCreationCost` continues reading only the 5m/1h prices.
- Embedded and fallback pricing contain the contract values for Sol, Terra, and Luna; dynamic LiteLLM parsing preserves the dedicated field, including a numeric zero.
- The real `RecordUsage` path forwards `Result.ServiceTier=priority` into billing and persists `ServiceTier`, the dedicated `CacheCreationCost`, matching `TotalCost`, and rate-adjusted `ActualCost`.
- Path audit found exactly the eight Allowed Paths and no Denied Path changes. Diff precision review found no unrelated refactor or behavior expansion.

## Executed Checks

- Exact required-test discovery - PASS: exactly six required test names were listed.
  `go test ./internal/service -list $requiredPattern`
- Exact six-test execution - PASS (`internal/service`, six tests).
  `go test ./internal/service -run $requiredPattern -count=1`
- Pricing and RecordUsage regressions - PASS (`internal/service`, `1.579s`).
  `go test ./internal/service -run "^(TestParsePricingData_.*|TestGetModelPricing_.*|TestOpenAIGatewayServiceRecordUsage_GPT56SeparatesCacheWriteForBillingAndStats|TestOpenAIGatewayServiceRecordUsage_ServiceTier.*)$" -count=1`
- WebSocket cache-creation and usage regressions - PASS (`internal/service/openai_ws_v2`, `5.797s`).
  `go test ./internal/service/openai_ws_v2 -run "CacheCreation|Usage" -count=1`
- Default service package compile - PASS (`internal/service`, no tests to run).
  `go test ./internal/service -run "^$" -count=1`
- Allowed-path audit - PASS: `git diff --name-only 2b3b5514e..07399e50d` returned exactly the eight contract paths.
- Implementation whitespace/error audit - PASS: `git diff --check 2b3b5514e..07399e50d` returned no findings.
- Clean-worktree preflight before report creation - PASS: `git status --porcelain --untracked-files=all` returned no paths.
- Manual semantic review - PASS: Standard/Priority/Flex selection, explicit-zero overrides, the exclusive `272000` boundary, 5m/1h preservation, and RecordUsage totals were checked directly against implementation and focused assertions.

## Unverified Risks

- No live OpenAI billing request was issued. Pricing selection and persistence are verified through deterministic service, resolver, WebSocket regression, and RecordUsage tests.
- Unit-tag suites were intentionally excluded because the approved contract records unrelated existing `-tags=unit` compile drift and explicitly denies using those suites as S69 evidence.
- The full service suite, race testing, deployment, push, and merge into `main` were outside this QA contract.
- The Agent Matrix model deviation remains a process-level exception, not a code-level gap; final acceptance still belongs to the main Codex Evaluator.

## Recommendation

- PASS. S69 satisfies the approved contract and can proceed to Final Evaluator closeout. This report does not authorize merging into `main`, pushing, deployment, or container replacement.

## Bug Owner Recommendation

- `none`

## Root Cause

- `none`

## Retest Scope

- None required.

## Knowledge Promotion

- `none`
