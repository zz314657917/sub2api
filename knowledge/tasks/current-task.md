# 当前任务快照

最后更新：2026-06-03 02:41 +08:00

## 背景

- 仓库主工作区：`F:/mcplugins/sub2api`。
- 本轮 upstream 同步只在独立 worktree：`E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`。
- 当前分支：`codex/upstream-main-openai-ws-usage-dedup-s2k`。
- 目标仍是分批同步 `upstream/main` 的稳定修复，不直接 merge `upstream/main`。
- 继续遵守 P/G/E：先 contract，再实现，再 QA 证据和 handoff。

## 当前目标

- 在不触碰 schema、迁移、frontend、gateway handler、OpenAI WS/Responses bridge 的前提下，同步小范围上游稳定修复。
- 本次 S2k 目标：port upstream `1e2193c3d fix: avoid websocket usage dedup conflicts` 的 service/test 子集。

## 本次已完成

- S2g 已完成并提交：
  - `00a11a676 docs: approve claude count_tokens sync`
  - `9083dd5fd fix(gateway): allow claude count_tokens validation`
- S2h 已完成并提交：
  - `39d6f22aa docs: approve openai failover body remap sync`
  - `2624cda59 fix(openai): reparse failover request body mappings`
- S2i 已完成并提交：
  - `c4e409517 docs: approve oauth 401 no-write test sync`
  - `5e129b6ac test(oauth): assert 401 handler preserves credentials`
- S2j 已完成并提交：
  - `ae0b3bfc9 docs: approve openai oauth refresh enrichment sync`
  - `b9646e6eb fix(openai): enrich oauth refresh credentials`
  - `de1c7fee7 docs: update upstream sync handoff`
- S2k 已完成并提交：
  - `6a926e2a0 docs: approve openai ws usage dedup sync`
  - `332ae3d4d fix(openai): avoid ws usage dedup conflicts`
  - contract 文件：`docs/workflow/tasks/upstream-main-openai-ws-usage-dedup-s2k.md`
  - `backend/internal/service/openai_gateway_service.go`
  - `backend/internal/service/openai_gateway_record_usage_test.go`
- S2k QA/结果证据已提交：
  - `docs/workflow/worker-results/upstream-main-openai-ws-usage-dedup-s2k-result.md`
  - `docs/workflow/qa-reports/upstream-main-openai-ws-usage-dedup-s2k-qa.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`

## 已确认事实

- 当前本地与 `upstream/main` 差异仍很大：最后一次状态检查中 `git rev-list --left-right --count HEAD...upstream/main` 输出为 `317 370`；后续若继续追加 handoff/doc 提交，左侧计数会随本地提交自然增加。
- 计数不会因手工语义 port 自动下降；当前同步采用 contract + QA 证据而不是直接 merge。
- S2k 已确认可以收敛为 `OpenAIGatewayService.RecordUsage` 小修：
  - 非 WS OpenAI usage 仍优先使用 `ctxkey.ClientRequestID`。
  - `OpenAIWSMode=true` 且 `Result.RequestID` 非空时，billing/log request id 使用 upstream response id，避免同一 WS connection 多 turn 去重冲突。
  - 本地 `RequestIDOverride` 仍在最后覆盖。
- 本轮补充确认等价，无需重复 port：
  - `8a999f438`：WS terminal events 已不再被 token event 分类，本地已有对应测试。
  - `2bd3125d`：usage worker context 保留已在本地，含 `wrapUsageRecordTaskContext` 和 request id/client request id 测试。
  - `df2b02e61`：group account available/rate-limited count 口径本地已等价，`GetByID`/`GetAccountCount`/`loadAccountCounts` 已共用调度可用口径。
  - `69305a609`：ops 本地客户端限制错误 SLA 排除本地已等价，含 API key auth/business-limit/upstream 反例测试。
- 这些候选已确认本地等价或无需重复 port：
  - `a6117429`, `26ca73a`, `2c14efeaa`, `6acb46c11`, `1d47fd630`, `b15375dfb`, `56e96fdd8`
  - `f1cc83e0e`, `a66f771cb`, `0cfabaa82`
  - `0a521f09f`, `20f534078`, `89dffdd2e`, `6010c3cca`, `1e6d0b602`
  - `888cd8092`, `d3d5843b9`, `a9c7a3a09`
  - `32ea9cfe`, `b9509e823`, `ed2aac25a`, `6aec50501`, `0daf0e613`, `b65dde634`, `27600b1d2`
- 这些候选已明确延后：
  - `a39163519`：OpenAI key generated config 默认模型升级到 `gpt-5.5`，属于产品/配置策略。
  - `003b2786d`：目标测试文件属于 deferred apicompat bridge 测试链。
  - `08e19bb15`, `d7bed40dd`, `08061717b`, `2a075a85b`：OpenAI WS bridge/failover/WS image tool 注入规模较大。
  - `5fd9a3509`：当前本地 pricing resource 仍匹配旧断言，不能只改测试。
  - `0560340bd`：admin create-user balance pointer 触及 DTO、默认余额语义和前端表单，需单独评估。
  - `a01686c63`, `a31b50748`, `33ac8eb27`, `ddf91e9a7`, `ed1b57c59`, `f7ac5e593`, `1e406fed5`, `0f8e2d093`, `bb4c1abe2`：涉及多模块 gateway/scheduler/config/frontend、API/DTO 安全响应形态、Ent/migration 或较大协议语义，需单独 Sprint 评估。
  - `cbdfedab3`：AES encryptor test-only 候选，属于低风险测试补强但不是 Sprint 1 指定 OpenAI/usage/ops 修复；可单独 test-only Sprint 处理。
  - `825834b5c`：admin/settings contract 小测试修复；当前本地 API contract 搜索未确认完全等价，若要处理可另开 test-only Sprint。
  - `e1b53fdeb`：notification email helper 写入错误检查；本地当前无 `notification_email_service.go` 同名路径，属于上游邮件通知链，随邮件功能 Sprint 评估。

## 待验证点

- S2k 实现/QA 和收尾 handoff 提交后 `git status --short --branch` 已确认 clean。
- S2k 未重跑 Docker runtime smoke 或全量后端测试；本轮只跑目标 service tests。
- 若继续下一批，需要重新从 `git log --cherry-pick --right-only HEAD...upstream/main --no-merges` 里筛候选，先判等价再写 contract。

## 当前结论

- S2k 已完成实现、目标 QA 和提交。
- 本轮没有触碰主工作区 `F:/mcplugins/sub2api`。
- 当前仍不建议直接 merge `upstream/main`；剩余大项主要是迁移、payment/subscription/channel-monitor 功能、OpenAI WS/Responses bridge 和 gateway 重构链。

## 下一步

- 如继续同步，先确认 `git status --short --branch` clean。
- 如继续同步，优先寻找文件少、无迁移、无 schema、无 bridge 链依赖的候选；每个候选先判本地等价。
- 大功能或迁移型补丁单独开 Sprint，不纳入当前小补丁批次。

## 验证记录

- `git diff --check`：通过。
- `go test ./internal/service -run "OpenAIGatewayServiceRecordUsage_(PrefersClientRequestIDOverUpstreamRequestID|WSModePrefersUpstreamRequestIDOverClientRequestID|GeneratesRequestIDWhenAllSourcesMissing)" -count=1`：通过。
