### PASS: upstream-v024-image25-oauth-s299

# QA Report

## Task ID
upstream-v024-image25-oauth-s299

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-v024-image25-oauth-s299.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
go test ./internal/pkg/openai -run 'TestDefaultModelsIncludeGPTImage25' -count=1 -> PASS
go test ./internal/service -run 'Test(OpenAIImagesResponsesDriverAndImageModels|OpenAIImagesRejectedDriverDoesNotCoolImageModel|GPTImage25.*|FetchOpenAIAccountModelsOAuth.*|AccountTestService_OpenAIImageOAuth.*)' -count=1 -> PASS
go build ./... -> PASS
npm.cmd exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts -> PASS (15/15)
npm.cmd run typecheck -> PASS
docker compose -f deploy/docker-compose.yml config --quiet -> PASS
docker compose -f deploy/docker-compose.dev.yml config --quiet -> PASS
docker compose -f deploy/docker-compose.local.yml config --quiet -> PASS
docker compose -f deploy/docker-compose.standalone.yml config --quiet -> PASS
PowerShell ConvertFrom-Json model_prices_and_context_window.json -> PASS
gofmt -d on allowed Go paths -> PASS
git diff --check on allowed paths -> PASS
git diff --name-only --diff-filter=U -> PASS
```
- manual checks:
```text
Responses driver rejection -> actionable OpenAIImagesUpstreamError, zero model cooldown, zero temp cooldown
Requested image-model rejection -> UpstreamFailoverError, exactly one model cooldown near 30 minutes, zero temp cooldown
Explicit price entry -> wins over Image 2.5 fallback
OAuth model mapping -> Flare/Sunburst added only when allowed by account mapping
ImageInputTokens -> existing implementation retained without duplicate changes
```

## Findings
- Initial independent QA found a real topology gap: local generic rate-limit
  handling does not treat the Codex plan-gated 400 response as a model cooldown.
  The first implementation therefore proved driver passthrough but did not
  preserve image-model failover.
- The remediation classifies the exact upstream message in the Images owner.
  Driver rejection bypasses image cooldown; requested image-model rejection
  writes one bounded cooldown and returns the existing failover error.
- No remaining blocking issue was found. Catalogs, OAuth picker, frontend
  whitelist, driver override, static/fallback pricing and Compose propagation
  match the approved contract.

## Bug Owner Recommendation
`original-worker`

## Root Cause
- `implementation-bug` (remediated and independently retested)

## Retest Scope
- Completed: both plan-gated rejection branches, focused backend/frontend,
  build, Compose config, JSON, formatting, exact diff and conflict checks.

## Unverified Risks
- Real OpenAI provider availability for the configured Responses driver and
  Image 2.5 models was not exercised.
- Containers were not started; deployment and push were not performed.
- The current runtime pricing struct has no separate image-cache input field;
  the official 2e-6 image-cache price remains represented in static pricing JSON.

## Knowledge Promotion
- `none`
