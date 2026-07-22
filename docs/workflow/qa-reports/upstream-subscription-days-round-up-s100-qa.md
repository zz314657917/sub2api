### PASS: upstream-subscription-days-round-up-s100

## Findings

- No blocking issue was found.
- The runtime change matches upstream `d0fa8c63f`: positive partial 24-hour
  durations round up, exact durations remain exact, and non-positive durations
  return zero.
- The upstream test build tag was removed locally so the new boundary test runs
  in the default service test set; the local `unit` aggregate has unrelated
  compile drift.

## Executed Checks

- Test discovery listed both `TestCalculateProgress_BasicFields` and
  `TestUserSubscriptionDaysRemainingAt`.
- Focused test command passed both required tests.
- Broader `TestCalculateProgress|TestUserSubscriptionDaysRemainingAt` service
  regression passed.
- `gofmt -d`, `git diff --check`, conflict-marker scan, exact allowlist audit,
  and unmerged-index check passed.

## Unverified Risks

- This keeps upstream 24-hour duration semantics; it does not redefine a day as
  a calendar day in the user's timezone.
- No deployment or running-container smoke was performed.

## Recommendation

- `PASS / source-only`: publish the scoped source, tests, contract, and evidence.
