# Task Contract

## Task ID
upstream-main-hardening-s3

## Role
Codex acts as Generator and Final Evaluator for this Sprint. Implement only the approved hardening subset of upstream fixes.

## Goal
Port the next low-risk hardening subset from `upstream/main` onto a dedicated local branch without directly merging `upstream/main`. Preserve local model market, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow changes.

## Success Criteria
- Selected upstream fixes are applied by cherry-pick or equivalent minimal porting.
- No Ent schema, SQL migration, frontend UI, public API field, production config, README/logo/deploy-only sync, or broad gateway refactor is introduced.
- API Key unauthorized access returns `404` instead of exposing key existence.
- API Key names are escaped before unsafe HTML can be rendered.
- User deletion and user API Key deletion are transactionally consistent.
- Redis, setup, EasyPay, and fixed quota window behavior remains interface-compatible while gaining the intended fixes.
- Skipped or deferred commits are documented with a reason.
- Target checks and feasible regression commands are executed and recorded.

## Context
- Repo: `F:/mcplugins/sub2api`
- Base branch: `main`
- Work branch: `codex/upstream-main-hardening-s3`
- Upstream source: `upstream/main`
- Baseline local commit: `6a091a835`
- Baseline upstream commit: `b76f9524b`

## Allowed Paths
- `backend/internal/service/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/payment/provider/**`
- `backend/internal/setup/**`
- `backend/internal/server/**`
- `backend/internal/pkg/**`
- `docs/workflow/tasks/upstream-main-hardening-s3.md`
- `docs/workflow/worker-results/upstream-main-hardening-s3-result.md`
- `docs/workflow/qa-reports/upstream-main-hardening-s3-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `.github/**`
- `deploy/**`
- `assets/**`
- `README*`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`

## Candidate Commits
- `0ae332961` fix stored XSS by escaping API Key names.
- `11b601717` return 404 for unauthorized API Key access.
- `585ff0944` replicate Redis Lua `TIME` effects for Redis 3.2-4.x.
- `8a56c9fa0` bootstrap Postgres setup with maintenance DB.
- `04deb819b` use `trade_status` for EasyPay query status.
- `bba86f97d` make `userRepo.Delete` reuse caller transaction.
- `705fe7d88` delete user API Keys when deleting a user.
- `fb0195f3d` normalize fixed quota windows on account edit.

## Explicitly Deferred
- Gateway, WS, Images, Responses, and apicompat fixes: next `gateway-compat-s4`.
- DingTalk OAuth, notification email templates, user-platform quota, channel-monitor API mode, and other migration-sized features: separate high-risk migration Sprints.
- Upstream deletion or restructuring of local product features and workflow knowledge.

## Constraints
- Do not direct-merge `upstream/main`.
- Prefer `git cherry-pick -x`; if conflicts would touch denied paths or broaden scope, stop that commit and document it as deferred.
- Keep local behavior when local code already contains an equivalent fix.
- Do not run code generation that rewrites denied generated code.
- Do not include generated frontend build output, Docker artifacts, `node_modules`, or unrelated temp files.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test ./internal/service ./internal/handler ./internal/repository ./internal/payment/provider ./internal/setup -run "APIKey|User|Delete|Redis|Concurrency|Session|EasyPay|Setup|Quota" -count=1
go test ./internal/service ./internal/handler ./internal/repository -count=1
go test ./internal/server/routes ./cmd/server -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-hardening-s3-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-hardening-s3-qa.md`.
- Update `docs/workflow/main-log.md` with contract, implementation, and QA events.

## Stop Rules
- Stop a candidate commit if it requires denied paths.
- Stop a candidate commit if conflict resolution requires new schema, migration, frontend UI, public API fields, production config, or broad gateway architecture changes.
- Stop Sprint implementation if the working tree cannot be returned to a clean state between candidate commits.
