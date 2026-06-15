# 当前任务快照

最后更新：2026-06-14 13:45 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前默认续做入口已从“上游低风险合成”切到 Studio Bridge / 落叶AI生产联调。
- 最近提交 `fe2f80be1 feat: add studio bridge integration` 已把 Sub2API 扩展为落叶AI的账号、充值、余额、配置和扣费真源。
- `chatgpt2api` 对应提交为 `47c9f72 feat: add luoye independent studio mode`，两边本地容器已重建并通过基础验证。

## 当前主线

- Studio Bridge / 落叶AI 生产联调仍是主链：
  - 配置落叶AI launch URL、充值回跳 URL、bridge internal secret、默认聊天/生图/视频分组。
  - 确认 `/chat-images` 与 `/studio-bridge/launch` 都能稳定进入落叶AI启动页。
  - 确认注册/登录完成后可带一次性 `launch_token` 回跳落叶AI并兑换本地 session。
- 支付治理已经进入同一条默认产品面：
  - 后台现在支持可配置充值套餐，不再只是假设固定充值面额或固定套餐展示。
  - 需要同时确认支付单创建、支付恢复、充值兑现和用户侧 `PaymentView` 对套餐配置的读取没有和 Studio Bridge 充值回跳语义脱节。
- 用户治理与后台画像继续前移：
  - 管理员用户列表现在已把注册/最近登录 IP 作为稳定字段的一部分。
  - 后续排查新用户福利、风控、异常充值或 OAuth 注册来源时，不应只看 user id / email，还要把 IP 画像一起视作默认排查面。
- 团队空间联调仍是稳定背景层：
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
- 2026-06-11 Studio Bridge 本地配置防丢已补齐：初始化时如果存在 `STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET`，且当前配置为空、禁用、缺 secret/group 或仍是 `example.com` 占位，会自动修复为本地默认 return URL、充值 URL 和 `127.0.0.1/localhost` allowed domains；默认生图分组不再硬编码为 `4`，会选择第一个 active 且 `allow_image_generation=true` 的 image group，聊天分组优先 text group、缺失时复用 image group；正式域名配置不会被覆盖。
- 2026-06-11 本地浏览器 smoke 确认：一次性本地用户从 `http://127.0.0.1:62080/chat-images` 成功跳转到落叶创艺 `http://127.0.0.1:8081/image`；`/api/v1/user/studio-bridge/launch`、落叶侧 `/auth/sub2api/launch` 和内部 redeem/user-summary 均 200；session-probe iframe 只请求 `/studio-bridge/session-probe`，未出现 `frame-ancestors 'none'` / CSP iframe 报错。
- 2026-06-11 上游 `v0.1.136` 小步合成继续在分支 `codex/upstream-v0136-partial-migration` 上进行；本轮仅补了两个确认缺失且低风险的点：OpenAI fallback 默认模型目录新增 `codex-auto-review`，以及公共 `Select` 下拉最大高度从 `max-h-60` 放宽到 `max-h-80`，避免 7+ 选项看似缺失。
- 2026-06-11 复核确认多条 `git cherry +` 候选其实已等价覆盖，不能重复迁：账号 credentials 脱敏与 `credentials_status`、账号重授权保留 Extra、OAuth 401 不覆盖 credentials JSONB、OpenAI compatible versioned base URL、API Key SSE body fallback、OpenAI images `n` 透传/审核错误透传、setup 完成后拦截、API key name XSS escape、content moderation 不自动封禁 admin、OpenAI fast policy 默认透传、Codex `call_` -> `fc_` 修复、Groups supported model scopes、EditAccountModal 旧凭证回退、i18n `@` 转义、Responses 路由选项、Ops deep link 初始化顺序等。
- 2026-06-12 上游 `v0.1.136` 小步合成继续：本轮补 `bf1a2d6 Align Codex usage stats with reset windows` 的等价修复，Codex 5h/7d `WindowStats` 查询起点改为优先使用上游 reset window 的 `ResetsAt - window`，缺失或过期 reset 时仍回退到原来的滚动 `now - window`；新增 `TestCodexWindowStatsStart` 覆盖 active reset、expired reset 和 nil progress。
- 2026-06-12 复核确认以下中小候选也已等价覆盖，暂不重复迁：长上下文 cache_read/cache_creation 倍率、Antigravity streaming `message_start.message.usage` 输入 token 采集、OpenAI WS terminal event 不计入 first-token、Gemini messages streaming tool_use->text block closure、count_tokens generation 字段过滤、Claude Code count_tokens UA-only 放行、Ops metrics 排除 count_tokens、WS usage dedup、并发 acquire 失败分类。
- 2026-06-13 管理员用户视图继续补齐用户治理字段：`users` 表和后台用户列表已把注册 IP / 最近登录 IP 纳入稳定数据面；后续排查注册来源、首充领取、异常登录或风控限制时，应默认把 IP 信息当作后台一等字段，而不是一次性调试信息。
- 2026-06-14 支付后台已支持可配置充值套餐：充值页、后台设置、支付恢复和兑现逻辑不再只服务固定金额流，而是围绕后台套餐配置运转；这说明“支付闭环”现在已从首充福利 bonus 继续推进到“套餐定义 + 支付回跳 + 恢复/兑现”的完整后台治理面。

