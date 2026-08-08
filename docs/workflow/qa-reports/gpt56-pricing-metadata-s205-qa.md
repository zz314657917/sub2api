### PASS: gpt56-pricing-metadata-s205

# GPT-5.6 Pricing Metadata Integration S205 QA

## Findings

- No release-blocking defect remains in the S205 scope.
- The upstream JSON patch applied on the frozen latest-main base with the same stable patch-id as
  `0616a297459e48f1fb3503b01c58cde670060c30`, and the commit retains the source footer for upstream
  `60f6dc91cf907841c09b4aa7f9f78874fd08579c`.
- Dynamic parsing now preserves all five S205 metadata fields. The real bundled fixture asserts exact Sol, Terra,
  and Luna values directly from `parsePricingData`, so static fallback policy cannot mask a parser regression.
- The implementation does not modify `billing_service.go`, handlers, routes, frontend, schema, migrations, or
  production configuration. Existing standard/priority/flex behavior is unchanged.

## Executed Checks

```text
go test ./internal/service -run 'TestParsePricingData|TestDefaultPricingIncludesGpt56|Test.*GPT56|Test.*GPT5' -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/service 0.110s

go test ./internal/service -run 'TestDefaultPricingGPT56FeedsBillingTierAndLongContextMatrix' -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/service 0.073s

go test ./internal/handler -run 'TestOpenAIResponsesWebSocket|TestOpenAIWS' -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/handler 1.098s

go test ./internal/service -count=1
-> PASS: ok github.com/Wei-Shaw/sub2api/internal/service 61.175s

gofmt -d backend/internal/service/pricing_service.go backend/internal/service/pricing_service_test.go
-> PASS: empty output

git diff --check 920de6d135d8873ec2d3322c4f6434deef2a8c1d..HEAD
-> PASS

stable patch-id comparison: 0616a2974 vs 78ad91f56
-> PASS: d8c2433637f22f1892040e4e19feb2546c26d26f

frozen-base ancestry, provenance footer, conflict-marker scan, unmerged-index check, exact changed-path allowlist,
JSON semantic values, and clean-worktree gate
-> PASS: QA_GATES_PASS changed=8
```

Diff precision review:

- Product diff is limited to one upstream JSON metadata commit plus parser fields, parser assignments, and one
  real-fixture regression.
- Workflow changes are limited to the S205 contract, status, log, worker result, and this QA report.
- No unrelated refactor, formatting churn, dependency, generated output, or denied path was included.

Official-source check:

- Current OpenAI pricing lists the selected Sol/Terra/Luna standard, cached-input, cache-write, output,
  long-context, Batch, Flex, and Fast values used by the fixture.
- The OpenAI changelog states Terra became 20% cheaper and Luna 80% cheaper on 2026-07-30, and identifies prompts
  exceeding 272K tokens as long-context for GPT-5.6 on 2026-08-05.

## Unverified Risks

- The targeted `-tags=unit` service billing suite does not compile on either S205 or the frozen
  `main@920de6d13` baseline because existing unit-only tests have stale function signatures, duplicate helpers,
  and removed fields. This is not attributed to S205, but no unit-tag PASS is claimed; the accepted evidence is
  the default-build parser-to-billing matrix plus the complete default service package.
- The repository still has no Batch billing selector. S205 retains Batch cache-write metadata but does not consume
  it at runtime, by contract.
- Flex billing still uses the existing generic `0.5x` multiplier. This matches the current GPT-5.6 official Flex
  values but remains an algorithmic rule rather than per-model field selection.
- No real provider, pricing refresh endpoint, container, deployment, staging, or production traffic was exercised.
- The configured `deepseek-v4-pro` QA worker was unavailable in the active collaboration environment. An
  independent Terra contract review was completed; final command execution and verdict were performed by the
  primary evaluator.

## Recommendation

- `PASS / local-regression`. The S205 branch is safe to integrate into local `main` as a normal fast-forward.
- Do not describe this as deployed or production-verified. Do not push or deploy without separate authorization.
