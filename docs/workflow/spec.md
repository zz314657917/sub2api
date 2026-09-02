## Upstream v0.1.185 Quota Reset Cooldown Addendum (S281)

- Adapt upstream `897faea33` in the local `accountRepository.ResetQuotaUsed`
  owner. Keep the existing repository interface and local quota/share-display
  SQL shape; make the single `UPDATE accounts` clear `rate_limited_at` and
  `rate_limit_reset_at` atomically with quota usage reset.
- Inspect `RowsAffected` and return `service.ErrAccountNotFound` for a missing
  or deleted account before enqueueing scheduler outbox work. On success,
  preserve local `model_rate_limits`, overload state, temporary-unschedulable
  state, fixed-reset configuration and share-display behavior, then enqueue
  the existing account-changed event and refresh the scheduler cache snapshot.
- Do not rename `AccountRepository.ResetQuotaUsed` or import the upstream split
  method: the local branch has many test doubles and service owners that would
  create unnecessary topology churn. Do not clear overload/temp-unschedulable
  fields; this fix is limited to account-level rate-limit cooldown.
- Prove SQL/row-count behavior with the existing sqlmock/recording executor
  tests and run repository/service compile gates. Real PostgreSQL integration,
  provider traffic, containers, deployment, commit and push are outside the
  worker scope. Contract:
  `docs/workflow/tasks/upstream-v0185-quota-reset-cooldown-s281.md`.

## Upstream v0.1.185 Gateway Pool Same-Account Retry Addendum (S280)

- Adapt upstream `b1e60ba45` in the two Anthropic compatibility forwarding
  paths. When a failover HTTP status is returned, preserve the result of local
  rate-limit handling and mark the `UpstreamFailoverError` retryable on the
  same account only when the account is in pool mode, the status is configured
  as pool-retryable, and rate-limit handling did not disable/unschedule it.
- Pass the mapped upstream model into `HandleUpstreamError` so existing
  model-specific error handling receives the same context as other gateway
  paths. Do not change status classification, retry counts, scheduler logic,
  account configuration semantics, response conversion or billing.
- Cover both Chat Completions and Responses compatibility entry points, plus
  negative cases for non-pool accounts and an explicitly empty pool retry-code
  list. Preserve every existing dirty path and `outputs/**`.
- No schema, migration, repository, frontend, dependency, provider, container,
  deployment, shared-data, commit or push action belongs to the worker scope.
  Contract:
  `docs/workflow/tasks/upstream-v0185-gateway-pool-retry-s280.md`.

## Upstream v0.1.184 Compatibility Fixes Addendum (S276)

- Port four independently testable upstream fixes without merging or
  cherry-picking the diverged `v0.1.184` history: Anthropic-to-Responses stream
  item lifecycle/content indexes (`8f5451587`), Anthropic streamed tool argument
  assembly (`da10822d7`), saved SMTP TLS fallback in test endpoints
  (`c31fe2ed9`), and custom-version suffix comparison (`9e7aff59d`).
- Adapt each behavior to the local owners. In particular, SMTP test handlers
  remain in the local monolithic `setting_handler.go`; do not import the
  upstream split-file topology or unrelated surrounding changes.
- Preserve Responses event ordering, usage/failover/billing behavior, explicit
  SMTP `true`/`false` overrides, and ordinary semantic-version comparisons.
- Schema, migrations, repositories, provider traffic, frontend, dependencies,
  VERSION, containers, deployment, shared data, push and `outputs/**` are
  excluded. Contract:
  `docs/workflow/tasks/upstream-v0184-compat-fixes-s276.md`.

## Upstream v0.1.184 Frontend Compatibility Addendum (S277)

- Adapt upstream `81e461f65` and `5778739cd` so only genuine local
  `datetime-local` strings are interpreted as local wall-clock times. Reject
  timezone-bearing, malformed and calendar-overflow strings instead of
  allowing `Date.parse` normalization; retain the browser's ordinary local/DST
  semantics for valid values.
- Redeem batch expiry must use the strict parser and serialize its seconds
  result to ISO only after parsing succeeds. Existing clear-mode and error
  text behavior remain unchanged.
- Adapt `c03776604` by removing the sole remaining
  `CLAUDE_CODE_ATTRIBUTION_HEADER=0` settings override. The generated Unix,
  CMD, PowerShell and settings JSON variants must retain
  `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` while not disabling Claude
  attribution.
- Test only these utility/view/modal owners. Backend, migrations, APIs,
  billing, lockfile/dependencies, Pixel Cafe, external services, containers,
  deployment, commit, push and `outputs/**` are excluded. Contract:
  `docs/workflow/tasks/upstream-v0184-frontend-compat-s277.md`.

## Upstream v0.1.184 Channel Pricing Normalization Addendum (S278)

- Adapt upstream `eb4237a2b` in the local `ModelPricingResolver`: after a
  literal channel-pricing miss, normalize known OpenAI/Codex model aliases and
  retry the channel lookup. This fixes requests such as
  `gpt-5.6-luna-high` or a known date-suffixed variant when the channel only
  prices `gpt-5.6-luna`.
- Literal lookup always wins so explicit variant pricing is not shadowed by a
  normalized base entry. Normalization is a no-op for non-OpenAI/unknown models
  and must not match unrelated channel pricing or alter model mapping.
