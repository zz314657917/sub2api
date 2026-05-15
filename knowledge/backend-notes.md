# 后端开发笔记

最后更新：2026-05-13

## 分层约定

- `cmd/server` 负责启动和依赖注入，不放业务逻辑。
- `server/routes` 负责路由注册和中间件组合。
- `handler` 负责 HTTP 接入、参数校验、认证上下文、状态码和 DTO。
- `service` 承载业务规则、调度、计费、支付和 provider 兼容逻辑。
- `repository` 负责 Ent、SQL、缓存、上游 HTTP client 和数据聚合。
- 外部 provider 或第三方系统逻辑尽量集中在 service/repository/integration 边界，不散落到 handler。

## 常见修改链路

新增或调整用户接口：

1. `backend/internal/server/routes/user.go`
2. `backend/internal/handler/*`
3. `backend/internal/handler/dto/*`
4. `backend/internal/service/*`
5. `backend/internal/repository/*`
6. 前端 `frontend/src/api/user.ts` 或对应 API client
7. 前端类型、页面和测试

新增或调整管理端设置：

1. `backend/internal/handler/admin/setting_handler.go`
2. `backend/internal/handler/dto/settings.go`
3. `backend/internal/service/setting_service.go`
4. `backend/internal/service/settings_view.go`
5. 前端 `frontend/src/api/admin/settings.ts`
6. `frontend/src/views/admin/SettingsView.vue`
7. 对应 service、handler、前端测试

账号调度/共享池相关：

- 用户路由集中在 `/user/accounts`。
- handler：`backend/internal/handler/user_account_handler.go`
- service：`backend/internal/service/user_account_service.go`、`account*` 相关文件。
- repository：`backend/internal/repository/account_repo.go` 及 usage/ledger 相关测试。
- 前端页面通常在 `frontend/src/views/user/MyAccountsView.vue`、dashboard、channel status 或 API client。

OpenAI / Codex / provider 兼容：

- Gateway handler：`gateway_handler.go`、`openai_gateway_handler.go`、`openai_chat_completions.go`、`openai_images.go`。
- 账号与模型映射：`account.go`、`account_service.go`、`account_test_service_*`、provider-specific service。
- 测试通常覆盖 endpoint normalization、stream failover、quota、passthrough、image controls。

图片生成：

- 用户路由在 `/user/image-creator`。
- handler：`image_creator_handler.go`
- service：`image_creator_service.go`
- repository：`image_creator_repo.go`
- 前端：`frontend/src/views/user/ImageCreatorView.vue`、`frontend/src/api/imageCreator.ts`

## 测试策略

- 单个 service/handler 改动优先跑关联测试文件。
- 改共享接口或 DTO 时补 API contract 或前后端测试。
- 改 interface 时搜索所有 stub/mock，补全新方法。
- 改 migration 或 Ent schema 时跑生成和相关 repository 测试。
- 改 provider 调度时至少覆盖成功、不可调度、quota/limit、错误透传和 failover。

## 风险点

- Go interface 新增方法会影响大量测试 stub。
- 模型映射和账号批量更新容易跨 provider 误伤，特别是 OpenAI/Codex 与 Gemini/Antigravity 混选。
- 计费、余额、共享池、排行榜奖励属于资金或额度相关逻辑，改动前后都要补测试证据。
- 中间件和 backend mode guard 影响面大，不能只看单个 handler。
