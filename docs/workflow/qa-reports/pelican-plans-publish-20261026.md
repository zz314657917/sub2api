# 鹈鹕计划提交前复核

日期：2026-10-26。基线：`ae097dfe790503677e6576fe8b835e69b2882bdb`。

## Findings

精确整理21个前后端、迁移和测试路径：

- 鹈鹕“管理测试计划”和账号“定时测试”加载后直接打开编辑弹窗；空列表直接新建，可在弹窗下拉切换计划或新建，也可返回列表。
- 优先测试账号按下拉选择顺序保存，先选优先，不可用顺延，最后有限自动调度；不改变全局账号优先级，provider 尝试后不换号重试。
- 采用既有对话框处理焦点和滚动锁，保存失败保留草稿，保存期间禁止重复提交和关闭，隔离过期会话响应。
- 包含迁移 `252_pelican_preferred_accounts.sql`，为计划增加非空有序账号数组，旧计划保留自动调度。

未发现本次提交范围中的新增阻断问题。原独立验收报告见 `pelican-plan-accounts-20261024.md`。

## Executed Checks

精确源码快照排除了其他未提交代码，前后端和迁移依赖齐全：

- 后端 Pelican service/repository/handler：101顶层、368含子项通过；10项真实数据库测试因未配置测试 DSN 跳过。
- 前端 Pelican 和定时测试面板：87项通过，包含单独运行的6项旧结果回归；类型检查、目标 ESLint和生产构建通过。
- 后端完整生产构建、目标 Go 文件 gofmt、精确暂存 diff 检查通过。
- 此前桌面1440px与窄屏390px模拟 API 浏览器验证通过；本轮没有重复浏览器验收，不把旧证据当作真实后端联调。

最新隔离提交快照及日志：`E:/codex-runtime/sub2api-publish-20261026`。原浏览器和入口验证证据：`E:/codex-runtime/pelican-plan-accounts-20261024`、`E:/codex-runtime/pelican-plan-entry-20261024`。

## Unverified Risks

迁移252仅随代码提交，尚未执行。未真实 PostgreSQL、provider及前后端联调，未部署或更新容器。

## Recommendation

本批可提交推送。后续部署须包含迁移252，再验证真实账号顺序及计划保存/执行。
