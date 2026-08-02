### PASS: usage-admin-frontend-s137

## Findings

- 未发现未解决的 contract 违例、功能回归或安全问题。
- 最终复审修正了两处问题：原 Ops 弹窗中的共享错误表格改为默认不可排序，只有管理员 Usage 错误 Tab 显式开启服务端排序；管理员错误 handler 对 `category`、`sort_by`、`sort_order` 增加固定白名单校验。
- 改动仅落在 S137 allowlist；未触及 schema、migration、route、permission、authentication、deployment、container 或 production configuration。

## Executed Checks

- `go test ./internal/handler/admin -run 'Test(GetUserBreakdown|OpsErrorHandler)' -count=1`：PASS。
- `go test ./internal/repository -run 'TestGetUserBreakdownStats' -count=1`：PASS。
- 4 个聚焦 Vitest 文件：4/4 files、18/18 tests PASS。
- `corepack.cmd pnpm --dir frontend run typecheck`：PASS。
- changed-file ESLint：PASS。
- `corepack.cmd pnpm --dir frontend run build`：PASS，1109 modules transformed。
- Playwright mock desktop `1440x1000`：三 Tab、错误 400 筛选、状态排序、第 2 页、独立列、详情、排行输入 Token 排序、Top 100 和用户下钻 PASS；`clientWidth=scrollWidth=1440`。
- Playwright mock mobile `390x844`：错误和排行移动卡片、错误详情 PASS；`clientWidth=scrollWidth=390`。
- 请求证据包含 `status_codes=400`、`sort_by=status_code`、错误 `page=2`、排行 `sort_by=input_tokens&limit=100`，以及下钻后的 `user_id=9`；最终控制台 0 error / 0 warning。
- `git diff --check`、冲突标记、未合并索引、allowlist、禁止触面和排序白名单检查：PASS。
- Ops 原调用兼容性：组件回归证明默认无可点击排序列；Usage 调用显式开启服务端排序，热更新后的浏览器快照仍显示模型、状态码和时间排序头。

## Unverified Risks

- 未连接真实 PostgreSQL，也未使用真实管理员登录态调用后端 API；本轮运行态证据来自聚焦 Go 测试和浏览器 mock。
- 未执行部署、容器更新、生产 migration 或生产数据量性能验证。
- production build 保留仓库既有的 Browserslist 过期、动态/静态 import、chunk size 和 Node shell deprecation 警告；本轮没有新增依赖或构建配置。

## Recommendation

- S137 可作为一个本地隔离分支提交收口。不得在当前步骤合入脏 `main`、push、部署、更新容器或执行生产 migration；后续集成前应重新检查主工作树重叠改动。
