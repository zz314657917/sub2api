### DONE: upstream-v0200-composite-gateway-selection-s295-p1

## 实现

- 新增 `CompositeModelRoute` Ent schema、迁移 `241_composite_model_routes.sql`、生成代码、仓储和窄的 Composite 路由管理 service。
- 分组创建和更新请求接受 `platform=composite`；管理员可在既有 `/api/v1/admin/groups/:id` 资源下列出、创建、更新、删除和静态预览路由。
- `preview` 只筛选并返回已启用、模型和 endpoint 相符的有序注册记录；未接入账号、`model_mapping`、Gateway 选路、上下文或 provider 分发。
- 分组级级联删除会在原有事务内软删除该分组的 Composite 路由。
- 修复了 `Update`/`Delete` 在成员归属检查阶段把仓储错误误报为路由不存在的问题；该错误现会原样传播。
- Ent generator 曾因 Windows 锁定中断，本轮以本地 `upstream/main` 的同实体生成片段精确补齐 `runtime/runtime.go` 的 hook、interceptor、默认值和 validator 初始化。未改动任何合同外生成文件。
- `go generate ./cmd/server` 被两项既有 Wire provider 缺失阻断，未产出文件；仅在合同允许的 `cmd/server/wire_gen.go` 补入 P1 的 `CompositeModelRouteRepository -> CompositeRouteAdminService -> GroupHandler` 注入，随后以 server 编译验证。

## 已执行命令

- `gofmt -w`（合同列出的 P1 Go 文件）通过。
- `go test ./internal/service -run 'CompositeRoute' -count=1` 通过。
- `go test ./internal/handler/admin -run 'CompositeRoute' -count=1` 通过。
- `go test ./internal/server/routes -run 'CompositeRoute' -count=1` 通过。
- `go test ./cmd/server -run '^$' -count=1` 通过。
- `go build ./...` 通过。
- `git diff --check` 通过；`git ls-files -u` 无输出。

## 未执行与风险

- 未执行 `241_composite_model_routes.sql`，未访问 PostgreSQL、Redis、provider、容器或部署环境。
- `go generate ./ent` 没有重跑：此前 Windows 锁会写入合同外 Ent 文件；当前生成集通过编译，但完整可再生性仍需先解除该锁。
- `go generate ./cmd/server` 的现存失败为缺少 `[]service.ProxyRepository` 和 `service.newUserTrialConsumer` provider，均不在 P1 允许范围内。
- P2/P3 的 resolver、账号归属、请求 context 和 Gateway/provider dispatch 保持未实现。
- 独立 Terra QA 尚未获得，最终 QA verdict 见同任务报告。

## knowledge_candidates

- 无；当前限制、生成器故障和 QA 状态已同步到项目任务快照。
