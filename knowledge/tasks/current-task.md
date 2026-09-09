# 当前任务快照

最后更新：2026-09-09

## 使用文档配置更新（2026-09-08，独立小修）

- 后续精简：首页改为获取密钥、填写配置、启动验证三步；移除重复信息卡、桌面端说明及“首次必做”标记，安装、文件夹查找、API 示例、排障默认折叠。完整配置与后台自定义地址仍保留。
- 精简验收：TutorialView 9/9、frontend typecheck 通过；Playwright 桌面 1440x1000、手机 390x844 截图、折叠开关及 Claude 切换通过，无横向溢出。截图位于 outputs/tutorial-compact-*.png。
- 浏览器 session tutorial-compact-0908，daemon PID 52352，Chrome PID 54384，独立 profile playwright_chromiumdev_profile-K0zdjZ；已 close 且精确归属进程为零。其他任务浏览器未操作。

- 按用户提供的模板更新 Codex：首页与完整教程共用 OpenAI / 3Z API 配置，主模型 gpt-6-astra，review_model gpt-5.5，默认地址 https://ai.3zapi.com；附加参数按用户模板保留，未完成真实客户端兼容验证。
- auth.json 示例只包含 JSON，启动命令移至说明；后台默认配置与定向测试同步。新增 migration 240，只替换匹配旧默认值的教程内容及已保存的 Codex 教程地址，保留其他自定义内容。
- 已执行：TutorialView Vitest 9/9、`public-pages` Vitest 10/10、frontend typecheck、Go TestQuickstart、git diff --check 均通过。`public-pages` 最初保留旧 `.top/v1` Codex 文案断言，已仅改为 `.com` 根地址的当前契约后重测通过。
- 浏览器验收：任务会话 `codex-tutorial-20260908`、隔离 in-memory profile、初始 PID `53260`；在 `1440x900` 与 `390x844` 打开 `/tutorial`。配置区显示 `gpt-6-astra`、`review_model = "gpt-5.5"`、3Z API 根地址及独立 auth.json；移动端 `document/body scrollWidth == clientWidth == 390`。截图为 `output/playwright/codex-tutorial-20260908/.playwright-cli/page-2026-09-08T10-10-25-348Z.png` 和 `page-2026-09-08T10-11-20-345Z.png`。会话已 close，PID 已归零。
- Vite PID `54088` 在验收前提供 HTTP 200，完成后按精确 PID 结束，`127.0.0.1:62080` 已拒绝连接；不操作其他任务资源。
- 本地提交：d0820d4b9（fix(tutorial): refresh Codex config and simplify quickstart），仅含 8 个教程配置、页面、迁移及测试文件；提交前 19 项前端测试与暂存区 diff 检查通过。混有其他任务记录的本文件未纳入该提交。
- 未执行：真实 PostgreSQL 迁移、真实客户端请求、推送和部署。原有 Sprint 状态及其他脏改保留；下方为原有恢复快照。

## 当前恢复快照（优先于下方历史记录）

- 当前 Sprint 为 `upstream-v0200-gpt6-astra-s297`，phase 为 `qa`；本轮补 S296/S297 独立 Terra 验收及状态记录，不推进发布。
- 检查基线 `42ab6da534c2a26c4d30176f9228f76e51bae839`；相对本地缓存 `origin/main` 领先 51 个提交，未刷新远端。下方旧 HEAD、领先数和容器健康记录仅代表历史时点。
- S296/S297 独立 Terra QA 已完成主体检查：后端定向 25/28 项、前端 25/28 项，以及共享 server 编译、后端 build、前端 typecheck/build 均通过。两份 QA 报告已落入 `docs/workflow/qa-reports/`；运行态/API 未验证，总体仍为 `BLOCKED`。测试基于含既有脏改的当前工作树，不代表干净提交快照。
- S296 格式命令引用 4 个不存在路径，已更正为当前 owner 的只读检查；独立修订审查、精确 15 路径格式检查和 handler/admin 门禁均通过。handler 包无匹配测试，仅确认编译通过；未重复已通过的公共构建。
- S295 原合同引用的是未并入当前分支的 Composite 功能链，而非本地误删文件。已获批准的 S295-P1 仅实现 Ent schema/migration、repository、后台 API 与静态 preview；定向 service/handler/routes、server 编译和 `go build ./...` 均通过，未执行迁移也未接入 Gateway。`go generate ./cmd/server` 被两个既有 provider 缺失阻断，完整 Ent 生成因 Windows 锁风险未重放；独立 Terra QA 未取得，P1 QA 为 `BLOCKED`。resolver、context 和 gateway dispatch 继续留给 P2/P3；主 Sprint 仍为 S297 QA，不能改 `status.md` 宣称 S295 完成。
- S293 R4 仍阻断：专属 PostgreSQL DSN 未设置，Docker Linux engine pipe 不可用，未发现本地 5432/6379 监听。Ollama 已安装 `qwen3guard-gen:0.6b-q4km`，但尚未执行本轮 Guard smoke；不能继续沿用“仅安装 embedding”的旧环境结论。
- S293 R4 资源复核（2026-09-09）：发现 `sub2api-s293-test-postgres`、`sub2api-s293-test-redis` 和配套测试服务均 healthy，专属网络/卷归属明确；但 Windows 宿主访问 Docker 内网地址 `172.18.0.3:5432`/`172.18.0.2:6379` 均超时，集成测试在连接初始化阶段失败，未执行 migration 239 或写入数据。需补宿主端口映射或同网运行测试；Guard 仍未做 runtime smoke。
- 本轮不修改已有业务脏改，不 push、部署、更新容器或操作数据库；S293 真实 PG/Redis/双实例/认证浏览器验收仍需具备专属资源。

