### PASS: upstream-v0216-gateway-20261015-integration

2026-10-17。本批选定的上游网关三项修复已完成独立合同复审、Developer、不同 QA 及保护性主树集成。前端六项和测试兼容前置亦已完成。本结论覆盖本批选定行为的本地源码合入；没有整体合并上游版本，没有提交、推送或部署。

## Findings

未发现本批阻断问题。网关只修改两个生产 owner 和四个测试文件：

- nil/非 OpenAI 账号在账号、渠道、全局开关前拒绝自动 Codex image bridge；显式用户工具与 OpenAI 开关优先级不变。
- Gemini tool-result 保留 base64/URL 图片、文字与图片顺序、工具名称绑定、错误标记和多个结果。
- 原生 Gemini 精确缺缓存 403 早于 policy、重试和账号处置透传。支持现有 OAuth 解包、过滤后的响应头及 request ID；探测限制 2 MiB，读取失败/达到上限不确认，最终以循环确认标记为准。普通凭据 403 和 overload 策略保留。

GPT-6.1 Sol 订阅凭据免模型长上下文附加费、缺失 API-key 元数据默认 500k 已存在于本地提交 `592d413876d8aa37b3cb1b98d1851b7e07767ead`，本轮只复验；API-key/管理员自定义价格与显式/OAuth 上下文值保留，没有修改历史账单或实际账号权益。

主树默认标签测试仍被原有无关脏文件 `backend/internal/service/gateway_beta_test.go:138` 的 `undefined: newTestGatewayServiceForBeta` 阻断。合入前后错误一致，该文件原字节未变。没有用 overlay、改 tag、隐藏文件或跳过测试掩盖此限制，不声称全库默认测试 PASS。

## Executed Checks

| 环节 | 实际结果 |
| --- | --- |
| 独立网关 QA default 专项 | 66 顶层 / 572 含子测试 PASS |
| 独立网关 QA apicompat/openai 全包 | 262 / 387 PASS |
| 独立网关 QA unit 专项及额外精确策略 | 9 / 106 与 2 / 8 PASS |
| 九个新增顶层用例精确执行 | 9 / 131 PASS；Developer 另已各重复三次通过 |
| 独立候选完整生产 build、scoped gofmt/diff | PASS |
| 集成主树完整 `go build -p 1 ./...` | PASS |
| 主树 canonical unit 联合专项 | 76 / 624 PASS，零 skip |
| 主树 apicompat/openai 全包 | 262 / 387 PASS，零 skip |
| 主树默认标签专项 | 编译退出1，既有 beta helper 缺失；测试实际执行0 |

主树联合专项按六个改动文件动态发现全部50个测试名，再加入 GPT61、长上下文及策略回归；50个名称和九个新增用例逐一核对均实际执行。主树及候选测试未使用 overlay。最初主树统计脚本未处理空构建日志而停止，修正空值统计后原样重新执行全部主树命令，业务源码未修改；上述结果来自完成后的原生退出码及日志。

## Scope and Seal

- 上游 `v0.2.16` 在线 peel 为 `01842576b3d5c1e2545b5a99ad555d36e28ed306`，三项候选均为其祖先；没有导入上游拆分架构、缓存 CRUD、handler、billing、ledger 或迁移。
- 网关六文件 seal SHA256：`253018f19e5f1733b69ede2226cade9bacdcb8cb9a3d6e7055d9f55c250c067b`。复制前五个既有主树目标与 base 原始字节一致，新文件不存在；复制后六文件与独立 QA seal 一致。
- 本次网关复制前后12112个非目标文件哈希零漂移，前端15文件保持 seal，六文件测试前置保持 seal。
- 对照最初12111文件快照，当前12118文件中27个获准业务/测试路径与各批 seal 一致；只有获准路径及三份主控流程/交接文档发生预期差异，非预期改动为空。七份新测试解释文件数增长。
- HEAD 仍为 `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`，暂存区为空；原有 apicompat、Claude、admin、gateway、UsageProgressBar 等无关脏改保留。没有提交、推送或部署。

## Evidence

证据根目录：`E:/codex-runtime/upstream-v0216-20261015/`。

- `gateway-integration-result.json`、`gateway-main-before.json`、`gateway-main-target-backup/`、`gateway-sealed-files.json`。
- `main-backend-checks.json`、`main-final-check-summary.json`、`main-gateway-selected-tests.json`、`main-backend-*.log`。
- `final-backend-main-audit.json`、`release-verification.json`。
- 独立 QA：`docs/workflow/qa-reports/upstream-v0216-gateway-20261015.md`。
- 前端及测试前置集成：同目录 `upstream-v0216-frontend-20261015-integration.md`、`upstream-v0216-test-compat-20261015-integration.md`。

## Unverified Risks

仅有本地编译、确定性转换、mock 转发和相关回归证据；没有真实供应商、生产数据库、真实收费/权益、浏览器或容器部署验收。前端15文件的此前独立测试、类型、lint和生产构建证据仍有效，本轮只核验其封存字节。默认标签 beta 编译阻断保留。

## Recommendation

本批选定行为本地源码合入 PASS，可审查未提交改动。提交、推送和部署须另按用户授权执行；不要将本报告解释为全量上游合并、默认全库测试通过或运行态已上线。
