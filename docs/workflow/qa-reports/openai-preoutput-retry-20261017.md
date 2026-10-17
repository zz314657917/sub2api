### PASS: openai-preoutput-retry-20261017

2026-10-17，独立 QA，按当前 Agent Matrix 与 approved contract/review 验收修订冻结候选。此结论仅适用于下列 isolated candidate 与 sealed 文件，不代替控制方的 main 集成及最终裁决。

- Candidate：`C:/Users/Administrator/.codex/worktrees/openai-preoutput-retry-20261017/sub2api`。C 盘位置是控制方确认的 host-managed worktree 例外。
- HEAD/base：`de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`。
- Fresh QA evidence：`E:/codex-runtime/openai-preoutput-retry-20261017/qa-final-20261017T173025/`。
- Seal：该 evidence 目录内 `qa-seal.json`；逐命令结果为 `results.json`、精简计数为 `summary.json`、测试/风险映射为 `coverage.json`。每条命令另有 `.process.json`、`.stdout.log`、`.stderr.log`。
- 工具链：`go1.26.3 windows/amd64`；有效 `GOFLAGS`、`GOWORK` 均为空。使用默认标签完整 package 测试，无 overlay、无源文件列表替代，无源码或测试改写。

## Findings

修订候选未发现阻止本 contract 交付的新问题。前轮两项取消回归已通过实际运行验证修复：

1. 输出前取消/EOF 仍禁止新增 dispatch；输出后取消不再无条件中断可读 upstream 的 terminal usage drain。同步、ticker、passthrough 的文本/图片场景均通过。
2. handler 不再在所有 error 分支入口直接因取消丢弃 usage；原 adapter 的 result+error partial usage 提交及图片结果的既有 usage 流程得到保留，实际 repository 记录恰好一次。

保留两个明确边界：

- **候选 unit-tag 基线失败并未修复。** 本轮首先独立执行 tagged compile probe，service 仍因五份旧 fixtures 失败，handler compile 成功。错误位置与原 `baseline-tagged-compile.log` 一致；这是本 contract 已批准保留的 baseline，不是通过的 tagged suite。
- **既存 nil-on-error 限制不变。** HEAD 原有 ordinary/passthrough Forward wrapper 在 processor 返回 error 时仍返回 `nil, err`，不会传播 processor partial result。本任务未新增或修复该架构；本 PASS 不宣称所有错误流都能结算 partial usage。验证的是 readable terminal success 的 drain/usage 保留，以及既有能够返回 result+error 的 adapter 路径保留。

## Executed Checks

收到控制方修订版明确 `CANDIDATE READY` 后，独立读取 Developer 冻结报告、核对全部六个 owner SHA256，再运行以下 fresh commands。Developer 的旧候选通过日志与 focused 9/29 不作为本轮替代证据。

计数格式为“顶层 PASS / 含子测试 PASS”。五轮重复按实际执行次数累计；表中各行存在测试重叠，不应相加解释为独立测试数量。所有实际验收测试 **skip=0、fail=0**；每包均有非零顶层执行。

| Fresh check | Exit | 顶层 / 含子测试 | Evidence prefix |
| --- | --- | --- | --- |
| service canonical | 0 | 65 / 127 | `service` |
| Anthropic/non-streaming canonical | 0 | 45 / 47 | `adapters` |
| handler canonical | 0 | 52 / 151 | `handler` |
| repeat5，service | 0 | 45 / 310；9 个顶层、62 个含子测试名称，各执行五次 | `repeat5` |
| repeat5，handler | 0 | 100 / 340；20 个顶层、68 个含子测试名称，各执行五次 | `repeat5` |
| cancellation 专项，service | 0 | 4 / 13 | `cancellation` |
| cancellation 专项，handler | 0 | 5 / 16 | `cancellation` |
| OAuth429 原策略回归 | 0 | 8 / 8 | `oauth429` |
| 完整 production build | 0 | `go build -p 1 ./...`，非测试数量项 | `build` |
| diff whitespace | 0 | `git diff --check` | `diff-check` |
| 六份 owner 格式检查 | 0 | `gofmt -l` stdout/stderr 均为空；只读，未格式化 | `gofmt` |
| 单独保留的 unit-tag compile probe | 1 | service baseline build failure；handler compile only，无测试执行 | `tagged-compile` |

