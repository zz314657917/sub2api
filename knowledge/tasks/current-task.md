# 当前任务快照

最后更新：2026-07-08 01:22 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前任务：执行“Sub2API 全前端统一暖白 / 陶土 / 黑灰风格 + 首页内嵌登录注册多智能体计划”。
- 目标：保留当前 MiMo/Xiaomi 首页方向和鼠标 reveal 效果，在首页右侧内嵌注册/登录卡片；同时把公共页、认证页、基础控制台 token、setup/key usage/404 收口到暖白、陶土、黑灰极简风格。
- 多智能体：主 Codex 执行集成与验证；子智能体 `019f3b25-8bc8-7d61-9c03-c7f49a755f37` 只读复核 `/models`、`/setup` 和旧色残留，未修改文件。
- 明确边界：未 stage、未提交、未回滚 unrelated dirty files；仍需提交时必须 scoped staging，不能 `git add .`。

## 本轮已完成

- 首页与认证：
  - 新增 `frontend/src/components/auth/AuthAccessPanel.vue`，复用登录/注册表单。
  - `/home` 默认首屏改为左侧产品文案 + 右侧认证卡片；未登录默认展示注册，可切换登录；已登录展示仪表盘入口。
  - `/login`、`/register` 继续保留路由，但加载 `HomeView.vue` 并传 `authMode`，不再走旧独立暗色页；`redirect`、OAuth、邀请码、优惠码、Turnstile、2FA、协议弹窗等业务逻辑仍由原表单承接。
  - `homeContent` 自定义首页只在 `route.path === '/home' && !!homeContent.value` 时生效，避免吞掉 `/login`、`/register`。
  - 认证卡片顶部补动态标题/副标题；登录和注册的快捷登录区域移动到邮箱表单前，GitHub/Google/OIDC/LinuxDo 仍按后台开关显示；底部补“继续即表示同意”条款提示，并只挑核心条款/政策链接避免卡片底部过长。
  - 注册表单输入框、提示、校验态和提交按钮改用与登录一致的 `auth-input` / `auth-submit-button` 体系；桌面端认证卡片设置稳定最小高度，默认登录/注册切换不再跳高，同时邀请码/优惠码/Turnstile 开启时仍可继续向下扩展。
  - 首页内嵌登录/注册切换时禁用原生 `autofocus`，避免表单 remount 后浏览器自动聚焦邮箱输入框并带动页面滚动；独立 `/login`、`/register` 仍保留自动聚焦。
- 认证页壳：
  - `LoginView.vue`、`RegisterView.vue` 支持 `embedded` / `showTitle`。
  - `AuthLayout.vue` 改为暖白浅色认证壳；忘记密码、重置密码、邮箱验证和 OAuth callback 继续复用。
- 公共页：
  - `PublicTopNav.vue` 为 fixed 毛玻璃顶栏，覆盖首页、教程、模型广场和法律页；右上角未登录态移除“注册领取试用”和“登录”入口，已登录态保留“前往仪表盘”，并保留联系客服。
  - `/models`、`/tutorial`、法律页保持 `PublicRevealBackdrop` 和暖白公共页壳。
  - `ModelPlazaView.vue` 增加 `normalizeCatalog()`，异常接口/错误代理返回非标准结构时进入错误态，不再因 `groups` 缺失崩溃。
