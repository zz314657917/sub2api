### PASS: upstream-main-v0140-api-key-acl-denial-s31

# QA Report

## Task ID
upstream-main-v0140-api-key-acl-denial-s31

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-v0140-api-key-acl-denial-s31.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial|TestAPIKeyAuthIPRestrictionUsesForwardedClientIPWhenProxyTrusted" -count=1 -> pass
```
- manual checks:
```text
Reviewed upstream 56c62c59c against local api_key_auth middleware -> pass
Confirmed local code continues to use ip.GetTrustedClientIP(c), which delegates to Gin ClientIP() and trusted proxy configuration -> pass
Confirmed 82576e0a3 is already present locally in ensureEmailAuthIdentity via assignment to outer err -> pass
Confirmed 65fa72892 is already present locally in both chat-completions forwarding paths via handleOpenAIUpstreamTransportError -> pass
Confirmed latest v0.1.140 tail commits require frontend/payment/OAuth/Grok/migration/version scope and are skipped for this small sprint -> pass
Confirmed staged scope will exclude existing Ent, migration, proxy/account, frontend, service, handler, and knowledge dirty files -> pending staged audit
```

## Findings
- 未发现明确问题。

## Bug Owner Recommendation
codex-planner

## Root Cause
none

## Retest Scope
- If API key ACL client-IP extraction or server trusted-proxy configuration changes again, rerun the three S31 middleware tests.

## Knowledge Promotion
none
