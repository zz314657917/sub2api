### PASS: home-conversion-accessibility-s58

## Scope
- 首页注册主路径、移动首屏、API 地址复制、可信信息、登录入口、协议标题异常回退和 FAQ 可访问性。
- 未修改后端、数据库、路由、部署或容器。

## Executed Checks
- `npm.cmd run test:run -- home-theme public-pages public-smoke auth-theme src/utils/__tests__/legalDocuments.spec.ts`
  - PASS：5 files / 43 tests。
- `npm.cmd run build`
  - PASS：`vue-tsc -b` 与 Vite production build 通过。
  - 保留既有 dynamic/static import、chunk size 和 Browserslist 警告。
- Chrome + Playwright，预览 `http://127.0.0.1:62087`，代理 `http://127.0.0.1:62080`：
  - 桌面 `1440x1000` 与移动 `390x844` 均无横向溢出。
  - 移动 `/home` 初始认证面板 `display:none`；点击“注册领取试用”后显示面板并聚焦 email 输入框。
  - `/login` 直接显示登录面板且登录 tab 选中。
  - API 入口复制后状态显示“API 地址已复制”，剪贴板内容与当前入口一致。
  - 后台协议标题为 `????` 时，首页与认证面板显示本地化回退标题，页面无 `????`。
  - FAQ 展开收起与 `aria-controls` 关联正常。
- `git diff --check`
  - PASS；仅 workflow/status 文件存在既有 LF/CRLF future-conversion warning。

## Evidence
- `output/playwright/home-s58-desktop.png`
- `output/playwright/home-s58-mobile-initial.png`
- `output/playwright/home-s58-mobile-expanded.png`

## Unverified Risks
- 数据库中的协议标题原始值仍是 ASCII `?`；本 Sprint 只做显示层安全回退。
- 未更新本地 `62080` 容器；最终用户运行态仍需在后续部署批次构建 embed 二进制并替换容器。

## Verdict
- PASS。代码、测试、生产构建和浏览器验收证据满足 S58 contract。
