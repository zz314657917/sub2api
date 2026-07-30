# 当前任务快照

最后更新：2026-07-27 20:32 +08:00

## 当前任务（S115-S119）

- 用户确认需要合入上游 `v0.1.165` 的两项功能：usage `session_id` 持久化和
  ChatGPT Live 实时网关。
- S115 已完成显式客户端会话标识提取、usage 全路径传播、单条/批量/扫描/DTO
  映射和迁移 `195_add_usage_log_session_id.sql`。
- S116 已完成 opt-in ChatGPT Live：SDP 创建、Codex 路由别名、sideband WebSocket、
  Redis call/controller 状态、并发租约、usage `request_type=live` 和管理端开关。
- S117 修复管理员设置局部更新覆盖未提交字段；S118 保留 Gemini pool 失败转移时的
  同账号重试资格；S119 保留客户端名为 `web_search` 的普通 function tool。
- 已按功能边界整理为四条本地提交：`90f2fad21`、`d1e37d7bc`、`9a5e8e7ab`、
  `4672daa0c`；尚未推送。
- 关联证据：
  `docs/workflow/qa-reports/upstream-v0165-usage-session-id-s115-qa.md`、
  `docs/workflow/qa-reports/upstream-v0165-chatgpt-live-s116-qa.md`。

## 背景

- v0.1.165 采用选择性适配；S114 的 Claude Opus 5、Grok Responses 清理和图片
  诊断改动已在当前工作树中，不能与本次 S115/S116 证据混淆。
- 工作树还存在用户并行的 group-buy、`knowledge/**`、`outputs/**` 等改动，
  以及独立 `codex/group-buy-lifecycle-refund-hardening-s110` 工作树；本任务未回滚、
  清理或混入这些内容。
- P/G/E 当前 phase 为 `done`，qa_mode 为 `runtime`；本轮已完成本地提交与整合，
  未推送、未部署、未更新容器。

## 当前目标

- 收口 S115-S119 的本地集成、QA 证据和会话交接，保持 ChatGPT Live 默认关闭，
  并保留真实运行态边界。

## 本次已完成

- `session_id` 支持显式请求头优先级、trim、控制字符拒绝和 255 Unicode 字符上限；
  不从 prompt、cache key、request hash 或 API-key ID 派生。
- usage 写入、批量/best-effort 写入、查询扫描和 DTO 已加入 nullable `session_id`。
- Live 仅允许标准 OpenAI OAuth；API key、PAT、Agent Identity 和非 OpenAI group 被拒绝。
- Live 的 `allow_live` 默认 false；租约续期失败会结束会话、写 usage 并释放并发槽，
  不会永久阻断同账号重试。
- Redis call round-trip 已保存加密 DeviceCheck attestation，Live 创建沿用本地 OpenAI
  内容审核门禁，避免 sideband 丢认证材料或新增入口绕过审核。
- Ent 生成、focused Go、全仓编译、前端 typecheck/1091 模块 build、gofmt 和 diff
  检查通过；contract 状态、status、main-log 和两份 QA 报告已更新。
- S117 设置回归、S118/S119 Gemini 回归、完整 Gemini 服务测试、整合后全仓仅编译、
  格式和冲突标记扫描均通过；S117-S119 的 workflow 合并冲突已按时间保留记录。

## 已确认事实

- `session_id` 是 usage 记录里的客户端关联标识，不是 ChatGPT Live 的专用字段。
- ChatGPT Live 是实时媒体/语音会话：客户端先提交 SDP，服务端向 ChatGPT Live 上游
  建立 call，再通过 sideband WebSocket 转发控制事件；它不改变普通 `/v1/responses`。
- Windows 下 `liveattestation` 明确返回 DeviceCheck unavailable，不伪造 macOS
  attestation；这是 fail-closed，不代表 Windows 可直接运行真实 Live。
- group 的 allow_live、usage request_type=live 和三条迁移均为向前兼容改动，旧 group
  默认为关闭，旧 usage 的 session_id 为空。
- S117-S119 已整合到本地 `main`，未混入 S114、group-buy、`knowledge/**` 或 `outputs/**`
  的并行改动；`origin/main` 仍停留在提交前基线。

## 待验证点

- macOS Apple Silicon + 官方 ChatGPT app 的 DeviceCheck attestation、真实 ChatGPT
  上游 SDP/sideband、断线重连和租约过期 smoke 未执行。
- Redis call 状态和 live lease 的运行态验证未执行；当前测试环境 Redis 指向
  `127.0.0.1:1`，完整路由套件不能作为运行态证据。
- 真实 PostgreSQL migration/usage 持久化、认证态浏览器、部署和容器刷新未执行。

## 当前结论

- `PASS / local-integrated`：S115-S119 源码、生成代码、定向测试、全仓编译和前端构建
  已完成，未发现明确源码级阻断问题。
- 可以保留在本地工作树；在 macOS/Redis/真实 OAuth 环境 smoke 之前，不宣称 ChatGPT
  Live 已上线。

## 下一步

1. 如需运行态验收，准备 macOS Apple Silicon、官方 ChatGPT app、Redis、真实 OAuth
   账号后执行 Live SDP/sideband/lease/usage smoke -> 验证：日志、Redis key 和 usage
   记录。
2. 如需发布，单独执行推送和远端 parity -> 验证：只推送四条已整理提交，不包含
   group-buy、knowledge、outputs 和 S110 工作树改动。
