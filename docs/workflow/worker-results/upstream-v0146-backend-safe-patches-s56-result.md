### PASS: upstream-v0146-backend-safe-patches-s56

Changed files:
- Payment response sanitization: `payment_order.go`, `payment_order_result_test.go`.
- OpenAI upstream models URL support: `upstream_models.go`, `upstream_models_test.go`.

Implementation notes:
- Cherry-picked `e76e0499d` and `f881ff7cb` with `-x` and no conflicts.
- Kept the batch backend-only; no frontend, deploy, migration, Ent, or container files were changed.
- Excluded clean-but-low-value test-only `b197ba61c` from this batch to keep S56 focused on runtime fixes.

Commands run:
- `go test ./internal/service -run "Test.*(Payment.*Order|PaymentOrder|NUL|UpstreamModels|ModelsURL|OpenAIModels)" -count=1` PASS.
- `git diff --check origin/main..HEAD` PASS.

Risks / follow-up:
- No full backend suite was run.
- No live payment-provider or upstream-model endpoint smoke was run.
- Remaining upstream candidates still need separate conflict resolution or product review.
