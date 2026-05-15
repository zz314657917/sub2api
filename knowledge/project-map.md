# 项目地图

最后更新：2026-05-13

## 顶层目录

- `backend/`：Go 后端，包含 HTTP handler、service、repository、Ent schema、迁移和 server wiring。
- `frontend/`：Vue 3 前端，包含公共页面、用户控制台、管理后台、API client、状态管理、i18n 和测试。
- `deploy/`：Docker、systemd、一键安装脚本、配置模板和部署说明。
- `docs/`：支付、后台支付集成等专题文档。
- `assets/`：README、赞助商、品牌相关静态资产。
- `tools/`：本地开发辅助脚本，例如 dev 清理脚本。
- `knowledge/`：项目协作知识库。
- `.gocache/`、`.venv/`、`output/`：本地缓存或输出，搜索时通常要排除。

## 后端结构

- `backend/cmd/server/`：主程序入口、版本、Wire 依赖注入生成文件。
- `backend/cmd/jwtgen/`：JWT 辅助命令。
- `backend/ent/`：Ent ORM 生成代码和 schema。
- `backend/migrations/`：数据库迁移。
- `backend/internal/server/`：Gin server、路由注册和 API contract 测试。
- `backend/internal/handler/`：HTTP 接入层，负责参数、认证上下文、响应和 gateway 接入。
- `backend/internal/handler/admin/`：管理端 handler。
- `backend/internal/handler/dto/`：请求/响应 DTO。
- `backend/internal/service/`：核心业务，账号调度、认证、计费、支付、图片生成、OpenAI/Anthropic/Gemini/Antigravity 等能力主要在这里。
- `backend/internal/repository/`：数据访问、缓存、迁移执行、上游 HTTP client 和聚合查询。
- `backend/internal/integration/`：外部系统或 provider 集成边界。
- `backend/internal/middleware/`：认证、模式保护、请求中间件。
- `backend/internal/testutil/`：测试辅助。

## 前端结构

- `frontend/src/router/`：路由、导航守卫、标题解析。
- `frontend/src/api/`：后端 API client；管理端 API 在 `frontend/src/api/admin/`。
- `frontend/src/views/public/`：公共页面，例如 `/home`、`/tutorial`、`/models`、法律文档。
- `frontend/src/views/auth/`：登录、注册、OAuth/OIDC/微信/LinuxDo 回调。
- `frontend/src/views/user/`：用户控制台页面，例如 dashboard、keys、usage、payment、chat、image creator。
- `frontend/src/views/admin/`：管理后台页面和设置。
- `frontend/src/components/layout/`：控制台布局、顶部栏、侧栏、表格页布局。
- `frontend/src/components/common/`：通用组件，例如客服弹窗、表格等。
- `frontend/src/stores/`：Pinia stores。
- `frontend/src/i18n/locales/`：中英文文案。
- `frontend/src/styles/` 和 `frontend/src/style.css`：全局样式、控制台 UI 样式。
- `frontend/src/__tests__/`：跨页面或主题类测试。

## 关键产品域

- 认证与身份绑定：邮箱、GitHub/Google/OIDC/LinuxDo/微信等 OAuth 链路，主要在 `auth_*` handler/service、auth store 和回调页面。
- 账号与共享池：管理员账号、用户账号共享池、容量池、使用汇总，重点看 `user_account_*`、`account_*` 和 `/user/accounts` 路由。
- Gateway 与兼容：OpenAI、Anthropic、Gemini、Antigravity、Bedrock 等请求转发和模型映射，重点看 `gateway_*`、`openai_*`、provider service。
- 计费与用量：API Key、usage、pricing、billing cache、dashboard 聚合和排行榜奖励。
- 支付：EasyPay、支付宝、微信、Stripe、Airwallex 相关 handler/service/frontend views。
- 公共页：公共导航、Matrix 背景、教程页、模型广场和首页一致性。
- 管理设置：`setting_service`、`settings_view`、admin settings API 和 `SettingsView.vue`。

## 搜索建议

- 文本搜索优先 `rg`。
- 全仓库搜索时排除 `.git`、`.gocache`、`.venv`、`output`、`frontend/node_modules`。
- 查接口先从路由开始，再顺 handler -> service -> repository。
- 查前端页面先从 `frontend/src/router/index.ts` 确认路由，再查 view、API client、store 和测试。
