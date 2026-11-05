### DONE: key-fast-policy-20261105

Backend Developer 实现、测试及 R3 最小修订已完成，源码冻结，交独立 QA。DONE 表示开发交付完成；独立 QA 和 Final Evaluator 尚未给出产品验收结论。

## 合同与冻结

- 候选：`E:/codex-worktrees/sub2api/key-fast-policy-20261105`。
- 证据目录：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend`。
- 最终 R3 合同完整 SHA256：`7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61`，与独立 R3 PASS 的 `contract_sha256` 一致。
- 原 R2 合同 SHA256：`f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce`；早期证据保留。
- HEAD：`0f6b739881ccf9e4ad423f464aeedcc9abd7f48a`。
- 产品冻结清单：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/source-manifest.json`，共 **32 文件：31 Go + 1 SQL**，包含原 31 文件及唯一获准旧测试。
- 产品文件映射摘要 `manifestDigest`：`6247ba8661f3e21fdff0632a07a23d0f7780c3d94786b740d7f29ecfa42dcb5c`。算法为按清单键顺序对 `JSON.stringify(files)` 的 UTF-8 字节计算 SHA256。
- `source-manifest.json` 文件自身 SHA256：`36833b026f808c54a705f1f8e713e2b71868b3e223db4b94a336791da097dd1e`。
- 原 31 文件冻结清单归档：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/source-manifest-r2-frozen.json`，原映射摘要 `5ef299af72e6a8b92c1bcaddfd2bd62f13ddf347310ad9c440e5b5e2ab99ba4a`。这 31 文件与当前逐文件 SHA256 全部一致，业务源码及新测试未因 R3 改变。
- 本报告单独交付，不放入产品文件映射，避免循环哈希；报告自身哈希由外部 `freeze-receipt.json` 绑定。

## 实现

- 新字段 `openai_fast_policy` 支持 `passthrough`、`force`、`filter`；默认及历史空值为透传。创建及更新 API 区分遗漏与显式值；更新遗漏保留；null、非字符串、空字符串和未知字符串均返回 400。
- ent schema、八个对应生成文件、迁移 255、DTO/mapper、domain、repository create/update/auth Select/hydration 已贯通。policy 使用独立 update mask，保留 quota/usage 等既有字段所有权。
- 保留 owner 检查及 Cafe seat/membership 托管密钥限制。auth snapshot 版本从 20 升为 21，覆盖双向序列化、旧版本 miss/reload、缓存失效；旧版本 L2 snapshot 不再提升进 L1。
- middleware 将已鉴权密钥策略写入类型安全 request context；有效密钥切换覆盖旧 context。客户端 body/header 不能替代可信策略来源。
- 新策略 helper 独立于大网关文件，仅实际 OpenAI 账号生效。优先级为 global > group force > key > client。force 覆盖任何 tier；filter 仅去除规范化 fast/priority，保留 flex、ultrafast 等其他值。
- HTTP 及 normal/passthrough WS 的每次 `response.create` 应用策略，包括后续省略 tier 的帧；其他 WS 事件保持原字节。出站 OAuth routing hint 使用最终 body。
- raw Chat 及 Responses→raw Chat 的流式/非流式 tier metadata 均重取最终策略输出，覆盖原非 nil 值的替换和删除；沿用既有计费与协议流程。
- ent 生成先输出仓库外，仅导入新字段必要差异；无关 ID 扫描生成漂移恢复候选原字节。未修改 module/deps、价格或旧 prerequisite。

## R3 唯一旧测试改动

路径：`E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_policy_group_fast_20261103_test.go:141`。

```diff
-	require.Equal(t, 20, snapshot.Version)
+	require.Equal(t, 21, snapshot.Version)
```

原文件 7428 字节，修改后仍为 7428 字节；仅零起始 offset 6258 从 ASCII `0` 改为 `1`，其余字节完全一致。group force、JSON 往返、RouteGroups、duplicate、旧版本拒绝及其他断言全部保留。

- 修改前 SHA256：`a19f5c449b78e0864a8c1ee1e992b59fb449e01077be512c98f040fe55f6bbe9`。
- 修改后 SHA256：`38da2ee67af08d0fe910295debf076e0e3efa4fd3292c2ce70928cd2410f6f52`。
- 原字节归档：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/r3-original/openai_policy_group_fast_20261103_test.go`。
- 字节比较、旧冻结及 prerequisite 核验：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/r3-integrity.json`。

## 真实执行计数与日志

计数来自 Go JSON 的带 `Test` 的 pass/fail/skip 事件，含 subtest，不含 package 结束事件；不是将 subtest 冒充顶层测试函数。权威计数：`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/test-counts.json`。

以下日志均位于上述证据目录：

| 执行与范围 | PASS 事件 | FAIL 事件 | SKIP | 顶层 PASS / FAIL | 日志 |
| --- | ---: | ---: | ---: | ---: | --- |
| 新测试 default，R3 前最终运行 | 479 | 0 | 0 | 14 / 0 | `new-default-final.jsonl` |
| 新测试 unit，R3 前最终运行 | 479 | 0 | 0 | 14 / 0 | `new-unit-final.jsonl` |
| 既有定向回归 default，保留原结果 | 233 | 1 | 0 | 56 / 1 | `regression-default.jsonl` |
| 既有定向回归 unit，保留原结果 | 272 | 2 | 0 | 83 / 2 | `regression-unit.jsonl` |
| 原始输入 overlay，两失败项独立归因 | 1 | 1 | 0 | 1 / 1 | `baseline-two-tests.jsonl` |
| R3 GroupFast 定向回归 default | 158 | 0 | 0 | 11 / 0 | `groupfast-r3-default.jsonl` |
| R3 GroupFast 定向回归 unit | 170 | 0 | 0 | 13 / 0 | `groupfast-r3-unit.jsonl` |

每种配置的新测试 479 事件为 **14 个顶层 + 465 个 subtest**：service 446、handler 24、DTO 1、repository 3、middleware 5。它们验证真实本地 HTTP/WS transport、最终 metadata、middleware/context/cache、repository/service/handler；没有调用真实 provider。

R3 只改旧测试单字节，按主控要求只重跑 GroupFast，不重跑 479 或扩大回归。新增两次运行均 `exitCode=0`，stderr 为空；命令、时间和环境记录分别为 `groupfast-r3-default-command.json`、`groupfast-r3-unit-command.json`。不同轮次计数没有合并为一轮“全通过”，早期失败也未改写。

## 命令

工作目录固定 `E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend`。以下为日志对应的参数；固定工具链，未使用 PATH 默认 Go：

```powershell
$env:GOROOT='E:/tools/go1.26.10/go'
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$go='E:/tools/go1.26.10/go/bin/go.exe'
$packages=@('./internal/service','./internal/handler','./internal/handler/dto','./internal/repository','./internal/server/middleware')

