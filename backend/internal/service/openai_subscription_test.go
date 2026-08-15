package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func TestFetchChatGPTSubscriptionExpiresAt(t *testing.T) {
	const wantExpiresAt = "2026-06-10T02:52:15Z"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/subscriptions", r.URL.Path)
		require.Equal(t, "acc_123", r.URL.Query().Get("account_id"))
		require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"plan_type":    "plus",
			"active_until": wantExpiresAt,
			"will_renew":   true,
			"id":           "sub_123",
		})
	}))
	defer server.Close()

	oldURL := chatGPTSubscriptionsURL
	chatGPTSubscriptionsURL = server.URL + "/backend-api/subscriptions"
	t.Cleanup(func() { chatGPTSubscriptionsURL = oldURL })

	got := fetchChatGPTSubscriptionExpiresAt(context.Background(), func(proxyURL string) (*req.Client, error) {
		return req.C().SetTimeout(5 * time.Second), nil
	}, "access-token", "", "acc_123")

	require.Equal(t, wantExpiresAt, got)
}

func TestFetchChatGPTAccountInfo_SkipsExpiredWorkspaceCandidate(t *testing.T) {
	expiredAt := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/accounts/check/v4-2023-04-27", r.URL.Path)
		require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accounts": map[string]any{
				"org-expired-workspace": map[string]any{
					"account": map[string]any{
						"plan_type":  "self_serve_business_usage_based",
						"is_default": true,
					},
					"entitlement": map[string]any{
						"expires_at": expiredAt,
					},
				},
				"personal-account": map[string]any{
					"account": map[string]any{
						"plan_type": "free",
					},
				},
			},
		})
	}))
	defer server.Close()

	oldURL := chatGPTAccountsCheckURL
	chatGPTAccountsCheckURL = server.URL + "/backend-api/accounts/check/v4-2023-04-27"
	t.Cleanup(func() { chatGPTAccountsCheckURL = oldURL })

	got := fetchChatGPTAccountInfo(context.Background(), func(proxyURL string) (*req.Client, error) {
		return req.C().SetTimeout(5 * time.Second), nil
	}, "access-token", "", "org-expired-workspace")

	require.NotNil(t, got)
	require.Equal(t, "free", got.PlanType)
	require.Empty(t, got.SubscriptionExpiresAt)
}

func TestFetchChatGPTAccountInfo_SkipsDeactivatedWorkspaceCandidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/accounts/check/v4-2023-04-27", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accounts": map[string]any{
				"org-deactivated-workspace": map[string]any{
					"account": map[string]any{
						"plan_type":      "self_serve_business_usage_based",
						"is_default":     true,
						"is_deactivated": true,
					},
				},
				"personal-account": map[string]any{
					"account": map[string]any{
						"plan_type": "pro",
					},
				},
			},
		})
	}))
	defer server.Close()

	oldURL := chatGPTAccountsCheckURL
	chatGPTAccountsCheckURL = server.URL + "/backend-api/accounts/check/v4-2023-04-27"
	t.Cleanup(func() { chatGPTAccountsCheckURL = oldURL })

	got := fetchChatGPTAccountInfo(context.Background(), func(proxyURL string) (*req.Client, error) {
		return req.C().SetTimeout(5 * time.Second), nil
	}, "access-token", "", "org-deactivated-workspace")

	require.NotNil(t, got)
	require.Equal(t, "pro", got.PlanType)
}

func TestShouldApplyChatGPTAccountInfoPlanType(t *testing.T) {
	require.False(t, shouldApplyChatGPTAccountInfoPlanType("k12", "self_serve_business_usage_based"))
	require.False(t, shouldApplyChatGPTAccountInfoPlanType("free", "team"))
	require.False(t, shouldApplyChatGPTAccountInfoPlanType("", ""))
	require.True(t, shouldApplyChatGPTAccountInfoPlanType("", "pro"))
}

func TestFetchChatGPTAccountInfo_ReportsAccountID(t *testing.T) {
	acct := map[string]any{"account": map[string]any{"account_id": "personal-account", "plan_type": "pro"}}
	info := &ChatGPTAccountInfo{}
	fillAccountInfo(info, acct, "default")
	require.Equal(t, "personal-account", info.AccountID)
}

func TestFetchChatGPTAccountInfo_WorkspaceEntitlementDoesNotOverridePersonalSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accounts": map[string]any{
				"workspace-account": map[string]any{
					"account":     map[string]any{"account_id": "workspace-account", "plan_type": "pro"},
					"entitlement": map[string]any{"expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339)},
				},
			},
		})
	}))
	defer server.Close()
	oldURL := chatGPTAccountsCheckURL
	chatGPTAccountsCheckURL = server.URL
	t.Cleanup(func() { chatGPTAccountsCheckURL = oldURL })

	info := fetchChatGPTAccountInfo(context.Background(), func(string) (*req.Client, error) {
		return req.C().SetTimeout(time.Second), nil
	}, "access-token", "", "workspace-account")
	require.NotNil(t, info)
	require.Equal(t, "workspace-account", info.AccountID)
	require.False(t, chatGPTAccountInfoBelongsToTokenAccount(
		&OpenAITokenInfo{ChatGPTAccountID: "personal-account", PlanType: "pro"}, info))
}

