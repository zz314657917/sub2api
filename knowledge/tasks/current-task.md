# 当前任务快照

最后更新：2026-07-21 23:04 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- S89 已完成 API Key 编辑弹窗加宽、桌面双栏和右侧多分组路由布局。
- S90 根据系统公开设置控制“账号池策略”字段可见性。

## 当前目标

- 已完成：`account_share_enabled === false` 时隐藏完整账号池策略字段。
- 已完成：开启或字段缺失时继续显示，兼容旧版公开设置响应。
- 已完成：隐藏状态不重置已有策略，创建/更新 payload 语义不变。

## 本次已完成

- `KeysView` 复用已加载的 `publicSettings.account_share_enabled`，没有增加
  API 请求或第二套开关来源。
- 账号池策略的 label、Select 和 hint 作为一个整体条件渲染。
- 参数化测试覆盖关闭、开启、缺失字段；编辑回归覆盖隐藏的
  `private_only` 仍原样提交。

## 已确认事实

- KeysView 定向 Vitest：`2 files / 18 tests` 通过。
- 前端类型检查通过。
- 生产构建通过，转换 `1088 modules`。
- `git diff --check` 和未合并索引检查通过。
- 本地 `3000` Vite 预览已关闭。

## 当前结论

- `PASS / source-only`：S89 与 S90 前端改动、测试和构建闭环。
- S89 已部署到本地 `62080` 容器；S90 尚未部署。
- 当前改动未提交、未推送。

## S91 当前目标

- 将模型匹配从 API Key 多分组路由迁移到管理员维护的分组
  `model_match_patterns`。
- 用户端移除模型匹配输入；后端拒绝普通用户提交旧字段。
- 路由和 `/v1/models` 统一使用分组规则；切换迁移在规则未补齐时阻断，
  补齐后事务清理旧 API Key 规则并刷新认证缓存。
- S91 contract：`docs/workflow/tasks/group-model-match-centralization-s91.md`。

## S91 已完成

- `groups.model_match_patterns` 已加入 Ent schema、迁移、管理员 DTO 和分组编辑。
  管理员规则会 trim、转小写、去重、排序，并拒绝空规则；`*` 表示全部模型。
- API Key 路由编辑页已移除旧 `model_patterns` 输入；普通用户 Create/Update
  提交该字段会被后端拒绝，仓储普通保存也不会写回旧规则。
- 多分组路由和 `/v1/models` / `/v1/model-catalog` 共用分组规则，按模型匹配后
  选择最低 priority，同优先级按 weight；渠道 `restrict_models` 语义未改。
- 已加入 S91 migration preflight/switch API。未配置的有效活跃分组会阻断切换；
  切换事务内清理 API Key 旧规则，并刷新相关认证缓存；缓存快照版本已升级。

## S92 当前目标

- 用户已确认继续简化 API Key 多分组路由：普通用户只选择分组、拖拽
  顺序、启用/删除；优先级由顺序自动生成。
- 用户端不再显示优先级输入、权重、冷却秒数、生图范围、文本范围或智能
  路由预设。保存时固定写入 `weight=1`、`cooldown_seconds=30`，丢弃旧
  scope 字段；后端兼容字段暂不删除。
- S92 contract：`docs/workflow/tasks/key-route-priority-only-s92.md`。

## S92 已完成

- 用户 API Key 多分组路由已收敛为分组选择、拖拽排序、启用/删除；预设、
  权重、冷却秒数、生图/文本范围和用户模型匹配均不再显示。
- 旧路由按原 priority 稳定排序并连续重编号；保存统一写
  `priority=index+1`、`weight=1`、`cooldown_seconds=30`，只保留 enabled，
  丢弃 `model_patterns`、`image_only`、`text_only`。
- 重复分组保存被拒绝；S90 账号池策略隐藏与原值提交回归保持通过。
- S92 contract：`docs/workflow/tasks/key-route-priority-only-s92.md`；QA：
  `docs/workflow/qa-reports/key-route-priority-only-s92-qa.md`。

## S93 当前目标

- 在“系统设置 -> 外部接入”增加默认 Key 兜底分组；新用户由系统生成的
  默认 API Key 将该分组写入基础 `group_id`，专用默认路由继续优先。
- 提供显式回填按钮，仅补齐已有用户尚未分组的默认 Key，不自动迁移数据。
- S93 contract：`docs/workflow/tasks/default-key-fallback-group-s93.md`。

## S93 已完成

- `studio_bridge_luoye_ai.default_fallback_group` 已完成后端设置、DTO、管理端
  启用分组下拉和保存回读；不存在、格式错误或未启用的分组会在写入时拒绝。
- 新默认 Key 基础分组和专用路由共用系统创建权限豁免；普通用户创建 Key 的
  分组权限校验保持不变。兜底仍校验分组状态、平台、routing scope 和 S91
  `Group.MatchesModel`。
- `POST /api/v1/admin/settings/default-key-fallback/backfill` 只更新每个用户
  最低 ID、未软删除且 `group_id IS NULL` 的默认 Key；已有分组、后续 Key、
  专用路由和其他字段不变，实际变更 key 的认证缓存逐个失效。
- 管理端回填要求选择值已保存，执行前确认，完成后显示实际更新数量。
- S93 最终结论：`PASS / source-only`；未提交、未推送、未部署、未更新容器。

## 下一步

1. 如用户明确要求提交/推送，复核 S89-S93 精确路径后做
   scoped staging；如要求更新容器，另走 `local-docker-update-guard`。

## 验证记录

- 2026-07-21 13:49 +08:00：KeysView Vitest `18/18`、typecheck PASS。
- 2026-07-21 13:50 +08:00：生产 build、diff、未合并索引检查 PASS。
- 2026-07-21 19:28 +08:00：S91 migration `-run TestS91GroupModelMatchMigration -v` 两个测试 PASS；
  S91 routing/API rejection `7/7`、gateway model catalog 全量定向 PASS。
- 2026-07-21 19:29 +08:00：S91 前端 `3 files / 20 tests`、typecheck、production build
  `1088 modules`、`git diff --check`、未合并索引检查 PASS；已知全量基线失败未纳入 S91。
- 2026-07-21 22:43 +08:00：S92 定向 Vitest `2 files / 16 tests` PASS；typecheck、production
  build `1089 modules`、源代码控件审计、`git diff --check`、未合并索引检查 PASS。
- 2026-07-21 23:04 +08:00：S93 default-tag service/handler、PostgreSQL repository
  integration、SettingsView `23/23`、typecheck、production build `1089 modules`、
  diff/conflict/unmerged 检查 PASS；unit-tag service 聚合仍被既有测试漂移阻塞编译。
