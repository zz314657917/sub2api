# Task Contract

## Task ID
home-conversion-accessibility-s58

## Role
你是 P/G/E 流程里的首页体验收口执行者。只执行本 contract，不扩大到后端、数据库或部署链路。

## Goal
根据桌面与移动端首页评审结果，收口首页注册主路径、移动端首屏密度、API 地址复制、可信信息、登录入口、协议标题异常回退和基础可访问性。

## Success Criteria
- 未登录首页的主 CTA 不再跳离当前可见认证流程，而是展开、滚动并聚焦内嵌认证面板。
- `/login`、`/register` 仍直接展示对应认证面板，不受首页移动端折叠逻辑影响。
- 移动端默认首页首屏不再连续堆叠完整注册表单，主次 CTA 清晰且无横向溢出。
- API 入口支持复制并提供明确成功反馈。
- 首页展示不依赖虚构额度的可信信息，并提供模型广场入口。
- 后台协议标题为空或仅含 `?` / Unicode replacement character 时，前端按 document id 使用本地化安全标题；不修改数据库原始数据。
- 未登录用户可以从公共顶栏直接进入登录流程。
- FAQ 按钮与答案区域具有完整 `aria-controls` / `id` 关联，轮换营销文案不使用 live region。

## Allowed Paths
- `frontend/src/views/HomeView.vue`
- `frontend/src/views/public/components/PublicTopNav.vue`
- `frontend/src/components/auth/AuthAccessPanel.vue`
- `frontend/src/utils/legalDocuments.ts`
- `frontend/src/i18n/locales/zh/home.ts`
- `frontend/src/i18n/locales/en/home.ts`
- `frontend/src/__tests__/home-theme.spec.ts`
- `frontend/src/__tests__/public-pages.spec.ts`
- `frontend/src/__tests__/public-smoke.spec.ts`
- `frontend/src/utils/__tests__/legalDocuments.spec.ts`
- `docs/workflow/tasks/home-conversion-accessibility-s58.md`
- `docs/workflow/status.md`
- `docs/workflow/qa-reports/home-conversion-accessibility-s58-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/**`
- `frontend/src/router/**`
- `frontend/src/views/auth/**`
- `frontend/src/stores/**`
- `deploy/**`
- 数据库、migration、容器和生产配置。

## Constraints
- 保留 `homeContent` 自定义首页逻辑。
- 不写死试用金额、有效期、无需绑卡等未由公开配置证明的承诺。
- 不回滚用户已有改动，不做无关格式化。
- 协议乱码只做显示层安全回退；原始数据修复必须另行确认。
- 不 stage、不 commit。

## Acceptance Commands
```powershell
cd frontend
npm.cmd run test:run -- home-theme public-pages src/utils/__tests__/legalDocuments.spec.ts
npm.cmd run build
cd ..
git diff --check
```

## Runtime Acceptance
- Chrome 桌面 `1440x1000` 与移动 `390x844` 检查 `/home`。
- 两种视口均无横向溢出。
- 移动端首次进入 `/home` 时完整认证面板默认折叠；点击主 CTA 后面板出现并聚焦首个输入框。
- `/login` 直接显示登录面板。
- API 地址复制按钮可用并显示复制完成状态。
- FAQ 展开收起正常。

## Stop Rules
- 需要修改数据库协议数据、后端公开设置结构或生产 API 地址时停止并单独报告。
- 需要改变注册、登录、OAuth、验证码或支付业务逻辑时停止。