- 全局/控制台基础风格：
  - `style.css` 全局 token 改为暖白/陶土/黑灰，按钮、输入、卡片、选择文本等不再使用旧绿主色。
  - `tailwind.config.js` 的 `primary-*` palette 从旧绿色改为陶土，减少用户/后台旧绿控件残留。
  - `console-ui.css` 已收口到暖白浅色控制台壳，dark mode 保留 slate；成功/危险/支付品牌按钮保留语义色。
  - `KeyUsageView.vue`、`SetupWizardView.vue`、`NotFoundView.vue` 已浅色化。
  - 2026-07-07 19:44 继续补齐控制台/用户侧旧色残留：`onboarding.css`、管理员/用户账号测试弹窗、管理员/用户用量 tooltip、管理员退款弹窗、用户订单退款/发票入口、订阅进度小组件、全局 `badge-purple` / `code-block` 已改为暖白/陶土体系。
  - 2026-07-07 20:38 继续收口深层控制台剩余界面：账号创建/编辑/批量编辑/重授权/统计/状态/用量条、后台用户/分组/账号、分组/渠道通用控件、风控和 Ops 图表/日志/详情弹窗均统一到暖白面板、细边框、陶土 focus 和浅色 tooltip；成功/警告/危险/支付/平台品牌语义色保留。
  - 2026-07-07 21:05 继续收口支付/订阅/后台订单界面：`PaymentView` 从蓝色 SaaS 定价氛围切到暖白/陶土，用户订单退款入口、后台订单重试/退款、开票审批、套餐编辑、订阅管理提示和支付统计图表均统一到当前控制台浅色体系；支付渠道品牌色、成功/失败/警告状态色保留。
  - 2026-07-07 22:07 继续收口管理控制台首屏：`DashboardView`、`ModelDistributionChart`、`TokenUsageTrend` 的图标框、面板背景、tooltip、环图/折线/柱图 palette 从默认蓝绿紫改为暖白/陶土/低饱和鼠尾草体系；顶栏公告、公告铃铛弹窗、余额胶囊和侧栏“领额度”标记同步改为陶土色，保留危险/成功等必要语义色。
  - 排行榜页已复核为当前暖白风格基本一致，本轮未再扩大修改。
- Dev/预览修复：
  - `frontend/vite.config.ts` 把 dev proxy 从裸 `/setup` 改为 `'/setup/'`，裸 `/setup` 页面走 Vue Router，`/setup/status` 等 API 仍代理到后端。
  - `router/index.ts` 登录页 `titleKey` 从不存在的 `common.login` 改为 `auth.signIn`，消除 i18n warning。
- 测试同步：
  - 更新 `auth-theme`、`home-theme`、`public-pages`、`public-smoke`、`console-theme`、`style-theme` 等静态主题断言。

## 验证记录

