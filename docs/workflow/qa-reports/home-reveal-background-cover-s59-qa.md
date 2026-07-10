### PASS: home-reveal-background-cover-s59

## Findings
- 原 hero 背景使用 `min(2048px, max(100vw, 90rem)) auto`，超宽屏超过 2048px 后停止增长；同时只按宽度缩放，不能保证覆盖 hero 高度。
- 原移动端使用固定 `inset: 4.2rem 0 auto`、`height: 18rem` 和 `background-size: 48rem auto`，不同视口会出现上下空带或错位。
- 原 canvas mask 基础透明度为 1，未揭示区域完全遮住背景，视觉上会被误判为背景未铺满。

## Executed Checks
- `npm.cmd run test:run -- home-theme public-pages`
  - PASS：2 files / 29 tests。
- `npm.cmd run build`
  - PASS：`vue-tsc -b` 与 Vite production build 通过。
  - 保留既有 dynamic/static import、chunk size、Browserslist 和 Node `DEP0190` 警告。
- Chrome + Playwright，预览 `http://127.0.0.1:62087/home`，代理 `http://127.0.0.1:62080`：
  - 超宽 `2560x900`：hero、背景层和 canvas 均为 `2560x678`，`background-size: cover`，无横向溢出。
  - 桌面 `1366x768`：hero 与背景层均为 `1366x605`，无横向溢出。
  - 移动 `390x844`：hero、背景层和图片层均为 `390x574.875`，mask 隐藏，静态背景透明度为 `0.12`，无横向溢出。
- `git diff --check`
  - PASS；仅 workflow/status 和 handoff 文件存在既有 LF/CRLF future-conversion warning。
- 变更边界审计：
  - S59 代码只修改 `PublicRevealBackdrop.vue` 与对应 `home-theme.spec.ts` 断言。
  - 工作区其余首页改动属于已存在的 S58 未提交内容，本 Sprint 未回滚或扩写 denied paths。

## Evidence
- `output/playwright/home-s59-wide-2560-final.png`
- `output/playwright/home-s59-desktop-1366-final.png`
- `output/playwright/home-s59-mobile-390-final.png`

## Unverified Risks
- `background-size: cover` 会按视口比例裁切图片边缘，这是避免空带的预期行为；若必须完整展示原图，需要改为独立图片布局并接受留白或重新制作多比例素材。
- 本 Sprint 未更新本地 `62080` 容器，当前验证基于 `62087` Vite 预览。

## Recommendation
- PASS。使用容器级 `cover` 后，超宽、常规桌面和移动视口均无背景空带；桌面保留局部 reveal，触屏、窄屏与 reduced-motion 使用低对比度静态背景。