& $go test -json @packages -run KeyFast20261105 -count=1
& $go test -tags unit -json @packages -run KeyFast20261105 -count=1

$regressionPattern='GroupFast20261103|OpenAIFastPolicy|APIKeyAuthCache|ApplyOpenAIFastPolicy|WSResponseCreate|PolicyEnforcingFrameConn|APIKeyService_.*(Cache|Snapshot)|APIKeyUpdate_|UpdateQuotaUsed_Exhausted|TokenMultiplierCap|ServiceTierCostMultiplier|CalculateCostWithServiceTier|OpenAICodexRoutingHint|OpenAIOAuthHTTPBuildersSendRoutingHint|OpenAIWSHeadersSendOAuthRoutingHint|ForwardAsRawChatCompletions_.*(Usage|Disconnect|Cancel)|ForwardResponses_ForceChatCompletions'
& $go test -json @packages -run $regressionPattern -count=1
& $go test -tags unit -json @packages -run $regressionPattern -count=1

& $go test -overlay E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/baseline-overlay.json -tags unit -json ./internal/service -run '^Test(ForwardAsRawChatCompletions_ForcesStreamUsageUpstreamAndPassesUsageDownstream|GroupFast20261103AuthSnapshotAndDuplicate)$' -count=1

& $go test -json @packages -run GroupFast20261103 -count=1
& $go test -tags unit -json @packages -run GroupFast20261103 -count=1

