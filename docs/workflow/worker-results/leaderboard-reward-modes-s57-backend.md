### DONE: leaderboard-reward-modes-s57

## changed files
- `backend/migrations/186_leaderboard_reward_modes.sql`
- `backend/internal/service/domain_constants.go`
- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/service/usage_leaderboard_reward.go`
- `backend/internal/service/usage_service_leaderboard_reward_test.go`
- `backend/internal/service/leaderboard_lottery_runner.go`
- `backend/internal/service/wire.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/server/api_contract_test.go`
- `backend/cmd/server/wire.go`
- `backend/cmd/server/wire_gen.go`
- `backend/cmd/server/wire_gen_test.go`

## commands run
- `go test ./internal/service -run "Test.*Leaderboard.*Reward|Test.*RedPacket|Test.*Lottery" -count=1` PASS
- `go test ./internal/repository -run "Test.*Leaderboard.*Reward|Test.*Leaderboard.*Packet|Test.*Leaderboard.*Lottery" -count=1` PASS
- `go test ./internal/handler -run "Test.*Leaderboard.*" -count=1` PASS
- `go test -tags=unit ./internal/server -run "TestAPIContracts" -count=1` PASS
- `go test ./cmd/server -run "TestProvideCleanup|^$" -count=1` PASS
- `git diff --check -- backend/migrations/186_leaderboard_reward_modes.sql backend/internal/service backend/internal/repository/usage_log_repo.go backend/internal/repository/usage_log_repo_request_type_test.go backend/internal/pkg/usagestats/usage_log_types.go backend/internal/handler/usage_handler.go backend/internal/handler/usage_handler_leaderboard_test.go backend/internal/handler/admin/setting_handler.go backend/internal/handler/dto/settings.go backend/internal/server/api_contract_test.go backend/cmd/server/wire.go backend/cmd/server/wire_gen.go backend/internal/service/wire.go` PASS

## risks
- `backend/cmd/server/wire_gen_test.go` was added to the contract allowed paths after review because the new lottery runner changes the cleanup provider signature.
- Frontend and `docs/workflow/main-log.md` are also dirty from parallel work and were not modified by this backend worker.
- Lottery settlement can run from the scheduler runner or lazy user reads/claims; DB unique constraints and idempotent attach paths are relied on for duplicate protection.