- Prove the behavior through the existing usage-record path: exact model,
  effort suffix, date suffix, subscription group, exact-variant precedence,
  and unrelated-model fallback. Keep all persisted billing fields and costs
  unchanged except for selecting the intended channel price.
- No schema, migration, repository SQL, billing algorithm rewrite, provider
  traffic, frontend, dependencies, VERSION, containers, deployment, shared
  data, commit or push is part of this addendum. Contract:
  `docs/workflow/tasks/upstream-v0184-channel-pricing-s278.md`.

## Upstream v0.1.184 Group Limit Partial Update Addendum (S279)

- Adapt upstream `9f1effd71` so omitted daily/weekly/monthly group limit fields
  preserve their current values. Explicit null clears only the named limit and
  numeric values update only the named limit.
- Preserve the local Room-managed invariant: a `room_managed` subscription
  group always has unlimited group-level limits because its Room plan owns the
  managed API-key quotas. Do not change the local `<=0` unlimited convention.
- The local service owner is the currently dirty `admin_service.go`; only its
  `UpdateGroup` limit block may change. Existing Pixel Cafe quota-reset work is
  protected by an external baseline snapshot and exact no-index diff review.
- No frontend, repository, schema, migration, billing, provider, container,
  deployment, shared-data, commit or push action belongs to this addendum.
  Contract: `docs/workflow/tasks/upstream-v0184-group-limit-partial-s279.md`.

## Usage Billing Multiplier Breakdown Addendum (S275)

- Keep the persisted composite `usage_logs.rate_multiplier`, `total_cost`,
  `actual_cost`, balance deduction and APIMart `7 * 1.2` settlement behavior
  unchanged for backward compatibility.
- Add a read-only usage API projection that separates the user/group/image
  pricing multiplier from the APIMart balance-conversion multiplier. The split
  applies only to image usage under the same account/model trigger as current
  billing; ordinary usage remains `pricing = rate_multiplier`,
  `conversion = 1`.
- User and admin usage views, details and exports show the pricing multiplier
  as the configured rate and show balance conversion separately when present.
  Older API responses fall back to the legacy composite multiplier.
- No schema, migration, repository SQL, historical rewrite, billing-path,
  provider, container, deployment, shared-data, commit, push or `outputs/**`
  operation is included. Contract:
  `docs/workflow/tasks/usage-billing-multiplier-breakdown-s275.md`.

# Workflow Spec

## S108 Addendum: user usage column menu layer

### Goal

- Keep the user usage column settings menu above the sticky table header and
  record rows while it is open.

### Scope Boundary

- Dynamically elevate only the user usage filter card while the existing
  column menu is open, following the established admin usage solution.
- Do not change table-wide z-index values, menu behavior, filtering, export,
  persistence, backend APIs, deployment, or containers.

### Acceptance Boundary

- Focused view regression, frontend typecheck/build, diff/path gates, and a
  desktop browser smoke must confirm that the menu is not obscured.

## S107 Addendum: x/text security dependency update

### Goal

- Adapt upstream `c5971a6fc` so `golang.org/x/text` reaches `v0.39.0` and
  GO-2026-5970 is removed from the backend module graph.

### Scope Boundary

- Upgrade only the eight `golang.org/x/*` modules selected by the upstream
  security commit, preserving the local direct/indirect dependency topology.
- Do not change Go source, generated code, schema, frontend, deployment,
  containers, VERSION, or unrelated dependencies.

### Acceptance Boundary

- Exact module-version and checksum review, `go mod verify`, backend build,
  focused/broad Go regression attempts, vulnerability scan, diff, conflict,
  and unmerged-index gates must pass or record unrelated baselines explicitly.

## S106 Addendum: selective upstream small fixes

### Goal

- Port five isolated upstream behavior fixes for scheduler quota metadata,
  monitor decrypt-failure scheduling, subscription validity-unit display,
  usage multiplier precision, and promo expiry local-time prefill.

### Scope Boundary

- Preserve all local scheduler quota dimensions, including local-only monthly
  quota fields, while continuing to filter unrelated account metadata.
- Treat only API-key decryption failure as terminal for channel-monitor
  scheduling; ordinary failures keep retrying.
- Keep validity labels aligned with backend day/week/month billing semantics,
  retain meaningful multiplier decimals, and use the existing local-time
  formatter for promo edit values.
- Do not change persistence, schema, billing, payment execution, dependencies,
  deployment, containers, VERSION, or unrelated upstream behavior.

### Acceptance Boundary

- Focused and broader unit-tag Go tests, focused Vitest regressions, frontend
  typecheck/build, formatting, exact allowlist, conflict-marker, unmerged-index,
  and `git diff --check` gates must pass.

## S105 Addendum: filter admin accounts by OpenAI plan type

### Goal

- Distinguish OpenAI `Plus`, `Pro`, `K12`, `Team`, `Free`, `Other`, and
  `Unrecognized` accounts in account management and filter them consistently.

### Scope Boundary

- Filter the persisted `credentials.plan_type` in the repository before count
  and pagination; a selected plan category implicitly limits the query to
  OpenAI accounts.
- Propagate the same `plan_type` through list, owner/share list, filtered bulk
  edit, filtered share-status changes, and filtered export.
- Keep `share_display_tier` display-only. Do not rewrite credentials, add
  manual plan editing, change import/OAuth enrichment, add schema/migrations,
  or touch scheduler, gateway, billing, deployment, and containers.

