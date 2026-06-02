# Task Contract

## Task ID
upstream-main-safe-patches-s1

## Role
Codex acts as the Generator and Final Evaluator for this Sprint. Apply only the low-risk upstream patch sync defined here; do not make broader upstream architecture decisions.

## Goal
Port a conservative Sprint 1 subset of `upstream/main` fixes onto current local `main` without directly merging `upstream/main`. Preserve local Canvas, tickets, billing/payment, and public UI changes.

## Success Criteria
- Selected fixes are applied by cherry-pick or equivalent minimal porting.
- No new Ent schema, Ent generated code, SQL migrations, public API surface, production config, README/logo/deploy-only sync, or broad protocol bridge rewrite is introduced.
- Skipped upstream commits are documented with a reason and next-Sprint classification.
- Backend and frontend target checks pass, followed by the agreed Sprint 1 verification commands where feasible.

## Context
- Repo: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch: `main`
- Work branch: `codex/upstream-main-safe-patches-s1`
- Upstream source: `upstream/main`
- Main worktree `F:/mcplugins/sub2api` has unrelated dirty files and must not be modified by this task.

## Allowed Paths
- `backend/internal/service/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/pkg/**`
- `backend/internal/server/**`
- `backend/internal/domain/**`
- `backend/resources/model-pricing/**`
- `backend/go.mod`
- `backend/go.sum`
- `frontend/src/**`
- `frontend/package.json`
- `frontend/pnpm-lock.yaml`
- `docs/workflow/tasks/upstream-main-safe-patches-s1.md`
- `docs/workflow/worker-results/upstream-main-safe-patches-s1-result.md`
- `docs/workflow/qa-reports/upstream-main-safe-patches-s1-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Any public database schema, generated Ent code, production secrets/config, or unrelated local UI changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Prefer cherry-pick with conflict resolution that keeps local behavior when local code already contains an equivalent fix.
- Skip commits that require denied paths or broad bridge/gateway/apicompat rewrites.
- Skip commits that fail because prerequisite architecture is absent and would require widening Sprint 1.
- Do not touch `F:/mcplugins/sub2api` main worktree.

## Candidate Themes
- OAuth 401 credential preservation and OpenAI OAuth refresh enrichment.
- Usage/billing correctness: long-context cache multipliers, OpenAI 5h usage semantics, usage context preservation.
- Small OpenAI/WS/Responses bugfixes, excluding the Codex Responses bridge redesign and WS/HTTP bridge recovery rewrite.
- API Key Responses SSE fallback, group custom `/v1/models` list, ops local business limit reasons.

## Acceptance Commands
```powershell
git status --short --branch
go test ./internal/service ./internal/handler ./internal/repository -count=1
go test ./internal/server/routes ./cmd/server -count=1
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run lint:check
docker build -f tmp/Dockerfile.runtime-prebuilt -t sub2api-dev:runtime-prebuilt --build-arg ALPINE_IMAGE=sub2api-local-alpine:3.21 --build-arg POSTGRES_IMAGE=sub2api-local-postgres:18-alpine .
```

## Output
- Write `docs/workflow/worker-results/upstream-main-safe-patches-s1-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-safe-patches-s1-qa.md`.
- Update `docs/workflow/main-log.md` with major events.

## Stop Rules
- Stop if a selected commit requires Denied Paths.
- Stop if resolving a conflict requires changing schema/API/config beyond this contract.
- Stop if validation fails for reasons that cannot be fixed inside Allowed Paths.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
