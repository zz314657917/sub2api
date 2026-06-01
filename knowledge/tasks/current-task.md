# 当前任务快照

最后更新：2026-06-01 18:45 +08:00

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

### 本批验证记录

- `cd backend && go test -tags unit ./internal/service -run TestIsUpstreamModelNotFoundError|TestRateLimitService_HandleUpstreamError_ModelNotFound|TestRateLimitService_HandleUpstreamError_NonModel404 -count=1`：通过。
- `cd backend && go test ./internal/service -run TestIsUpstreamModelNotFoundError -count=1`：通过，确认非 unit service 包可编译。
- `git diff --check`：通过。

### 下一步候选

- 继续合低风险网关安全修复时，优先看 `count_tokens`、WS compatibility、usage context 这类不引入数据库模型的补丁。
- 账号配额 5h/7d 自动暂停、风控运行态、DingTalk OAuth、迁移编号重排仍保持排除，除非用户明确切到该主题。

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