- `npm.cmd run test:run -- auth-theme`：通过。
- `npm.cmd run test:run -- home-theme`：通过。
- `npm.cmd run test:run -- public-pages`：通过。
- `npm.cmd run test:run -- public-smoke`：通过。
- `npm.cmd run test:run -- console-theme`：通过。
- `npm.cmd run test:run -- style-theme`：通过。
- 合并主题回归：`npm.cmd run test:run -- public-pages public-smoke home-theme auth-theme style-theme console-theme`：6 files / 45 tests passed。
- 路由/认证回归：`npm.cmd run test:run -- src/router/__tests__/guards.spec.ts src/router/__tests__/title.spec.ts public-smoke public-pages`：4 files / 58 tests passed。
  - `npm.cmd run build`：通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - `git diff --check`：通过；仅有 Windows LF/CRLF 未来转换提示：`style.css`、`console-ui.css`、`public-page.css`、`knowledge/tasks/current-task.md`。
  - 2026-07-07 16:04 认证卡片收尾：`npm.cmd run test:run -- auth-theme home-theme public-smoke`、`npm.cmd run test:run -- auth-theme`、`npm.cmd run build`、`git diff --check` 通过；本地 `http://127.0.0.1:62086/api/v1/settings/public` 确认 `github_oauth_enabled=false`，因此本地不会显示 GitHub 按钮，需后台开启后才显示。
  - 2026-07-07 16:08 注册按钮圆角统一：`RegisterView.vue` 的创建账户按钮改用与登录按钮一致的 `auth-submit-button` 胶囊样式，并补 `auth-theme` 断言；`npm.cmd run test:run -- auth-theme`、`npm.cmd run build`、`git diff --check` 通过。
  - 2026-07-07 16:15 顶栏 CTA 收尾：`PublicTopNav.vue` 未登录态右上角移除 `/register` 黑色 CTA，仅保留 `/login`；`npm.cmd run test:run -- public-smoke public-pages home-theme`、`git diff --check` 通过。
  - 2026-07-07 16:30 认证卡片圆角/高度/顶栏收尾：`RegisterView.vue` 注册输入框切换到与登录一致的 `auth-input`；`AuthAccessPanel.vue` 桌面端固定最小高度；`PublicTopNav.vue` 未登录态右上角不再显示 `/login`；`npm.cmd run test:run -- auth-theme public-smoke public-pages home-theme`、`npm.cmd run build`、`git diff --check` 通过。
  - 2026-07-07 16:44 认证卡片回抖修复：`LoginView.vue`、`RegisterView.vue` 邮箱输入改为 `:autofocus="!embedded"`；`npm.cmd run test:run -- auth-theme`、`npm.cmd run build` 已通过；`git diff --check` 复测通过。
  - 2026-07-07 19:44 控制台旧色补齐：`npm.cmd run test:run -- console-theme style-theme` 通过，2 files / 13 tests passed。
  - 2026-07-07 19:45 主题回归：`npm.cmd run test:run -- auth-theme home-theme public-pages public-smoke console-theme style-theme` 通过，6 files / 51 tests passed。
  - 2026-07-07 19:45 `npm.cmd run build` 通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - 2026-07-07 19:46 `git diff --check` 通过；仅有 Windows LF/CRLF 未来转换提示。
  - 2026-07-07 20:28 深层控制台主题回归：`npm.cmd run test:run -- console-theme style-theme` 通过，2 files / 17 tests passed。
  - 2026-07-07 20:29 公共/认证/首页回归：`npm.cmd run test:run -- auth-theme home-theme public-pages public-smoke` 通过，4 files / 34 tests passed。
  - 2026-07-07 20:30 `npm.cmd run build` 通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - 2026-07-07 20:31 `git diff --check` 通过；仅有 Windows LF/CRLF 未来转换提示。
  - 2026-07-07 20:36 容器替换后复跑 `npm.cmd run test:run -- console-theme style-theme` 通过，2 files / 17 tests passed。
  - 2026-07-07 21:05 支付/订单主题回归：`npm.cmd run test:run -- console-theme style-theme` 通过，2 files / 18 tests passed；`npm.cmd run test:run -- auth-theme home-theme public-pages public-smoke` 通过，4 files / 38 tests passed。
  - 2026-07-07 21:05 支付相关组件/页面回归：`npm.cmd run test:run -- src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/admin/orders/__tests__/AdminInvoiceRequestsView.spec.ts` 通过，5 files / 23 tests passed。
  - 2026-07-07 21:05 `npm.cmd run build` 通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - 2026-07-07 21:10 `git diff --check` 通过；仅有 Windows LF/CRLF 未来转换提示。
  - 2026-07-07 21:56 dashboard 首屏主题回归：`npm.cmd run test:run -- console-theme style-theme` 通过，2 files / 20 tests passed。
  - 2026-07-07 21:56 公共/认证/首页回归：`npm.cmd run test:run -- auth-theme home-theme public-pages public-smoke` 通过，4 files / 38 tests passed。
  - 2026-07-07 21:57 dashboard 图表/页面定向回归：`npm.cmd run test:run -- src/views/admin/__tests__/DashboardView.spec.ts src/components/charts/__tests__/ModelDistributionChart.spec.ts src/components/charts/__tests__/TokenUsageTrend.spec.ts` 通过，3 files / 9 tests passed；`ModelDistributionChart.spec.ts` 已同步 Others chart color 预期到新暖色 palette。
  - 2026-07-07 21:57 `npm.cmd run build` 通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - 2026-07-07 21:58 `git diff --check` 通过；仅有 Windows LF/CRLF 未来转换提示。
- 子智能体只读验收：确认 `/models` 源码保留 `PublicTopNav`、`h1`、暖白公共页壳；确认旧 Matrix/PixelIcon 在当前 public/auth/console 入口基本无明显引用；指出 `/setup` 预览问题来自 dev proxy。

## Chrome / 本地预览检查

