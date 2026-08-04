---
task_id: upstream-v0168-kimi-k3-s132-integration
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Port the Kimi K3 model-recognition slice from S132 onto the current mainline.
Exact K3 IDs must choose compatible accounts, use the intended fallback pricing
and thinking protocol, and permit the two documented Moonshot upstream hosts.

## Success Criteria

- Default outbound-host allowlists include only `api.moonshot.ai` and
  `api.moonshot.cn` in addition to the existing Kimi host; only the example
  configuration is updated, never a live configuration.
- An OpenAI OAuth account without an explicit mapping rejects exact bare K3
  Code IDs so compatible account selection may continue; explicit mappings,
  normal OpenAI API-key accounts, and unrelated aliases keep current behavior.
- Exact K3 model IDs resolve the declared fallback price and require thinking
  passback. K3-like unknown names and client `[1m]` syntax do not acquire a
  fallback price accidentally.

## Context

- Repo: `F:/mcplugins/sub2api`
- Integration worktree: `E:/codex-worktrees/sub2api-s132-kimi-k3-integration-20260804`
- Baseline: `main@a2055945d`
- Source behavior: `ee19f3ece feat(kimi): add K3 model support`
- Source QA: `d72a8477e:docs/workflow/qa-reports/upstream-v0168-kimi-k3-s132-qa.md`

## Allowed Paths

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_wildcard_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/thinking_protocol.go`
- `backend/internal/service/thinking_protocol_test.go`
- `deploy/config.example.yaml`
- `docs/workflow/tasks/upstream-v0168-kimi-k3-s132-integration.md`
- `docs/workflow/qa-reports/upstream-v0168-kimi-k3-s132-integration-qa.md`

## Denied Paths

- `backend/migrations/**`
- `backend/go.mod`
- `backend/go.sum`
- `frontend/**`
- `backend/internal/**/passkey*`
- `backend/internal/**/model_plaza*`
- `backend/internal/**/openai_codex_models*`
- `config*.yaml`
- `deploy/docker-compose*.yml`
- `Dockerfile*`
- `docker-compose*.yml`
- `outputs/**`
- `output/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `C:/Users/Administrator/.codex/memories/**`
- Any live configuration, container, provider account, deployment, or production state.

## Constraints

- Do not make a network request to Kimi/Moonshot during implementation or QA.
- Do not treat `[1m]` as a billable Kimi model ID and do not use broad `k3`
  substring matching.
- Do not port the source branch's Passkey, Model Plaza, Codex manifest, or
  superseded workflow files.
- The upstream pricing values are release-source data; their current merchant
  prices are not a deployment claim and must be called out as unverified.

## Acceptance Commands

Run from `backend`:

```powershell
gofmt -w <changed Go files>
go test ./internal/config -run '^TestLoadDefaultSecurityToggles$' -count=1
go test ./internal/service -run '^(TestGetFallbackPricing_FamilyMatching|TestResolveThinkingProtocol|TestAccountIsModelSupported)$' -count=1
go test ./... -run '^$'
go build ./...
```

Run from the worktree root:

```powershell
git diff --check
git ls-files -u
git diff --name-only main...HEAD
```

## Output

- One independently reviewable Kimi integration commit, an approved contract,
  and an evaluator report with findings, executed checks, unverified risks,
  and recommendation.

## Stop Rules

- Stop if actual runtime configuration, host allowlist architecture, routing,
  billing persistence, a migration, provider credential, deployment, or
  container change is required.
- Stop if the current mainline cannot retain the exact-ID and `[1m]` boundary.
- Stop before push, remote deletion, Docker operation, deployment, or any
  production write.

## Contract Review

`PASS / contract-approved`: the source diff applies cleanly to the current
mainline after excluding its obsolete workflow document. Existing default-host,
account-model-selection, fallback-pricing, and thinking-protocol seams make
the three success criteria directly testable. The scope neither changes live
configuration nor adds an outbound call; merchant pricing freshness remains an
explicit unverified runtime concern rather than a reason to broaden this
source-level port.
