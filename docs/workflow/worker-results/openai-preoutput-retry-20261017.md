### DONE: openai-preoutput-retry-20261017

2026-10-17，独立 Developer 实现报告。修订候选 **CANDIDATE READY / CODE FROZEN**；这不是独立 QA 或最终集成 PASS。

## Scope and changes

Worktree: `C:/Users/Administrator/.codex/worktrees/openai-preoutput-retry-20261017/sub2api`

Base: `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`

仅修改以下五份业务/测试文件，另新增本报告：

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_stream_retry.go`（新）
- `backend/internal/service/openai_stream_retry_test.go`（新，默认标签）
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_stream_retry_test.go`（新，默认标签）

`backend/internal/service/openai_gateway_service_test.go` 保持原样。未修改 scheduler、compact keepalive helper、adapter、routes、dependencies、config、其他 workflow/status/spec/task、主工作区或 index。未执行 stage/commit/reset/push、provider/database/container 操作。

## Implementation

- `OpenAIStreamRetryError` 通过 typed wrapper 标记非 pool OpenAI Responses streaming 的 replay-safe generic failure；仅 ordinary/passthrough streaming 调用新 constructor，原 shared constructor 保持原策略，non-streaming/Anthropic 不获得新 generic policy。
- handler 使用请求级三次预算，跨账号不重置；新增等待依次为 500ms、1s、1.5s。既有 pool、capacity、overload、OAuth429 的计数、优先级与状态分类不变。
- 临时 target 通过现有 validated pinned scheduler 重新选择和正常 acquire/wait。保留原 caller pin；task-local context 在已有 eligibility owner 中保留 private/shared 限制。先释放旧账号槽位再 backoff；目标失效时排除目标并消耗原 switch budget，有限 fallback，不复用旧 account snapshot。
- ordinary/passthrough lifecycle preamble 独立缓冲，使用已有 MaxLineSize 作累计上限，追加前检查；不会进入可自动 flush 的 4KiB output writer。超限是普通 too-large 错误，不获得 generic 或普通 failover replay。
- 普通 comment 单独记账，与现有 compact comment 核算共存；仅可证明的完整 comment 写入不触发 replay boundary。未知 header-only、partial write 和语义输出仍禁止 replay。
- 未改变实际 max-account-switches 默认值。generic-only、持续可用时 switches=0 为最多四次，switches=10 为最多十四次；既有特殊策略不属于这一公式。新增 backoff 总计最多三秒，不是请求网络/排队总时长上限。

## Controller cancellation finding and focused correction

控制方与独立 QA 在旧候选中发现：新增的 handler 统一 `ctx.Err()` 提前返回会跳过已有 partial usage/图片结果处理；stream 无条件取消判断会截断已输出后的 terminal usage drain。旧 CANDIDATE READY 已撤回，以下修复完成后重新冻结：

- handler 将取消判断放在 failover/retry 入口，普通 error 在已有 partial usage 提交后才因取消返回；有图片的 result 继续进入原正常 usage 分支。没有新增 billing/result propagation 架构。
- stream 新增取消保护只对尚未有语义/未知 committed output、且未发现 client disconnect 的请求立即生效；输出后取消停止下游写入，继续解析仍可读的 detached upstream。ticker 在此状态下禁用已经 ready 的 ctx.Done channel，避免循环抢占 terminal/usage 事件。
- **保留 HEAD 原有限制**：ordinary/passthrough Forward wrapper 在 processor 返回 error 时仍返回 `nil, err`，不传播 processor partial result。本任务没有修复或改变这一 pre-existing 行为，也不声称它现在可记录所有错误流的 partial usage。

新增实证测试：