- 当前可用预览：`http://127.0.0.1:62086/`，PID `23480`，显式 `VITE_DEV_PROXY_TARGET=http://127.0.0.1:62080`。
- 已停止本轮临时 `62085`；旧 `62083`、`62084` 可能仍是此前会话/预览进程，不作为本轮最终验收依据。
- `http://127.0.0.1:62086/setup`：返回 Vite dev HTML，包含 `/src/main.ts`，不再被后端旧构建/VitePress 抢走。
- `http://127.0.0.1:62086/setup/status`：返回后端 JSON `{"needs_setup":false,"step":"completed"}`。
- `http://127.0.0.1:62086/api/v1/model-market/catalog`：返回后端模型目录 JSON。
- Chrome 检查 `http://127.0.0.1:62085/models`（修复前端口）：在正确 API 代理下有 fixed 毛玻璃 `.public-top-shell`、标题“模型定价”、11 个模型卡、暖白背景、无 VitePress；后续 62086 API 代理同样正常。
- Chrome 检查 `http://127.0.0.1:62085/login?redirect=/dashboard`：首页壳内嵌认证卡片存在，默认登录 tab，有邮箱和密码输入框，无横向溢出。
- Chrome 检查 `http://127.0.0.1:62086/home`：未登录顶栏链接仅有品牌、首页、教程、模型广场，无 `/login` 和 `/register`；认证卡片注册/登录切换高度均为 496px，heightDelta=0，输入框圆角 10px，tab/提交按钮圆角 999px。
- Chrome 复测 `http://127.0.0.1:62086/home`：注册 tab 切到登录 tab 后 `scrollY` 从 0 到 0，`heightDelta=0`，焦点停留在 tab 按钮，不再自动跳到邮箱输入框。
- 2026-07-07 19:55 已更新本地 `62080` 容器：
  - `local-docker-update-guard` 获取并释放 `sub2api` 锁成功。
  - Docker Hub metadata 拉取失败后，改用本地 `go1.26.3 windows/amd64` 编译 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags embed`，再基于当前 runtime 镜像替换 `/app/sub2api` 并 commit 新镜像。
  - 当前容器：`sub2api`，镜像 `sub2api:codex-20260707-1948-warm-ui`，状态 `healthy`，端口 `127.0.0.1:62080->8080/tcp`。
  - 回滚镜像 tag：`sub2api:before-codex-20260707-1948-warm-ui` 指向替换前镜像 ID。
  - HTTP smoke：`http://127.0.0.1:62080/health` 返回 200 / `{"status":"ok"}`；`/home`、`/models` 返回 200 且包含 Vue app root。
  - Chrome 抽查 `http://127.0.0.1:62080/home`：暖白 body、fixed 公共顶栏、reveal 背景存在，`scrollY=0`。
  - Chrome 抽查 `http://127.0.0.1:62080/models`、`/tutorial/`：暖白 body、公共顶栏、reveal 背景存在，`scrollWidth == clientWidth`，无横向溢出。
- 2026-07-07 20:32 深层控制台收口后再次更新本地 `62080` 容器：
  - `local-docker-update-guard` 续用并释放 `sub2api` 锁成功。
  - 使用已编译的 Linux/amd64 embed 二进制 `backend/.codex-tmp-build/sub2api-linux-amd64` 替换容器内 `/app/sub2api` 并重启。
  - 当前运行容器：`sub2api`，Docker 配置仍显示原始镜像 `sub2api:codex-20260707-1948-warm-ui`，状态 `healthy`，端口 `127.0.0.1:62080->8080/tcp`。
  - 新镜像 tag：`sub2api:codex-20260707-2027-warm-console`，image id `e08fadf6c392`。
  - 回滚镜像 tag：`sub2api:before-codex-20260707-2027-warm-console`，image id `2bdaac52bfd1`。
  - HTTP smoke：`/health` 返回 200 / `{"status":"ok"}`；`/home`、`/models`、`/tutorial`、`/dashboard`、`/my-accounts`、`/usage`、`/orders`、`/admin/accounts`、`/admin/groups`、`/admin/users`、`/admin/risk-control` 均返回前端入口 HTML。
  - Chrome 抽查：`/dashboard`、`/my-accounts`、`/usage`、`/orders`、`/admin/accounts`、`/admin/groups`、`/admin/users`、`/admin/ops` 都是暖白控制台壳，`scrollWidth == clientWidth`，未命中旧暗色节点选择器。
  - 当前 `homeContent` 已启用，`/home` 展示自定义首页 `mapleAI 智能解决方案`，仍有 fixed 公共顶栏、暖白 body 和 reveal 背景；这符合“保留 `homeContent` 自定义首页逻辑”的边界。
  - 当前公开设置未返回 `risk_control_enabled=true`，所以访问 `/admin/risk-control` 会被路由守卫导向 `/admin/settings`；这是配置行为，不是页面加载失败。
