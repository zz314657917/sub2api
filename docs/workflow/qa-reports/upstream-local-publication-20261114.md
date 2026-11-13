# Gemini 与上游四项修复本地提交验收

日期：2026-11-14。结论：PASS，含一项已独立复现的既有失败。

本次将本会话两批已审核改动整理为单一本地提交：Gemini 价格目录与计费兜底、
可配置同账号重试、非流式 thought 内容隔离，以及 refresh token 原子消费、
账号 hydration 失败释放并发槽、价格 hash 不一致重试、Sonnet 5.5 缓存读价对齐。
提交只包含 21 个产品/测试文件、本报告和原 Gemini 差异说明表。

独立候选基于 `75fa7b0456d3d8ecc8640c96b07baf2b45bf502c`，使用该提交原始
go.mod/go.sum。从两批审核补丁精确提取任务差异，未将共享文件中的其他模型适配、
网络重试或未提交依赖带入。原主目录字节保持，任务之外的修改继续留在工作区。

## 已执行检查

Go 1.26.10，GOROOT 指向对应安装，GOTOOLCHAIN=local，GOWORK=off。

| 检查 | 实际结果 |
| --- | --- |
| config/repository/service 默认定向回归 | 867 pass，0 fail |
| 同范围 unit 定向回归 | 1219 pass，1 个既有失败 |
| Sonnet55FourFixes 默认与 unit 专项 | 各 3 pass，0 fail |
| `go build -mod=readonly ./...` | exit 0 |

unit 唯一失败：
`TestAPIKeyResolveForModelRequestAllowsGeminiModelsOnOpenAIGroup`，
`api_key_routing_test.go:408`，`Expected value not to be nil`。
在完全未应用本任务补丁的同一基准独立 worktree 中执行相同测试，复现同一断言。
保留原断言，不修复、不删测、不以跳过测试伪装全部通过。

实际定向正则：
`Test(Gemini|RefreshToken|AuthService|GatewayServiceNewSelectionResult|Pricing|ParsePricing|GetModelPricing|NewModelBilling)|Test.*(Gemini|ChannelImage|ImagePrice|ComputeTokenBreakdown)`。
三个包为 `./internal/config ./internal/repository ./internal/service`；
默认和 unit 都使用 `-json -mod=readonly -count=1`，unit 增加 `-tags unit`。
Sonnet 专项使用 `-run '^TestSonnet55FourFixes'`。

逐项合入前的独立合同与 QA 记录保持历史语义，未伪造独立角色执行测试。
本次提交验收由主控执行。原 Gemini 说明表见
`outputs/upstream-gemini-20261111/report.md`；其中“下一批四项”是当时建议，
现已由本次四项修复落实。该表与价格均来自固定上游源码快照，不代表实时官方报价。

## 证据与边界

提交候选、精确提取记录、完整日志、候选冻结和提交回读：
`E:/codex-runtime/sub2api/upstream-publication-20261114/`。
四项修复独立 QA、原问题红/绿回归及证据包：
`C:/Users/Administrator/.codex/state/plugins/codex-security/scans/sub2api/artifacts-f4e9a931fe062ea28e7fd64dd3bebde3fb49ae843638e41b97d462044f9e3578/artifacts/upstream-four-fixes-20261113/`。

未推送、部署、更新容器或连接真实 provider/生产 Redis/数据库。
miniredis 回归不等于真实 Redis 验收；未运行全库测试及完整服务/浏览器。
家族撤销与签发之间的既有竞态不在本次单 token 消费修复范围。
