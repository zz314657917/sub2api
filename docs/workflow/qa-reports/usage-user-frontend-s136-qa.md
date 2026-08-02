### PASS: usage-user-frontend-s136

# QA Report

## Findings

- 未发现阻断 S136 的实现缺陷。用户 Usage 分析、错误请求 opt-in、严格脱敏详情、设置持久化映射及桌面/移动布局均符合 contract。
- 最终抽查修复了两项移动端问题：分组/端点圆环与表格横向挤压，以及缓存 Token tooltip 和超长分组徽标造成的页面横向溢出。
- 最终抽查修复了错误列表缓存一致性：用户修改日期后重新进入错误 Tab 会按当前条件刷新，不再复用旧列表。
- `frontend/src/stores/app.ts` 的 public-setting fallback 字段是 fail-closed 所必需；已通过 contract amendment 加入精确允许路径，未扩展其他 store 行为。

## Executed Checks

- 聚焦 Vitest：8 个测试文件、41/41 PASS；覆盖分析过滤传播、设置开关、fail-closed、错误懒加载和重新进入刷新、筛选/排序/分页、列设置、详情脱敏以及三类分布图禁止管理员下钻。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- changed-file ESLint：PASS。
- `corepack.cmd pnpm --dir frontend run build`：PASS，Vite production build 共转换 1106 modules。
- Playwright 本地 mock smoke：桌面 `1440x1000` 与移动 `390x844` 均实际打开 `/usage`；统计卡、三类分布图、趋势、错误筛选、移动错误卡片、详情按钮和详情弹窗可见且可操作。
- 超长模型、分组和端点 mock：三张卡保持在移动端 `33..357px` 内容边界，页面 `clientWidth=390`、`scrollWidth=390`；桌面保持横向圆环/表格布局，端点长文本仅在表格容器内滚动。
- 浏览器控制台：0 error、0 warning。
- 用户错误 DTO/UI 字段复核：仅包含 id、时间、模型、入站端点、状态、分类、平台、消息、Key 名称/删除状态、错误响应和可选上游状态码；无 IP、User-Agent、邮箱、账户、上游地址、重试、owner/source 或 API Key 前缀。
- `git diff --check`、未合并索引、冲突标记、构建产物状态及 contract 允许路径检查：PASS。
- diff 精准性检查：变更均可追溯到 S136 用户 Usage、用户错误 UI、设置开关、响应式修复、测试或 workflow 证据；未发现无关重构或格式化。

## Unverified Risks

- 浏览器 smoke 使用会话内 API mock，不是真实后端登录态；未执行真实 `/usage/errors`、Settings 保存或 PostgreSQL 数据流。
- 未部署、未更新容器、未执行生产 migration、未 push，也未合入脏主工作树。
- Vite 仍报告既有 dynamic/static import、chunk size、Browserslist 数据陈旧和 Node child-process deprecation 警告；本 Sprint 未新增依赖或构建配置，构建本身通过。
- 管理员错误请求工作台和管理员 Usage 用户 Token 排行仍属于 S137，不在 S136 中。

## Contract Compliance

- S136 只修改批准的用户 Usage、共享图表、Settings/public-setting 类型、i18n、聚焦测试和 workflow 文档路径。
- 用户错误功能默认关闭、设置不可用时 fail closed，且只有用户主动进入已启用错误 Tab 后才请求数据。
- 用户模式保留实际消费列、隐藏账户成本，并禁止调用管理员 `user-breakdown` API。
- 未修改后端、schema、migration、管理员 Usage 页面、管理员 ops 页面、Token 排行、部署、容器或生产配置。

## Recommendation

`PASS / frontend + mocked-browser`。可进行授权范围内的 S136 本地提交，并以该提交为基线进入 S137；暂不 merge 脏主工作树、不 push、不部署。