### Acceptance Boundary

- Repository integration tests cover known aliases, K12, other, unrecognized,
  OpenAI scoping, total, and pagination behavior.
- Service/handler tests prove list, filtered bulk, and export propagation;
  frontend tests cover filter snapshots and normalized badge labels.
- Focused Go/Vitest, typecheck, production build, formatting, exact allowlist,
  conflict-marker, unmerged-index, and `git diff --check` gates must pass.

## S104 Addendum: preserve OpenAI plan type across inactive workspaces

### Goal

- Preserve a token-derived OpenAI/K12 plan type when ChatGPT
  `accounts/check` also returns inactive workspace billing metadata.

### Scope Boundary

- Skip deactivated, disabled, deleted, inactive, suspended, and expired
  workspace candidates when selecting fallback account information.
- Apply the `accounts/check` plan type only when the token did not provide one;
  keep subscription-expiry, email, privacy, timeout, and logging behavior.
- Do not change Codex session identity matching, PAT/Agent Identity support,
  persistence, scheduler routing, gateway behavior, frontend, migrations,
  billing, deployment, or containers.

### Acceptance Boundary

- Focused service tests cover expired/deactivated candidate rejection and
  token-plan preservation while retaining empty-plan fallback behavior.
- Formatting, exact allowlist, conflict-marker, unmerged-index, and
  `git diff --check` gates must pass.

## S82 Addendum: clarify OpenAI WS account-mode prerequisite

### 一句话目标

- 明确账号级 OpenAI Responses WS mode 只有在全局 `gateway.openai_ws.mode_router_v2_enabled=true` 时生效，避免管理员配置后仍走 legacy 路由却缺少提示。

### 边界与验收

- 只修改 README、示例配置注释、中英文账号帮助文案与定向 locale 测试；不修改任何运行时配置值或路由代码。
- 本地账号级模式继续限定为 `off / ctx_pool / passthrough`；上游新增的账号级 `http_bridge` 不属于本地能力，不能写入帮助文案。
- 本地 `http_bridge_enabled` 仍只是大首包 HTTP fallback 开关，与账号级模式保持区分。
- 定向 Vitest、typecheck、production build、精确十一项路径审计和 protected-hash gate 必须通过。

## S83 Addendum: minute-level subscription expiry display

### 一句话目标

- 在管理员和用户订阅页面将有效期时间显示到分钟，保留现有日期 locale、失效状态和剩余天数逻辑。

### 边界与验收

- 复用 `formatDate` 新增分钟级 helper，只替换管理员订阅视图和本地 `UserSubscriptionsPanel` 的有效期展示；不修改 API、后端、计费、时区或 UsageView。
- 无效日期继续显示空字符串，状态文案和剩余天数计算保持不变。
- 定向 Vitest、typecheck、production build、十项路径审计和 protected-hash gate 必须通过。

## S84 Addendum: buffered Anthropic JSON content type

### 一句话目标

- 修复 OpenAI-compatible Anthropic buffered 响应被上游 SSE header 污染的问题，确保非流式 JSON 响应声明 `application/json`。

### 边界与验收

- 只在 buffered 转换路径 `c.JSON` 前覆盖 Content-Type；流式 SSE、响应 body、usage、计费和 failover 不变。
- 按本地函数签名重写回归测试，不直接复制上游测试参数。
- 定向 Go 测试、gofmt、八项路径审计、冲突/diff 和 protected-hash gate 必须通过。

## S85 Addendum: avoid cache billing on same-account retry

### 一句话目标

- 同账号重试期间不因 sticky/bound session 单独强制缓存计费；真正切换账号或上游显式要求时仍保持缓存计费。

### 边界与验收

- 只修改 failover 状态中的 `ForceCacheBilling` 判定和对应 handler 测试；不修改重试次数/延时、账号切换、临时封禁、错误分类或计费计算。
- 定向及 broader handler 测试、gofmt、八项路径审计、冲突/diff 和 protected-hash gate 必须通过。

## S81 Addendum: renew expired admin subscription assignments

### 一句话目标

- 管理员重新分配已过期的同分组订阅时，复用原记录并从当前时间开启新周期，而不是返回成功但继续保持过期。

### 边界与验收

- 仅修改 admin assignment 的复用分支；单个/批量分配都复用同一逻辑。
- 新周期重置 starts/expires、active 状态、日/周/月窗口与用量，并保留原 ID 和分配来源。
- suspended 无论是否已过期都不得自动恢复；有效 active 订阅的幂等复用/冲突语义保持。
- 管理员相同备注不重复追加，不同备注追加一次；购买/兑换使用的 `AssignOrExtendSubscription` 仍记录每次事件，即相同备注继续追加。
- 不修改 repository、schema、migration、handler、frontend、billing、payment、redeem 或部署配置。

## S80 Addendum: Redis Compose command continuation

### 一句话目标

- 把上游 `be74deae7` 的 Redis 启动参数续行修复移植到本地三个内置 Redis 的 Compose，确保 RDB、AOF、fsync 与可选密码参数真正传给同一次 `redis-server` 调用。

### 边界与验收

