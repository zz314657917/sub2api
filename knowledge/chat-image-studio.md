# 聊天生图工作台

最后更新：2026-06-09

## 入口与定位

- 默认用户入口：`/chat-images`，现在用于启动外部落叶创艺创作站。
- 跳转页主文件：`frontend/src/views/user/LuoyeAILaunchView.vue`。
- 原生工作台入口：`/chat-images/native`，作为 Sub2API 内部备用工作台保留。
- 原生工作台主文件：`frontend/src/views/user/ChatImageStudioView.vue`。
- 后端生图接口：`/user/image-creator/tasks`、`/user/image-creator/tasks/:id`、`/user/image-creator/images/:id/file`。
- 默认链路：用户进入 `/chat-images` 后，前端调用 `/api/v1/user/studio-bridge/launch` 生成一次性 `launch_token`，再跳转到落叶创艺站点配置的回跳地址。

## 会话与图片任务的关系

- 工作台“会话”是前端本地状态，使用 `localStorage` key `sub2api:chat-image-studio:v1` 保存。
- 会话最多保留 20 个，每个会话最多保留 120 条消息。
- 删除会话只会从前端 `sessions` 和 `localStorage` 删除对应聊天记录；它不是后端图片任务删除接口。
- 图片任务是后端持久化任务，存储在图片生成任务链路中；图片库和任务队列来自 `/user/image-creator/tasks` 的服务端列表。

## 正在生成时删除会话会怎样

按 2026-05-15 的实现：

- UI 层把 `busy = sending || generating`，会话删除按钮在忙碌时禁用，`deleteSession(id)` 也会在 `busy` 时直接返回。
- 因此在当前工作台正常操作路径里，图片正在生成时用户不能直接删除会话。
- 如果通过刷新、另一个标签页、本地存储手工修改或后续代码绕过了这个限制，删除会话不会自动取消后端图片任务。
- 后端用户侧图片任务接口当前没有 `DELETE` 或 `cancel` 能力；任务只在 `pending`、`running`、`succeeded`、`failed` 之间流转。
- 前端轮询完成后会用 `activePollMessageId` 找对应消息。若会话/消息已经不存在，结果不会再写回那条聊天消息；但 `imageTasks` / 图片库仍可通过服务端任务列表刷新看到已生成结果。

结论：当前语义是“删除本地聊天会话 ≠ 取消后端图片任务，也 ≠ 删除图片库结果”。正常 UI 防止生成中删除；异常绕过后，任务大概率继续跑，结果从任务队列/图片库恢复，而不是从已删会话恢复。

## 相关后端事实

- 路由注册在 `backend/internal/server/routes/user.go`：
  - `POST /user/image-creator/tasks`
  - `GET /user/image-creator/tasks`
  - `GET /user/image-creator/tasks/:id`
  - `GET /user/image-creator/images/:id/file`
- handler 接口 `imageCreatorService` 只暴露创建、列表、查询和文件下载。
- `ImageCreatorRepository` 没有用户侧取消任务方法；只有创建、查询、领取待执行任务、标记运行/成功/失败、图片清理和过期清理。
- `CreateTask` 创建 `pending` 任务后，在当前配置允许时会异步 `ProcessTask`；worker 也会领取 `pending` 任务处理。

## 如果后续要改成“删会话取消生图”

需要先明确产品语义，再改接口：

- 只取消当前生成中的任务，还是同时删除历史任务和图片文件。
- 删除会话是否应该影响图片库，还是只从聊天流里移除。
- 是否新增 `canceled` 状态，或复用 `failed` 并写入取消原因。
- 后端需要取消接口、权限校验、任务状态转换和 worker/generator 的取消检查。
- 前端需要确认弹窗、主动停止轮询、清理 active task 状态，并处理任务已开始但上游不可取消的提示。

## 验证依据

- 已检查 `frontend/src/views/user/ChatImageStudioView.vue` 的 `busy`、`deleteSession`、`startTaskPolling`、`pollImageTask`、`localStorage` 恢复/持久化逻辑。
- 已检查 `frontend/src/api/imageCreator.ts` 只提供创建、列表、查询、下载。
- 已检查 `backend/internal/server/routes/user.go`、`backend/internal/handler/image_creator_handler.go`、`backend/internal/service/image_creator_service.go`、`backend/internal/repository/image_creator_repo.go` 的图片任务接口和状态流转。
- 本轮是知识库更新，未改运行时代码，未跑前后端测试。
