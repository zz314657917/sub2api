# 当前任务快照

最后更新：2026-06-10 04:18 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前默认续做入口已从“上游低风险合成”切到 Studio Bridge / 落叶AI生产联调。
- 最近提交 `fe2f80be1 feat: add studio bridge integration` 已把 Sub2API 扩展为落叶AI的账号、充值、余额、配置和扣费真源。
- `chatgpt2api` 对应提交为 `47c9f72 feat: add luoye independent studio mode`，两边本地容器已重建并通过基础验证。

## 当前主线

- Studio Bridge 生产配置：
  - 配置落叶AI launch URL、充值回跳 URL、bridge internal secret、默认聊天/生图/视频分组。
  - 确认 `/chat-images` 与 `/studio-bridge/launch` 都能稳定进入落叶AI启动页。
  - 确认注册/登录完成后可带一次性 `launch_token` 回跳落叶AI并兑换本地 session。
- 内部接口生产联调：
  - 余额/充值摘要、使用记录摘要。
  - `reserve / commit / refund` 幂等扣费。
  - 余额不足、任务失败、任务取消时落叶AI能给出明确提示和退款/释放预扣。
- 团队空间联调：
  - 团队空间由落叶AI记录 actor/payer。
  - Sub2API 侧仍作为扣费和余额真源。

## 已稳定事实

- OpenWebUI 不再作为当前用户侧“聊天生图”入口；用户侧入口应进入落叶AI启动链路。
- `/studio-bridge/launch` 是 `/chat-images` 的 alias，避免注册/登录 redirect 到 404。
- 充值、用户、余额、默认分组和内部通信配置仍应由 Sub2API 管理后台维护。
- Studio Bridge `reserve / commit / refund` 已从 Redis 财务状态改为数据库账本表 `studio_bridge_charges`，以 `(app_id, charge_key)` 做唯一幂等键；重复 reserve/commit/refund 不重复扣退，fingerprint 冲突拒绝。
- partial refund 后 commit 只写净消费 usage log；commit 后不允许再 refund 原 reserved 单，避免已确认消费被回退。
- 之前的上游低风险合成仍是背景候选，但不再是最前主线；后续合上游时仍需保护模型市场、APIMart 计费、工单、Canvas、Chat/Image Studio 和 Studio Bridge 定制。
- 2026-06-10 本地验收确认：落叶创艺通过 Studio Bridge launch/redeem 能进入 `/image`；内部扣费接口的 reserve/commit/refund 幂等成立；余额不足返回明确错误；普通用户直打落叶创艺协议 API 会被独立模式 403 拦截，不能绕过创作任务扣费。
- 2026-06-10 排查确认：Sub2API 用户侧 `/usage` “统计和分页有数据但表格空白”不是扣费没入账，数据库 `usage_logs` 和 `/api/v1/usage` 都有记录；根因是 `UsageView.vue` 对 Studio Bridge 记录中的 `duration_ms = null` 等字段直接 `.toFixed()` / `.toLocaleString()`，导致 `DataTable` 渲染中断。
- 2026-06-10 已修复用户使用记录展示：`UsageView.vue` 对金额、Token、耗时、CSV 导出和详情弹层做 null-safe 格式化；`UsageLog.duration_ms` 前端类型改为 `number | null`；新增回归测试覆盖 `duration_ms: null` 的 Studio Bridge 行。
- 2026-06-10 为了重包容器，顺手补齐 `backend/internal/service/api_key_service.go` 中 `validateAPIKeyRouteGroups(..., false)` 的遗漏调用参数；该点属于此前默认 API Key / 默认分组改造的编译断点，保持普通更新路径继续校验分组权限。

## 下一步

- 正式域名上线前，先在后台填好 bridge internal secret、落叶AI launch URL、充值回跳 URL 和默认分组。
- 用真实账号验证注册/登录、充值、落叶AI创作扣费、使用记录、团队空间最小闭环。
- 如继续上游合成，必须单独开 Sprint，避免覆盖 Studio Bridge 和落叶AI入口。
- 真实支付回调、真钱充值、真实上游创作成功/失败扣退、网络超时/DB 故障注入和迁移演练仍需 staging 或生产环境验证；本地验收不触碰真钱支付。
- 若用户浏览器仍看到旧的使用记录空表，先强刷 `/usage`；当前本地容器已更新到新前端资源，正常无需清数据库或改扣费记录。

## 证据入口

- Sub2API 提交：`fe2f80be1 feat: add studio bridge integration`
- 落叶AI提交：`47c9f72 feat: add luoye independent studio mode`
- Sub2API 验证：
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run test:run -- public-smoke`
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run build`
  - `cd F:/mcplugins/sub2api/backend && go test ./...`
  - `cd F:/mcplugins/sub2api/backend && go test -tags=integration ./internal/repository -run "TestStudioBridgeRepository" -count=1`
  - HTTP `/health` 与 `/studio-bridge/launch` 检查通过
- chatgpt2api 验证：
  - `cd F:/java/chatgpt2api/web && npm.cmd run lint`
  - `cd F:/java/chatgpt2api/web && npm.cmd run build`
  - `cd F:/java/chatgpt2api && go test ./...`
- 2026-06-10 02:38 复核：
  - `cd F:/mcplugins/sub2api/backend && go test ./...` 通过。
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run test:run -- public-smoke` 通过。
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run build` 通过，仅有既有 Vite chunk、Browserslist 和 Node deprecation 警告。
  - `git diff --check` 无 whitespace 错误，仅 LF/CRLF 工作区提示。
  - 本地 `sub2api` 容器 healthy，`http://127.0.0.1:62080/` 返回 200；最近日志未见真实 panic/fatal/500，`apiError` 命中是静态资源文件名。
- 2026-06-10 04:18 使用记录修复复核：
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run test:run -- src/views/user/__tests__/UsageView.spec.ts` 通过，12 tests。
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run build` 通过，仅有既有 Vite chunk、Browserslist 和 Node deprecation 警告。
  - `cd F:/mcplugins/sub2api/backend && go test ./internal/service -run TestAPIKey -count=1` 通过。
  - `git diff --check -- frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/UsageView.spec.ts frontend/src/types/index.ts backend/internal/service/api_key_service.go` 通过。
  - 已重新编译 Linux amd64 embed 二进制、重包 `sub2api:local`，新镜像 ID `98794421cc65`；`docker compose ... up -d --no-deps --force-recreate sub2api` 后容器 healthy，`http://127.0.0.1:62080/health` 返回 `{"status":"ok"}`。
  - 数据库 `usage_logs` 当前 `COUNT=158`、`MAX(id)=215`；Playwright 打开 `http://127.0.0.1:62080/usage` 检查到 `tableCount=1`、`rowCount=8`，页面包含 `gpt-5.5` 和 `gpt-image-2`，浏览器控制台 0 errors。
  - 截图证据：`F:/mcplugins/sub2api/output/playwright/sub2api-usage-records-restored.png`。
