# Sub2API 知识库入口

最后更新：2026-08-06

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

- 当前默认续做主线已经继续前移到 2026-08-03 的 `Usage S135-S138` 与 `Pixel Cafe S139+`；`S111/S112` 和并行 `group-buy` 已退成更早一层背景。
- 现在判断“仓库在做什么”，先看 `docs/workflow/status.md` 和 `knowledge/tasks/current-task.md`，再用 `knowledge/tasks/timeline.md` 补足最近阶段历史。
- `knowledge/tasks/current-task.md` 应优先记录“现在默认从哪条主线继续”，不再适合停留在 2026-07-12 的 release push / fast-forward / cleanup 清单。
- 遇到入口摘要与 workflow 文档冲突时，先以 `docs/workflow/status.md` 的当前 Sprint/phase 为准，再用 `knowledge/05-current-focus.md` 判断稳定主线，用 `knowledge/tasks/current-task.md` 判断当前会话快照。
- 当前工作区仍存在并行 `group-buy` dirt：主工作树有 `frontend/src/views/admin/group-buy/AdminGroupBuyView.vue` 和对应测试目录未提交，且另有 `codex/group-buy-lifecycle-refund-hardening-s110` 独立工作树；开始新任务前先执行 `git status --short`，确认哪些文件属于本轮目标。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 当前默认心智

- 当前最靠前的用户面变化已经前移到 2026-08-03 的 Usage 与 Pixel Cafe 多阶段收口；`S111` 只作为更早背景保留。
- 当前最靠前的后端兼容变化已经前移到 `S112`，但它已经退成背景层；后续继续做设置、网关、账号能力或 UI 时要先看最新 workflow 状态。
- `S106-S109`、独立 Agent Identity S108、`S76-S81`、`S77` 和排行榜按小时刷新都已进入更早一层的稳定背景；继续排查设置、网关、账号能力或 UI 时不能回退到 7 月 20 日之前的 `S82-S86 publish pending` 心智。
- 当前 workflow phase 是 `done`；`S111` 已 `PASS / published`，`S112` 是 `PASS / source-only` 但功能提交和发布收口都已在 `origin/main`。这仍不等于已部署、已更新容器或已做真实登录态/上游 smoke。
- Studio Bridge / 落叶AI、暖白前端统一、共享账号渠道状态可见性、首充 bonus only、以及更早的上游 safe patches 仍然成立，但它们已经退成当前集成线之前的稳定背景层。
- 继续做聊天生图、嵌入工作区、模型市场、OpenAI/Codex 网关兼容或排行榜相关工作时，不要只看单个页面或单个 Sprint；通常要把最新 workflow 状态、默认主线、并行 `group-buy` 边界和旧的产品背景一起看成一条连续链路。

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
