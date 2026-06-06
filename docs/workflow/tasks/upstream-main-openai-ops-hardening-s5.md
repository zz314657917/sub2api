# Task Contract

## Task ID
upstream-main-openai-ops-hardening-s5

## Role
Codex acts as Generator and Final Evaluator for this Sprint. Implement only the approved low-risk OpenAI/Ops/admin backend hardening subset of upstream fixes.

## Goal
Port selected upstream backend fixes for OpenAI gateway stream termination, Codex usage snapshot self-healing, proxy quality status classification, group description persistence, and Claude Code client recognition onto a dedicated isolated branch without directly merging `upstream/main`. Preserve local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow changes.

## Success Criteria
- Selected upstream fixes are applied by cherry-pick or equivalent minimal porting.
- No Ent schema, SQL migration, frontend UI, public API field, production config, README/logo/deploy-only sync, DingTalk, notification email, user-platform quota, Channel Monitor API mode, upstream model sync service, or broad gateway refactor is introduced.
- OpenAI messages stream fallback handles missing terminal events without hanging or losing a deterministic terminal response.
- Codex usage snapshot handling can self-heal stale used-percent values and respects lock semantics.
- Allowed proxy quality statuses are classified as pass instead of warn.
- Administrators can clear a group description and persist the empty value.
- Claude Code clients are recognized from the billing block where prompt-only detection is insufficient.
- Skipped or deferred commits are documented with a reason.
- Target checks and feasible regression commands are executed and recorded.

## Context
- Repo: `F:/mcplugins/sub2api`
- Isolated worktree: `E:/codex-worktrees/sub2api/upstream-main-openai-ops-hardening-s5`
- Base branch: `main`
- Work branch: `codex/upstream-main-openai-ops-hardening-s5`
- Upstream source: `upstream/main`
- Baseline local commit: `b708d0552`

## Allowed Paths
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_compat_model_test.go`
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_gateway_service_codex_snapshot_test.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_proxy_quality_test.go`
- `backend/internal/service/admin_service_group_test.go`
- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `backend/internal/service/testdata/security_monitor_system_prompt.txt`
- `docs/workflow/tasks/upstream-main-openai-ops-hardening-s5.md`
- `docs/workflow/worker-results/upstream-main-openai-ops-hardening-s5-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-ops-hardening-s5-qa.md`
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

## Candidate Commits
- `8e27ff20a` fix(openai): handle missing messages stream terminal.
- `86d9b6bff` fix(openai): self-heal stale Codex used% snapshots + lock semantics.
- `32ef47110` fix: treat allowed proxy quality statuses as pass not warn.
- `bc7ce1857` fix(group): 管理员清空分组描述时正确持久化.
- `d626ccce1` fix: recognize claude code clients via billing block, not just prompt.

## Explicitly Deferred
- Broad upstream gateway architecture refactor chain beyond the listed commits.
- Image rate-limit cooldown/failover changes that touch runtime block, context keys, or broader rate-limit semantics.
- TTFT sample weighting or other commits that require SQL migrations.
- HTTP2 timeout changes that require config/repository/deploy scope expansion.
- DingTalk, notification emails, user-platform quota, Channel Monitor API mode, frontend UI, Ent/migration, version/sponsors, and CI-only changes.
- Upstream deletion or restructuring of local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow features.

## Constraints
- Do not direct-merge `upstream/main`.
- Work only inside the isolated worktree for this Sprint.
- Prefer `git cherry-pick -x`; if conflicts would touch denied paths or broaden scope, stop that commit and document it as deferred.
- Keep local behavior when local code already contains an equivalent fix.
- Do not run code generation that rewrites denied generated code.
- Do not include generated frontend build output, Docker artifacts, `node_modules`, or unrelated temp files.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status b708d0552..HEAD
git diff --name-only b708d0552..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"
go test ./internal/service -run "OpenAI|Codex|Proxy|Group|Claude|Terminal|Snapshot|Quality" -count=1
go test ./internal/handler ./internal/service -run "OpenAI|Gateway|Group|Proxy|Claude|Terminal" -count=1
go test ./internal/service ./internal/handler -count=1
```

Run `go test ./internal/server/routes ./cmd/server -count=1` only if this Sprint touches route/server wiring.

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-ops-hardening-s5-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-openai-ops-hardening-s5-qa.md`.
- Update `docs/workflow/main-log.md` with contract, implementation, and QA events.

## Stop Rules
- Stop a candidate commit if it requires denied paths.
- Stop a candidate commit if conflict resolution requires new schema, migration, frontend UI, public API fields, production config, or broad gateway architecture changes.
- Stop Sprint implementation if the working tree cannot be returned to a clean state between candidate commits.
