### PASS: group-fast-20261103

task_id: group-fast-20261103
verdict: PASS
status: approved
contract_sha256: d030d97471765c44c50a9d92652af3605729da086b50e76adc4bc584799c724f
previous_pass_contract_sha256: a12680fc17e908538ac053c009626ddec94b09ca5ea4151ed9a4784cbb032238
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
reviewer: independent-contract-reviewer
reviewer_role_model: gpt-6.1-sol (Agent Matrix)
last_verified: 2026-11-03

## Focused Re-review: i18n Allowed Paths (2026-11-03)

- Verdict: PASS。新最终原始字节绑定当前 contract_sha256；原 PASS 哈希保留于 previous_pass_contract_sha256，仅作历史记录。
- 变更严格限于 Allowed Paths 的两项：`frontend/src/i18n/locales/zh/admin/overview.ts` → `frontend/src/i18n/locales/zh/admin/groups.ts`；`frontend/src/i18n/locales/en/admin/overview.ts` → `frontend/src/i18n/locales/en/admin/groups.ts`。初审未识别这两条上游路径与本地布局的差异，本次更正已核实。
- 原始字节核对：只读读取合同为严格 UTF-8，在内存中将两条新路径各一次逆替换回旧路径，不改 BOM、换行或其他内容；逆替换字节 SHA256 精确等于原 PASS 哈希 `a12680fc17e908538ac053c009626ddec94b09ca5ea4151ed9a4784cbb032238`。因此除这两条路径外，合同字节完全不变。
- 本地两个 groups.ts 均为现存文件，内容为管理分组文案；两个 admin/index.ts 均 `import groups from './groups'`，确认是实际管理端 locale 入口。未将普通用户的 locales/{zh,en}/groups.ts 混入范围。
- 产品行为、验收要求、所有其他 Allowed/Denied Paths 及角色边界不变；无必须修正项或新增阻断。本次仅更新本独立报告，未修改合同/产品、未运行产品验收或启动其他 agent。下列初审检查作为原审查记录保留。

## Findings

未发现必须修正的合同问题。PASS 仅批准下述最终合同字节进入隔离开发，不代表实现、QA、数据库迁移或部署通过。

- Contract Checked: `F:/mcplugins/sub2api/docs/workflow/tasks/group-fast-20261103.md`。多次直接对原始文件执行 SHA256，包含现有 approved/PASS 元数据，均为上述值；未修改或重新序列化合同。任何字节变化均使本报告失效，须独立重审。预置 approved 仅在本报告 PASS 后生效。
- 范围明确：默认关闭的分组布尔字段、管理表单、可信 OpenAI 请求注入及原全局策略优先级；明确排除 Free Fast、全局规则扩展、价格/计费重写、依赖变更及部署。Allowed/Denied Paths 与本地所有者对应，没有要求复制上游完整文件。
- 上游已核实：GitHub 官方仓库 PR #6443，标题 `Add group-level OpenAI Fast enforcement`，merged=true，merge_commit_sha=`a4fb58e428a0c5d12087d1de2e6d91b56b7db8b2`，merged_at=`2026-09-02T01:07:06Z`。本地 Git 对象与该 SHA、标题吻合；该提交是固定 upstream `4a4a83ee5ec86b19d3a2dd9b6471de5128963bec` 的祖先。核对了 merge 第一父提交至 merge 的实际字段、缓存及策略差异。
- 本地数据链可落地：Ent schema/generated、迁移、handler request/DTO、service inputs/domain、repository 创建/更新与 API-key 分组投影、duplicate、auth snapshot 均在 allowlist。上游 admin_group.go 在本地由 admin_service.go 承担，合同已正确适配；cloneGroupForDuplicate 当前复制整个 Group，需验证新字段保留，无需无意义重写。
- 信任链已有接口：`group.go:IsGroupContextValid` 要求 ID>0、Hydrated、非空 Platform/Status；`server/middleware/api_key_auth.go:setGroupContext` 将可信分组写入 request context 的 ctxkey.Group。可在允许修改的 service 内复用，结合实际 account.Platform 与分组平台判定，无需增加客户端策略字段或扩大 middleware 修改范围。
- HTTP 路径可覆盖：openai_gateway_service.go 同时存在 native Responses 的内联 evaluateOpenAIFastPolicy 分支与 passthrough 的 applyOpenAIFastPolicyToBody 调用；Chat/Messages 各有共用策略调用。合同明确列出这些入口，允许文件覆盖全部调用点。只改共用 helper 而漏掉 native 内联分支不满足合同。
- WS 路径可覆盖：openai_ws_forwarder.go 的 ingress 公共解析点处理首帧及后续 turn；openai_ws_v2_passthrough_adapter.go 分别处理首帧及后续 response.create，并从最终 policy 输出记录 usage tier。两者均在 allowlist；合同要求省略 tier 仍注入、其他帧不变，保留现有 session policy 快照和重连边界。
- 兼容性约束充分：未启用/不可信/不支持平台及非 OpenAI 上游保持原行为；先强制 priority，再执行原全局 filter/block；省略 update 保留字段、切换至不支持平台归零；价格和 usage ownership 不变。未知 tier 的覆盖与默认关闭负向用例均可判定。

