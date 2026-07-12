# Task Contract: upstream-gpt56-bare-alias-catalog-s72

## Task ID

`upstream-gpt56-bare-alias-catalog-s72`

## Status

`draft`

## Role

Generator adapts only the bare GPT-5.6 alias, model catalog, whitelist, and OpenCode configuration subset from upstream `de28eba3c`.

## Goal

Route bare `gpt-5.6` to Sol for Codex OAuth and billing candidates while exposing the alias consistently in backend/frontend catalogs.

## Success Criteria

- `gpt-5.6`, `openai/gpt-5.6`, and `gpt5.6` normalize to `gpt-5.6-sol`.
- Bare GPT-5.6 accepts current known suffixes plus local `max`; unknown suffixes such as `ultra`, `solstice`, and `terrain` remain unknown passthrough values and never become Sol billing candidates.
- Explicit Sol/Terra/Luna names and legal suffixes remain mapped to their original models; Terra/Luna never collapse to Sol.
- OAuth/nil-account upstream normalization maps the bare alias to Sol; API-key compatible forwarding keeps the original bare name.
- `usageBillingModelCandidates` adds canonical Sol only for a recognized bare alias.
- Backend default models contain exactly one `gpt-5.6` entry named `GPT-5.6 (Sol)`.
- Frontend whitelist/preset and OpenCode configuration expose bare GPT-5.6; bare/Sol/Terra/Luna include `xhigh` and `max` with existing context/output limits.

## Allowed Paths

- `backend/internal/pkg/openai/constants.go`
- `backend/internal/pkg/openai/constants_test.go`
- `backend/internal/service/openai_model_alias.go`
- `backend/internal/service/openai_model_alias_test.go`
- `backend/internal/service/openai_model_mapping_test.go`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/composables/__tests__/useModelWhitelist.spec.ts`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`
- `docs/workflow/worker-results/upstream-gpt56-bare-alias-catalog-s72-result.md`

## Denied Paths

- Billing/pricing, usage extraction/repository, migrations, handlers, DTO/settings, payment/subscription, Ent/Wire, deployment, and production configuration.
- `openai_codex_transform.go`, `knowledge/**`, and global memories.

## Constraints

- Manually adapt the narrow subset; do not cherry-pick `de28eba3c`.
- Reuse existing suffix helpers and keep `max` local to GPT-5.6 alias logic.
- Unknown suffix means unrecognized passthrough, not a new HTTP error.
- Do not create independent bare-alias pricing.

## Acceptance Commands

```powershell
Push-Location backend
$pattern = "^(TestDefaultModelsIncludeBareGPT56Alias|TestNormalizeKnownOpenAICodexModel_BareGPT56RoutesToSol|TestNormalizeKnownOpenAICodexModel_GPT56RejectsUnknownSuffixes|TestIsOpenAIGPT56Model_BareAliasSuffixMatrix|TestUsageBillingModelCandidates_BareGPT56IncludesSol|TestNormalizeOpenAIModelForUpstream_BareGPT56AccountScopes)$"
$listed = @(go test ./internal/pkg/openai ./internal/service -list $pattern | Where-Object { $_ -match '^Test' })
if ($LASTEXITCODE -ne 0 -or $listed.Count -ne 6) { throw "S72 required test discovery failed: $($listed -join ', ')" }
go test ./internal/pkg/openai ./internal/service -run $pattern -count=1
if ($LASTEXITCODE -ne 0) { throw "S72 required tests failed" }
go test ./internal/service -run "GPT56|UsageBillingModelCandidates|NormalizeOpenAIModelForUpstream" -count=1
if ($LASTEXITCODE -ne 0) { throw "S72 service regressions failed" }
Pop-Location
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts"
if ($LASTEXITCODE -ne 0) { throw "S72 frontend tests failed" }
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
if ($LASTEXITCODE -ne 0) { throw "S72 frontend typecheck failed" }
$dirty = @(git status --porcelain --untracked-files=all)
if ($dirty.Count -gt 0) { throw "S72 acceptance requires a clean committed worktree: $($dirty -join ', ')" }
$base = (git merge-base HEAD codex/upstream-v0151-followups-s71-s73).Trim()
$allowed = @(
  "backend/internal/pkg/openai/constants.go", "backend/internal/pkg/openai/constants_test.go",
  "backend/internal/service/openai_model_alias.go", "backend/internal/service/openai_model_alias_test.go",
  "backend/internal/service/openai_model_mapping_test.go", "frontend/src/composables/useModelWhitelist.ts",
  "frontend/src/composables/__tests__/useModelWhitelist.spec.ts", "frontend/src/components/keys/UseKeyModal.vue",
  "frontend/src/components/keys/__tests__/UseKeyModal.spec.ts", "docs/workflow/worker-results/upstream-gpt56-bare-alias-catalog-s72-result.md"
)
$unexpected = @(git diff --name-only "$base..HEAD" | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -gt 0) { throw "S72 path audit failed: $($unexpected -join ', ')" }
git diff --check "$base..HEAD"
if ($LASTEXITCODE -ne 0) { throw "S72 diff check failed" }
```

## Output

- Write `docs/workflow/worker-results/upstream-gpt56-bare-alias-catalog-s72-result.md` with the required verdict first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if billing, usage, migration, settings, or transform changes are required.
- Stop if unknown suffixes map to Sol or explicit Terra/Luna map to Sol.
- Stop if bare alias requires an independent price.
