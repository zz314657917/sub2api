### PASS: upstream-v0216-test-compat-20261015-integration

2026-10-17。独立合同复审、Developer 和不同 QA 均完成；主控仅将六个已封存的 service 测试文件定向复制进主工作区，没有生产逻辑修改。

## Findings

未发现阻断集成问题。原五个夹具修复暴露 external service_test 的邮箱绑定 helper/仓库 stub 漂移，Developer 按原范围停手后，经独立合同复审批准第六文件。未削弱既有断言；13 个邮箱绑定测试函数保持原文，余额 stub 对意外调用立即失败。

QA 首次范围核验将已提交的 CRLF Python 工具文件误判为修改；复核证明该文件与 HEAD 原始字节及同提交 pristine 工作树完全一致。修正 QA 检查脚本后独立执行验收，原失败日志保留，没有修改工具文件或跳过核验。

## Executed Checks

- 独立 QA 原样执行：impacted 10 顶层/33 含子测试、related 66/131、邮箱绑定 13/13、GPT-6.1 Sol 21/364、额外 Gemini policy 2/8，全部 PASS，零 fail/skip；完整后端生产 build、六文件 gofmt 只读检查和 diff 检查通过。
- QA 核对六文件完整 diff，其余 4664 个受跟踪文件与同基线 pristine 原始字节相同；QA 执行前后 4670 个受跟踪文件、HEAD/index 均未改变。
- 封存清单 SHA256：`4610fe32d5c0606773db999ea872f321f74e336ead4ce42eefbf3278c8e7ce20`。六个主树目标在复制前与指定 base 原始字节一致，复制后与 QA seal 一致。
- 集成前后 12111 个非目标文件哈希零漂移；前端已验收 15 文件哈希不变。HEAD 仍为 `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`，暂存区为空。未提交、推送或部署。
- 相同六个封存文件已复制进独立网关候选作为只读前置，哈希及范围复核通过；网关业务实施和验收属于下一批。
- 主树合入后原样执行 `go test -p 1 -tags unit ./internal/service -run '^TestShouldFailoverGeminiUpstreamError$' -count=1` 退出0；没有 overlay。默认标签 GPT61 专项仍在既有无关脏文件 `gateway_beta_test.go:138` 因 `undefined: newTestGatewayServiceForBeta` 编译失败，退出1；该文件未改动。原日志 `main-post-compat-unit-baseline.log` / `main-post-compat-default-baseline.log`。

证据：`E:/codex-runtime/upstream-v0216-20261015/test-compat-integration-result.json`、`test-compat-main-before.json`、`test-compat-main-target-backup/`、`test-compat-sealed-files.json`、`gateway-prerequisite.json`；独立报告为 `docs/workflow/qa-reports/upstream-v0216-test-compat-20261015.md`。

## Unverified Risks

本 PASS 只覆盖六文件测试兼容及保护性集成，不代表主树全部无关脏改已通过回归，也不代表网关三项修复、真实 provider、数据库或部署验收。主树最终构建/专项在网关合入后另行执行并记录。

## Recommendation

测试兼容前置本地集成 PASS；继续已批准的独立网关 Developer/QA 门禁。
