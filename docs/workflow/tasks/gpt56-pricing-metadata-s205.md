# GPT-5.6 Pricing Metadata Integration S205

## Task ID
gpt56-pricing-metadata-s205

## Role
You are the Generator for this P/G/E Sprint. Implement only this approved contract and do not make architecture decisions outside its scope.

## Goal
Apply the already-reviewed upstream pricing metadata commit to the latest local `main`, preserving its upstream provenance. Restore the missing dynamic parser coverage for the GPT-5.6 long-context metadata and retain the two newly added cache-write tier metadata fields without changing the existing billing-tier algorithm.

## Success Criteria
- The latest-main branch contains the exact JSON metadata from `0616a297459e48f1fb3503b01c58cde670060c30`, including its `(cherry picked from commit 60f6dc91...)` provenance.
- `LiteLLMRawEntry` and `parsePricingData` preserve `long_context_input_token_threshold`, `long_context_input_cost_multiplier`, and `long_context_output_cost_multiplier` into `LiteLLMModelPricing`.
- `LiteLLMRawEntry`, `LiteLLMModelPricing`, and `parsePricingData` preserve `cache_creation_input_token_cost_batches` and `cache_creation_input_token_cost_flex` as metadata.
- Real-fixture regressions prove Sol, Terra, and Luna retain standard, priority, cache, long-context, batch-cache-write, and flex-cache-write values after dynamic parsing.
- Standard, priority, flex, cache-write, and long-context billing regressions remain green. No Batch billing path or tier-selection behavior is introduced.

## Context
- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/gpt56-pricing-metadata-s205`
- Base: `main@920de6d135d8873ec2d3322c4f6434deef2a8c1d`
- Upstream candidate: local provenance commit `0616a297459e48f1fb3503b01c58cde670060c30`, sourced from upstream merge `60f6dc91cf907841c09b4aa7f9f78874fd08579c`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`, `backend/internal/service/pricing_service.go`, `backend/internal/service/pricing_service_test.go`
- Official reference: `https://developers.openai.com/api/docs/pricing` and the 2026-07-30 / 2026-08-05 entries in `https://developers.openai.com/api/docs/changelog`

