### PASS: gpt61-context-subscription-20261012-integration

## Findings

未发现本次已批准六个 Go 文件的范围、候选哈希或主树集成缺陷。2026-10-13（Asia/Shanghai）Controller 已审阅四个生产 owner 的精确 diff、两份测试及独立 QA 的执行证据，并定向整合到 `F:/mcplugins/sub2api` 的本地 `main`。本结论限于受控测试和生产构建，不是部署或真实供应商验收。

- 基线：`167ea06927643856ea1354871761d82f721f07bd`。
- 指定上游对象：`07185812655ec2e0740ca76882879c0b9ae5ee0a`、`43768680d38ccdbd8fffc2cb91d3c1fc778a61c6`；按本地既有单体 gateway 架构适配，没有整仓 merge/cherry-pick。
- 独立 Contract Reviewer、Developer、QA 均按矩阵使用 `gpt-6.1-sol`。合同 PASS、Developer DONE、QA PASS；QA 会话退出码 0。Controller 没有修改委派的业务源码或测试。
- 主树原样 service 测试仍被其他任务的 `backend/internal/service/gateway_beta_test.go:138:7: undefined: newTestGatewayServiceForBeta` 阻塞。这不是本次新增失败，也没有修复、覆盖、暂存或提交该文件。首次主树失败保存在 `E:/codex-runtime/gpt61-context-subscription-20261012/main-focused-canonical.log`。

### 接收行为及保持不变的边界

1. 仅对已识别 GPT-6.1 Sol 的 API-key manifest 缺失/null `context_window` 补齐 `500000`；保留显式窗口、最大窗口、压缩阈值、其他字段/模型。OAuth manifest 原始字节和权益不变；客户端 ETag 对应实际返回 body。
2. 按最终发送模型及解析后的上游 OpenAI OAuth/setup-token 凭据判断模型价卡长上下文附加费豁免，合格 shadow 使用母账号凭据身份，但保持选中账号的账单和统计归属。客户 Sub2API 订阅、billing alias、response model 不授予此豁免。
3. API-key 阶梯、管理员自定义 interval/flat/per-request/image/video 定价、倍率、费用覆盖、去重和结算保持原行为；账号统计只在模型文件定价分支传播豁免，更高优先级规则不变。
4. 没有修改客户端配置、定价目录、前端、账号开关、数据库、历史用量、依赖或部署文件。

## Executed Checks

### 独立 QA：原样隔离候选

`E:/codex-worktrees/sub2api/gpt61-context-subscription-20261012` 上独立 QA 使用默认标签，没有 overlay：

- 合同 canonical service tests：PASS。
- 新增 15 个顶层测试各重测 10 次：150 个顶层 PASS、3580 个含子测试 PASS、0 FAIL、0 skip。
- 带既有 ETag 测试的补充专项：101 个顶层 PASS、474 个含子测试 PASS。
- OpenAI 包测试、完整 `go build -p 1 ./...`、gofmt、scoped diff check：PASS。
- 六个候选 owner 的初始/最终哈希一致，QA 未修改业务源码或测试。

完整独立报告：`F:/mcplugins/sub2api/docs/workflow/qa-reports/gpt61-context-subscription-20261012.md`。

### 定向复制与保护检查

执行 `node E:/codex-runtime/gpt61-context-subscription-20261012/integrate.mjs`：exit 0。

- 复制前确认本地 `main`、固定 HEAD、空暂存区、shared owner 原像，以及 Developer 和 QA 封存哈希。
- 六个 Go 文件与已验收候选字节一致。
- 定向复制前后核对 **12100 个无关跟踪/未跟踪文件**，无意外变化；不是仅凭 Git status 判断。
- 在任务专属 `integration-backup/` 保存原 owner，未删除原有文件。
- Developer/QA 报告按原字节复制到主树专属路径；没有覆盖混合工作流或交接文件。

证据：`E:/codex-runtime/gpt61-context-subscription-20261012/integration-result.json`、`integration-immediate-hashes.json`。

### 主工作树 fresh verification

执行 `node E:/codex-runtime/gpt61-context-subscription-20261012/verify-main.mjs`：exit 0。Go `go1.26.3 windows/amd64`。

在 `F:/mcplugins/sub2api/backend`：

