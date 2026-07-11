# 当前任务快照

最后更新：2026-07-11 17:06 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`main`。
- 本轮主线：检查并分批适配上游 `v0.1.150` / `v0.1.151` 可合并内容，使用 Planner / Generator / Evaluator 与独立 worktree 控制范围。
- 发布边界：S65-S70 已合入并推送到 `main`；未部署、未更新本地容器。

## 当前目标

- S70 已完成并通过独立 QA；S65-S70 已发布到 `origin/main`，已合并 feature 分支也已清理。
- 支付并发补丁 `fc66a30ff` 继续单独审计，不与普通协议/运行时补丁混合。

## 本次已完成

- S67：完成 GPT-5.6 `max`、MCP/custom/tool_search bridge 与 Ops 流内错误日志；首轮 QA 的动态 output index/reasoning lifecycle finding 已由 fix1 关闭并复测 PASS。
- S68a：完成账号级 `codex_image_generation_explicit_tool_policy` 后端策略与 OAuth/setup-token/API Key 专用 UI，支持 `strip/remove/drop` 归一和 nested fallback/cleanup；独立 QA PASS。
- S68b：完成 flat `image_generation`、`image_gen` namespace、Responses Lite `additional_tools`、matching `tool_choice` 在 managed HTTP、API-key/OAuth passthrough、parsed WS、WS passthrough 与 Spark 的一致 strip。
- S68b 初始 review 发现 `OpenAIWSIngressModePassthrough` 绕过 parsed strip；fix1 已补首帧/后续帧 adapter strip、invalid raw JSON 与 OAuth actual forwarded-body 覆盖。
- 组合实现提交：`7593079a9`、`19066c93d`；QA 报告提交：`c2529cd4e`。
- 已清理 15 个 S66-S68 临时 worktree/branch；清理前均确认工作树干净且相对当前 HEAD 无 patch-unique commit。
- S69：修复 GPT-5.6 Sol/Terra/Luna 在 `service_tier=priority` 时 cache creation 仍按 Standard 价计费的问题，贯通 static/dynamic/embedded JSON、channel/interval override、272k long-context、Flex 与真实 `RecordUsage`。
- S69 实现提交 `d5a1aef0b`，worker result `07399e50d`，独立 QA 报告 `f7a2d67a9`，workflow 收口 `0a962b30e`。
- `deepseek-v4-pro` 外部 worker 返回 model 404；按用户已明确授权多智能体，Generator/QA 改用当前可用协作 agent，并在 result/QA 报告中记录偏差。
- S69 worker 与 QA 临时分支在 `git cherry codex/upstream-latency-health-column <branch>` 全部显示 patch-equivalent (`-`) 后删除；当前只剩集成工作树。
- S70：新增平台用户 `exclude_from_leaderboard` 管理开关，覆盖 migration 190、Ent、admin DTO/API、前端编辑弹窗、排行榜/冠军/徽章/奖励过滤和缓存失效；独立 QA PASS。
- S70 实现提交 `997f423ff`，workflow/QA 提交 `4dc2b462e`，均已推送到 `origin/codex/leaderboard-participation-exclusion-s70`。
- S65-S70 主线合并提交为 `d6ff6a158`；合并后 backend 定向回归、frontend `6 files / 62 tests`、typecheck、migration tests 和 `git diff --check` 均 PASS，`origin/main` 已核对到同一提交。

## 已确认事实

- 仅精确匹配 flat `type=image_generation` 与 namespace `type=namespace,name=image_gen`；非图片 namespace、普通 function、custom `imagegen`、`tool_choice:auto` 均保留。
- 账号策略默认/未知值为 `allow`；`strip/remove/drop` 忽略大小写和首尾空格并归一为 `strip`；仅 Codex client 命中。
- HTTP passthrough 同时替换 `body` 与实际交给 forwarder 的 `originalBody`；API-key 与 OAuth recorder 均证明上游收到 stripped body。
- WS parsed ingress 与 passthrough mode 都通过 upstream capture 验证；passthrough 首帧和后续 text frame 在 fast policy/hooks/usage metadata/actual relay 前 strip，非 text frame保持原样。
- S67 apicompat custom/tool_search/namespace/tool-choice 与 messages fallback 保持性测试通过。
- GPT-5.6 cache-write 每 token：Sol Standard/Priority `6.25e-6 / 12.5e-6`，Terra `3.125e-6 / 6.25e-6`，Luna `1.25e-6 / 2.5e-6`。
- Priority 先选专用 cache-write 价，再在输入侧总量 `Input + CacheCreation + CacheRead > 272000` 时乘 `2.0`；`==272000` 不触发。Flex 继续取 Standard 后乘 `0.5`。
- channel/interval `CacheWritePrice` 同时覆盖 Standard/Priority 且保留显式零；5m/1h breakdown 不读取新的 flat Priority 字段。

## 待验证点

- 未向真实 OpenAI/Codex 上游发送 HTTP 或 WS 请求；当前证据来自 in-process HTTP recorder 与 WS capture。
- 未运行完整 `internal/service` 套件；该套件存在既有 `group_peak_rate` 时区断言漂移，本轮按合同只跑定向回归与 compile-only。
- 未执行 race、生产部署、本地容器更新、`main` 合并或远端推送。
- 未发送真实 OpenAI Priority 请求；S69 证据来自 deterministic service/resolver/WS/RecordUsage 测试。
- `-tags=unit` suite 仍有既有无关编译漂移，按 S69 contract 未作为验收证据。

## 当前结论

- `leaderboard-participation-exclusion-s70` 最终 `PASS / done`；S65-S70 发布前多智能体复核无业务阻断，已合入并推送主线。
- S70 QA 报告：`docs/workflow/qa-reports/leaderboard-participation-exclusion-s70-qa.md`。
- 当前集成分支可继续做剩余 upstream 候选审计，但任何新候选都应另立 contract；不得把 `fc66a30ff` 混入普通补丁波次。

## 下一步

1. 选择下一独立 Sprint -> 验证：重新审计 remaining upstream 候选并另立 contract；`fc66a30ff` 继续保持支付高风险隔离边界。

## 验证记录

- S69 exact test discovery：6/6；六项必需测试 PASS。
- pricing/RecordUsage service regressions、WS `CacheCreation|Usage`、default service compile：PASS。
- `git diff --name-only 2b3b5514e..07399e50d`：恰好 8 个 contract allowed paths；`git diff --check` PASS。
- 独立 code review 与 fresh QA：PASS，无阻断 finding。
- S69 cleanup audit：worker 两笔和 QA 一笔均 `git cherry = -`；清理后 `git worktree list` 仅剩当前集成工作树。
- 主线清理审计：本地只剩 `main`；origin 只剩 `origin/main`，已删除完全合并的 S69、S70 和旧 `group-buy-v1` 分支。
