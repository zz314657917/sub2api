### PASS: key-fast-policy-20261105

## Findings

**独立后端验收通过，未发现本次实现引入的新阻断。** 本结论限于下述冻结候选与本地后端验证；已独立复现并保留一项既存 RawChat 回归失败，不将该失败或失败的 package 记为通过。

- `TestForwardAsRawChatCompletions_ForcesStreamUsageUpstreamAndPassesUsageDownstream` 在候选 unit 回归与独立原字节 baseline 中均于 `backend/internal/service/openai_gateway_chat_completions_raw_test.go:167` 报 `stream usage incomplete: missing terminal event`。测试全文 SHA 与候选初始清单一致，未修改旧测试或放宽断言。两次错误输出逐字一致，见 `baseline-failure-comparison.json`。
- 原字节 baseline 使用 R3 派生 overlay，包括原 GroupFast `20` 断言；独立核对并复制 23 个既有 Go 文件原字节，其 SHA 均匹配 `candidate-before.json`，8 个新增 Go 文件仅在 baseline overlay 中映射为空 package。baseline 实际结果为 GroupFast PASS、RawChat 同错 FAIL。没有在候选磁盘替换或恢复源码。
- 候选 GroupFast 旧测试相对原字节唯一变更确为获批的 `20→21`；新版本断言及其余组策略、快照、duplicate 和权限回归通过。它与上述既存 RawChat 失败的归因分开。
- 32 个后端交付文件与 Developer 最终清单、主控 39-source 冻结逐项一致。未发现 QA 写集越界；没有修改业务源码、旧测试或主目录。

## Executed Checks

### 冻结、范围与工具链

- Contract 完整 SHA256：`7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61`；已核对 R3 独立 PASS、Agent Matrix 与 Backend DONE。矩阵及合同指定 QA 模型为 `gpt-6.1-sol`，QA 与 Developer 分别执行并保存独立证据。
- 候选：`E:/codex-worktrees/sub2api/key-fast-policy-20261105`。
- 主控冻结：`E:/codex-runtime/sub2api/key-fast-policy-20261105/candidate-freeze.json`，文件 SHA256：
  `b04c489046bdf948697558ee7a84d1841e624fa65fd6657183abd3cfbbdef9a7`。
- **39-source** SHA256：
  `5226bfffd299cfdd4f17875d2acb0dae3e51fba2cfee713c0c31f33c3c6fa923`。
  包括后端 32 文件、前端 7 文件；前端本轮仅核对原字节哈希。
- **3826-input** SHA256：
  `8eb2d607d846b8059df6fe917e145e7db57c0ebc52c49b0fad2c85cc4ce8688d`。
  两个集合摘要均按冻结清单对应映射的 `JSON.stringify` UTF-8 字节重算；逐文件 SHA 与摘要在测试前、全部测试和构建后均匹配，零漂移。
- Developer 最终 `source-manifest.json` 文件 SHA256：
  `36833b026f808c54a705f1f8e713e2b71868b3e223db4b94a336791da097dd1e`。
  32 文件全部匹配主控 freeze，不把原 dirty 依赖和六个旧 untracked prerequisite 计入本次产品交付。
- 验收权威为 **当前候选的实际字节**。没有将候选称为主目录全部 4928 文件的原字节克隆。静态检查中发现的三个后端主目录/候选换行差异均已核对为候选初始即存在，LF 归一化后相等；相对 `candidate-before.json` 的其余 2744 个后端输入无漂移。主目录原字节及跨目录集成保护仍由主控负责。
- 全部 Go 命令仅使用 `E:/tools/go1.26.10/go/bin/go.exe`；实测 `go version go1.26.10 windows/amd64`。`GOROOT=E:/tools/go1.26.10/go`、`GOTOOLCHAIN=local`、`GOPROXY=off`、`GOSUMDB=off`，命令使用 `-mod=readonly`；复用现有 module/build 缓存，没有安装依赖或下载工具链。
- 首次阶段只读审查；收到主控完整 freeze 通知并校验全部输入后才启动测试与构建。

### 独立执行计数

计数来自本 QA 的 Go JSON `run/pass/fail/skip` 事件，顶层与 subtest 分开，不含 package 结束事件。每一行均已实际执行；所有列出的测试运行均无 SKIP、无 `no tests to run`，五包运行的每个包都有实际匹配测试。不同构建标签和补证运行不合并为去重测试数。

