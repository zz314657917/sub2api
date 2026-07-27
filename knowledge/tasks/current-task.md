# 当前任务快照

最后更新：2026-07-28 00:58 +08:00

## 当前任务（S122）

- 修复 Sub2API 网关模型目录遗漏渠道定价模型的问题，使 chatgpt2api Studio Bridge 能读取仅在渠道定价中配置的具体模型，例如 `kimi-k3`。
- 改动位于隔离工作树 `E:/codex-worktrees/sub2api/channel-pricing-model-catalog-s122`，不触碰主工作树现有用户改动和其它 worktree。
- Contract：`docs/workflow/tasks/channel-pricing-model-catalog-s122.md`。
- QA：`docs/workflow/qa-reports/channel-pricing-model-catalog-s122-qa.md`。

## 本次完成

- `GatewayService.GetAvailableModels` 在账号 `model_mapping` 之外，合并当前 group/platform 活跃渠道的 `SupportedModels()`。
- pricing-only 具体模型会进入 `/v1/models` 与 `/v1/model-catalog`；通配符、空值、其它平台模型不会进入。
- 平台过滤后仍要求存在可调度账号；没有账号、没有活跃渠道或渠道读取失败时保留原回退行为。
- 未修改计费、渠道限制、模型映射执行、路由选择、数据库、前端、部署或容器。

## 验证记录

- `TestGetAvailableModels*` 定向回归 `count=10`：PASS。
- `internal/handler` 全包：PASS。
- chatgpt2api Sub2API Canvas 模型目录三组回归：PASS。
- `internal/service` 全包仅复现既有五组 `TestPeakMultiplier*` 聚合时区/顺序失败；这些测试单独运行 PASS。
- `gofmt`、精确路径、冲突标记、未合并索引和 `git diff --check`：PASS。

## 当前结论

- `PASS / published`：功能提交 `ee5f8abbe` 已推送到 `origin/main`，远端一致性已验证。
- 尚未使用真实渠道数据做认证态 Studio Bridge 联调，也未部署或更新容器。

## 下一步

1. 部署 Sub2API 后，用真实渠道定价中的 `kimi-k3` 调用 `/v1/model-catalog`，再刷新 chatgpt2api Canvas 模型列表。
2. 不把主工作树现有用户改动或其它 worktree 内容混入 S122。
