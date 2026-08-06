# 当前任务快照

最后更新：2026-08-06 +08:00

## 背景

- Pixel Cafe 已按独立 Sprint 交付：S143 页面/开关，S144 schema 与 migration 201，S145 管理 Room API，S146 管理工作台，S147 用户端只读发现，S148 用户下单与 Seat 锁定，S149 paid-full 激活/disabled 受管 Key/strict Binding，S150 固定账号 auth-cache/gateway 路由。
- 主工作树有多轮未提交改动。各 Sprint 仅按自己的 contract allowlist 裁决；不要把全量 `git status`、Ent/schema/migration 201 或已有 generated changes 归因给 S150。
- S157/S158 已补齐隔离 PostgreSQL/Redis 与最后一 Seat 运行态证据；Docker Desktop 曾在明确授权后重启，原有容器已恢复健康。

## 当前目标

- v0.1.171 选择性上游整合已收口到本地 `main`：S181-S201 分为三组行为提交
  `21c2d33d4`、`20c56753a`、`290a815ba` 和一组 regression/triage 提交 `194edd3f7`；此前 S202
  媒体工具输出桥接保持在 `d6792b966`。本轮无推送、部署、数据库或生产操作，本地主线仍未推送到
  `origin/main`。
- 四个 `codex/upstream-v0171-*` 临时分支和全部 Git worktree 注册已删除，`backup/*` 保留。S182 的目录已删除；
  `E:/codex-worktrees/sub2api/upstream-v0171-financial-integrity-s181` 与
  `E:/codex-worktrees/sub2api/upstream-v0171-integration-s183` 是系统策略拒绝递归删除后留下的非 Git
  残留目录，不能绕过策略强删。主工作树仅保留用户的未跟踪 `outputs/` JSON 文件。

- S177 本地分支整合已收口并推送到 `origin/main@09c9971e7`。在既有合入基础上，本轮按独立合同快进了 classifier 回归
  `c92dcc13d`、发布安全资源 `618cc3bf9` 与 OpenAI 代理流熔断 `199be5cba`。S132、Passkey、Kimi K3、
  Model Plaza、Codex manifest、SMTP 与其余已覆盖行为均不重复合入；远端追溯引用和 `backup/*` 保留。
  基线测试夹具修正已作为 `09c9971e7` 追加：安全审计 AST consumer 名称、异步图片 uploader 与注册风控显式 `RemoteAddr`。
  未执行远端删除、数据库、部署或生产操作。QA：
  `docs/workflow/qa-reports/upstream-v0169-classifier-regression-integration-qa.md`、
  `docs/workflow/qa-reports/upstream-v0169-release-security-integration-qa.md`、
  `docs/workflow/qa-reports/upstream-v0169-proxy-stream-circuit-integration-qa.md`。

- S176 已完成源代码实现、聚焦测试和受保护的本地容器更新：用户页移除“今日使用用户”卡片及独立轮询；管理员可配置标题、说明和整个标题区显示；默认值保持原页面。合同：`docs/workflow/tasks/pixel-cafe-phase30-presentation-settings-s176.md`。本机 `sub2api` 已更新为 `sub2api:codex-20260805-main-36e35a7bb`，在 `127.0.0.1:62080` healthy；旧镜像保留为 `sub2api:rollback-before-codex-20260805-main`。
- S176 已补齐登录态桌面浏览器证据并关闭为 `PASS / browser-local`：截图证明“今日使用用户”卡片已移除且 Room 区域保留；证据在 `output/playwright/pixel-cafe-s176/group-buy-desktop.png`。QA：`docs/workflow/qa-reports/pixel-cafe-phase30-presentation-settings-s176-qa.md`。

- S180 排行榜展示实现已完成：动态“你的战绩”移入左侧新生成的无文字长条横幅，右侧疯狂星期四图片和独立战绩卡已删除，右栏只保留完整 Top 10/奖励信息；文案“等待开奖，下一名距离你”收紧为“领先下一名”。聚焦 `38/38` 测试、typecheck、build 与 Git integrity 均通过。
- S180 当前为 `BLOCKED / browser-resource`：本地 synthetic `/leaderboard` 已打开，但在完成桌面/移动端 DOM 几何断言和截图前，用户报告机器 CPU 饱和。S180 的 `lb-s180` 与 `62081` 进程已关闭；高占用来源是另一个 `beads-task026-browserqa` 会话，未越权关闭。QA：`docs/workflow/qa-reports/leaderboard-record-banner-s180-qa.md`。

## 已确认事实

- `main` 已包含 S202 及 S181-S201 选择性整合，而 `origin/main` 仍是 `d6f05667b`；本轮 `git diff --check`、`git fsck --no-dangling`、聚焦 Go 回归、前端 62 个 Vitest、typecheck 与 production build 均通过。主工作树只保留未跟踪的 `outputs/` JSON 文件。
- 本地分支仅剩 `main` 与三个 `backup/*`。三个临时集成 worktree 已删除；原始宽分支已删除，其 4 个未提交 Prompt Audit 文件保存在命名 `stash@{0}`。
- Windows 本机策略拒绝递归删除 `E:/codex-worktrees/sub2api-v0169-behavior-wide` 的 993 项脱离 Git 的残留副本；它不在 `git worktree list`，且不含 `.git`。不得绕过该策略强删。

## 待验证点

- S180 仍需在 CPU 异常解除后只启动一个最小 Playwright 会话，完成 `1920x1080` 与 `390x844` 的横向溢出、文字/图片碰撞、Top 10 行数和截图验收。
- 如需恢复已归档 Prompt Audit WIP，先审阅 `stash@{0}`，再建立新的独立合同、从当前 `main` 重建并运行验收；不得直接恢复到主工作树。

## 当前结论

- 分支整理已完成：所有可证明覆盖或已验证独立切片已本地合入，临时分支/worktree 已清理，主线无冲突索引。仅剩 Windows 策略阻断的非 Git 残留目录和可恢复的命名 stash。

