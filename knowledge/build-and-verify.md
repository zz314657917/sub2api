# 构建与验证

最后更新：2026-06-17

## 基本原则

- Windows 本地优先使用原生命令：`go`、`npm.cmd`、`pnpm.cmd`。
- PowerShell 执行中文输出前先设置 UTF-8。
- 前端项目声明使用 pnpm，但当前会话历史中也有 `npm.cmd run ...` 的验证记录；改依赖时以 `pnpm-lock.yaml` 为准，不要只改 `package.json`。
- 不要把 `.gocache/`、`.venv/`、`output/` 当成源码扫描目标。

PowerShell UTF-8：

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
```

## 后端

目录：`backend/`

常用命令：

```powershell
go test -tags=unit ./...
go test -tags=integration ./...
go test ./...
golangci-lint run ./...
go generate ./ent
go build -ldflags="-s -w" -trimpath -o bin/server ./cmd/server
```

说明：

- 改 Ent schema 后必须运行 `go generate ./ent`，并提交生成文件。
- Windows 没有 `make` 时，直接执行 `backend/Makefile` 中的原始 Go 命令。
- 集成测试可能依赖 PostgreSQL、Redis 或 Testcontainers；失败时先区分环境问题和代码问题。

## 前端

目录：`frontend/`

常用命令：

```powershell
pnpm.cmd install
pnpm.cmd run dev
pnpm.cmd run build
pnpm.cmd run typecheck
pnpm.cmd run lint:check
pnpm.cmd run test:run
```

按文件跑 Vitest：

```powershell
pnpm.cmd exec vitest run src/__tests__/public-pages.spec.ts
```

如果当前环境实际使用 npm 脚本且依赖已安装，也可临时跑：

```powershell
npm.cmd run test:run -- src/__tests__/public-pages.spec.ts
npm.cmd run build
```

注意：

- CI 使用 pnpm 和 `pnpm-lock.yaml`，新增依赖后必须更新锁文件。
- `npm` 与 `pnpm` 混用可能导致 `node_modules` 冲突；出现 EPERM 或奇怪解析问题时优先清理后用 pnpm 重装。
- `npm.ps1` 被策略拦截时用 `npm.cmd`。

## 视觉验证

- 公共页常用本地入口：`/home`、`/tutorial`、`/models`。
- 本仓库本地前端预览只使用固定端口 `http://127.0.0.1:62080/`。
- 不要自行启动其他 `sub2api` 前端预览端口；如果 `62080` 未运行，先告知用户并等待明确要求后再启动。
- 不要为了前端预览自行启动 `sub2api` Docker/后端/数据库服务；需要后端能力时先说明当前依赖和影响，等用户确认。
- 前端 UI 改动完成后建议用浏览器或 Playwright 看桌面与移动端，特别是公共页、控制台侧栏、弹窗和表格页。

## 提交前最小检查

- `git diff --check`
- 后端改动：对应 `go test`，必要时补 `golangci-lint run ./...`
- 前端改动：对应 Vitest，必要时补 `typecheck` / `build`
- 改 package 依赖：确认 `pnpm-lock.yaml`
- 改 Ent schema：确认生成代码
- 改跨端 API contract：同时检查后端 DTO/路由、前端 API client、类型和测试

## 当前高频验证入口

- 账号共享、容量池、排行榜或用户侧展示改动：

```powershell
cd frontend
npm.cmd run test:run -- ChannelStatusView.capacityPools
npm.cmd run typecheck
npm.cmd run build
```

- 后端共享池、调度过滤、repository 查询改动：

```powershell
cd backend
go test ./internal/service/...
go test ./internal/repository/...
```

- 双仓库聊天生图链路、COS、launch/redeem 相关改动：
  这类改动通常要同时回看 `F:/java/chatgpt2api` 的 `go test ./...`、`corepack.cmd pnpm --dir web lint`、`corepack.cmd pnpm --dir web build`，不要只验证 `sub2api` 单仓库。

- Studio Bridge 后台配置、自修复、session-probe、sidebar launch 或使用记录空表相关改动：

```powershell
cd backend
go test ./internal/service ./internal/server
go test -tags=integration ./internal/repository -run "TestStudioBridgeRepository" -count=1

cd ../frontend
npm.cmd run test:run -- public-smoke
npm.cmd run build
```

  如本地 62080/8081 预览已可用，再补最小人工检查：
  - 打开 `http://127.0.0.1:62080/chat-images`
  - 确认能跳到落叶AI `/image`
  - 确认网络里 `POST /api/v1/user/studio-bridge/launch`、落叶侧 `/auth/sub2api/launch`、`redeem/user-summary` 均成功
  - 确认 iframe 只请求 `/studio-bridge/session-probe`，没有 `frame-ancestors 'none'` / CSP 报错

- 首充福利、福利页、充值页、注册 IP 或管理员用户列表相关改动：

```powershell
cd backend
go test ./internal/service ./internal/handler ./internal/repository

cd ../frontend
npm.cmd run test:run -- WelfareView.first-recharge SettingsView.first-recharge UsersView
npm.cmd run build
```

- 可配置充值套餐、支付恢复、支付兑现或用户支付页套餐展示相关改动：

```powershell
cd backend
go test ./internal/service ./internal/handler -run "TestPayment|TestRecharge|TestSetting" -count=1

cd ../frontend
npm.cmd run test:run -- PaymentView SettingsView paymentFlow paymentWechatResume
npm.cmd run build
```

- 用户注册 IP / 最近登录 IP、后台用户画像或注册来源排查相关改动：