- 只修改 `docker-compose.yml`、`docker-compose.local.yml`、`docker-compose.dev.yml` 的 Redis command；本地/开发文件是与上游主文件同构的行为补齐。
- 保持 Redis image、容器名、healthcheck、`REDISCLI_AUTH`、volume、network、端口和其他服务完全不变。
- `docker-compose.standalone.yml` 使用外部 Redis，不修改。
- 验收只做 Docker Compose 静态渲染、空/非空受控密码命令检查和路径审计；不启动、更新、重启或删除任何容器/volume。
- 任意特殊字符密码的 shell 安全化、真实 Redis `CONFIG GET`、磁盘增长评估和部署均属于后续独立任务。

## S79 Addendum: upstream v0.1.161 low-risk compatibility

### 一句话目标

- 在不整体合并 `v0.1.161` 的前提下，把四个互不扩张的低风险行为移植到本地：Antigravity 付费 tier 保留、Anthropic 监控文本块提取、Claude Code `[1m]` 后缀归一化、套餐有效期动态单位文案。

### 边界与验收

- Antigravity 的异常状态与原因继续保留，但已识别的 `Pro/Ultra` 不再被 `IneligibleTiers` 覆盖为 `Abnormal`。
- Anthropic monitor 只拼接 `content[]` 中的 `type=text` 块；thinking/tool 块不参与 challenge 判断。
- `[1m]` 只在 Anthropic 请求模型末尾按大小写不敏感方式剥离，可处理重复后缀；归一化必须进入实际 handler 转发 body，其他协议和中间后缀保持不变。
- 现有 payment locale key 不改名，只去掉“天/days”的写死文案，继续由 `validity_unit` 表达单位。
- 不修改 deploy/Compose、Responses SSE、Grok media、routing、subscription assignment、migration、billing、security、VERSION、lockfile 或 `knowledge/**`。

## S75 Addendum: admin usage column-menu stacking

### 一句话目标

- 修复管理端使用记录中“列设置”及筛选下拉被固定表头/记录遮挡的问题，让筛选卡片的浮层稳定位于表格之上。

### 边界与验收

- `UsageFilters` 当前为 `z-30`，而 `DataTable` 固定表头最高为 `z-index: 220`；仅在 `showColumnDropdown` 为真时，从 `UsageView` 向筛选组件传入 `z-[221]`。
- 不修改 `DataTable`、`UsageTable`、筛选状态、菜单交互或请求逻辑，也不引入 Teleport。
- 视图测试锁定菜单打开/关闭时筛选卡片的动态层级；跑定向 Vitest、typecheck、production build 与 `git diff --check`。

## S74 Addendum: support ticket user context

### 一句话目标

- 让管理员在工单管理中直接看到用户的用户名和注册邮箱，并可从工单详情以只读方式打开现有用户信息弹窗，查看最近使用、订阅订单和充值/余额流水。

### 边界与验收

- 管理端工单列表和详情只增加 `{ id, username, email }` 用户摘要；用户侧工单接口不能新增该字段，完整 `AdminUser` 只能在管理员主动点击后通过既有 `/admin/users/:id` 接口获取。
- 资料摘要必须由工单查询一次性关联用户表获得，不能在前端按工单逐条请求用户，避免 N+1。
- 管理端必须使用独立的 ticket DTO mapper 填充摘要；原有用户侧 mapper 和序列化结果必须继续不含 `user` 摘要，并由定向测试锁定。
- `UserBalanceHistoryModal` 在工单页以 `hideActions` 只读模式复用；最近使用请求最近 30 条，并以固定高度的纵向滚动区域展示，不应撑高整个弹窗。
- 不修改数据库 schema、支付/余额/订阅业务、鉴权路由或用户端工单界面。

## S45 Addendum: affiliate risk scoring and alert scanner

### 一句话目标

- 做一个最小可用的邀请返佣风控扫描器：扫描间隔默认 `20m` 且可在后台设置调整，扫描最近 `12h`，按风险评分写入 `ops_alert_events`，并在高风险时冻结邀请奖励兑现；第一版不自动封号、不自动禁用 API key、不回滚历史奖励。

### 当前结论

- 这不是单条规则封禁，而是“风险评分 + 告警 + 奖励兑现冻结”。
- 默认扫描窗口从原先讨论的 `24h` 收敛为 `12h`；扫描周期默认 `20m`，但必须加入后台设置，方便运营自行调整。
- 扫描间隔设置建议限制在 `5-1440` 分钟，非法值回退 `20m`，避免过频扫描压数据库。
- 高风险处理只冻结返佣变现路径：
  - 阻止被邀请人首次 API 调用奖励 claim。
  - 阻止邀请返佣 quota 转余额。
  - 不扣回已发 ledger、不移除邀请关系、不封用户、不禁用 API key。

### 主要触达面

- 数据源：`users.register_ip`、`users.last_login_ip`、`usage_logs.ip_address`、`user_affiliates.inviter_id`、`user_affiliate_ledger.action = 'api_call_reward'`。
- 后端服务：新增 affiliate risk scanner、IPv6 `/64` 归一化、风险评分和去重。
- 后台设置：新增扫描间隔设置项，不需要新增独立风控页面。
- 性能索引：补齐 `users(created_at)`、`user_affiliate_ledger(action, created_at)`、`usage_logs(ip_address, created_at) WHERE ip_address <> ''`。
- 持久化：新增风险冻结记录或等价状态，用于拦截 claim/transfer。
- 运维告警：复用 `ops_alert_events`，并尽量复用现有 ops email 通知路径。
- 启动调度：按现有后台服务模式启动，使用 Redis leader lock / heartbeat 风格。