- S175 contract approved：修复 S174 运行中明确观测到的通用 best-effort usage batch-state `input_idx` JSON 类型错误。现有同步 fallback 已保证 S174 durable Cafe usage attribution，但批处理路径每次触发 warning/降级，需要以 repository SQL 类型最小修复和 fresh S174 rerun 证明直接 batch state decode 恢复；不得修改 usage 业务字段、schema/migration、gateway/provider、支付或共享资源。合同：`docs/workflow/tasks/pixel-cafe-phase29-usage-batch-state-s175.md`。
- S175 已关闭为 `PASS / runtime-isolated`：`backend/internal/repository/usage_log_repo.go` 仅在 state query 的 synthetic `input_idx` 参数增加 `::integer`；新增 `usage_log_repo_best_effort_state_test.go` 锁定 state/non-state 查询边界。聚焦 repository、verbose fresh PostgreSQL Gateway（单次及三次）、相邻 Gateway/Cafe 路由、gofmt、`git diff --check` 和未合并索引均通过；所有 fresh Gateway 日志不再出现 batch-state decode warning。QA：`docs/workflow/qa-reports/pixel-cafe-phase29-usage-batch-state-s175-qa.md`。

- S174 已关闭为 `PASS / runtime-isolated`：fresh PostgreSQL、真实 `RegisterGatewayRoutes`、API-key middleware、`GatewayHandler`/`GatewayService`、`repository.NewHTTPUpstream` 与本进程 loopback Anthropic terminal 证明有效 managed Key 的实际上游认证和 durable `usage_logs.account_id` 均等于 Room Binding 固定 Account。Binding 到期及 Account disabled 在转发/usage 写入前 `CAFE_ACCOUNT_UNAVAILABLE` fail-closed。单次与连续三次 fresh rerun 均通过；QA：`docs/workflow/qa-reports/pixel-cafe-phase28-gateway-usage-s174-qa.md`。
- S174 的 disposable Ent PostgreSQL 仅补齐当前 native UsageLogRepository 需要的 migration-only usage columns、daily rollup table 与幂等索引，不修改真实 schema/migration。每次运行均观察到通用 best-effort batch state `input_idx` JSON decode 警告；同步 fallback 仍成功写入，并由 SQL account attribution 断言证明。该效率/观测问题需独立 product-source Sprint，不能将 S174 描述为完全消除了通用 usage 批处理风险。
- S174 未验证真实 provider、merchant/payment sandbox、容量、staging/canary、部署、回滚或生产；不能把 loopback 结果表述为外部 provider 或生产 readiness。

- S173 已关闭为 `PASS / runtime-isolated`：fresh PostgreSQL/Redis 以临时 JWT/用户/Key/Binding/Account 真实装配 Cafe My Rooms user routes、JWT middleware、API-key middleware 与 gateway-auth preflight。两用户 My Rooms 隔离、匿名拒绝、fixed Account pin、B L1 stale 命中、A 删除 Redis L2 并经 Pub/Sub 失效 B、inactive 拒绝、valid re-enable 恢复和 Binding 到期 fail-closed 均通过；核心 smoke 连续三次 fresh rerun 通过。preflight terminal 只返回本地 `204`，没有 provider 调用、现有 Key 或共享数据库/Redis 访问。QA：`docs/workflow/qa-reports/pixel-cafe-phase27-jwt-gateway-redis-s173-qa.md`。
- S173 未验证完整 provider handler/出站流量、payment sandbox、usage 写入、性能、staging、部署、回滚或生产；不能将其表述为真实 gateway/provider 或生产 readiness。

- S172 已关闭为 `PASS / runtime-isolated`：满员激活与可重复激活最终生成 `active` managed Key；用户可在受保护边界内修改名称/IP 并切换 `active/inactive`，legacy `disabled -> active` 仅在完整有效权益事实下允许。fresh Testcontainers PostgreSQL 的 10 个场景覆盖成功、终态/所有权/保护字段拒绝、Round/Seat/Binding/到期/Account/Group fail-closed、无部分写入和 20 次并发启用收敛；S159 payment callback 最终状态已同步为 active。
- S172 已完成 service/repository 实现、L1/L2/Pub/Sub cache-call 回归、activation/payment/repository/middleware/pinned gateway/server compile 和 Git integrity 检查。QA：`docs/workflow/qa-reports/pixel-cafe-phase26-managed-key-state-s172-qa.md`。
- S172 未验证真实 Redis 网络传播、真实 JWT/gateway/provider/payment sandbox、性能、部署、回滚或生产；没有启用、读取或写入本机现有 Cafe managed Key。

- S171 已关闭为 `PASS / runtime-isolated`：fresh Testcontainers PostgreSQL 同时释放 100 个不同用户的最后一 Seat 请求，得到恰好 1 个 winner/99 个 `CAFE_SEAT_UNAVAILABLE`；100 次 paid-full activation retry 全部成功收敛为一个四 Seat active Round、四个 disabled managed Key、四个 active strict Binding 和一个 activation event。实际数据库同时在途调用限制为 24；核心测试单轮及连续 3 轮均通过，未改产品代码或现有 S158/S159/S170。

- S170 已关闭为 `PASS / runtime-isolated`：fresh Testcontainers PostgreSQL 经实际 Cafe 注册路由/JWT、UserRepository、SQL-backed idempotency、PaymentConfig/DefaultLoadBalancer/PaymentService/GroupBuy/Cafe order service 验证一笔 pending PaymentOrder、locked Seat/reservation、`shares_locked` 事件和 `ORDER_CREATED` 审计；同一 `Idempotency-Key` 的重放携带 `X-Idempotency-Replayed: true` 且不重复持久化事实。临时 EasyPay `popup` 只用合成 `*.invalid` 配置构造跳转 URL；缺少必填配置的独立夹具使 Order `failed`、Seat `released`、reservation 清零，未调用外部 HTTP、真实商户、共享容器或启用 Cafe Key。