- 2026-07-07 21:10 支付/订单/订阅收口后再次更新本地 `62080` 容器：
  - `local-docker-update-guard` 获取并释放 `sub2api` 锁成功。
  - 重新执行 `frontend/ npm.cmd run build` 后，用本地 `go build -tags embed ./cmd/server` 重新生成 Linux/amd64 embed 二进制，替换容器内 `/app/sub2api` 并重启。
  - 当前运行容器：`sub2api`，Docker 配置仍显示原始镜像 `sub2api:codex-20260707-1948-warm-ui`，状态最终复查为 `healthy`，端口 `127.0.0.1:62080->8080/tcp`。
  - 新镜像 tag：`sub2api:codex-20260707-2105-warm-payment`，image id `38ca8053f76d`。
  - 回滚镜像 tag：`sub2api:before-codex-20260707-2105-warm-payment`，image id `4f4ddbba955f`。
  - HTTP smoke：`/health` 返回 200 / `{"status":"ok"}`；`/home`、`/purchase`、`/orders`、`/admin/orders`、`/admin/orders/invoices`、`/admin/orders/plans`、`/admin/subscriptions` 均返回前端入口 HTML。
  - Chrome 抽查：`/home`、`/purchase`、`/orders`、`/admin/orders`、`/admin/orders/invoices`、`/admin/orders/plans`、`/admin/subscriptions` 均为暖白背景或控制台浅色壳，支付/订单导航高亮为陶土色，`scrollWidth == clientWidth`，无横向溢出。
- 2026-07-07 22:04 dashboard 首屏收口后再次更新本地 `62080` 容器：
  - `local-docker-update-guard` 获取并释放 `sub2api` 锁成功。
  - 重新执行 `frontend/ npm.cmd run build` 后，用本地 Linux/amd64 embed 二进制替换容器内 `/app/sub2api` 并重启；本轮曾误生成 `backend/backend/.codex-tmp-build`，已只清理该本轮临时目录，未动更早的 `backend/backend/.codex-runtime`。
  - 当前运行容器：`sub2api`，Docker 配置仍显示原始镜像 `sub2api:codex-20260707-1948-warm-ui`，最终复查 `health=healthy`，端口 `127.0.0.1:62080->8080/tcp`。
  - 新镜像 tag：`sub2api:codex-20260707-2205-warm-dashboard`，image id `8387f21a0fe2`。
  - 回滚镜像 tag：`sub2api:before-codex-20260707-2205-warm-dashboard`，image id `2bdaac52bfd1`。
  - HTTP smoke：`/health`、`/home`、`/admin/dashboard`、`/admin/accounts`、`/models` 均返回 200。
  - Chrome 抽查 `http://127.0.0.1:62080/admin/dashboard`：已登录态可进入后台，`hasDashboard=true`，`cardCount=14`；公告 label、余额胶囊、侧栏“领额度”、首个统计图标和 dashboard title icon 均为陶土/暖白色系；body 背景为 `rgb(250, 249, 245)`。

## 相关文件

