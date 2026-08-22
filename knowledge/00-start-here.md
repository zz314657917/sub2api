# Sub2API 知识库入口

最后更新：2026-08-20

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

- 当前默认续做主线已从 `S220/S221/S222` 继续前移到 workflow status 中的 `S234 upstream-v178-ui-polish`；S234 当前因前端依赖环境阻塞，未授权业务提交或推送。S220/S221/S222 仍是已合入的稳定后端基线。
- 现在判断“仓库在做什么”，先看 `docs/workflow/status.md` 和 `knowledge/tasks/current-task.md`，再用 `knowledge/tasks/timeline.md` 补足最近阶段历史。
- `knowledge/tasks/current-task.md` 现在应优先承载 S234 的阻塞快照；如果入口摘要与 workflow 文档冲突，先以 `docs/workflow/status.md` 的当前 Sprint/phase 为准，再用 `knowledge/05-current-focus.md` 判断稳定主线。
- 当前主工作树剩余的用户未提交内容已经收敛到 `frontend/src/components/account/EditAccountModal.vue`、对应测试和 `outputs/`；开始新任务前先执行 `git status --short`，确认是否仍保持这组边界。
- 当前默认边界是：`origin/main` 已与本地 `main` 同步；不继续 push、不部署、不更新容器、不触碰 `outputs/`，也不覆盖用户 account-modal dirty。若要继续新的上游切片或数据库影响改动，需先取得相应授权。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 当前默认心智

- 当前最靠前的默认主线已经前移到 `S220 分组定价与长上下文账户 veto + S221 Codex fingerprint convergence + S222 分组用量日汇总`。继续接手分组定价、账号能力、OpenAI/Codex 调度、fingerprint、usage dashboard 或数据库聚合相关工作时，不应再把仓库先理解成 8 月 12 日的 `S211/S212` 收口仓。
- `S220` 已把分组 built-in pricing、group long-context 开关、OpenAI group/account intersection 和非 OpenAI Grok veto 修正并入当前后端基线；`S221` 已把 opt-in `codex_fingerprint_mode` 收口到本地 gateway/account 编辑链路；`S222` 已把分组日汇总、失效重算、timezone rebuild、DST 边界与 advisory lock 多副本排斥并入 dashboard 稳态。
- 当前 workflow phase 仍是 `done`，但这里表示 `v0.1.177` 已授权切片已本地验收并普通 push；不是“允许继续随手合更多上游”或“允许动共享数据库/生产环境”。新的上游 commit/tag 仍需先走 contract，数据库影响只在用户明确授权后继续。
- 当前 dirty 边界也需显式保留：用户未提交内容只剩 account-modal 两个文件和 `outputs/`，其 patch-id 需要在后续主线操作前后保持不变。
- `S219` turn-state、`S218` remote compaction v2、`S217` quota correctness、`S211/S212`、Studio Bridge / 落叶AI、Pixel Cafe、排行榜与更早的 safe patches 仍然成立，但都已退成这轮 `v0.1.177` 已授权切片之前的稳定背景层。

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