- S169 已关闭为 `PASS / runtime-isolated`：fresh Testcontainers PostgreSQL、实际 `RegisterUserRoutes`、JWT middleware、AuthService、UserRepository 与 Cafe public service 证明两个隔离用户只看到各自 My Rooms membership，公共 Room/overview 不泄露夹具中的 private metadata；匿名先被 JWT 拦截，feature-disabled 返回 `CAFE_DISABLED`，付款前未勾协议返回 `CAFE_AGREEMENT_REQUIRED` 且 PaymentOrder/Seat 无写入。为维持真实 UserRepository，临时数据库补齐既有迁移中的 `user_avatars` 表形状；未改 migration 或共享数据库。未证明成功订单创建、支付/provider、Cafe Key 激活、Redis、性能、部署或生产。

- S168 已关闭为 `PASS / mocked-browser`：隔离浏览器仅用本地 `/api/v1/**` 合成 fixture 和合成 auth cache 驱动实际 Vue/Router `/group-buy`。功能开启时页面正确进入 Pixel Cafe；切换 Claude 区、键盘选 Room、选空 Seat、勾选协议后构造本地订单请求并进入等待支付。`390x844` 无横向溢出，匿名访问重定向登录，关闭 `pixel_cafe_enabled` 时回退旧 Token 拼团。该结果仅证明浏览器 UI；不证明真实 JWT/API、订单持久化/Seat 锁定、支付/provider、Key 激活/路由、数据库/Redis、性能、部署或生产可用性。

- S167 已关闭为 `PASS / handler-isolated`：两临时加密 Alipay 实例使用不同生成 RSA key pair/synthetic app ID。真实 Gin POST route、smartwalle RSA2 verifier 与 PaymentService 证明第二个绑定实例有效回调仅一次完成 Cafe GroupBuy，完整 form-body 回放不重复分发或审计，第一实例 key 对第二实例订单返回 `400 verify failed` 且不改变 Order/audit/dispatch。Alipay provider 的已有测试经 `-tags=unit` 实际执行；该 Sprint 未修改 webhook、SDK/provider、支付路由、Payment/GroupBuy/Cafe 生产状态机。

- S166 已关闭为 `PASS / handler-isolated`：两临时加密 EasyPay 实例使用不同 synthetic `pid`/`pkey`。真实 Gin GET route、MD5 verifier 与 PaymentService 证明第二个绑定实例有效回调仅一次完成 Cafe GroupBuy，完整 query 回放不重复分发或审计，第一实例 key 对第二实例订单返回 `400 verify failed` 且不改变 Order/audit/dispatch。该 Sprint 不访问 EasyPay/商户、不使用真实凭据、不启用 Cafe Key、不改 schema/routes/deployment/共享容器；S159 保持下游激活证据。
- S166 明确未实现带空格 `out_trade_no` 的兼容：仅在 handler 裁剪会让 provider notification 到 PaymentService 时仍携带原始 OrderID，属于跨 handler/provider/service 的独立合同，不能用半修复掩盖。

- S165 已关闭为 `PASS / handler-isolated`：两临时加密 WeChat Pay 实例使用不同生成 RSA key pair/public-key ID/API v3 key。真实 Gin route 与 SDK verifier 证明第一排序候选拒绝、第二候选验签并 AES-GCM 解密 `TRANSACTION.SUCCESS` 后仅一次完成 Cafe GroupBuy；同 body/header 回放不重复分发或审计。第一实例 RSA 签名配第二实例密文返回 `400 verify failed`，Order/audit/dispatch 均不变。该 Sprint 不访问 WeChat API/商户、不使用真实证书或凭据、不启用 Cafe Key、不改 schema/routes/deployment/共享容器；S159 保持下游激活证据。

- S164 已关闭为 `PASS / handler-isolated`：真实 Gin Airwallex webhook 和两临时加密实例证明 `merchant_order_id` 只做订单绑定候选选择；正确第二实例 HMAC 签名成功且仅一次 GroupBuy dispatch，完全相同 body/header 回放不重复处理，第一实例签名对第二实例订单 fail-closed。该 Sprint 不访问 Airwallex API/商户、不启用 Cafe Key、不改 schema/routes/deployment/共享容器；S159 保持下游激活证据。

- S153 已关闭为 `PASS / source-level`：Cafe 专属 lifecycle 对未满到期 Round 原子关闭并把 paid Seat 交给既有幂等退款状态机；Cafe provider pending refund 独立对账；`activating` 与 paid-full `open` Round 都会重试既有激活服务。
- S152 已关闭为 `PASS / source-level`：到期的 active Cafe Round 会事务性收回 Binding、managed Key 与 Seat 并完成 Round，成功提交后失效 auth cache；关联不一致 Round 回滚且不阻塞其他 Room。Cafe Seat 已从普通汇总权益刷新中排除。
- S152 不启用 Cafe Key，也不处理退款、未满 Round timeout、激活补偿、账号迁移、自动下一轮、Lobby/Pixi 或生产运行态。
- S159 已关闭为 `PASS / runtime-isolated`：错误金额通知保持 `pending/locked/reserved`；合法通知经真实 Payment -> GroupBuy -> Cafe activation 链完成订单、Seat、Round、managed Key 与 strict Binding；S172 已把该隔离回归的最终 Key 状态更新并复测为 `active`，精确重放不增加审计、事件或授权事实。
- S160 已关闭为 `PASS / runtime-isolated`：普通 GroupBuy provider-pending reconciliation 会隔离 Cafe refund；Cafe lifecycle 经真实 Payment query-finalizer 保持 provider `pending` 可重试，后续 `success` 仅完成一次 Order/Seat/refund，终态重放无额外 query/event。
- S161 已关闭为 `PASS / runtime-isolated`：真实 Cafe lifecycle 初始 provider-refund 路径在隔离 PostgreSQL 中经 `processSeatRefund -> PrepareRefund -> ExecuteRefund` 提交一次本地 fake `pending`；pending 重放不二次提交，后续 Cafe-only query `success` 只完成一次 Order/Seat/refund/event。
- S162 已关闭为 `PASS / handler-isolated`：真实 Gin Stripe webhook 路径把原始 body 和规范化 signature header 交给本地 verifier；有效回调经真实 PaymentService 完成 Order 并只分发一次 GroupBuy，回放不重复分发/审计，伪造签名在变更前返回 400。
- S163 已关闭为 `PASS / handler-isolated`：Stripe webhook 现在在 provider lookup 前解析创建 PaymentIntent 时写入的 `metadata.orderId`。两个临时加密 Stripe 实例的本地 SDK 签名证明 order-bound 第二实例成功、第一实例签名被拒绝，回放不重复 GroupBuy 分发或审计。