- 新增：`frontend/src/components/auth/AuthAccessPanel.vue`。
- 认证/首页/路由：`frontend/src/views/HomeView.vue`、`frontend/src/views/auth/LoginView.vue`、`frontend/src/views/auth/RegisterView.vue`、`frontend/src/components/layout/AuthLayout.vue`、`frontend/src/router/index.ts`。
- 公共页：`frontend/src/views/public/ModelPlazaView.vue`、`TutorialView.vue`、`LegalDocumentView.vue`、`components/PublicTopNav.vue`、`components/PublicRevealBackdrop.vue`、`public-page.css`。
- 全局/控制台/setup：`frontend/src/style.css`、`frontend/src/styles/console-ui.css`、`frontend/tailwind.config.js`、`frontend/vite.config.ts`、`frontend/src/views/KeyUsageView.vue`、`frontend/src/views/setup/SetupWizardView.vue`、`frontend/src/views/NotFoundView.vue`。
- 本轮继续补齐：`frontend/src/styles/onboarding.css`、`frontend/src/components/account/AccountTestModal.vue`、`frontend/src/components/admin/account/AccountTestModal.vue`、`frontend/src/components/admin/usage/UsageTable.vue`、`frontend/src/views/user/UsageView.vue`、`frontend/src/components/admin/payment/AdminRefundDialog.vue`、`frontend/src/views/user/UserOrdersView.vue`、`frontend/src/components/common/SubscriptionProgressMini.vue`。
- 深层控制台收口：`frontend/src/components/common/BaseDialog.vue`、`frontend/src/components/account/CreateAccountModal.vue`、`EditAccountModal.vue`、`BulkEditAccountModal.vue`、`ReAuthAccountModal.vue`、`AccountStatsModal.vue`、`AccountStatusIndicator.vue`、`AccountUsageCell.vue`、`UsageProgressBar.vue`、`AccountQuotaInfo.vue`、`frontend/src/components/admin/account/AccountActionMenu.vue`、`frontend/src/views/admin/AccountsView.vue`、`GroupsView.vue`、`UsersView.vue`、`RiskControlView.vue`、`frontend/src/views/admin/ops/components/*`。
- 支付/订单/订阅收口：`frontend/src/views/user/PaymentView.vue`、`frontend/src/views/user/UserOrdersView.vue`、`frontend/src/views/admin/orders/AdminOrdersView.vue`、`AdminInvoiceRequestsView.vue`、`AdminPaymentPlansView.vue`、`frontend/src/views/admin/SubscriptionsView.vue`、`frontend/src/components/admin/payment/OrderStatsCards.vue`、`DailyRevenueChart.vue`。
- dashboard 首屏收口：`frontend/src/views/admin/DashboardView.vue`、`frontend/src/components/charts/ModelDistributionChart.vue`、`TokenUsageTrend.vue`、`frontend/src/components/common/AnnouncementBell.vue`、`frontend/src/components/layout/HeaderAnnouncementCarousel.vue`、`AppSidebar.vue`、`frontend/src/styles/console-ui.css`、`frontend/src/components/charts/__tests__/ModelDistributionChart.spec.ts`。
- 测试：`frontend/src/__tests__/auth-theme.spec.ts`、`console-theme.spec.ts`、`home-theme.spec.ts`、`public-pages.spec.ts`、`public-smoke.spec.ts`、`style-theme.spec.ts`。
- i18n：`frontend/src/i18n/locales/zh/auth.ts`、`en/auth.ts`、`zh/home.ts`、`en/home.ts`。
- build 刷新：`backend/internal/web/dist/**` 因 `npm.cmd run build` 重新生成。
- Unrelated dirty/untracked 仍包括 `.codex-tmp-*`、`DESIGN-apple.md`、backend/admin/leaderboard 等旧脏改，未处理。

## 2026-07-08 S54 多智能体收口补充

- 用户明确要求“启用多智能体开发”，本轮按 `docs/workflow/tasks/ui-warm-leftovers-s54.md` 继续执行暖白 / 陶土 / 黑灰前端收口。
- 第一轮 worker：
  - `admin-a` 收口 `SettingsView`、`PromoCodesView`、`TutorialPagesView`、`BackupView` 等后台配置页面残留。
  - `monitor-b` 收口 monitor 模板管理对话框普通紫色 badge。
  - `user-c` 收口用户邀请、兑换、订阅、支付状态、profile 通知/TOTP 等绿色/蓝紫残留。
- 第二轮 worker：
  - 后台分组/用户弹窗：`GroupRPMOverridesModal`、`GroupRateMultipliersModal`、`GroupAccountPriorityModal`、`UserBalanceHistoryModal`、`UserAllowedGroupsModal`、`GroupReplaceModal`、`ErrorPassthroughRulesModal`、`TLSFingerprintProfilesModal`、`UserEditModal`。
  - 用户账户/仪表盘/profile：`MyAccountsView`、`UserApiKeyOnboardingDialog`、`UserDashboardAccountUsage`、`UserDashboardPerformanceStats`、`UserDashboardQuickActions`、`UserDashboardRecentUsage`、`ProfileIdentityBindingsSection`。
  - 用户支付/福利/订单结果：`WelfareView`、`PaymentResultView`、`StripePaymentView`、`StripePopupView`、`AirwallexPaymentView`、`UserOrdersView`。
  - 主控发现 worker 曾删除仪表盘“兑换码”快捷入口；用户随后明确要求删除兑换码，因此最终保留删除结果。
- 第三轮 worker：
  - 认证/common/setup：`ForgotPasswordView`、`EmailVerifyView`、`RegisterView`、`ResetPasswordView`、`SetupWizardView`、`WechatOAuthSection`、`PendingOAuthCreateAccountForm`、`AnnouncementBell`、`SubscriptionProgressMini`、`VersionBadge`。
  - 后台 usage/risk/发票/兑换：`UsageStatsCards`、`UsageTable`、`admin/RedeemView`、`RiskControlView`、`AdminInvoiceRequestsView`。
  - `KeyUsageView` 公共查询页从旧 emerald/indigo 装饰切到陶土/鼠尾草/暖白；主控确认只改颜色 token，未动请求或统计计算。
