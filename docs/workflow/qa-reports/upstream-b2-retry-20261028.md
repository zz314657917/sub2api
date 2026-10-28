# upstream-b2-retry-20261028 本地合入与验收

日期：2026-10-28  
结论：**PASS / 已精确合入本地工作树**

## Findings

B2 已按批准合同从上游 `c591412c42ab195dacdede9eb19f9cbfee221144` 适配到当前本地拓扑，完成空流安全重试和请求级 overload / `Retry-After` 行为：

- Responses sync/ticker/passthrough 只缓冲明确已知的空文本、空 reasoning、空 refusal、空内容片段和可重试前置错误帧。
- unknown、malformed、缺失/null/数字/对象字段、空白文本、工具公告/完成和 encrypted reasoning 仍作为输出边界，已有输出后不重放。
- OpenAI API-key/OAuth/setup-token 的合资格 HTTP overload 支持请求级标记和最多 8 秒的安全 `Retry-After` 等待，保留原 headers、账号选择、并发、重试次数和取消行为。
- 请求级 overload 不写账号健康失败、不临时禁用账号、不伪造成功采样；真实 auth/quota/429、WS、原生 Anthropic、CN 和其他协议继续沿用旧策略。
- Responses 的原 `nil,error` partial 策略、raw Chat 的既有 partial usage 和 Images 的已完成子调用不重放行为均保留。

仅精确复制冻结的 15 个 B2 文件。主树中的并行 frontend 7 文件及已有 dirty 内容保留，没有被覆盖。

## Executed Checks

- 独立 QA：`qa/qa-report.md` 为 PASS，`qa/qa-seal.json` SHA256 为 `8447f83782308e296301c3c9b417cd3ce5ca803ebd625f4b95098710f96dee99`；绑定 15 个源文件、33 个前置和 71 项证据。
- B2 固定回归：service default `1155 pass / 1 原失败`，service unit `1200 / 2 原失败`，handler default `279 / 7 原失败`，handler unit `279 / 7 原失败`，build exit 0。原 17 项失败名称、具体断言和执行集合与基线一致，零 skip/missing。
- B2 新测试：19 个 `TestUpstreamB2*` 声明在 default/unit 各执行 3 次，合计每标签 1890 个 pass 事件，audit PASS。
- 独立外置探针：第一轮 default/unit 各 44 项通过；第二轮加入数值 status 冲突边界后各 49 项通过；覆盖异常帧、Retry-After 极值、资格冲突、真实 handler 重试、slot/body、健康采样、usage/charge、输出后不重放和部分图片结果。
- A1–A7 交叉回归：新声明 default/unit 各 3 次，每标签 912 个 pass 事件，32 声明 audit PASS；原有 A 回归 default `1266`、unit `1503`、packages `104`，build exit 0。
- 格式与范围：15 个 Go 文件 `gofmt -l` 无输出，`git diff --check` 无输出；R2 保护门禁最终 PASS，主树 12,225 原文件保护和更新后的 12,226 文件保护均通过，HEAD/index 未变。

## Unverified Risks

未执行真实 provider、真实数据库、容器、部署或生产请求验收。基线中原有 17 项失败仍然存在，未被豁免。

## Recommendation

本地代码和测试验收完成，可以进入后续提交整理；本轮未执行提交、推送或部署。