### 评分和分级

- 同一邀请人 `12h` 内邀请 `>=3` 个账号：`+25`。
- 邀请人和被邀请人登录 IP 相同：`+40`。
- IPv6 同 `/64`：`+35`。
- 注册 IP 分散但登录 IP 或 `/64` 聚合：`+25`。
- 注册后 `30m` 内触发 API 奖励：`+20`。
- 多个被邀请账号邮箱像批量生成：`+10`。
- 被邀请关系已撤销/禁用但存在 `api_call_reward`：`+30`。
- `>=50` 为 `P3` 告警，`>=70` 为 `P2` 告警并冻结兑现，`>=90` 为 `P1` 高风险冻结兑现。

### 当前阻塞与风险

- 当前主工作树仍有 payment、welfare voucher、settings、billing、gateway、frontend payment/i18n、knowledge 脏改。
- S45 会触达 affiliate、ops、wire、migration、后台任务启动链路和最小 settings 表单；必须在干净 worktree 或收口脏树后实现。
- `OpsService.CreateAlertEvent` 只负责写事件；现有 email 发送在 alert evaluator 内部，S45 实现时必须避免复制大段邮件逻辑。
- migration 编号不能直接假定；实现前要重新检查 tracked/untracked migrations。
- 当前已有 `usage_logs(user_id, created_at)`、`usage_logs(created_at)`、`user_affiliates(inviter_id)` 等基础索引，但缺少上述风控扫描专用索引；S45 实现必须补窄范围 migration，避免按 IP/时间或 ledger action/time 扫全表。

### 推荐执行计划

1. 评审并批准 `docs/workflow/tasks/affiliate-risk-alerts-s45.md`。
2. 在干净 worktree 开发，先实现 IPv6 `/64` 归一化和评分纯函数测试。
3. 增加 risk repository 查询最近 `12h` 邀请/登录/API 奖励/usage IP 聚合数据。
4. 增加 scan-specific indexes migration：`users(created_at)`、`user_affiliate_ledger(action, created_at)`、`usage_logs(ip_address, created_at) WHERE ip_address <> ''`。
5. 增加 scanner 服务：读取后台配置的扫描间隔，默认 `20m`，Redis leader lock、ops heartbeat、去重写 `ops_alert_events`。
6. 增加风险冻结持久化，并在 `ClaimInviteeAPICallReward` / `TransferAffiliateQuota` 前拦截。
7. 增加后台设置项，允许调整扫描间隔。
8. 跑定向 Go 测试、frontend typecheck、migration 编号检查、`git diff --check` 和 denied-path audit。

### 明确不在 S45 范围内

- 不做前端新页面；只允许在后台设置中增加扫描间隔控制，运维中心先复用现有告警列表。
- 不自动封号。
- 不自动禁用 API key。
- 不扣回历史奖励。
- 不删除或撤销邀请关系。
- 不混入支付、福利券、Studio Bridge、OpenAI image/video 或前端 payment 脏改。

## S43 Addendum: upstream v0.1.143 group peak-rate synthesis plan

### 一句话目标

- 把上游 `v0.1.143` 的订阅分组高峰时段倍率能力拆成一个独立产品级合成批次，先完成边界和验收计划，再决定是否进入 schema/migration + billing/gateway + frontend 实现。

### 当前结论

- 本地尚未合入订阅分组高峰时段倍率能力。
- 本地现有能力包括普通 `rate_multiplier`、`image_rate_multiplier` 和用户专属分组倍率/RPM；未发现 `peak_rate_enabled`、`peak_start`、`peak_end`、`peak_rate_multiplier` 这套字段和链路。
- 上游该能力不是小补丁，不能混入 S38a-S42 这类安全兼容提交。

### 上游来源

- `915c60b15 feat(group): 订阅分组新增可选的高峰时段倍率，以支持智谱等coding plan的高峰时段`
- `1034f576d fix: 高峰倍率全链路透传、计费术语修正与边界处理`
- `11a3da65c fix(group): harden peak-rate config handling and label peak windows with server timezone`

### 主要触达面

- 数据层：`backend/ent/schema/group.go`、Ent 生成代码、group migration。
- 后端接口：admin group create/update/list DTO、available channels、public settings server timezone。
- 服务层：group validation/normalization、API key auth cache/group hydration、gateway/openai gateway usage recording。
- 计费层：token 计费倍率叠加高峰因子，图片按次计费不受高峰因子影响。
- 前端：admin GroupsView、GroupBadge、GroupOptionItem、AvailableChannelsTable、Payment/Subscriptions/Keys 页面和 i18n。

### 当前阻塞与风险

- 当前主工作树仍有 payment、welfare voucher、settings、billing、gateway、frontend payment/i18n、knowledge 脏改。
- 上游高峰倍率触达路径与当前脏文件重叠：`backend/internal/handler/dto/settings.go`、`backend/internal/handler/payment_handler.go`、`backend/internal/service/billing_service.go`、`backend/internal/service/gateway_service.go`、`backend/internal/service/openai_gateway_service.go`、`backend/internal/service/setting_service.go`、`frontend/src/types/index.ts`、`frontend/src/types/payment.ts`。
- 上游 migration 是 `158_add_group_peak_rate_multiplier.sql`，本地 migration 已推进到更高编号且存在未提交 migration 工作；实现时必须使用本地下一安全编号。
- 上游允许 `peak_rate_multiplier=0` 表示高峰免费策略；本地实现前需要确认是否接受该产品语义。