- `TestOpenAIStreamRetryCancellationAfterOutputDrainsUsage`：同步、ticker、passthrough 各自文本/图片，真实语义写出触发取消，再允许上游提供 terminal usage；断言 input=13、output=7、cached=3、image=1（图片场景）。
- `TestOpenAIPreOutputRetryCancellationRecordsExistingPartialUsage`：通过真实 Responses handler 和既有 raw-chat adapter 的 result+error 路径，在输出后取消且读取 usage，再返回 context.Canceled；断言实际 UsageLogRepository 收到记录恰好一次、token 明细准确、无额外 dispatch。未修改该 adapter。
- `TestOpenAIPreOutputRetryCancellationAfterOutputRecordsTerminalUsage`：真实 handler 同步/ticker/passthrough 文本与图片六种场景；输出触发取消后，上游继续可读，断言 terminal usage 实际记录恰好一次、图片计数准确、仅一次 dispatch 和正常资源释放。
- 同批回归继续覆盖取消/EOF 竞争、backoff 取消、wait/acquire 错误及等待中取消，保证 pre-output 取消无 replay。

## Commands and evidence

执行目录均为该 worktree 的 `backend`，`git diff --check` 从 worktree 根目录执行。日志根目录：

`E:/codex-runtime/openai-preoutput-retry-20261017/`

计数为 RUN 对应的顶层 PASS / 包含子测试的 PASS；重复命令按执行次数累计。全部成功测试无 skip。

| Command | Result / counts | Evidence |
| --- | --- | --- |
| `go test -p 1 -tags unit ./internal/service ./internal/handler -run '^$'` | 旧 service fixtures 编译失败；handler 仅 compile，不计验收测试 | `developer-tagged-compile.log` |
| `go test -json -p 1 ./internal/service -run 'TestOpenAI(Stream\|Streaming\|Passthrough\|NonStreaming)\|TestNonStreamingTerminalFailureFailover\|TestOpenAI.*Retry' -count=1` | 旧候选 PASS，64 / 120，exit=0 | `developer-canonical-service.jsonl` |
| `go test -json -p 1 ./internal/service -run 'TestForwardAsAnthropic_\|TestHandleAnthropic\|TestNonStreamingTerminalFailureFailover' -count=1` | 旧候选 PASS，45 / 47，exit=0 | `developer-canonical-adapters.jsonl` |
| `go test -json -p 1 ./internal/handler -run 'Test.*(StreamRetry\|PreOutput\|SameAccount\|Failover\|Streaming)' -count=1` | 旧候选 PASS，50 / 143，exit=0 | `developer-canonical-handler.jsonl` |
| `go test -json -p 1 ./internal/service ./internal/handler -run 'TestOpenAIStreamRetry\|TestOpenAIPreOutputRetry' -count=5` | 旧候选 PASS，130 / 575，exit=0；session 47505 自然结束，未 kill/restart | `developer-canonical-repeat.jsonl` |
| `go test -json -p 1 ./internal/service -run '^TestS282' -count=1` | 旧候选 OAuth429 PASS，8 / 8，exit=0 | `developer-oauth429.jsonl` |
| `go build -p 1 ./...` | 旧候选完整 production build PASS，exit=0 | `developer-build.log`；exec session 38010 返回 BUILD_EXIT=0 |
| 下方 cancellation focused command | **最终修订候选 PASS，9 / 29，exit=0** | `developer-cancellation-fix.jsonl`；session 74435 |
| `git diff --check` | 最终修订候选 PASS，exit=0 | Developer 冻结前执行 |

上表中的 `\|` 是 Markdown 表格转义；实际 shell 正则使用普通 `|`，没有更改测试选择表达式。

最终修订候选的 focused command：

```powershell
go test -json -p 1 ./internal/service ./internal/handler -run 'TestOpenAIStreamRetryCancellation|TestOpenAIPreOutputRetryCancellation|TestOpenAIStreamingClientDisconnectDrainsUpstreamUsage|TestOpenAIStreamingContextCanceledReturnsIncompleteErrorWithoutInjectingErrorEvent|TestOpenAIPreOutputRetryCancelEOF|TestOpenAIPreOutputRetryCancelBackoff|TestOpenAIPreOutputRetrySlotFailures' -count=1
```

