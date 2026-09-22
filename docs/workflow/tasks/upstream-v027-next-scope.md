# v0.2.7 后续批次范围

本批基线：38a5a6040；隔离工作树：E:/codex-worktrees/sub2api/upstream-v027-next。

- 当前实施 bcc73f8d4：DeepSeek Responses 转 Chat 时补缺失 reasoning_content，并保留输入明文。上游独立 pipeline 在本地不存在，接入本地 fallback；共享 bridge 只增加 DeepSeek opt-in，默认转换不变。
- 当前实施 8e34ca5e3：暂停调度的 active OAuth 账号仍刷新 token。上游分页 SQL owner 在本地不存在，本地修复 TokenRefreshService.listActiveAccounts 的二次过滤；不解除管理员暂停。
- 后续候选 18d483c2a：分组用量尾段查询优化，须另审 SQL 水位、时区、无效状态回退和实际查询计划；不混入本批。
- 后续候选 a9ed21898：兑换记录分页，须前后端兼容旧数组客户端并验收稳定排序；不混入本批。
- a9c7c6e8b 的手机模型入口显示行为本地已用常显 /models 图标实现；辅助标签完整性可单独补查。
- 插件宿主能力、Seedance 和依赖升级保留在独立评估范围；不从相关大提交连带引入。

主工作区已有商店、支付等未提交内容，不覆盖或纳入本批。临时补丁文件保留为未跟踪文件，按精确路径提交时排除。默认不部署或调用真实付费供应商。
