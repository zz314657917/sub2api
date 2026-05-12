# 当前任务快照

最后更新：2026-06-02 11:30 +08:00

## 当前仓库状态

- 项目主仓库：`F:/mcplugins/sub2api`。
- 当前上游关键修复移植工作区：`F:/mcplugins/.codex-worktrees/sub2api-v0133-batch3`，分支 `codex/upstream-v0.1.133-batch3`。
- 本轮目标仍是按主题小步移植 `v0.1.133` 风险/安全修复，不做整包 merge。
- `docs/workflow/status.md` 当前是 `done`，对应 `sub2api-canvas-core` 已收口；仓库没有正在推进中的单一 Sprint。
- 当前工作区同时存在多条并行改动线，`knowledge/tasks/current-task.md` 需要同时说明“本次正在做的局部任务”和“仓库默认稳定主线”，避免误把一次性首页任务当成整个仓库的默认续做方向。

## 当前上游移植进度

### 已完成 checkpoint

- `codex/upstream-v0.1.133-batch3` 已有 checkpoint commit：`fix: port auth and oauth account safety fixes`。
- 本批继续移植模型 404 风险修复：上游模型不存在 404 不再直接走账号级临时不可调度，改为写账号+模型维度 `model_rate_limits`，默认冷却 30 分钟。
- OpenAI Responses、Chat Completions、Anthropic Messages 兼容路径，以及原生 Anthropic/Gateway/Bedrock 错误路径，已把当前请求的上游模型传入 `RateLimitService.HandleUpstreamError`。
- 新增 `model_not_found_error.go` 识别模型不存在 404 body，避免普通 endpoint 404 被误判为模型级冷却。
- 新增 unit 测试覆盖模型 404 写模型级冷却、写入失败仍触发 failover、普通 404 仍走既有临时不可调度规则。
- 本批继续补齐低风险 ops 指标修复：`ops_metrics_collector.queryErrorCounts` 现在排除 `is_count_tokens = FALSE`，避免 `count_tokens` 非生成类计数错误污染管理端错误率、SLA 错误和上游 429/529 指标。
- 已确认多项候选修复本地已存在：OpenAI `response.failed` 流式终态补偿、`count_tokens` generation-only 字段过滤、`context_management` sanitize、Gemini Messages `tool_use` 后接 `text` 的内容块关闭、系统更新 `already_up_to_date`、并发 acquire 失败分类、usage request context、敏感 credentials redaction、OIDC compat email、setup 初始化后拦截。
- 本批继续移植 pool-mode 同账号重试状态码可配置后端逻辑：新增 `pool_mode_retry_status_codes` credentials 解析和 `Account.IsPoolModeRetryableStatus`，所有本地 Anthropic/OpenAI/Embeddings/Images 相关 failover 路径统一使用账号级判断；未配置时仍保持默认 `401/403/429`。
- 已确认 `32ea9cfe1` API key Responses SSE body fallback 与 `cae93ae13` Responses force chat completions fallback 已在本地完整存在，未重复提交。
- 已确认 `aae20ef43` OIDC verified-email fast path 加固已在本地完整存在，并通过定向测试。
- 本批继续移植 OpenAI endpoint capability gate 后端逻辑：新增 `OpenAIEndpointCapability` 与 `openai_capabilities` credentials 解析，OpenAI Responses/Chat/Messages/WS 路由要求 `chat_completions`，Embeddings 路由要求 `embeddings`；调度器、粘性会话、previous_response sticky 和本地用户池路径都会过滤不匹配账号。未配置 `openai_capabilities` 时保持默认兼容行为；本轮不合并账号弹窗 UI 改动。
- 本批继续移植 Antigravity 低风险 usage 修复：`5a317eed5 fix: capture antigravity message_start usage` 让流式透传同时读取 `message_start.message.usage` 和 `message_delta.usage`，避免输入侧 token 丢失。
- OpenAI Images `n` 参数修复在本地已按策略等价存在：非 `dall-e-3`、非一图模型会透传 `tools.0.n`；`gpt-image-2` 继续按本地拆分为多次单图请求。新增 `9c8192485 test: cover openai images n passthrough` 固化正向回归测试。
- Channel Monitor Responses reasoning 输出修复当前不适合本批直接移植：本地尚未引入上游 `MonitorAPIMode` / `providerOpenAIResponsesPath` / API mode 持久化与前端配置链路，单独套用文本提取函数没有调用入口；应与 Channel Monitor API 模式专题一起评估。
- 本批已按用户确认移植 OpenAI/Codex 账号 5h/7d 用量阈值自动暂停专题：
  - 后端使用现有 `accounts.extra` 和 `ops_advanced_settings` JSON，不新增数据库表/字段，不引入 `user_platform_quota` 数据模型。
  - 支持账号级 `auto_pause_5h_threshold` / `auto_pause_7d_threshold`，支持全局默认 `openai_account_quota_auto_pause.default_threshold_5h/default_threshold_7d`。
  - 支持账号级 `auto_pause_5h_disabled` / `auto_pause_7d_disabled` 覆盖全局默认，实现单账号、单窗口豁免。
  - 调度接入覆盖普通 OpenAI 账号选择、load-aware、advanced scheduler TopK 初筛、session sticky recheck、previous_response sticky。
  - scheduler snapshot 保留 Codex 用量快照字段和 auto-pause 配置字段，避免缓存瘦身导致调度层看不到阈值/用量。
  - 前端已补 Ops 高级设置入口和 OpenAI 账号编辑入口，可配置全局默认阈值、账号级阈值和单窗口禁用。
