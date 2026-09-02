---
type: contract-review
scope: project
status: approved
task_id: upstream-v0185-anthropic-fallback-s287
verdict: PASS
base_commit: 409a93110
reviewer: final-evaluator
last_verified: 2026-09-02
---

### PASS: upstream-v0185-anthropic-fallback-s287

## Review

- 本地直连路径已有 `sanitizeAnthropicBodyForBetaTokens` 且在最终 beta 计算后、CCH 签名前调用，新增两个字段可复用该边界，不需要改调用者。
- fallback 字段来自客户端，不应被默认 mimic beta 打开；保留/剥离条件可由纯函数测试覆盖，且不会影响 context-management 或无关字段。
- 本地 Bedrock beta 过滤白名单已存在，fallback beta 不会进入最终 tokens；在现有 `sanitizeBedrockFieldsForBetaTokens` 中无条件清除两个字段与本地拓扑一致。
- 上游同 commit 的 CC 专用 sanitizer 在本地不存在，本 contract 明确不引入不适用代码。

## Gate Checks

- success_criteria_testable: `PASS`
- allowed_paths_explicit: `PASS`
- denied_paths_explicit: `PASS`
- acceptance_commands_executable: `PASS`
- base_commit_confirmed: `PASS` (`409a93110`)

## Approval

允许在 contract allowlist 内手工适配；不得整体 cherry-pick 上游提交或注入 fallback beta 到默认 header。