func TestEnrichTokenInfo_WorkspaceEntitlementDoesNotOverridePersonalSubscription(t *testing.T) {
	const personalExpiry = "2027-03-01T00:00:00Z"
	var subscriptionAccountIDs []string
	server := newS217SubscriptionFixture(t, map[string]any{
		"accounts": map[string]any{
			"workspace-account": map[string]any{
				"account":     map[string]any{"account_id": "workspace-account", "plan_type": "pro"},
				"entitlement": map[string]any{"expires_at": "2026-04-01T00:00:00Z"},
			},
		},
	}, personalExpiry, &subscriptionAccountIDs)
	defer server.Close()

	tokenInfo := &OpenAITokenInfo{
		AccessToken:      "access-token",
		ChatGPTAccountID: "personal-account",
		OrganizationID:   "workspace-account",
		PlanType:         "pro",
	}
	(&OpenAIOAuthService{privacyClientFactory: newS217LocalPrivacyClientFactory()}).enrichTokenInfo(context.Background(), tokenInfo, "")

	require.Equal(t, []string{"personal-account"}, subscriptionAccountIDs)
	require.Equal(t, personalExpiry, tokenInfo.SubscriptionExpiresAt)
}

func TestEnrichTokenInfo_SameAccountDoesNotRepeatPersonalSubscriptionLookup(t *testing.T) {
	const entitlementExpiry = "2027-04-01T00:00:00Z"
	var subscriptionAccountIDs []string
	server := newS217SubscriptionFixture(t, map[string]any{
		"accounts": map[string]any{
			"personal-account": map[string]any{
				"account":     map[string]any{"account_id": "PERSONAL-ACCOUNT", "plan_type": "pro"},
				"entitlement": map[string]any{"expires_at": entitlementExpiry},
			},
		},
	}, "", &subscriptionAccountIDs)
	defer server.Close()

	tokenInfo := &OpenAITokenInfo{AccessToken: "access-token", ChatGPTAccountID: "personal-account", OrganizationID: "personal-account", PlanType: "pro"}
	(&OpenAIOAuthService{privacyClientFactory: newS217LocalPrivacyClientFactory()}).enrichTokenInfo(context.Background(), tokenInfo, "")

	require.Empty(t, subscriptionAccountIDs)
	require.Equal(t, entitlementExpiry, tokenInfo.SubscriptionExpiresAt)
}

func TestEnrichTokenInfo_MissingAccountIDPreservesCompatibilityFallback(t *testing.T) {
	const entitlementExpiry = "2027-05-01T00:00:00Z"
	var subscriptionAccountIDs []string
	server := newS217SubscriptionFixture(t, map[string]any{
		"accounts": map[string]any{
			"workspace-account": map[string]any{
				"account":     map[string]any{"account_id": "workspace-account", "plan_type": "pro"},
				"entitlement": map[string]any{"expires_at": entitlementExpiry},
			},
		},
	}, "", &subscriptionAccountIDs)
	defer server.Close()

	tokenInfo := &OpenAITokenInfo{AccessToken: "access-token", OrganizationID: "workspace-account", PlanType: "pro"}
	(&OpenAIOAuthService{privacyClientFactory: newS217LocalPrivacyClientFactory()}).enrichTokenInfo(context.Background(), tokenInfo, "")

	require.Empty(t, subscriptionAccountIDs)
	require.Equal(t, entitlementExpiry, tokenInfo.SubscriptionExpiresAt)
}

func newS217LocalPrivacyClientFactory() PrivacyClientFactory {
	return func(string) (*req.Client, error) { return req.C().SetTimeout(time.Second), nil }
}

// newS217SubscriptionFixture redirects every enrichTokenInfo HTTP request,
// including the best-effort privacy PATCH, to this local server.
func newS217SubscriptionFixture(t *testing.T, accounts map[string]any, activeUntil string, subscriptionAccountIDs *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/backend-api/accounts/check/v4-2023-04-27":
			_ = json.NewEncoder(w).Encode(accounts)
		case "/backend-api/subscriptions":
			*subscriptionAccountIDs = append(*subscriptionAccountIDs, r.URL.Query().Get("account_id"))
			_ = json.NewEncoder(w).Encode(map[string]any{"active_until": activeUntil})
		case "/backend-api/settings/account_user_setting":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	oldSettings, oldAccounts, oldSubscriptions := openAISettingsURL, chatGPTAccountsCheckURL, chatGPTSubscriptionsURL
	openAISettingsURL = server.URL + "/backend-api/settings/account_user_setting"
	chatGPTAccountsCheckURL = server.URL + "/backend-api/accounts/check/v4-2023-04-27"
	chatGPTSubscriptionsURL = server.URL + "/backend-api/subscriptions"
	t.Cleanup(func() {
		openAISettingsURL, chatGPTAccountsCheckURL, chatGPTSubscriptionsURL = oldSettings, oldAccounts, oldSubscriptions
	})
	return server
}
