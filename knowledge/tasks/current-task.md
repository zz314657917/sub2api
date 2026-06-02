# 当前任务快照

最后更新：2026-06-03 02:17 +08:00

## 背景

- 仓库主工作区：`F:/mcplugins/sub2api`。
- 本轮 upstream 同步只在独立 worktree：`E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`。
- 当前分支：`codex/upstream-main-openai-oauth-refresh-enrichment-s2j`。
- 目标仍是分批同步 `upstream/main` 的稳定修复，不直接 merge `upstream/main`。
- 继续遵守 P/G/E：先 contract，再实现，再 QA 证据和 handoff。

## 当前目标

- 在不触碰 schema、迁移、frontend、gateway handler、OpenAI WS/Responses bridge 的前提下，同步小范围上游稳定修复。
- 本次 S2j 目标：port upstream `eba204632 fix: enrich OpenAI OAuth token refresh` 的可收敛 service/wiring 子集。

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
- S2j contract 已完成并提交：
  - `ae0b3bfc9 docs: approve openai oauth refresh enrichment sync`
  - contract 文件：`docs/workflow/tasks/upstream-main-openai-oauth-refresh-enrichment-s2j.md`
- S2j 实现已完成，待提交：
  - `backend/internal/service/openai_oauth_service.go`
  - `backend/internal/service/openai_privacy_service.go`
  - `backend/internal/service/openai_oauth_service_refresh_test.go`
  - `backend/internal/service/openai_subscription_test.go`
  - `backend/internal/service/wire.go`
  - `backend/cmd/server/wire_gen.go`
- S2j QA/结果证据已写入，待提交：
  - `docs/workflow/worker-results/upstream-main-openai-oauth-refresh-enrichment-s2j-result.md`
  - `docs/workflow/qa-reports/upstream-main-openai-oauth-refresh-enrichment-s2j-qa.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`

## 已确认事实

- 当前本地与 `upstream/main` 差异仍很大：S2j contract 提交后 `git rev-list --left-right --count HEAD...upstream/main` 为 `310 370`。
- 计数不会因手工语义 port 自动下降；当前同步采用 contract + QA 证据而不是直接 merge。
- S2j 已确认可以收敛为 OpenAI OAuth/privacy service + 最小 Wire provider 补丁：
  - no-refresh-token existing access-token path 会保留 `subscription_expires_at` 并运行 enrichment。
  - `fetchChatGPTSubscriptionExpiresAt` 使用 `httptest` 覆盖，不访问真实 ChatGPT 服务。
  - `wire.go` 新增 `ProvideOpenAIOAuthService`，`wire_gen.go` 只做对应 provider 调用和 `privacyClientFactory` 位置调整。
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

## 待验证点

- S2j 实现/QA 提交后需再确认 `git status --short --branch` clean。
- S2j 未重跑 Docker runtime smoke 或全量后端测试；本轮只跑目标 service tests 和 `cmd/server` 编译 smoke。
- 若继续下一批，需要重新从 `git log --cherry-pick --right-only HEAD...upstream/main --no-merges` 里筛候选，先判等价再写 contract。

## 当前结论

- S2j 已完成实现和目标 QA，准备提交实现/QA commit。
- 本轮没有触碰主工作区 `F:/mcplugins/sub2api`。
- 当前仍不建议直接 merge `upstream/main`；剩余大项主要是迁移、payment/subscription/channel-monitor 功能、OpenAI WS/Responses bridge 和 gateway 重构链。

## 下一步

- 提交 S2j 实现/QA：验证 `git status --short --branch` clean。
- 如继续同步，优先寻找文件少、无迁移、无 schema、无 bridge 链依赖的候选；每个候选先判本地等价。
- 大功能或迁移型补丁单独开 Sprint，不纳入当前小补丁批次。

## 验证记录

- `git diff --check`：通过。
- `go test ./internal/service -run "OpenAIOAuthService_RefreshAccountToken_NoRefreshTokenUsesExistingAccessToken|FetchChatGPTSubscriptionExpiresAt|OpenAI.*Refresh|OpenAI.*Privacy|OpenAI.*Subscription|BuildAccountCredentials|RefreshIfNeeded" -count=1`：通过。
- `go test ./cmd/server -run TestNonExistent -count=1`：通过。