## Allowed Paths
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/pricing_service_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/gpt56-pricing-metadata-s205.md`
- `docs/workflow/worker-results/gpt56-pricing-metadata-s205-result.md`
- `docs/workflow/qa-reports/gpt56-pricing-metadata-s205-qa.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/service/billing_service.go`
- `backend/internal/handler/**`
- production configuration, container, deployment, credential, and provider-integration paths
- any path not listed under Allowed Paths

## Constraints
- Keep the change minimal; do not refactor pricing or service-tier architecture.
- Do not change the existing standard, priority, or flex calculation behavior.
- Do not add a Batch route, Batch request classification, or Batch billing selection path.
- Do not alter the already-correct Sol/Terra/Luna core rates or fallback rates.
- Do not push, deploy, update containers, call a real provider, or touch the main worktree's `outputs/` directory.
- Preserve existing user changes and do not use destructive Git commands.

## Acceptance Commands
```powershell
Set-Location "E:/codex-worktrees/sub2api/gpt56-pricing-metadata-s205/backend"
$env:GOTMPDIR = "F:/mcplugins/sub2api/.tmp/go-build"
go test ./internal/service -run 'TestParsePricingData|TestDefaultPricingIncludesGpt56|Test.*GPT56|Test.*GPT5' -count=1
go test ./internal/handler -run 'TestOpenAIResponsesWebSocket|TestOpenAIWS' -count=1

Set-Location "E:/codex-worktrees/sub2api/gpt56-pricing-metadata-s205"
$base = "920de6d135d8873ec2d3322c4f6434deef2a8c1d"

git merge-base --is-ancestor $base HEAD
if ($LASTEXITCODE -ne 0) { throw "S205 HEAD is not based on the frozen main commit" }

$messages = (git log --format=%B "$base..HEAD") -join "`n"
$sourceRecord = "(cherry picked from commit 60f6dc91cf907841c09b4aa7f9f78874fd08579c)"
if (-not $messages.Contains($sourceRecord)) { throw "upstream provenance record is missing" }

$formatDiff = gofmt -d backend/internal/service/pricing_service.go backend/internal/service/pricing_service_test.go
if ($LASTEXITCODE -ne 0 -or $formatDiff) {
    $formatDiff
    throw "gofmt produced a diff"
}

git diff --check "$base..HEAD"
if ($LASTEXITCODE -ne 0) { throw "committed diff check failed" }

$markers = rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' .
if ($LASTEXITCODE -eq 0) {
    $markers
    throw "conflict markers found"
}
if ($LASTEXITCODE -ne 1) { throw "conflict-marker scan failed" }

$unmerged = git ls-files -u
if ($unmerged) {
    $unmerged
    throw "unmerged index entries found"
}

$allowed = @(
    "backend/resources/model-pricing/model_prices_and_context_window.json",
    "backend/internal/service/pricing_service.go",
    "backend/internal/service/pricing_service_test.go",
    "docs/workflow/spec.md",
    "docs/workflow/status.md",
    "docs/workflow/main-log.md",
    "docs/workflow/tasks/gpt56-pricing-metadata-s205.md",
    "docs/workflow/worker-results/gpt56-pricing-metadata-s205-result.md",
    "docs/workflow/qa-reports/gpt56-pricing-metadata-s205-qa.md"
)
$changed = @(git diff --name-only "$base..HEAD")
$denied = @($changed | Where-Object { $_ -notin $allowed })
if ($denied.Count -gt 0) {
    $denied
    throw "S205 changed paths outside the contract allowlist"
}

$pricing = Get-Content -LiteralPath "backend/resources/model-pricing/model_prices_and_context_window.json" -Raw -Encoding UTF8 | ConvertFrom-Json
$expected = @{
    "gpt-5.6-sol" = @{ cache_creation_input_token_cost_batches = 3.125e-6; cache_creation_input_token_cost_flex = 3.125e-6; long_context_input_token_threshold = 272000; long_context_input_cost_multiplier = 2.0; long_context_output_cost_multiplier = 1.5 }
    "gpt-5.6-terra" = @{ cache_creation_input_token_cost_batches = 1.25e-6; cache_creation_input_token_cost_flex = 1.25e-6; long_context_input_token_threshold = 272000; long_context_input_cost_multiplier = 2.0; long_context_output_cost_multiplier = 1.5 }
    "gpt-5.6-luna" = @{ cache_creation_input_token_cost_batches = 1.25e-7; cache_creation_input_token_cost_flex = 1.25e-7; long_context_input_token_threshold = 272000; long_context_input_cost_multiplier = 2.0; long_context_output_cost_multiplier = 1.5 }
}
foreach ($model in $expected.Keys) {
    $entry = $pricing.PSObject.Properties[$model].Value
    if ($null -eq $entry) { throw "missing pricing entry: $model" }
    foreach ($field in $expected[$model].Keys) {
        $actual = [double]$entry.PSObject.Properties[$field].Value
        $want = [double]$expected[$model][$field]
        if ([math]::Abs($actual - $want) -gt 1e-18) { throw "unexpected ${model}.${field}: $actual != $want" }
    }
}
```

Manual evidence:
- Review the upstream provenance commit and confirm its JSON hunk is unchanged.
- Review the scoped diff for parser-only metadata retention and absence of billing-tier behavior changes.

## Output
- Write `docs/workflow/worker-results/gpt56-pricing-metadata-s205-result.md` using the worker-result format.
- The first line must be `### DONE: gpt56-pricing-metadata-s205`, `### BLOCKED: gpt56-pricing-metadata-s205`, or `### FAILED: gpt56-pricing-metadata-s205`.
- List changed files, commands run, test results, risks, contract compliance, and knowledge candidates.
- Do not write directly to `knowledge/**` or global memories.

## Stop Rules
- Stop if the upstream JSON patch no longer applies cleanly to latest `main`.
- Stop if correct behavior requires changing billing tier selection, handlers, schema/migrations, production configuration, or any denied path.
- Stop if tests expose an unrelated baseline failure that cannot be separated from this Sprint.
- Stop on ownership conflict or if preserving the main worktree's existing changes becomes impossible.

## Budget
- worker_mode: `current-agent-generator`
- qa_worker_mode: `independent-current-agent-review`
- configured matrix worker model `deepseek-v4-pro` is unavailable in the active collaboration tool; role separation and evidence gates remain mandatory.
- worktree_root: `E:/codex-worktrees`

## Worker Output
- Same requirements as `Output`.