## 背景

- 用户要求持续筛选并选择性合入上游改动；始终禁止整体 merge、rebase 与 cherry-pick。
- 已完成 S281--S285 的 429/限流队列以及 S287、S289；当前从刷新后的 v0.2.0 候选中继续挑选独立小批。

## 当前目标

- 当前 Sprint：`upstream-v0200-claude-fable-5-1-s294`。
- Workflow phase：`qa`，独立 QA 已解除编译/类型检查阻断；仍保留已知 unit fixture 漂移记录。
- S294 已完成 Fable 5.1 的行为级本地适配；Fable 定向 backend 测试和前端 Vitest 47/47 通过，`go build ./...`、前端 typecheck 和 production build 现已通过。`go test -tags unit ./...` 仍有已知 repository fixture/test drift。
- S294 证据：`docs/workflow/worker-results/upstream-v0200-claude-fable-5-1-s294-result.md`、`docs/workflow/qa-reports/upstream-v0200-claude-fable-5-1-s294-qa.md`。
- 本地产品已按行为边界分批提交：Prompt Audit `3281801e1`、Fable catalogs/usage `74a5d2c35`、Pixel Cafe reservation `3ff5ae8f0`、Fable pricing/rate-limit `a92afcde3`；未 push。
- S290 已按修订合同完成独立 QA 和最终裁决；四个前端文件已提交为 `7cacdbab1`。S266 内容审核的产品提交已在主线等价存在，其任务、结果和 QA 证据已通过 `12e52216e` 合并回主线谱系；`origin/main` 已同步至 `5b95e68dd`。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## S292 已完成

- 新增可选 `prompt_audit_config.rules`，与现有配置版本、Redis 失效通知和原子快照共用热加载路径。
- `Safe` 只能 `Allow`，`Controversial` 只能 `Warn/Block`，`Unsafe` 固定 `Block`；分类规则只能升级到 `Warn/Block`，不能降级放行。
- 同步 `GuardEvaluator` 与异步 `Runner` 均在聚合前应用同一规则；定向 securityaudit、handler/admin、service 测试和 `go build ./...` 通过；无前端规则编辑器、真实 Guard provider smoke 或推送；产品代码已本地提交为 `cc4acbcac`。

## S293 规划

- 已写入 `docs/workflow/plans/prompt-audit-policy-matrix-s293.md`，拆为 A 纯求值核心、B 持久化发布回滚、C 管理 API/前端、D shadow/灰度/最终 QA。
- 总合同草稿保留在 `docs/workflow/tasks/prompt-audit-policy-matrix-s293.md`；A-D 均已完成本地实现和对应 QA 记录。

## S293 修复进行中（2026-09-05）