## S152 审计结论

- 现有 `GroupBuyLifecycleService` 已由 `wire.go` 每 60 秒启动，普通 `ExpireRounds`、`AdminCloseRound`、`AdminProcessRefunds` 对 Cafe Round 都返回 `CAFE_ROUND_LIFECYCLE_DEFERRED`，因此不能直接复用普通关闭/退款路径。
- `RefreshExpiredEntitlements` 和 `RefreshUserEntitlement` 当前按 active Seat 汇总，必须在 S152 增加 `Round.CafeRoomID IS NULL` 过滤，否则 Cafe 到期会污染普通 `UserSubscription`。
- S152 contract/QA：`docs/workflow/tasks/pixel-cafe-phase8-expiry-s152.md`、`docs/workflow/qa-reports/pixel-cafe-phase8-expiry-s152-qa.md`，结论 `PASS / source-level`。
- 现有 60 秒 `GroupBuyLifecycleService` 在普通 Round expiry 后、旧 entitlement refresh 前运行 Cafe expiry；这里记录的是 S152 当时的 `disabled` 约束，已由 S172 的受控 `active/inactive` 状态机取代。

## 本次已完成

- 本轮将 Model Plaza 提交 `aad349f84` 与 S169 独立修复提交 `e3b0ad86c` 快进合入本地 `main`。后者覆盖 Anthropic count-token 参数清理、Composite 可用渠道展开、OPS 成功日志、Claude Code security-monitor 识别及不可调度账号 token refresh 跳过；合同和 QA 分别在 `docs/workflow/tasks/upstream-v0169-independent-fixes-s169-integration.md`、`docs/workflow/qa-reports/upstream-v0169-independent-fixes-s169-integration-qa.md`。
- 已删除本地 S132 分支及其 Git worktree 注册。Windows 留下的非 Git 残留目录删除被本地安全策略阻断；目录不在 worktree 列表且不含 `.git`，未对其执行绕过性清理。

- S173 新增唯一 `integration` route 测试 `backend/internal/server/routes/cafe_jwt_gateway_redis_smoke_integration_test.go`。它复用 disposable S169 Cafe/JWT PostgreSQL 组装并自行启动 fresh Redis，使用两个真实 APIKeyService 实例和本地 preflight terminal 证明跨实例 cache/Pub/Sub、状态变更、pin 与 Binding TTL 语义。单次和连续三次运行均清理 PostgreSQL、Redis 和 Ryuk；Docker 前后原有 9 个容器 ID 未变。

- S172 修改 `api_key_service.go`、`api_key_repo.go` 与 `cafe_room_activation_service.go`，并新增/扩展 managed-Key service/cache/activation/PostgreSQL 测试。事务更新按 Round -> Seat -> Binding -> Key 校验权益事实，成功后失效 auth L1/L2；仅在批准 allowlist 内工作，未触碰 Ent/schema/generated、migration、provider、前端、共享容器或现有 Key。
- S172 fresh PostgreSQL 测试结束后所有 disposable PostgreSQL/Ryuk 资源均清理；Docker Desktop 重启前后保持相同 9 个既有容器和容器 ID，healthcheck 容器保持 healthy。

- S171 新增唯一 integration-tagged 测试 `backend/internal/service/cafe_room_concurrency_postgres_integration_test.go`。测试复用现有真实订单事务、激活状态机和 fresh PostgreSQL helper，完整分类 100 个 Seat 请求并校验 100 次激活重试的 Round/Seat/Key/Binding/event 一致性；连续 3 轮复测额外覆盖 300+300 次调用。Docker 前后保持原有 9 个运行容器且无 Testcontainers 残留，未启用任何 Cafe Key。

- S170 新增唯一 integration-tagged route 测试 `backend/internal/server/routes/cafe_order_creation_postgres_integration_test.go`，不改产品代码。为忠实模拟现有 migration 057，测试专属数据库在 Ent schema 创建后仅补 `idempotency_records.created_at/updated_at` 的 `DEFAULT NOW()`；成功路径验证真实 JWT 路由、SQL 幂等、PaymentOrder/Seat/事件/审计和同 Key 重放，失败路径验证真实 EasyPay 构造失败后的 durable Seat release。Docker Desktop 重启后原有 9 个运行容器均保持运行，具备 healthcheck 的容器为 healthy；测试结束后无 Testcontainers 容器残留。

- S169 新增唯一 integration-tagged route 测试 `backend/internal/server/routes/cafe_authenticated_postgres_integration_test.go`，不改产品代码。该测试使用一次性 PostgreSQL，测试结束即终止；通过实际 Gin 用户路由、JWT、用户查询和 Cafe 查询验证 HTTP ownership 与付款前协议拒绝。

- S168 已完成真实浏览器的本地 fixture 验收，不改任何 Pixel Cafe 生产源文件。最终干净会话拦截全部本地 API route，控制台 `0 errors`；订单请求体确认 `seat_no=1`、`payment_type=alipay`、`agreement_accepted=true` 且带 `Idempotency-Key`。截图为 `output/playwright/pixel-cafe-s168/clean-desktop-payment-waiting.png` 与 `output/playwright/pixel-cafe-s168/clean-mobile-390-selected-room.png`。

