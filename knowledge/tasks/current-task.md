# 当前任务快照

最后更新：2026-09-04 10:00 +08:00

## 背景

- 用户要求持续筛选并选择性合入上游改动；始终禁止整体 merge、rebase 与 cherry-pick。
- 已完成 S281--S285 的 429/限流队列以及 S287、S289；当前从刷新后的 v0.2.0 候选中继续挑选独立小批。

## 当前目标

- 当前 Sprint：`upstream-v0200-ops-proxy-attribution-s291c`。
- Workflow phase：`build`。
- S290 已按修订合同完成独立 QA 和最终裁决；四个前端文件已提交为 `7cacdbab1`。S266 内容审核的产品提交已在主线等价存在，其任务、结果和 QA 证据已通过 `12e52216e` 合并回主线谱系；`origin/main` 已同步至 `5b95e68dd`。
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

## 待验证点

- 启用渠道定价入口的 `IntervalRow` 仍未跑真实浏览器 smoke；后续单独打开该入口并在桌面与移动视口检查布局，不能通过修改分组弹窗的 `hide-token-intervals=true` 来规避。
- `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266` 是已注销的任务依赖残留；如需释放空间，先验证它仍不在 `git worktree list` 且无关联进程，再删除该精确目录。

## 当前结论

- S290 在修订后的可达浏览器范围内为 `PASS`，产品提交为 `7cacdbab1`；S266 证据谱系合并与 S290/S266 定向回归均通过，已推送 `origin/main@5b95e68dd`。分组弹窗保留既有 `hide-token-intervals=true`，因此启用渠道定价入口的 `IntervalRow` 真实浏览器 smoke 明确拆为后续任务。
- Kimi native Responses、Claude Fable 5.1 和 reasoning-effort scope 均为独立大功能，不与 S290 混合。

## S291-A 合同

- 上游最新代理归因链拆为 S291-A/B/C；当前只进入 S291-A 核心事件与队列边界。
- 合同和独立审查均为 PASS：
  `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291a.md`、
  `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291a-review.md`。
- S291-A 允许修改 Ops 事件/队列核心及对应测试；所有网关调用点留给后续合同。
- S291-A build 已完成：定向测试、完整 `internal/service`、`go build ./...`、
  `git diff --check` 和未合并索引检查均通过；结果见
  `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291a-result.md`。
- S291-B 合同和独立审查均为 PASS，当前只补本地 Gateway/Gemini HTTP 调用点；
  OpenAI/WS/provider 剩余调用点延后至 S291-C。
- S291-B build 已完成：Gateway 单体 22 个事件点与 Gemini 兼容入口补齐代理归因，
  定向测试和 `go build ./...` 通过；结果见
  `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291b-result.md`。
- S291-C 合同和独立审查均为 PASS，当前补 OpenAI/Grok/WS 错误事件调用点。
- S291-C build 已完成：OpenAI/Grok/WS 现有生产事件点和 WS fallback unknown
  语义已覆盖，定向及完整 service 测试、构建通过；Antigravity 单体逻辑另拆 S291-D。

## 下一步

1. 若需释放空间，可手动删除已注销的 `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266` 目录；该目录只含任务依赖残留，Git 已无注册。删除前验证其不重新出现在 `git worktree list`。
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
- 发布与整理：`git push origin main` 成功，远端从 `6050139a3` 更新为 `5b95e68dd`；S266 的三个 PGE 分支及 S280/S281 已合入分支已删除，两个干净 S266 子工作树和父工作树的 Git 注册已删除。其父目录仍留有未注册依赖文件，因宿主拒绝递归删除而未强行清理；所有脏/冲突/备份工作树均保留。
