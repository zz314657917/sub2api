---
repo: sub2api
project_type: web
qa_mode: runtime
last_verified: 2026-06-17 11:20 +08:00
---

# Workflow Spec

## S18 Addendum: APIMart task webhook

### 一句话目标

- 在不改变普通图片同步代理兼容行为的前提下，为 APIMart 视频/长任务接入任务完成 webhook，让 Sub2API 能在任务终态时主动完成本地状态落库、结算和失败退款。

### 当前背景

- Sub2API 当前仍是 Studio Bridge / 落叶AI的账号、余额、分组和扣费真源。
- `chatgpt2api` / 落叶创作台负责任务体验，但任务成功扣费、失败退款和使用记录最终应回到 Sub2API 侧闭环。
- APIMart task webhook 只在任务 `completed` / `failed` 等终态后回调；因此它适合补强长任务可靠结算，不适合作为普通同步 OpenAI 图片接口的直接替代。
- 本地视频任务已经有 `openai_video_tasks`、预扣、`/v1/tasks/:task_id` 查询结算和失败退款逻辑，S18 应复用这条链路，而不是新增并行账本。

### 明确不在范围内

- 不把普通 `/v1/images/generations` 同步响应改为异步 `task_id` 返回。
- 不覆盖客户请求里已经带的 `webhook` 字段；客户 webhook fan-out/relay 另开任务。
- 不新增数据库迁移，不写真实公网域名或 secret，不改 Studio Bridge / chatgpt2api 协议。
- 不整体重构 Image Creator 或 APIMart 图片轮询逻辑。

### 验收标准

- webhook 接收端有 secret 校验、body 大小限制、脱敏日志和幂等处理。
- APIMart 视频任务只在配置完整且请求未带客户 webhook 时注入 Sub2API callback URL。
- 成功终态只结算一次；失败终态只退款一次；重复回调不重复扣退。
- 现有 `/v1/tasks/:task_id` 查询结算仍作为兜底可用。
- 定向后端测试和 `git diff --check` 通过，denied-path audit 不触碰图片同步代理、迁移、Ent、公共页、支付页、Canvas、Studio 前端等禁止范围。

## S19 Draft Addendum: upstream v0.1.137 postfixes

### 一句话目标

- 在 S15-S17 已完成的基础上，继续筛出 `v0.1.137` 中不碰迁移、不碰前端、不覆盖本地产品定制的后端小修，作为 S18 之后的候选小步迁移。

### 当前背景

- 上游 `v0.1.137` 的安全/兼容主干已经通过 S15/S16/S17 小步迁入。
- 仍有少量后端补丁有合并价值，但不应和当前 S18 APIMart webhook 实现混在一起。
- 当前 S19 只是 follow-up contract 草案；`docs/workflow/status.md` 仍以 S18 `contract-draft` 为当前合法动作。

### 候选范围

- OpenAI failover 复用原始错误体。
- Anthropic window cooldown 保留。
- Account repository 列表参数限制与 refresh candidates SQL 修复。

### 明确不在范围内

- 不整体 merge/rebase `upstream/main`。
- 不碰 Ent、migrations、VERSION、wire 生成物或前端。
- 不把 OpenAI image failover、token refresh retry amplification、OAuth promo signup、scheduler outbox dedup/cleanup、cyber policy、channel monitor jitter、Claude OAuth system prompt blocks 混入本轮。

### 验收标准

- 定向后端 service / repository / server contract 测试通过。
- `git diff --check` 通过。
- denied-path audit 返回 `NO_DENIED_PATHS`。
- worker/result 和 QA report 说明每个上游候选是 `ported`、`equivalent` 还是 `skipped`。

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
- 以上边界仅适用于 S15-S17 这条上游小步迁移链路；后续统一 API Key / APIMart 图片网关 / 前端导航合并已经触达 `wire_gen.go`、Studio Bridge repo、公共页、模型市场、Keys 和 Settings 等路径，不能再用本节作为当前 `origin/main..HEAD` 的 denied-path 证明。
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
- 上游小步 Sprint 的 denied-path audit 应返回 `NO_DENIED_PATHS`；若当前批次是产品合并或 UI/网关主线合并，则必须改为列出实际触达路径和对应验证，不能沿用旧审计结论。
- lockfile 扫描无已知需规避版本残留，例如 `form-data@4.0.5`。
- 迁移结果不触碰本轮禁止路径；若后续合并触达曾经的禁止路径，workflow/knowledge 必须明确说明这是新批次范围，而不是继续复用旧 Sprint 证据。
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
