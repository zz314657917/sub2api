### PASS: upstream-v027-remaining

# 主控复核与验证记录

日期：2026-09-22。基线 `d6e6c34718e6eee6388391b346f40e1492e81d57`，隔离工作树 `E:/codex-worktrees/sub2api/upstream-v027-port`。

## 实际执行

- 主控逐项读取后端与前端业务 diff、新增 helper 和关键测试，核对各批允许路径；未发现业务文件越界。
- `go test ./internal/pkg/apicompat ./internal/service ./internal/handler -run 'TestV027|TestFetchCodexModelsManifest|TestOpenAIGatewayService_BindHTTPResponseAccount|CN.*403|403.*CN' -count=1`：三个包通过。
- `go build ./...`：退出 0。
- `git diff HEAD --check`：退出 0；无未合并路径；暂存区为空。
- 独立 UI QA 完整执行合同 14 个 Vitest 文件，24 个用例通过；vue-tsc 通过。
- 私有构建环境使用原锁文件安装真实依赖，源码及 manifest 哈希一致，生产构建通过。详见 `outputs/upstream-v027-ui/build-report.md`。
- 独立 Edge profile 的真实组件浏览器 smoke 通过。主控查看桌面与 390px 截图；临时 harness/config/stub 已删除。清理证明见 `outputs/upstream-v027-ui/runtime-report.md`。

## 审查导致的修正

- CN 配额冷却最初遗漏进程内调度通知；修复为成功持久化后通知，双写失败不通知，独立 QA 复验通过。
- 前端测试补齐剪贴板 throw 后清理、订单从第 3 页切筛选回第 1 页、真实 TOTP 错误提取、退款等于/超过余额边界。
- Gemini 报告最初高估 identity mapping 覆盖；已补精确同名映射不被思考变体覆盖的测试，独立最终复验 PASS。

## 交付边界

没有提交、推送本轮改动，没有合并 main、部署或操作共享数据库/容器。未验证真实付费供应商、支付、Redis 和数据库集成。本轮本地 fake HTTP/SSE 与组件 smoke 不能替代这些环境的验证。
`docs/workflow` 新增合同和报告受仓库忽略规则影响，文件已保存在工作树内；日后提交需显式列入，禁止将 outputs 或共享依赖 junction 整体加入提交。