### 推荐执行计划

1. 先收口或隔离当前 payment/welfare/settings/gateway/frontend dirty tree。
2. 在干净 branch/worktree 上开启实现 Sprint，不直接 cherry-pick 三个上游提交。
3. 先做 schema、migration、Ent 生成和 DTO/mappers。
4. 再做 group service 的校验、归一化、server timezone 计算和 API key cache 透传。
5. 然后接入 gateway/openai gateway 计费快照，锁定“高峰只影响 token 倍率，不影响图片按次计费”。
6. 最后接前端 admin/user display 和 i18n，跑后端定向测试、frontend typecheck/Vitest、`git diff --check` 与 staged denied-path audit。

### 明确不在 S43 范围内

- 不修改业务代码。
- 不新增 migration 或 Ent 生成代码。
- 不触碰当前 dirty payment/welfare/settings/gateway/frontend 文件。
- 不把 post-`v0.1.143` 的 `a5638a4e5`、ops realtime stats、redeem invitation fix 混入高峰倍率计划。

## S35 Addendum: upstream v0.1.142 merge plan

### 一句话目标

- 在不整体 merge `upstream/main` / `v0.1.142` 的前提下，把上游 `v0.1.142` 与本地长期 fork 的差异拆成可评审、可验证、可暂停恢复的分批合并计划。

### 当前背景

- GitHub latest release 已确认为 `v0.1.142`，tag 提交为 `60da9ba17`，发布时间为 2026-07-01。
- 本地 `main` 已经长期分叉，直接 `git merge --no-commit v0.1.142` 在临时 worktree 中触发大量冲突，集中在 Ent 生成代码、account/proxy schema、gateway、payment、usage 和前端视图。
- 本地主工作树当前仍有未提交的福利券、设置、用户代理、知识库等改动，且 `main` 相对 `origin/main` ahead；合入上游补丁前必须先收口或隔离这些本地变更。
- 上轮只读筛选已经证明，多个 `v0.1.138..v0.1.142` 小补丁可以在干净本地 `HEAD` 上通过 `git apply --check --3way`，但 Grok、Spark shadow、Codex detect、Anthropic dateline / Sonnet5 属于大功能链路，需要另开 Sprint。

### 合并策略

- 禁止整段 merge/rebase `v0.1.142`。
- 继续采用“小批次 port + contract + 定向测试 + denied-path audit”的方式推进。
- 每一批只允许包含同一业务域的上游补丁；支付、OpenAI/Codex 网关、usage billing、订阅、前端 API base 不能混成一个提交。
- 当前 dirty tree 未收口前，不启动代码迁移；若必须先 port，应使用干净 worktree，并且 contract 必须声明不会触碰主工作树脏改。

### 推荐批次

1. `S36 payment-refund-safe-bundle`
   - 候选提交：`c6f375d3a`、`b1403e8b2`、`55242ffac`、`65ad7df4f`、`7316d8302`、`93a3bf307`、`930326116`。
   - 目标：订阅金额、汇率换算、退款 pending、支付卡片和币种显示的安全修复。
   - 风险：触碰支付前后端与退款流程，必须定向跑 payment service / handler / frontend vitest。

2. `S37 openai-codex-gateway-safe-bundle`
   - 候选提交：`9491de0a3`、`ae5e980dd`、`65fa72892`、`0a97a5f46`、`2b49d662c`、`011278204`、`e5f7836bf`、`73de2ea7f`、`b28a22333`、`82553c4dc`、`7a38c6621`。
   - 目标：OpenAI/Codex transport failover、tool args 去重、Codex image bridge、GPT-5.5 Pro Codex 名称保留、count_tokens bridge 等网关兼容修复。
   - 风险：部分行为与本地既有 S23-S30 迁移重叠，必须逐项判定 `ported / equivalent / skipped`。

3. `S38 billing-subscription-safe-bundle`
   - 候选提交：`9f5b57fc9`、`03727ac36`、`fd004bdd8`。
   - 目标：余额计费防透支、订阅撤销软删除、account query `Count` 污染修复。
   - 风险：`9f5b57fc9` 会触碰 `usage_billing_repo.go`、`billing_cache_service.go`、`gateway_service.go`，与当前福利券/usage billing 脏改重叠；必须等福利券工作树收口后再做，或独立干净 worktree 迁移。

4. `S39 frontend-small-fixes`
   - 候选提交：`2a58a57a7`、`8c2d9b9a1`。
   - 目标：前端 direct requests 使用配置的 API base；是否移除 `gpt-5.3-codex` 默认模型由本地产品策略决定。
   - 风险：触碰前端 i18n、Settings、KeyUsage 等文件；需要前端 typecheck / targeted vitest。

5. 独立大功能 Epic，不混入小补丁批次：
   - Grok subscription / media / OAuth / quota 链路。
   - OpenAI Spark shadow account。
   - `codex_cli_only` engine fingerprint 加固与 app-server 配置。
   - Anthropic OAuth dateline 指纹抹除与 Sonnet5 适配。

### 明确不在范围内

- 不整体 merge `v0.1.142`。
- 不在 S35 修改任何业务代码。
- 不触碰 Ent、migrations、wire、生产配置、Docker/deploy、README/assets，除非后续独立 Sprint contract 明确批准。
- 不把当前未提交的福利券、发票、用户代理、知识库改动纳入上游合并批次。

