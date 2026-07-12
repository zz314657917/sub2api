# Task Contract: upstream-gpt56-bare-alias-catalog-s72

## Task ID

`upstream-gpt56-bare-alias-catalog-s72`

## Status

`approved`

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
- Required backend assertions cover `ultra`, `solstice`, and `terrain` across normalization, `isOpenAIGPT56Model`, and billing candidates, plus nil/OAuth-to-Sol and API-key passthrough behavior.
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
- Add exact frontend tests named `exposes bare GPT-5.6 in whitelist and preset mappings without collapsing explicit variants` and `generates OpenCode config for bare and explicit GPT-5.6 max variants`.
- Frontend acceptance may create and must remove a temporary junction to `F:/mcplugins/sub2api/frontend/node_modules`; do not install or update dependencies.

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
$frontendModules = Join-Path $PWD "frontend/node_modules"
$sharedModules = "F:/mcplugins/sub2api/frontend/node_modules"
$createdModulesJunction = $false
if (-not (Test-Path -LiteralPath $frontendModules)) {
  if (-not (Test-Path -LiteralPath $sharedModules)) { throw "Shared frontend dependencies are unavailable" }
  New-Item -ItemType Junction -Path $frontendModules -Target $sharedModules | Out-Null
  $createdModulesJunction = $true
}
try {
  $whitelistTest = "exposes bare GPT-5.6 in whitelist and preset mappings without collapsing explicit variants"
  $openCodeTest = "generates OpenCode config for bare and explicit GPT-5.6 max variants"
  $whitelistList = @(cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest list src/composables/__tests__/useModelWhitelist.spec.ts -t `"$whitelistTest`" --no-color")
  if ($LASTEXITCODE -ne 0 -or @($whitelistList | Where-Object { $_ -match ([regex]::Escape($whitelistTest) + '$') }).Count -ne 1) { throw "S72 whitelist test discovery failed" }
  $openCodeList = @(cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest list src/components/keys/__tests__/UseKeyModal.spec.ts -t `"$openCodeTest`" --no-color")
  if ($LASTEXITCODE -ne 0 -or @($openCodeList | Where-Object { $_ -match ([regex]::Escape($openCodeTest) + '$') }).Count -ne 1) { throw "S72 OpenCode test discovery failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts -t `"$whitelistTest`""
  if ($LASTEXITCODE -ne 0) { throw "S72 whitelist test failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/keys/__tests__/UseKeyModal.spec.ts -t `"$openCodeTest`""
  if ($LASTEXITCODE -ne 0) { throw "S72 OpenCode test failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
  if ($LASTEXITCODE -ne 0) { throw "S72 frontend typecheck failed" }
} finally {
  if ($createdModulesJunction) { Remove-Item -LiteralPath $frontendModules -Force }
}
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
