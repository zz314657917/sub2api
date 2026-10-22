package admin

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCodexV0217SameUserDifferentWorkspaces(t *testing.T) {
	svc := newCodexImportMemoryAdminService(nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	req := CodexSessionImportRequest{SkipDefaultGroupBind: boolPtr(true)}
	result, err := handler.importCodexSessions(context.Background(), req, []codexImportEntry{
		{Index: 1, Value: buildCodexRefreshImportValue(t, "workspace-1", "shared-user", "refresh-1")},
		{Index: 2, Value: buildCodexRefreshImportValue(t, "workspace-2", "shared-user", "refresh-2")},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Created, "batch dedup must include workspace")
	require.Zero(t, result.Updated)
	require.Zero(t, result.Skipped)
	require.Zero(t, result.Failed)
	require.Len(t, svc.createdAccounts, 2)
	require.Empty(t, svc.updatedAccounts)

	// A later request must update only the identity in the selected workspace.
	result, err = handler.importCodexSessions(context.Background(), req, []codexImportEntry{
		{Index: 1, Value: buildCodexRefreshImportValue(t, "workspace-2", "shared-user", "refresh-updated")},
	})
	require.NoError(t, err)
	require.Zero(t, result.Created)
	require.Equal(t, 1, result.Updated)
	require.Zero(t, result.Failed)
	require.Len(t, svc.updatedAccounts, 1)
	require.EqualValues(t, 101, svc.updatedAccounts[0].id)
	require.Equal(t, "refresh-1", svc.accounts[0].Credentials["refresh_token"], "other workspace must remain untouched")
	require.Equal(t, "refresh-updated", svc.accounts[1].Credentials["refresh_token"])
}

func TestCodexV0217SameWorkspaceDifferentUsersAndIdempotency(t *testing.T) {
	svc := newCodexImportMemoryAdminService(nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	req := CodexSessionImportRequest{SkipDefaultGroupBind: boolPtr(true)}
	value := buildCodexRefreshImportValue(t, "workspace", "user-1", "refresh-1")
	result, err := handler.importCodexSessions(context.Background(), req, []codexImportEntry{
		{Index: 1, Value: value},
		{Index: 2, Value: buildCodexRefreshImportValue(t, "workspace", "user-2", "refresh-2")},
		{Index: 3, Value: value},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Created)
	require.Zero(t, result.Updated)
	require.Equal(t, 1, result.Skipped, "same composite identity remains idempotent within a batch")
	require.Zero(t, result.Failed)
	result, err = handler.importCodexSessions(context.Background(), req, []codexImportEntry{{Index: 1, Value: value}})
	require.NoError(t, err)
	require.Zero(t, result.Created)
	require.Equal(t, 1, result.Updated)
	require.Zero(t, result.Failed)
	require.Len(t, svc.createdAccounts, 2)
	require.Len(t, svc.updatedAccounts, 1)
	require.EqualValues(t, 100, svc.updatedAccounts[0].id)
	require.Equal(t, "refresh-2", svc.accounts[1].Credentials["refresh_token"])
}

func TestCodexV0217CompositeIdentityEncodingAndWeakFallback(t *testing.T) {
	left := buildCodexImportIdentityKeys("ab:c", "shared-user", "", "", "refresh")
	right := buildCodexImportIdentityKeys("ab", "c:shared-user", "", "", "refresh")
	require.NotEqual(t, left[0], right[0], "delimiter-containing identifiers must not collide")
	require.Equal(t, buildCodexStoredIdentityKeys(" workspace ", " user ", "", ""),
		buildCodexStoredIdentityKeys("workspace", "user", "", ""))
	require.Equal(t, []string{"user:user"}, buildCodexStoredIdentityKeys("", "user", "", ""))
	require.Equal(t, []string{"email:a@example.com"}, buildCodexStoredIdentityKeys("", "", " A@EXAMPLE.COM ", ""))
	require.Equal(t, []string{"account:workspace"}, buildCodexStoredIdentityKeys("workspace", "", "", ""))

	for _, tc := range []struct {
		name, storedUser, incomingUser string
	}{
		{"stored_user_missing", "", "user"},
		{"incoming_user_missing", "user", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := service.Account{ID: 1, Credentials: map[string]any{
				"chatgpt_account_id": "workspace", "chatgpt_user_id": tc.storedUser,
			}}
			index := buildCodexAccountIndex([]service.Account{account})
			found, key := index.Find(buildCodexImportIdentityKeys("workspace", tc.incomingUser, "", "new-token", "refresh"), tc.incomingUser)
			require.NotNil(t, found)
			require.EqualValues(t, 1, found.ID)
			require.Equal(t, "account:workspace", key, "weak account fallback must remain available")
		})
	}
}

func TestCodexV0217TokenOnlyUpgradeIsWorkspaceScoped(t *testing.T) {
	token1 := buildCodexAccessToken(t, "workspace-1", "shared-user", time.Now().Add(time.Hour))
	token2 := buildCodexAccessToken(t, "workspace-2", "shared-user", time.Now().Add(time.Hour))
	svc := newCodexImportMemoryAdminService([]service.Account{
		{ID: 11, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Credentials: map[string]any{
			"chatgpt_account_id": "workspace-1", "chatgpt_user_id": "shared-user", "access_token": token1,
		}},
		{ID: 12, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Credentials: map[string]any{
			"chatgpt_account_id": "workspace-2", "chatgpt_user_id": "shared-user", "access_token": token2,
		}},
	})
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	result, err := handler.importCodexSessions(context.Background(), CodexSessionImportRequest{SkipDefaultGroupBind: boolPtr(true)}, []codexImportEntry{
		{Index: 1, Value: buildCodexRefreshImportValue(t, "workspace-2", "shared-user", "refresh-upgrade")},
	})
	require.NoError(t, err)
	require.Zero(t, result.Created)
	require.Equal(t, 1, result.Updated)
	require.Zero(t, result.Failed)
	require.Len(t, svc.updatedAccounts, 1)
	require.EqualValues(t, 12, svc.updatedAccounts[0].id)
	require.Equal(t, "refresh-upgrade", svc.accounts[1].Credentials["refresh_token"])
	require.NotContains(t, svc.accounts[0].Credentials, "refresh_token")
	require.Equal(t, []string{"access:" + codexTokenFingerprint(token1)},
		buildCodexImportIdentityKeys("workspace-1", "shared-user", "", token1, ""))
}
