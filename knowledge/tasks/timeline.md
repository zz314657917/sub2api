# 项目时间轴

## 2026-05-27 01:56 +08:00 - 用户使用记录与新人引导收口复核

- 当前阶段：用户侧使用记录运维视图已提交推送，新人 API Key 引导弹窗完成后续文案与模型标签细化，并完成一次多智能体代码审查。
- 本段重点：`a5e998cb4` 完成用户侧 usage 表格、消费分组筛选、统计口径统一、列设置、详情抽屉和 CSV 字段扩展；`793c352ef` 到 `e73c60474` 连续收口新人福利弹窗，补余额/试用/奖励文案、头图、客服联系、能力标签和中英文 i18n。
- 已完成：当前 `main` / `origin/main` 已同步到 `e73c60474 fix(user): update onboarding model tags`；工作区干净；最近提交范围 `HEAD~1..HEAD` 只改 `UserApiKeyOnboardingDialog.vue`、对应 spec、`en.ts`、`zh.ts`。
- 关键决策：使用记录不新增数据库字段，继续复用 `usage_logs.group_id`、已 hydrate 的 `UsageLog.group` 和用户侧可用分组；新人奖励弹窗的能力标签从 `Claude Code` 调整为更宽的 `Claude`，并新增 `OpenClaw`。
- 验证记录：此前使用记录主线通过 `go test ./internal/handler -run UsageHandler`、`corepack.cmd pnpm exec vitest run src/views/user/__tests__/UsageView.spec.ts src/components/user/dashboard/__tests__/UserApiKeyOnboardingDialog.spec.ts src/components/common/__tests__/DateRangePicker.spec.ts`、`corepack.cmd pnpm run typecheck`、`git diff --check`；本次复核另执行 `git diff --check HEAD~1..HEAD` 和 `npm.cmd run test:run -- src/components/user/dashboard/__tests__/UserApiKeyOnboardingDialog.spec.ts` 通过。
- 遗留问题：默认 CI 的 `FRONTEND_CRITICAL_VITEST` 未包含 `UserApiKeyOnboardingDialog.spec.ts`；`toContain('Claude')` 断言可能被页面其他 `Claude Code` 文案误命中；未做真实浏览器视觉验收，4 个 pill 在窄屏弹窗内的换行仍需截图确认。
- 下一步：建议把 `UserApiKeyOnboardingDialog.spec.ts` 纳入默认关键前端测试，或明确 CI 有全量 Vitest；将 `Claude` 断言改为针对 pill 节点的精确断言；用真实登录态打开 dashboard/usage 做一轮桌面与窄屏视觉确认。

## 2026-05-21 20:10 +08:00 - 教程 CMS 与 UI/UX 评审修复完成

- 当前阶段：公共教程区已从硬编码 Vue 页面升级为后台可编辑教程 CMS，并完成一轮 UI/UX 评审后的公开页和后台管理体验收口。
- 本段重点：新增 `tutorial_pages` 后端实体、迁移、默认种子、公开 API 和管理员 API；前台 `/tutorial`、`/tutorial/:slug` 读取已发布教程，支持 Markdown + 短代码和旧 hash 跳转；后台新增 `/admin/tutorials` 管理入口。
- 已完成：公开页 API 失败/空列表不再静默 fallback，而是显示可重试提示并继续渲染内置教程；教程阅读宽度收敛、移动端文章页减轻总览和 sidebar；后台发布/下线增加确认和行级 busy，关闭编辑时拦截未保存修改，草稿隐藏“打开前台”，列表加载失败显示可重试错误态。
- 关键决策：教程 CMS 独立于 `custom_menu_items` 可见性；默认种子和前端 fallback 只作为升级兜底，不覆盖管理员编辑内容；图片上传资产库不在本轮范围，先支持 URL 或现有静态资源路径。
- 验证记录：`go test ./internal/handler ./internal/service ./internal/repository -run "Tutorial|Page|Settings" -count=1`、`npm.cmd run test:run -- public-pages`、`npm.cmd run test:run -- public-smoke`、`npm.cmd run test:run -- SettingsView`、`npm.cmd run build`、`git diff --check` 均通过；build 仍只有既有 Vite dynamic import/chunk size/`DEP0190` 警告；Playwright CLI 已截图抽查桌面、平板和移动教程页。
- 遗留问题：真实后端联调管理员登录态下的 `/admin/tutorials` 创建/发布流程还可做一次人工冒烟；后续若要上传教程图片，需要补资产库或复用现有上传通道；根目录两个未跟踪调试 PNG 仍未处理。
- 下一步：提交前复核 `frontend/src/views/admin/SubscriptionsView.vue` 既有改动不要混入本轮；如继续产品化，优先补教程图片上传、后台预览与前台样式更严格一致、以及真实管理员流程 E2E。

## 2026-05-21 14:36 +08:00 - 官方稳定修复选择性合并并推送

