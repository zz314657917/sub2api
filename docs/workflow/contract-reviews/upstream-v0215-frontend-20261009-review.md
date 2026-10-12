---
status: approved
verdict: PASS
task_id: upstream-v0215-frontend-20261009
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
reviewer: current-session-default
last_verified: 2026-10-10
---
### PASS: upstream-v0215-frontend-20261009

## Findings
- **原 FAIL 已修复，修订复核 PASS。** 合同两个 locale allowed paths 已替换为实际存在、被各自 `admin/index.ts` 导入并导出的 `frontend/src/i18n/locales/en/admin/usage.ts` 和 `frontend/src/i18n/locales/zh/admin/usage.ts`。
- Success Criteria 已明确仅向这些已加载模块增加 neutral long-context keys，禁止 broad upstream locale replacement；原 locale allowlist/接线阻碍消除，无剩余必须修订的合同问题。PASS 仅指合同审查，不代表实现或 QA 通过。
- 其余 numeric/lifecycle 入口均存在，与两个 backend 合同业务 owner 无重叠；不要求修全库、不要求安装缺失依赖。

## Executed Checks
- 修订复核仅重读 frontend 合同和本报告，核对上述两个 locale 文件存在及中英文 `admin/index.ts` 的 `import usage from './usage'` / `usage` 导出；未重查或改动其他合同、报告和已委派业务文件。以下为前轮执行证据。
- 读取 spec、三合同、isolated status/matrix；核对十六个上游 commit 的路径及 d5f490e10 具体 diff。
- 检查本地 `UsageTable.vue:195-198` 的硬编码 x2、`locales/en.ts` 和两个 `admin/index.ts` 的加载链、既有 `admin/usage.ts`。
- 18 个现存精确业务路径 SHA256 均等于外置 `baseline.json`。两个 group modal 的 copied dirty `onUnmounted` 清理、`clearTimeout(searchTimeout)`、document listener removal 必须保留。
- 确认 node_modules 是指向共享原仓库依赖的 junction；检查安装的 Vite/Vitest 缓存逻辑和主控外置配置。本 reviewer 未执行前端测试、构建或浏览器。

## Gates / Unverified Risks
- 使用主控已提供配置，具体新/相关测试文件必须展开，不把 PATHS 当字面命令：
  `node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs <明确测试文件列表>`
- 该配置显式 root 为本 worktree，cacheDir 在 E runtime，test.cache=false；默认 Vitest 命令会写 junction 下的 `node_modules/.vite/vitest`，不得省略外置配置。不得写共享依赖、缓存或安装包。
- 保留 `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` 证据；主控已确认的 `@airwallex/components-sdk` 缺依赖是独立 baseline blocker，不能安装依赖或修改 manifests/lockfiles。本轮新增类型错误仍必须失败，不得用 baseline blocker 掩盖。
- production Vite build 显式指定 E runtime outDir；默认 vite.config.ts 的 outDir 是 `../backend/internal/web/dist`，不能直接用默认产物路径。需要临时配置时也只能外置，不能修改本轮 allowlist 外的配置。
- 主控报告的六个 UsageProgressBar baseline tests PASS 不等于本合同新测试 PASS。须为所有变更添加 mounted/deferred-promise、numeric 边界测试，并报告逐项覆盖和非零执行数量。
- 验收只声明 mounted interaction，不声明浏览器视觉、真实 API 或部署通过；baseline 与 task delta 分别检查。
- 用户已授权本轮 current-session-default native agents，不改矩阵；合同修订复审 PASS 后，主控再批准 frontmatter。

## Recommendation
修订后的合同可由主控批准 frontmatter 并派开发；保留上述外置缓存、baseline 保护及独立 QA 门禁。本次仅更新本 frontend review，不改合同、其他报告、业务文件、知识库或依赖，不提交、推送或部署。

## Revision Review — 2026-10-10
### PASS: upstream-v0215-frontend-20261009 — narrow UsageTable expected exception
- 只读复核本次合同扩展：allowlist 新增 `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`；Constraints 仅允许长上下文 marker expected 和必要的 mock translation labels 适配，明确不删除测试、不改其他 copied 内容。扩展直接对应原 Success Criteria 的中性 long-context 文案，不扩大产品范围或其他 owner。
- 静态核对：旧测试第 148 行仍 `toBe('x2')`，mock messages 缺新增 long-context labels；当前组件使用 `admin.usage.longContext` 和 `longContextPricingTooltip`，中英文已加载 usage.ts 含对应 tooltip。因此主控报告的 18 PASS / 1 FAIL 是本轮 expected 行为变化导致，不是 baseline 预失败，不得跳过该用例或用 baseline blocker 掩盖。
- 此旧测试文件当前 SHA256 与 immutable copied baseline 一致、相对 HEAD 无 diff。修改后的增量必须只含 marker expected 和确有必要的对应 mock labels；保留测试名称、所有其他断言与 copied 内容。mock 文案须匹配真实已加载中性 locale，不能为了绿测继续返回 x2 或仅检查元素存在。
- **修后必须重跑旧 UsageTable.spec.ts 全部 19 个相关测试并全部通过：** 从 frontend 执行 `node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs src/components/admin/usage/__tests__/UsageTable.spec.ts`，记录完整非零数量、退出码和日志，不得只用 `-t` 跑失败用例替代全文件。主控所报新增 88 tests PASS 不替代该门禁；新/相关回归及独立 QA 仍须按原合同执行。
- 既有 typecheck/build Airwallex 缺依赖限制不在本轮重审；继续单列未验证项，禁止安装依赖、修改共享 junction 或伪造 full/production build PASS。本 revision PASS 只批准合同最小测试扩展，不是修复执行或产品验收 PASS。
- 本次仅追加本 frontend review；未改测试、合同、其他报告、源代码或依赖，未运行测试。
