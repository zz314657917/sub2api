# Sub2API 知识库入口

最后更新：2026-05-15

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
8. 改聊天生图工作台或图片任务体验：读 `knowledge/chat-image-studio.md`。
9. 遇到老坑或环境差异：读 `knowledge/known-pitfalls.md`。

## 事实源分工

- `README*.md`：面向用户和部署者的产品说明。
- `DEV_GUIDE.md`：本地开发环境、CI、历史坑点和命令速查。
- `docs/`：支付等专题文档；目前 `.gitignore` 默认忽略多数 `docs/*`，只有少数文档例外。
- `knowledge/`：面向 AI/开发协作的项目知识入口和长期结论。
- `knowledge/tasks/current-task.md`：当前工作快照，只保留“现在做到哪、下一步是什么”。
- `knowledge/tasks/timeline.md`：阶段历史和关键决策，按倒序追加。

## 当前仓库状态提示

- 截至 2026-05-13，工作区存在大量未提交业务改动，涉及后端账号/设置/图片生成服务，以及前端公共页、控制台、客服弹窗、ChatStudio 等。
- 截至 2026-05-15，聊天生图主线已转为 Sub2API 原生 `/chat-images` 工作台；会话删除与后端图片任务的关系见 `knowledge/chat-image-studio.md`。
- 新任务开始前先执行 `git status --short`，确认当前改动是否属于本轮任务。
- 不要清理、回滚或格式化与当前目标无关的文件。

## 知识维护规则

- 长期稳定、后续会复用的项目结论写入 `knowledge/`。
- 临时进度写 `knowledge/tasks/current-task.md`。
- 阶段归档写 `knowledge/tasks/timeline.md`。
- 不写入密钥、token、账号、私有地址或未标注的猜测。
- 写文档后用 UTF-8 回读，避免中文乱码。
