### DONE: upstream-cn-providers-s226-a

# Worker Result

## Task ID

`upstream-cn-providers-s226-a`

## Status

`done`

## Base And Head

- base: `98daf5b8d9008c9db6753631a62ede9a3ff8ca6d`
- implementation head: `ba7c00c78a746097a192cd8b086b3dea799a1c34`

## Summary

- Added Kimi, Zhipu, and DeepSeek platform constants; `payg|coding` account
  modes; and `chat_completions|anthropic|responses` protocol values.
- Added provider/mode/protocol-aware default Base URL accessors. Explicit
  `credentials.base_url` remains authoritative; missing or invalid protocol
  values retain Chat Completions behavior, and Responses remains DeepSeek-only.
- Added protocol-aware API-key and Anthropic auth-scheme access plus model-list
  request construction, without adding or enabling any gateway route.

## Changed Files

- `backend/internal/domain/constants.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_service.go`
- `backend/internal/service/anthropic_apikey_auth.go`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/upstream_models.go`
- `backend/internal/service/cn_provider_foundation_test.go`
- `docs/workflow/worker-results/upstream-cn-providers-s226-a-result.md`

## Commands Run

```text
go test -list ^<each of 8 S226-A tests>$ ./internal/service -> PASS, 8/8 discoverable
go test ./internal/service -run <8 S226-A tests> -count=10 -> PASS (0.084s)
go test ./internal/service -count=1 -> PASS (60.343s, exit 0)
go test ./cmd/server -run '^$' -count=1 -> PASS (5.549s)
gofmt -w <7 Go implementation/test paths> -> PASS
gofmt -d <7 Go implementation/test paths> -> PASS
git diff --check 98daf5b8d9008c9db6753631a62ede9a3ff8ca6d...HEAD -> PASS
git diff --name-only <base>...HEAD -> exactly 7 implementation/test paths before this report
git diff --name-only --diff-filter=U -> PASS (empty)
git ls-files -u -> PASS (empty)
git merge-base --is-ancestor 901a0439f1575a45c29150e04d2ccc3ed87f4948 upstream/main -> PASS
git merge-base --is-ancestor 4b667ccd45747162ac58545cdbb6f6d88737bf04 upstream/main -> PASS
git merge-base --is-ancestor e7285453829930e2889432dcd18d7a0c2ba18481 upstream/main -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 0.084s
ok github.com/Wei-Shaw/sub2api/internal/service 60.343s
ok github.com/Wei-Shaw/sub2api/cmd/server 5.549s [no tests to run]
```

## Risks

- Upstream `AllowedQuotaPlatforms` and `AllowedSchedulingThresholdPlatforms`
  additions are intentionally excluded: this checkout lacks the quota product
  and generic threshold prerequisite; those deltas belong to later authorized
  work, not S226-A.
- `IsOpenAICompatible` remains openai/grok-only. Expanding it would alter
  scheduler/gateway eligibility before S226-C, contrary to this batch boundary.
- No provider requests were sent. Model-list tests only construct requests.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
