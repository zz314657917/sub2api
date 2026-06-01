# 当前任务快照

最后更新：2026-06-01 21:20 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 本轮任务围绕 API Key 创建弹窗的智能路由模式，以及后续“模型感知智能分组路由”运行时能力。
- 工作区有多条并行未提交改动，包括 support ticket、ent 生成文件、workflow 文档、用户/福利/支付等；本轮只处理 Key 智能路由相关前后端文件，不回滚其他改动。

## 当前目标

- 一个 API Key 能按请求模型和生图意图自动选择分组，不要求用户为每类模型创建专用 Key。
- 前端保留模式卡片，自动生成 `multi_group_routes`；高级手动路由可配置模型匹配、仅生图、排除生图。
- 后端不新增数据库列，扩展现有 `multi_group_routes` JSON，兼容旧 Key。

## 本次已完成

- 后端 `APIKeyMultiGroupRoute` 增加 `model_patterns`、`image_only`、`text_only` JSON 字段。
- 后端路由 resolver 支持模型 pattern、图片意图、`image_only` / `text_only`、旧 priority/weight fallback。
- OpenAI/Anthropic/Gemini 相关 handler 在解析 model/imageIntent 后执行二次分组解析并更新 `api_key` 上下文。
- `/v1/messages`、`/v1/messages/count_tokens`、responses、chat completions、embeddings、images 等路由分发前预读 JSON model，避免先进入错误平台 handler。
- 自检智能体发现的问题已处理：
  - 延后模型感知端点的分组级订阅/余额校验，避免默认分组在最终路由前误拦截。
  - `Abort` 后显式 `return`，避免错误响应后继续进入 handler。
  - body 预读失败时直接 abort，`MaxBytesError` 返回 413。
  - 同一 group 多 scope 冷却改为取该 group 所有启用 route 的最大 cooldown。
  - `messages/count_tokens` 增加模型感知二次路由。
- 前端 `KeysView.vue` 智能自动、价格优先、速度优先、成功率优先生成通用/生图规则；手动路由支持模型匹配和生图开关。
- 前端类型和中英文 keys 文案已同步。

## 已确认事实

- 生图路由只允许 OpenAI 平台且 `allow_image_generation=true` 的分组参与。
- 生图价格优先在有独立生图倍率时按 `image_rate_multiplier` 排序，否则按通用有效倍率排序。
- 旧 `multi_group_routes` 缺少新字段时按零值处理，不会强制进入模型规则。
- 前端允许同一 group 按 `text_only` / `image_only` 拆成多条 route；后端 normalize/validate 已按 scope 兼容。

## 待验证点

- 真实登录态下打开 API Key 创建/编辑弹窗，检查桌面和窄屏布局、底部提交按钮固定、手动路由展开区域不溢出。
- 用实际账号池 smoke：同一 Key 调用 `/v1/responses` 文本模型、生图模型、`/v1/images/generations`、`/v1/messages/count_tokens`，确认最终命中的分组和计费上下文符合规则。
- 如果未来要精确到“命中 route 的 cooldown”，需要把 resolver 返回值从 group 扩展为 route+group；本轮先按 group 统一冷却。

## 当前结论

- 模型感知智能分组路由已实现并通过相关自动化验证。
- 自检智能体指出的 P1/P2/P3 问题均已修复或以兼容方案收口。
- 当前未做数据库迁移，API 请求格式保持兼容。

## 下一步

- 视觉验证：登录本地前端后打开 API Key 创建弹窗 -> 验证创建/编辑模式、智能模式卡片、手动路由模型规则输入在桌面和移动窄屏均正常。
- 运行时 smoke：准备至少两个可用分组，一个通用文本组和一个允许生图的 OpenAI 组 -> 使用同一 API Key 分别请求文本和生图模型 -> 验证日志中的 group_id 分别命中预期分组。
- 提交前复核：用 `git diff --stat` 和 `git diff --check` 确认只纳入本轮相关文件，避免混入并行 support ticket / ent / workflow 改动。

## 验证记录

- `go test ./internal/server/routes ./internal/server/middleware ./internal/service ./internal/handler`：通过。
- `npm.cmd run test:run -- KeysView`：通过，`KeysView.createQuery.spec.ts` 10 个测试通过。
- `npm.cmd run build`：通过；仅有既有 Browserslist 过旧、Vite dynamic/static import、chunk size 警告。
- `git diff --check`：通过；仅有 docs/workflow 文件 LF/CRLF warning。
- 自检智能体 `019e833a-f7ff-72a0-8a60-2c5a6b74d67c` 已完成并关闭。
