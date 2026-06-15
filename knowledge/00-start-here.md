# Sub2API 知识库入口

最后更新：2026-06-11

## 项目定位

Sub2API 是 AI API 网关平台，用于把上游 AI 账号、订阅额度和模型能力封装成用户侧 API Key。核心职责包括认证、账号调度、计费、限流、请求转发、支付、后台管理和公共展示页。

技术栈：

- 后端：Go，Gin，Ent，PostgreSQL，Redis。
- 前端：Vue 3，Vite，TypeScript，Pinia，TailwindCSS，Vitest。
- 部署：Docker / systemd / release binary，部署资料在 `deploy/`。

## 先读顺序

1. 当前继续做或掉线恢复：读 `knowledge/tasks/current-task.md`。
2. 最近阶段历史：读 `knowledge/tasks/timeline.md`。
3. 了解任务快照和时间轴分工：读 `knowledge/task-state.md`。
4. 找项目结构和模块入口：读 `knowledge/project-map.md`。
5. 需要跑测试或构建：读 `knowledge/build-and-verify.md`。
6. 改后端：读 `knowledge/backend-notes.md`。
7. 改前端或公共页面：读 `knowledge/frontend-notes.md`。
8. 改 Studio Bridge、聊天生图工作台、Canvas、嵌入式登录或图片任务体验：先读 `knowledge/tasks/current-task.md` 与 `knowledge/tasks/timeline.md`，再补读 `knowledge/studio-bridge-luoye.md`、`knowledge/chat-image-embedded-workspace.md` 和 `knowledge/chat-image-studio.md`。
9. 遇到老坑、环境差异或旧入口语义：读 `knowledge/known-pitfalls.md`。

## 事实源分工

- `README*.md`：面向用户和部署者的产品说明。
- `DEV_GUIDE.md`：本地开发环境、CI、历史坑点和命令速查。
- `docs/`：支付等专题文档；目前 `.gitignore` 默认忽略多数 `docs/*`，只有少数文档例外。
- `knowledge/`：面向 AI/开发协作的项目知识入口和长期结论。
- `knowledge/tasks/current-task.md`：当前工作快照，只保留“现在做到哪、下一步是什么”。
- `knowledge/tasks/timeline.md`：阶段历史和关键决策，按倒序追加。

## 当前仓库状态提示

- 当前默认续做主线已经从 6 月上旬的 OpenAI 网关稳态、account capability routing、用户控制台和 `key/base-url` 归一，继续前移到 Studio Bridge / 落叶AI生产联调。
- `knowledge/tasks/current-task.md` 现在同时记录“当前会话正在处理的局部任务”和“仓库默认主线分流”。如果你看到其他入口还停在 gateway prompt cache、模型广场、首页、文案或 i18n 整理，不要误判成仓库整体主线仍停在那一层；最近更高频的默认开发成本已经落到 Studio Bridge、真实充值/扣费闭环和跨仓库联调链路。
- 遇到入口摘要与任务快照冲突时，先用 `knowledge/05-current-focus.md` 判断稳定主线，再用 `knowledge/tasks/current-task.md` 判断当前会话具体在做哪一条线。
- 当前工作区仍可能同时存在并行主线的未提交改动，开始新任务前先执行 `git status --short`，确认哪些文件属于本轮目标。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 当前默认心智

- OpenAI 网关稳态、账号能力路由、控制台状态面板和 `key/base-url` 归一仍然成立，但它们已从“最近主线”退成 Studio Bridge 之前的稳定背景层。
- 现在更值得优先理解的是 Sub2API 如何作为落叶AI的账号、充值、余额、默认分组、bridge internal secret 和扣费真源，以及 `/chat-images` / `/studio-bridge/launch` 如何成为当前用户侧创作入口。
- 2026-06-10~2026-06-11 已稳定的新事实，不应继续只埋在任务快照里：
  - Studio Bridge 本地配置在 env secret 存在且配置为空、禁用、缺 secret/group 或仍是 `example.com` 占位时，会自动修复成本地默认 launch / recharge / allowed domains 配置。
  - 默认生图分组不再硬编码为 `4`，会动态选择第一个 active 且允许生图的 image group；聊天分组优先 text group，缺失时复用 image group。
  - `session-probe` iframe/CSP 已成为默认验收面的一部分；本地浏览器 smoke 需确认只请求 `/studio-bridge/session-probe`，且不会再出现 `frame-ancestors 'none'`。
  - `/usage` 表格空白不代表扣费没入账；Studio Bridge 记录里 `duration_ms = null` 等字段会让前端格式化炸掉，相关修复已进入稳定知识。
- 继续做聊天生图、嵌入工作区、模型市场或上游合成时，不要只看单个前端页面；通常要把 launch token、会话恢复、默认分组、预扣/确认/退款、团队空间 actor/payer 和 OpenAI 路径背景一起当成一条链路理解。

## 知识维护规则

- 长期稳定、后续会复用的项目结论写入 `knowledge/`。
- 临时进度写 `knowledge/tasks/current-task.md`。
- 阶段归档写 `knowledge/tasks/timeline.md`。
- 不写入密钥、token、账号、私有地址或未标注的猜测。
- 写文档后用 UTF-8 回读，避免中文乱码。

<!-- codex:pge-workflow:start -->
## Planner / Generator / Evaluator Workflow

- 本仓库的交付流程产物位于 `docs/workflow/`。
- 默认 Agent Matrix：`docs/workflow/agent-matrix.md`；命中 `P/G/E`、`Agent Matrix`、`worker` 或 `测试 worker` 时按矩阵分工执行。
- 当前阶段阅读顺序：`docs/workflow/status.md` -> `docs/workflow/agent-matrix.md` -> `docs/workflow/spec.md` -> 当前 Sprint 的 contract/review/qa/fix-log。
- 会话暂停、续做或换人接手时，仍优先更新 `knowledge/tasks/current-task.md` 作为事实源；阶段完成或需要保留最近重点时追加 `knowledge/tasks/timeline.md`。
- 小型一次性修改可显式绕过该流程；多 Sprint 或需要验收门禁的任务默认启用。
<!-- codex:pge-workflow:end -->