### 验收标准

- S35 只交付合并计划和下一步 contract 草案，不做代码迁移。
- `docs/workflow/status.md` 进入 `contract-draft`，下一合法动作是评审 S35 plan contract。
- `docs/workflow/tasks/upstream-main-v0142-merge-plan-s35.md` 明确成功标准、允许路径、禁止路径、候选批次、验收命令和 stop rules。
- `git diff --check` 覆盖本轮 workflow 文档。

## S18 Addendum: APIMart task webhook

### 一句话目标

- 在不改变普通图片同步代理兼容行为的前提下，为 APIMart 视频/长任务接入任务完成 webhook，让 Sub2API 能在任务终态时主动完成本地状态落库、结算和失败退款。

### 当前背景

- Sub2API 当前仍是 Studio Bridge / 落叶AI的账号、余额、分组和扣费真源。
- `chatgpt2api` / 落叶创作台负责任务体验，但任务成功扣费、失败退款和使用记录最终应回到 Sub2API 侧闭环。
- APIMart task webhook 只在任务 `completed` / `failed` 等终态后回调；因此它适合补强长任务可靠结算，不适合作为普通同步 OpenAI 图片接口的直接替代。
- 本地视频任务已经有 `openai_video_tasks`、预扣、`/v1/tasks/:task_id` 查询结算和失败退款逻辑，S18 应复用这条链路，而不是新增并行账本。

### 明确不在范围内

- 不把普通 `/v1/images/generations` 同步响应改为异步 `task_id` 返回。
- 不覆盖客户请求里已经带的 `webhook` 字段；客户 webhook fan-out/relay 另开任务。
- 不新增数据库迁移，不写真实公网域名或 secret，不改 Studio Bridge / chatgpt2api 协议。
- 不整体重构 Image Creator 或 APIMart 图片轮询逻辑。

### 验收标准

- webhook 接收端有 secret 校验、body 大小限制、脱敏日志和幂等处理。
- APIMart 视频任务只在配置完整且请求未带客户 webhook 时注入 Sub2API callback URL。
- 成功终态只结算一次；失败终态只退款一次；重复回调不重复扣退。
- 现有 `/v1/tasks/:task_id` 查询结算仍作为兜底可用。
- 定向后端测试和 `git diff --check` 通过，denied-path audit 不触碰图片同步代理、迁移、Ent、公共页、支付页、Canvas、Studio 前端等禁止范围。

## S19 Draft Addendum: upstream v0.1.137 postfixes

### 一句话目标

- 在 S15-S17 已完成的基础上，继续筛出 `v0.1.137` 中不碰迁移、不碰前端、不覆盖本地产品定制的后端小修，作为 S18 之后的候选小步迁移。

### 当前背景

- 上游 `v0.1.137` 的安全/兼容主干已经通过 S15/S16/S17 小步迁入。
- 仍有少量后端补丁有合并价值，但不应和当前 S18 APIMart webhook 实现混在一起。
- 当前 S19 只是 follow-up contract 草案；`docs/workflow/status.md` 仍以 S18 `contract-draft` 为当前合法动作。

### 候选范围

- OpenAI failover 复用原始错误体。
- Anthropic window cooldown 保留。
- Account repository 列表参数限制与 refresh candidates SQL 修复。

### 明确不在范围内

- 不整体 merge/rebase `upstream/main`。
- 不碰 Ent、migrations、VERSION、wire 生成物或前端。
- 不把 OpenAI image failover、token refresh retry amplification、OAuth promo signup、scheduler outbox dedup/cleanup、cyber policy、channel monitor jitter、Claude OAuth system prompt blocks 混入本轮。

### 验收标准

- 定向后端 service / repository / server contract 测试通过。
- `git diff --check` 通过。
- denied-path audit 返回 `NO_DENIED_PATHS`。
- worker/result 和 QA report 说明每个上游候选是 `ported`、`equivalent` 还是 `skipped`。

## 一句话目标

- 在不覆盖本地 Studio Bridge / 支付治理 / Canvas / 公共页等产品定制的前提下，把上游 `v0.1.137` 的低风险安全、兼容、计费兜底和管理员运维能力按独立 Sprint 小步迁入，并为后续继续评估候选 patch 保留清晰边界。

## 当前背景

- 本地 `sub2api` 已不再处于“直接跟上游 merge”阶段；`main` 与 `upstream/main` 分叉较大。
- 仓库当前产品主线仍是 Studio Bridge / 落叶AI生产联调、支付套餐与用户治理。
- 同期存在一条独立工程主线：从上游按 Sprint 迁入低风险 patch，但显式保护本地定制，不碰 Ent/migrations/VERSION，不整体 merge `upstream/main`。
- 2026-06-17 已完成三个连续 Sprint：
  - `upstream-main-v0137-safe-patches-s15`
  - `upstream-main-v0137-small-compat-s16`
  - `upstream-main-openai-quota-reset-s17`

## 当前范围

- 当前 workflow spec 只描述 2026-06-17 这条上游小步合成链路。
- 不覆盖 Studio Bridge / 支付套餐 / 用户 IP / 首充福利这类产品主线的完整需求文档。
- 不把“后续也许会迁”的候选 patch 说成已批准范围。

