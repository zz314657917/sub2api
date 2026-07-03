---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 14:43 +08:00
---

# Task Contract

## Task ID
upstream-main-v0141-codex-reasoning-preserve-s48

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `73de2ea7` so Codex OAuth requests preserve `reasoning` input items across turns while stripping replay-unsafe `rs_*` ids and backfilling missing `summary` fields.

## Success Criteria
- `reasoning` items in Codex input are no longer dropped by `filterCodexInput`.
- `reasoning.id` is stripped even when reference preservation is enabled, preventing upstream `rs_*` lookup failures under `store=false`.
- `encrypted_content`, `content`, existing `summary`, and other reasoning fields are preserved verbatim.
- Missing or nil `summary` on reasoning items is backfilled as an empty array.
- Existing message, tool-call, `call_id`, and function-call-output pairing behavior remains unchanged.
- No frontend, i18n, Ent, migrations, deploy, README, `.github`, or knowledge files are modified.

## Allowed Paths
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `docs/workflow/tasks/upstream-main-v0141-codex-reasoning-preserve-s48.md`
- `docs/workflow/worker-results/upstream-main-v0141-codex-reasoning-preserve-s48-result.md`
- `docs/workflow/qa-reports/upstream-main-v0141-codex-reasoning-preserve-s48-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of `upstream/main` or the full release.
- Do not change OpenAI gateway account selection, billing, compact routing, image bridge behavior, or frontend usage views.
- Do not change tool/function call id normalization except to preserve existing behavior around `call_id` pairing.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestFilterCodexInput|TestApplyCodexOAuthTransform" -count=1
cd ..
git diff --check -- backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0141-codex-reasoning-preserve-s48.md docs/workflow/worker-results/upstream-main-v0141-codex-reasoning-preserve-s48-result.md docs/workflow/qa-reports/upstream-main-v0141-codex-reasoning-preserve-s48-qa.md
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/|deploy/|knowledge/|\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0141-codex-reasoning-preserve-s48`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0141-codex-reasoning-preserve-s48-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0141-codex-reasoning-preserve-s48-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if preserving reasoning requires touching gateway routing, billing, compact model selection, image bridge behavior, or frontend files.
- Stop if any `reasoning` item keeps an `rs_*` id after filtering.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 14:43 +08:00.
- Verdict: approved.
- Reason: the upstream patch is narrow, backend-only, fixes a concrete Codex OAuth multi-turn reasoning regression, and current local code still drops `reasoning` items.
