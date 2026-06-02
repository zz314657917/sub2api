# 当前任务快照

最后更新：2026-06-03 02:52 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 用户要求：模型广场给所有模型补 `官方/参考价`，并把视频模型改成厂商真实官方报价口径，不再使用 APIMart 或其他中转站价格冒充官方价。
- 当前改动聚焦公开页 `/models` 及其现有 `/channels/available` 数据源。
- `tmp-ui-check/` 仍是未跟踪截图证据目录，本轮未处理。

## 当前目标

- 在用户侧可用渠道模型 DTO 中返回 `reference_pricing`。
- 前端模型广场优先展示 `reference_pricing`，按输入/输出官方价格式化。
- 保留“我们的价格”和倍率折算逻辑，不让官方/参考价参与真实扣费或倍率折算。

## 本次已完成

- 后端 `SupportedModel` 增加 `ReferencePricing`，由 `PricingService.GetModelPricing()` 从现有 LiteLLM 官方/参考价资源合成。
- `/channels/available` 白名单 DTO 增加 `reference_pricing` 字段。
- `PricingService` 增加 Claude 小数/连字符别名兼容，例如 `claude-sonnet-4.5` 可匹配 `claude-sonnet-4-5`。
- 前端 `UserSupportedModel` 类型增加可选 `reference_pricing`。
- `ModelPlazaView.vue` 优先展示接口返回的官方/参考价，token 价格显示为 `输入 ¥x/M / 输出 ¥y/M`。
- 未登录 fallback 模型目录补了 OpenAI/Anthropic 聊天模型参考价，视频模型继续沿用原有细分参考价说明。
- 已按用户后续要求移除模型表格里的 `提示词缓存` 徽标，不再展示 prompt caching 标签。
- 已修复登录用户在渠道列表为空、页面使用 fallback 模型目录时的分组计数：模型卡片优先按真实 `availableGroups` 的同平台分组计算，不再把 fallback 示例里的 4 个 OpenAI 分组显示为用户可用分组。
- 视频模型官方价展示已改为厂商真实口径：
  - `kling-v3-omni`：可灵官方 Credit 消耗按 `¥0.098/Credit` 估算为 `¥0.588-¥1.568/秒`，细分 720P/1080P、音频、视频输入。
  - `kling-v2-6`：可灵官方 Credit 消耗按 `¥0.098/Credit` 估算为 `¥0.294-¥0.98/秒`，声音控制额外约 `¥0.196/秒`。
  - `wan2.7`：阿里百炼中国内地官方价 `¥0.6-1/秒`，细分 720P/1080P。
  - `veo3.1-fast`：Google Vertex AI 官方价 `$0.08-0.30/秒`，细分 720P/1080P/4K 和视频/视频+音频。
  - `doubao-seedance-2.0`：火山方舟官方 token 价 `¥28-51/M tokens`，细分含视频/无视频和分辨率。
- 模型广场底部说明改为“官方价按厂商公开口径展示，单位可能不同”，避免把 Credits/s、美元秒价、人民币 token 价混成同一价格单位。
- 已从模型广场展示层移除 OpenAI `gpt-5.2` / `gpt-5.3` 系列：fallback 目录不再列出这些模型，登录态 `/channels/available` 返回这些模型时也会在前端展示聚合前过滤。
- 视频模型规格价格展示已从横向 chips 改为逐行明细：官方/参考价和我们的价格各自按“规格名 + 价格”纵向排列，更接近参考站逐规格展示方式。
- 有逐行规格明细的视频模型不再显示价格列顶部汇总价，只保留每个规格自己的价格；无规格明细的普通模型仍显示单行价格。
- 已核对 APIMart 定价页视频分类，修正视频模型 fallback 的 `我们的价格`：原先误用了 APIMart 表格里的“官方价格”列，现在改为使用 APIMart “我们的价格”列，再按 `* 7` 折算人民币展示。
- 已按用户要求隐藏视频分区的 `官方/参考价` 列；聊天等非视频分区仍保留官方/参考价展示。
- 已按用户要求隐藏视频分区计费说明里的黄色官方参考说明，只保留 `按规格/秒计费` / `按规格/次计费` 这类计费方式。
- 已修正 `veo3.1-fast` 模型广场展示单位：Veo 视频模型按秒计价，页面不再特判为 `/次`。
- 已修复 `videoTierUnit(model)` 在统一返回 `/秒` 后产生的 vue-tsc 未使用参数错误。
- 已补充 Seedance 2.0 视频变体 fallback 展示价：`doubao-seedance-2.0-fast`、`doubao-seedance-2.0-fast-face`、`doubao-seedance-2.0-face`，与已有 `doubao-seedance-2.0` 一起按 APIMart “我们的价格”列折算人民币/秒展示。
- 已在模型广场四个 Seedance 2.0 变体名称下方补充用途差异：
  - `doubao-seedance-2.0`：标准版，质量优先，支持 1080P。
  - `doubao-seedance-2.0-fast`：快速版，适合草稿和批量试错。
  - `doubao-seedance-2.0-fast-face`：快速真人版，支持真人上传/人像素材，最高 720P。
  - `doubao-seedance-2.0-face`：真人版，支持真人上传/人像素材和 1080P。