- 当前阶段：官方仓库差异已按“低风险稳定修复优先”完成一批 cherry-pick，避免整仓 merge 覆盖本地共享展示、容量池、公共页和生图工作台定制。
- 本段重点：合入 OpenAI 图片错误透传与 `n` 参数、账号删除调度缓存清理、报错账号取消调度、refresh token reused 不重试、thinking block 重试、Ops SLA 分类与本地调度容量标记。
- 已完成：临时分支 `codex/pick-upstream-stability` 合回 `main`，生成并推送 `7b0b05c4b merge: pick upstream stability fixes` 到 `origin/main`；远端与本地 `main` 已对齐。
- 关键决策：跳过 `d3d5843b9 fix(channel-monitor): 兼容 Responses reasoning 输出`，因为它依赖未合并的 channel monitor API mode / 模板协议管理大功能；fetch 后新增的 `a613a587b feat: add subscription expiry email toggle` 属于邮件通知配置、migration 和前端设置页功能线，本轮暂不混入稳定修复批次。
- 验证记录：临时分支和合回 `main` 后均执行 `go test ./internal/handler -run "Ops|OpenAIImages|ErrorLogger|Gateway|Gemini|ChatCompletions|Responses" -count=1`、`go test ./internal/service -run "OpenAIImages|Image|TokenRefresh|Thinking|GatewayRequest|GatewayService|Ops|ChannelMonitor" -count=1`、`go test ./internal/repository -run "Account|Scheduler|Delete|Unschedule" -count=1`、`go test ./internal/server/middleware -run "APIKeyAuth|IPRestriction" -count=1` 通过；`git diff --check` 通过；`git push origin main` 成功。
- 遗留问题：未跑后端 `go test ./...` 全量，也未跑前端 typecheck/build；根目录 `capsule_area_preview.png`、`capsule_mask_debug.png` 仍是未跟踪调试图片；官方邮件通知开关、channel monitor 大功能、兑换码增强、钉钉 OAuth、ACL 真实 IP 开关、Bedrock Claude Code 兼容、用量平台拆分等仍未合入。
- 下一步：如继续合官方，按主题单独评估 `a613a587b` 邮件通知开关或 channel monitor API mode，不建议直接整仓合并。

## 2026-05-21 11:30 +08:00 - 排行榜趋势与共享容量池配置提交前收口

- 当前阶段：近期共享展示、容量池和排行榜主线进入一轮合并提交前自检与验证。
- 本段重点：排行榜接口新增近 10 日 Token 趋势、输入/输出 Token 和每百万 Token 成本；前端排行榜新增趋势图、Token 指标 tooltip 和每日奖励进度；共享展示卡片改为“移动到容量池”语义，并同步管理端/用户侧账号数保存。
- 已完成：补后端 DTO/service/repository/handler 测试；补前端 `LeaderboardView`、`EditAccountModal`、`MyAccountsView`、公共模型广场断言；模型广场在已登录但渠道为空时使用 fallback catalog 且隐藏倍率组。
- 关键决策：根目录 `capsule_area_preview.png`、`capsule_mask_debug.png` 视为调试图片，不进入提交；排行榜趋势暂复用现有全局趋势查询，用户时区只用于窗口计算，严格用户时区分桶留作后续独立改造。
- 验证记录：执行 `go test ./internal/handler ./internal/repository ./internal/service`、`go test ./...`、`npm.cmd run test:run -- LeaderboardView EditAccountModal public-pages leaderboard-theme`、`npm.cmd run test:run -- MyAccountsView.importFile EditAccountModal LeaderboardView public-pages leaderboard-theme`、`npm.cmd run typecheck`、`npm.cmd run build`、`git diff --check` 均通过；build 仅有既有 Vite 动态导入和大 chunk 警告。
- 遗留问题：未做真实浏览器视觉检查；趋势分桶严格用户时区语义未实现；两个根目录调试 PNG 后续应删除或移动到正式资产目录。
- 下一步：提交本轮 19 个源码/测试文件；后续用真实数据打开 `/leaderboard` 和共享池页面做桌面/移动端视觉复核。

## 2026-05-18 00:20 +08:00 - 知识入口从聊天生图转向共享展示与容量池主线

- 当前阶段：知识入口收口到最近两天真正高频的产品主线，而不是继续停在双仓库聊天生图闭环。
- 本段重点：补 `knowledge/05-current-focus.md`，并把 `current-task`、`build-and-verify` 从旧 `/chat-images` / COS 语境推进到账号共享展示、容量池聚合、排行榜文案和 cockpit 导入。
- 已完成：新增 `knowledge/05-current-focus.md`；更新 `knowledge/build-and-verify.md` 和 `knowledge/tasks/current-task.md`，明确容量池验证入口、fake 演示账号边界和双仓库知识分工。
- 关键决策：聊天生图与 COS 仍是稳定历史结论，但当前默认主线应改为“共享产品化 + 容量池展示收口”，避免后续续做继续误判优先级。
- 验证记录：本轮只更新知识文件，依据最近提交主题和既有时间轴事实收口；未新增代码验证命令。
- 遗留问题：如果共享账号、排行榜或 cockpit 继续独立演进，后续可能需要再拆更细的专题知识页，而不是持续堆在任务快照和时间轴。
- 下一步：如继续补知识，优先把共享展示、容量池和用户侧文案的长期结论沉淀到稳定专题页。