- 主控额外收口：
  - `PaymentProviderDialog` webhook 提示、`ProviderCard` 启用态/编辑 hover、`StripePaymentInline` 金额提示色。
  - `PaymentView` 会员进度、价格数值、深色模式蓝色高光和成功提示。
  - `RiskControlView` 中 worker slot 的 active/idle 状态重新区分：active 陶土，idle 鼠尾草，避免状态信息丢失。
- 本轮刻意保留：
  - 支付品牌色：支付宝、微信、Stripe。
  - 平台/分类品牌色：OpenAI、Gemini、Antigravity 等用于明确平台识别的色块。
  - 危险红、警告黄、错误/风险等级等真实语义色。

## 2026-07-08 S54 验证与容器

- 主题测试：
  - `npm.cmd run test:run -- console-theme style-theme`：通过，2 files / 23 tests passed。
  - `npm.cmd run test:run -- auth-theme home-theme public-pages public-smoke`：通过，4 files / 38 tests passed。
- 构建：
  - `npm.cmd run build`：通过；仍有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
- Diff：
  - `git diff --check`：通过；仅有 Windows LF/CRLF 未来转换提示。
- 本地 `62080` 容器：
  - 使用 `local-docker-update-guard` 获取并释放 `sub2api` 锁成功。
  - 重新执行 `frontend/ npm.cmd run build` 后，用 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags embed -o .codex-tmp-build/sub2api-linux-amd64 ./cmd/server` 生成 Linux/amd64 embed 二进制。
  - 容器内 `/app/sub2api` 已替换为本轮新二进制；容器内 hash `9d2f187ec416f1dde9f4b4cd9dcdffc2b05eafc9fa04244ee5d2f788ac06f195` 与本地构建产物一致。
  - 当前运行容器：`sub2api`，状态 `healthy`，端口 `127.0.0.1:62080->8080/tcp`。Docker 配置仍显示基础镜像 tag `sub2api:codex-20260707-1948-warm-ui`，这是因为本轮采用保留容器并替换内置二进制的方式。
  - 新镜像 tag：`sub2api:codex-20260708-0115-warm-ui-workers`，image id `sha256:1bd6abe4d38b97078f7064d3a109681fdd2c66074b4aa5fdb74e1013422d96bc`。
  - 回滚镜像 tag：`sub2api:before-codex-20260708-0115-warm-ui-workers`，image id `sha256:199a62e25e2ac4d3eece5d46a90a6cd1995a3200b3f2627a16adcb021741df69`。
  - HTTP smoke：`/health`、`/home`、`/dashboard`、`/admin/accounts`、`/admin/risk-control`、`/admin/ops`、`/keys`、`/usage`、`/monitor`、`/tickets`、`/purchase`、`/subscriptions`、`/profile`、`/redeem`、`/affiliate`、`/key-usage` 全部返回 200。

## 2026-07-08 首页已登录账号入口收口

- 用户要求“登录状态右上角的名字按钮移除，放到左侧的账号框里面”。
- 已完成：
  - `PublicTopNav.vue` 移除 `/home` 已登录态右上角用户名 chip；`showDashboardButton` 仅在已登录且非 `/home` / `/` 的公共页显示，因此 `/models`、`/tutorial` 等仍保留“前往仪表盘”。
  - `HomeView.vue` 在已登录的 `home-account-workbench` 账号工作台卡片 header 内加入 `home-account-user-chip`，显示 `authStore.user.username`，缺失时回退 `authStore.user.email`，再缺失时回退仪表盘文案。
  - `home-theme`、`public-pages`、`public-smoke` 断言已同步，防止 `public-user-chip` / `showHomeUserName` 回流。
- 验证：
  - `npm.cmd run test:run -- home-theme public-pages public-smoke`：通过，3 files / 34 tests passed。
  - `npm.cmd run build`：通过；仍只有既有 Vite dynamic/static import chunk 警告、大 chunk 警告、Browserslist 过期提示和 Node `DEP0190` 警告。
  - `git diff --check -- frontend/src/views/HomeView.vue frontend/src/views/public/components/PublicTopNav.vue frontend/src/__tests__/home-theme.spec.ts frontend/src/__tests__/public-pages.spec.ts frontend/src/__tests__/public-smoke.spec.ts backend/internal/web/dist`：通过。
- 本地 `62080` 容器：
  - 使用 `local-docker-update-guard` 获取并释放 `sub2api` 锁成功。
  - 用 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags embed -o .codex-tmp-build/sub2api-linux-amd64 ./cmd/server` 重新生成 Linux/amd64 embed 二进制。
  - 容器内 `/app/sub2api` 已替换；本地和容器 hash 均为 `9a52449069982b2d81d46e9a57957430ace724cadc7bcceba2c2789eccbad299`。
  - 当前运行容器：`sub2api`，状态复查为 `healthy`，端口 `127.0.0.1:62080->8080/tcp`。
  - 新镜像 tag：`sub2api:codex-20260708-0316-home-user-chip-card`，image id `sha256:0d901874e697d08af6065b5b999d964005603f8385881636cab3019e2e09dce7`。
  - 回滚镜像 tag：`sub2api:before-codex-20260708-0316-home-user-chip-card`，image id `sha256:8d91a6494e0c06b1df828b1194340c3b0e364aa0ad872b8acab29d44db8d82c5`。
  - HTTP smoke：`/health`、`/home`、`/models`、`/tutorial`、`/dashboard` 全部返回 200。
  - Chrome 抽查 `http://127.0.0.1:62080/home`：已登录态显示账号工作台卡片，`public-user-chip` 数量为 0，`.public-auth-button` 数量为 0，`.home-account-user-chip` 在账号卡片内显示当前用户标识；无横向溢出。

