# Studio Bridge / 落叶AI 联调

最后更新：2026-06-28

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

## 与图片输入能力的当前关系

- 当前落叶创艺侧的图片继续编辑、参考图复用和跨工作台送电商，已经不应只按“chatgpt2api 前端自己持有本地文件”理解；Sub2API 作为 bridge 真源，最近也开始决定某些上游账号是否必须把本地图片转成对象存储 URL。
- 2026-06-21 已确认：这条能力不是按 APIMart 或某个平台名硬编码，而是按上游账号 `extra` 决定：
  - `image_input_transport=object_url`
  - `image_upload_limit_bytes`
  - `image_url_fields_supported`
- 这意味着后续再排查“落叶侧继续编辑失败 / 某些参考图请求在上游被 1MB multipart 限制拦住 / 同样功能在不同账号表现不同”时，不能只查 launch/redeem 或 session-probe，还要回到 Sub2API 账号能力配置与对象存储可达性。

## 与支付/治理面的当前关系

- 当前 Studio Bridge 主线已经不只包含 launch/redeem、余额和扣费；它还和首充福利、可配置充值套餐、注册 IP / 最近登录 IP 这些支付治理面共同组成默认后台知识。
- 换句话说，后续如果做“落叶AI 生产联调”知识补写，不应再把它拆成 bridge 一页、支付一页、用户治理一页互不相关；默认心智应是同一条用户入口 -> 充值/余额 -> 创作扣费 -> 后台治理链路。

## 与 6 月下旬上游安全小补丁的关系

- 到 2026-06-26，Sub2API 的默认续做入口已经不再只有 Studio Bridge / 支付治理；`upstream-main-v0138-followup-safe-patches-s21/s22` 也进入了当前稳定工程背景层。
- 这意味着后续再排查落叶AI生产联调时，不能把 bridge 问题和上游兼容问题完全分开看：
  - Spark `image_generation` tool strip 会直接影响图片意图识别。
  - usage cache token 明细会影响后台用量/排行类解释。
  - Responses / ChatCompletions 工具参数去重、`refresh_token_invalidated` 非重试和 transport failover 会影响 OpenAI/Codex 路径的稳定性。
- 但这条工程线当前仍明确不覆盖支付/订阅/余额预扣、前端产品页或 migrations；不要把 S21/S22 误说成 Studio Bridge 全面产品升级。

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
- 遇到某些账号的继续编辑、参考图或 mask 上传在上游失败：再加一层排查账号 `extra` 是否启用了 `image_input_transport=object_url`、是否声明支持 `image_urls / mask_url`、对象存储 presigned URL 是否可达；不要只在落叶前端或 launch 链路里找原因。

## 仍未验证的边界

- 真实支付回调、真钱充值。
- 真实上游创作成功扣费 / 失败退款。
- 团队共享额度和团队成员真实 E2E。
- 网络超时、数据库故障注入和迁移演练。
