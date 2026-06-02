# 当前任务快照

最后更新：2026-06-03 01:42 +08:00

## 背景

- 仓库主工作区：`F:/mcplugins/sub2api`。
- 本轮 upstream 同步只在独立 worktree：`E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`。
- 当前分支：`codex/upstream-main-claude-count-tokens-s2g`。
- 目标仍是分批同步 `upstream/main` 的稳定修复，不直接 merge `upstream/main`。
- 继续遵守 P/G/E：先 contract，再实现，再 QA 证据和 handoff。

## 当前目标

- 在不触碰 schema、迁移、frontend、gateway handler、OpenAI WS/Responses bridge 的前提下，同步小范围上游稳定修复。
- 本次 S2g 目标：语义 port upstream `bf3787de1 fix(gateway): allow Claude Code count_tokens`。

## 本次已完成

- 新建并提交 S2g contract：
  - `00a11a676 docs: approve claude count_tokens sync`
  - contract 文件：`docs/workflow/tasks/upstream-main-claude-count-tokens-s2g.md`
- 实现 S2g：
  - `backend/internal/service/claude_code_validator.go`
  - `backend/internal/service/claude_code_validator_test.go`
- 行为变化：
  - `/v1/messages/count_tokens` 在 Claude Code User-Agent 匹配后可跳过普通 `/v1/messages` 的 system prompt/header/metadata 严格校验。
  - 非 Claude Code User-Agent 仍被拒绝。
  - 普通 `/v1/messages` 严格校验未放宽。
- 写入 QA/结果证据：
  - `docs/workflow/worker-results/upstream-main-claude-count-tokens-s2g-result.md`
  - `docs/workflow/qa-reports/upstream-main-claude-count-tokens-s2g-qa.md`

## 已确认事实

- 当前本地与 `upstream/main` 差异仍很大：`git rev-list --left-right --count HEAD...upstream/main` 在提交前为 `304 370`。
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

## 待验证点

- S2g 实现提交后需再确认 `git status --short --branch` clean。
- 若继续下一批，需要重新从 `git log --cherry-pick --right-only HEAD...upstream/main --no-merges` 里筛候选，先判等价再写 contract。
- Docker runtime smoke 已在 S1 做过；S2g 是纯 service validator 小修，本轮未重跑 Docker。

## 当前结论

- S2g 已完成实现和目标 QA，准备提交实现/QA commit。
- 本轮没有触碰主工作区 `F:/mcplugins/sub2api`。
- 当前仍不建议直接 merge `upstream/main`；剩余大项主要是迁移、payment/subscription/channel-monitor 功能、OpenAI WS/Responses bridge 和 gateway 重构链。

## 下一步

- 提交 S2g 实现/QA：验证 `git status --short --branch` clean。
- 如继续同步，优先寻找文件少、无迁移、无 schema、无 bridge 链依赖的候选；每个候选先判本地等价。
- 大功能或迁移型补丁单独开 Sprint，不纳入当前小补丁批次。

## 验证记录

- `git diff --check`：通过。
- `go test ./internal/service -run ClaudeCodeValidator -count=1`：通过。
- `go test ./internal/service ./internal/handler -run "ClaudeCode|CountTokens" -count=1`：通过。