- 多智能体自检后已补一轮 hardening：
  - Ops 设置弹窗会归一化旧响应缺失的 `openai_account_quota_auto_pause`，避免前端直接解引用崩溃。
  - `SettingService` 的 OpenAI quota auto-pause 缓存加入 revision guard，避免后台旧 DB 刷新覆盖刚保存的新阈值。
  - 补测 previous_response sticky 命中 quota 暂停账号时会失效并清理绑定；补测真实 `SettingService` 注入链路读取全局默认阈值。
- 已按用户要求合入 `664e9fdcd feat(usage): 用户用量按平台拆分 + UsersView 列设置可配置 + 用量列排序`：
  - 后端用户 Dashboard stats 与 UsersView 批量用量统计增加 `by_platform`，平台口径沿用 `group.platform` 优先、再 fallback `account.platform`。
  - 前端用户 Dashboard 增加按平台拆分卡；管理端 UsersView 增加平台用量子列、列设置版本迁移和本页用量排序。
  - 本地 i18n 已拆分，未恢复上游大单文件；新增文案已迁入 `dashboard.ts` 和 `admin/users.ts`。

### 本批验证记录

- `cd backend && go test -tags unit ./internal/service -run TestIsUpstreamModelNotFoundError|TestRateLimitService_HandleUpstreamError_ModelNotFound|TestRateLimitService_HandleUpstreamError_NonModel404 -count=1`：通过。
- `cd backend && go test ./internal/service -run TestIsUpstreamModelNotFoundError -count=1`：通过，确认非 unit service 包可编译。
- `cd backend && go test ./internal/service -run TestOpsMetricsCollectorQueryErrorCountsExcludesCountTokens -count=1`：通过。
- `cd backend && go test -tags unit ./internal/service -run 'Test(GetPoolModeRetryStatusCodes|IsPoolModeRetryableStatus_Account)' -count=1`：通过。
- `cd backend && go test ./internal/service -run TestOpenAI.*Pool -count=1`：通过。
- `cd backend && go test ./internal/handler -run 'TestOIDCOAuthCallbackVerifiedEmailFastPath|TestTryOIDCVerifiedEmailFastPath' -count=1`：通过，确认 OIDC 加固已存在。
- `cd backend && go test ./internal/service -run 'TestAccountSupportsOpenAIEndpointCapability|TestOpenAIGatewayService_SelectAccountWithScheduler_DefaultDisabled_EmbeddingsSkipsChatOnlyAccount|TestOpenAIGatewayService_SelectAccountByPreviousResponseID_CapabilityMismatchKeepsSticky' -count=1`：通过。
- `cd backend && go test ./internal/service -run 'TestOpenAIGatewayService_SelectAccountWithScheduler|TestOpenAIGatewayService_SelectAccountByPreviousResponseID|TestAccountSupportsOpenAIEndpointCapability|TestOpenAISelectAccountWithLoadAwareness' -count=1`：通过。
- `cd backend && go test ./internal/handler -run 'Test.*Embeddings|Test.*OpenAI.*Responses|Test.*OpenAI.*ChatCompletions' -count=1`：通过。
- `cd backend && go test ./internal/service -run "TestExtractSSEUsage|TestAntigravity.*Passthrough|Test.*Passthrough"`：通过。
- `cd backend && go test ./internal/service -run "TestOpenAIGatewayService.*Images|TestBuildOpenAIImagesResponsesRequest|Test.*Image"`：通过。
- `cd backend && go test -tags unit ./internal/service -run "TestRunCheckForModel|Test.*ChannelMonitor"`：通过。
- `cd backend && go test ./internal/service/...`：通过。
- `cd backend && go test ./internal/handler/...`：通过。
- `cd backend && go test -tags unit ./internal/repository -run TestBuildSchedulerMetadataAccount_KeepsQuotaAutoPauseFields -count=1`：通过。
- `cd backend && go test ./internal/service -run "TestOpenAIGatewayService_SelectAccountForModelWithExclusions_.*(AutoPause|Threshold|Global|Disable|Window)|TestOpenAIGatewayService_SelectAccountWithScheduler_LoadBalanceTopKExcludesQuotaPaused|TestGetOpenAIQuotaAutoPauseSettings|TestSetOpenAIQuotaAutoPauseSettings|TestUpdateOpsAdvancedSettings_PushesQuotaAutoPauseSink" -count=1`：通过。
- `cd backend && go test ./internal/pkg/apicompat/...`：通过。
- `cd backend && go test ./internal/server/...`：通过。
- `cd frontend && npm.cmd run typecheck`：通过。当前移植 worktree 未安装 `node_modules`，验证时临时创建 `frontend/node_modules` junction 指向主仓库依赖，命令完成后已删除 junction。
- `cd frontend && npm.cmd run test:run -- EditAccountModal OpsSettingsDialog`：通过，实际匹配并运行 `EditAccountModal.spec.ts` 与 `BulkEditAccountModal.spec.ts` 共 26 个测试；当前没有单独的 `OpsSettingsDialog` spec。
- `cd backend && go test ./internal/service -run "TestOpenAIGatewayService_SelectAccountForModelWithExclusions_UsesSettingServiceGlobalDefaultThreshold|TestOpenAIGatewayService_SelectAccountByPreviousResponseID_QuotaAutoPausedMiss|TestRefreshOpenAIQuotaAutoPauseSettings_DoesNotOverwriteNewerSinkValue|TestGetOpenAIQuotaAutoPauseSettings|TestSetOpenAIQuotaAutoPauseSettings|TestUpdateOpsAdvancedSettings_PushesQuotaAutoPauseSink" -count=1`：通过。
- `cd frontend && npm.cmd run test:run -- OpsSettingsDialog`：通过，新 `OpsSettingsDialog.spec.ts` 覆盖旧响应缺失 OpenAI quota auto-pause 配置时的默认回填与保存。
- `cd frontend && npm.cmd run typecheck`：通过。当前移植 worktree 未安装 `node_modules`，验证时临时创建 `frontend/node_modules` junction 指向主仓库依赖，命令完成后已删除 junction。
- `cd backend && go test ./internal/service/...`：通过。
- `cd backend && go test ./internal/repository ./internal/handler/admin ./internal/pkg/usagestats ./internal/service -run "Test.*Dashboard|Test.*Usage|Test.*User"`：通过。
- `cd frontend && npm.cmd run test:run -- UsersView Dashboard usage`：通过，17 个文件 65 个测试；`admin UsageView` 测试期间有既有 mock 缺少 `getModelStats` 的 stderr，但测试通过。
- `cd frontend && npm.cmd run typecheck`：通过。当前移植 worktree 未安装 `node_modules`，验证时临时创建 `frontend/node_modules` junction 指向主仓库依赖，命令完成后已删除 junction。
- `git diff --check`：通过。