```powershell
cd backend
go test ./internal/service ./internal/handler ./internal/repository -run "TestAuth|TestUser" -count=1

cd ../frontend
npm.cmd run test:run -- UsersView
npm.cmd run build
```

- 默认 API key、默认分组、route groups 权限或 Studio Bridge 默认 key 路由相关改动：

```powershell
cd backend
go test ./internal/service -run "TestAPIKeyService|TestStudioBridge" -count=1
go test ./internal/repository -run "TestStudioBridgeRepository" -count=1
```

- 上游 `v0.1.137` 小步合成、安全/兼容/计费兜底、thinking 协议或 OpenAI quota/reset 相关改动：

```powershell
cd backend
go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback|Test.*GenerateSessionHash|TestParseGatewayRequest|TestOpenAIQuota" -count=1
go test -tags=unit ./internal/handler -run "TestDetectInterceptType_MaxTokensOneHaiku|TestSendMockInterceptResponse_MaxTokensOneHaiku" -count=1
go test -tags=unit ./internal/handler/admin -run "TestOpenAIOAuthHandler.*Quota" -count=1
go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionUsesGenericMessageForBlacklistDenial" -count=1
go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1
go test ./internal/pkg/apicompat -count=1

cd ../frontend
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts"
```

## 近期稳定结论

- 2026-05-16~2026-05-17 的高频改动面已经从早期 `/chat-images` 跳转闭环，转向账号共享展示、容量池聚合展示、排行榜文案和 cockpit 导入。
- 容量池相关展示验证至少要覆盖：
  - OpenAI Free/Plus/Pro/Team 是否按套餐聚合，而不是按 display name 散开。
  - 只展示有账号的池子/分组。
  - 剩余百分比、`5h`/`7d` 窗口和 i18n 文案是否正确。
- 本地 fake 演示账号只用于页面预览，生产调度查询必须排除；涉及该逻辑时，要同步检查后端过滤和前端展示，不要只看 UI。
- 2026-06-10~2026-06-11 的高频验证面已经进一步前移到 Studio Bridge 本地配置自修复、动态默认 group、session-probe iframe/CSP、使用记录 null-safe 渲染、首充福利 bonus 和 sidebar launch。
- 当前本地 Studio Bridge smoke 不再只验证 `/studio-bridge/launch` 是否 200；更有价值的检查是 launch/redeem/user-summary 是否通、`session-probe` 是否被正确嵌入、以及是否仍能稳定进入落叶AI `/image`。
- 如果改动触达福利、充值或用户治理，不要只跑支付或用户单边测试；首充福利 bonus、注册 IP 和福利页已经进入同一稳定后台面。
- 如果改动触达可配置充值套餐，不要只看支付成功回调；至少同时确认后台设置、用户支付页套餐展示、支付恢复和兑现逻辑。
- 如果改动触达用户 IP 字段，不要只看数据库 migration 或后台表格；至少同时确认注册/登录链路写入、DTO 映射和用户列表展示没有脱节。
- 如果改动触达默认 API key / route groups / Studio Bridge 默认分组，不要只看前端设置页；要同时确认普通更新路径的 route groups 权限校验没有被遗漏。
- 2026-06-17 的高频验证面又补进了上游 `v0.1.137` 小步合成：
  - `form-data@4.0.6` 锁定、token refresh 不可重试错误、zstd、SSE `event:error` failover、thinking 过滤、Responses sticky hash、OpenAI `/responses` probe 和 OpenAI quota/reset 都已有定向测试入口。
  - 这类改动当前默认按“低风险 patch + 定向回归”验证，不按“整仓全量通过后再判断”理解。
  - 如果需要复跑前端 Vitest，优先用 `corepack.cmd pnpm --dir frontend exec vitest run ...` 或 `npm.cmd run test:run -- --pool=threads --poolOptions.threads.singleThread=true`；不要再用 Jest 风格 `--runInBand`。
- 2026-06-17 的上游 S15-S17 Sprint 都显式保护本地定制：验证这些 Sprint 本身时除了测试本身，还要看 `git diff --check`、denied-path audit 和 lockfile scan，确认没有误碰 Ent/migrations/VERSION、Studio Bridge、Canvas、支付页、公共页或模型市场。
- 但当前 `main` 已在后续统一 API Key / APIMart 图片网关 / 前端导航与设置页合并中触达 `wire_gen.go`、Studio Bridge repo、公共页、模型市场、`KeysView` 和 `SettingsView`。评估 `origin/main..HEAD` 时必须列出真实触达路径和对应验证，不能沿用 S15-S17 的 `NO_DENIED_PATHS`。

## 已知验证噪声

- 当前任务记录中出现过 Vite chunk / Node `DEP0190` 警告；如果 build 通过且警告为既有问题，记录为残余风险即可。
- 当前任务记录中出现过 `PaymentView.vue` 并行改动导致 `typecheck` 大量模板引用缺失；遇到时先确认是否属于当前任务改动。
- 本地 browser smoke 经常会把问题表现成“launch 成功但页面里没有余额/会话不同步”；这类情况先排查 `session-probe` iframe、CSP `frame-ancestors` 和 parent origin，而不是先判定 redeem 或余额接口失效。
- `npm.cmd run test:run -- --runInBand` 在当前 Vitest 环境下会因为不支持 Jest 参数而失败；这属于命令噪声，不应误判为产品回归。
- 前端全量 Vitest 目前仍可能被 Studio/Canvas/导航/支付等既有产品面失败污染；如果本轮只是上游小步合成，优先看定向测试、QA 报告和 denied-path audit。若本轮本来就是产品合并或 UI 合并，则不要用 denied-path audit 掩盖真实触达范围。