## 2026-05-17 08:16 +08:00 - /monitor 容量池按套餐展示与提交归档

- 当前阶段：`/monitor` 渠道状态里的账号容量池展示已完成一轮收敛、验证、提交并推送到 `origin/main`。
- 本段重点：OpenAI 共享容量池按套餐聚合展示；百分比文案改成显示剩余、隐藏 30D；没账号的池子/分组隐藏；本地 fake 演示账号用于页面预览且被排除出生产调度查询。
- 已完成：提交并推送 `5dfba2e3 fix: simplify capacity pool window display`、`1218bdcc fix: group shared openai pools by plan`、`867addcf fix: refine capacity pool monitor display`、`6efa3849 fix: exclude demo fake accounts from scheduling`、`7c63ea4f chore: clarify leaderboard badge labels`；`main` 已与 `origin/main` 对齐。
- 关键决策：OpenAI Plus/Pro/Team/Free 组优先按套餐名聚合，而不是按用户自定义 display name 拆成 `1`、`8000` 等池子；套餐组只展示 `5h` 和 `7d` 窗口；percent-only 场景统一展示剩余百分比；空账号池不展示。
- 验证记录：执行过 `npm.cmd run test:run -- ChannelStatusView.capacityPools`、`npm.cmd run typecheck`、`npm.cmd run build`、`go test ./internal/repository -run '^$'`、`git diff --check`；本地接口曾确认容量池返回 OpenAI Free/Plus/Pro/Team 均为 `healthy`，且 Pro 已合并为单一套餐组。
- 遗留问题：本地 Docker 镜像仍可能是旧构建，容器重建会丢失手动拷贝的后端二进制；fake 演示账号数据存在本地数据库，不是迁移或种子数据；`notDemoFakeAccountPredicate` 目前依赖 `extra.demo_fake_account=true` 这个 JSON 标记。
- 下一步：如继续做容量池，优先补一个正式 demo/seed 开关或管理入口，避免靠手工 SQL 造假账号；如要重新部署本地后端，优先用正式镜像构建链路替代手动 `docker cp`。

## 2026-05-16 20:22 +08:00 - 双仓库提交部署与 COS 图片链路归档

- 当前阶段：Sub2API + chatgpt2api 双服务已完成自定义仓库提交、远程部署与 COS 对象存储接入说明。
- 本段重点：`sub2api` 和 `chatgpt2api` 改动已分别推送到用户仓库；chatgpt2api 独立目录迁移到 `F:/java/chatgpt2api`；远程部署采用两个自定义镜像和同一 redeem secret 串联。
- 已完成：`sub2api` 推送 `8489b012 feat: expand account sharing and image workspace`；`chatgpt2api` 推送 `e753ae6 feat: integrate sub2api image workspace`；说明了部署 env、Caddy 反代、COS 存储流程、下载走 COS/CDN、删除同步删远端对象，以及 Referer 防盗链建议。
- 关键决策：两仓库分开维护和部署，不把 chatgpt2api 放在 sub2api `tmp/`；主图展示/下载直接走 COS/CDN，不经站点中转；用防盗链、登录态和每用户 50 张限制控制流量风险。
- 验证记录：chatgpt2api 在 `F:/java/chatgpt2api` 执行 `go test ./...`、`corepack.cmd pnpm --dir web lint`、`corepack.cmd pnpm --dir web build`、`git diff --check` 通过；sub2api 执行后端 `go test ./...`、前端 `npm.cmd run typecheck`、相关 vitest 41 个测试、`npm.cmd run build`、`git diff --check` 通过；两个仓库本地/远端 HEAD 已核对一致。
- 遗留问题：远程真实生图后的 COS 对象可访问、防盗链白名单、下载按钮在不同浏览器下的 Referer 行为仍需线上实测；缩略图仍由站点本地提供，未迁移到 COS。
- 下一步：线上用真实登录态从 Sub2API “聊天生图”入口跳转到 chatgpt2api，测试 1 张和 8 张生图、历史恢复、COS 访问、防盗链拦直链，以及清理超过 50 张时远端对象删除。

## 2026-05-16 16:07 +08:00 - chatgpt2api 接入腾讯 COS/S3 兼容图片存储