实际命令，除 Git/gofmt 位于候选根目录外，cwd 均为候选 `backend`：

```powershell
go test -p 1 -tags unit ./internal/service ./internal/handler -run '^$' -count=1
go test -json -p 1 ./internal/service -run 'TestOpenAI(Stream|Streaming|Passthrough|NonStreaming)|TestNonStreamingTerminalFailureFailover|TestOpenAI.*Retry' -count=1
go test -json -p 1 ./internal/service -run 'TestForwardAsAnthropic_|TestHandleAnthropic|TestNonStreamingTerminalFailureFailover' -count=1
go test -json -p 1 ./internal/handler -run 'Test.*(StreamRetry|PreOutput|SameAccount|Failover|Streaming)' -count=1
go test -json -p 1 ./internal/service ./internal/handler -run 'TestOpenAIStreamRetry|TestOpenAIPreOutputRetry' -count=5
go test -json -p 1 ./internal/service ./internal/handler -run 'TestOpenAIStreamRetryCancellation|TestOpenAIPreOutputRetryCancellation|TestOpenAIStreamingClientDisconnectDrainsUpstreamUsage|TestOpenAIStreamingContextCanceledReturnsIncompleteErrorWithoutInjectingErrorEvent|TestOpenAIPreOutputRetryCancelEOF|TestOpenAIPreOutputRetryCancelBackoff|TestOpenAIPreOutputRetrySlotFailures' -count=1
go test -json -p 1 ./internal/service -run '^TestS282' -count=1
go build -p 1 ./...
git diff --check
gofmt -l backend/internal/service/openai_gateway_service.go backend/internal/service/openai_stream_retry.go backend/internal/service/openai_stream_retry_test.go backend/internal/service/openai_gateway_service_test.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_stream_retry_test.go
```

### 风险矩阵与真实 dispatch

下列测试均经本轮 fresh canonical/专项执行，新增 retry 测试同时经过 repeat5；`coverage.json` 可定位对应 PASS 名称与子场景。结合实现 diff、helper 调用点和既有 validated scheduler 路径只读复核，未发现扩围。

