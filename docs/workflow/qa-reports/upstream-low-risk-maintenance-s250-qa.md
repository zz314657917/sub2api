### PASS: upstream-low-risk-maintenance-s250

## 独立性

- 本报告由独立 QA 在隔离 worktree
  `E:/codex-worktrees/sub2api/upstream-low-risk-maintenance-s250-qa` 执行。
- QA 基于 `943815cd4`（并审计此前 `9a4fd1966`、`d10926da3`）复跑；未修改业务、状态、知识库、main-log 或其他文件。

## 范围与实现审计

- `9a4fd1966` 仅修改 `frontend/package.json` 与 `frontend/pnpm-lock.yaml`；直接依赖和 Mermaid 传递路径均解析为 `dompurify 3.4.14`，并以 override 阻止低版本解析。
- `d10926da3` 仅修改 Ops collector 及其默认标签测试。无具体 cgroup 内存上限时返回完整宿主机 used/total/percent 三元组；存在具体限制时返回完整 cgroup 三元组，避免混用来源。
- `943815cd4` 仅修改管理员用户编辑弹窗、其测试和中英文用户 locale。`0` 可保存为 unlimited；负数和小数均被拒绝，并显示 `0 = unlimited` 提示。
- 三个候选的业务 owner 与任务 allowlist 一致；`7e326ac28..HEAD` 中额外的 workflow contract/status/main-log/result 文件为候选业务提交之前的流程证据。当前 QA 新增仅本报告。

## 已执行命令与结果

| 命令 | 结果 |
| --- | --- |
| `corepack pnpm --dir frontend install --frozen-lockfile --ignore-scripts` | PASS；lockfile up to date，未再生成。 |
| `corepack pnpm --dir frontend why dompurify` | PASS；direct、Mermaid 与 `@types/dompurify` 均为 `3.4.14`。 |
| `corepack pnpm --dir frontend exec vitest run src/components/admin/user/__tests__/UserEditModal.spec.ts` | PASS；1 文件、3 tests。 |
| `corepack pnpm --dir frontend run typecheck` | PASS。 |
| `corepack pnpm --dir frontend run build` | PASS；Vite build 20.11s。 |
| `go test ./internal/service -run "TestResolveMemoryStats" -count=10` | PASS；5.691s。 |
| `go test ./internal/service -count=1` | PASS；64.630s。 |
| `go test ./cmd/server -run '^$' -count=1` | PASS；5.745s compile-only。 |
| `gofmt -d`（两项 Ops owner） | PASS；无输出。 |
| `git diff --check`、冲突标记扫描、`git ls-files -u`、暂存/工作区 diff 检查 | PASS；无格式错误、冲突标记、未合并索引或未经授权的工作区修改。 |

## 风险与未执行项

- 生产构建有既存 Browserslist 过期、动态 import 与 chunk-size 警告，但退出码为 0；本 Sprint 未触及相关构建拓扑。
- 未执行真实 provider、部署、容器、共享/生产数据库、浏览器会话或 push；均在契约禁止范围内。

## 结论

- 独立 QA 通过；三项 S250 低风险维护切片满足契约验收边界，可进入 Controller 最终裁决/集成门禁。
