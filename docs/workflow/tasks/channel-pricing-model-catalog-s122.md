---
task_id: channel-pricing-model-catalog-s122
status: contract-approved
owner: Codex
qa_mode: runtime
---

# Task Contract

## Goal

让 Sub2API 的 `/v1/models` 与 `/v1/model-catalog` 在账号模型映射之外，同时读取当前分组关联渠道中的具体定价模型，使 chatgpt2api Studio Bridge 能发现仅配置在渠道定价中的模型，例如 `kimi-k3`。

## Success Criteria

- 当前分组存在可调度账号时，模型目录合并账号 `model_mapping` 与渠道 `SupportedModels()`。
- 渠道定价中的具体模型即使没有账号模型映射，也会出现在 `/v1/models` 和 `/v1/model-catalog` 的候选中。
- 通配符、空模型和其它平台的定价模型不进入当前平台目录。
- 没有可调度账号、没有关联活跃渠道或渠道读取失败时，保留现有回退行为。
- 模型列表短缓存、排序和去重语义保持不变。

## Context

- Repo: `F:/mcplugins/sub2api`
- chatgpt2api 已通过 Sub2API 标准 `/v1/model-catalog` 和 `/v1/models` 读取模型，无需添加 `kimi-k3` 特判。
- `Channel.SupportedModels()` 已定义渠道支持模型为 mapping 与 pricing 的并集，并排除通配符。

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_hotpath_optimization_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/channel-pricing-model-catalog-s122-contract-review.md`
- `docs/workflow/tasks/channel-pricing-model-catalog-s122.md`
- `docs/workflow/worker-results/channel-pricing-model-catalog-s122-result.md`
- `docs/workflow/qa-reports/channel-pricing-model-catalog-s122-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- 计费金额计算、模型映射执行、渠道限制、账号调度、API Key 路由策略。
- 数据库 schema、migrations、frontend、deployment、containers 和 VERSION。
- 主工作树现有未提交改动及其它 worktree。

## Constraints

- 不硬编码 `kimi-k3` 或任何厂商模型特例。
- 只合并当前 group 和 platform 可见的具体渠道模型。
- 渠道目录补充不得让无可调度账号的分组看起来可用。
- 保持现有 15 秒模型目录缓存与显式失效接口不变。

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/channel-pricing-model-catalog-s122/backend
go test ./internal/service -run "TestGetAvailableModels" -count=1
go test ./internal/handler -run "TestGatewayModelCatalog|TestGatewayModels" -count=1
go test ./internal/service ./internal/handler -count=1
cd E:/codex-worktrees/sub2api/channel-pricing-model-catalog-s122
git diff --check
```

## Output

- 记录实现文件、测试结果、未验证风险和发布状态。

## Stop Rules

- 若必须修改计费、路由匹配、渠道限制、数据库或前端，停止并重新裁决。
- 若远端 `main` 在推送前发生变化，先重新验证，不强制推送。
