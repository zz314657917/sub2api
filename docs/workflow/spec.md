---
repo: sub2api
project_type: generic
qa_mode: runtime
last_verified: pending
---

# Product Spec

## 一句话需求
- 继续迁移旧版生图 Canvas 能力，让 sub2api 的 Canvas 工作台能复用现有 API Key 和 ImageCreator 任务队列提交真实生图运行。

## 目标与非目标
- 目标：
  - Canvas run 创建后能识别 `text_to_image` / `image_to_image` 节点，并转换为现有 ImageCreator task。
  - 复用 ImageCreatorService 的 API Key 归属、分组生图权限、并发限制、任务队列、图片保存和 gateway 计费链路。
  - Canvas run 输出中记录节点到 ImageCreator task 的映射，便于前端展示运行结果和后续轮询。
  - 前端 `/canvas` 支持选择可用 API Key、编辑节点参数、运行前保存画布、展示最近运行状态。
- 非目标：
  - 本阶段不实现完整节点执行引擎、跨节点数据流调度或并行/串行队列编排。
  - 本阶段不把 Canvas run 取消级联到 ImageCreator task。
  - 本阶段不迁移旧版裁剪、外扩、mask、历史和模板完整能力。

## 技术方案
- 后端给 Canvas repository 增加 run running / complete 状态更新方法。
- CanvasService 注入 ImageCreatorService，`CreateRun` 保存 run 后立即标记 running，解析当前 Canvas 文档的可执行图片节点并调用 `CreateTask`。
- `text_to_image` 使用节点 config 或上游文本/提示词节点生成 prompt；`image_to_image` 使用节点 config 或上游图片节点的 `referenceImageId`，通过受保护的当前用户图片读取接口加载参考图 bytes。
- 前端 `CanvasView` 复用用户 API Key 列表接口，只传 `api_key_id`，不暴露明文 key。

## Sprint 计划
- Sprint 2：Canvas 真实运行链路最小闭环，后端排 ImageCreator task，前端节点参数/API Key/运行状态展示。
