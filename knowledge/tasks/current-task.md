# 当前任务快照

最后更新：2026-07-11 14:13 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/upstream-latency-health-column`。
- 本轮主线：检查并分批适配上游 `v0.1.150` / `v0.1.151` 可合并内容，使用 Planner / Generator / Evaluator 与独立 worktree 控制范围。
- 发布边界：当前成果未合入 `main`、未推送、未更新本地容器。

## 当前目标

- S68 已完成；下一阶段先清理已完成的 S66-S68 临时 worktree/branch，再审计剩余 upstream v0.1.151 候选。
- 支付并发补丁 `fc66a30ff` 继续单独审计，不与普通协议/运行时补丁混合。

## 本次已完成

- S67：完成 GPT-5.6 `max`、MCP/custom/tool_search bridge 与 Ops 流内错误日志；首轮 QA 的动态 output index/reasoning lifecycle finding 已由 fix1 关闭并复测 PASS。
- S68a：完成账号级 `codex_image_generation_explicit_tool_policy` 后端策略与 OAuth/setup-token/API Key 专用 UI，支持 `strip/remove/drop` 归一和 nested fallback/cleanup；独立 QA PASS。
- S68b：完成 flat `image_generation`、`image_gen` namespace、Responses Lite `additional_tools`、matching `tool_choice` 在 managed HTTP、API-key/OAuth passthrough、parsed WS、WS passthrough 与 Spark 的一致 strip。
- S68b 初始 review 发现 `OpenAIWSIngressModePassthrough` 绕过 parsed strip；fix1 已补首帧/后续帧 adapter strip、invalid raw JSON 与 OAuth actual forwarded-body 覆盖。
- 组合实现提交：`7593079a9`、`19066c93d`；QA 报告提交：`c2529cd4e`。

## 已确认事实

- 仅精确匹配 flat `type=image_generation` 与 namespace `type=namespace,name=image_gen`；非图片 namespace、普通 function、custom `imagegen`、`tool_choice:auto` 均保留。
- 账号策略默认/未知值为 `allow`；`strip/remove/drop` 忽略大小写和首尾空格并归一为 `strip`；仅 Codex client 命中。
- HTTP passthrough 同时替换 `body` 与实际交给 forwarder 的 `originalBody`；API-key 与 OAuth recorder 均证明上游收到 stripped body。
- WS parsed ingress 与 passthrough mode 都通过 upstream capture 验证；passthrough 首帧和后续 text frame 在 fast policy/hooks/usage metadata/actual relay 前 strip，非 text frame保持原样。
- S67 apicompat custom/tool_search/namespace/tool-choice 与 messages fallback 保持性测试通过。

## 待验证点

- 未向真实 OpenAI/Codex 上游发送 HTTP 或 WS 请求；当前证据来自 in-process HTTP recorder 与 WS capture。
- 未运行完整 `internal/service` 套件；该套件存在既有 `group_peak_rate` 时区断言漂移，本轮按合同只跑定向回归与 compile-only。
- 未执行 race、生产部署、本地容器更新、`main` 合并或远端推送。

## 当前结论

- `upstream-codex-imagegen-namespace-strip-s68b` 最终 `PASS / done`。
- QA 报告：`docs/workflow/qa-reports/upstream-codex-imagegen-namespace-strip-s68b-qa.md`。
- 当前集成分支可继续做剩余 upstream 候选审计，但任何新候选都应另立 contract；不得把 `fc66a30ff` 混入普通补丁波次。

## 下一步

1. 清理 S66-S68 已完成临时 worktree/branch -> 验证：`git worktree list` 不再包含已合入集成分支且工作树干净的临时项，保留有未合并/脏改的引用。
2. 刷新并审计 upstream v0.1.151 剩余提交 -> 验证：按可独立移植 / 依赖较大 / 单独高风险分类，明确本地祖先依赖与冲突面。
3. 如用户明确要求合入主线 -> 验证：先审当前分支相对 `main`/`origin/main` 的提交范围，再 merge、定向回归、push verify；默认不自动执行。

## 验证记录

- 原 S68b primary/policy/apicompat/fallback/compile acceptance：全部 PASS。
- Fix1 focused strip 与 WS passthrough effort/billing/session 回归：全部 PASS。
- `TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughImageNamespaceStripAcrossTurns -count=3`：PASS。
- `git diff --check 5869f7b08..HEAD`：PASS。
- Allowed/Denied path audit：10 个允许 backend source/test 路径、workflow 证据路径，无 bridge/apicompat source/fallback source/frontend/billing/migration/deploy 越界修改。