## 下一步

- 正式域名上线前，先在后台填好 bridge internal secret、落叶AI launch URL、充值回跳 URL 和默认分组。
- 补一轮后台支付治理验收：确认可配置充值套餐、首充福利 bonus、支付恢复和用户侧支付页文案在同一套配置下工作一致，不要只验支付成功回调。
- 如继续排查用户异常或风控反馈，优先从管理员用户列表核对注册 IP / 最近登录 IP，再结合福利领取、OAuth 注册和充值记录判断，不要只看账户基础字段。
- 用真实账号验证注册/登录、充值、落叶AI创作扣费、使用记录、团队空间最小闭环。
- 本地重启后若再遇到 `STUDIO_BRIDGE_DISABLED`，优先检查 env secret 是否注入、active image group 是否存在且允许生图、以及设置是否被正式域名配置覆盖；不要再手动硬编码默认 group `4`。
- 如继续上游合成，必须单独开 Sprint，避免覆盖 Studio Bridge 和落叶AI入口；下一轮优先只看仍未确认的中小稳定修复，暂不孤立迁 `service_quota_enabled` 这类依赖功能链的尾巴，也不碰 `0acf00c4 Add admin compliance acknowledgement gate` 这种产品行为级门禁。
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
- 2026-06-11 Studio Bridge 本地配置防丢验证：
  - `cd F:/mcplugins/sub2api/backend && go test ./internal/service ./internal/server` 通过。
  - `git diff --check` 通过。
  - 本地 `sub2api:local` 容器已更新并 healthy，`http://127.0.0.1:62080/health` 返回正常。
  - `HEAD http://127.0.0.1:62080/studio-bridge/session-probe?...parent_origin=http://127.0.0.1:8081` 返回 200，CSP 含 `frame-ancestors http://127.0.0.1:8081`。
  - 浏览器 smoke 从 `62080/chat-images` 成功跳到 `8081/image`；网络记录显示 `POST /api/v1/user/studio-bridge/launch`、落叶侧 `/auth/sub2api/launch`、`/api/v1/internal/studio-bridge/redeem` 和 `user-summary` 均 200；性能记录中 62080 资源只有 `/studio-bridge/session-probe?...`，没有根路径 iframe 请求。
- 2026-06-11 上游 `v0.1.136` 小步合成验证：
  - `cd F:/mcplugins/sub2api/backend && go test ./internal/pkg/openai` 通过。
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run test:run -- src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/admin/__tests__/groupsSupportedModelScopes.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts` 通过，24 tests；仅有既有 Browserslist 数据过期提示。
  - `cd F:/mcplugins/sub2api/frontend && npm.cmd run typecheck -- --pretty false` 通过。
  - `git diff --check` 通过，仅 knowledge 文件 LF/CRLF 工作区提示。
- 2026-06-12 上游 `v0.1.136` 小步合成验证：
  - `cd F:/mcplugins/sub2api/backend && go test ./internal/service -run "TestCodexWindowStatsStart|TestBuildCodexUsageProgressFromExtra|TestResolveOpenAICodex7dTempBlock" -count=1` 通过。
  - 2026-06-13~2026-06-14 的支付/用户治理扩展尚未在本页补新的真实浏览器或联调证据；当前先以提交主题和测试落点确认它们已进入默认知识面，后续如继续推进应优先补套餐配置、支付恢复和用户 IP 字段的定向验收记录。
