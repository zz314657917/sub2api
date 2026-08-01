# 当前任务快照

最后更新：2026-08-01 13:07 +08:00

## 背景

- S135 后端基础已提交为 `d48370f75`，S136 在隔离 worktree `E:/codex-worktrees/sub2api/usage-user-frontend-s136` 完成用户 Usage 前端。
- 主工作树存在重叠用户改动，本轮禁止直接 merge、stash 或回滚主工作树。

## 当前目标

- 收口 `usage-user-frontend-s136` 本地提交，并以该提交为基线进入 S137。
- S137 处理管理员错误请求工作台和管理员 Usage 用户 Token 排行。

## 本次已完成

- `/usage` 已加入统计卡、Token 趋势、模型/分组/端点分布，统一传播日期、API Key、分组、模型、请求类型、计费模式、时区和日/小时粒度。
- 已加入默认关闭的用户错误请求 Tab、懒加载、筛选、排序、分页、列设置、移动卡片、详情按钮和严格脱敏详情。
- Settings 已接入 `allow_user_view_error_requests`；设置缺失或运行时关闭均 fail closed。
- 已修复移动端分布图挤压、缓存 tooltip/长分组横向溢出，以及错误 Tab 重新进入时的旧数据问题。

## 已确认事实

- 用户错误 DTO/UI 不含 IP、User-Agent、邮箱、账户、上游地址、重试、owner/source 或 API Key 前缀。
- 用户模式不调用管理员 `user-breakdown` API；保留实际消费，隐藏账户成本。
- Playwright mock 下 `390x844` 和 `1440x1000` 页面宽度、布局和交互通过，控制台 0 error / 0 warning。

## 待验证点

- 真实登录后端 smoke -> 验证：实际读取 `/usage`、切换错误 Tab、打开详情并由管理员保存开关。
- 部署/容器/生产 migration -> 验证：仅在后续明确授权后执行并记录运行态证据。
- S137 管理员界面 -> 验证：独立 contract、聚焦 Vitest、typecheck/build 和管理员桌面/移动 browser smoke。

## 当前结论

- S136 最终裁决为 `PASS / frontend + mocked-browser`，详见 `docs/workflow/qa-reports/usage-user-frontend-s136-qa.md`。
- 可做授权范围内的本地提交；不得 merge 脏主工作树、push、部署或更新容器。

## 下一步

1. 本地提交 S136 -> 验证：提交只含 contract 允许路径，提交后工作树干净。
2. 从 S136 提交创建隔离 S137 worktree -> 验证：基线包含 S135/S136 且不触碰主工作树。
3. 起草并审核 S137 contract -> 验证：管理员错误工作台和 Token 排行的 API、权限、响应式与验收边界明确。

## 验证记录

- 聚焦 Vitest：8 文件、41/41 PASS。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- changed-file ESLint：PASS。
- `corepack.cmd pnpm --dir frontend run build`：PASS，1106 modules。
- Playwright mock：移动/桌面、超长文本、错误卡片/详情及控制台检查 PASS。
- `git diff --check`、冲突标记、未合并索引、敏感字段和允许路径检查：PASS。
