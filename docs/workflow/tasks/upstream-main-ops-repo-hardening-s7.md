# Task Contract

## Task ID
upstream-main-ops-repo-hardening-s7

## Role
Codex acts as Generator and Final Evaluator for this Sprint. Implement only the approved Ops/repository/account hardening subset of upstream fixes.

## Goal
Port selected upstream fixes for Ops SLA exclusion, repository group/account counting, refresh-token retry classification, and scheduler account unscheduling onto a dedicated isolated branch without directly merging `upstream/main`. Preserve local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow changes.

## Context
- Repo: `F:/mcplugins/sub2api`
- Isolated worktree: `E:/codex-worktrees/sub2api/upstream-main-ops-repo-hardening-s7`
- Base branch: `main`
- Work branch: `codex/upstream-main-ops-repo-hardening-s7`
- Upstream source: `upstream/main`
- Baseline local commit: `c3625ce46`

## Candidate Commits
- `ae6ee23e2` fix: 调整 Ops 错误分类的 SLA 排除逻辑.
- `271aba1ab` fix(ops): exclude IP-denied access from SLA.
- `69305a609` fix(ops): 排除本地客户端限制错误的 SLA 计数.
- `ab6510f1a` fix(repo): 为公告查询添加分页上限，优化分组按账户数排序的数据加载.
- `5465003d0` test(group): 补充分组列表可用账号数与总账号数统计正确性的集成测试.
- `df2b02e61` fix: 修正分组账号可用计数口径.
- `49b415e33` fix: mark reused refresh tokens non-retryable.
- `202aab8e6` fix(accounts): unschedule errored accounts.

## Allowed Paths
- `backend/internal/handler/ops_error_logger*`
- `backend/internal/server/middleware/api_key_auth*`
- `backend/internal/service/ops_upstream_context*`
- `backend/internal/service/token_refresh_service*`
- `backend/internal/repository/announcement_repo*`
- `backend/internal/repository/group_repo*`
- `backend/internal/repository/account_repo*`
- `docs/workflow/tasks/upstream-main-ops-repo-hardening-s7.md`
- `docs/workflow/worker-results/upstream-main-ops-repo-hardening-s7-result.md`
- `docs/workflow/qa-reports/upstream-main-ops-repo-hardening-s7-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `.github/**`
- `assets/**`
- `README*`

## Constraints
- Prefer `git cherry-pick -x`.
- If a candidate requires denied paths, new API fields, schema/migration, frontend changes, or broad refactor, stop that candidate and record `DEFERRED`.
- Do not run code generation that rewrites denied generated files.
- Do not include generated frontend build output, Docker artifacts, `node_modules`, or unrelated temp files.
- Keep local behavior when local code already contains an equivalent fix.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status c3625ce46..HEAD
git diff --name-only c3625ce46..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"
go test ./internal/handler ./internal/server/middleware ./internal/service -run "Ops|SLA|IP|Denied|Client|Token|Refresh|Scheduler|Account" -count=1
go test ./internal/repository -run "Announcement|Group|Account|Available|Sort|Count" -count=1
go test ./internal/repository ./internal/service ./internal/handler ./internal/server/middleware -count=1
```

Run `go test ./internal/server ./cmd/server -count=1` only if this Sprint touches route/server wiring or API contract expectations.

## Output
- Write `docs/workflow/worker-results/upstream-main-ops-repo-hardening-s7-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-ops-repo-hardening-s7-qa.md`.
- Update `docs/workflow/main-log.md` with contract, implementation, QA, and integration events.

## Stop Rules
- Stop a candidate commit if it requires denied paths.
- Stop a candidate commit if conflict resolution requires new schema, migration, frontend UI, public API fields, production config, or broad architecture changes.
- Stop Sprint implementation if the working tree cannot be returned to a clean state between candidate commits.
