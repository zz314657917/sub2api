### DONE: upstream-v0151-codex-luna-identity-s60b

changed_files:
- backend/internal/service/account_test_service.go
- backend/internal/service/account_test_service_openai_compact_test.go
- backend/internal/service/account_usage_service.go
- backend/internal/service/openai_gateway_messages.go
- backend/internal/service/openai_ws_forwarder_ingress_test.go
- backend/internal/service/openai_codex_version_consistency_test.go
- docs/workflow/worker-results/upstream-v0151-codex-luna-identity-s60b-result.md

commands_run:
- git status --short
- git rev-parse --verify 09f71cfdc && git rev-parse --verify 9d8ef743a
- go test -v ./internal/pkg/openai ./internal/service -run 'TestPairCodexClientIdentity|TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders|TestOpenAIGatewayServiceForwardAsAnthropicMappedNonCodexOmitsOriginator|TestIsOpenAIWSClientDisconnectError|TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestCodexVersionConstants_Consistency'
- git diff --stat -- knowledge
- git add backend/internal/service/account_test_service.go backend/internal/service/account_test_service_openai_compact_test.go backend/internal/service/account_usage_service.go backend/internal/service/openai_gateway_messages.go backend/internal/service/openai_ws_forwarder_ingress_test.go backend/internal/service/openai_codex_version_consistency_test.go docs/workflow/worker-results/upstream-v0151-codex-luna-identity-s60b-result.md
- git commit -m 'Complete S60b Codex Luna identity closure'
- git status --short

key_output:
- git rev-parse verified both required base commits: 09f71cfdc23344204a066f23d707a019b9a0c1a0 and 9d8ef743ae2cda815d86126cd86480f0d1b54b6d.
- go test -v ./internal/pkg/openai ./internal/service -run 'TestPairCodexClientIdentity|TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders|TestOpenAIGatewayServiceForwardAsAnthropicMappedNonCodexOmitsOriginator|TestIsOpenAIWSClientDisconnectError|TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestCodexVersionConstants_Consistency' => PASS for all targeted tests.
- git diff --stat -- knowledge => no output, proving zero net diff under knowledge/.

summary:
- Updated the OAuth image account-test path to choose its final custom or default User-Agent first, then run enforceCodexIdentityHeaders.
- Replaced the literal image-path opencode originator with the shared default Codex originator before identity enforcement.
- Added an Allowed focused assertion in account_test_service_openai_compact_test.go covering image account-test identity repair for a Luna-style UA that resolves to codex-tui.
- Left denied paths untouched and kept the rest of the pending S60b Allowed fixes intact for the closure commit.

risks:
- Acceptance coverage is intentionally narrow to the S60b closure set. No broader package sweep was added beyond the requested targeted tests.

contract_compliance:
- Did not modify knowledge/**, docs contracts, migrations, pricing, billing, or other denied paths.
- Added the image account identity assertion only in an Allowed test path.
- Ensured the targeted acceptance set includes PairCodexClientIdentity, explicit Luna recorder, ForwardAsAnthropic final no-originator recorder, Windows reset, image account identity, and exact 0.144.1 constants.
- knowledge/ audit shows zero net diff.