- 本轮续修已补严格草稿 CAS、并发快照版本不倒退、历史 ID/裁剪校验、分块规则归因，以及保存/发布/回滚的前端版本续用和迟到响应保护；管理页 Shadow 使用真实本地解析基线，分别比较 active/candidate，兼容旧接口。
- 最新代码级证据：独立 Terra 复核完成；审计/路由/middleware/migrations 定向测试、后端构建、前端 43/43 和 production build 通过。全量 Go 仍有既有 repository 32/34 fixture 失败，另有 usagestats 测试程序启动被系统拒绝。
- 新 PG 往返测试和 migration 239 夹具已准备，因专属 DSN 未设置明确 SKIP。任务浏览器预览服务启动被工具策略拒绝，未启动浏览器，无任务自有运行资源遗留。
- 当前结论仍为发布 `BLOCKED`；下一步仅在确认专属 PG/Redis/Guard 测试资源后执行 R4，不改共享环境。最新证据：`docs/workflow/qa-reports/prompt-audit-policy-matrix-s293-remediation-continuation-qa.md`。以下为本轮前的修复记录。

- 复核确认 F1-F8 后已创建修复计划 `docs/workflow/plans/prompt-audit-policy-matrix-s293-remediation.md`。
- R1 已修复事件 INSERT 参数错位、Unsafe/未知分类安全底线、策略输入边界、规则归因与多分块聚合；R2 已加入草稿 `BaseConfigVersion` 发布冲突校验；R3 已分离本地策略编辑与已保存草稿，并补充规则 safety/categories/groups/models/providers 编辑；代码已本地提交为 `3281801e1`。
- R1/R2/R3 合同、审查和 worker 结果文档位于 `docs/workflow/tasks/`、`docs/workflow/contract-reviews/`、`docs/workflow/worker-results/`；当前只完成代码级修复和定向测试，尚未完成 R4。
- 已验证：securityaudit/routes/middleware/migrations 定向测试、`go build ./...`、Prompt Audit Vitest 32/32、`vue-tsc --noEmit`、前端 production build、仓库外五项原始复现均通过。
- 未验证：专属 PostgreSQL 事件读回与并发事务、Redis 多实例收敛、认证浏览器真实后端流程、Qwen3Guard `0.6b` runtime。不得据此宣布 S293 可发布。

## S293-A 已完成

- 策略支持 `defaults + rules[] + priority + safety/category/group/model/provider scope`。
- 兼容 S292 Map；同步和异步路径共用求值逻辑；结果保留 `matched_rule_id`。
- `securityaudit`、`handler/admin`、`service` 测试和 `go build ./...` 通过；未做真实 provider/browser smoke。

## S293-B/C/D 已完成

- B 增加最多 20 个策略历史版本、草稿 CAS、预览、发布、回滚，并复用 settings、PostgreSQL advisory lock 和 Redis 热加载；旧 `PUT /admin/prompt-audit/config` 立即保存行为保留。
- C 增加后台策略 API 和 Prompt Audit 规则编辑器，锁定 Safe/Unsafe 安全底线，支持分类动作、自定义规则、OWASP 标签、草稿/预览/发布/历史/回滚。
- D 增加 copy-on-write `ShadowEvaluate`、shadow 次数/差异指标，以及事件中的 `matched_rule_id`、`owasp_tags`；OWASP 仅作解释标签，不可降低风险。
- 本地后端定向测试、`go build ./...`、Prompt Audit Vitest 31/31、`vue-tsc --noEmit`、前端 production build、`git diff --check` 和未合并索引检查通过。
- 新迁移 `backend/migrations/239_prompt_audit_policy_explanations.sql` 尚未执行；运行态探测发现本机仅有 `qwen3-embedding:0.6b`，PostgreSQL/Redis 和 Docker 不可用。S293 已完成任务专用 Playwright 浏览器 smoke（session `s293-policy-20260904b`，页面使用本地 mock API）：验证草稿预览、发布 toast/版本更新、历史版本回滚确认与 toast，并在 `1440x900`、`390x844` 检查无横向溢出；截图位于 `output/playwright/s293-policy-1440x900.png` 和 `output/playwright/s293-policy-390x844.png`。session、专用 profile/daemon 与 Vite 已清理；真实 PostgreSQL/Redis 多实例、认证浏览器生命周期和 Qwen3Guard `0.6b` runtime smoke 仍未验证。
- 2026-09-04 19:46 续验：重新执行 `go test ./internal/securityaudit ./internal/server/routes ./internal/server/middleware ./migrations -count=1`、Prompt Audit Vitest 32/32、`npm.cmd run typecheck` 和 `npm.cmd run build` 均通过；任务专用浏览器使用 mock API 再次验证试运行/预览，桌面与移动端 `scrollWidth == clientWidth`，session/Vite/专用浏览器资源已清理。

