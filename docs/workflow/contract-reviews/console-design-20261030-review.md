### PASS: console-design-20261030

独立契约审查，2026-10-30。审查对象 SHA256：`2db0e5531990928af05120171fa9f2932c48b67a4ba09bfb96214edb5b0de9ad`。此结论仅批准进入开发，不是实现或运行验收通过。

## Findings

未发现阻断开发的契约缺口。隔离候选、脏文件保护、禁止业务写操作、独立 QA、冻结与集成时重查均有明确边界；构建命令已将产物移到外部证据目录。

实施及 QA 需落实以下已约定要求：

- 模型统计 API 确实存在：`frontend/src/api/usage.ts` 的 `usageAPI.getDashboardModels` 使用 GET `/usage/dashboard/models`，支持日期与 timezone；后端只返回当前用户请求模型统计。契约中的 “if available” 条件已满足，模型排行应作为必验项，不能以 API 不存在跳过。
- `getDashboardStats()` 不接受 timezone，而日期趋势与模型接口接受 timezone。实现应明确今日卡片与图表的日界线，使用现有只读日期统计接口或清楚标注统计时区，避免无说明地混用服务端今日与浏览器日期；无需扩大 API client/backend 修改范围。
- 范围切换的迟到响应测试应覆盖趋势、模型、近期记录及启用时的账户汇总，至少模拟旧请求晚于新请求返回；不能只验证模型标题或日期选择器变化。
- 新页面保留原有 onboarding 的显式创建入口和跳过持久化行为；禁止自动触发创建与领取。浏览器 fixture 验收应阻断未声明的 API 请求，避免 Vite 代理访问真实后端。

## Executed Checks

- 只读检查 AGENTS、Agent Matrix、当前 workflow status/spec、任务全文及其 SHA256；本任务是并发独立分支，主任务前置信息需保留。
- 核对实际国际化目录：`zh/en` 的 `dashboard.ts`、`common.ts`、`nav.ts` 均存在，且已有聚合导入；允许路径充分，不需要新增 locale 入口。
- 核对 DashboardView、现有 dashboard 组件、AppLayout/Header/Sidebar、usage API、相关类型和后端 GET 路由实现。所需实际费用、请求数、token、活跃 key 字段与 Chart.js 依赖均已存在。
- 核对 package.json 的 typecheck/build/test:run 命令存在；核对 Vite 构建/开发代理配置。未运行构建或浏览器，本轮没有产品代码 diff。

## Unverified Risks

深浅色、手机溢出、admin/dense workspace、动态菜单、feature flags、键盘焦点与竞态行为仍需冻结候选上的独立 QA；fixtures 不证明真实后端数据正确。基线测试成功由主控另行保留，本审查不冒充执行证据。

## Recommendation

主控可将契约更新为 approved/PASS 后分派 Developer；以上要求纳入实现与 QA 证据。完成前仍需独立 QA 和主控最终裁决。
