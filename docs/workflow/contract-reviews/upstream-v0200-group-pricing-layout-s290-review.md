---
type: contract-review
scope: project
status: approved
task_id: upstream-v0200-group-pricing-layout-s290
verdict: PASS
base_commit: 6050139a3
reviewer: final-evaluator
last_verified: 2026-09-02
---

### PASS: upstream-v0200-group-pricing-layout-s290

# Contract Review

## Task ID

`upstream-v0200-group-pricing-layout-s290`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0200-group-pricing-layout-s290.md`

## Findings

- 原 QA 报告继续保留 `BLOCKED`：它准确记录了旧合同要求在分组弹窗显示
  `IntervalRow` 的不可达性，并在 Contract Revision Note 中要求本次独立复审；
  未把该覆盖空洞倒写为历史 QA 通过。
- 修订合同、spec addendum 与 status 一致确认创建/编辑两处调用都固定
  `:hide-token-intervals="true"`。因此浏览器验收只覆盖实际可见的六项默认
  Token 价格；不改变该 prop、不保存表单，也不以构造区间行来改变分组计价语义。
  共享 `IntervalRow` 的 `pricing-interval-grid` 保留为源码断言，enabled-route
  的真实浏览器 smoke 被明确拆为后续任务，仍是未验证风险而非被掩盖的通过项。
- 受保护脏改门禁已具备可执行性并已实测：合同中的
  `git diff -- $protected | git hash-object --stdin` 返回
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，与冻结基线一致；四个
  受保护路径的 `git diff --check` 亦通过。
- 静态复核最新版 `try/finally`：浏览器仅访问本机 Vite，发现并记录带
  `--user-data-dir` 的新建浏览器进程，拒绝默认 Chrome profile；两个视口均包含
  snapshot、overflow 断言和截图。`finally` 关闭命名 session、停止记录的 Vite PID，
  仅按已确认的 profile/session 清理并断言无残留 owned browser/cliDaemon。
- `npx.cmd` 可用（11.8.0），实际 `playwright-cli --help` 明确支持
  `-s=<session>`，因此合同的 CLI session 语法可执行。
- `base_commit` 与当前 `6050139a397e...` 一致。四个业务文件与当前用户脏改不重叠；
  `wide` 是既有 `BaseDialog` 宽度，定向 Vitest、`typecheck` 和 `build` 脚本均存在。
  上游补丁分叉，禁止直接 cherry-pick 正确。
- 源级验收已明确要求断言两个 `wide` 弹窗、两个可换行标题行、非收缩添加控件和
  两个响应式网格标记；保留六字段、输入/emit/校验和计价语义不变。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes` (`6050139a3`)
- openspec_traceable: `not-applicable`

## Approval

Contract is approved with the revised browser boundary. The prior QA `BLOCKED`
report is not a release PASS; next legal action is independent QA against the
revised contract, including the reachable create/edit six-control browser smoke
and the existing static `IntervalRow` sentinel. No production URL, credential,
shared data, merge, push, deployment or container action is authorized.
