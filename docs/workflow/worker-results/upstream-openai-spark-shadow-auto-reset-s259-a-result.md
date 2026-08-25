### DONE: upstream-openai-spark-shadow-auto-reset-s259-a

## Scope

- 实现提交：`54c1b75a5 feat(openai): add spark shadow credential inheritance`
- 严格限定在 contract 的 Ent schema/generated owners、两份 additive migration、账号映射、凭据 resolver、OpenAI HTTP/WS 调用点与默认 tag 测试；未修改前端、handler、route/Wire、config、调度、quota、自动重置或主工作区。
- 先前 Terra Developer 已停止，原因是缺少真实 HTTP/WS/WSv2 出站证据。Controller 在相同 allowlist 内补充 fail-closed 检查与单一 outbound regression 文件；未扩大产品范围。

## Behavior

- `Account` 增加 `parent_account_id`、`quota_dimension`，Ent 代码由 schema 生成；两份 migration 以约束和并发索引保证自指、非法维度、global shadow、孤立父引用以及每父 Spark shadow 的数据库级拒绝。
- resolver 对普通账号返回自身；shadow 只接受非 shadow 的 OpenAI OAuth 父账号。父解析失败时 HTTP request builder 返回错误；普通 WS、WS ingress 与 WSv2 passthrough 在拨号前返回错误，禁止匿名或空 header 出站。
- token、ChatGPT account header、OAuth 请求目标和 WS 自定义 User-Agent 使用 resolved parent；shadow 不持久化 credential refresh。普通 OAuth 后续尝试使用自己的 credential，不复用先前 shadow parent。

## Outbound evidence

- HTTP fake upstream：shadow 的 Authorization 和 `chatgpt-account-id` 来自父账号；同一 service 随后转普通 OAuth 时两项均替换为普通账号值；父账号缺失时 recorder 没有新增请求。
- 普通 WS fake dialer：dial header 和首个 `response.create` 使用父账号；后续普通 OAuth 拨号使用自身 header/frame；不合格父账号的 direct WS entry 没有新增拨号。
- WSv2 passthrough fake dialer：dial header 来自父账号，且真实 relay 捕获首个 `response.create`、中间 `session.update` 与后续 `response.create` 三帧；缺父 resolver 返回 nil header，未建立 passthrough dial。

## Verification

```text
go test ./internal/service -list 'Test(ResolveCredentialAccount|PersistAccountCredentials.*Shadow|OpenAI.*Shadow.*Credential|.*Shadow.*Outbound)'
7 default-tag tests discovered

go test ./internal/service -run 'Test(ResolveCredentialAccount|PersistAccountCredentials.*Shadow|OpenAI.*Shadow.*Credential|.*Shadow.*Outbound)' -count=10 -timeout=3m
PASS (2.135s)

go test ./internal/repository -run 'TestAccount.*Shadow' -count=10 -timeout=3m
PASS (0.061s)

go test ./internal/service -count=1 -timeout=3m
PASS (66.190s)

go test ./cmd/server -run '^$' -count=1
PASS (0.071s)

gofmt, git diff --check, exact allowlist, cached/unmerged index
PASS
```

## Review and residual risk

- `bdf7ead15` is an ancestor of `upstream/main`, `6f972145b`, and `96b160d9`; this is behavior-level adaptation, not a cherry-pick of the 95-file upstream product chain.
- Primary worktree was inspected read-only and retains its user-owned Pixel Cafe/Group/Settings/knowledge/assets/outputs state. No database, provider, browser, container, deployment, push, or migration execution occurred.
- `docs/workflow/tasks/upstream-openai-spark-shadow-auto-reset-s259-a.md` remains an unstaged pre-existing line-ending-only worktree state and is deliberately outside this implementation commit.
- S259-B scheduling/admin, S259-C reset automation, S259-D frontend, and S259-E/S258 parent-identity revalidation remain separate gates.

## Contract compliance

- Allowed paths: PASS.
- Denied paths and external state: PASS.
- Secrets: no real credential, parent ID, or provider detail is recorded.
- knowledge_candidates: none; this is task-local and needs later integration/QA before becoming stable guidance.
