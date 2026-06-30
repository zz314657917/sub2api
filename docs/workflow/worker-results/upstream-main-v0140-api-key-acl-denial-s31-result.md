### DONE: upstream-main-v0140-api-key-acl-denial-s31

# Worker Result

## Task ID
upstream-main-v0140-api-key-acl-denial-s31

## Status
done

## Summary
- Ported the local-relevant API key ACL denial message change from upstream `56c62c59c`.
- Updated ACL denial responses to include the resolved trusted client IP, falling back to `unknown` if no IP is available.
- Updated existing IP restriction regressions from generic denial text to concrete client-IP text.
- Added a trusted-proxy regression proving Gin trusted proxy configuration can surface the forwarded client IP while the untrusted-header test still rejects spoofed headers.
- Confirmed `82576e0a3` and `65fa72892` are already effectively present locally and were not re-ported.

## Changed Files
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `docs/workflow/tasks/upstream-main-v0140-api-key-acl-denial-s31.md`

## Commands Run
```text
gofmt -w backend/internal/server/middleware/api_key_auth.go backend/internal/server/middleware/api_key_auth_test.go -> pass
go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial|TestAPIKeyAuthIPRestrictionUsesForwardedClientIPWhenProxyTrusted" -count=1 -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/server/middleware	0.855s
```

## Risks
- Verification is unit-level; no live reverse-proxy deployment was exercised.
- Returning the client IP in the error response is intentionally limited to the resolved client IP and does not expose whitelist/blacklist rule details.
- Existing unrelated dirty proxy/account/frontend/knowledge files remain outside this sprint.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
