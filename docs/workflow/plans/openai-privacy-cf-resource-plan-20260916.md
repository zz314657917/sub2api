# OpenAI privacy/account-check compatibility — design and resource gate

## Scope evidence

Candidate 9eb120dd40509d17df17bb61ceef22554d9ca45b changes four upstream files.
Current local req_client_pool.go still uses ImpersonateChrome. Repository search
found Impersonate:true only in CreatePrivacyReqClient; OpenAI/Gemini/Grok OAuth
factories reuse the pool but do not enable impersonation. No local business
change has been made for this candidate.

The local privacy service already keeps account selection, entitlement fallback,
proxy and 15-second request context handling. Preserve those local behaviors;
do not import the upstream service wholesale or change token refresh paths.

## Proposed bounded implementation

- Switch the enabled impersonation branch to the req library Firefox profile,
  preserving pool key, reuse, proxy validation and per-client timeout semantics.
- Recognize trimmed case-insensitive cf-mitigated=challenge for 403/503 privacy
  responses; retain existing body-marker fallback and ordinary JSON failures.
- Emit warning-level failed account/subscription events with typed status and
  cf_challenge. Do not promote potentially sensitive raw upstream bodies or
  credential-bearing error strings into new warning fields.
- Add localhost request/response tests for settings, accounts/check and
  subscriptions, including no challenge, challenge-header-only, body markers,
  transport error, success, request timeout/proxy/auth and cleanup boundaries.
- Test actual CreatePrivacyReqClient profile headers and unaffected OAuth
  factories, cache separation/reuse and invalid proxy behavior. Header assertions
  are local evidence, not proof that Cloudflare accepts the TLS profile.

## Required before full acceptance

User confirms a dedicated account and test environment, approved credential
source (never echoed or stored in reports), egress/proxy scope and allowed actions.
Privacy verification PATCH disables training for that test account, so the
account mutation must be explicitly in scope. Account/subscription GET tests
must also be authorized; no existing production credentials are auto-discovered.

Real-provider acceptance must exercise those three production code paths and
record redacted status/challenge/result evidence. A localhost fake, UA string,
or successful build cannot satisfy this gate. On CF challenge, record blocked
state and do not automate challenge circumvention or repeat unbounded requests.

Current status: design prepared; resource request sent to user. No approved
implementation contract and no code, provider, database or container action.