| 风险 | 本轮证据与结论 |
| --- | --- |
| R01：generic eligibility | `TestOpenAIStreamRetryPolicy`、`DisconnectDispatch` 验证 read error/missing terminal/generic response.failed；typed `OpenAIStreamRetryError` 区分新策略，仅 ordinary/passthrough streaming 使用新 constructor。 |
| R02：四次实际调用/固定目标 | `FourAttempts` 的 ordinary/passthrough 实际序列均为 `[1,1,1,1]`；`PinnedSuccess` 在另一账号可用时仍为 `[1,1]`，只返回成功 attempt ID。`LocalHTTP` 经实际 httptest HTTP upstream，断言 upstream 两次、账号 `[1,1]`、成功输出。 |
| R03：请求级预算/上界 | `FourteenAttempts` 实际为 `[1,1,1,1,2,3,4,5,6,7,8,9,10,11]`，取得 14 个 slot；`SingleAccountNoAlternative` 不刷新预算。`InvalidTargetSharedBudget` 为 `[1,2,2,2,3]`，仍按请求级 500ms/1s/1.5s progression；未修改原有效 switch 默认值。 |
| R04：pool | `PoolDispatch` 的 zero/empty/default/missing_count_default/custom_429/custom_excludes_429/generic_default/generic_custom_502 八种场景，实际 dispatch 分别为 1/1/3/4/2/1/1/1，保持旧配置决定。 |
| R05：特例/headers | `SpecialPolicies` 实际 capacity=6、overloaded=4，后者保留 1s/2s/3s 延时；credentials 旧 failover 为 `[1,2]`。原 `TestS282*` 八项、stream 429/pool 回归及 `TestCopyFailoverRetryAfterValidatesAndCopiesHeader` 通过。 |
| R06：live target 失效 | `InvalidTargetSharedBudget` 确认失效 A 后选择 B；`InvalidTargetNoFreeSwitch` 在 switches=0 时仅调用 A 一次并保留 upstream error。静态确认重新 GetByID、eligibility/group/transport/capability/runtime-block 校验，无直接复用旧 account snapshot。 |
| R07：pin/private/group | `OriginalPinPrivateFirst` 的 original_pin/private_first/lost_ownership/wrong_group 实际序列为 `[1]` / `[1,1]` / `[1,2]` / `[1,2]`；`PinAndPoolRestrictions` 验证原 pin 不覆盖、private/shared 用户归属限制。 |
| R08：slots/wait | `FourAttempts` 每次 dispatch 校验已取得 slot、旧 slot 与 body 已释放，并检查 release 到下一 dispatch 间隔覆盖 backoff。`SlotFailures` 覆盖 acquire_error/wait_error/cancel_wait；fixture 在每个实际 handler smoke 后检查 account/user exactly-once release、held=0、waitAdded=waitRemoved、body close 数等于 upstream 调用数。 |
| R09：取消/usage | backoff、wait、EOF/cancel 竞争均通过且无额外调用；取消后 drain 与实际记录见下段专项证据。 |
| R10：>4KiB preamble | `Preamble` 覆盖 sync/ticker/passthrough × missing_terminal/read_error/response_failed/comment；`PinnedSuccess` 及 `TickerKeepalive` 验证失败 attempt 的大型生命周期 preamble 不泄漏，重试成功仅出现成功 ID。 |
| R11：累计/单行 overflow | `PreambleBound` 覆盖累计恰好上限与多一字节，上限在 append 前判断；`OverflowNoDispatch`、`OversizedLine` 的 ordinary/passthrough 均只有一次实际 upstream，无 generic/普通 failover replay、无失败 ID 泄漏。 |
| R12：keepalive/未知写入 | comment、ticker keepalive 之后失败再成功均通过；`CommentAccounting` 包含 compact comment 共存、header-only/未知写入/partial write 的保守 replay 边界。 |
| R13：输出后禁止 replay | `NeverAfterOutput` 的 ordinary/passthrough text/reasoning/tool/image/成功 terminal 后 read error 均仅实际调用一次；`OutputBoundary` 和旧 post-output/drain/quarantine 回归通过。 |
| R14：负向分类 | policy/context-window 旧测试及新 generic eligibility table、bare error conservative classifier 通过；显式 client/auth/account-state 不获新 generic same-account policy，credentials 旧 failover 保留。HTTP 403/409/413/502 owner 未扩展 generic HTTP 行为。 |
| R15：adapter/non-streaming | 45/47 canonical 回归及 shared constructor 未获 typed generic wrapper 的断言通过；未新增 adapter read-error retry 能力。 |
| R16：scope/freeze/build | 完整构建、diff、gofmt、非零/无 skip、精确五轮计数及全部文件前后哈希检查通过，见下方 seal 与隔离证据。 |

### 取消修复的专项证据

本轮单独重跑 cancellation 命令 **9 顶层 / 29 含子测试 PASS**，且对应新增测试在 canonical 与五轮重复中也通过：

- `TestOpenAIStreamRetryCancellationAfterOutputDrainsUsage`：sync/ticker/passthrough × 文本/图片六种场景，真实 downstream semantic write 触发 cancel，readable-after-cancel body 才允许继续提供 terminal；断言 input=13、output=7、cached=3、image=1（图片场景），并确认 terminal 数据已读尽。
- `TestOpenAIPreOutputRetryCancellationAfterOutputRecordsTerminalUsage`：同六种场景经过真实 Responses handler/service；每种只 dispatch 一次，实际 `UsageLogRepository` 记录一次、token 明细正确、图片计数为 1。fixture 同时检查资源释放。
- `TestOpenAIPreOutputRetryCancellationRecordsExistingPartialUsage`：沿现有 raw-chat adapter 的 result+error 路径，输出后取消、读取 usage 再返回 `context.Canceled`；实际 repository 恰好记录一次、无新增 dispatch、无额外 response.failed。
- `CancellationWinsEOF`、`CancelEOF`、`CancelBackoff`、`SlotFailures` 与既有 canceled/disconnect drain 回归通过。输出前取消仍禁止 retry，未因修复恢复而放松 replay 保护。

上述证明 usage 收集及记录路径保留；不声称真实供应商、生产数据库扣费或所有 error partial result 的传播能力。

### 隔离、scope 与旧失败证据

