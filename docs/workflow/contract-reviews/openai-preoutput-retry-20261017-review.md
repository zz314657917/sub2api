---
type: contract-review
status: approved
task_id: openai-preoutput-retry-20261017
verdict: PASS
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
reviewer: independent-contract-reviewer
worker_model: gpt-6.1-sol
last_verified: 2026-10-17
---

### PASS: openai-preoutput-retry-20261017

## Contract Checked

`docs/workflow/tasks/openai-preoutput-retry-20261017.md`，修订后的 draft/pending，以及用户本轮补充的有界 preamble 缓冲要求。独立 contract 门禁通过；这不是实现验收或 runtime PASS。仅更新本评审文件，不代改 contract frontmatter 或全局 workflow。

## Findings

未发现阻止进入实现阶段的 contract 缺口。此前五项 findings 已由修订内容及用户补充要求解决：

| 评审项 | 修订后的验收边界 |
| --- | --- |
| 有限 generic 重试 | 非 pool、ordinary/passthrough Responses streaming 专用；每请求最多三次新增机会，跨账号共享计数，延时按请求内第 1/2/3 次依次为 500ms/1s/1.5s，不因换号重置。使用结构化 metadata/wrapper，不依赖错误文本或特殊延时识别策略。 |
| pool 与特例 | 保留 count=0、空状态列表、默认及自定义列表；capacity/overloaded/OAuth429 的既有优先级、上限、headers 和 Retry-After 不改变。默认标签测试验证 eligibility、limit 与实际 dispatch。 |
| pinning 与可用性 | 临时 retry target 经现有 validated pinned scheduler 重新校验；不覆盖原 pin/private-first，不直接复用旧快照。目标失效时保留 last error、排除目标并消耗已有 switch budget，再有限 fallback。busy/wait failure 可提前终止。 |
| 并发与取消 | 每次正常获取账号槽位，sleep 前释放；覆盖 account/user release、wait counter、acquire/wait 失败、backoff/wait 取消及 EOF/cancel 竞争。禁止重置计数或因重新选择而产生免费换号。 |
| 提交边界 | 独立 preamble 缓冲，只有可证明为 comment-only 的 keepalive 可提前写出；覆盖大于 4KiB、ticker/同步、passthrough 和 keepalive 后失败再成功；失败尝试 lifecycle IDs 不可泄漏。text/reasoning/tool/image 提交后不 replay；成功 terminal 后不得因后续 read error 重试。 |
| 调用范围 | generic eligibility 不扩至 buffered/non-streaming 或 Anthropic adapter。增加 `TestForwardAsAnthropic_` / `TestHandleAnthropic` 回归命令；不要求修改 denied adapter 文件。 |
| 测试与隔离 | 默认标签的非零测试和完整 production build 为 canonical gate；记录旧 tagged fixture 失败，不修复、不隐藏。仅 task allowlist 内局部修改；独立 Developer/QA，不更改其他任务状态。 |

### 本次明确采用的负载上界

对于仅触发新 generic 策略的请求，上界为 `maxAccountSwitches + 1 + 3` 次实际上游调用。因此 switches=10 时最多 14 次，新增 backoff 累计最多三秒；switches=0 且目标持续可用时最多四次。单账号耗尽后不能因为无替代账号而刷新三次预算。原有 capacity/pool 等策略不属于此 generic-only 公式，混合策略不得宣称统一十四次封顶。

验收应在新默认标签 handler 测试中断言 switches=0 的四次封顶、switches=10 的十四次封顶，以及跨账号共享三次 budget 和 500ms/1s/1.5s progression；通过实际 dispatch 序列和计数证明，不能只测试常量。目标失效、取消、槽位错误均允许少于该上界。网络时间和槽位排队不计入三秒，不承诺请求总耗时。

文案非阻塞澄清：contract 的“default10”不应用来修改现有默认值。当前 handler 构造器 fallback 为 3，正数配置可覆盖（`handler/openai_gateway_handler.go:302-306`）。本次批准的是保持原有效配置不变，并以 switches=10 作为十四次上界测试条件。

### 用户补充要求：preamble 缓冲必须有界