## 已确认事实

- `reference_pricing` 只用于展示，不写入渠道价格，不影响 billing、倍率和扣费链路。
- 公开页的旧硬编码参考价仍作为视频模型和缺失接口字段的兜底。
- 视频模型的 fallback `我们的价格` 使用 APIMart 定价页“我们的价格”列的美元价，再按页面现有 `* 7` 逻辑折算人民币；`官方/参考价` 单独展示厂商官方口径，两者不强行同单位换算。
- 前端对旧后端响应兼容：`supportedModel.reference_pricing ?? null`。
- 本轮复用当前 `http://127.0.0.1:62080/models` 做浏览器检查，未保存截图。

## 验证记录

- `go test -tags=unit ./internal/service -run "TestFillReferencePricing|TestListAvailable|TestPricingNeedsFallback|TestSynthesizePricingFromLiteLLM|TestGetModelPricing_ClaudeDecimalAliasMatchesHyphenatedPricing" -count=1`：通过。
- `go test -tags=unit ./internal/handler -run "TestToUserSupportedModels|TestUserAvailableChannel_FieldWhitelist|TestBuildPlatformSections" -count=1`：通过。
- `go test ./internal/service ./internal/handler -run TestNonExistent -count=1`：通过，仅编译目标包，无测试运行。
- `corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/public-pages.spec.ts`：通过，9 个用例。
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/ChatStudioView.spec.ts`：通过，7 个用例；仅有 Browserslist 数据较旧提示。
- `corepack.cmd pnpm --dir frontend run typecheck`：通过。
- `git diff --check -- <本轮相关文件>`：通过。
- 移除提示词缓存徽标后，重新执行 `corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/public-pages.spec.ts`：通过，9 个用例。
- 移除提示词缓存徽标后，重新执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 修复 fallback 分组计数后，重新执行 `corepack.cmd pnpm --dir frontend exec vitest run src/__tests__/public-pages.spec.ts`：通过，9 个用例。
- 修复 fallback 分组计数后，重新执行 `corepack.cmd pnpm --dir frontend run typecheck`：通过。
- 修复 fallback 分组计数后，重新执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 视频模型官方价口径调整后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 视频模型官方价口径调整后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 视频模型官方价口径调整后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 移除 OpenAI 5.2/5.3 展示后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 移除 OpenAI 5.2/5.3 展示后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 移除 OpenAI 5.2/5.3 展示后，浏览器检查 `http://127.0.0.1:62080/models`：OpenAI 可见模型为 `gpt-5.5`、`gpt-5.4`、`gpt-5.4-mini`，`gpt-5.2/5.3` 系列为空。
- 移除 OpenAI 5.2/5.3 展示后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 将可灵 `Credits/s` 改为人民币/秒展示后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 将可灵 `Credits/s` 改为人民币/秒展示后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 将可灵 `Credits/s` 改为人民币/秒展示后，浏览器检查 `http://127.0.0.1:62080/models`：页面不再包含 `Credits/s`，`kling-v3-omni` 显示 `¥0.588-¥1.568/秒`，`kling-v2-6` 显示 `¥0.294-¥0.98/秒`。
- 将视频价格 chips 改为逐行明细后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 将视频价格 chips 改为逐行明细后，浏览器检查 `http://127.0.0.1:62080/models`：`kling-v3-omni` 官方价列为 6 行，我们价格列为 8 行。
- 将视频价格 chips 改为逐行明细后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 移除视频规格列顶部汇总价后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 移除视频规格列顶部汇总价后，浏览器检查 `http://127.0.0.1:62080/models`：`kling-v3-omni` 官方价和我们的价格列顶部 `.model-price-value` 数量均为 0，仍保留官方 6 行和我们 8 行明细。
- 移除视频规格列顶部汇总价后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 修正视频 fallback `我们的价格` 后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 修正视频 fallback `我们的价格` 后，浏览器检查 `http://127.0.0.1:62080/models`：`doubao-seedance-2.0 480P-input` 显示 `¥0.308/秒`，`kling-v3-omni default` 显示 `¥0.4704/秒`，符合 APIMart “我们的价格”列美元价按 `* 7` 折算。
- 修正视频 fallback `我们的价格` 后，执行 `npm.cmd run build`：通过；仅有项目既有 Vite chunk、Browserslist 和 Node `DEP0190` 警告。
- 修正视频 fallback `我们的价格` 后，执行 `git diff --check`：通过，仅有 `knowledge/tasks/current-task.md` 的既有 CRLF warning。
- 隐藏视频分区 `官方/参考价` 列后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 隐藏视频分区 `官方/参考价` 列后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 隐藏视频分区 `官方/参考价` 列后，浏览器检查 `http://127.0.0.1:62080/models`：Video 分区表头为 `模型 / 我们的价格 / 计费说明`，首行不再显示官方参考价明细。
- 隐藏视频黄色官方参考说明后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 隐藏视频黄色官方参考说明后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 隐藏视频黄色官方参考说明后，浏览器检查 `http://127.0.0.1:62080/models`：Video 分区 `.model-price-note` 数量为 0，首行只剩价格和 `按规格/秒计费`。
- 修正 `veo3.1-fast` 单位后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 修正 `veo3.1-fast` 单位后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 修正 `veo3.1-fast` 单位后，浏览器检查 `http://127.0.0.1:62080/models`：Veo 行显示 `default¥1.26/秒`、`extend¥0.56/秒`、`按规格/秒计费`。
- 修复 `videoTierUnit` 未使用参数后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 修复 `videoTierUnit` 未使用参数后，执行 `npm.cmd run typecheck`：通过。
- 修复 `videoTierUnit` 未使用参数后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 补充 Seedance 2.0 变体后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 补充 Seedance 2.0 变体后，执行 `npm.cmd run typecheck`：通过。
- 补充 Seedance 2.0 变体后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 补充 Seedance 2.0 变体后，浏览器检查 `http://127.0.0.1:62080/models`：可见 `doubao-seedance-2.0`、`doubao-seedance-2.0-fast`、`doubao-seedance-2.0-fast-face`、`doubao-seedance-2.0-face` 四行，均按 `/秒` 展示。
- 补充 Seedance 2.0 变体说明后，执行 `npm.cmd run test:run -- public-pages`：通过，9 个用例。
- 补充 Seedance 2.0 变体说明后，执行 `npm.cmd run typecheck`：通过。
- 补充 Seedance 2.0 变体说明后，执行 `git diff --check -- frontend/src/views/public/ModelPlazaView.vue frontend/src/__tests__/public-pages.spec.ts`：通过。
- 补充 Seedance 2.0 变体说明后，浏览器检查 `http://127.0.0.1:62080/models`：四个 Seedance 行均显示对应说明，行 `scrollWidth` 与可见宽度一致，未被内容横向撑开。

## 下一步

- 如需视觉验收：启动前端预览并访问 `/models`，检查视频模型 `官方/参考价` 逐行明细在桌面和窄屏下不溢出。
- 如需上线当前运行容器：重新 build 前端/后端并替换本地 `sub2api-dev` 容器。
