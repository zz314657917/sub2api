### PASS: hide-empty-subscriptions-s138

# QA Report

## Findings

- 未发现阻断 S138 的实现缺陷。订阅请求加载期间保留原 spinner，请求成功返回空数组后整个订阅面板退出布局。
- 有订阅时继续走原卡片渲染和续费跳转逻辑；下方用量分析与记录表不依赖订阅面板是否存在。
- 实现只复用组件现有 `loading` 和 `subscriptions` 状态，并移除了空状态删除后不再使用的 `Icon` import。

## Executed Checks

- 聚焦 Vitest：`frontend/src/views/user/__tests__/UsageView.spec.ts` 1 file、26/26 PASS；覆盖加载态、空结果隐藏、下方分析/表格保留，以及既有非空订阅与续费行为。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- changed-file ESLint：PASS，0 warnings。
- `corepack.cmd pnpm --dir frontend run build`：PASS，Vite production build 共转换 1109 modules。
- `git diff --check`、未合并索引、冲突标记与 contract 允许路径检查：PASS。
- 最终 diff review：业务变更仅位于订阅面板和 UsageView 聚焦测试；其余均为 S138 workflow 证据。

## Unverified Risks

- 未执行真实登录态浏览器 smoke；当前 UI 结论来自组件行为审查、聚焦 Vitest 和 production build。
- 已提交于隔离分支，尚未合入主工作树、push、部署或更新本地容器；当前运行容器不包含 S138。
- Vite 仍报告既有 Browserslist、dynamic/static import、chunk size 和 Node child-process deprecation 警告；本 Sprint 未新增依赖或构建配置，构建本身通过。

## Contract Compliance

- 只修改批准的订阅面板、UsageView 聚焦测试和 S138 workflow/交接文档。
- 未修改后端 API、持久化、schema、migration、路由、认证、计费、管理员页面、部署或容器。
- 空结果仅隐藏顶部订阅区域；请求失败时原 toast 行为和请求收尾逻辑保持不变。

## Recommendation

`PASS / source-level + production-build`。S138 已在隔离分支提交；集成、push 和本地容器刷新仍需用户明确授权，并应分别执行对应检查流程。
