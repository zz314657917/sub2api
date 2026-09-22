# v0.2.7 主线集成验收

2026-09-22。此记录优先于此前各批报告中的“未提交、未合入”历史状态。

- 源代码保存提交：`fe12c25ca`，先前 Anthropic 提交：`d6e6c3471`。
- 最新 main 基线：`796003b8d`。使用精确业务补丁迁入隔离集成副本，未整体合并上游、未 cherry-pick。
- 集成代码与验收证据提交：`c03ed7f43`。49 个业务/测试文件逐一比对源工作树一致。
- 独立后端 QA：完整 apicompat、V027/OpenAI/CN 定向集合、Gemini unit 回归及 Go build 通过。
- 独立前端 QA：14 文件、24 用例、vue-tsc 与生产构建通过；确认最新 main 的 Select 与页码/订单筛选兼容。浏览器证据沿用上轮已通过并完成清理的组件 smoke。
- 集成仅新增专属 workflow 文件，不覆盖 main 中正在修改的 status、main-log、current-task 等共享文档。outputs、依赖 junction、私有构建产物均未提交。
- 本轮仅本地合入，不推送、不部署；真实供应商、支付、Redis/数据库未验证。

独立报告：`upstream-v027-main-integration-backend.md`、`upstream-v027-main-integration-ui.md`。
