# Task Contract: upstream-openai-fast-flex-user-scope-s71

## Task ID

`upstream-openai-fast-flex-user-scope-s71`

## Status

`draft`

## Role

Generator adapts only the user-scoped OpenAI Fast/Flex policy behavior from upstream `f2966530c` to the current local gateway architecture.

## Goal

Allow a Fast/Flex policy rule to target trusted platform user IDs, with user-scoped rules evaluated before global rules across HTTP and WebSocket paths.

## Success Criteria

- `OpenAIFastPolicyRule` carries optional `user_ids` through service settings, DTO, bulk settings API, and frontend state.
- Empty or omitted `user_ids` means a global rule. Non-empty `user_ids` matches only the trusted positive `ctxkey.APIKeyUserID`; request body/header user values never participate.
- User-scoped rules as a group run before global rules regardless of array position. Within each group, configuration order and current first-match semantics remain unchanged.
- Missing, invalid, or different user identity skips user rules and continues to global rules.
- User, scope, tier, and model matching intersect; current whitelist/fallback action semantics remain unchanged.
- Zero, negative, or duplicate IDs are rejected by backend validation. The UI does not silently convert invalid user input into a global rule.
- Managed HTTP, API-key/OAuth passthrough HTTP, parsed WebSocket, and passthrough WebSocket first/follow-up frames use the same evaluator and trusted identity.
- UI loads, clones, edits, removes, and submits exact IDs; an empty list becomes global only after an explicit clear.

## Context

- Upstream reference: `f2966530c feat(openai): support user-scoped Fast/Flex policy`.
- Local code already has trusted `ctxkey.APIKeyUserID`; do not add a second user identity key or change auth middleware.
- Local evaluator lives in `openai_gateway_service.go`; upstream file layout no longer matches.

## Allowed Paths

- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/admin/admin_helpers_test.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_fast_policy_test.go`
- `backend/internal/service/openai_fast_policy_ws_test.go`
- `backend/internal/server/middleware/openai_fast_policy_forwarding_test.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `docs/workflow/worker-results/upstream-openai-fast-flex-user-scope-s71-result.md`

## Denied Paths

- `backend/internal/pkg/ctxkey/**`, API-key auth implementation/tests, and all WS relay/adapter implementation files.
- Ent, migrations, repositories, Wire, billing, pricing, model alias/catalog, payment, subscription, deployment, VERSION, Docker, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Reuse `ctxkey.APIKeyUserID`. Do not trust client-controlled body/header identity.
- Preserve `pass/filter/block`; do not introduce `force_priority` or another action.
- Preserve current scope, tier, whitelist, fallback, and group-internal first-match behavior.
- Do not repair unrelated unit-tag or full-suite drift.

## Acceptance Commands

```powershell
Push-Location backend
$servicePattern = "^(TestOpenAIFastPolicyUserScope_MatchingAndPrecedenceMatrix|TestOpenAIFastPolicyUserScope_ValidationAndRoundTrip|TestOpenAIWSUserScopedFastPolicy_ParsedIngressFirstAndFollowupFrames|TestOpenAIWSUserScopedFastPolicy_PassthroughFirstAndFollowupFrames)$"
$serviceList = @(go test ./internal/service -list $servicePattern)
if ($LASTEXITCODE -ne 0 -or @($serviceList | Where-Object { $_ -match '^Test' }).Count -ne 4) { throw "S71 service test discovery failed" }
go test ./internal/service -run $servicePattern -count=1
if ($LASTEXITCODE -ne 0) { throw "S71 service tests failed" }
go test ./internal/server/middleware -run "^TestAPIKeyAuthForwardsUserScopedOpenAIFastPolicyToManagedAndPassthroughHTTP$" -count=1
if ($LASTEXITCODE -ne 0) { throw "S71 HTTP forwarding test failed" }
go test ./internal/handler/admin -run "^TestOpenAIFastPolicySettingsFromDTO_PreservesUserIDs$" -count=1
if ($LASTEXITCODE -ne 0) { throw "S71 DTO test failed" }
go test ./internal/service -run "Test.*OpenAIFastPolicy|Test.*PassthroughBilling" -count=1
if ($LASTEXITCODE -ne 0) { throw "S71 service regressions failed" }
Pop-Location
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts -t `"loads, edits, and submits OpenAI Fast policy user IDs without widening scope`""
if ($LASTEXITCODE -ne 0) { throw "S71 frontend test failed" }
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
if ($LASTEXITCODE -ne 0) { throw "S71 frontend typecheck failed" }
$dirty = @(git status --porcelain --untracked-files=all)
if ($dirty.Count -gt 0) { throw "S71 acceptance requires a clean committed worktree: $($dirty -join ', ')" }
$base = (git merge-base HEAD codex/upstream-v0151-followups-s71-s73).Trim()
$allowed = @(
  "backend/internal/handler/dto/settings.go", "backend/internal/handler/admin/admin_helpers_test.go",
  "backend/internal/service/settings_view.go", "backend/internal/service/setting_service.go",
  "backend/internal/service/openai_gateway_service.go", "backend/internal/service/openai_fast_policy_test.go",
  "backend/internal/service/openai_fast_policy_ws_test.go", "backend/internal/server/middleware/openai_fast_policy_forwarding_test.go",
  "frontend/src/api/admin/settings.ts", "frontend/src/views/admin/SettingsView.vue",
  "frontend/src/views/admin/__tests__/SettingsView.spec.ts", "frontend/src/i18n/locales/en/admin/settings.ts",
  "frontend/src/i18n/locales/zh/admin/settings.ts", "docs/workflow/worker-results/upstream-openai-fast-flex-user-scope-s71-result.md"
)
$unexpected = @(git diff --name-only "$base..HEAD" | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -gt 0) { throw "S71 path audit failed: $($unexpected -join ', ')" }
git diff --check "$base..HEAD"
if ($LASTEXITCODE -ne 0) { throw "S71 diff check failed" }
```

## Output

- Write `docs/workflow/worker-results/upstream-openai-fast-flex-user-scope-s71-result.md` with the required DONE/BLOCKED/FAILED first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if trusted user identity is missing from real HTTP/WS paths and implementation would require auth or relay changes.
- Stop if migration, repository, billing, alias, new action, or client-controlled identity is required.
- Stop if user precedence changes group-internal first-match/fallback behavior.
- Stop if invalid UI input can silently widen a rule to global scope.
- Do not include `knowledge/05-current-focus.md` or any other main-worktree change.