3. 如需容器或生产验证，另开授权任务 -> 验证：部署前后健康检查和回滚证据。

## 验证记录

- `go generate ./ent`、focused Go（Live/session/usage）、repository focused、路由
  定向测试、`go test ./... -run '^$'` 均通过。
- 前端 `corepack.cmd pnpm --dir frontend run typecheck` 通过；生产构建通过，1091
  modules transformed。
- `gofmt -d`（无输出）和 `git diff --check` 通过。
- 整合后：S117 设置回归、`TestGeminiPoolMode`、完整 `TestGemini`、
  `go test ./... -run '^$'`、格式和冲突标记扫描均通过。
- 未执行完整路由运行态、真实 Redis/PostgreSQL、真实 ChatGPT 上游、认证态浏览器、
  部署和容器刷新。

## 2026-07-29 公共教程快速接入向导收尾

- `/tutorial` 首页现只保留参考图式的快速接入向导：ChatGPT / Codex 优先的平台和终端切换，Base URL / 鉴权 / 协议 / 模型信息卡，五步命令示例，桌面端说明、cURL 示例和错误码排查（含 `CAPACITY` 官方算力不足）。无实际差异的教程模式控制已移除。原 CMS 目录、搜索和分类移至 `/tutorial?view=library`；详情路由和 Markdown 交互保持不变。
- 本轮只涉及 `frontend/src/views/public/TutorialView.vue` 与 `frontend/src/views/public/__tests__/TutorialView.spec.ts` 的教程改动；工作树仍保留其他用户并行修改，未清理、回滚或混入。
- 工作树预览当前运行在 `http://127.0.0.1:62081/`；`62080` 仍由现有 Docker 服务占用。桌面端首屏、`390x844` 移动端无横向溢出，首页不再渲染旧目录，`?view=library` 才显示目录；平台/终端切换已完成浏览器 DOM 验收；真实浏览器剪贴板权限未作为通过证据。
- 本轮验证：`corepack.cmd pnpm --dir frontend exec vitest run src/views/public/__tests__/TutorialView.spec.ts`（7/7）、`corepack.cmd pnpm --dir frontend run typecheck`、`corepack.cmd pnpm --dir frontend run build`（1099 modules transformed）、教程文件 `git diff --check` 均通过。

## 2026-07-29 公共教程快速接入后台配置（S125）

- `/admin/tutorials` 新增“快速接入配置”入口。管理员可加载、修改、保存或恢复完整的公开教程 JSON；其中平台名称、Base URL、鉴权/协议/模型提示、标题说明、桌面端说明、API 说明和错误码均可维护。配置仅接受纯文本和 HTTP(S) Base URL，不支持 HTML。
- 新增公开读取 `GET /api/v1/tutorials/quickstart-config`，以及管理员 `GET/PUT /api/v1/admin/tutorials/quickstart-config` 和 `POST /api/v1/admin/tutorials/quickstart-config/reset`。数据写入既有 settings 存储的 `quickstart_tutorial_config`，无需迁移。
- 前台 `/tutorial` 会读取公开配置，并将新的 Base URL 同步写入信息卡、Claude 环境变量、Codex `config.toml` 和 cURL 示例。若还没有单独保存配置，会优先使用已有 `api_base_url` 推导 Claude 根地址和 Codex `/v1` 地址；请求失败或无效配置则回退内置配置。
- Base URL 约定已按用户反馈统一为根地址：ChatGPT / Codex 和 Claude 都显示 `https://<domain>`，不再默认显示 `/v1`。当站点 `api_base_url` 是 `https://<domain>/v1` 时，快速接入会自动去除末尾 `/v1`；Codex 示例由根地址访问 `/responses`，与网关的裸 Responses 路由一致。已保存的自定义 URL 保持原值，管理员可在快速接入配置中自行修改。
- ChatGPT / Codex 的第 2 步已补充“下载 ChatGPT Desktop（Windows / macOS）”官方链接，指向 `https://developers.openai.com/codex/app#getting-started`；固定官方链接不暴露为后台富文本字段，Claude 分支不显示该链接。定向 Vue 测试 8/8 和前端 typecheck 均通过。
- 第 3 步现为“找到或创建 config.toml”：Windows 明确 `%USERPROFILE%\.codex\config.toml` 和资源管理器打开命令，macOS/Linux 明确 `~/.codex/config.toml` 及隐藏目录的进入方式；额外提示避免保存成 `config.toml.txt`。Claude 分支不再错误显示 `.codex` 路径，改为环境变量配置提示。定向 Vue 测试 8/8 和前端 typecheck 均通过。
- 验证：`go test ./internal/service -run "QuickstartTutorial" -count=1`、`go test ./... -run "^$"`、定向 Vitest（2 files / 9 tests）、`pnpm typecheck`、生产 build（1100 modules）、`git diff --check`、冲突标记扫描均通过。Playwright 已检查 `/tutorial` 的桌面和 `390x844` 移动布局，以及 Claude/终端切换；本地预览仍连接旧后端，因此只验证了内置回退路径。
- 未执行：运行新后端后的认证态后台保存/恢复、真实 settings 数据库存储、部署、容器更新、推送和生产浏览器验收。工作树仍有用户其他并行改动，本轮未清理、回滚或混入。