| 运行 | 顶层 PASS | 顶层 FAIL | subtest PASS | subtest FAIL | exit | 结论 |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| 新 KeyFast default | 14 | 0 | 465 | 0 | 0 | PASS，479 个 PASS 事件 |
| 新 KeyFast unit | 14 | 0 | 465 | 0 | 0 | PASS，479 个 PASS 事件 |
| 定向旧回归 default | 143 | 0 | 219 | 0 | 0 | PASS，362 个 PASS 事件 |
| 定向旧回归 unit | 201 | 1 | 246 | 0 | 1 | 447 个 PASS，唯一 RawChat 既存失败 |
| 独立原字节 baseline unit | 1 | 1 | 0 | 0 | 1 | GroupFast PASS；RawChat 同错 FAIL |
| 更新 mask/用量/Cafe 补充 default | 6 | 0 | 19 | 0 | 0 | PASS |
| 更新 mask/用量/Cafe 补充 unit | 6 | 0 | 19 | 0 | 0 | PASS |
| 独立 OAuth WS 补证 default | 1 | 0 | 14 | 0 | 0 | PASS |
| 独立 OAuth WS 补证 unit | 1 | 0 | 14 | 0 | 0 | PASS |
| 独立非 OpenAI 边界补证 default | 1 | 0 | 56 | 0 | 0 | PASS |
| 独立非 OpenAI 边界补证 unit | 1 | 0 | 56 | 0 | 0 | PASS |

新测试每种标签分包为：service 顶层 6/subtest 440，handler 2/22，DTO 1/0，repository 3/0，middleware 2/3。

定向旧回归由 `regression-selection.json` 记录的 28 份测试源文件、203 个静态候选顶层名称形成精确正则；不同 build tags 决定实际编译和执行数，以表内 JSON 实测计数为准。覆盖原组/global 策略、快照和缓存、API CRUD 校验、仓储、raw Chat/fallback、WS ingress/replay/error/cancel、最终 routing hint、tier metadata 与现有 priority/flex 计费。补充六个顶层用例专门验证 update mask、显式 usage reset、quota exhausted 状态更新、token cap 及 Cafe 生命周期保护。

### 合同逐项覆盖

| 合同要求 | 实际验证与证据 |
| --- | --- |
| schema/default/合法三态及旧空值 | 审查 schema 与八个必要 Ent 生成文件；SQLite 实际 create/read/update/list/auth roundtrip；Ent 默认 passthrough 和非法值拒绝；DTO 空值兼容。 |
| create/update/get/list、omit/null/类型 | Handler 真正请求/响应测试覆盖创建省略、三值、更新省略保留、显式切回 passthrough、get/list；POST/PUT 的 null、数字、布尔、数组、对象、空值、未知值、大小写及空白非法字符串均 400。 |
| owner/Cafe 保护 | 非 owner update 403/get 404，禁止持久化；seat/membership 均拒绝 policy 修改；既有安全更名和生命周期保护仍通过。 |
| repository/auth Select/SQL update mask | SQLite 实际往返与 PostgreSQL sqlmock 投影；policy-only SQL 仅声明 updated_at/policy，保留 quota_used、usage 及未请求 policy；相关旧 mask 回归通过。 |
| Redis/cache/version/invalidation | miniredis 真实序列化、读取及删除；v21 双向快照，0/19/20 拒绝并重新加载；L1/L2 命中、更新失效与发布；middleware 的后续请求实际从 Redis 恢复策略。 |
| 可信 context 与防 spoof | 真实 middleware→typed request context→HTTP upstream；伪造 body/header 不覆盖服务端 filter；字符串同名 key、错误类型不生效；effective key 切换刷新 context。 |
| actual OpenAI 限制与优先级 | 新矩阵 192 个组合覆盖 key/group/global/client；额外四平台×两策略×七 tier=56 个子测试，对照既有 passthrough 验证非 OpenAI HTTP/WS 行为不被新 key preference 改写。 |
| 多 HTTP 入口和 raw Chat 最终 metadata | 每种标签实际执行 224 个 HTTP 矩阵子测试：native APIKey、native OAuth、OAuth passthrough、Chat、Messages、raw Chat、Responses→raw Chat，均含 stream/nonstream。断言实际 outbound body、结果 tier、usage、OAuth routing hint；覆盖原非 nil flex/priority 替换和 key/global filter 清除；global block 零 upstream 调用。 |
| normal/passthrough WS 与 followup | 新测试每种标签 20 个实际 relay 子测试，覆盖首帧、后续省略 tier、不沿用上一轮 metadata、group/global 优先级和 block；其他事件 byte-unchanged 有明确断言。既有 replay/error/cancel/断连 drain 回归保持。 |
| OAuth WS routing 与逐轮 metadata | 仓库外独立探针每种标签 14 个子测试，双 relay 验证 key force/filter、group/global 覆盖、首帧握手 routing hint、followup tier 和 metadata；global block 零上游写入。 |
| 计费一致性 | 最终请求 tier 与 `OpenAIForwardResult.ServiceTier` 一致；原 priority/flex 价格、cost-with-tier、WS passthrough billing 用例通过。审查确认沿用原 RecordUsage 所有权和价格链，无本次价格/历史账单修改。 |
| migration 255 | 只读核对编号唯一、添加 NOT NULL/default passthrough、CHECK 三态和 Ent 描述一致；没有执行真实数据库迁移。 |

