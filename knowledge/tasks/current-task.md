# 当前任务快照

最后更新：2026-05-29 03:50 +08:00

## 背景

- 项目主仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/sub2api-studio-layout`。
- 用户要求：继续把旧版 `chatgpt2api /image` 生图能力迁入 sub2api，并起多个智能体并行开发迁移。
- 本轮明确范围：只使用 sub2api 用户体系，不迁旧账号/RBAC；不做公开图库、发布/取消公开或 visibility/share 字段。

## 当前目标

- 补齐图片库高级筛选、图片库参考图复用、提示词市场/收藏、生图存储治理，以及 Canvas 的后端/API 和前端工作台骨架。
- 保持迁移分批提交，避免一个不可验收的大提交。

## 本次已完成

- 图片库增强已提交：扩展当前用户图片列表筛选、返回宽高/比例/格式/任务元数据，`/image-manager` 增加搜索、日期、比例、格式、分辨率等筛选。
- 参考图复用已提交：图片库“用作参考图”跳转 `/chat-images?mode=image&reference_image_id=...`，`/chat-images` 通过后端校验后的引用文件接口加载 reference image。
- 提示词市场/收藏已提交：迁移 `banana-prompt-quicker`、`awesome-gpt-image-2-prompts` 拉取/搜索/预览/套用，并新增当前用户收藏 API `GET/POST/DELETE /api/v1/user/prompt-favorites`。
- 生图存储治理已提交：新增管理员 API `GET/POST /api/v1/admin/image-creator/storage-governance`，统计图片、过期图片、孤儿文件、预览缓存、缩略缓存，并支持清理动作。
- Canvas 后端已提交：新增 `canvas_documents`、`canvas_nodes`、`canvas_edges`、`canvas_runs` 表和 `/api/v1/user/canvases`、`/api/v1/user/canvas-runs`、`/api/v1/user/canvas/models` API。
- Canvas 前端已提交：新增 `/canvas`、侧边栏入口、Canvas API 适配层、基础节点画布、节点列表、模型选择、保存、运行队列/历史骨架、模板入口占位。

## 已确认事实

- 当前提交序列包含：
  - `47e0b5489 feat(images): enhance image library filters`
  - `b03e09354 feat(images): add prompt market favorites`
  - `d810a93bf feat(canvas): add backend canvas and storage governance APIs`
  - `ce961c84a feat(canvas): add canvas workspace UI`
- Canvas 当前是可保存/打开/排队记录的工作台骨架；真实图像运行引擎、节点拖拽连线编辑器、模板库和高级图像编辑还未完整实现。
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
- 动作：打开 `/canvas`。
  验证：能新建/保存/打开 Canvas，添加/删除节点，提交运行记录；确认当前只是骨架，不包含完整旧版 Canvas 编辑能力。

## 当前结论

- 本轮迁移的图片库、参考图复用、提示词市场/收藏、存储治理和 Canvas 基础能力已经完成代码实现并通过自动化验证。
- 完整旧版 Canvas 能力尚未全部迁完，下一阶段应继续做真实运行编排、节点交互编辑、模板和高级图像编辑。

## 下一步

- 动作：实现 Canvas 真实运行链路，接入 sub2api API Key、模型目录、计费、并发和图片任务服务。
  验证：`go test ./internal/service ./internal/handler ./internal/repository -run "Canvas" -count=1`，并手动运行文生图/图生图流程。
- 动作：补 Canvas 前端交互编辑能力，包括拖拽、连线、节点参数面板、运行结果回填。
  验证：`npm.cmd run test:run -- CanvasView`，并用浏览器检查桌面/移动端布局。
- 动作：迁移 Canvas 图像编辑能力，包括裁剪、外扩、mask、历史和模板。
  验证：用旧版典型流程逐项对照验收。

## 验证记录

- `go test ./internal/service ./internal/handler -run "ImageCreator" -count=1`，通过。
- `npm.cmd run test:run -- ImageManagerView ChatImageStudioView public-smoke AppSidebar`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "ImageCreator|PromptFavorite" -count=1`，通过。
- `npm.cmd run test:run -- promptMarket ChatImageStudioView`，通过。
- `go test ./internal/service ./internal/handler ./internal/repository -run "ImageCreator|PromptFavorite|Canvas" -count=1`，通过。
- `go test ./cmd/server -count=1`，通过。
- `npm.cmd run test:run -- canvas CanvasView AppSidebar public-smoke`，通过。
- `go test ./...`，通过。
- `npm.cmd run lint:check`，通过。
- `npm.cmd run build`，通过；仅有既有 Vite dynamic import/chunk size 警告。
- `git diff --check`，通过。