```text
go test -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext' -count=1
```

首次 exit 1：上文已记录的 unrelated dirty fixture 编译错误。保留该失败，不声称原样主树 suite PASS。

```text
go build -p 1 ./...
go test -json -overlay E:/codex-runtime/gpt61-context-subscription-20261012/main-test-overlay.json -p 1 ./internal/service -run 'TestFetchCodexModelsManifest|TestOpenAISubscription|TestGPT61|TestBillingLongContext20261009|TestOpenAIGatewayServiceRecordUsage|TestCalculateCostUnified_ModelLongContext|TestCodexModelsManifestETagMatchesWeakAndMultipleValues' -count=1
go test -p 1 ./internal/pkg/openai -count=1
```

均 exit 0；生产完整构建 **没有 overlay**。主树专项实际执行 101 个顶层 PASS、474 个含子测试 PASS，包括全部 15 个新增顶层测试。

外置测试 overlay **只**用固定基线的 `gateway_beta_test.go` 内容避开该文件的无关脏新增测试；脚本确认替代字节与固定 Git 对象一致。没有替换任何生产文件、本次源码、本次测试或相关断言，没有修改主树 fixture。Overlay 文件位于仓库外，不纳入提交。

根目录六个 changed Go files 的 `gofmt -l` 和 scoped `git diff --check` 均 exit 0、无输出。验证结束重新核对六个 accepted owner 哈希及 12100 个保护文件：匹配，无无关漂移。

执行命令、退出码、时长、测试名称与日志路径：`E:/codex-runtime/gpt61-context-subscription-20261012/main-check-results.json`。

### 接收候选 SHA256

以下路径均以 `F:/mcplugins/sub2api/` 为根，哈希与 Developer/QA 候选一致：

| 文件 | SHA256 |
| --- | --- |
| `backend/internal/service/openai_codex_models_service.go` | `e7f2f0b35487a660df2fd9158d3c820defb552d4cb71513ef5ac240e0d50e223` |
| `backend/internal/service/openai_codex_models_service_test.go` | `487646c1b9ba1181048c3ee61bfafb1f2ea29c8976e47319c085f960bbc80c58` |
| `backend/internal/service/billing_service.go` | `634daaf96ca3224c181d23be395be2b883e7b45f7916decc63a01bd6d42cc47a` |
| `backend/internal/service/openai_gateway_service.go` | `386046ccd48bf235ca7078344c68fe057b997cf004550946c6e34118c3ecc245` |
| `backend/internal/service/account_stats_pricing.go` | `50218f79bcf7956142ad09300f95931119d1b4dfc29cef32a0b2a11f30df33fd` |
| `backend/internal/service/openai_subscription_long_context_test.go` | `a491fdebf872bc6400f5ec71566990694a24248028c8eda99fdf710d8cdb0d7d` |

## Unverified Risks

- 受控 HTTP stub/httptest、内存价卡及 fake repository 结算不等于真实供应商 entitlement/收费、真实数据库事务、客户端或部署验收。
- 未执行全量 tagged unit suite、全量 service suite、race、性能测试或浏览器验证；未调用付费 provider、生产数据库或容器。
- API-key 的 500k 仅是缺失 metadata 兜底，不保证上游实际支持 500k，也不会覆盖 OAuth 权益或显式 272k。
- 历史使用记录、已有客户端配置和运行中的容器没有被更新。
- 混合脏改的 `status.md`、`spec.md`、`main-log.md`、`current-task.md` 不属于本次提交；数字商店分支不合 main。

## Recommendation

**PASS：可按用户授权创建精确范围的本地 `main` 提交。** 提交仅含六个 Go 文件及合同、独立合同评审、Developer 报告、独立 QA 报告、本整合报告五份任务专属文档；提交前重新核对空暂存区、HEAD、候选哈希和实际暂存范围，提交后核对提交路径及原文件哈希。

本报告封存时，代码与主树验证已完成；最终提交号及提交后保护结果由 Controller 执行 `commit-mainline.mjs` 后记录在 `E:/codex-runtime/gpt61-context-subscription-20261012/mainline-commit.json`，并追加项目交接快照。**没有 push 或部署授权，不执行推送、容器更新或历史账单变更。**
