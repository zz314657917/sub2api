---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v0200-composite-gateway-selection-s295-p1
worker_model: gpt-5.6-terra
base_commit: 42ab6da534c2a26c4d30176f9228f76e51bae839
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-08
---

# S295-P1 Composite 路由持久化与后台管理合同

## Task ID

`upstream-v0200-composite-gateway-selection-s295-p1`

## Role

Generator 在当前本地拓扑中实现 Composite 分组的持久化和受认证后台管理数据面。上游代码只作为行为参考；不得 merge、rebase 或 cherry-pick 其历史。

## Goal

让管理员可以创建 `platform=composite` 的分组，并为其维护模型路由记录。路由记录可通过后台只读预览列出候选项，但本批不参与任何账号选择、请求上下文写入或上游请求分发。

## Success Criteria

- 新增 `CompositeModelRoute` Ent schema、生成代码和前向迁移 `241_composite_model_routes.sql`。迁移含软删除、活跃记录唯一约束、索引和本地已支持的八种具体目标平台；本批绝不执行迁移。
- 新增路由仓储和专用管理 service；创建、更新、删除、按分组列出都只允许 `platform=composite` 的分组，并校验 public model、match type、endpoint、具体 target platform 与 priority。分组级删除会在同一事务中软删除其路由。
- 管理员 API 仅在既有 admin auth/audit middleware 下暴露：
  `GET/POST /api/v1/admin/groups/:id/composite-routes`、
  `PUT/DELETE /api/v1/admin/groups/:id/composite-routes/:route_id`、
  `POST /api/v1/admin/groups/:id/composite-routes/preview`。
- `preview` 仅返回请求 model/endpoint 的已启用候选路由，按仓储既有 priority/id 顺序展示；不返回账号、不读 `model_mapping`、不做平台 detector 或最终调度决策。无候选时返回空列表而非猜测目标平台。
- 创建和更新分组的请求校验接受 `composite`；既有非 Composite 分组、Gateway 选择、请求体、账号绑定、计费和缓存行为保持不变。
- 生成后的 Wire 图可编译；定向 service、handler、route 测试、server 编译、backend build、格式和 diff/冲突检查通过，或记录已有无关基线。

## Context

- Repo: `F:/mcplugins/sub2api`
- Local base: `42ab6da534c2a26c4d30176f9228f76e51bae839`。
- Planner evidence: `docs/workflow/plans/upstream-v0200-composite-gateway-selection-s295-replan.md`。
- Upstream references: `07619d491` (schema/migration), `6ba96841e` (repository/domain) and `a008b63c1` (admin registry). The current branch does not contain that feature chain.
- Local topology keeps group administration in `handler/admin/group_handler.go`. A narrow Composite route admin service is preferred over widening the large `AdminService` interface and every existing test double.

## Allowed Paths

### Persistence and generated Ent ownership

- `backend/ent/schema/composite_model_route.go`
- `backend/ent/client.go`
- `backend/ent/compositemodelroute.go`
- `backend/ent/compositemodelroute/compositemodelroute.go`
- `backend/ent/compositemodelroute/where.go`
- `backend/ent/compositemodelroute_create.go`
- `backend/ent/compositemodelroute_delete.go`
- `backend/ent/compositemodelroute_query.go`
- `backend/ent/compositemodelroute_update.go`
- `backend/ent/ent.go`
- `backend/ent/hook/hook.go`
- `backend/ent/intercept/intercept.go`
- `backend/ent/migrate/schema.go`
- `backend/ent/mutation.go`
- `backend/ent/predicate/predicate.go`
- `backend/ent/runtime/runtime.go`
- `backend/ent/tx.go`
- `backend/migrations/241_composite_model_routes.sql`
- `backend/internal/repository/composite_model_route_repo.go`
- `backend/internal/repository/group_repo.go`
- `backend/internal/repository/wire.go`

### Service, admin handler and route registration

