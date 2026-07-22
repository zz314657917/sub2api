### PASS: expired-subscription-route-skip-s99

## Findings

- 未发现阻塞 S99 的明确问题。
- Diff 精准性检查通过：后端变更只覆盖请求级失效订阅标记、路由过滤、模型目录过滤和既有绑定保留；前端 S99 只补失效分组快照、标签和禁选状态。`KeysView` 同时保留了用户此前未提交的紧凑单行路由布局、滚动和重复分组隐藏改动，没有回滚或混入无关页面。

## Executed Checks

- `go test ./internal/service -run "TestS99" -count=1`：PASS。
- `go test -tags=unit ./internal/server/middleware -run "TestS99" -count=1`：PASS。
- `go test ./internal/handler -run "TestS99|TestGatewayModels_MultiGroupRoutesAggregateRoutableModels|TestGatewayModels_CustomModelsListCanReturnEmptyWhenSelectionsUnavailable" -count=1`：PASS。
- `go test ./internal/service -run "TestS88|TestS91|TestS93|TestAPIKey.*Route|Test.*ResolveFor.*Request" -count=1`：PASS。
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/KeysView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts`：2 files / 18 tests PASS。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- `corepack.cmd pnpm --dir frontend run build`：PASS，Vite 转换 1089 modules；仅有既有动态导入、chunk size、Browserslist 和 Node deprecation 警告。
- `gofmt -d` 覆盖全部 S99 Go 文件：`GOFMT_CLEAN`。
- `git diff --check`、冲突标记扫描、未合并索引检查：PASS。
- 浏览器只读检查：`http://127.0.0.1:62100/keys` 可访问，但当前浏览器没有登录态，被重定向到 `/login?redirect=/keys`；未尝试登录或提交表单。

## Unverified Risks

- 未在真实登录态下打开含失效订阅路由的 Key 编辑弹窗；标签、禁选、保留提交由 Vitest 覆盖，实际视觉和交互仍缺一轮已登录浏览器 smoke。
- 未连接真实 PostgreSQL/Redis 构造“订阅过期后调用、续费后恢复”的端到端场景；当前证据来自服务、中间件和 handler 定向测试。
- 本地 `62080` 容器未更新，因此当前运行容器尚不包含 S99 源码。

## Recommendation

- `可继续`：S99 达到 source-only 交付标准。只有在需要本地运行态验收时，才进入容器更新和已登录 smoke；本轮不提交、不推送、不更新容器。
