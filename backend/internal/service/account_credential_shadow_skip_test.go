package service

import (
	"context"
	"testing"
)

type shadowCredentialRepo struct {
	AccountRepository
	accounts map[int64]*Account
}

func (r *shadowCredentialRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	if account, ok := r.accounts[id]; ok {
		return account, nil
	}
	return nil, ErrAccountNotFound
}

func TestResolveCredentialAccountShadowUsesEligibleOpenAIParent(t *testing.T) {
	parentID := int64(100)
	parent := &Account{ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test-parent-token"}}
	repo := &shadowCredentialRepo{accounts: map[int64]*Account{parentID: parent}}
	child := &Account{ID: 101, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: &parentID}

	resolved, err := resolveCredentialAccount(context.Background(), repo, child)
	if err != nil {
		t.Fatalf("resolve eligible parent: %v", err)
	}
	if resolved != parent {
		t.Fatal("credential resolver did not return the parent account")
	}
}

func TestResolveCredentialAccountShadowFailsClosedForShadowParent(t *testing.T) {
	parentID := int64(100)
	grandparentID := int64(99)
	repo := &shadowCredentialRepo{accounts: map[int64]*Account{
		parentID: {ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: &grandparentID},
	}}
	_, err := resolveCredentialAccount(context.Background(), repo, &Account{ID: 101, ParentAccountID: &parentID})
	if err == nil {
		t.Fatal("shadow chain must fail closed")
	}
}

func TestPersistAccountCredentialsSkipsShadow(t *testing.T) {
	parentID := int64(100)
	child := &Account{ID: 101, ParentAccountID: &parentID, Credentials: map[string]any{}}
	repo := &shadowCredentialRepo{}
	if err := persistAccountCredentials(context.Background(), repo, child, map[string]any{"access_token": "test-child-token"}); err != nil {
		t.Fatalf("persist shadow credentials: %v", err)
	}
	if len(child.Credentials) != 0 {
		t.Fatal("shadow credentials must not be stored")
	}
}

func TestOpenAIShadowCredentialUsesParentToken(t *testing.T) {
	parentID := int64(100)
	parent := &Account{ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "test-parent-token"}}
	svc := &OpenAIGatewayService{accountRepo: &shadowCredentialRepo{accounts: map[int64]*Account{parentID: parent}}}
	child := &Account{ID: 101, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: &parentID, Credentials: map[string]any{}}
	token, mode, err := svc.GetAccessToken(context.Background(), child)
	if err != nil {
		t.Fatalf("get shadow token: %v", err)
	}
	if token != "test-parent-token" || mode != "oauth" {
		t.Fatal("shadow did not use its parent OAuth credential")
	}
}
