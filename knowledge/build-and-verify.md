# 构建与验证

最后更新：2026-05-13

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
- 当前任务快照中记录过固定预览端口 `http://127.0.0.1:62080/`，但实际端口以正在运行的 Vite 输出为准。
- 前端 UI 改动完成后建议用浏览器或 Playwright 看桌面与移动端，特别是公共页、控制台侧栏、弹窗和表格页。

## 提交前最小检查

- `git diff --check`
- 后端改动：对应 `go test`，必要时补 `golangci-lint run ./...`
- 前端改动：对应 Vitest，必要时补 `typecheck` / `build`
- 改 package 依赖：确认 `pnpm-lock.yaml`
- 改 Ent schema：确认生成代码
- 改跨端 API contract：同时检查后端 DTO/路由、前端 API client、类型和测试

## 已知验证噪声

- 当前任务记录中出现过 Vite chunk / Node `DEP0190` 警告；如果 build 通过且警告为既有问题，记录为残余风险即可。
- 当前任务记录中出现过 `PaymentView.vue` 并行改动导致 `typecheck` 大量模板引用缺失；遇到时先确认是否属于当前任务改动。