- S153 新增 `CafeRoomLifecycleService` 并由既有 60 秒 GroupBuy ticker 唯一调用；普通 `ReconcilePendingProviderRefunds` 过滤 Cafe refund，防止管理/普通 lifecycle 越过 Cafe 边界。未满关闭只触及 `open -> failed`、`paid -> refund_pending` 和 `locked -> released`，实际退款均复用 `processSeatRefund`、PaymentOrder 与 GroupBuyRefund 的幂等边界。
- S153 聚焦回归覆盖到期关闭与余额退款、Cafe provider pending 对账隔离、`activating` 重试、paid-full `open` pre-claim 补偿，以及完整 Cafe lifecycle 对旧 expiry hook 的优先级。Wire generation、`cmd/server`、格式与 Git integrity 通过。
- S151 已完成：以 `GroupBuySeat.UserID -> Round.CafeRoom -> Plan/BoundAPIKey` 为安全查询边界，复用现有 page/page_size envelope。API 只返回 Room/Plan/Round/Seat 与 source/用户一致的 managed Key 安全元数据；无 Key、跨用户 Key 或 source 不一致 Key 都返回 `null`，不返回 Key/Account/Binding/其他用户信息。
- S151 My Rooms service/handler/route、Pixel Cafe 进行中/历史 UI、聚焦 Go/Vitest、typecheck、lint、1119-module build 和 Git integrity checks 均通过；没有 schema/migration、Key enable、支付、lifecycle、路由、Lobby/Pixi 或部署改动。

- S150 将 auth snapshot 升级至 v17，保存 Cafe managed source、Binding ID、Pinned Account ID 与 Binding expiry；旧版本缓存自动 miss。
- `GetByKeyForAuth` 只接受一个 active/current strict Binding，并验证 Key、user、group、Seat、Round、Room、assigned Account 和时间边界。缺失、重复、未来、过期或不一致 Binding 都不会写入 pin。
- L1/L2 auth cache 的 TTL 被 Binding expiry 截断；到期快照在缓存命中时删除并回源，不能在过期后继续固定到旧 Account。
- 普通和 Google/Gemini 原生 API-key middleware 对 Cafe managed Key 缺少有效 pin 时返回 `CAFE_ACCOUNT_UNAVAILABLE`。模型重解析保留 pin 且移除多组路由、账号池、sticky/fallback。
- Claude、OpenAI/Codex、Gemini/Antigravity、Grok/general、OpenAI 图片和 WebSocket `previous_response_id` 路径全部固定到 Binding Account，不会通过 sticky、model route、候选池或 provider fallback 换号。
- S150 worker result、QA report、workflow status、main log 已写入；没有 Key enable、commit、push、部署、容器更新、生产 migration、真实 JWT/支付/provider 调用。

## 已确认事实

- S173 QA：`docs/workflow/qa-reports/pixel-cafe-phase27-jwt-gateway-redis-s173-qa.md`，结论 `PASS / runtime-isolated`。真实 Redis transport、JWT Cafe HTTP 与 gateway-auth preflight 已在 temporary Testcontainers 中验证；完整 provider gateway traffic、payment sandbox、性能、部署、回滚和生产仍未验证。

- S172 QA 为 `PASS / runtime-isolated`。无标签 managed/cache 测试、fresh PostgreSQL 状态机/失效/并发、S159 activation callback、repository managed-source/auth、普通和 Google/Gemini Cafe middleware、pinned gateway、server compile、gofmt/diff/conflict/unmerged 及 Docker 前后检查均通过。`-tags=unit` 的整个 service 包仍被既有重复 helper、billing 参数和 OpenAI runtime-block 编译错误阻断，不计为 S172 PASS。

- S171 QA：`docs/workflow/qa-reports/pixel-cafe-phase25-postgres-concurrency-s171-qa.md`，结论 `PASS / runtime-isolated`。高争用下最后 Seat、激活 claim/completion、managed-source 唯一性和 active Binding 唯一性在真实 PostgreSQL 中收敛；这是 24 个同时在途调用的正确性压力，不是性能或生产容量证据。

- S170 QA：`docs/workflow/qa-reports/pixel-cafe-phase24-server-order-creation-s170-qa.md`，结论 `PASS / runtime-isolated`。通过实际注册路由/JWT、UserRepository、SQL-backed idempotency 和真实支付/拼团/Cafe service 组合证明成功下单、语义幂等重放和 provider-construction failure release；未触及产品业务源文件、migration/schema/provider、共享容器、真实商户、Key/Binding、Redis、支付 callback 或生产数据。

- S169 QA：`docs/workflow/qa-reports/pixel-cafe-phase23-authenticated-http-s169-qa.md`，结论 `PASS / runtime-isolated`。真实注册路由/JWT/UserRepository/Cafe service 和 fresh PostgreSQL 的 ownership/redaction/negative HTTP 路径已验证；成功订单、支付、Key 启用、Redis、性能和发布未验证。

- S168 QA：`docs/workflow/qa-reports/pixel-cafe-phase22-mocked-browser-user-flow-s168-qa.md`，结论 `PASS / mocked-browser`。页面单测 `10/10` 和 typecheck 通过；未访问真实 Cafe API、JWT、支付、Key、数据库/Redis 或共享容器。

