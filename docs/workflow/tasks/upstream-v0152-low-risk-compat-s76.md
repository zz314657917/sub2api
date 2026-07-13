# Task Contract: upstream-v0152-low-risk-compat-s76

## Task ID

`upstream-v0152-low-risk-compat-s76`

## Status

`approved`

## Role

Generator manually adapts three independent low-risk subsets from upstream `v0.1.152`: Fast/Flex user search selection (`0464856c4`), Grok Composer reasoning sanitization (`aeb34d200`), and platform-aware no-account diagnostics (`8a22dc734`).

## Goal

Improve administration and Grok compatibility without importing the release's migration, billing, account-model, prompt-cache, or deployment changes.

## Success Criteria

- Fast/Flex rules replace raw user-ID entry with debounced email search while continuing to persist exact positive `user_ids`.
- Existing IDs hydrate to email labels when available; deleted or unresolved users remain visible and removable without widening the rule to global scope.
- Duplicate, non-positive, and non-integer IDs are not emitted by the selector.
- Grok Composer aliases `grok-composer`, `grok-composer-2.5-fast`, `composer-2.5`, including provider prefixes, strip top-level `reasoning`, `reasoning_effort`, and `reasoningEffort` before forwarding.
- Other Grok models preserve all reasoning fields.
- OpenAI-compatible account selection diagnoses against the API key group's actual platform and rewrites Grok selection log text without changing client-facing routing semantics.
- Responses, Messages, Chat Completions, and WebSocket paths keep their existing capacity/failover behavior; Count Tokens adopts the same platform-aware 404/503 no-account classification already used by the other HTTP paths.

## Allowed Paths

- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/views/admin/settings/OpenAIFastPolicyUserSelector.vue`
- `frontend/src/views/admin/settings/__tests__/OpenAIFastPolicyUserSelector.spec.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/i18n/__tests__/openaiFastPolicyLocales.spec.ts`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_test.go`
- `backend/internal/service/openai_gateway_grok_s76_test.go`
- `backend/internal/handler/no_account_error.go`
- `backend/internal/handler/no_account_error_test.go`
- `backend/internal/handler/no_account_error_s76_test.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-v0152-low-risk-compat-s76.md`
- `docs/workflow/worker-results/upstream-v0152-low-risk-compat-s76-result.md`
- `docs/workflow/qa-reports/upstream-v0152-low-risk-compat-s76-qa.md`

## Denied Paths

- Ent, migrations, repositories, Wire, billing, pricing, account schema/type support, prompt caching, payment, subscription, deployment, VERSION, Docker, and production configuration.
- Existing Fast/Flex backend policy evaluation, API-key authentication, WebSocket relay implementation, and user search API implementation.
- `knowledge/**` and global memories.

## Constraints

- Manually adapt the three subsets; do not merge or cherry-pick the full release.
- Preserve the current Fast/Flex settings namespace fixed during S71.
- Reuse `adminAPI.usage.searchUsers` and `adminAPI.users.getById`; do not add an endpoint.
- Sanitize only the known Composer aliases and only top-level reasoning fields.
- Platform-aware diagnostics must not alter account selection or failover. Count Tokens may adopt the shared classifier's existing 404/503 response contract; other routes keep their current response behavior.
- Do not install or update dependencies. A temporary junction to the main worktree's existing `frontend/node_modules` is allowed and must be removed after verification.

## Acceptance Commands

```powershell
Push-Location backend
$servicePattern = "^(TestPatchGrokResponsesBodySanitizesComposerReasoningParameters|TestPatchGrokResponsesBodyPreservesReasoningForOtherModels)$"
$serviceTests = @(go test ./internal/service -list $servicePattern | Where-Object { $_ -match '^Test' })
if ($LASTEXITCODE -ne 0 -or $serviceTests.Count -ne 2) { throw "S76 Grok test discovery failed: $($serviceTests -join ', ')" }
go test ./internal/service -run $servicePattern -count=1
if ($LASTEXITCODE -ne 0) { throw "S76 Grok sanitizer tests failed" }
$handlerPattern = "^(TestClassifyOpenAICompatibleNoAccountError_GrokUsesGrokPlatform|TestOpenAICompatibleSelectionErrorForLog_PreservesOtherPlatforms)$"
$handlerTests = @(go test ./internal/handler -list $handlerPattern | Where-Object { $_ -match '^Test' })
if ($LASTEXITCODE -ne 0 -or $handlerTests.Count -ne 2) { throw "S76 handler test discovery failed: $($handlerTests -join ', ')" }
go test ./internal/handler -run $handlerPattern -count=1
if ($LASTEXITCODE -ne 0) { throw "S76 handler diagnostics tests failed" }
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
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/settings/__tests__/OpenAIFastPolicyUserSelector.spec.ts src/i18n/__tests__/openaiFastPolicyLocales.spec.ts"
  if ($LASTEXITCODE -ne 0) { throw "S76 frontend tests failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts -t `"loads, edits, and submits OpenAI Fast policy user IDs without widening scope`""
  if ($LASTEXITCODE -ne 0) { throw "S76 Settings integration test failed" }
  cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
  if ($LASTEXITCODE -ne 0) { throw "S76 frontend typecheck failed" }
} finally {
  if ($createdModulesJunction) { Remove-Item -LiteralPath $frontendModules -Force }
}

git diff --check
if ($LASTEXITCODE -ne 0) { throw "S76 diff check failed" }
```

## Output

- Write `docs/workflow/worker-results/upstream-v0152-low-risk-compat-s76-result.md` with the implementation verdict and executed checks.
- Write `docs/workflow/qa-reports/upstream-v0152-low-risk-compat-s76-qa.md` with an independent final evidence summary.
- Return the changed-path audit and any unverified runtime risk.

## Stop Rules

- Stop if any slice requires migration, Ent, billing, new account types, prompt caching, or a new user API.
- Stop if existing Fast/Flex IDs cannot be preserved exactly.
- Stop if Grok sanitization would affect non-Composer models.
- Stop if platform-aware diagnostics require changing public response contracts or failover semantics.