## 本次已完成

- S281--S285 已分别提交为 `c886cdcac`、`f48b4b77f`、`bb3d3bca6`、`65bf61f5a`、`b686353c3`。
- S287 和 S289 已提交为 `e6845b4ea`、`6050139a3`，均有独立 QA PASS。
- 刷新后 `upstream/main=b1748c4ea`；本地与上游历史仍大幅分叉，继续行为级筛选。

## S294 Fable 5.1

- 上游 `b3f796972`、`32ac921f2`、`34b8bf1a6` 未整体合入；已按合同行为级移植到本地模型目录、OAuth、计费、7d_oi 限流/用量和前端 owner。
- Fable 5.1 已加入 Claude/Antigravity/Bedrock catalog/mapping、OpenCode 配置和 fallback pricing；7d_oi 只写入 `claude-fable-5` 家族级 model limit，不误伤同账号其他 Anthropic 模型。
- Antigravity Claude 用量汇总已补 `claude-fable-5-1`；新增测试覆盖该配额聚合。
- focused QA PASS；full acceptance `BLOCKED`，不标记 Sprint PASS，不 push、不部署、不更新容器、不操作数据库或真实 provider。

## 已确认事实

- `343858021` 在本地已等价，不单独合入；`05ea883e2` 依赖缺失的 Group schema/Ent 状态。
- `3510aa22b` 的 reasoning-effort 按模型范围功能依赖至少 43 文件的策略基础、迁移和 Ent，不是小批，暂缓。
- `1a33dc8cc` 的四处布局意图仍在本地有效，但其 patch 因组件拓扑分叉失败；可作为 S290 的四文件手工适配。
- 最新全量 `go test ./...` 除既有 repository fixture 漂移外均通过；失败为 `account_repo_upstream_billing_probe_update_test.go:559` 的 32/34 列不一致，与本轮前端及 S287/S289 无交集。

## 保护边界

- 保留 `backend/internal/pkg/apicompat/*.go` 的既有脏改。
- 保留 `backend/internal/service/admin_service.go` 的 Pixel Cafe hunk。
- 保留 `frontend/pnpm-lock.yaml`、`frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue` 与 `outputs/**`。
- Controller 冻结的受保护 dirty diff hash：`0e467987fd7aec5fc451983bdb8f8216f97ba69c`。

## 已验证

- S290 仅修改合同列出的四个前端文件；受保护脏改 hash 仍为
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，无冲突索引。
- 任务专属 Chrome profile 已在本机 `/admin/groups` 打开创建、编辑分组的定价表单；六项默认 Token 价格在 `1440x900` 与 `390x844` 无文档或弹窗横向溢出。两张表单均取消，未保存或改动共享数据，且 session/profile/Vite 已清理。
- 定向 Vitest 2/2、typecheck、production build 和四文件 `git diff --check` 均通过。截图位于
  `E:/codex-runtime/pge/sub2api/s290/browser-smoke-20260902-retest/artifacts`。

## 待验证点

- 启用渠道定价入口的 `IntervalRow` 仍未跑真实浏览器 smoke；后续单独打开该入口并在桌面与移动视口检查布局，不能通过修改分组弹窗的 `hide-token-intervals=true` 来规避。
- `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266` 是已注销的任务依赖残留；如需释放空间，先验证它仍不在 `git worktree list` 且无关联进程，再删除该精确目录。

## 当前结论

- 2026-09-05 本地容器更新已完成：`sub2api:local` 现指向 `sha256:6e21f30abb0599303d5b56cb943f217990fd4091501c8e94b31e9547cf75a51c`，容器 `sub2api` 在 `127.0.0.1:62580` 为 `healthy`，`/health` 返回 200，重启次数为 0；239 审计策略解释迁移已在专属本地 PostgreSQL 落库。回滚镜像保留为 `sub2api:rollback-20260905-before-s293-r1`（旧 S290 镜像）。本次使用本机 Linux ELF 二进制覆盖运行时层，未修改受保护的 `frontend/pnpm-lock.yaml`；完整多阶段 Docker 构建因该既有 lockfile 与 `package.json` overrides 不一致而未采用。
- 当前本地 `HEAD=a92afcde3`，相对 `origin/main=43b38cc32` ahead 16；本轮不 push。
- S290 在修订后的可达浏览器范围内为 `PASS`，产品提交为 `7cacdbab1`；S266 证据谱系合并与 S290/S266 定向回归均通过，已推送 `origin/main@5b95e68dd`。分组弹窗保留既有 `hide-token-intervals=true`，因此启用渠道定价入口的 `IntervalRow` 真实浏览器 smoke 明确拆为后续任务。
- Kimi native Responses、Claude Fable 5.1 和 reasoning-effort scope 均为独立大功能，不与 S290 混合。

