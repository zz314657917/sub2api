### DONE: upstream-response-model-audit

已按批准 contract 完成 Developer 实现，未改变计费、路由、账号状态或客户端请求/下游改写语义。独立 QA 正在对最终冻结的 WSv2 relay 修复复验；本报告不替代 QA 结论。

## 变更范围

工作区状态共 67 个实现文件：Ent 生成及 schema 9 个、后端 service/repository/handler/迁移与测试 45 个、前端 API/UI/i18n/测试 13 个。新增文件为两条迁移、响应模型观察器、三组 service 测试、PostgreSQL 测试和 Admin handler 测试；其余为允许清单内既有文件。未提交、未推送、未部署。

## 协议矩阵与证据边界

| 路径 | 静态接入检查 | 实际运行证据 |
| --- | --- | --- |
| OpenAI HTTP | `Forward` 在原始响应改写前观察 | 本地 stub transport 的 paired `Forward + RecordUsage`：`TestUpstreamResponseModelHTTPNonInterference` |
| OpenAI SSE | passthrough/conversion 在原始 SSE payload 观察 | 本地 stub transport 的 paired `stream:true Forward + RecordUsage`：`TestUpstreamResponseModelSSENonInterference` |
| OpenAI WSv2 | relay、adapter 和逐 turn callback 接入 | repository capture-dialer/local frame fixture 的 paired `Forward + RecordUsage`：`TestUpstreamResponseModelWSNonInterference`；relay package 覆盖 foreign ID、id-less、terminal override、callback/fallback 关联 |
| OpenAI HTTP bridge、legacy WS | 观察器已接在各自生产转发路径 | 静态接入及定向本地单元测试；未对外部 WS provider 运行 |
| Anthropic、Gemini、Antigravity（HTTP/SSE/协议转换） | 观察器已在对应生产转发和改写前路径接入 | 定向 service 测试验证本地 fixture；未逐协议调用真实 provider |

所有 WS 证据均来自本仓库 local fixture/capture-dialer，不是付费 provider 或外部 WebSocket server 的运行证明。

## 已实际执行的验证

- `go test ./internal/service -run '^TestUpstreamResponseModel' -count=1 -v`：PASS。
- `go test ./internal/service/openai_ws_v2 -count=1`：PASS（最终 relay 隔离修复后复跑）。
- `go test ./internal/repository -run '^TestUpstreamResponseModelPostgres' -count=1 -v`：PASS，使用 controller 提供的 disposable task PostgreSQL DSN；覆盖旧 NULL、重复迁移、batch/best-effort、分页、true/false/NULL、统计/趋势、模型/分组及 invalid partial index recovery。
- `go test ./internal/handler/admin -run 'TestAdminUsageHandlersPropagateUpstreamModelMismatch|TestAdminUsageResponseModelAuditFieldsAreRawAndTriState' -count=1 -v`：PASS（10 个表格子用例及 DTO 暴露）。
- 前端冻结前已实际通过 focused Vitest 34/34、ESLint、`vue-tsc --noEmit` 与 Vite build；QA 已完成隔离 profile 的 mock 浏览器验收和精确清理。
- `go build ./...`：PASS（最终 build cycle）。`git diff --check`：PASS（最终 relay 修复后复跑）。

`go test -tags unit ./internal/service -run '^TestUpstreamResponseModel' -count=1` 未作为本变更失败：controller 在编辑前已复现该 tags 基线 fixture 编译失败（duplicate `stringPtr`、过时 billing/countTokens/proxy 参数与字段），本任务未修改这些基线错误，且未以该命令的失败或零测试作为通过证据。

## 最终 WSv2 relay 修复

最终冻结文件为 `backend/internal/service/openai_ws_v2/passthrough_relay.go` 和 `backend/internal/service/openai_ws_v2/passthrough_relay_internal_test.go`。显式 response ID 保留原有 `activeTurn`、response ID、duration、usage 和计费生命周期；id-less 事件只观察独立 audit owner。无 ID terminal 不写 `lastResponseID`，没有 audit owner 时清空审计结果；`enrichResult` 仅在审计模型来源 ID 与结果 ID 相同（两者均空也可）时输出，无法关联时保持 NULL。已完成 turn 的 callback 不会被后续 id-less terminal/delta 覆盖。

## 约束与未覆盖边界

Developer 未执行 provider 调用、容器生命周期操作、依赖变更或生产部署。PostgreSQL 证据只使用 controller 提供的空白 disposable task 数据库。真实 provider/外部 WS server 验证不在本轮执行范围。主控收口补记：独立 QA 已对最终 relay 代码复验 PASS，见 docs/workflow/qa-reports/upstream-response-model-audit-qa.md；主控已精确清理任务 PostgreSQL。