### 下一步候选

- 继续合低风险网关安全修复时，优先看剩余 OpenAI/Anthropic API compatibility、ops 指标、认证状态处理这类不引入数据库模型的补丁；不要再按 `git cherry +` 盲合，先判断是否已由本地手工实现等价覆盖。
- Channel Monitor Responses API 模式、邮件模板/通知系统、user-platform quota、风控运行态、DingTalk OAuth 都应按单独专题做 schema/UI/迁移影响评估。
- 账号 5h/7d 自动暂停已完成本轮最小闭环；仍未合并 user-platform quota 数据模型和账号配额自动暂停以外的大范围配额体系。
- 风控运行态、DingTalk OAuth、迁移编号重排仍保持排除，除非用户明确切到该主题。

## 当前默认主线分流

1. 稳定主线仍以聊天生图/嵌入工作区链路为主
   - 近 2026-05-29 的高频提交仍集中在 `chat image workspace migration`、`/canvas`、OpenWebUI launch / redeem、用户图片库、prompt market、模型目录与定价同步。
   - 继续做这条线时，优先读 `knowledge/05-current-focus.md`、`knowledge/chat-image-embedded-workspace.md` 和 `knowledge/tasks/timeline.md`，不要只看本文件后半段的首页任务记录。

2. 当前会话层面的活跃任务是公共首页与多语言维护
   - 最近一次明确在做的单项任务，是默认首页信息架构/视觉重排，以及 `zh.ts` / `en.ts` 的维护性拆分。
   - 这属于当前工作区里的局部前端任务，不代表仓库整体主线已经从聊天生图/嵌入工作区切回公共首页。

## 当前局部任务

