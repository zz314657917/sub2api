### BLOCKED: upstream-baseline-20261003

实现已完成，但合同测试未能进入测试执行阶段，故本 worker 不宣称验收通过。

变更：

- `backend/internal/service/billing_token_cost_request.go`
  - 增加 `TokenCostRequest.GroupID`。
  - 显式 `GroupID` 优先；缺省时回退 `Group.ID`。
  - resolver 在仅有 `GroupID` 时进入统一计费路径。
- `backend/internal/service/billing_token_cost_group_id_test.go`
  - 复用 `pricing_group_id_only_test.go` 的渠道定价 helper。
  - 覆盖真实渠道金额、显式 ID 优先、`Group.ID` 回退、nil 回退及 catalog fallback。

命令与结果：

- `backend: go test -tags unit ./internal/service -run 'Test.*(TokenCost|GroupID|LegacyLongContext)' -count=1`：退出码 1，未命中测试执行；测试包编译被现有基线错误阻断，包括 `stringPtr` 重定义、`computeTokenBreakdown`/`calculateCostInternal` 参数不匹配、`buildCountTokensRequest` 返回值不匹配及 `Proxy` 字段缺失。
- `backend: go build ./...`：退出码 0，成功。
- `root: git diff --check`：退出码 0，成功（仅有既有 `docs/workflow/status.md` 行尾转换警告）。
- 精确路径审计：本次 worker 新增/修改为两个合同允许业务文件；未修改共享 `F:` 工作区、未提交、未 push。隔离树原有 `docs/workflow/status.md` 改动保留。

限制：由于测试编译基线失败，新增金额回归、precedence/fallback/nil 和既有 long-context 测试未获得本命令的运行时证据；完整 build 已通过。
