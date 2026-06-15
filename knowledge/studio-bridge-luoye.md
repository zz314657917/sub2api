# Studio Bridge / 落叶AI 联调

最后更新：2026-06-11

## 当前定位

- `sub2api` 当前不只是网关；在 Studio Bridge / 落叶AI 主线下，它还是账号、余额、充值、默认分组、bridge internal secret 和扣费真源。
- 当前用户侧入口默认走 `/chat-images` 或 `/studio-bridge/launch`，而不是旧 OpenWebUI 入口。
- `chatgpt2api` / 落叶创艺负责创作台 UI 和任务体验，但 actor/payer、余额状态和真实扣费语义仍要回到 Sub2API 侧理解。

## 稳定入口链路

1. 用户访问 `/chat-images`
2. 前端调用 `/api/v1/user/studio-bridge/launch`
3. Sub2API 生成一次性 `launch_token`
4. 浏览器跳到落叶创艺 `/auth/sub2api/launch`
5. 落叶创艺调用 Sub2API 内部 `redeem`
6. 落叶创艺建立本地 session，并继续拉 `user-summary` 等 bridge 数据

稳定事实：

- `/studio-bridge/launch` 是 `/chat-images` 的 alias，避免注册/登录 redirect 到 404。
- `session-probe` 已进入默认链路：iframe 只应请求 `/studio-bridge/session-probe`，不应再尝试根路径 iframe。
- 本地 smoke 已验证 `62080 -> 8081` 跳转、launch/redeem/user-summary 200 和 CSP `frame-ancestors` 放通。

## 配置真源与默认值

- 充值回跳 URL、落叶AI launch URL、bridge internal secret、默认聊天/生图/视频分组都由 Sub2API 管理后台维护。
- 如果存在 `STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET`，且当前配置为空、禁用、缺 secret/group 或仍是 `example.com` 占位，初始化会自动修复成本地 bridge 默认配置。
- 本地自动修复只面向空配置、禁用、缺 secret/group 或占位配置；正式域名配置不会被覆盖。
- 自动修复后的 allowed domains 会放开 `127.0.0.1` / `localhost` 本地联调入口。

## 默认分组语义

- 默认生图分组不再硬编码为 `4`。
- 当前规则是优先选择第一个 active 且 `allow_image_generation=true` 的 image group。
- 默认聊天分组优先 text group；如果没有可用 text group，会复用 image group。
- 如果没有可用 image group，不应强行启用 bridge；应保留 `STUDIO_BRIDGE_GROUP_REQUIRED` 这类显式错误。

## 扣费与账本语义

- `reserve / commit / refund` 已不是临时接口；它们是默认联调面。
- Studio Bridge 账本已落到数据库表 `studio_bridge_charges`，唯一幂等键为 `(app_id, charge_key)`。
- 重复 reserve / commit / refund 不应重复扣退；fingerprint 冲突要拒绝。
- partial refund 后，commit 只写净消费 usage log。
- commit 后不允许再 refund 原 reserved 单，避免已确认消费被回退。

## 使用记录与前端展示

- `/usage` 统计和分页有数据但表格空白，不等于扣费没入账。
- 2026-06-10 已确认：数据库 `usage_logs` 和 `/api/v1/usage` 有记录时，前端仍可能因 `duration_ms = null`、金额或 token 字段空值格式化而让 `DataTable` 渲染中断。
- 当前稳定修复是对金额、Token、耗时、CSV 导出和详情弹层做 null-safe 格式化，并允许 `UsageLog.duration_ms` 为 `number | null`。

## 最小验证清单

- 后端：
  - `cd backend && go test ./...`
  - `cd backend && go test -tags=integration ./internal/repository -run "TestStudioBridgeRepository" -count=1`
- 前端：
  - `cd frontend && npm.cmd run test:run -- public-smoke`
  - `cd frontend && npm.cmd run build`
- 定向回归：
  - `cd frontend && npm.cmd run test:run -- src/views/user/__tests__/UsageView.spec.ts`
  - `cd backend && go test ./internal/service ./internal/server`
- 最小本地 smoke：
  - `HEAD /studio-bridge/session-probe?...parent_origin=http://127.0.0.1:8081` 返回 200，且 CSP 放开 8081
  - 浏览器从 `http://127.0.0.1:62080/chat-images` 成功进入 `http://127.0.0.1:8081/image`
  - 网络记录确认 `launch`、`redeem`、`user-summary` 200

## 现在优先排查什么

- 遇到 `STUDIO_BRIDGE_DISABLED`：先查 env secret 是否注入、配置是否仍是占位值、active image group 是否存在且允许生图、正式域名配置是否覆盖了本地默认值。
- 遇到余额不足、任务失败或取消：先从 `reserve / commit / refund` 和 usage log 语义排查，不要先怀疑纯前端展示。
- 遇到落叶创艺团队空间问题：不要只在 `chatgpt2api` 查；`actor/payer`、余额和扣费真源仍在 Sub2API。
- 遇到浏览器 launch 成功但页面打不开：先查 `session-probe`、CSP `frame-ancestors`、allowed domains 和 launch / redeem / user-summary 调用链。

## 仍未验证的边界

- 真实支付回调、真钱充值。
- 真实上游创作成功扣费 / 失败退款。
- 团队共享额度和团队成员真实 E2E。
- 网络超时、数据库故障注入和迁移演练。
