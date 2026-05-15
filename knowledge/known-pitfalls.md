# 已知坑点

最后更新：2026-05-13

## 搜索与本地目录

- `.gocache/` 里有 Go 依赖缓存，递归搜索会很慢。
- `.venv/`、`output/`、`frontend/node_modules/` 也应从常规源码搜索中排除。
- 推荐搜索：

```powershell
rg "关键字" -g "!/.git/**" -g "!/.gocache/**" -g "!/.venv/**" -g "!/output/**" -g "!/frontend/node_modules/**"
```

## 文档与 Git 忽略

- `.gitignore` 当前忽略 `docs/*`，但例外允许 `docs/PAYMENT.md`、`docs/PAYMENT_CN.md`、`docs/ADMIN_PAYMENT_INTEGRATION_API.md`。
- `.gitignore` 当前忽略 `knowledge/tasks/`，所以任务快照和时间轴默认是本地协作状态，不一定会提交。
- `knowledge/` 根下的长期知识页未被忽略，可以作为项目知识库入口。

## 编码

- PowerShell 读写中文文档时显式 UTF-8。
- 看到中文变成无意义字符组合，或出现 Unicode replacement character 时，先按 UTF-8 重读，不要基于乱码内容判断。

## 前端依赖

- 项目开发指南强调前端包管理使用 pnpm。
- CI 使用 `pnpm install --frozen-lockfile`，改依赖必须提交 `pnpm-lock.yaml`。
- 如果曾用 npm 安装导致 `node_modules` 冲突，清理后用 pnpm 重装。
- Windows 上 `npm.ps1` 可能被执行策略拦截，使用 `npm.cmd` 或 `pnpm.cmd`。

## 数据库与 Windows 环境

- 本地 PostgreSQL 开发配置见 `DEV_GUIDE.md`。
- Windows 上 psql 连 `localhost` 可能先走 IPv6，建议使用 `127.0.0.1`。
- PowerShell 会解释 bcrypt hash 里的 `$`，执行包含 hash 的 SQL 时优先写入 SQL 文件再 `psql -f`。
- `psql -f` 对中文路径可能不稳定，必要时复制到纯英文路径。

## Go / Ent

- 改 Ent schema 后必须 `go generate ./ent`。
- Go interface 新增方法后，所有测试 stub/mock 都要补方法。
- Windows 没有 make 时直接使用 Makefile 中的底层命令。

## 业务坑

- OpenAI 账号和 Gemini/Antigravity 等不同 provider 混合批量修改时，模型白名单或映射容易被跨平台策略覆盖。
- Codex 或 OpenAI 新模型更新快于默认映射时，临时透传映射可能比改核心映射表风险更低。
- 计费、余额、共享池、排行榜奖励属于额度/资金相关链路，必须有明确测试或人工验证记录。
- 公共页视觉调整容易和控制台侧栏/i18n 改动混在同一工作区，提交前要按主题复核 diff。
