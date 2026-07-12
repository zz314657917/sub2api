# 当前任务快照

最后更新：2026-07-12 12:30 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 主线基线：`main` / `origin/main` 均为 `d471e58dc`；主工作树仅保留用户自己的 `knowledge/05-current-focus.md` 修改。
- 当前集成分支：`codex/upstream-v0151-followups-s71-s73`。
- 发布边界：本轮未合入 `main`、未推送、未部署、未更新本地容器。

## 当前目标

- S71-S73 已按 `用户级 Fast/Flex -> 裸 gpt-5.6 alias/catalog -> legacy request_type fallback` 顺序完成集成和独立 QA。
- Workflow 已收口，6 个临时 Generator/QA worktree 与 branch 已通过 patch-equivalence 审计并清理；当前只等待用户明确决定是否合入主线。
- 支付并发补丁 `fc66a30ff` 继续保持独立高风险审计边界。

## 本次已完成

- S71：新增 `OpenAIFastPolicyRule.user_ids`，只读取可信 `ctxkey.APIKeyUserID`；用户规则组优先于全局规则，组内 first-match、scope/tier/whitelist/fallback 语义不变。
- S71：managed HTTP、API-key passthrough、OAuth passthrough、parsed WS、passthrough WS 首帧/后续帧均通过真实上游 capture；无效非空用户 ID 不会静默扩大成全局规则。
- S71：独立 review 找到 en/zh key 误放 `betaPolicy`；修复提交 `24c22c9a9` 将五个 key 移入 `openaiFastPolicy`，真实 locale 回归测试和复审均 PASS。
- S71 集成提交 `f1997dcf9`，QA 报告提交 `489d66ba7`。
- S72：裸 `gpt-5.6`、`openai/gpt-5.6`、`gpt5.6` 在 nil/OAuth 路径归一到 Sol，API-key compatible forwarding 保留原名；`ultra/solstice/terrain` 不误映射、不进入 Sol billing candidates。
- S72：后端目录、前端 whitelist/preset、OpenCode bare/Sol/Terra/Luna 的 `xhigh/max` 与既有限额完成；集成提交 `a8e3ee4a4`，QA 报告提交 `bf2a1dc77`。
- S73：`GetUserBreakdownStats` 使用 `ul` alias 的 Sync/Stream/WS v2 legacy fallback；非零 `request_type` 仍权威，RequestType+Stream 额外 AND、七列 scan、排序/LIMIT 和排行榜边界保持不变。
- S73 集成提交 `40b710e73`，QA 报告提交 `5e4b08af9`。

## 验证记录

- 三个 Sprint 的 worker result 首行均为 DONE，独立 QA 报告首行均为 PASS。
- 组合后端精确 discovery：S71 `4`、S72 `6`、S73 `2`；HTTP/WS、DTO、service、repository、leaderboard 和 compile-only 回归全部 PASS。
- 组合前端：`SettingsView.spec.ts`、`useModelWhitelist.spec.ts`、`UseKeyModal.spec.ts` 共 `3 files / 46 tests` PASS；`vue-tsc --noEmit` PASS。
- `d471e58dc..5e4b08af9` 的 35 个业务/合同/QA 路径全部属于批准集合；`git diff --check`、冲突标记、临时 junction 清理审计 PASS。
- 组合证据：`docs/workflow/qa-reports/upstream-v0151-followups-s71-s73-qa.md`。
- 6 个 Generator/QA 分支的 `git cherry` 均只有 `-` patch-equivalent 条目；对应 clean worktree 和本地分支已删除，当前仅保留主工作树与集成 worktree。

## 待验证点

- 未调用真实 OpenAI/Codex HTTP 或 WebSocket 上游；当前证据来自 in-process HTTP server 与真实本地 WS relay capture。
- S73 未连接真实 PostgreSQL；SQL 行为由 SQLMock 与精确 SQL 片段断言覆盖。
- 未执行 race、无关完整后端套件、生产部署、本地容器更新、`main` 合并或远端推送。
- 前端仍有既有 `router-link` stub 与 stale `caniuse-lite` warning，不影响测试结果。

## 当前结论

- `upstream-v0151-followups-s71-s73` 最终裁决：`PASS / done`。
- 三项均已在隔离集成分支完成，可进入显式主线合并决策；不能描述为已发布。

## 下一步

1. 等待用户明确授权后，再把 `codex/upstream-v0151-followups-s71-s73` 合入 `main` 并决定是否推送。