按控制方最新明确指示，取消回归修复后没有重跑整个 canonical/repeat/build；独立 QA 将对下方最终冻结 hashes 运行新 canonical 与五轮重复。旧日志只能证明旧候选的对应检查，不替代最终候选验收。

### Retained failed evidence

- `developer-new-tests.jsonl`：v1 有两个实际失败测试，JSON 中另含父测试/包 fail 事件，未隐藏或覆盖。capacity fixture 原文不是既有 classifier 识别的特例，修正为已有 capacity 测试的 `Selected model is at capacity...`；compact coexist fixture 未设置 client-stream 标记，补齐既有 helper 的启动条件。没有删除或降低断言。
- `developer-new-tests-v2.jsonl`：上述修正后的旧候选，23 / 109 PASS。
- unit-tag probe 保留五组原有 fixture 漂移：`ops_health_score_test.go`、`billing_service_test.go`、`billing_service_unified_test.go`、`gateway_context_management_test.go`、`proxy_update_probe_invalidation_test.go`。没有修复、overlay、排除或宣称 tagged suite 通过。

## Dispatch/resource coverage already authored

默认标签测试包含：单账号四次、单账号无替代账号不刷新预算、十一账号十四次、A,A 成功、有替代账号仍固定 A、A 失效后 B 且跨账号共享三次 backoff、失效 pin 无免费 switch、原 pin/private-first/归属/group 限制、pool count=0/空列表/默认列表/默认 count/自定义列表/generic 502 保持既有决定、capacity 六次/overload 四次及原延时、凭据旧 failover、acquire/wait 失败与取消、exactly-once account/user/wait releases、sleep 前已释放账号槽位、body close、>4KiB preamble/累计上限边界与超限/单行超限、comment/ticker 后重试不泄漏失败 ID、text/reasoning/tool/image/terminal 后无 replay、实际 httptest HTTP upstream。

最终 acceptance 以独立 QA 的运行与审查为准。

## Frozen SHA256

| Path | SHA256 |
| --- | --- |
| `backend/internal/service/openai_gateway_service.go` | `20f071dd2354fd284e0a68e6a6c05d51b17d7e284b7f18d3d2409dde18c687c1` |
| `backend/internal/service/openai_stream_retry.go` | `cf168a022a467c26130e6915045fb8001323541fb897dd8d9b174832ce5999da` |
| `backend/internal/service/openai_stream_retry_test.go` | `46143424d67a54dfa019145a4c63ec599d9a2ebb19caa0651bbe5ad3298e90c4` |
| `backend/internal/handler/openai_gateway_handler.go` | `84463c34bb351b6745a3c8bd51bdbe9197cc9d59673d37c3cffb2f0b55505f93` |
| `backend/internal/handler/openai_stream_retry_test.go` | `857d7ae0e8dcb01836067237f0893d960ec1c8d68e1ff66d2385a231df5358e2` |
| `backend/internal/service/openai_gateway_service_test.go`（未修改） | `526e2e9c548aa627fdd24cbd08dc3b48831ee4363b737b5fa495b9e83ebf0342` |

## Boundaries and remaining ownership

- 没有真实 provider、生产数据库、Redis、container、部署或提交证据，也未执行这些操作。
- 本地 stub/httptest 调用和准确计数是 runtime smoke 证据；不等于真实供应商可用性。
- 控制方负责 routes 集成回归、主工作区 drift/index protection、最终复制集成和 verdict；Developer 未动主工作区。
- 独立 QA 必须对最终冻结候选重新验收，尤其 cancellation/usage 边界。本报告没有写入 QA verdict，也未代替独立 QA。
- 无新增 scheduler/auth/billing 架构；无 knowledge/memory 写入或需要提升的稳定知识候选。
