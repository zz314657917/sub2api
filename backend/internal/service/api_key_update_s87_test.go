package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type s87APIKeyRepo struct {
	APIKeyRepository
	key         *APIKey
	updated     *APIKey
	updateCalls int
}

func (r *s87APIKeyRepo) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *r.key
	clone.IPWhitelist = append([]string(nil), r.key.IPWhitelist...)
	clone.IPBlacklist = append([]string(nil), r.key.IPBlacklist...)
	return &clone, nil
}

func (r *s87APIKeyRepo) Update(_ context.Context, key *APIKey) error {
	clone := *key
	clone.IPWhitelist = append([]string(nil), key.IPWhitelist...)
	clone.IPBlacklist = append([]string(nil), key.IPBlacklist...)
	r.updated = &clone
	r.updateCalls++
	return nil
}

func (r *s87APIKeyRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return []APIKey{*r.key}, &pagination.PaginationResult{Total: 1}, nil
}

func TestS87APIKeyIPRestrictions(t *testing.T) {
	oldWhitelist := []string{"10.0.0.0/8"}
	oldBlacklist := []string{"192.168.1.0/24"}
	empty := []string{}
	newWhitelist := []string{"203.0.113.7"}
	newBlacklist := []string{"198.51.100.0/24"}

	tests := []struct {
		name          string
		req           UpdateAPIKeyRequest
		wantWhitelist []string
		wantBlacklist []string
	}{
		{name: "omitted preserves both", wantWhitelist: oldWhitelist, wantBlacklist: oldBlacklist},
		{name: "clear whitelist only", req: UpdateAPIKeyRequest{IPWhitelist: &empty}, wantWhitelist: empty, wantBlacklist: oldBlacklist},
		{name: "clear blacklist only", req: UpdateAPIKeyRequest{IPBlacklist: &empty}, wantWhitelist: oldWhitelist, wantBlacklist: empty},
		{name: "replace whitelist only", req: UpdateAPIKeyRequest{IPWhitelist: &newWhitelist}, wantWhitelist: newWhitelist, wantBlacklist: oldBlacklist},
		{name: "replace blacklist only", req: UpdateAPIKeyRequest{IPBlacklist: &newBlacklist}, wantWhitelist: oldWhitelist, wantBlacklist: newBlacklist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &s87APIKeyRepo{key: &APIKey{
				ID: 7, UserID: 42, Key: "s87-key", Status: StatusActive,
				IPWhitelist: oldWhitelist, IPBlacklist: oldBlacklist,
			}}
			svc := &APIKeyService{apiKeyRepo: repo}
			updated, err := svc.Update(context.Background(), 7, 42, tt.req)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if updated == nil || repo.updated == nil {
				t.Fatal("expected updated API key")
			}
			if got := updated.IPWhitelist; !equalStrings(got, tt.wantWhitelist) {
				t.Fatalf("whitelist = %#v, want %#v", got, tt.wantWhitelist)
			}
			if got := updated.IPBlacklist; !equalStrings(got, tt.wantBlacklist) {
				t.Fatalf("blacklist = %#v, want %#v", got, tt.wantBlacklist)
			}
			if repo.updateCalls != 1 {
				t.Fatalf("Update repository calls = %d, want 1", repo.updateCalls)
			}
		})
	}

	for _, invalid := range []struct {
		name string
		req  UpdateAPIKeyRequest
	}{
		{name: "invalid whitelist", req: UpdateAPIKeyRequest{IPWhitelist: ptrStrings("not an ip")}},
		{name: "invalid blacklist", req: UpdateAPIKeyRequest{IPBlacklist: ptrStrings("not an ip")}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			repo := &s87APIKeyRepo{key: &APIKey{ID: 7, UserID: 42, Key: "s87-key", Status: StatusActive}}
			svc := &APIKeyService{apiKeyRepo: repo}
			if _, err := svc.Update(context.Background(), 7, 42, invalid.req); err == nil {
				t.Fatal("Update() error = nil, want invalid IP error")
			}
			if repo.updateCalls != 0 {
				t.Fatalf("repository Update calls = %d, want 0", repo.updateCalls)
			}
		})
	}
}

func ptrStrings(values ...string) *[]string { return &values }

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