- 当前阶段：Sub2API -> chatgpt2api 独立生图工作台开始补齐生产部署图片存储能力。
- 本段重点：新增 `internal/imagestore` S3 兼容对象存储适配，生成图片本地落盘后上传 COS，返回给前端和图库的 URL 优先使用对象存储/CDN 地址，同时保留本地路径用于元数据、缩略图和每用户 50 张治理。
- 已完成：新增 COS/S3 env 配置示例；生成结果新增内部 `local_url` 用于图库记录，用户可见 `url` 指向 COS；元数据保存 `storage_backend`、`object_key`、`object_url`；图片删除、过期清理、容量清理和每用户上限清理会同步删除远端对象。
- 关键决策：不再让浏览器依赖服务器本地 `/images/...` 作为主展示地址；但本地文件和元数据仍保留，避免缩略图、历史列表、复用参数、50 张上限失效。
- 验证记录：`go test ./...` 通过；`corepack.cmd pnpm --dir web lint` 通过；`corepack.cmd pnpm --dir web build` 通过，仅既有 Vite 大 chunk warning；`git diff --check` 通过，仅 CRLF warning；chatgpt2api 8081 已重启，`/health` 正常。
- 遗留问题：当前本地未配置真实腾讯云 COS 密钥，只用 httptest S3 兼容服务覆盖上传/删除；部署后需配置真实 COS bucket/CDN/public read，再做 1 张和 8 张真实生图确认对象 URL 可访问。

## 2026-05-16 18:05 +08:00 - chatgpt2api 每用户图片上限默认 50

- 当前阶段：在 Sub2API -> chatgpt2api 独立生图工作台链路上补齐图片存储治理。
- 本段重点：chatgpt2api 原本只按保留天数和全局容量清理，没有每用户张数限制；本轮新增 `CHATGPT2API_IMAGE_MAX_SAVED_PER_USER`，默认 50。
- 已完成：后端配置、设置页字段、自动清理入口和手动治理 API 均接入每用户上限；清理按 owner 保留最新图片，超出后删除最旧图片及其缩略图、参考图和元数据；公开图也计入该用户 50 张硬限制。
- 关键决策：50G 腾讯云 COS 容量下，默认每用户 50 张更适合早期控制增长；值设为 0 可关闭按用户张数清理。
- 验证记录：`go test ./internal/config ./internal/service ./internal/httpapi` 通过；`corepack.cmd pnpm --dir web lint` 通过；`corepack.cmd pnpm --dir web build` 通过，仅有既有 Vite 大 chunk 提示；`git diff --check` 通过，仅 CRLF warning。
- 遗留问题：chatgpt2api 仍未实现 COS/S3 对象存储后端，当前图片仍保存在本地 `data/images`；后续接腾讯云 COS 需另做对象存储适配或复用 Sub2API 存储层。

## 2026-05-16 14:49 +08:00 - ZyphrZero chatgpt2api 本地容器/8 张生图闭环

- 当前阶段：Sub2API 跳转到 ZyphrZero chatgpt2api 独立生图工作台的后端闭环已跑通，重点转向部署镜像和真实浏览器入口确认。
- 本段重点：检查现有容器后确认只有 `sub2api-dev`、Postgres、Redis 可复用，没有包含本地改动的 chatgpt2api 容器；本地以 `go run -tags=embed ./internal` 启动 chatgpt2api 8081 完成测试。
- 已完成：修复 8 张生图链路，任务层上限从 4 放宽到 8；Sub2API 桥接层改为按输出槽并发发起 8 个 `n=1` 请求并按 index 合并，避免当前 gateway 对 `n>1` 实际只返回 1 张。
- 关键决策：不依赖上游兼容接口的多图 `n` 行为；独立工作台负责 8 槽并发与结果合并，Sub2API 继续负责 API Key、余额和扣费。
- 验证记录：`go test ./internal/service ./internal/httpapi ./internal/protocol` 通过；`git diff --check` 通过且仅有 CRLF warning；真实 1 张生图成功；真实 8 张任务 `codex-parallel-8-*` 最终 `outputs=8`、8 个槽全部 `success`。
- 遗留问题：尚未用真实浏览器从 Sub2API `/chat-images` 人工点击进入 chatgpt2api `/image`；生产部署需构建包含当前本地改动的自有 chatgpt2api 镜像并配置数据目录持久化。
- 下一步：用浏览器登录态验证 `/chat-images` 按钮跳转和 `/image` 页面交互；构建/部署自有 chatgpt2api 镜像；部署后再做 1 张和 8 张生图、历史恢复和用量归属检查。

# 2026-05-21 19:46 +08:00 - 教程页后台可编辑 CMS 落地

- 当前阶段：把公共教程页从硬编码 Vue 模板升级为后台可维护的教程 CMS。
- 本段重点：管理员可在 `/admin/tutorials` 管理教程标题、slug、分组、排序、草稿/发布状态和 Markdown 正文；公共 `/tutorial` 和 `/tutorial/:slug` 自动读取已发布内容。
- 已完成：新增 `tutorial_pages` 迁移和默认种子；新增教程 service/repository/handler/routes/wire；新增公开 API 和管理员 API；公共教程页改为数据驱动并保留 Matrix 风格；新增 Markdown + 短代码渲染工具；新增后台教程管理页；支持旧 hash 跳转到 slug；后台无发布内容时使用内置 fallback。
- 关键决策：教程 CMS 独立于 `custom_menu_items`，不复用自定义菜单页可见性；正文格式采用 Markdown + 最小短代码，不做完整可视化块编辑器；种子迁移使用 `ON CONFLICT (slug) DO NOTHING`，避免覆盖管理员编辑内容；`/tutorial` 保留为目录入口，单篇正文走 `/tutorial/:slug`。
- 验证记录：`go test ./internal/handler ./internal/service ./internal/repository -run 'Tutorial|Page|Settings' -count=1`、`npm.cmd run test:run -- public-pages`、`npm.cmd run test:run -- public-smoke`、`npm.cmd run test:run -- SettingsView`、`npm.cmd run build`、`git diff --check` 均通过；build 仍有既有 Vite dynamic import/chunk size/`DEP0190` 警告。
- 遗留问题：未做浏览器截图验收；图片上传资产库不在本轮范围；本地已有 `frontend/src/views/admin/SubscriptionsView.vue` 未提交改动和两个根目录调试 PNG，提交时需与教程 CMS 变更区分。

