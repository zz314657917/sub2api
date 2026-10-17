### PASS: openai-preoutput-retry-20261017-integration

2026-10-17，控制方最终复核。用户授权的断流重试优化已集成至本地主工作区；未提交、推送、部署或调用真实供应商。

## Findings

最终集成未发现阻止交付的新问题。普通及 passthrough Responses 流在尚未输出语义内容的 generic 断流、缺少终态或临时 response.failed 时，非 pool 账号获得整条请求共享的最多三次新增重试，等待 500ms、1s、1.5s；保留现有有限换号。原账号仍可用时优先重试原账号，并重新校验账号资格和并发槽位。已有文本、推理、工具、图片或有效终态输出后禁止重放。

失败尝试的 lifecycle preamble 使用有累计大小上限的独立缓冲，心跳单独记账；成功时不泄漏失败尝试的 response ID。超限、取消、策略/上下文错误不获得新 generic retry。池模式与既有 capacity/overload/OAuth429 规则保持原行为。取消后已有内容的 usage drain 和现有 partial/image usage 路径已实测保留。

## Executed Checks

- 独立合同 review、Developer、不同 QA 均完成。独立 QA fresh service 65顶层/127含子测试、adapter45/47、handler52/151、取消9/29、OAuth429 8/8；五轮重复累计145/650，全部PASS、零skip。完整生产build、gofmt和scope检查通过。
- 集成前 drift gate 拦住同文件并发改动，没有先覆盖再补救。另一项已授权上游任务在 `isCodexImageGenerationBridgeEnabled` 增加了三行 nil/nonOpenAI guard；仓库外三方合并无冲突，独立 QA 逐字节复核 merged 相对 accepted 仅多出原三行，全部 retry 改动保留。
- 精确复制五份业务/测试文件。四份与原QA seal一致；`openai_gateway_service.go` 使用补充评审通过的 merged SHA256 `e4a676ca9b6a3f7b34eb044e50ae9881463210fa8f696adc03d2c48163cfadd8`。主树已有16个并发路径变化均保留，复制前后4668个非目标tracked文件哈希不变。
- 主工作区运行 canonical tagged 专项：

```powershell
go test -json -p 1 -tags unit ./internal/service ./internal/handler -run 'TestOpenAIStreamRetry|TestOpenAIPreOutputRetry|TestOpenAIStreaming|TestNonStreamingTerminalFailureFailover|TestForwardAsAnthropic_|TestHandleAnthropic|TestOpenAIGatewayService_CodexBridgeRejectsNilAndNonOpenAI|TestOpenAIGatewayServiceForward_CodexBridgeNonOpenAI|TestOpenAIGatewayService_CodexImageGenerationBridgeOverridePrecedence' -count=1
```

service89顶层/181含子测试，handler20/68，合计109/249，exit0、零skip。

- 主工作区 `go test -p 1 ./internal/server/routes -run 'Test.*(KeyRouteFirstResponse|FirstResponseWriter)' -count=1 -v`：15顶层/28含子测试，exit0、零skip。基线与候选同类路由回归也均通过。
- 主工作区 `go build -p 1 ./...` exit0；精准 `git diff --check` 及五文件 `gofmt -l` 通过。HEAD仍为 `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`，索引内容未变，无staging/commit/push。
- 最终二次哈希检查与集成结果一致，无非目标tracked漂移。

证据根目录：`E:/codex-runtime/openai-preoutput-retry-20261017/`。

- `qa-final-20261017T173025/qa-seal.json`：原候选独立QA封存和真实命令计数。
- `merge-review/review.md`、`proof.json`：并发三行修复的独立合并复核。
- `integration-copy.json`、`integration-final.json`：复制及最终文件/Git保护证据。
- `main-integrated-tests.jsonl`、`main-integrated-route-tests.log`、`main-integrated-build.log`：最终主工作区运行证据。

## Unverified Risks

- 仅实现和验证本地代码，线上实例尚未部署、未对真实供应商故障验收。
- generic-only且max switches=10时，上限为11个账号初始/换号尝试加3次新增重试，即14次调用；新增等待最多3秒，不代表网络/排队总耗时上限。账号不可用、取消和现有路由超时均可提前结束。
- pool账号不受新增generic策略覆盖；non-streaming/Anthropic adapter没有新增generic断流重试能力。普通HTTP 403/409/413/502没有统一改为三次重试。
- 既存 ordinary/passthrough Forward wrapper 在processor返回error时丢弃partial result的限制保持不变；不声称所有错误流现在都能结算partial usage。
- 隔离候选的旧unit-tag fixture编译失败如实保留；另一并发任务修复了main对应fixtures，最终main tagged专项实际执行通过。main默认标签的既有beta fixture问题未被修改或用overlay掩盖；不以main tagged通过声称全量所有标签测试通过。

## Recommendation

本次授权的本地代码优化完成，可进入另行授权的发布/部署流程。保留现有池配置和输出后禁止重放边界。
