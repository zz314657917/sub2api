package admin

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCodexImportEntryAcceptsAgentIdentityAuthJSON(t *testing.T) {
	value := buildAgentIdentityImportValue(t, "runtime-import", "team-import", "user-import", "")
	identity := value["agent_identity"].(map[string]any)
	identity["email"] = "agent@example.invalid"
	identity["plan_type"] = "team"
	identity["chatgpt_account_is_fedramp"] = false

	item, err := normalizeCodexImportEntry(codexImportEntry{Index: 1, Value: value})
	require.NoError(t, err)
	require.True(t, item.IsAgentIdentity)
	require.Equal(t, service.OpenAIAuthModeAgentIdentity, item.Credentials["auth_mode"])
	require.Equal(t, "runtime-import", item.Credentials["agent_runtime_id"])
	require.Equal(t, identity["agent_private_key"], item.Credentials["agent_private_key"])
	require.Equal(t, "team-import", item.Credentials["chatgpt_account_id"])
	require.Equal(t, "user-import", item.Credentials["chatgpt_user_id"])
	require.NotContains(t, item.Credentials, "access_token")
	require.NotContains(t, item.Credentials, "refresh_token")
	require.NotEmpty(t, item.WarningTexts)
}

func TestNormalizeCodexImportEntryAcceptsAgentIdentityCamelCaseJSON(t *testing.T) {
	privateKey := buildAgentIdentityPrivateKey(t)
	item, err := normalizeCodexImportEntry(codexImportEntry{
		Index: 1,
		Value: map[string]any{
			"authMode": "agent_identity",
			"agentIdentity": map[string]any{
				"agentRuntimeId":          "runtime-camel",
				"agentPrivateKey":         privateKey,
				"taskId":                  "task-camel",
				"accountId":               "team-camel",
				"chatgptUserId":           "user-camel",
				"chatgptAccountIsFedramp": "true",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, item.IsAgentIdentity)
	require.Equal(t, "runtime-camel", item.Credentials["agent_runtime_id"])
	require.Equal(t, "task-camel", item.Credentials["task_id"])
	require.Equal(t, true, item.Credentials["chatgpt_account_is_fedramp"])
}

func TestNormalizeCodexImportEntryAcceptsSnakeAgentIdentityAuthMode(t *testing.T) {
	item, err := normalizeCodexImportEntry(codexImportEntry{
		Index: 1,
		Value: map[string]any{
			"auth_mode":         "agent_identity",
			"agent_runtime_id":  "runtime-root",
			"agent_private_key": buildAgentIdentityPrivateKey(t),
			"task_id":           "task-root",
			"account_id":        "team-root",
			"chatgpt_user_id":   "user-root",
		},
	})
	require.NoError(t, err)
	require.True(t, item.IsAgentIdentity)
	require.Equal(t, "team-root", item.Credentials["chatgpt_account_id"])
}

func TestNormalizeCodexImportEntryRejectsInvalidAgentIdentityPrivateKeyWithoutEcho(t *testing.T) {
	const invalidPrivateKey = "synthetic-invalid-private-key"
	_, err := normalizeCodexImportEntry(codexImportEntry{
		Index: 1,
		Value: map[string]any{
			"auth_mode": "agentIdentity",
			"agent_identity": map[string]any{
				"agent_runtime_id":  "runtime-invalid",
				"agent_private_key": invalidPrivateKey,
				"account_id":        "team-invalid",
				"chatgpt_user_id":   "user-invalid",
			},
		},
	})
	require.EqualError(t, err, "agent identity private key 格式无效")
	require.NotContains(t, err.Error(), invalidPrivateKey)
}

func TestImportCodexSessionsCreatesAgentIdentityWithoutOAuthExpiry(t *testing.T) {
	svc := newCodexImportMemoryAdminService(nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	result, err := handler.importCodexSessions(context.Background(), CodexSessionImportRequest{
		SkipDefaultGroupBind: boolPtr(true),
	}, []codexImportEntry{{
		Index: 1,
		Value: buildAgentIdentityImportValue(t, "runtime-no-expiry", "team-no-expiry", "user-no-expiry", "task-no-expiry"),
	}})
	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	require.Zero(t, result.Failed)
	require.Len(t, svc.createdAccounts, 1)
	created := svc.createdAccounts[0]
	require.Nil(t, created.ExpiresAt)
	require.Nil(t, created.AutoPauseOnExpired)
	require.Equal(t, service.OpenAIAuthModeAgentIdentity, created.Credentials["auth_mode"])
	require.NotContains(t, created.Credentials, "access_token")
	require.NotContains(t, created.Credentials, "refresh_token")
}

func TestBuildCodexAgentIdentityKeysUseChatGPTAccountOnly(t *testing.T) {
	require.Equal(t, []string{"account:team-a"}, buildCodexAgentIdentityKeys("team-a"))
	require.Nil(t, buildCodexAgentIdentityKeys("  "))
}

func TestCodexAgentIdentityIndexSeparatesTeamsForSameUser(t *testing.T) {
	existing := service.Account{
		ID: 1,
		Credentials: map[string]any{
			"auth_mode":          service.OpenAIAuthModeAgentIdentity,
			"chatgpt_account_id": "team-a",
			"chatgpt_user_id":    "same-user",
			"agent_runtime_id":   "runtime-a",
		},
	}
	index := buildCodexAccountIndex([]service.Account{existing})

	matched, _ := index.Find(buildCodexAgentIdentityKeys("team-b"), "same-user")
	require.Nil(t, matched)

	matched, matchedKey := index.Find(buildCodexAgentIdentityKeys("team-a"), "same-user")
	require.NotNil(t, matched)
	require.Equal(t, int64(1), matched.ID)
	require.Equal(t, "account:team-a", matchedKey)
}

func TestImportCodexSessionsKeepsAgentIdentityTeamsSeparate(t *testing.T) {
	svc := newCodexImportMemoryAdminService(nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	entries := []codexImportEntry{
		{Index: 1, Value: buildAgentIdentityImportValue(t, "runtime-a", "team-a", "same-user", "task-a")},
		{Index: 2, Value: buildAgentIdentityImportValue(t, "runtime-b", "team-b", "same-user", "task-b")},
	}

	result, err := handler.importCodexSessions(context.Background(), CodexSessionImportRequest{
		SkipDefaultGroupBind: boolPtr(true),
	}, entries)
	require.NoError(t, err)
	require.Equal(t, 2, result.Created)
	require.Zero(t, result.Updated)
	require.Zero(t, result.Skipped)
	require.Len(t, svc.createdAccounts, 2)
}

func TestImportCodexSessionsMergesAgentIdentityRuntimesForSameTeamAndUser(t *testing.T) {
	first := buildAgentIdentityImportValue(t, "runtime-a", "team-a", "same-user", "task-a")
	second := buildAgentIdentityImportValue(t, "runtime-b", "team-a", "same-user", "task-b")
	firstIdentity := first["agent_identity"].(map[string]any)
	existing := service.Account{
		ID:       41,
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_mode":          service.OpenAIAuthModeAgentIdentity,
			"agent_runtime_id":   firstIdentity["agent_runtime_id"],
			"agent_private_key":  firstIdentity["agent_private_key"],
			"task_id":            firstIdentity["task_id"],
			"chatgpt_account_id": firstIdentity["account_id"],
			"chatgpt_user_id":    firstIdentity["chatgpt_user_id"],
		},
	}
	svc := newCodexImportMemoryAdminService([]service.Account{existing})
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := handler.importCodexSessions(context.Background(), CodexSessionImportRequest{
		SkipDefaultGroupBind: boolPtr(true),
	}, []codexImportEntry{{Index: 1, Value: second}})
	require.NoError(t, err)
	require.Zero(t, result.Created)
	require.Equal(t, 1, result.Updated)
	require.Len(t, svc.updatedAccounts, 1)
	require.Equal(t, "runtime-b", svc.updatedAccounts[0].input.Credentials["agent_runtime_id"])
	require.Equal(t, "task-b", svc.updatedAccounts[0].input.Credentials["task_id"])
}

func TestImportCodexSessionsCreatesFourteenAgentIdentityUsersInOneTeam(t *testing.T) {
	const accountCount = 14
	entries := make([]codexImportEntry, 0, accountCount)
	for i := 1; i <= accountCount; i++ {
		entries = append(entries, codexImportEntry{
			Index: i,
			Value: buildAgentIdentityImportValue(
				t,
				fmt.Sprintf("runtime-%02d", i),
				"team-shared",
				fmt.Sprintf("user-%02d", i),
				fmt.Sprintf("task-%02d", i),
			),
		})
	}
	svc := newCodexImportMemoryAdminService(nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	result, err := handler.importCodexSessions(context.Background(), CodexSessionImportRequest{
		SkipDefaultGroupBind: boolPtr(true),
	}, entries)
	require.NoError(t, err)
	require.Equal(t, accountCount, result.Created)
	require.Zero(t, result.Updated)
	require.Zero(t, result.Skipped)
	require.Zero(t, result.Failed)
	require.Len(t, svc.createdAccounts, accountCount)
	for i, account := range svc.createdAccounts {
		require.Equal(t, "team-shared", account.Credentials["chatgpt_account_id"])
		require.Equal(t, fmt.Sprintf("user-%02d", i+1), account.Credentials["chatgpt_user_id"])
	}
}

func TestCodexAgentIdentitySeenMapPreservesAAfterBForAThenBThenA(t *testing.T) {
	seen := map[string]codexSeenIdentity{}
	keys := buildCodexAgentIdentityKeys("team-shared")
	markCodexIdentitySeen(seen, keys, 1, "user-a")
	require.NotContains(t, seen[keys[0]].entries, codexSeenIdentityEntry{index: 2, userID: "user-b"})
	require.False(t, codexSeenIdentityMatchesForTest(seen, keys, "user-b"))

	markCodexIdentitySeen(seen, keys, 2, "user-b")
	duplicateIndex, ok := firstSeenCodexIdentity(seen, keys, "user-a")
	require.True(t, ok)
	require.Equal(t, 1, duplicateIndex)
}

func TestSanitizeCodexImportCredentialExtrasProtectsAgentIdentityCredentials(t *testing.T) {
	out := sanitizeCodexImportCredentialExtras(map[string]any{
		"agent_runtime_id":  "runtime-override",
		"agent_private_key": "private-key-override",
		"agentPrivateKey":   "camel-private-key-override",
		"task_id":           "task-override",
		"label":             "allowed",
	})
	require.Equal(t, map[string]any{"label": "allowed"}, out)
}

func codexSeenIdentityMatchesForTest(seen map[string]codexSeenIdentity, keys []string, userID string) bool {
	_, ok := firstSeenCodexIdentity(seen, keys, userID)
	return ok
}

func buildAgentIdentityImportValue(t *testing.T, runtimeID, accountID, userID, taskID string) map[string]any {
	t.Helper()
	identity := map[string]any{
		"agent_runtime_id":  runtimeID,
		"agent_private_key": buildAgentIdentityPrivateKey(t),
		"account_id":        accountID,
		"chatgpt_user_id":   userID,
	}
	if taskID != "" {
		identity["task_id"] = taskID
	}
	return map[string]any{
		"auth_mode":      "agentIdentity",
		"agent_identity": identity,
	}
}

func buildAgentIdentityPrivateKey(t *testing.T) string {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(der)
}