## 2026-05-16 13:53 +08:00 - ZyphrZero chatgpt2api 接入 Sub2API 统一生图扣费

- 当前阶段：从旧 gpt2api 方案切到与参考站 `/image` 匹配的 `ZyphrZero/chatgpt2api`，把它作为独立生图工作台。
- 本段重点：用户仍从 Sub2API 聊天生图入口进入；chatgpt2api 用 Sub2API launch token 建立本地会话，并用用户自己的 API Key 调 Sub2API gateway 生图。
- 已完成：新增 chatgpt2api 侧 `sub2api` 外部会话、launch/redeem、绑定存储、`/auth/sub2api/launch` 路由、前端回调页；图片生成/编辑任务在 `sub2api` 会话下绕过本地积分并转发到 Sub2API gateway；`auto`/`codex-gpt-image-2` 归一为 `gpt-image-2`；网关调用改为直连 HTTP client；gateway override 环境变量优先于 redeem 返回值。
- 关键决策：不让用户在 chatgpt2api 再注册，不做两套积分；Sub2API 负责账号、API Key、余额与扣费，chatgpt2api 负责更完整的生图 UI 和任务展示。
- 验证记录：在 `tmp/chatgpt2api-zyphr` 执行 `GOPROXY=https://goproxy.cn,direct go test ./...`、`corepack.cmd pnpm --dir web build`、`corepack.cmd pnpm --dir web lint`、`git diff --check` 均通过；`git diff --check` 仅有 Windows CRLF 提示。
- 遗留问题：尚未用真实浏览器从 Sub2API `/chat-images` 点击进入新工作台；尚未用真实 API Key 在新接入版本做 1 张和 8 张生图；生产部署前需保护 chatgpt2api 数据目录/数据库中的 API Key 绑定。
- 下一步：配置 chatgpt2api 的 `CHATGPT2API_SUB2API_REDEEM_URL`、`CHATGPT2API_SUB2API_REDEEM_SECRET` 和可选 gateway override；配置 Sub2API launch 目标到 chatgpt2api；再做真实 UI 跳转、1 张生图和 8 张生图闭环。

## 2026-05-16 12:18 +08:00 - gpt2api 生图桥接后端闭环验证

- 当前阶段：sub2api 聊天生图入口已切到 gpt2api `/image`，开始真实生图链路验证。
- 本段重点：使用本地用户 16 的 active API Key 经 launch/redeem 自动登录 gpt2api，再调用 gpt2api `/api/v1/gen/image` 生成 `gpt-image-2`。
- 已完成：确认 sub2api/gpt2api 四个本地服务在线；第一轮任务 `d4d9031b9a2747d4a6bc9c8aa3` 成功创建但上游 account 1 返回 524，最终 gpt2api 状态 3；第二轮短提示任务 `8c1334ef9664431d856a82c2d9` 成功，结果 1 张。
- 关键决策：失败任务不视为桥接失败，日志显示 gpt2api 已正确调用 sub2api `/v1/images/generations`，失败点是上游 OpenAI 图片请求超时；后续需要区分“上游偶发 524”和“页面展示问题”。
- 验证记录：成功任务状态 2、进度 100、结果 1 张；缓存图片 `/api/v1/gen/cached/generated/2026/05/16/8c1334ef9664431d856a82c2d9_0.png` HTTP 200，大小约 1.25 MB；sub2api 日志对应 `/v1/images/generations` 状态 200。
- 遗留问题：尚未通过真实浏览器 UI 从 `/chat-images` 点按钮进入 gpt2api `/image` 并看图；尚未做 8 张生图验证；当前账号池只有 account 1 在分组 2 可用，偶发 524 时会导致无账号可切。
- 下一步：人工刷新浏览器，从 sub2api `/chat-images` 进入 gpt2api `/image`，确认结果展示/历史；再发 8 张任务观察成功率和部分成功展示。

## 2026-05-16 12:04 +08:00 - gpt2api 跳转落点对齐参考站 /image

