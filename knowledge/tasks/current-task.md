# 当前任务快照

最后更新：2026-07-23 21:30 +08:00

## 背景

- 当前主线已包含 S106-S109，以及独立的 OpenAI Agent Identity 功能提交
  `6b87a2d2b`。
- `upstream-openai-agent-identity-s108` 与
  `user-usage-column-menu-layer-s108` 是两个不同任务；编号碰撞，但合同、代码和
  证据路径互不相同。

## 当前目标

- Agent Identity 验收证据归档、主线推送和本地分支清理已完成。
- 当前无待合并工作；新上游变更必须另开 contract。
- 未修改业务代码、依赖、数据库、部署或容器。

## 本次已完成

- 确认功能提交 `6b87a2d2b` 是当前 `main` 的祖先。
- 从旧收口分支仅恢复主线缺失的 QA report 和 worker result；没有 cherry-pick
  会覆盖 S106-S109 新状态的 `3ddc0d6ab`。
- 证据提交 `68dc78661` 已推送并验证 `HEAD`、`origin/main`、GitHub `main`
  一致，分叉计数为 `0 0`。
- 已删除本地 `codex/upstream-openai-agent-identity-s108`；当前只剩 `main`
  和主工作树。
- 已在 workflow 文档中记录 Agent Identity 与菜单层级两个 S108 的编号碰撞。

## 已确认事实

- Agent Identity 支持 Ed25519 assertion/task 注册与一次恢复、snake/camel
  导入、无普通 access token 的 K12/Team 账号、HTTP/images/WS/quota/test 路径，
  以及私钥/assertion 脱敏。
- S109 模型定价和图片输入计费、S106 小修、S107 安全依赖升级、用户 usage
  菜单层级 S108 均已发布。
- 旧分支的业务价值和独有验收证据均已在主线保留；其余内容是过期状态快照。

## 待验证点

- 真实 K12 Agent Identity、外部 OpenAI 请求、race detector、部署、容器和
  登录态浏览器 smoke 不在本轮授权范围，仍未执行。
- `go test -tags=unit ./internal/service` 的既有编译漂移，以及 govulncheck
  剩余 Go 标准库/AWS SDK 问题，需要独立 Sprint 处理。

## 当前结论

- `PASS / published`：Agent Identity 功能和验收证据均已进入主线，冗余本地
  分支/worktree 已清理。
- 当前 P/G/E phase 为 `done`；本轮不再继续混入新的上游变更。

## 下一步

1. 如继续同步上游，先盘点候选并为下一项批准独立 contract。
2. 安全升级另开 Sprint，处理 Go patch 版本和 AWS SDK 剩余漏洞。
3. 如需验证真实 K12，单独授权脱敏凭据、外部网络和运行态 smoke。

## 验证记录

- fresh `go test`：Agent Identity Admin/DTO、service 模式、server compile PASS。
- 三项并发/恢复测试以 `-count=10` PASS。
- `git diff --check`、业务路径为 0、冲突标记、未合并索引和精确暂存门禁 PASS。
- `68dc78661` 推送后 `HEAD = origin/main = GitHub main`，分叉 `0 0`。
- 清理后 `git branch -vv` 仅有 `main`，`git worktree list` 仅有主工作树。