- S153 QA：`docs/workflow/qa-reports/pixel-cafe-phase8b-refund-compensation-s153-qa.md`，结论 `PASS / source-level`。active/completed Round、Key/Binding reclaim、账号迁移、自动下一轮、前端、Schema 和 Key enable 都未纳入本 Sprint。
- S150 QA：`docs/workflow/qa-reports/pixel-cafe-phase7-fixed-account-s150-qa.md`，结论 `PASS / source-level`。
- S149/S150 关闭时生成的 Key 为 `disabled` 且没有 enable transition；S172 已将新激活 Key 改为 `active` 并增加受控 `active/inactive` 状态机，路由/额度/到期/删除等生命周期保护继续生效。
- Repository 回归证明 source-seat mismatch 和 Binding expiry 均会清零 pin；cache 回归证明 Binding 到期截断 L1/L2 TTL，过期缓存会删除并重新读取认证事实。
- Provider 回归证明 pinned 请求在 Claude、OpenAI/Codex、Gemini/Antigravity、Grok/general、image 与 WebSocket continuation 中不切换 Account；选账号失败保留 `CAFE_ACCOUNT_UNAVAILABLE`，模型不可用保留 `model_not_found`。
- `go test ./... -count=1` 只失败于既有 `internal/service/openai_compat_model_test.go:1877`：期望 `missing terminal event`，实际 `upstream error: 502 (failover)`；非 S150 allowlist。
- S159 QA：`docs/workflow/qa-reports/pixel-cafe-phase13-payment-callback-activation-s159-qa.md`；没有修改生产实现。现有脏 `group_buy.go`、`cafe_room_activation_service.go` 等不归因于 S159。
- S160 QA：`docs/workflow/qa-reports/pixel-cafe-phase14-provider-pending-refund-s160-qa.md`；仅新增隔离 PostgreSQL 测试，现有脏 `group_buy.go`、`cafe_room_lifecycle_service.go` 等不归因于 S160。
- S161 QA：`docs/workflow/qa-reports/pixel-cafe-phase15-initial-provider-refund-s161-qa.md`；仅新增隔离 PostgreSQL 测试，未修改 Cafe、GroupBuy 或 Payment 生产实现，既有脏生产文件不归因于 S161。
- S162 QA：`docs/workflow/qa-reports/pixel-cafe-phase16-webhook-handler-s162-qa.md`；仅新增 handler 测试，未修改 webhook、provider selection、Payment 或 Cafe 生产实现，既有脏生产文件不归因于 S162。
- S163 QA：`docs/workflow/qa-reports/pixel-cafe-phase17-stripe-webhook-instance-selection-s163-qa.md`；仅修改 `payment_webhook_handler.go` 的 Stripe order-id 解析并扩展 handler 测试，未修改 provider SDK、支付路由、Payment/GroupBuy/Cafe 生产状态机。
- S165 QA：`docs/workflow/qa-reports/pixel-cafe-phase19-wxpay-webhook-candidate-selection-s165-qa.md`；仅扩展共享 handler 测试，未修改 webhook、WeChat SDK/provider、支付路由、Payment/GroupBuy/Cafe 生产状态机。该本地 RSA/AES-GCM 证据不代表 merchant、endpoint、sandbox 或生产回调验证。
- S166 QA：`docs/workflow/qa-reports/pixel-cafe-phase20-easypay-webhook-pinned-instance-s166-qa.md`；仅扩展共享 handler 测试，未修改 webhook、EasyPay provider、支付路由、Payment/GroupBuy/Cafe 生产状态机。该本地 MD5 证据不代表 merchant、endpoint、sandbox 或生产回调验证。
- S167 QA：`docs/workflow/qa-reports/pixel-cafe-phase21-alipay-webhook-pinned-instance-s167-qa.md`；仅扩展共享 handler 测试，未修改 webhook、Alipay SDK/provider、支付路由、Payment/GroupBuy/Cafe 生产状态机。该本地 RSA2 证据不代表 merchant、endpoint、sandbox 或生产回调验证。

## 待验证点

- 真实 JWT/API 与服务端订单路径 -> 验证：在独立非生产环境，使用明确授权的受管 Key/Binding，断言身份、服务端 Seat 锁定、Order 幂等和实际路由事实；不得把 S168 浏览器 fixture 当作该证据。
- 支付 sandbox、真实 gateway 使用、性能、部署与回滚 -> 验证：分别建立已批准合同；只使用临时隔离 managed Key，不操作本机现有 Key。

- 非生产真实 service smoke -> 验证：用有效 Cafe managed Key/Binding 跑真实 JWT、provider 请求与 usage 写入，断言使用的 `account_id` 恒等于 pin，Binding 到期后立即拒绝且不回退。
- S153 PostgreSQL/payment -> 验证：并发执行相同 Cafe lifecycle tick，断言只创建一个 GroupBuyRefund/支付退款；以 payment sandbox 验证 provider-pending 对账与失败隔离。

- 非生产真实 service smoke -> 验证：用有效 Cafe managed Key/Binding 跑真实 JWT、provider 请求与 usage 写入，断言使用的 `account_id` 恒等于 pin，Binding 到期后立即拒绝且不回退。
- 支付 sandbox -> 验证：feature gate、幂等回放、callback、失败释放与 activated/active Key 状态。
- Provider signature/HTTP callback and provider-pending refund -> 验证：下一独立 contract，仍禁止启用 Cafe managed Key。
- Initial provider refund submission/partial-refund real response -> 验证：需 payment sandbox contract；S160 只验证 pending 后 query 收敛。
- Provider signature/HTTP、真实 sandbox refund、失败/partial response 与并发初始提交 -> 验证：分别建立独立 contract；S161 仅验证进程内 fake 的初始 `pending`、防重复提交和后续 query success 收敛。
- Provider-specific cryptography、merchant callback、Stripe multi-instance selection/order binding 与 provider fallback -> 验证：建立独立 contract；S162 仅验证本地 verifier double 的 handler 排序、HTTP response 和一次性 dispatch。
- Live Stripe merchant callback/endpoint configuration、其他 provider 多实例选择、secret rotation/replay window 与 provider unavailable policy -> 验证：分别建立独立 contract；S163 仅验证本地 Stripe SDK、两实例 order-bound selection 与 handler replay。
- 退款/到期/收回、My Rooms、Lobby/Redis/Pixi -> 验证：分别建立 contract、运行态和浏览器验收矩阵；不得混入 S150。

## 当前结论

- S176 实现和本地容器已就绪，当前不能标记 `done`：必须在可用浏览器中确认 `/group-buy` 无“今日使用用户”卡片、标题/说明可见且可隐藏，并确认 `/admin/settings` 三个控件存在；当前结论为 `BLOCKED / browser-tool`。

- Pixel Cafe 的设计、已批准本地实现与分层测试已完成至 S175：34 份 Pixel Cafe 合同均有对应 QA；Phase 0 的历史基线阻断已由 S142 的全量基线修复通过结果取代。Room 管理/发现、Seat 锁定与订单、支付 callback、active managed Key/Binding、用户受控 active/inactive、严格固定账号路由、My Rooms、Lobby/Pixi、退款/生命周期、浏览器 UI、真实 JWT/HTTP ownership、隔离服务端下单、bounded PostgreSQL 并发、Redis/JWT/gateway-auth 和 registered gateway usage 均有相应层级证据。
- S174/S175 已补齐临时隔离 Key 的 registered gateway forwarding、固定 Account usage attribution 和 best-effort batch-state decoding；fresh Gateway PostgreSQL 单次及三次回归均无 decode warning。仍不能表述为“所有功能都已在生产环境验证”。
- S159 仅证明验签后支付服务边界；Pixel Cafe 仍不能宣称全部完成或生产可用。
- S160 仅证明隔离 fake-provider 的 pending query 收敛；不构成真实商户退款、HTTP webhook 或生产可用证据。
- S161 仅证明隔离 fake-provider 的初始退款提交及 pending-to-query-success 状态机；不构成真实商户、payment sandbox、HTTP webhook 或生产可用证据。
- S162 仅证明本地 verifier double 下的 webhook handler 顺序和 HTTP replay；不构成真实 provider 签名、商户回调、多实例选择或生产可用证据。`go test -tags=unit ./internal/handler` 当前被 `payment_handler_resume_test.go:161` 的既有缺失 `strconv` 导入阻断。
- S163 仅证明本地 Stripe SDK signature 和 two-instance order-bound selection；不构成真实 Stripe merchant、endpoint、网络重试、其他 provider 或生产可用证据。`go test -tags=unit ./internal/handler` 仍被 `payment_handler_resume_test.go:161` 的既有缺失 `strconv` 导入阻断。