- 当前阶段：sub2api 聊天生图入口已能跳转 gpt2api，继续修正跳转后页面不一致问题。
- 本段重点：参考站是 `https://img.imgwwo.top/image`，而本地此前登录成功后进入 `/create/image` 的新版 Studio 页，导致视觉和交互不一致。
- 已完成：gpt2api 用户前端新增 `/image` 路由挂载旧 `CreateImagePage`；首页、图片导航、登录/注册默认落点、Sub2API launch 成功落点均改为 `/image`；旧 `/create/image` 仍保留为备用 Studio 页。
- 关键决策：只改落点与导航，不删除 `/create/image`；同时把 `CreateImagePage` 的模型列表改为 `gpt-image-2`，避免页面对了但请求旧 `img-v3/img-real` 模型。
- 验证记录：gpt2api `@kleinai/user typecheck`、`build` 通过；`http://127.0.0.1:8081/image` 返回 200；相关文件 `git diff --check` 通过。
- 遗留问题：尚未用真实登录态从 sub2api `/chat-images` 点击到 gpt2api `/image` 做人工视觉确认；尚未做真实 1 张/8 张生图展示验证。
- 下一步：刷新/重启 gpt2api 前端后，从 sub2api `/chat-images` 进入，确认最终 URL 为 `/image` 且页面视觉对齐参考站；随后测 `gpt-image-2` 1 张和 8 张。

## 2026-05-16 11:47 +08:00 - 聊天生图入口切到独立生图工作台

- 当前阶段：sub2api -> gpt2api 跳转生图链路继续收口到正确用户入口。
- 本段重点：`/chat-images` 现在就是新版生图工作台启动页，用户从“聊天生图”选择 API Key 后跳转 gpt2api；旧 `/open-webui/launch` 只做兼容重定向。
- 已完成：更新前端路由、中文/英文文案和 smoke 断言；旧原生聊天生图保留在 `/chat-images/native` 作为备用入口。
- 关键决策：保留内部 `openWebUI` API/组件命名，避免牵动后端接口和部署 env；用户可见文案统一改为“生图工作台”。
- 验证记录：`npm.cmd run test:run -- public-smoke AppSidebar OpenWebUILaunchView` 通过，3 个测试文件 18 个断言；`npm.cmd run typecheck` 通过；`http://127.0.0.1:62080/chat-images` 返回 200；相关文件 `git diff --check` 通过。
- 遗留问题：尚未用真实登录态从 `/chat-images` 人工点击按钮完成“选择 API Key -> gpt2api `/create/image`”闭环；尚未做真实 `gpt-image-2` 1 张和 8 张生图展示验证。
- 下一步：用真实浏览器登录态打开 `/chat-images` 点击进入工作台；随后在 gpt2api 发起 1 张和 8 张生图，确认图片展示、历史恢复和用量归属。

## 2026-05-16 11:22 +08:00 - gpt2api 跳转登录浏览器闭环验证

- 当前阶段：sub2api -> gpt2api 一键登录和 API Key 绑定已进入本地闭环验证。
- 本段重点：真实后端入口 `/api/v1/user/open-webui/launch` 能生成 gpt2api 回调 URL；gpt2api 能兑换 token、签发登录态、写入前端 `klein:token` 并进入 `/create/image`。
- 已完成：发现并修复 gpt2api 回调页在 React dev `StrictMode` 下 effect 二次执行导致一次性 token 被重复兑换、第二次 401 的问题；同一 token 现在复用同一个兑换 Promise。
- 关键决策：不取消 React StrictMode，也不让 sub2api token 可重复兑换；前端回调页只做同 token 幂等去重，保留一次性 token 的安全语义。
- 验证记录：直连 sub2api redeem 成功；真实 launch 接口生成 URL 后由浏览器打开，最终 URL 为 `http://127.0.0.1:8081/create/image`，且 `klein:token` 已写入；gpt2api `@kleinai/user lint/typecheck/build` 均通过；`git diff --check` 仅 CRLF warning。
- 遗留问题：尚未用人工真实登录态点击 sub2api 页面按钮；尚未消耗真实 API Key 做 `gpt-image-2` 1 张/8 张生图展示验证。
- 下一步：在 `http://127.0.0.1:62080/open-webui/launch` 用真实登录态点击选择 API Key 并跳转；随后在 gpt2api `/create/image` 做 1 张和 8 张生图闭环。

## 2026-05-16 10:10 +08:00 - gpt2api 接入 Sub2API 跳转登录与生图桥接

