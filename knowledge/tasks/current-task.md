# 当前任务快照

最后更新：2026-09-02 18:21 +08:00

## 背景

- 用户要求持续筛选并选择性合入上游改动；始终禁止整体 merge、rebase 与 cherry-pick。
- 已完成 S281--S285 的 429/限流队列以及 S287、S289；当前从刷新后的 v0.2.0 候选中继续挑选独立小批。

## 当前目标

- 当前 Sprint：`upstream-v0200-group-pricing-layout-s290`。
- Workflow phase：`done`。
- S290 已按修订合同完成独立 QA 和最终裁决；四个前端文件已提交为 `7cacdbab1`。S266 内容审核的产品提交已在主线等价存在，其任务、结果和 QA 证据已通过 `12e52216e` 合并回主线谱系。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## 本次已完成

- S281--S285 已分别提交为 `c886cdcac`、`f48b4b77f`、`bb3d3bca6`、`65bf61f5a`、`b686353c3`。
- S287 和 S289 已提交为 `e6845b4ea`、`6050139a3`，均有独立 QA PASS。
- 刷新后 `upstream/main=5097b3145`（v0.2.0）；本地与上游历史仍大幅分叉，继续行为级筛选。

## 已确认事实

- `343858021` 在本地已等价，不单独合入；`05ea883e2` 依赖缺失的 Group schema/Ent 状态。
- `3510aa22b` 的 reasoning-effort 按模型范围功能依赖至少 43 文件的策略基础、迁移和 Ent，不是小批，暂缓。
- `1a33dc8cc` 的四处布局意图仍在本地有效，但其 patch 因组件拓扑分叉失败；可作为 S290 的四文件手工适配。
- 最新全量 `go test ./...` 除既有 repository fixture 漂移外均通过；失败为 `account_repo_upstream_billing_probe_update_test.go:559` 的 32/34 列不一致，与本轮前端及 S287/S289 无交集。

## 保护边界

- 保留 `backend/internal/pkg/apicompat/*.go` 的既有脏改。
- 保留 `backend/internal/service/admin_service.go` 的 Pixel Cafe hunk。
- 保留 `frontend/pnpm-lock.yaml`、`frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue` 与 `outputs/**`。
- Controller 冻结的受保护 dirty diff hash：`0e467987fd7aec5fc451983bdb8f8216f97ba69c`。

## 已验证

- S290 仅修改合同列出的四个前端文件；受保护脏改 hash 仍为
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，无冲突索引。
- 任务专属 Chrome profile 已在本机 `/admin/groups` 打开创建、编辑分组的定价表单；六项默认 Token 价格在 `1440x900` 与 `390x844` 无文档或弹窗横向溢出。两张表单均取消，未保存或改动共享数据，且 session/profile/Vite 已清理。
- 定向 Vitest 2/2、typecheck、production build 和四文件 `git diff --check` 均通过。截图位于
  `E:/codex-runtime/pge/sub2api/s290/browser-smoke-20260902-retest/artifacts`。

## 当前结论

- S290 在修订后的可达浏览器范围内为 `PASS`，产品提交为 `7cacdbab1`；S266 证据谱系合并与 S290/S266 定向回归均通过，等待已授权的 `origin/main` 推送。分组弹窗保留既有 `hide-token-intervals=true`，因此启用渠道定价入口的 `IntervalRow` 真实浏览器 smoke 明确拆为后续任务。
- Kimi native Responses、Claude Fable 5.1 和 reasoning-effort scope 均为独立大功能，不与 S290 混合。

## 下一步

1. 完成已获准的 `origin/main` 推送，并只清理已合入且干净的 PGE 分支/工作树；不更新容器或部署。
2. 若继续前端定价验收，另开任务验证启用渠道定价入口的 `IntervalRow` 桌面/移动布局，不改变分组弹窗的计费语义。

## 验证记录

- `git fetch upstream --prune`：PASS，`upstream/main=5097b31457e6dc9f49e5f5c9c72b925ce79543b3`。
- `node C:/Users/Administrator/.codex/scripts/codex-workflow.mjs pge-doctor --repo . --strict`：20/20 PASS。
- `go test ./...`：除既有 repository fixture 32/34 列漂移外，其余包 PASS；未修改工作区。
- 受保护业务脏改普通 diff hash：`0e467987fd7aec5fc451983bdb8f8216f97ba69c`，本轮前保持不变。
- S290：定向 Vitest 2/2、typecheck 与 production build 均 PASS；独立 QA 使用任务专属 Chrome profile
  `E:\codex-runtime\pge\sub2api\s290\browser-smoke-20260902-retest\chrome-profile` 在本机 `/admin/groups`
  验收创建/编辑定价表单，`1440x900`/`390x844` 下六项默认 Token 价格均无横向溢出并保存截图。表单取消未保存；关闭后 session、task-owned browser/cliDaemon 与 5174 监听均为零，最终 verdict 为 `PASS`。
- 合并后回归：S290 布局与 S266 风险控制前端测试共 3 文件/9 项 PASS，`typecheck` 和 production build PASS；S266 的 service、handler、admin handler、repository、migration 与 server compile 定向命令均 PASS。
