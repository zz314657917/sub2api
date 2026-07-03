---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 13:43 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-codex-import-identity-s46

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `a5638a4e5` so Codex session imports match by `chatgpt_user_id` before the shared ChatGPT account id, preventing one team member's import from overwriting another member's account credentials.

## Success Criteria
- Identity keys are ordered by strength: `user:` first, email fallback only when user/account are absent, access-token fingerprint, then shared `account:` last.
- Existing account lookup keeps all candidates for shared `account:` keys and skips candidates whose stored `chatgpt_user_id` conflicts with the imported user id.
- Legacy accounts without stored `chatgpt_user_id` can still be matched by shared `account:` key and backfilled, with a warning.
- In-batch duplicate detection applies the same account-key conflict rule so different team members are not treated as duplicates.
- Large Codex session imports get a request-scoped 120s frontend timeout without changing global API client timeout.

## Allowed Paths
- `backend/internal/handler/admin/account_codex_import.go`
- `backend/internal/handler/admin/account_codex_import_test.go`
- `frontend/src/api/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-v0143-codex-import-identity-s46.md`
- `docs/workflow/worker-results/upstream-main-v0143-codex-import-identity-s46-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-codex-import-identity-s46-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/service/**`
- `backend/internal/repository/**`
- `frontend/src/i18n/**`
- `frontend/src/views/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of upstream `v0.1.143` or `upstream/main`.
- Do not change account create/update service semantics outside Codex import matching.
- Do not change global API client timeout.
- Do not introduce migrations, Ent generation, or wire changes.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44
go test ./internal/handler/admin -run "TestCodexIdentity|TestParseCodexSessionImport|TestNormalizeCodexImport|TestResolveCodexImport|TestMergeCodexImport|TestImportCodex" -count=1
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
git diff --check
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|backend/internal/service/|backend/internal/repository/|frontend/src/i18n/|frontend/src/views/|deploy/|knowledge/|\\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-codex-import-identity-s46`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-codex-import-identity-s46-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-codex-import-identity-s46-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if account-key fallback no longer updates legacy accounts missing `chatgpt_user_id`.
- Stop if different `chatgpt_user_id` values under the same shared `chatgpt_account_id` can match or deduplicate each other.
- Stop if staged diff includes denied paths.
- Stop if frontend timeout change affects global `apiClient`.

## Review Result
- Reviewed at: 2026-07-03 13:43 +08:00.
- Verdict: approved.
- Reason: the upstream fix is narrow, security/data-integrity relevant, and directly prevents credential overwrite between ChatGPT team members while preserving legacy fallback behavior.
