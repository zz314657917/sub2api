---
phase: build
current_sprint: model-market-catalog-v9
total_sprints: 1
pending_action: close-model-market-catalog-v9
project_type: web
qa_mode: runtime
approval_required: false
last_verified: 2026-06-06
---

# Workflow Status

- 当前阶段：`build`
- 当前 Sprint：`model-market-catalog-v9`
- 当前目标：把 6 月初已经进入主线的模型市场公开目录、后台目录维护、分组倍率换算、`¥` 展示口径，以及 Doubao / 视频官方价同步收口成稳定默认入口。
- Task contract：当前没有单独 sprint contract；以 `knowledge/tasks/current-task.md` 和近期目标测试记录为事实源。
- 本次结论：进行中。`knowledge/tasks/current-task.md` 已推进到 `version=11`、模型市场后台维护和公开 `/models` 展示链路，并补齐 `gpt-image-2-official` 人民币余额扣费倍率。
- 当前已稳定进入默认主线的事实：
  - `/models` 已公开可访问，不再依赖登录。
  - 前台与后台都不展示 APIMart 品牌文案，但内部仍保留上游兼容命名。
  - 模型市场目录由后台 `model_market_catalog` 管理，不再以前台硬编码为主。
  - 公开价展示已切到 `¥`，并支持 `模型市场分组倍率 × 账号分组有效倍率` 的“我们的价格”换算。
  - `gpt-image-2-official`、Doubao Seedance 图像/视频，以及其他视频分组的展示规则、隐藏列策略和默认目录版本都已推进到 `version=11`。
  - `gpt-image-2-official` 模型市场目录行价本身保存 APIMart 官方参考价，例如 `default=¥0.2109`、`2576x3216 · 中=¥0.1408`；前台展示加价由后台模型市场分组 `price_multiplier` 控制，不在目录行价里预乘。
  - `gpt-image-2-official` 的 APIMart 上游真实任务 cost 按 `data.cost × 7 × 1.2 × 账号分组图片有效倍率` 扣用户余额。
- 目标验证命令：
  - `go test -tags=unit ./internal/service ./internal/handler -run "TestSettingService_GetModelMarketCatalog|TestSettingService_SetModelMarketCatalog|TestNormalizeModelMarketCatalog|TestSettingHandler_GetModelMarketCatalog" -count=1`
  - `corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/public-pages.spec.ts src/__tests__/public-smoke.spec.ts`
  - `corepack.cmd pnpm --dir frontend run typecheck`
  - `corepack.cmd pnpm --dir frontend run build`
  - `git diff --check`
- 运行态回读：
  - `http://127.0.0.1:8080/health`
  - `http://127.0.0.1:8080/api/v1/model-market/catalog`
  - 如本地容器/预览环境可用，补 `http://127.0.0.1:8080/models` 的最小页面回读。
- 下一合法动作：关闭本轮模型市场目录 Sprint，或继续在后台目录 JSON 中补新分组/新规格，不再回退到前台模板硬编码。
- 状态推进规则：先 `spec-approved`，再进入当前 Sprint 的 `contract-draft -> contract-approved -> build -> qa -> fix -> retest -> done`。