## Executed Checks

- 已读取 AGENTS.md、status、Agent Matrix、相关 spec 与合同全文；核对 HEAD 等于 base_commit；只读检查本地相关 service、repository、middleware、前端测试入口、package.json 和 Ent generator。
- 已使用 Get-FileHash -Algorithm SHA256 检查最终合同；git show / git diff 检查固定上游差异；git merge-base --is-ancestor 核实 ancestry（退出码 0）；通过 GitHub REST API 读取 PR 元数据。
- 已确认本地 apiKeyAuthSnapshotVersion=19，旧版本检查会拒绝快照；groupFromAPIKeyAuthSnapshot 同时用于主 Group 与 RouteGroups。UpdateGroup 已调用 InvalidateAuthCacheByGroupID。合同“stale caches must not silently lose it”要求升级本地版本并覆盖这两种投影，不能照抄上游版本号或只添加 JSON 字段。
- 当前迁移最高数字编号 253，254 未占用；合同要求写入前重查，符合并发边界。固定 Go 可执行文件、已有 Vitest 入口均存在；现有 mounted group duplicate/model-pricing 和 WS ingress/passthrough 用例可复用。
- 验收可执行：backend 内固定 Go 执行 focused tests（default/unit、-count=1、匹配数非零）和 build；frontend 可用现有 Vitest/vue-tsc/ESLint/Vite。vite.config.ts 默认输出到 backend/internal/web/dist，执行时必须按合同显式 --outDir 指向 evidence；lint 必须采用无 --fix 入口。Ent 生成后逐路径核对仅新增字段，禁止依赖/生成噪声。
- 已只读查看 main-before.json 头部和 prerequisites.json；其 HEAD 与合同一致，后者登记三个未跟踪测试。没有读取 .env。完整冻结清单、候选字节及依赖 junction 的独立一致性验收仍由 Controller/QA 承担，本报告不冒充已全量验证。
- 本轮未运行产品测试、构建、Ent 生成、浏览器或迁移；仅写本独立报告，没有产品增量可作实现 diff 验收，未启动其他 agent。

## Gate Checks

- success_criteria_testable: yes
- allowed_paths_explicit: yes
- denied_paths_explicit: yes
- acceptance_commands_executable: yes (现有工具/入口可用；开发与 QA 须记录实际完整命令及计数)
- worker_model_confirmed: yes (Agent Matrix 与 contract 配置一致；实际调用身份由主控保证)
- base_commit_confirmed: yes
- openspec_traceable: not-applicable (spec_ref 与当前功能规格可追溯)

## Unverified Risks / Recommendation

- 必须修正项：无。
- 可以按当前合同开展隔离开发。主控拟拆分两个 disjoint Developer scopes 时，应显式分配共享 domain/input/Ent/cache 文件的单一 owner，保持本合同总 allowlist；各自完成后冻结组合候选，再交独立 QA。
- QA 必须实际证明新版本拒绝旧快照、主 Group/RouteGroups 往返与管理更新失效；native/透传/Chat/Messages 和两条 WS 路径的强制、负向及 filter/block 优先级；两个 mounted 表单保存。以上为现有合同履约要点，不是新增批准条件或扩大范围。
- 真实 provider、数据库执行和正在运行的服务均未验收且不在本次授权内；保留原 dirty 输入/HEAD/index、未跟踪测试和候选 hash 的门禁继续有效。工作流全局 phase 当前 done 不能视为该任务已开发/验收，主控应在实际派工前按流程记录该任务阶段。
