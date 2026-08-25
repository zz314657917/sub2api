### PASS: upstream-openai-spark-shadow-auto-reset-s259-a

## Findings

- 未发现明确问题。独立复验基于 `f5a73cf7d..d5877bb16` 与实际测试执行；未将 Developer/Controller 的测试结论作为通过证据。
- 范围：相对 `f5a73cf7d` 的产品改动为 23 个文件，均在 S259-A `Allowed Paths`；另有唯一允许的 worker result 报告。未见 frontend、handler、routes/Wire、config、scheduler、quota/reset、UI 或依赖改动。Ent 生成文件、两份 additive migration、repository/service owners 与唯一新增 outbound test 均在允许范围内。
- Resolver：shadow 仅通过 `parent_account_id` 查询父账号，拒绝 shadow 链、非 OpenAI 平台及非 OAuth 父账号；普通账号直接返回自身。shadow 的凭据持久化入口直接跳过。
- Fail-closed：HTTP raw/Chat/Messages 的两个上游 request builder 先解析 credential account，失败即返回、不会构造请求；普通 WS、入站 WS 与 WSv2 passthrough 都由 header builder 返回 nil，并在 acquire/dial 前拒绝。普通 WS 和 WSv2 passthrough 的真实 fake-dial 路径均已独立运行并捕获。
- Outbound：HTTP fake upstream 验证 child 使用父账号 Authorization 和 account header，随后普通 OAuth 尝试替换为自身；普通 WS 验证拨号 header 与首个 `response.create`，不合格父账号零 dial；WSv2 passthrough 验证拨号 header、首帧 `response.create`、中间 `session.update`、后续 `response.create`，缺父时 header 构造失败且零 dial。报告不记录凭据或真实账号标识。
- 上游：`git merge-base --is-ancestor bdf7ead15 6f972145b`、`96b160d9`、`upstream/main` 均退出 0；`bdf7ead15` 显示为 95 文件大功能链，因此本切片的行为级适配不应机械 cherry-pick。

## Executed Checks

```text
git diff --name-only/check f5a73cf7d..d5877bb16
PASS: exact allowlist; diff check clean

gofmt -d <all changed Go files>
PASS: no output

git diff --cached --name-only; git ls-files -u; conflict-marker scan
PASS: no cached paths, no unmerged entries, no conflict markers

Push-Location backend
GOTMPDIR=backend/.tmp/s259-a-qa-go-build
go test ./internal/service -list 'Test(ResolveCredentialAccount|PersistAccountCredentials.*Shadow|OpenAI.*Shadow.*Credential|.*Shadow.*Outbound)'
PASS: 7 default-tag tests discovered

go test ./internal/service -run 'Test(ResolveCredentialAccount|PersistAccountCredentials.*Shadow|OpenAI.*Shadow.*Credential|.*Shadow.*Outbound)' -count=10 -timeout=3m
PASS: 2.133s

go test ./internal/repository -run 'TestAccount.*Shadow' -count=10 -timeout=3m
PASS: 0.065s

go test ./internal/service -count=1 -timeout=3m
PASS: 65.886s

go test ./cmd/server -run '^$' -count=1
PASS: 0.060s (no tests to run)
Pop-Location
```

## Primary Worktree Protection

- 只读快照 `F:/mcplugins/sub2api`：索引 patch hash 为空索引 `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`，工作树 patch hash 为 `1e3ef43cd768ae5fb510c47cdf636e3187e5e372`，porcelain 状态 122 行（已有 Pixel Cafe/Group/Settings/knowledge/assets/outputs 等用户改动）。本 QA 未在该路径执行写入、暂存或测试。

## Residual Risks

- 本轮按合同未执行真实数据库 migration、provider、container、browser、deploy 或真实 parent 运行态验证；迁移约束与 fake-upstream 行为只具代码级证据。
- QA 新增的 `.tmp/s259-a-qa-go-build` 为测试编译临时目录，未纳入提交。

## Recommendation

- 可继续进入约定的后续父身份运行态门禁；不应以本代码级 QA 代替真实 parent 的 S259-E/S258 验证。

## knowledge_candidates

- none: task-local shadow credential adaptation，需后续真实 parent 验证后才适合沉淀。