& $go build -o E:/codex-runtime/sub2api/key-fast-policy-20261105/server.exe ./cmd/server
```

构建于 R3 前成功，日志 `build.log`。原 31 文件 gofmt 和 scoped diff 检查成功，日志 `gofmt-check.log`、`diff-check.log`。R3 对唯一变动旧测试执行 `E:/tools/go1.26.10/go/bin/gofmt.exe -l`，并从候选根执行 `git -c core.safecrlf=false diff --check -- <changed-paths.txt 中32条路径>`；两者退出 0、日志为空，记录在 `gofmt-r3-check.log`、`diff-r3-check.log`、`r3-checks.json`。

## 回归失败与基线边界

1. **GroupFast 为本次版本升级引起的断言冲突，不能归为既存失败。** `TestGroupFast20261103AuthSnapshotAndDuplicate` 在 `E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_policy_group_fast_20261103_test.go:141` 原错误为 `expected: 20 / actual: 21`；default/unit 的原失败日志保留。原始输入 overlay 中此测试 PASS。经 R3 独立合同批准后仅将 20→21；本轮 GroupFast default 158、unit 170 PASS 已覆盖修订，其他断言未动。

2. **RawChat 有独立原始输入复现证据。** `TestForwardAsRawChatCompletions_ForcesStreamUsageUpstreamAndPassesUsageDownstream` 在 `E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_gateway_chat_completions_raw_test.go:167` 报 `stream usage incomplete: missing terminal event`。候选 unit 回归及原始输入 overlay 中路径、行号、错误一致；该旧测试及 terminal 协议未改。此项仍是保留的基线失败，不宣称已修复或当前全部回归零失败。

3. 主控早先完整旧基线 `GroupFast20261103|OpenAIFastPolicy|APIKeyAuthCache` 为 191 个 PASS 事件，日志 `E:/codex-runtime/sub2api/key-fast-policy-20261105/baseline-complete-backend.jsonl`，计数 `E:/codex-runtime/sub2api/key-fast-policy-20261105/baseline-complete-counts.json`。该 regex **不包含 RawChat terminal 测试**，因此 RawChat 的既存归因依据是单独 overlay 复现，而不是上述 191 事件。原 missing-prerequisite 失败日志仍保留在 `E:/codex-runtime/sub2api/key-fast-policy-20261105/baseline-backend.jsonl`。

## Overlay 与源码一致性证据

基线复现使用 **Go 临时 `-overlay`，没有在候选磁盘上替换源码，也不存在“替换后恢复”步骤**。

- `E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/baseline-overlay.json` 将 22 个已修改既有 Go 文件映射到证据目录 `original/` 的原字节，将 8 个新增 Go 文件映射到空 package 文件；原 30 项映射保持不变。
- 已逐一核验 22 个原字节映射的 SHA256 与 `E:/codex-runtime/sub2api/key-fast-policy-20261105/candidate-before.json` 对应值相同。
- overlay 运行后生成的旧 31 文件 manifest 已归档；本轮 R3 前全部 31 个当前文件哈希再次与该 manifest 一致，R3 后这 31 文件也完全不变。证据是 `source-manifest-r2-frozen.json`、当前 `source-manifest.json` 和 `r3-integrity.json`。
- 为使 R3 后还可复用原始基线，另存 `E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-backend/baseline-overlay-r3-repro.json`：保留原 30 映射，仅追加旧 GroupFast 测试至本轮归档的原 20 断言文件。它是复现准备文件，**本轮未执行**；早期 overlay 和日志均未覆盖。R3 后若独立 QA 重现原始基线，应使用这一派生 overlay。

## 变更源码清单

以下仅含本任务拥有增量的 32 个后端文件；不是候选或主仓库所有 dirty 文件列表。逐文件最终字节与 SHA256 见产品 manifest。

```text
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/apikey.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/apikey/apikey.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/apikey/where.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/apikey_create.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/apikey_update.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/migrate/schema.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/mutation.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/runtime/runtime.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/ent/schema/api_key.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/handler/api_key_handler.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/handler/api_key_policy_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/handler/dto/api_key_policy_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/handler/dto/mappers.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/handler/dto/types.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/pkg/ctxkey/ctxkey.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/repository/api_key_policy_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/repository/api_key_repo.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/server/middleware/api_key_auth.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/server/middleware/api_key_policy_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key_auth_cache.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key_auth_cache_impl.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key_fast_policy.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key_policy_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/api_key_service.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_gateway_chat_completions_raw.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_gateway_responses_chat_fallback.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_gateway_service.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_http_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_policy_group_fast_20261103_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/internal/service/openai_ws_key_fast_20261105_test.go
E:/codex-worktrees/sub2api/key-fast-policy-20261105/backend/migrations/255_api_key_openai_fast_policy.sql
```

此外只交付本报告 `E:/codex-worktrees/sub2api/key-fast-policy-20261105/docs/workflow/worker-results/key-fast-policy-20261105-backend.md`；仓库外生成缓存、原字节、日志与复现脚本均为证据，不计入产品写集。

## 范围保护、风险与交接

- 本轮比较候选原始清单中的其余 **2744 个 backend 文件**，哈希均未变；六个旧 untracked prerequisite 与 `untracked-prerequisites.json` 的原始 SHA256 全部一致，均不在交付 manifest。证据 `r3-integrity.json`。
- 未修改 frontend、主仓库、共享合同/评审/状态文件、旧 prerequisite、module/deps、计费价格；未提交、推送、部署或运行真实数据库迁移。
- 原始失败日志、R1/R2 测试失败和最终成功日志保留；`compile-initial.jsonl` 的 no-tests-to-run 不计为通过。
- 未验收：真实 PostgreSQL 迁移/数据库往返、生产 Redis/provider、生产部署、race 及独立 QA 的更广端到端验证。既存 RawChat terminal 失败继续留给主控单独处理。
- **源码停止修改，交独立 QA**。任何后续源码修改先通知主控停止 QA，并重新生成受影响冻结证据。
- `knowledge_candidates`：鉴权 snapshot 新字段需同步版本及旧版本 L2→L1 提升门禁；raw Chat 的 tier metadata 应取最终策略 body。仅供主控审核，未写入知识库或全局 memory。