## S291-A 合同

- 上游最新代理归因链拆为 S291-A/B/C；当前只进入 S291-A 核心事件与队列边界。
- 合同和独立审查均为 PASS：
  `docs/workflow/tasks/upstream-v0200-ops-proxy-attribution-s291a.md`、
  `docs/workflow/contract-reviews/upstream-v0200-ops-proxy-attribution-s291a-review.md`。
- S291-A 允许修改 Ops 事件/队列核心及对应测试；所有网关调用点留给后续合同。
- S291-A build 已完成：定向测试、完整 `internal/service`、`go build ./...`、
  `git diff --check` 和未合并索引检查均通过；结果见
  `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291a-result.md`。
- S291-B 合同和独立审查均为 PASS，当前只补本地 Gateway/Gemini HTTP 调用点；
  OpenAI/WS/provider 剩余调用点延后至 S291-C。
- S291-B build 已完成：Gateway 单体 22 个事件点与 Gemini 兼容入口补齐代理归因，
  定向测试和 `go build ./...` 通过；结果见
  `docs/workflow/worker-results/upstream-v0200-ops-proxy-attribution-s291b-result.md`。
- S291-C 合同和独立审查均为 PASS，当前补 OpenAI/Grok/WS 错误事件调用点。
- S291-C build 已完成：OpenAI/Grok/WS 现有生产事件点和 WS fallback unknown
  语义已覆盖，定向及完整 service 测试、构建通过；Antigravity 单体逻辑另拆 S291-D。
- S291-A 至 S291-E 已完成本地代理归因集成：核心 event/legacy/queue 边界、
  Gateway/Gemini、OpenAI/WS/Grok、Antigravity 与最终遗漏点均已覆盖。独立 Terra QA 已通过定向、
  完整 service、`go build ./...`、diff/冲突、allowlist 与 100 个生产事件构造点扫描；未推送。

## 下一步

1. S291 独立 QA 已完成；需发布时另行请求推送，不得夹带任何未提交的用户改动。
2. 若继续前端定价验收，另开任务验证启用渠道定价入口的 `IntervalRow` 桌面/移动布局，不改变分组弹窗的计费语义。

## 验证记录

- `git fetch upstream --prune`：PASS，`upstream/main=5097b31457e6dc9f49e5f5c9c72b925ce79543b3`。
- `node C:/Users/Administrator/.codex/scripts/codex-workflow.mjs pge-doctor --repo . --strict`：20/20 PASS。
- `go test ./...`：除既有 repository fixture 32/34 列漂移外，其余包 PASS；未修改工作区。
- 受保护业务脏改普通 diff hash：`0e467987fd7aec5fc451983bdb8f8216f97ba69c`，本轮前保持不变。
- S290：定向 Vitest 2/2、typecheck 与 production build 均 PASS；独立 QA 使用任务专属 Chrome profile
  `E:\codex-runtime\pge\sub2api\s290\browser-smoke-20260902-retest\chrome-profile` 在本机 `/admin/groups`
  验收创建/编辑定价表单，`1440x900`/`390x844` 下六项默认 Token 价格均无横向溢出并保存截图。表单取消未保存；关闭后 session、task-owned browser/cliDaemon 与 5174 监听均为零，最终 verdict 为 `PASS`。
- 合并后回归：S290 布局与 S266 风险控制前端测试共 3 文件/9 项 PASS，`typecheck` 和 production build PASS；S266 的 service、handler、admin handler、repository、migration 与 server compile 定向命令均 PASS。
- 发布与整理：`git push origin main` 成功，远端从 `6050139a3` 更新为 `5b95e68dd`；S266 的三个 PGE 分支及 S280/S281 已合入分支已删除，两个干净 S266 子工作树和父工作树的 Git 注册已删除。其父目录仍留有未注册依赖文件，因宿主拒绝递归删除而未强行清理；所有脏/冲突/备份工作树均保留。