- 当前阶段：从“只修 sub2api 原生生图”扩展到“部署/改造 gpt2api 作为独立生图前台，并复用 sub2api 用户 API Key”。
- 本段重点：gpt2api 已新增 Sub2API launch token 自动登录、影子用户绑定、API Key 加密保存、`gpt-image-2` 生图桥接和 8 张上限；本地服务链路为 sub2api `62080/8080` + gpt2api `8081/17180`。
- 已完成：新增 gpt2api `sub2api_user_binding` 迁移、launch service/repo/model、前端回调页、登录/注册跳转 Sub2API、OpenAI-compatible image count 8、Sub2API gateway 生图桥接；自检后补齐桥接成功结算、登录/注册页 Hook 顺序、ESLint 9 flat config。
- 关键决策：gpt2api 侧不再做独立注册主流程，用户从 sub2api 跳转后使用自己的 API Key；当前默认不需要积分，但桥接成功仍支持未来收费模型的结算。
- 验证记录：gpt2api `go test ./internal/router ./internal/handler ./internal/service ./pkg/config`、`@kleinai/user lint/typecheck/build` 均通过；本地 `17180/readyz` ready true，`8081/api/v1/auths/sub2api/launch?token=abc` 307 到 `/auth/sub2api/launch?token=abc`，`62080/open-webui/launch` 返回 200。
- 遗留问题：尚未用真实浏览器登录态完成一次 launch token 兑换和真实生图；生产部署前需确认反代路径、migration 和环境变量配置。
- 下一步：用真实 sub2api 登录态打开 `http://127.0.0.1:62080/open-webui/launch`，验证 gpt2api 自动登录；再测 `gpt-image-2` 1 张和 8 张生成，确认图片展示和用量归属。

## 2026-05-15 23:35 +08:00 - 聊天生图会话删除语义沉淀

- 当前阶段：原生 `/chat-images` 工作台首版后，补充会话与后端图片任务的知识库事实。
- 本段重点：确认“生成图片时删除会话”的真实项目行为，避免套用 ChatGPT 官方产品语义。
- 已完成：新增 `knowledge/chat-image-studio.md`；更新 `knowledge/00-start-here.md`、`knowledge/frontend-notes.md`、`knowledge/backend-notes.md` 和 `knowledge/tasks/current-task.md`。
- 关键决策：当前记录为已验证实现事实，不改运行时代码；现语义是删除本地聊天会话不等于取消后端图片任务。
- 验证记录：代码检查 `ChatImageStudioView.vue`、`frontend/src/api/imageCreator.ts`、`routes/user.go`、`image_creator_handler.go`、`image_creator_service.go`、`image_creator_repo.go`；未跑测试，因为本轮只更新知识库。
- 遗留问题：若产品希望“删会话即取消生图”，后续需新增后端取消接口/状态、worker 取消检查和前端确认/轮询清理。
- 下一步：继续做 `/chat-images` 真实登录态 UI、生图闭环和旧入口策略验证。

## 2026-05-14 02:57 +08:00 - 教程页转为文档三栏并强化 CC Switch

- 当前阶段：公共教程页从卡片式快速开始继续转为更像文档站的阅读体验。
- 本段重点：参考 EasyRouter CC Switch 文档的信息架构，改为左侧工具目录、中间正文、右侧本页目录；CC Switch 教程改成控制台一键导入 Provider 优先。
- 已完成：更新 `frontend/src/views/public/TutorialView.vue`，新增 `docAnchors` 右侧锚点目录、浅色文档样式、CC Switch Provider/MCP/Prompts/多平台说明、`ccswitch://` Deep Link 接入步骤、macOS/Linux/Web 安装方式和手动兜底；更新 `frontend/src/__tests__/public-pages.spec.ts` 覆盖新结构。
- 关键决策：不复制 EasyRouter 品牌和长文案，只学习文档型排版与 CC Switch 信息层级；落叶网络的固定接口地址仍为 `https://ai.3zapi.top`。
- 验证记录：`npm.cmd run test:run -- public-pages` 通过；`npm.cmd run typecheck` 通过；`npm.cmd run build` 通过但仍有既有 Vite chunk / `DEP0190` 警告；`git diff --check` 通过；Playwright CLI 已验证桌面、移动端和 CC Switch 段截图，保存到 `output/playwright/tutorial-doc-layout-*.png`。
- 遗留问题：后续如继续优化，可把教程数据抽成数组驱动，降低大模板维护成本。

## 2026-05-14 02:52 +08:00 - shadcn 深色二阶段收敛与控制台灰雾修正

- 当前阶段：在已有首页、教程页、模型广场和控制台布局上继续收敛视觉系统，不重写业务页面。
- 本段重点：新增/复用 public 与 console surface token，统一深色表面、细边框、8px 圆角、低阴影、hover/focus；修正 driver.js 自动引导遮罩导致控制台视觉发灰的风险。
- 已完成：更新 `public-page.css`、`HomeView.vue`、`PublicTopNav.vue`、`ModelPlazaView.vue`、`TutorialView.vue`、`console-ui.css`、`useOnboardingTour.ts`、`onboarding.css`；公共 CTA、导航、模型筛选/卡片、教程 panel/card/sidebar 和控制台 header/sidebar/card/table/input/button/dropdown 均收敛到 token 化样式。
- 关键决策：不引入 `shadcn-vue`，不改路由/API/权限；保留已有 `--mc-*`、PixelIcon 节点和测试入口；控制台发灰优先处理引导层透明度与 dark surface，而不是关闭 onboarding 功能。
- 验证记录：公共页与控制台样式测试共 32 个断言通过；`npm.cmd run build` 通过；`git diff --check` 通过；Playwright CLI 已验证 `/home`、`/models`、`/tutorial#quick-start` 并保存截图到 `output/playwright/62080-*-after.png`。控制台登录后页面因当前 Playwright 会话未持有登录态未实测，需要用户浏览器刷新后复核。
- 遗留问题：提交前仍需复核当前未提交文件是否拆分为“公共页视觉收敛”和“控制台/onboarding 修正”；控制台 `/dashboard`、`/admin/dashboard`、`/admin/settings` 建议用已有登录态做最终人工视觉确认。