- 初始 scope 精确为两份 tracked 修改、三份 untracked 新文件；恰为下面五份 accepted code/test paths。第六份 `openai_gateway_service_test.go` 哈希与 baseline 一致，未修改。
- `workspace-before.json` / `workspace-after.json` 对 **4,673** 个 tracked/untracked 文件路径核对，`workspace-drift.json` 为 `{}`；HEAD 与 index diff SHA256 前后相同。ignored 的 Developer report 另外校验哈希，QA 最终报告是本轮唯一候选写入文档。
- 未更改 scheduler/adapter/routes、旧 tagged fixtures、其他 workflow/status/tasks、dependencies/config、main 或主 index。QA 运行脚本、解析结果与证据只写指定 runtime；未写业务或测试源文件。
- `retained-developer-evidence.json` 保存既有 `developer-*` 日志 SHA256，结束复查均不变；包括失败的 `developer-new-tests.jsonl`（v1）、后续 v2、旧 canonical/repeat/build 与修订 focused 日志，没有覆盖、删除或以新日志冒充旧结果。
- candidate tagged stderr 的错误位置列表与原 baseline 相同：`ops_health_score_test.go`、`billing_service_test.go`、`billing_service_unified_test.go`、`gateway_context_management_test.go`、`proxy_update_probe_invalidation_test.go`。没有 overlay、修复或排除 fixture。
- 控制方报告：另一活动任务已修复 main fixtures，main 当前 tagged **compile-only** exit=0，日志为 runtime 根目录 `main-current-tagged-compile.log`；candidate routes 15 顶层/28 含子测试 PASS。它们是控制方证据，不计入本 QA service/handler 数量。本报告不声称当前 main 仍被旧 tagged baseline 阻塞。
- 控制方说明 main 另有九项并行修改，六份本任务 main owner 及 HEAD/index 未变。本 QA 不导入这些并行 fixture/workflow 修改；其最新字节保护、main 非零 tagged 回归、集成后 build/routes 仍由控制方负责。

## Accepted SHA256

下列五份文件与 Developer 修订冻结报告完全一致，QA 开始/结束及 seal 前复查稳定。哈希为文件实际字节 SHA256；仅此精确业务/测试集合可作为本 QA 接受的集成候选。

| Accepted path | SHA256 |
| --- | --- |
| `backend/internal/service/openai_gateway_service.go` | `20f071dd2354fd284e0a68e6a6c05d51b17d7e284b7f18d3d2409dde18c687c1` |
| `backend/internal/service/openai_stream_retry.go` | `cf168a022a467c26130e6915045fb8001323541fb897dd8d9b174832ce5999da` |
| `backend/internal/service/openai_stream_retry_test.go` | `46143424d67a54dfa019145a4c63ec599d9a2ebb19caa0651bbe5ad3298e90c4` |
| `backend/internal/handler/openai_gateway_handler.go` | `84463c34bb351b6745a3c8bd51bdbe9197cc9d59673d37c3cffb2f0b55505f93` |
| `backend/internal/handler/openai_stream_retry_test.go` | `857d7ae0e8dcb01836067237f0893d960ec1c8d68e1ff66d2385a231df5358e2` |

未修改、仅作 regression owner 保护的文件：

| Unchanged path | SHA256 |
| --- | --- |
| `backend/internal/service/openai_gateway_service_test.go` | `526e2e9c548aa627fdd24cbd08dc3b48831ee4363b737b5fa495b9e83ebf0342` |

Developer/QA 文档哈希由 `qa-seal.json` 单独记录，不混入业务文件 map；seal 不计算自身递归哈希。

## Unverified Risks

- 本 PASS 是 frozen isolated candidate 的 scoped default-tag 验收及完整 backend production build；没有运行全仓所有测试或已知失败的完整 unit-tag suite。
- 没有真实 provider、生产数据库、容器、部署或线上可用性证据。runtime smoke 使用 task-owned httptest/local in-process upstream 与 repository stubs。
- 新 generic-only 上界是 switches=10 时至多 14 次 upstream，加三次 backoff 共三秒；capacity/pool/overload 混合策略不受该统一公式约束。网络、排队、原有超时不属于这三秒，不宣称请求总时长封顶。
- wrapper nil-on-error 为明确保留的旧限制；main 最新并行修复和集成后状态尚需控制方完成非零回归与保护性复查。

## Recommendation

**PASS：可交控制方按 seal 精确集成五份代码/测试文件，再执行其 main hash/index 保护、非零 tagged retry/cancel 回归、完整构建与 routes 门禁。** 本 QA 未 stage/commit/push、未操作 main、provider、数据库或容器；本报告不是发布或部署授权。