## 下一步

1. 恢复一个可操作的本地浏览器会话并完成 S176 手动验收 -> 验证：捕获桌面截图，确认 `/group-buy` 无“今日使用用户”、可见/隐藏标题区均保留房间内容，`/admin/settings` 显示标题/说明/开关控件。
2. 浏览器验收通过后更新 S176 QA/status/main-log/current-task 为 `PASS / browser` -> 验证：QA 首行 PASS、截图存在、容器健康和回滚标签仍可见；本轮 Docker guard 已释放。
3. 仅在明确需要恢复 Prompt Audit WIP 时审阅 `stash@{0}`，并另建合同与隔离 worktree；不得直接恢复或合并旧宽分支内容。

1. 无需新增本地 Pixel Cafe 功能 Sprint；验证：已批准的 S143-S175 合同和 QA 已闭合，S175 当前 workflow 为 `done`。
2. 如需继续提升 readiness，先取得商户 sandbox 凭据和 callback endpoint 的明确授权；验证：在非生产环境完成支付创建、验签回调、失败释放、重放和 Key 激活端到端证据，禁止生产商户或生产写入。
3. 获得非生产部署授权后，再定义性能、staging/canary 与回滚合同；验证：获得负载、观测、容器发布及回滚演练事实，不操作本机现有 Key。

## 验证记录

- 分支整合本轮：classifier、发布安全资源与代理流熔断均完成合同 allowlist 审计、定向测试、`git diff --check` 和无冲突索引检查；代理切片额外通过 `go test -p 1 ./... -run '^$'` 与 `go build ./...`。快进后 `main@199be5cba` 仅保留原有 `main-log.md`/`outputs/` 脏改动；`git worktree list` 仅剩主工作树，`git branch --no-merged main` 仅保留预留的 `backup/pre-s121-split-4161d254b`。

- S177：`main@8ced00f75`；S156-S161 后端全仓 compile-only、聚焦 alias/OAuth/notification/ops
  tests、Wire generation、前端 4 files / 43 tests、typecheck、1864-module build、gofmt/diff/unmerged
  checks PASS。主线的 Retry-After、429/pool retry、Messages failover 和 compact keepalive 回归也通过；
  detached 429 WIP 只会回退较新的保护，已存为 `abeba42fc` stash 后删除。清理前四份脏改动已保存为命名
  stash；本地保留分支只剩 S132、S169 和三个 backup refs。
- S176：focused Go/Vitest/typecheck/build、Linux amd64 build、gofmt/diff/unmerged checks、guarded image promotion and Compose recreation、`/health` 200、PostgreSQL/Redis health and public-settings HTTP fields all PASS; browser screenshot blocked by local browser transports. Temporary binary cleanup was rejected by local shell safety policy; image/rollback tag and data volume remain retained.