本次 PASS 包含用户明确新增的实现与测试约束：为每次尝试的 pending preamble 设累计字节上限，可复用现有配置的 max SSE line size，无需新配置。单行 scanner 限制不能代替累计限制；应在保存下一段数据前检查，不能先无限追加再事后检查。

超限须明确以 too-large 失败结束，既不 generic retry 也不借普通 failover 继续重放，并丢弃当前 pending preamble；不能通过提前 flush lifecycle 数据规避上限。释放尝试资源并保留取消优先级。默认标签测试须覆盖：累计大于 4KiB 但低于配置上限时安全重试；多行各自未超限、累计达到边界与超过边界时行为明确；超限后无额外上游调用、无失败 preamble 泄漏。各尝试结束后清理 pending 数据，不能跨重试累积。

这些要求可在既有 allowed service/helper/test 文件内完成；不需要新调度架构。此前要求的 44 次封顶测试已被请求级三次预算取代，不再适用。

## Executed Checks

- 本轮重读修订 contract 与原 review，结合本会话已只读检查的 Agent Matrix、service/handler/scheduler/account/helper 源码及现有测试重新裁决，未实施代码。
- 前轮只读检查时，HEAD 与 contract base 均为 `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`，`git cat-file -t` 确认对象为 commit。该项不是本轮重新验证的仓库状态。
- 前轮 main 与托管 worktree 的五个关键文件 SHA256 一致：`openai_gateway_service.go`、`openai_gateway_handler.go`、`openai_account_scheduler.go`、`openai_gateway_messages.go`、`openai_gateway_service_test.go`。worktree 为 `C:/Users/Administrator/.codex/worktrees/openai-preoutput-retry-20261017/sub2api`。并行工作可能继续变化，Developer/QA 仍须按 contract 检查 drift。
- `openai_gateway_service_test.go`、`openai_gateway_handler_test.go`、`openai_nonstreaming_terminal_failure_failover_test.go`、`openai_compat_model_test.go` 有未加 build tag 的现有覆盖。pool 配置专门测试文件带 `//go:build unit`，故其关键边界需要在允许的新默认标签测试中覆盖。
- Controller 最新提供 baseline：default-tag service+handler compile PASS；六个现有 streaming cases PASS、无 skips；tagged service compile 因五个旧 fixtures 失败，tagged handler compile PASS。本 reviewer 未独立执行这些命令，baseline 不等于新增行为验收。
- 本轮仅修改此 review 文件，未运行 Go 测试、编译、provider 调用或 Git mutation。

## Gate Checks

- success_criteria_testable: `yes`，含用户补充的有界 preamble 要求。
- allowed_paths_explicit: `yes`。
- denied_paths_explicit: `yes`。
- acceptance_commands_executable: `yes`，按源码与现有默认标签测试静态核对；执行结果由 Developer/QA 提供。
- worker_model_confirmed: `yes`，采用用户明确指定的当前 Agent Matrix `gpt-6.1-sol`。
- base_commit_confirmed: `yes`，前轮确认对象存在，本轮 contract base 未变。
- openspec_traceable: `not-applicable`，本任务以 task-local contract 为边界。

## Unverified Risks

- default-tag 测试及 production build 尚须 Developer/独立 QA 实际执行并报告非零测试数。Contract 选择默认标签作为本任务 canonical gate 是合理的；无需修复或绕过其他任务的 tagged fixtures。
- 新 generic 策略的结构化标识、共享三次 budget、原 pin/private-first 保留、live eligibility、缓冲累计上限及释放行为，均仍需实现后由独立 QA 验证。现有 failover classifier 的 true 不能直接当作新 generic eligibility。
- 本轮没有 implementation diff，不能给出实现正确性或 runtime PASS。

## Recommendation

**PASS：修订 contract 连同用户明确补充的有界 preamble 要求，足以进入后续独立 Developer 实现门禁。** 本轮不实施代码。主控交付实现说明时应携带上述缓冲上限及十四次 dispatch 验证要求，并由主控更新本任务自身的批准状态；无需修改其他任务或全局 workflow。

保留全局 workflow 的 phase/current_sprint 及并行 upstream 任务状态。本评审只更新本文件；不要求修复旧 tagged fixtures，亦不授权任何提交或部署。