## 已完成 Sprint 范围

### S15: 安全 / 兼容 / 计费兜底

- 锁定前端 `form-data` 到 `4.0.6`。
- token refresh 增加不可重试错误分类。
- 上游响应支持 zstd。
- 非流式 2xx 非 JSON 与 SSE `event:error` 进入 failover，并保留原始错误体。
- `tool strict` 缺省补 `false`。
- 国产模型 fallback pricing 和图像输入 token 计费补齐。
- DeepSeek `reasoning_effort=max` 归一到 `xhigh`。
- Anthropic thinking block 过滤改为按 mapped upstream model 分流。

### S16: 小兼容补丁

- Responses API sticky hash 在缺少旧字段时以 `input` 兜底。
- Claude Code `max_tokens=1` 的 Haiku 流式探测拦截补齐。
- OpenAI APIKey `/responses` probe 增加工具能力校验。
- API Key ACL 拒绝信息补充实际 client IP。

### S17: OpenAI OAuth 上游 quota/reset

- 新增管理员 OpenAI OAuth 账号上游 WHAM quota 查询入口。
- 新增 rate-limit reset credit 消费入口。
- 后端复用 token provider、privacy client factory 和账号代理解析。
- 前端仅在 OpenAI OAuth usage cell 展示上游 credits 查询/重置控件。

## 明确不在范围内

- 不整体 merge 或 rebase `upstream/main`。
- 不触碰 Ent schema、migrations、VERSION、wire 大链路生成物。
- 不覆盖本地 Studio Bridge、Canvas、支付页、公共页、模型市场、工单或 Chat/Image Studio 定制。
- 以上边界仅适用于 S15-S17 这条上游小步迁移链路；后续统一 API Key / APIMart 图片网关 / 前端导航合并已经触达 `wire_gen.go`、Studio Bridge repo、公共页、模型市场、Keys 和 Settings 等路径，不能再用本节作为当前 `origin/main..HEAD` 的 denied-path 证明。
- 不引入 migration-heavy、compliance gate、cyber policy、渠道监控 jitter、Claude OAuth system prompt blocks 等高风险链路。
- 不把前端全量 Vitest 失败修复混进本轮 Sprint；这应另开前端稳定化任务。

## 当前稳定工程边界

- 当前允许的上游迁移策略是“低风险 patch + 独立 Sprint + 定向验证 + denied-path audit”。
- 每轮 Sprint 都必须显式说明：
  - 为什么该补丁适合独立迁移
  - 为什么不会覆盖本地产品面
  - 需要哪些定向测试证明行为稳定
  - 哪些更大链路明确跳过
- 当前稳定结论不是“本地已接近上游”，而是“本地已有一条可持续的小步迁移方法”。

## 验收标准

- patch 迁入后，定向后端测试通过。
- 涉及前端控件或管理页时，定向 Vitest 通过。
- `git diff --check` 通过。
- 上游小步 Sprint 的 denied-path audit 应返回 `NO_DENIED_PATHS`；若当前批次是产品合并或 UI/网关主线合并，则必须改为列出实际触达路径和对应验证，不能沿用旧审计结论。
- lockfile 扫描无已知需规避版本残留，例如 `form-data@4.0.5`。
- 迁移结果不触碰本轮禁止路径；若后续合并触达曾经的禁止路径，workflow/knowledge 必须明确说明这是新批次范围，而不是继续复用旧 Sprint 证据。
- workflow 文档能说明“为什么这轮可以迁、为什么其他候选仍应跳过”。

## 当前证据入口

- `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
- `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
- `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
- `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
- `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
- `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`

## 下一步候选

- 若继续做上游迁移，需另开 Sprint 单独评估候选 patch。
- 当前可评估但未批准的方向包括：
  - OpenAI image failover
  - Anthropic window cooldown
  - account list parameter batching
  - token refresh retry amplification / outbox dedup
- 若继续做前端全量测试收口，应另开“前端稳定化”任务，不与上游 patch Sprint 混合。

- Earlier spec addenda were archived by pge-compact at 20260831T062627697Z.
## Upstream v0.2.0 Group Pricing Layout Addendum (S290)

- Hand-adapt upstream `1a33dc8cc` only where its responsive layout intent maps
  to the local six-field `PricingEntryCard` and flex-based `IntervalRow`.
  Make group create/edit dialogs wide and their model-pricing headers able to
  wrap without shrinking the add button. Preserve all inputs, emits, pricing
  conversion and billing behavior.
- The group create/edit callers intentionally pass `hide-token-intervals=true`.
  Therefore their browser acceptance covers the actually rendered six default
  Token-price inputs, not an unreachable `IntervalRow`. Keep the shared
  interval-grid source assertion in S290; a browser smoke for the enabled
  channel-pricing route is a separate follow-up and must not change group
  pricing semantics merely to satisfy this layout task.
- The component topology has diverged, so do not cherry-pick the four-file
  patch. Use a responsive grid consistent with local fields and prove the
  source-level layout sentinels, typecheck and production build. Browser
  verification must use a task-owned profile; no production credentials or
  shared data are permitted.
- Backend, schema, migration, router, auth, billing, i18n, dependencies,
  lockfile, Pixel Cafe, protected dirty files, containers, deployment, push and
  `outputs/**` are excluded. Contract:
  `docs/workflow/tasks/upstream-v0200-group-pricing-layout-s290.md`.