- `go generate ./ent`：PASS；`go generate ./cmd/server`：PASS。
- `go test ./internal/service -run 'Cafe.*Order|CafePublic|GroupBuy|Payment.*GroupBuy' -count=1`：PASS。
- `go test ./internal/handler ./internal/server/routes -run 'Cafe|GroupBuy|Idempotency' -count=1`：PASS。
- `go test ./cmd/server`：PASS。
- `npm.cmd run test:run -- src/features/pixelCafe/__tests__/PixelCafePage.spec.ts src/views/user/__tests__/GroupBuyView.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts`：3 files / 15 tests PASS。
- `npm.cmd run lint:check -- src/api/cafe.ts src/types/pixelCafe.ts src/features/pixelCafe/PixelCafePage.vue`、`npm.cmd run typecheck`、`npm.cmd run build`：PASS；生产构建为 1119 modules。
- `gofmt -d`、`git diff --check`、冲突标记扫描与 `git ls-files -u`：PASS；`git diff --check` 仅有既有 CRLF 转换警告。
- S149：`go test ./internal/service -run 'TestCafeRoomActivation|TestGroupBuy.*Cafe|TestAPIKey.*Managed' -count=1`、repository managed-source、`go test ./cmd/server`、service/repository compile-only、Wire generation、gofmt 和 Git integrity checks：PASS。
- S149：`go test ./... -count=1` 仅失败 `TestForwardAsAnthropic_MissingTerminalAfterClientDisconnectSkipsOpsAndFailover`（`openai_compat_model_test.go:1877`，期望 `missing terminal event`，实际 `upstream error: 502 (failover)`）；非 S149 文件。
- S150：service pinned provider/cache、repository Binding、ordinary/Google middleware 与 handler classifier 聚焦命令均 PASS；`go test ./cmd/server`、service/repository compile-only、`go generate ./cmd/server` 后复测、`gofmt`、`git diff --check`、冲突标记和 unmerged-index checks 均 PASS。
- S150：`go test ./... -count=1` 只保留同一个 `openai_compat_model_test.go:1877` 基线失败；无真实 PostgreSQL、JWT、provider、支付、部署或生产 smoke。
- S153：`go test ./internal/service -run 'TestCafeRoomLifecycle|TestCafeRoomExpiry|TestCafeRoomActivation|TestGroupBuyLifecycleService|TestGroupBuy.*Cafe' -count=1`、`go generate ./cmd/server`、`go test ./cmd/server`、gofmt、diff/conflict/unmerged checks 均 PASS。
- S153：`go test ./internal/service -count=1` 仍失败 `TestForwardAsAnthropic_MissingTerminalAfterClientDisconnectSkipsOpsAndFailover`（`openai_compat_model_test.go:1877`，期望 `missing terminal event`，实际 `upstream error: 502 (failover)`）；无真实 PostgreSQL/payment/JWT/provider/deployment 验证。
- S151：`go test ./internal/service -run 'TestCafePublic.*(MyRooms|MyRoom)|TestCafeMyRooms' -count=1`、Cafe handler/routes compile、`go test ./cmd/server`、Pixel Cafe Vitest 8/8、目标 lint/typecheck/build、gofmt、diff/conflict/unmerged checks 均 PASS。
- S151：`go test ./... -count=1` 只保留 `internal/service/openai_compat_model_test.go:1877`（期望 `missing terminal event`，实际 `upstream error: 502 (failover)`）；非 S151 allowlist。无真实 JWT/PostgreSQL/browser/payment/provider/usage/deployment 验证。
- S159：`go test -tags=integration ./internal/service -run "TestCafePaymentCallbackPaidFullPostgresIntegration" -count=1 -timeout 120s`、Cafe activation/GroupBuy/notification 回归、repository/server compile-only、gofmt、diff/conflict/unmerged 与 Docker 前后快照均通过；无 provider signature/HTTP/JWT/refund/Key enablement/deployment 验证。
- S160：`go test -tags=integration ./internal/service -run "TestCafeProviderPendingRefundPostgresIntegration" -count=1 -timeout 120s`、Cafe refund/GroupBuy/Payment query 回归、repository/server compile-only、gofmt、diff/conflict/unmerged 与 Docker 前后快照均通过；无真实 provider/HTTP、初始退款提交、Key enablement/JWT/gateway/deployment 验证。
- S161：`go test -tags=integration ./internal/service -run "TestCafeInitialProviderRefundPostgresIntegration" -count=1 -timeout 120s`、Cafe lifecycle/provider refund/Payment query 回归、repository/server compile-only、gofmt、scoped diff/conflict/unmerged 与 Docker 前后快照均通过；无真实 provider/sandbox/HTTP、partial/failure/concurrent submission、Key enablement/JWT/gateway/deployment 验证。
- S162：`go test ./internal/handler -run "TestCafeWebhookStripe" -count=1` 与 handler 组合回归、server compile-only、gofmt、scoped diff/conflict/unmerged 均通过。`go test -tags=unit ./internal/handler -run "TestCafeWebhookStripe" -count=1` 被既有 `payment_handler_resume_test.go:161` 的 `undefined: strconv` 阻断，未作为 PASS；无真实 provider cryptography/merchant/multi-instance/sandbox、Cafe activation、Key enablement/JWT/gateway/deployment 验证。
- S163：两实例 encrypted synthetic Stripe config + Stripe SDK local signature 的 handler 回归、S162/S163 组合 handler 回归、server compile-only、gofmt、scoped diff/conflict/unmerged 均通过；无真实 merchant/endpoint/network/other provider/sandbox、Cafe activation、Key enablement/JWT/gateway/deployment 验证。
- S168：Playwright 隔离浏览器以本地 `/api/v1/**` fixture 验证功能开启 Cafe 流、Claude 区/键盘 Room/空 Seat/协议/等待支付、`390x844` 无横向溢出、匿名登录重定向和 feature-disabled 旧版回退；最终 clean session 控制台 `0 errors`。`npm.cmd run test:run -- src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`：10/10 PASS；`npm.cmd run typecheck`：PASS。无真实 JWT/API、持久化订单/Seat lock、支付、Key、数据库/Redis、性能、部署或生产验证。
- S169：`go test -tags=integration ./internal/server/routes -run "TestCafeAuthenticatedHTTPPostgresIntegration" -count=1 -timeout 180s`：PASS；路由静态回归、JWT middleware 单测、`cmd/server` compile-only、gofmt、scoped diff/conflict/unmerged 也通过。测试创建并终止独立 Testcontainers PostgreSQL；无共享容器/数据库、支付/provider、Key 启用、Redis、性能、部署或生产验证。
- S170：`go test -tags=integration ./internal/server/routes -run "TestCafeOrderCreationPostgresIntegration" -count=1 -timeout 180s`：PASS；路由静态、Cafe handler、Cafe order/GroupBuy service、EasyPay provider 和 `cmd/server` 回归通过。测试创建并终止独立 Testcontainers PostgreSQL；Docker 前后快照保持原有 9 个运行容器，未替换容器或镜像；无支付 callback、真实 merchant、Key 启用/路由、Redis、生命周期、并发、性能、部署或生产验证。
- S171：`go test -tags=integration ./internal/service -run "TestCafeRoomPostgresConcurrencyIntegration" -count=1 -timeout 240s` 与 `-count=3` 均 PASS；Cafe order/activation/GroupBuy 回归、`cmd/server` compile-only、gofmt、scoped diff/conflict/unmerged 和 Docker 前后检查通过。没有产品源修改；无启用 Key、真实 gateway/provider/payment、Redis、性能、部署或生产验证。
- S172：`go test -tags=integration ./internal/service -run "TestCafeManagedKey(PostgresIntegration|Enable|Concurrency)" -count=1 -timeout 240s` PASS；10 个 fresh PostgreSQL 场景覆盖状态切换、legacy disabled、终态/保护字段、六类权益失效、无部分写入和 20 次并发启用收敛。
- S172：`go test -tags=integration ./internal/service -run "TestCafePaymentCallbackPaidFullPostgresIntegration" -count=1 -timeout 120s`、managed/activation service、repository managed/auth、普通/Google middleware、pinned gateway、`cmd/server` compile-only、gofmt、diff/conflict/unmerged 和 Docker 前后检查 PASS。
- S172：`go test -tags=unit ./internal/service ...` 被既有 service package 编译错误阻断，不计为 PASS；真实 Redis、JWT/gateway/provider/payment sandbox、性能、部署和生产仍未验证。
