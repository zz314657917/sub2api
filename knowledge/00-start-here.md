# Sub2API 知识库入口

最后更新：2026-08-24

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

- 当前默认续做主线已前移到已发布的 `S252-S256 Pixel Cafe` 功能链；此前 S244-S259 选择性合入与 S248 已在同一 `origin/main` 历史中收口。共享/生产数据库、真实 provider 与部署仍未执行。
- 现在判断“仓库在做什么”，先看 `docs/workflow/status.md` 和 `knowledge/tasks/current-task.md`，再用 `knowledge/tasks/timeline.md` 补足最近阶段历史。
- `knowledge/tasks/current-task.md` 现在应优先承载 `S248 + Pixel Cafe + worktree cleanup` 的收口快照；如果入口摘要与 workflow 文档冲突，先以 `docs/workflow/status.md` 的当前 Sprint/phase 为准，再用 `knowledge/05-current-focus.md` 判断稳定主线。
- 当前主工作树的用户脏改已从旧的 account-modal 边界切回“仅保留未跟踪 `outputs/`”；并发 `S249/QA` worktree、detached `tutorial-nav-20260817` 与 `backup/pre-reorg-s240-s243-20260823` 仍需保留，开始新任务前先执行 `git status --short` 复核边界。
- 当前默认边界是：`origin/main` 与本地 `main` 已同步，且包含 `50ddfcc0b` 的 Pixel Cafe S252-S256 功能链；不触碰 `outputs/`，也不清理并发 S249/S258、教程导航和备份引用。继续新上游切片、发布或数据库影响改动时仍先取得相应授权。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 当前默认心智

- 当前最靠前的默认主线已经前移到 `S248 Google One 模型目录收口 + Pixel Cafe 账号选择器/房间详情批次 + S244-S248 清理完成`。继续接手 Gemini/Google One 账号能力、管理员可见模型目录、Pixel Cafe 房间管理或并发 worktree 清理相关工作时，不应再把仓库先理解成 8 月 20 日的 `S234` 阻塞仓。
- `S248` 已把 legacy Gemini Google One OAuth 账号的默认/可见模型目录收口到 2.0 Flash、2.5 Flash、2.5 Pro，同时保留显式模型映射与其他 Gemini/Antigravity 账号类型原语义；对应 focused x10、完整 geminicli/admin-handler/service、server compile、scope/provenance 与保护门禁均已通过。
- 受保护的 Pixel Cafe 批次已把账号选择器与房间详情并入当前主线；后续排查 Cafe 房间前后端行为时，应把它视为当前稳定产品面，而不是仍停在更早的静态 Lobby 或旧支付阶段。
- 当前 workflow phase 仍是 `done`，表示 S248、S244-S259 与 S252-S256 已完成本地验收并普通发布；不是“允许继续随手 push/publish”或“允许动共享数据库/生产环境”。新的上游 commit/tag 仍需先走 contract，数据库影响只在用户明确授权后继续。
- 当前工作树保护边界也已变化：主工作区默认只保留 `outputs/`，并发 `S249/QA` worktree、detached `tutorial-nav-20260817` 与备份引用必须保留；后续主线操作前后都要验证这些边界未被误清理。
- `S244-S247`、`S220-S222`、`S219`、`S218`、`S217`、Studio Bridge / 落叶AI、早期 Pixel Cafe、排行榜与更早的 safe patches 仍然成立，但都已退成这轮 `S248` 和并发 `S249` 之前的稳定背景层。

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
