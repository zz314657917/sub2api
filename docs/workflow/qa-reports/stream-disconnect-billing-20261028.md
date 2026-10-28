### PASS: stream-disconnect-billing-20261028 local integration

## Scope

选择性适配上游 `2b1f12938`、`94c7d6b02`、`4355236e0`、
`bd52e5d77` 的流式断开计费与收尾行为。上游固定为
`f178e5ae32659e8901644a9f2378a0439ffa36ae`；原主树基线
`1536c4849533460ab8c6ce81cf9e406c808e1cc4`。保留已验收 A1–A7/B2、
本地计费归属、首响应重试边界和全部无关 dirty。

- Gemini native/Messages/Chat 保留已观察累计 usage；图片只统计已完整
  写给客户端的交付，不将失败/partial write 或未完成请求默认算作一张。
- Anthropic native/passthrough/Chat/Responses 和 OpenAI
  Responses/passthrough 在断开或异常中断后保留可信 partial result/error；
  无 usage 不造账，failover 不携带可计费 partial result。
- Anthropic/OpenAI 读取具有 idle 与 total 上限，terminal 后关闭 body，
  退出时 join reader；同一次实际 body 只关闭一次。
- 已取消首响应尝试要求显式 Committed 证据，不能凭 prelude Size 或本地
  started 标记计费；已输出内容不重放。
- worker stopped/nil 使用有界同步兜底，显式 overflow drop/sample 策略
  保留，mandatory images 仍记录；trial、account/user slots 与队列恰好收尾一次。

## Findings

独立候选 QA 未发现未解决的本批生产缺陷。宽测试仍有原失败，不能宣称
全部 Go 测试通过；DeepSeek 本次未再失败也不视为本批根因修复。

首次保护检查因并行 onboarding 提交 `e004ab272`（4 个前端路径）和
2 个 Leaderboard dirty 正确 FAIL。原保护输入、FAIL 与原 HEAD/index
事实完整保留。补充保护审查只允许保存这 6 个前端路径的新字节，不扩大
产品写集。R2 独立保护 PASS 后已精确合回 24 文件，pre/post 检查通过。
随后并行排行榜提交 `646d791eb` 将原两条 dirty 原样提交，六前端 SHA
和 Git blob 均未变；原 final 正确拒绝新 HEAD/index。原 R2 输入/评审/
成功集成记录保持不变。final-only 身份补充独立 PASS 后，现场最终全树
检查也 PASS，未重复复制产品。当前 HEAD 为
`646d791eb79e68b1ef02e4179e6a11098a98c472`，index SHA 为
`e43a7bade3de935eac8696695179b991782ff433f9b64ec41368b121e81164a5`；
本批未创建提交、暂存、推送或部署。

## Executed Checks

证据根：`E:/codex-runtime/sub2api/stream-disconnect-billing-20261028`。
候选 freeze24 SHA256：
`ab0580e9e15be703d5861c5d130b0cbb0eb66e590f449e71802fdc0caa8cf46f`。
独立 QA seal SHA256：
`7f67e0dd5082bea5a42c0e89b2b5b836ac915eebe991be33275e5a59a86b2177`，
绑定 24 个源码/测试文件和 198 份证据，主控已全量核验。

| 独立 QA 检查 | default | unit |
| --- | ---: | ---: |
| service 宽回归 pass / old fail events | 1836 / 0 | 1904 / 4 |
| handler 宽回归 pass / old fail events | 391 / 7 | 391 / 7 |
| 新声明 52 个，各 count3 | 1098 pass events | 1098 pass events |
| A1–A7/B2 | 934 | 934 |
| 完整 apicompat | 293 | 323 |
| server/routes 首响应 | 28 | 28 |
| 外部 overlay probes | 188 | 188 |
| 外部关键链路 stress，163 keys × 10 | 1630 | 1630 |

固定命令的 baseline 比较零新增失败、零遗漏、零断言漂移、零 skip；
backend build、gofmt 和本批 diff 检查通过。详细命令、旧失败名称与断言、
loopback/handler/usage recorder 覆盖见外部 `qa/report.md`、
`qa/final-evidence-audit.json` 及封存原日志。

R3 只迁移批准的三条旧 ping 测试，其余原 23 个文件字节不变。
原 freeze23、旧 QA、首轮合同/探针/组合失败全部保留。QA 首轮五类
fixture/oracle 更正及不变断言在 `qa/fixture-corrections-r2.md`；
没有为探针通过修改冻结生产源码。未改 Antigravity keepalive 时序测试在
immutable baseline 独立 count10 为 8 pass / 2 fail，原失败不掩盖或豁免。

### Fresh main

`main-final-verification.json` 为 PASS。主树固定命令/比较结果与上述
QA 表一致：service 1836/1904、handler 各391通过，保留原 service-unit
4 个和两 tag 各7个 WS 失败，无新增失败/遗漏/断言漂移/skip。
主树新52声明各 tag count3、prior各934、protocol293/323、routes各28、
backend build、gofmt/diff、24源码及198证据审计全部 PASS。

原执行会话在宽测试及比较完成后中断，未把中断视为通过；原日志与
`main-final-session-interruption.json` 保留，只用 fresh
`main-final-resume-*` label 完成未确认的新测试/交叉回归/审计。
主树源码始终等于相同 freeze24，未为续跑修改实现或测试。

### 最终保护与独立批准

- `reconciliation/review-current.md/json`：R2 独立 PASS，原 R1 三项
  保护脚本缺口与原件保留在 `reconciliation/r1-fail/`。
- `reconciliation/integration-preflight.json`、`integration-postflight.json`
  与 `integration.json`：真实精确复制 24 文件，前后全部保护 PASS。
- `reconciliation/review-final-identity.md/json`：仅批准排行榜既有字节
  被提交后 main 的精确新身份；原 R2 审批与集成链保持完整。
- `reconciliation/final-protection.json`：正式最终 PASS。原/观察两套
  12230 文件主树快照、candidate/QA/baseline 全范围、四树45前置、
  24冻结源、198封存证据和六并行前端均核验通过。真实 HEAD/index
  在检查前后及范围检查中均核对，没有隐藏实际 Git 身份。
- 所有原失败、未完成的 stress、会话中断、首轮保护失败和 HEAD drift
  失败都保留；本地验收没有把它们改写为测试全绿。

## Unverified Risks

- 运行态证据限于本地 loopback、内存 fixture 和 Go overlay；未调用真实
  provider、线上 DB 或生产流量。
- Windows 无 C 编译器，race instrumentation unavailable；确定性双 tag
  count10 不能替代 race detector。
- 未提交、推送、部署、更新容器或执行数据库迁移。

## Recommendation

本批已完成本地选择性合入及验收：独立候选 QA、R2 精确集成、主树
fresh 组合、final-only 身份补充和最终全树保护均 PASS。保留原宽测试
失败及上述运行环境限制；提交、推送和部署不属于本次完成范围。
