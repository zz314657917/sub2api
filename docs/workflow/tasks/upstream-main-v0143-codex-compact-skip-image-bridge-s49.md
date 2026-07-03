---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 14:55 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-codex-compact-skip-image-bridge-s49

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `c797159bf` so Codex image-generation bridge injection is skipped for `/responses/compact`, preventing upstream 400 errors caused by unsupported `tool_choice` on compact requests.

## Success Criteria
- `/responses/compact` requests do not receive injected `tools:[{"type":"image_generation"}]`, `tool_choice`, or Codex image bridge instructions.
- Non-compact Codex requests keep existing image bridge behavior when the bridge is enabled and the group/account permits image generation.
- Explicit non-compact image-generation requests keep existing normalization and validation behavior.
- Compact-only model mapping and Codex OAuth compact transform behavior remain unchanged.
- No frontend, i18n, Ent, migrations, deploy, README, `.github`, or knowledge files are modified.

## Allowed Paths
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_image_generation_controls_test.go`
- `docs/workflow/tasks/upstream-main-v0143-codex-compact-skip-image-bridge-s49.md`
- `docs/workflow/worker-results/upstream-main-v0143-codex-compact-skip-image-bridge-s49-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-codex-compact-skip-image-bridge-s49-qa.md`
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
- Do not change image-generation permission semantics for normal `/responses` requests.
- Do not change compact account selection, compact model mapping, session-id behavior, billing, or passthrough behavior.
- Local code currently computes `isCompactRequest` after bridge injection; implement this as a minimal precomputed guard or equivalent narrow check, not a broad `Forward` refactor.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestOpenAIGatewayServiceForward_CodexBridge|TestOpenAIGatewayServiceForward_.*Image|TestOpenAIGatewayService_CodexImageGenerationBridge" -count=1
cd ..
git diff --check -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_image_generation_controls_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-codex-compact-skip-image-bridge-s49.md docs/workflow/worker-results/upstream-main-v0143-codex-compact-skip-image-bridge-s49-result.md docs/workflow/qa-reports/upstream-main-v0143-codex-compact-skip-image-bridge-s49-qa.md
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/|deploy/|knowledge/|\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-codex-compact-skip-image-bridge-s49`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-codex-compact-skip-image-bridge-s49-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-codex-compact-skip-image-bridge-s49-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if fixing compact requires changing unrelated image billing, gateway selection, OAuth transform behavior, or frontend files.
- Stop if `/responses/compact` still contains injected `tool_choice` or image_generation bridge instructions under bridge-enabled Codex requests.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 14:55 +08:00.
- Verdict: approved.
- Reason: the upstream patch is narrow, backend-only, and fixes a concrete compact endpoint incompatibility without touching local product UI or database schema.
