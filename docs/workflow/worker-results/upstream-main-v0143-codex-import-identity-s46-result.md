### DONE: upstream-main-v0143-codex-import-identity-s46

## Summary
- Ported upstream `a5638a4e5` Codex session import identity matching fix.
- Changed identity matching to prefer `chatgpt_user_id` before the shared `chatgpt_account_id`.
- Preserved legacy fallback: accounts missing `chatgpt_user_id` can still be claimed by shared account id and backfilled, with a warning.
- Kept all candidates for a shared account key so another team member cannot shadow a legacy account by row order.
- Applied the same identity conflict rule to in-batch duplicate detection.
- Added a request-scoped 120s timeout for the frontend Codex session import API call.

## Changed Files
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_import_test.go`
- `frontend/src/api/admin/accounts.ts`
- Workflow evidence files for S46.

## Commands Run
- `go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex" -count=1`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - Result: PASS.
- `git diff --check`
  - Result: PASS.

## Contract Compliance
- Did not merge all upstream `v0.1.143` or `upstream/main`.
- Did not touch account service/repository semantics, Ent, migrations, wire, i18n, views, deploy, `.github`, README, or knowledge paths.
- The frontend timeout is scoped to `importCodexSession`; global `apiClient` timeout remains unchanged.

## Risks
- If multiple legacy accounts share the same `chatgpt_account_id` and all lack `chatgpt_user_id`, the first matching candidate can still depend on query order; S46 preserves upstream behavior and warns only when claiming one legacy row.
- Reverse proxy or deployment timeouts shorter than 120s can still interrupt very large imports before the frontend timeout.