### 首页重排与价格公式模块

- 用户要求：参考 `https://xcode.best/` 首页截图，在当前首页下方增加更多内容；随后调整版式和文案，避免借鉴痕迹过重，并增加小图标；最新要求参考 `nextlevelbuilder/ui-ux-pro-max-skill` 设定一个设计风格。
- 后续要求：在首页明确加入“价格计算公式”部分，说明余额如何换算和消耗；把第一屏空出来，不让下方新增内容露进首屏；移除“模型接入，像控制台一样清晰 / gateway.config / 一把密钥、分组路由、账单回放”那段；再按 UI 设计师检查结果继续压缩高空卡区域。
- 当前工作区存在多项本轮前或并行的未提交改动；本轮首页任务只改默认首页和首页文案，不处理其他 payment / affiliate / tutorial 相关线。

### 本轮已完成

- `frontend/src/views/HomeView.vue` 默认首页新增价格公式模块，保留 `homeContent` 自定义首页分支不变。
- 视觉方向采用 `Enterprise Gateway` 信息架构 + `AI-Native / HUD` 语言，未引入外部依赖。
- 已移除旧 `AI-Native Command Center` 说明区，不再展示 `gateway.config` 终端预览和“一把密钥 / 分组路由 / 账单回放”三段卡片。
- 独立价格公式模块用 `1 人民币 = 1 美元`、分组倍率 `0.15x`、得到 `0.150 元人民币`、对应官方 `$1 API 用量` 的四步表达解释费用换算，并补充 Claude / OpenAI / Gemini 官方价目表入口。
- 已删除价格公式下方重复费用大卡和三张高空卡，把 CTA 和说明压缩进公式模块底部紧凑区。
- 已按用户要求删除 `Bento Workflow / 开发工具、Agent 和团队额度走同一套规则` 整块展示区，并清理对应 `codingTools` 文案、bento/workflow 样式和回归测试断言。
- 默认首页首屏改为独立 hero 视口：`home-main-stage` 使用 `100svh`，价格公式和后续内容从第二屏开始出现。

### i18n 维护性拆分

- `frontend/src/i18n/locales/zh.ts` 与 `frontend/src/i18n/locales/en.ts` 已改成聚合入口，单文件约 2KB。
- 顶层 domain 拆到 `frontend/src/i18n/locales/zh/*.ts` 与 `frontend/src/i18n/locales/en/*.ts`。
- `admin` 再拆到 `frontend/src/i18n/locales/{zh,en}/admin/*.ts`，避免形成新的超大 `admin.ts`。
- `i18n/index.ts` 的动态导入入口保持不变，因此运行时加载语义不变；这次拆分主要是维护性收益，不是为了显著缩小语言包 chunk。

## 验证记录

- `npm.cmd run test:run -- home-theme public-smoke`：通过，覆盖首页价格公式、hero 首屏和旧块不回归。
- `npm.cmd run test:run -- home-theme usageServiceTierLocales PaymentView public-smoke`：通过，29 个测试。
- `npm.cmd run typecheck`：通过。
- `npm.cmd run build`：通过；仅有既有 Vite dynamic import/chunk size 和 Node `DEP0190` 警告。
- `git diff --check`：通过；仅有若干 knowledge 文件既有 LF/CRLF warning。
- Node AST 结构检查：`zh/en` 顶层 39 个 key 对齐，`zh/en admin` 24 个子 key 对齐。
- Playwright 截图与 DOM 检查已覆盖首页桌面/移动端、价格模块和首屏高度，移动端无横向溢出。

## 当前工作区注意

- 本轮首页/i18n 相关改动集中在：
  - `frontend/src/views/HomeView.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
  - `frontend/src/i18n/locales/zh/`
  - `frontend/src/i18n/locales/en/`
  - `frontend/src/__tests__/home-theme.spec.ts`
  - `frontend/src/views/user/__tests__/PaymentView.spec.ts`
- `git status` 仍有其他并行改动未处理也未回滚，包括后端 payment / affiliate、部分 tutorial 公共页、`knowledge/00-start-here.md`、`knowledge/05-current-focus.md`、迁移脚本和教程静态资源。
- 继续新任务前先执行 `git status --short`，确认本轮只接手哪条线；不要把首页任务和聊天生图/嵌入工作区主线混成同一批提交。

## 下一步

- 如果继续做首页，可追加“支持模型 / 接入教程 / 常见问题”第三段，并在上线前核对示例倍率 `0.15x` 是否与站点真实默认分组一致。
- 如果继续做仓库默认主线，应转去聊天生图/嵌入工作区、`/canvas`、OpenWebUI launch / redeem、模型目录/定价同步，不要被本文件里的首页任务记录带偏。
