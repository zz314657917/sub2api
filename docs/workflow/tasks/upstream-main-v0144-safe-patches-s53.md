# Task Contract: upstream-main-v0144-safe-patches-s53

## Task ID

`upstream-main-v0144-safe-patches-s53`

## Role

Generator / Codex direct integration.

## Goal

Port the narrow backend-safe subset from upstream `v0.1.144` after S45-S52:

- `e5dc1f597` token refresh treats `token_expired` as non-retryable.
- `4dd3aee5c` OpenAI Responses usage billing records the mapped billing model.
- `6bd248fd1` Codex import avoids merging access-only accounts into existing full accounts.

## Success Criteria

- The three upstream commits are replayed with `cherry-pick -x` on `codex/upstream-main-v0144-s53-safe-patches`.
- Token refresh no longer retries `token_expired` errors.
- OpenAI Responses billing uses mapped model metadata where the request model maps to a different upstream/billing model.
- Codex session import refuses to merge access-only imports into existing full accounts and keeps identity matching behavior from S46.
- Workflow status and handoff clearly record S53 outcome, validation, skipped larger v0.1.144 items, and next action.

## Allowed Paths

- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_import_test.go`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-v0144-safe-patches-s53.md`
- `docs/workflow/worker-results/upstream-main-v0144-safe-patches-s53-result.md`
- `docs/workflow/qa-reports/upstream-main-v0144-safe-patches-s53-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `.github/**`
- `README*`
- unrelated `payment` / `welfare` paths
- frontend files
- Docker/container files
- any broad v0.1.144 feature not listed in the Goal

## Constraints

- Do not merge `upstream/main` or tag `v0.1.144` directly.
- Do not include the larger v0.1.144 items in this batch: usage log queue backpressure, group capacity batching, concurrency cleanup, Codex image tool policy, error request UI alignment, Anthropic Fable 7d_oi, deploy migration timeout, Grok UI fixes, README changes.
- If conflicts hit denied paths, stop and reassess instead of resolving silently.
- Preserve existing local behavior unless the selected upstream commits intentionally change it.
- Keep commits scoped; do not use `git add .`.

## Acceptance Commands

Run from `backend` unless noted:

- `go test ./internal/service -run "TestIsNonRetryableRefreshError|TestTokenRefreshService_RefreshWithRetry|TestOpenAIGatewayServiceRecordUsage|TestOpenAIGatewayService_.*Mapped|TestOpenAIGatewayService_Forward" -count=1`
- `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex" -count=1`
- `git diff --check`
- `rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" .`
- denied-path audit over `git diff --name-only origin/main..HEAD`

Optional if frontend remains untouched:

- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"` may be skipped with reason, because S53 denied frontend changes.

## Output

- Cherry-picked commits plus any minimal conflict-resolution commits if needed.
- `docs/workflow/worker-results/upstream-main-v0144-safe-patches-s53-result.md`
- `docs/workflow/qa-reports/upstream-main-v0144-safe-patches-s53-qa.md`
- Updated `docs/workflow/status.md`, `docs/workflow/main-log.md`, and `knowledge/tasks/current-task.md`.

## Stop Rules

- Stop if any cherry-pick conflicts in denied paths.
- Stop if validation fails outside known unrelated unit-test baseline.
- Stop if the diff includes frontend, deploy, migrations, README, payment, welfare, or unrelated repository changes.
