### PASS: key-fast-policy-20261105

Final Evaluator：本地实现、独立候选验收及主目录保护集成完成。
此判定覆盖本次密钥 Fast 策略，不代表生产迁移、真实 provider 或部署完成。

## Findings

- 未发现本次功能仍需修复的阻断项。新建/编辑密钥的高级设置新增
  OpenAI Fast 策略：透传（默认）、强制启用、过滤。
- `force` 将服务等级设为 `priority`；`filter` 仅移除规范化后的
  Fast/priority，保留其他等级；默认及旧密钥为 passthrough。
  全局策略最终裁决，管理员分组强制优先于密钥，密钥优先于客户端。
- 只影响实际 OpenAI 上游；策略来自已认证密钥。权限、托管密钥限制、
  既有价格、倍率、额度和用量写入所有权保持。已有 WS 会话需重连。
- 三处既存测试失败已独立复现并保留：前端 GroupBuyView 的 i18n
  mount、KeysView.createQuery 的延迟分组场景；后端 unit RawChat
  `stream usage incomplete: missing terminal event`。没有削弱原断言。
- 本次缓存版本升级引起的 GroupFast 旧断言冲突已修复，不能归为基线：
  R3 单独批准仅将版本预期 20 改为21，其余原文件字节与断言保持。

## Executed Checks

### 独立门禁与冻结

- 基准 HEAD：`0f6b739881ccf9e4ad423f464aeedcc9abd7f48a`。
- 最终 R3 合同完整 SHA256：
  `7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61`。
  独立 contract-review round3 PASS；原 FAIL、R2 PASS 均保留。
- 开发及独立 QA 按 Agent Matrix 分离。开发 DONE 不作为 QA PASS。
  独立前端、后端报告分别为同目录的
  `key-fast-policy-20261105-frontend.md`、
  `key-fast-policy-20261105-backend.md`。
- 完整候选冻结：`candidate-freeze.json`，文件 SHA256
  `b04c489046bdf948697558ee7a84d1841e624fa65fd6657183abd3cfbbdef9a7`。
  39 产品文件集合 SHA256
  `5226bfffd299cfdd4f17875d2acb0dae3e51fba2cfee713c0c31f33c3c6fa923`；
  3826 编译输入集合 SHA256
  `8eb2d607d846b8059df6fe917e145e7db57c0ebc52c49b0fad2c85cc4ce8688d`。
- 前端独立 seal 包含339项证据，SHA256
  `e7e7074d7d7a8d2bf2b1086d2a54ad9bb70d53d98e479775955687f7ec46cc73`；
  后端独立 seal 包含97项证据，SHA256
  `bf40aa51dfd0e6f3aa09800f41a4a8e9369b29624d48f44302f1df0e96c989e6`。
  主控已逐一重算所有证据及报告哈希，全部匹配。

### 已执行验证

| 检查 | 实际结果 |
| --- | --- |
| 独立新后端 default / unit，各一次 | 每次14顶层、465子测试通过，零失败/跳过 |
| 独立旧后端 default | 362 PASS 事件 |
| 独立旧后端 unit | 447 PASS，唯一 RawChat 既存失败 |
| 后端独立基线 | GroupFast PASS，RawChat 同行同错 FAIL |
| 后端附加保护验证，每种标签分别执行 | 更新mask/用量/Cafe 25、OAuth WS 15、非OpenAI边界57个PASS事件 |
| 独立新前端 / 分组Fast | 35/35、26/26 |
| 独立前端完整定向 / 原源码完整基线 | 402/404、367/369，同两旧失败；双文件基线24/26 |
| 独立类型、定向ESLint、前后端构建、Go格式与diff | 通过；前端独立构建259文件与冻结dist逐字节相同 |
| Controller production-dist 浏览器 | 25/25，三态CRUD/回显/刷新、桌面手机明暗、零页面错误 |
| 集成后主目录后端新功能与GroupFast | 25顶层、612子测试，共637 PASS事件；零失败/跳过 |
| 集成后主目录前端新功能与GroupFast | 61/61 |
| 集成后主目录类型、定向ESLint、前后端构建、gofmt与定向diff | 全部退出0 |

不同标签、重复验证和重叠套件不累加为去重测试数量。
后端覆盖 SQLite 实际持久化、SQL投影与更新mask、miniredis/cache失效、
认证context、真实本地HTTP/WS、优先级、followup、最终tier metadata、
OAuth routing hint及原计费链。独立QA补证探针使用仓库外overlay，
不修改冻结源码。

浏览器使用合成内存 CRUD fixture，不连接真实后端。R1 resize侧栏动画
瞬态溢出及R2初始手机视口导致的桌面定位超时均保留；R3设置起始视口、
等待原动画稳定后原25项断言通过。四截图已抽查并封存。session
`key-fast-policy-20261105`、专属profile、daemon33296、fixture44448
均已关闭并核查归零，其他预览与用户浏览器未操作。

### 保护集成

- 39份产品及4份开发/独立QA报告精确复制；原产品目标均先匹配快照，
  新目标先确认不存在。不是整棵候选目录覆盖。
- 4928份原跟踪文件受保护；除本任务产品及登记的Controller文档外，
  原文件字节保留，既有16项dirty范围与6份untracked Go prerequisite
  不作本任务交付。HEAD与逻辑index保持，未暂存、提交或推送。
- 候选初始检出有换行差异；实际编译输入中15份非产品文件与主目录
  仅CRLF/LF不同。已逐份归一化核对相等，并保留主目录原字节。
  候选与主目录使用各自原字节输入身份，没有宣称两者字节完全相同。
- 集成主目录3826输入集合SHA256：
  `107efe502ba8741167a2b452fc61f8150cd096d20e1f3145b1aea723d4f0528a`；
  产品39份与已验收候选精确一致，并在此主目录身份上执行fresh验证。

证据根目录：`E:/codex-runtime/sub2api/key-fast-policy-20261105`。
详见 `integration-receipt.json`、`main-inputs.json`、
`main-final/summary.json`、`main-final/backend-tests.jsonl`、
`main-final/frontend-tests.json`、两份 `controller-*-seal-verification.json`
及 `final-protection.json`。构建产物仅写入外部证据目录。

## Unverified Risks

- 未执行真实 PostgreSQL migration255、生产 Redis/provider、
  容器/服务更新或部署。SQLite/sqlmock/miniredis及受控HTTP/WS不替代生产验证。
- 三项基线失败仍存在，不能宣称全部旧套件全绿。
- Vite保留现有chunk/dynamic-import与Browserslist提示；未扩大改动修复。
- P/G/E doctor退出0，仍提示四份旧长文档及旧current-task未引用新任务；
  不在本次授权写集内改写共享knowledge或归档历史。

## Recommendation

本次源码交付 PASS，功能已合入当前工作区，可进入后续单独授权的发布流程。
线上生效仍需发布和迁移；本轮未提交、推送或部署。
