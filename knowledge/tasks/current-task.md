# 当前任务快照

最后更新：2026-06-03 02:03 +08:00

## 背景

- 仓库主工作区：`F:/mcplugins/sub2api`。
- 本轮 upstream 同步只在独立 worktree：`E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`。
- 当前分支：`codex/upstream-main-oauth-401-no-credentials-write-test-s2i`。
- 目标仍是分批同步 `upstream/main` 的稳定修复，不直接 merge `upstream/main`。
- 继续遵守 P/G/E：先 contract，再实现，再 QA 证据和 handoff。

## 当前目标

- 在不触碰 schema、迁移、frontend、gateway handler、OpenAI WS/Responses bridge 的前提下，同步小范围上游稳定修复。
- 本次 S2i 目标：test-only port upstream `be3613593 test(oauth): update OAuth 401 tests to match new no-write behavior`。

## 本次已完成

- S2g 已完成并提交：
  - `00a11a676 docs: approve claude count_tokens sync`
  - `9083dd5fd fix(gateway): allow claude count_tokens validation`
- S2h 已完成并提交：
  - `39d6f22aa docs: approve openai failover body remap sync`
  - `2624cda59 fix(openai): reparse failover request body mappings`
- S2h 行为变化：
  - `getOpenAIRequestBodyMap` 不再读取或写入 `OpenAIParsedRequestBodyKey`。
  - Service 层每次按传入 `body` 重新解析请求 map。
  - OpenAI failover 第二个账号不会复用第一个账号已重写过 `model` 的 legacy context cache。
  - Handler 侧 `OpenAIParsedRequestBodyKey` 常量和预校验/Claude Code helper 用途保留。
- 新建并提交 S2i contract：
  - `c4e409517 docs: approve oauth 401 no-write test sync`
  - contract 文件：`docs/workflow/tasks/upstream-main-oauth-401-no-credentials-write-test-s2i.md`
- 实现 S2i：
  - `backend/internal/service/ratelimit_service_401_test.go`
- S2i 行为/测试变化：
  - 生产 `ratelimit_service.go` 未改；本地实现已等价 upstream `6aec50501`，OAuth 401 不回写 credentials JSONB。
  - `OAuth401InvalidatorError` 断言 `updateCredentialsCalls == 0`。
  - 原 `OAuth401UsesCredentialsUpdater` 改为 `OAuth401DoesNotOverwriteCredentials`，断言不持久化请求开始时的 credentials 快照，同时仍设置 temp-unschedulable。
- 写入 QA/结果证据：
  - `docs/workflow/worker-results/upstream-main-oauth-401-no-credentials-write-test-s2i-result.md`
  - `docs/workflow/qa-reports/upstream-main-oauth-401-no-credentials-write-test-s2i-qa.md`

## 已确认事实

- 当前本地与 `upstream/main` 差异仍很大：`git rev-list --left-right --count HEAD...upstream/main` 在 S2i contract 后为 `308 370`。
- 计数不会因手工语义 port 自动下降；当前同步采用 contract + QA 证据而不是直接 merge。
- 这些候选已确认本地等价或无需重复 port：
  - `a6117429`, `26ca73a`, `2c14efeaa`, `6acb46c11`, `1d47fd630`, `b15375dfb`, `56e96fdd8`
  - `f1cc83e0e`, `a66f771cb`, `0cfabaa82`
  - `0a521f09f`, `20f534078`, `89dffdd2e`, `6010c3cca`, `1e6d0b602`
  - `888cd8092`, `d3d5843b9`
- 这些候选已明确延后：
  - `a39163519`：OpenAI key generated config 默认模型升级到 `gpt-5.5`，属于产品/配置策略。
  - `003b2786d`：目标测试文件属于 deferred apicompat bridge 测试链。
  - `08e19bb15`, `d7bed40dd`, `08061717b`：OpenAI WS bridge/failover 规模较大。
  - `5fd9a3509`：当前本地 pricing resource 仍匹配旧断言，不能只改测试。
- 这些候选本轮补充确认已等价或已处理：
  - `32ea9cfe`：API Key Responses SSE body fallback，本地已有 `TestHandleNonStreamingResponse_APIKeyFallsBackToSSEBodyWhenContentTypeIsWrong` 和对应实现。
  - `b9509e823`, `ed2aac25a`：long-context cache_read/cache_creation multiplier，本地 billing 实现和测试已存在。
  - `6aec50501`：OAuth 401 不覆盖 credentials JSONB，本地生产实现已存在；S2i 只同步了后续测试语义。
  - `f597c1581`：group custom `/v1/models` list 已在 migration Sprint 处理，不在小补丁流重复 port。

## 待验证点

- S2i 实现/QA 提交后需再确认 `git status --short --branch` clean。
- 若继续下一批，需要重新从 `git log --cherry-pick --right-only HEAD...upstream/main --no-merges` 里筛候选，先判等价再写 contract。
- Docker runtime smoke 已在 S1 做过；S2g 是纯 service validator 小修，本轮未重跑 Docker。
- S2i 是 unit-tag test-only 同步，本轮未重跑 Docker runtime smoke 或全量后端测试。

## 当前结论

- S2i 已完成实现和目标 QA，准备提交实现/QA commit。
- 本轮没有触碰主工作区 `F:/mcplugins/sub2api`。
- 当前仍不建议直接 merge `upstream/main`；剩余大项主要是迁移、payment/subscription/channel-monitor 功能、OpenAI WS/Responses bridge 和 gateway 重构链。

## 下一步

- 提交 S2i 实现/QA：验证 `git status --short --branch` clean。
- 如继续同步，优先寻找文件少、无迁移、无 schema、无 bridge 链依赖的候选；每个候选先判本地等价。
- 大功能或迁移型补丁单独开 Sprint，不纳入当前小补丁批次。

## 验证记录

- `git diff --check`：通过。
- `go test -tags unit ./internal/service -run "OAuth401InvalidatorError|OAuth401DoesNotOverwriteCredentials|OAuth401NoRefreshToken" -count=1`：通过。
