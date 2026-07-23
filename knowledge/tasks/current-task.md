# 当前任务快照

最后更新：2026-07-23 22:54 +08:00

## 背景

- 当前主线已包含已发布的 S106-S109 和独立 Agent Identity S108。
- GitHub 最新正式版 `v0.1.164` 已完成只读盘点；整版不适合直接合并。
- 另有 `codex/group-buy-lifecycle-refund-hardening-s110` 工作树，当前任务不得
  修改或清理它。

## 当前目标

- 收口 `upstream-v0164-small-fixes-s111`：手工移植 Grok CC Switch、Grok 402
  冷却、长限流按天显示和 GPT-5.6 Sol 默认排序。
- 保持 source-only：不提交、不推送、不部署、不更新容器，不触碰 S110。

## 本次已完成

- 确认 `v0.1.164` 是当前最新 release，并完成 43 个 release 提交的行为分级。
- 创建并通过 S111 contract review，明确四项适配和十个业务/测试路径。
- 完成四项本地适配及回归测试；Grok 402 默认标签测试补齐 repository
  persistence、调度阻断和过期恢复。
- 完成 focused/broad Go、Vitest `17/17`、typecheck、1090-module production
  build、目标 ESLint、gofmt 和静态门禁。
- 创建 S111 QA 报告并把 P/G/E phase 收口为 `done`。

## 已确认事实

- Grok CC Switch 现在导入 `grokbuild`，默认 `grok-4.5`，endpoint 仅保留一个
  `/v1`，homepage 继续使用本地归一化结果。
- Grok HTTP 402 现在暂停账号 30 分钟，原因是 `grok payment required`；
  401、403、429 和 5xx 策略未改。
- 模型限流使用共享 `formatCountdown`，超过 24 小时显示天数，tooltip 显示
  完整本地日期时间并保留 console 主题类。
- `gpt-5.6-sol` 已排到默认列表第一位；裸 `gpt-5.6` 别名仍仅出现一次。

## 待验证点

- 真实 CC Switch deeplink、真实 Grok 402 和管理员登录态浏览器 smoke 未执行。
- 完整 `go test -tags=unit ./internal/service` 仍受既有 `stringPtr`、billing
  helper 和 Grok runtime-block 测试编译漂移阻断。
- 提交、推送、部署和容器更新均未授权。

## 当前结论

- `PASS / source-only`：S111 四项补丁和批准范围内的验证已完成。
- 当前 P/G/E phase 为 `done`；未提交、未推送、未部署。

## 下一步

1. 等待用户决定是否提交/推送 -> 验证：若授权，精确暂存十个业务/测试路径
   和 workflow 证据，ignored contract/QA report 使用 `git add -f`，再审 cached
   allowlist 和 `git diff --cached --check`。
2. 如需真实运行态确认 -> 验证：另行授权后执行 CC Switch、Grok 402、登录态
   浏览器或部署/容器 smoke；这些不属于当前 source-only PASS。

## 验证记录

- Grok 402 默认标签回归 `-count=10`、默认模型/available-model focused tests、
  OpenAI/admin 包测试均通过。
- 前端 focused Vitest `2 files / 17 tests`、typecheck、production build（1090
  modules）和目标 ESLint 均通过。
- `gofmt -d`、exact allowlist、`git diff --check`、冲突标记和未合并索引检查
  均通过；构建只出现既有非阻断告警。
- `HEAD = origin/main = 7e2013fdd`；主工作树开始时干净。
