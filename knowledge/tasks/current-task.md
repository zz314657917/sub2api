# 当前任务快照

最后更新：2026-05-29 10:55 +08:00

## 背景

- 项目主仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/sub2api-studio-layout`。
- 用户要求：继续把旧版 `chatgpt2api /image` 生图能力迁入 sub2api，并起多个智能体并行开发迁移。
- 本轮明确范围：只使用 sub2api 用户体系，不迁旧账号/RBAC；不做公开图库、发布/取消公开或 visibility/share 字段。

## 当前目标

- 继续补齐图片库高级筛选、图片库参考图复用、提示词市场/收藏、生图存储治理，以及 Canvas 的后端/API 和前端工作台骨架。
- 本轮聚焦 Canvas 核心可用性：Canvas run 取消、运行队列状态、节点拖拽、连线创建/删除、画布缩放/平移/适配视图。
- 保持迁移分批提交，避免一个不可验收的大提交。

## 本次已完成

- 图片库增强已提交：扩展当前用户图片列表筛选、返回宽高/比例/格式/任务元数据，`/image-manager` 增加搜索、日期、比例、格式、分辨率等筛选。
- 参考图复用已提交：图片库“用作参考图”跳转 `/chat-images?mode=image&reference_image_id=...`，`/chat-images` 通过后端校验后的引用文件接口加载 reference image。
- 提示词市场/收藏已提交：迁移 `banana-prompt-quicker`、`awesome-gpt-image-2-prompts` 拉取/搜索/预览/套用，并新增当前用户收藏 API `GET/POST/DELETE /api/v1/user/prompt-favorites`。
- 生图存储治理已提交：新增管理员 API `GET/POST /api/v1/admin/image-creator/storage-governance`，统计图片、过期图片、孤儿文件、预览缓存、缩略缓存，并支持清理动作。
- Canvas 后端已提交：新增 `canvas_documents`、`canvas_nodes`、`canvas_edges`、`canvas_runs` 表和 `/api/v1/user/canvases`、`/api/v1/user/canvas-runs`、`/api/v1/user/canvas/models` API。
- Canvas 前端已提交：新增 `/canvas`、侧边栏入口、Canvas API 适配层、基础节点画布、节点列表、模型选择、保存、运行队列/历史骨架、模板入口占位。
- Canvas 运行链路已完成：CanvasService 注入 ImageCreatorService，`CreateRun` 会解析 `text_to_image` / `image_to_image` 节点，创建现有 ImageCreator task，并把 node -> task 映射写入 `canvas_runs.output`。
- Canvas 前端交互已增强：`/canvas` 支持选择可用 API Key、编辑节点 `prompt/text/model/size/quality/referenceImageId`，运行前保存画布，并展示最近运行状态、节点结果摘要/缩略和错误摘要。
- Canvas 运行结果回填已完成：`/canvas` 会解析 `canvas_runs.output.image_tasks`，调用现有 `/user/image-creator/tasks/:id` 轮询任务，并在节点上展示排队/运行/成功/失败、图片预览和错误信息；结果仅作为展示层 overlay，不自动写回 Canvas 文档。
- Canvas 核心编辑已完成：`/canvas` 支持节点拖拽、节点选择、边选择、连线创建、重复边抑制、删除边、删除节点时清理相关边、缩放、滚轮缩放、平移和适配视图；保存 payload 会带上节点坐标、边和 viewport。
- Canvas run 取消已完成：前端 API 新增 `cancelCanvasRun`，`CanvasRun` 映射 `canceled_at`；运行队列对 queued/running/pending run 显示取消按钮，取消后本地队列状态变为 canceled。取消仅作用 Canvas run，不级联取消 ImageCreator task。
- P/G/E 多智能体执行已完成：Worker A 负责 run cancel/API-client/后端测试，Worker B 负责前端编辑器，QA Worker 独立验收并写入 `docs/workflow/qa-reports/sub2api-canvas-core-qa.md`。

## 已确认事实

- 当前提交序列包含：
  - `47e0b5489 feat(images): enhance image library filters`
  - `b03e09354 feat(images): add prompt market favorites`
  - `d810a93bf feat(canvas): add backend canvas and storage governance APIs`
  - `ce961c84a feat(canvas): add canvas workspace UI`
  - `c0f0c8af9 feat(canvas): queue image creator tasks from canvas runs`
  - `403baa41c docs(workflow): record canvas run chain progress`
  - `b8dec861e feat(canvas): poll image task results`
- Canvas 当前可把图片节点运行提交到现有 ImageCreator task；图片实际生成、API Key 校验、分组生图权限、并发限制、gateway 计费和图片保存继续复用 ImageCreator 链路。
- Canvas 前端现在可轮询 ImageCreator task 结果并回填展示；为了避免隐藏持久化，轮询结果不修改保存 payload 中的 Canvas 文档。
- Canvas 还不是完整旧版节点执行引擎：取消 run 不会级联取消 ImageCreator task，模板库和高级图像编辑还未完整实现。
- 存储治理对象存储场景仅支持过期图片走既有删除逻辑；本地孤儿文件/缓存遍历在对象存储后端返回 `unsupported`。
- `wire_gen.go` 因仓库既有 Wire 生成阻塞，仍采用手工同步方式；`go test ./cmd/server` 已验证当前注入可编译。

## 待验证点

- 动作：用真实登录态打开 `/image-manager`。
  验证：筛选、搜索、加载更多、删除、下载、复制提示词、继续创作、用作参考图都按当前用户隔离。
- 动作：从 `/image-manager` 选择“用作参考图”进入 `/chat-images`。
  验证：参考图自动带入，能发起图生图，且无法跨用户引用图片。
- 动作：打开 `/chat-images` 提示词市场。
  验证：市场搜索、收藏/取消收藏、套用提示词、带参考图套用均正常。
- 动作：管理员打开设置页的“生图存储治理”卡片。
  验证：统计展示正确，清理动作有结果反馈；对象存储本地遍历动作显示不支持。
- 动作：用真实登录态打开 `/canvas`。
  验证：能新建/保存/打开 Canvas，选择 API Key，编辑节点参数，拖拽节点，创建/删除连线，缩放/平移/适配视图，提交运行后生成 ImageCreator task 映射，随后节点能显示 ImageCreator task 状态、预览图和失败错误；取消 queued/running run 后刷新仍为 canceled。

## 当前结论

- 本轮迁移的图片库、参考图复用、提示词市场/收藏、存储治理、Canvas 基础能力、Canvas 运行最小闭环、Canvas 结果轮询回填、Canvas 核心编辑和 Canvas run 取消已经完成代码实现并通过自动化验证。
- 完整旧版 Canvas 能力尚未全部迁完，下一阶段应继续做模板库和高级图像编辑。

## 下一步

- 动作：迁移 Canvas 模板库。
  验证：模板列表、创建模板、从模板创建 Canvas、一键套用节点结构。
- 动作：迁移 Canvas 图像编辑能力，包括裁剪、外扩、mask、历史。
  验证：用旧版典型流程逐项对照验收。

## 验证记录

- `go test ./internal/service ./internal/handler -run "ImageCreator" -count=1`，通过。
- `npm.cmd run test:run -- ImageManagerView ChatImageStudioView public-smoke AppSidebar`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "ImageCreator|PromptFavorite" -count=1`，通过。
- `npm.cmd run test:run -- promptMarket ChatImageStudioView`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "ImageCreator|PromptFavorite|Canvas" -count=1`，通过。
- `go test ./cmd/server -count=1`，通过。
- `npm.cmd run test:run -- canvas CanvasView AppSidebar public-smoke`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "ImageCreator|Canvas" -count=1`，通过。
- `go test ./cmd/server -count=1`，通过。
- `npm.cmd run test:run -- CanvasView canvas AppSidebar public-smoke`，通过。
- `go test ./...`，通过。
- `npm.cmd run lint:check`，通过。
- `npm.cmd run build`，通过；仅有既有 Vite dynamic import/chunk size 警告。
- `git diff --check`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "Canvas|ImageCreator" -count=1`，通过。
- `go test ./cmd/server -count=1`，通过。
- `npm.cmd run test:run -- CanvasView canvas`，通过，14 个测试。
- `npm.cmd run lint:check`，通过。
- `npm.cmd run build`，通过；仅有既有 Vite dynamic import/chunk size 和 Node DEP0190 警告。
- `git diff --check`，通过。
- 浏览器 smoke：`http://127.0.0.1:62080/canvas` 可加载到受保护路由，当前本地会话显示登录页且无前端 error log；未用真实登录态做 UI 手测。
- QA Worker 报告：`docs/workflow/qa-reports/sub2api-canvas-core-qa.md` 首行 `### PASS: sub2api-canvas-core`。
- `npm.cmd run test:run -- CanvasView canvas`，通过。
- `npm.cmd run lint:check`，通过。
- `npm.cmd run build`，通过；仅有既有 Vite dynamic import/chunk size 警告。
- `git diff --check`，通过。
