package service

import (
	"context"
	"errors"
	"testing"
)

type userErrorSettingRepoStub struct {
	values map[string]string
	err    error
}

func (s *userErrorSettingRepoStub) Get(context.Context, string) (*Setting, error) { return nil, s.err }
func (s *userErrorSettingRepoStub) GetValue(context.Context, string) (string, error) {
	return "", s.err
}
func (s *userErrorSettingRepoStub) Set(context.Context, string, string) error { return nil }
func (s *userErrorSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = s.values[key]
	}
	return out, nil
}
func (s *userErrorSettingRepoStub) SetMultiple(context.Context, map[string]string) error { return nil }
func (s *userErrorSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return nil, nil
}
func (s *userErrorSettingRepoStub) Delete(context.Context, string) error { return nil }

func TestIsUserErrorViewAllowedFailsClosed(t *testing.T) {
	if NewSettingService(&userErrorSettingRepoStub{err: errors.New("store unavailable")}, nil).IsUserErrorViewAllowed(context.Background()) {
		t.Fatal("settings errors must disable user error view")
	}
	if !NewSettingService(&userErrorSettingRepoStub{values: map[string]string{SettingKeyAllowUserViewErrorRequests: "true"}}, nil).IsUserErrorViewAllowed(context.Background()) {
		t.Fatal("explicit true should enable user error view")
	}
	if NewSettingService(&userErrorSettingRepoStub{values: map[string]string{}}, nil).IsUserErrorViewAllowed(context.Background()) {
		t.Fatal("missing setting must default to disabled")
	}
}