- `backend/internal/service/composite_model_route.go`
- `backend/internal/service/composite_route_admin_service.go`
- `backend/internal/service/composite_route_admin_service_test.go`
- `backend/internal/service/wire.go`
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/admin/group_handler_composite_route_test.go`
- `backend/internal/handler/wire.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/server/routes/composite_route_admin_test.go`
- `backend/cmd/server/wire_gen.go`

### Workflow evidence

- `docs/workflow/tasks/upstream-v0200-composite-gateway-selection-s295-p1.md`
- `docs/workflow/contract-reviews/upstream-v0200-composite-gateway-selection-s295-p1-review.md`
- `docs/workflow/worker-results/upstream-v0200-composite-gateway-selection-s295-p1-result.md`
- `docs/workflow/qa-reports/upstream-v0200-composite-gateway-selection-s295-p1-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `docs/workflow/status.md` and all pre-existing S295 contract/review files.
- `backend/migrations/240_refresh_codex_usage_tutorial.sql` and every migration other than `241_composite_model_routes.sql`.
- `backend/internal/service/gateway_service.go`, `backend/internal/service/openai_gateway_service.go`, all Gateway endpoint services, `backend/internal/pkg/ctxkey/**`, and all request dispatch/context propagation code.
- `backend/internal/server/http.go`, `backend/internal/server/router.go`, `backend/internal/server/routes/gateway.go`, provider handlers and middleware.
- `frontend/**`, `outputs/**`, `sub2api`, `pelican-bicycle.html`, all existing dirty business paths and every unlisted path.
- Database, Redis, provider traffic, container replacement, deployment, commit, push and shared runtime state.

## Constraints

- The new `composite` group platform is an administration-only value in P1. It must not become selectable by Gateway code until P3 has an approved resolver and dispatch contract.
- Preview is a registry inspection endpoint, not an acceptance claim about runtime selection. It must fail closed by returning no candidates for unknown inputs.
- Keep a narrow Composite route admin interface injected into `GroupHandler`; do not expand `AdminService` or alter its positional constructor callers.
- `go generate ./ent` and `go generate ./cmd/server` may update only the generated files listed above. If either generator needs an unlisted source/dependency change, stop and return to Planner.
- Preserve every existing dirty change byte-for-byte. Do not execute the new migration against any database.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go generate ./ent
go generate ./cmd/server
go test ./internal/service -run 'CompositeRoute' -count=1
go test ./internal/handler/admin -run 'CompositeRoute' -count=1
go test ./internal/server/routes -run 'CompositeRoute' -count=1
go test ./cmd/server -run '^$' -count=1
go build ./...
gofmt -w internal/repository/composite_model_route_repo.go internal/repository/group_repo.go internal/repository/wire.go internal/service/composite_model_route.go internal/service/composite_route_admin_service.go internal/service/composite_route_admin_service_test.go internal/service/wire.go internal/handler/admin/group_handler.go internal/handler/admin/group_handler_composite_route_test.go internal/handler/wire.go internal/server/routes/admin.go internal/server/routes/composite_route_admin_test.go

Set-Location F:/mcplugins/sub2api
git diff --check
git ls-files -u
git status --short
```

## Output

- Worker result 首行必须为 `### DONE: upstream-v0200-composite-gateway-selection-s295-p1`、`### BLOCKED: ...` 或 `### FAILED: ...`，并列出变更、命令、结果、风险和 `knowledge_candidates`。
- QA report 必须独立记录 allowlist、生成文件审计、命令、未验证运行态和 PASS/FAIL/BLOCKED。

## Stop Rules

- 需要修改 denied path、启用 Gateway 选路、读取账号 `model_mapping`、写请求 context、运行迁移或接触任何共享运行资源时，立即停止并回到 Planner。
- 若生成器修改未列出的文件、路由语义无法在不引入 resolver 的前提下定义，或定向失败需要无关修复，保留证据并标记 BLOCKED。