## 2026-05-14 02:29 +08:00 - 教程页补充 Linux/macOS 与左上目录

- 当前阶段：公共教程页继续从单一 Windows 口径扩展为跨平台接入文档。
- 本段重点：新增 Linux / macOS 环境配置段；代理地址统一为无尾斜杠 `https://ai.3zapi.top`；桌面端教程目录改为左上 sticky 导航并与总览内容同排展示，移动端保留横向滑动目录。
- 已完成：更新 `frontend/src/views/public/TutorialView.vue`，新增 Linux 基础环境、Linux Shell 配置、macOS 基础环境、macOS Shell 配置；Windows 手动配置文案明确为 Windows 专属；更新 `frontend/src/__tests__/public-pages.spec.ts` 覆盖新内容和布局。
- 关键决策：教程内容增多后，桌面端采用左侧目录降低回看成本；最新布局把目录提前到页面最外层左上，并将主容器放宽到 `max-width: min(118rem, calc(100vw - 1rem))`，减少宽屏两侧空白；移动端继续使用横向 chips，避免挤占内容宽度。
- 验证记录：`npm.cmd run typecheck` 通过；`npm.cmd run test:run -- public-pages` 通过，7 个测试；`git diff --check` 通过；`npm.cmd run build` 通过但仍有既有 Vite chunk / `DEP0190` 警告；Playwright CLI 已验证 `http://127.0.0.1:62080/tutorial#platforms`，截图为 `output/playwright/tutorial-platforms-desktop.png`、`output/playwright/tutorial-platforms-mobile.png` 和 `output/playwright/tutorial-left-sidebar-wide-expanded-full.png`。
- 遗留问题：本轮只改教程页和公共页测试；如后续继续扩展教程，可考虑把 Windows / Linux / macOS 的工具安装链接抽成更易维护的数据结构。

## 2026-05-13 13:01 +08:00 - 建立项目知识库入口

- 当前阶段：仓库开始把 AI 协作知识从零散 `docs/ai` 和任务快照收束到 `knowledge/`。
- 本段重点：新增长期知识入口、项目地图、构建验证、前后端笔记、已知坑点和任务状态说明。
- 已完成：建立 `knowledge/00-start-here.md`、`project-map.md`、`build-and-verify.md`、`backend-notes.md`、`frontend-notes.md`、`known-pitfalls.md`、`task-state.md`；旧 `docs/ai/00-start-here.md` 改为指向新知识库。
- 关键决策：`knowledge/` 根下放可提交长期知识；`knowledge/tasks/` 继续只放本地动态快照和时间轴，因为当前 `.gitignore` 忽略该目录。
- 验证记录：已 UTF-8 回读关键文档；`git check-ignore` 确认 `knowledge/` 根下知识页未被忽略，`knowledge/tasks/` 和 `docs/ai/` 仍被现有规则忽略；敏感词/乱码扫描无异常命中。
- 遗留问题：如果希望 `current-task.md`、`timeline.md` 或 `docs/ai/00-start-here.md` 入库，需要后续调整 `.gitignore`。
- 下一步：新会话优先从 `knowledge/00-start-here.md` 进入；开发任务再按后端、前端、验证或任务快照分流。

## 2026-05-13 01:15 +08:00 - 公共页面深色统一与教程页重做

- 当前阶段：公共首页体系继续打磨，重点落在教程页、模型广场和公共导航一致性。
- 本段重点：模型广场从浅色页统一到深色 Matrix 公共页面风格；教程页重写为可跟读的接入文档；左侧目录改为 sticky 并随阅读位置高亮。
- 已完成：新增/更新 `public-pages` 静态断言；教程页 8 步接入内容覆盖准备信息、注册试用、API 密钥、Base URL、Codex、SDK、兼容工具和用量排查；Playwright 截图保存到 `output/playwright/tutorial-sticky.png`。
- 关键决策：参考外部教程只取结构和信息组织方式，不复制品牌文案；公共页面继续使用同一套 `PublicTopNav` 与 `PublicMatrixBackdrop`，避免教程深色、模型广场浅色的割裂。
- 验证记录：`npm.cmd run test:run -- public-pages` 通过；`npm.cmd run build` 通过但仍有既有 Vite chunk/DEP0190 警告；`git diff --check` 通过；Playwright CLI 验证教程目录滚动后仍在视口内。
- 遗留问题：本地仍有 6 个未提交文件，提交前需要复核是否拆分公共页面改动与控制台侧栏/i18n 改动；模型广场还建议人工看一次真实数据下的筛选和卡片拥挤度。
- 下一步：复核 `/tutorial#quick-start`、`/models` 和相关控制台侧栏效果；确认后按主题拆分或合并提交并推送。