补充探针只存于 `qa-backend`，通过 Go `-overlay` 加入独立测试文件。冻结业务源码及原测试原字节不变；探针源码、overlay 和 SHA 另存为 `qa-oauth-ws-inputs.json`、`qa-provider-inputs.json`，不冒充 3826-input 的原组成部分。

### 构建和静态检查

- 独立 server build：`go.exe build -mod=readonly -o E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-backend/server.exe ./cmd/server`，exit 0。输出在本 QA 写集内，没有覆盖主控已有 server 产物。
- server.exe SHA256：`6dba3ef528cf2712393a6abf7e36fc08acba44bc1a2b60fb7211f064367c9c34`。
- 固定工具链的 `gofmt.exe -l` 检查全部 31 个交付 Go 文件，exit 0、stdout/stderr 空；未使用 `-w`。
- 对 32 个后端交付路径执行 scoped `git diff --check`，exit 0、stdout/stderr 空。新增文件另经源码审查、gofmt 与实际编译测试覆盖。
- 全部本 QA 启动的 Go jobs 已退出；未启动浏览器或前端服务。

### 可复核证据与命令

证据根：`E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-backend`。

- `before-tests-freeze-check.json`、`after-tests-freeze-check.json`：完整 3826/39 集合、逐文件核对结论和合同/manifest 身份。
- `test-counts.json`：分包、顶层/subtest、失败、skip、package 结果与空匹配检测。
- `new-default.jsonl`、`new-unit.jsonl`：独立新测试原始 Go JSON。
- `regression-default.jsonl`、`regression-unit.jsonl`、`regression-selection.json`：旧回归原始结果与精确测试选择。
- `supplemental-*.jsonl`、`qa-oauth-ws-*.jsonl`、`qa-provider-*.jsonl`：保护字段和独立补证。
- `baseline-independent-two.jsonl`、`baseline-overlay-r3.json`、`baseline-overlay-inputs.json`、`baseline-originals/`、`baseline-failure-comparison.json`：独立基线原字节、临时映射与同错归因。
- 每次 Go 执行对应的 `*-command.json` 和 `*.stderr.log`：完整可重放参数、固定工具链/缓存、cwd、PID、起止时间与 exit code。新测试为合同规定五包 `-run KeyFast20261105 -count=1`，分别 default/unit；基线仅执行合同指定两项。
- `coverage.json`、`gofmt-check.json`、`diff-check.json`、`toolchain-*.json`、`server-build-command.json`、`processes-final.json`：逐条覆盖、构建和静态检查。
- `qa-evidence-seal.json`：绑定本报告原字节 SHA、freeze 身份、全部 QA 证据和 server 产物；封存不包含自身，避免循环摘要。

## Unverified Risks

- 既存 RawChat terminal-event 测试仍失败，未修复；本任务允许按独立 baseline 分离，不能宣称旧 unit 回归全绿。
- 未执行真实 PostgreSQL migration、生产 Redis/provider、生产部署或 race；SQLite、sqlmock、miniredis 和受控本地 HTTP/WS 不能替代这些验证。
- 前端仅完成 7 份交付哈希核对，交互、typecheck/build/browser 由另一个独立 QA 负责。
- 本 PASS 绑定候选 3826 份实际编译输入。主目录的原有换行字节保护、39 文件精确集成以及主目录 fresh 验证由主控完成；本 QA 没有修改主目录，也没有据候选结果代签主目录验收。

## Recommendation

**可进入主控保护集成与最终审查；没有仍需修复的本任务后端阻断项。** 仅接受上述冻结 39-source/3826-input 身份，主控应保留非产品输入和主目录原字节，按计划完成精确集成及 fresh 关键测试/构建，并继续单列既存 RawChat 失败。任何后续产品字节变化均需要重新确认受影响验收。

本 QA 唯一仓库写入为本报告；其余输出在指定 `qa-backend`。未安装依赖、修改源码或旧测试、commit/push、更新 status/spec/knowledge、调用真实 provider、执行真实迁移或部署。