## 2026-07-08 S55 upstream 小批后端补丁

- 已完成：在隔离 worktree `E:/codex-worktrees/sub2api/upstream-s55-small-safe-patches` 基于 `origin/main` cherry-pick 并验证 5 个上游小补丁，随后通过 `E:/codex-worktrees/sub2api/main-merge-s55-r2` 合回并推送到 `origin/main`。
- 已合入远端主线：
  - `00b706341 merge: upstream v0.1.146 small safe patch s55`，当前 `origin/main` 已确认指向该提交。
  - S55 包含 `438f17be5` compact JSON/SSE usage 修复、`fd64d07e6` Codex function_call 非法 `item_*` id 剥离、`cbfeab964` Antigravity 默认生产 endpoint、`a1b2b32e0` usage_logs 队列溢出不静默丢弃、`f3a3a0869` 并发槽位清理优化。
- 验证：
  - `go test ./internal/service -run "Test.*(Compact|SSE|Codex|FunctionCall|Antigravity|Usage|Queue|Concurrency|Slot)" -count=1`：通过。
  - `go test ./internal/repository -run "Test.*(UsageLog|Concurrency)" -count=1`：通过。
  - `go test ./internal/config -count=1`：通过。
  - `go test ./internal/handler -run "Test.*(Gateway|Concurrency|Warmup|Fastpath|Hotpath)" -count=1`：通过。
  - `git diff --check origin/main..HEAD`：通过。
- 清理：
  - 已删除本轮临时 worktree：`main-merge-s55`、`main-merge-s55-r2`、`upstream-s55-small-safe-patches`。
  - 已删除本轮临时分支：`codex/main-merge-s55`、`codex/main-merge-s55-r2`、`codex/upstream-s55-small-safe-patches`。
- 注意：
  - S55 未更新本地 `62080` 容器；当前运行容器仍是前端首页账号入口收口后的镜像状态。
  - 本轮中途发现 `origin/main` 新增 `6f99b9d8f feat(user): merge subscriptions into usage page`，已改用 r2 merge worktree 基于最新远端重新合并，避免覆盖远端提交。
  - 冲突或大功能 upstream 候选仍暂缓：websearch 历史块过滤、请求体解析错误日志、image_gen namespace、Grok video、messages fallback、batch image 系列。

## 下一步

1. 如要让本地 `62080` 容器吃到 S55 后端补丁，需按 `local-docker-update-guard` 获取锁后重新构建/替换容器二进制并跑 `/health` smoke。
2. 如继续 upstream 合并，优先单独评估仍暂缓的冲突项；不要直接全量合 `upstream/main` 或 `v0.1.146`。
3. 如继续 UI 收口，建议按用户截图逐页处理后台公告管理、渠道/监控、设置页深层弹窗、订阅创建/编辑弹窗、风控开关开启后的 `/admin/risk-control` 实际内容；不要机械替换品牌色或危险/成功/警告语义色。
