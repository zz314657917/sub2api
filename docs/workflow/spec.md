---
repo: sub2api
project_type: web
qa_mode: runtime
last_verified: 2026-06-17 11:20 +08:00
---

# Workflow Spec

## 一句话目标

- 在不覆盖本地 Studio Bridge / 支付治理 / Canvas / 公共页等产品定制的前提下，把上游 `v0.1.137` 的低风险安全、兼容、计费兜底和管理员运维能力按独立 Sprint 小步迁入，并为后续继续评估候选 patch 保留清晰边界。

## 当前背景

- 本地 `sub2api` 已不再处于“直接跟上游 merge”阶段；`main` 与 `upstream/main` 分叉较大。
- 仓库当前产品主线仍是 Studio Bridge / 落叶AI生产联调、支付套餐与用户治理。
- 同期存在一条独立工程主线：从上游按 Sprint 迁入低风险 patch，但显式保护本地定制，不碰 Ent/migrations/VERSION，不整体 merge `upstream/main`。
- 2026-06-17 已完成三个连续 Sprint：
  - `upstream-main-v0137-safe-patches-s15`
  - `upstream-main-v0137-small-compat-s16`
  - `upstream-main-openai-quota-reset-s17`

## 当前范围

- 当前 workflow spec 只描述 2026-06-17 这条上游小步合成链路。
- 不覆盖 Studio Bridge / 支付套餐 / 用户 IP / 首充福利这类产品主线的完整需求文档。
- 不把“后续也许会迁”的候选 patch 说成已批准范围。

## 已完成 Sprint 范围

### S15: 安全 / 兼容 / 计费兜底

- 锁定前端 `form-data` 到 `4.0.6`。
- token refresh 增加不可重试错误分类。
- 上游响应支持 zstd。
- 非流式 2xx 非 JSON 与 SSE `event:error` 进入 failover，并保留原始错误体。
- `tool strict` 缺省补 `false`。
- 国产模型 fallback pricing 和图像输入 token 计费补齐。
- DeepSeek `reasoning_effort=max` 归一到 `xhigh`。
- Anthropic thinking block 过滤改为按 mapped upstream model 分流。

### S16: 小兼容补丁

- Responses API sticky hash 在缺少旧字段时以 `input` 兜底。
- Claude Code `max_tokens=1` 的 Haiku 流式探测拦截补齐。
- OpenAI APIKey `/responses` probe 增加工具能力校验。
- API Key ACL 拒绝信息补充实际 client IP。

### S17: OpenAI OAuth 上游 quota/reset

- 新增管理员 OpenAI OAuth 账号上游 WHAM quota 查询入口。
- 新增 rate-limit reset credit 消费入口。
- 后端复用 token provider、privacy client factory 和账号代理解析。
- 前端仅在 OpenAI OAuth usage cell 展示上游 credits 查询/重置控件。

## 明确不在范围内

- 不整体 merge 或 rebase `upstream/main`。
- 不触碰 Ent schema、migrations、VERSION、wire 大链路生成物。
- 不覆盖本地 Studio Bridge、Canvas、支付页、公共页、模型市场、工单或 Chat/Image Studio 定制。
- 不引入 migration-heavy、compliance gate、cyber policy、渠道监控 jitter、Claude OAuth system prompt blocks 等高风险链路。
- 不把前端全量 Vitest 失败修复混进本轮 Sprint；这应另开前端稳定化任务。

## 当前稳定工程边界

- 当前允许的上游迁移策略是“低风险 patch + 独立 Sprint + 定向验证 + denied-path audit”。
- 每轮 Sprint 都必须显式说明：
  - 为什么该补丁适合独立迁移
  - 为什么不会覆盖本地产品面
  - 需要哪些定向测试证明行为稳定
  - 哪些更大链路明确跳过
- 当前稳定结论不是“本地已接近上游”，而是“本地已有一条可持续的小步迁移方法”。

## 验收标准

- patch 迁入后，定向后端测试通过。
- 涉及前端控件或管理页时，定向 Vitest 通过。
- `git diff --check` 通过。
- denied-path audit 返回 `NO_DENIED_PATHS`。
- lockfile 扫描无已知需规避版本残留，例如 `form-data@4.0.5`。
- 迁移结果不触碰本轮禁止路径。
- workflow 文档能说明“为什么这轮可以迁、为什么其他候选仍应跳过”。

## 当前证据入口

- `docs/workflow/tasks/upstream-main-v0137-safe-patches-s15.md`
- `docs/workflow/worker-results/upstream-main-v0137-safe-patches-s15-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-safe-patches-s15-qa.md`
- `docs/workflow/tasks/upstream-main-v0137-small-compat-s16.md`
- `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`
- `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`
- `docs/workflow/tasks/upstream-main-openai-quota-reset-s17.md`
- `docs/workflow/worker-results/upstream-main-openai-quota-reset-s17-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-quota-reset-s17-qa.md`

## 下一步候选

- 若继续做上游迁移，需另开 Sprint 单独评估候选 patch。
- 当前可评估但未批准的方向包括：
  - OpenAI image failover
  - Anthropic window cooldown
  - account list parameter batching
  - token refresh retry amplification / outbox dedup
- 若继续做前端全量测试收口，应另开“前端稳定化”任务，不与上游 patch Sprint 混合。
