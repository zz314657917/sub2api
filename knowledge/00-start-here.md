# Sub2API 知识库入口

最后更新：2026-08-12

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

- 当前默认续做主线已经从 2026-08-03 的 `Usage S135-S138` / `Pixel Cafe S139+` 继续前移到 2026-08-12 的 `S211 标准分组时段倍率` 与 `S212 账号时段可用性`；较早的 Pixel Cafe、Usage、`S111/S112` 和并行 `group-buy` 只保留为背景层。
- 现在判断“仓库在做什么”，先看 `docs/workflow/status.md` 和 `knowledge/tasks/current-task.md`，再用 `knowledge/tasks/timeline.md` 补足最近阶段历史。
- `knowledge/tasks/current-task.md` 现在应优先承载 `S211/S212` 的会话快照；如果入口摘要与 workflow 文档冲突，先以 `docs/workflow/status.md` 的当前 Sprint/phase 为准，再用 `knowledge/05-current-focus.md` 判断稳定主线。
- 当前主工作树存在新的未提交实现面，已不再是 8 月初那批 Pixel Cafe / group-buy dirt：本轮高频改动集中在 `backend/internal/service/**`、`backend/internal/handler/**`、`frontend/src/components/account/**`、`frontend/src/views/admin/GroupsView.vue` 和 `docs/workflow/**`。开始新任务前先执行 `git status --short`，确认哪些文件属于本轮目标。
- `S211/S212` 的默认边界是：不 push、不部署、不更新容器、不触碰 `outputs/`；如需继续 source 收口或提交，只能在用户明确授权后单独执行。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 当前默认心智

- 当前最靠前的默认主线已经前移到 `S211 标准分组时段倍率 + S212 账号时段可用性`。继续接手分组、账号、调度、倍率、可用性或管理端弹窗相关工作时，不应再把仓库先理解成 8 月 3 日的 Usage / Pixel Cafe 收口仓。
- `S211` 已把 standard group 的 `peak_rate_*` 语义扩展到同日时段倍率；`S212` 已把账号级可用时间窗收口到 `accounts.extra` 与调度排除逻辑。这两条是当前账号/分组/调度默认背景，优先级高于更早的 Pixel Cafe 和排行榜语境。
- `S210`、`S209`、`S208`、`S207` 已退成上一层稳定工程背景，但仍直接影响网关稳态：streaming terminal audit、API key 输入校验、stream route cooldown 和记分/availability 小步上游合成都已进入当前后端基线。
- 当前 workflow phase 仍是 `done`，但这只表示本地 source / QA 收口完成；`S212` 明确未执行 push、部署、容器更新、共享资源变更和真实生产流量验证，不能把入口知识写成“已经发布”或“线上已验证”。
- 当前浏览器证据边界也需显式保留：`S211` 有本地桌面和 390px 视觉证据；`S212` 自动化和独立 QA 已通过，但 Vite-only 环境因后端 `500` / `ECONNREFUSED` 未取得认证后的账号弹窗视觉证据，不能误写成完整视觉 PASS。
- Usage、Pixel Cafe、排行榜、Studio Bridge / 落叶AI、暖白前端统一、共享账号渠道状态可见性、首充 bonus only、以及更早的上游 safe patches 仍然成立，但都已退成当前 `S211/S212` 之前的稳定背景层。

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
