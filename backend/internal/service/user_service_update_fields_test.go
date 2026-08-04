package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// 这些用例锁死"每个入口只声明自己真正要改的列"：
// 任何退回整行回写的改动都会让并发写入被陈旧快照覆盖，并在这里变红。

type updateFieldsUserRepoStub struct {
	UserRepository
	user         *User
	updateFields []UserUpdateFields
	avatarWrites int
}

func (s *updateFieldsUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	clone := *s.user
	return &clone, nil
}

func (s *updateFieldsUserRepoStub) Update(_ context.Context, user *User, fields UserUpdateFields) error {
	clone := *user
	s.user = &clone
	s.updateFields = append(s.updateFields, fields)
	return nil
}

func (s *updateFieldsUserRepoStub) UpsertUserAvatar(_ context.Context, _ int64, input UpsertUserAvatarInput) (*UserAvatar, error) {
	s.avatarWrites++
	return &UserAvatar{
		StorageProvider: input.StorageProvider,
		StorageKey:      input.StorageKey,
		URL:             input.URL,
		ContentType:     input.ContentType,
		ByteSize:        input.ByteSize,
		SHA256:          input.SHA256,
	}, nil
}

func updateFieldsBoolPtr(value bool) *bool { return &value }

func updateFieldsFloat64Ptr(value float64) *float64 { return &value }

func TestUpdateProfile_OnlyDeclaresRequestedColumns(t *testing.T) {
	username := "renamed"
	tests := []struct {
		name string
		req  UpdateProfileRequest
		want UserUpdateFields
	}{
		{
			name: "username only",
			req:  UpdateProfileRequest{Username: &username},
			want: UserUpdateFields{Username: true},
		},
		{
			name: "notify settings only",
			req:  UpdateProfileRequest{BalanceNotifyEnabled: updateFieldsBoolPtr(true)},
			want: UserUpdateFields{BalanceNotifySettings: true},
		},
		{
			name: "username and notify threshold",
			req:  UpdateProfileRequest{Username: &username, BalanceNotifyThreshold: updateFieldsFloat64Ptr(1.5)},
			want: UserUpdateFields{Username: true, BalanceNotifySettings: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &updateFieldsUserRepoStub{user: &User{ID: 7, Balance: 0.30, Status: StatusActive}}
			svc := NewUserService(repo, nil, nil, nil)

			_, err := svc.UpdateProfile(context.Background(), 7, tt.req)
			require.NoError(t, err)
			require.Equal(t, []UserUpdateFields{tt.want}, repo.updateFields)
		})
	}
}

// 只改头像时用户行没有任何列要写，不应产生一次整行更新。
func TestUpdateProfile_AvatarOnlySkipsUserRowWrite(t *testing.T) {
	repo := &updateFieldsUserRepoStub{user: &User{ID: 7, Balance: 0.30}}
	svc := NewUserService(repo, nil, nil, nil)

	avatar := "https://cdn.example.com/a.png"
	_, err := svc.UpdateProfile(context.Background(), 7, UpdateProfileRequest{AvatarURL: &avatar})
	require.NoError(t, err)
	require.Equal(t, 1, repo.avatarWrites, "avatar must still be stored")
	require.Equal(t, []UserUpdateFields{{}}, repo.updateFields, "no user column should be declared")
}

func TestChangePassword_OnlyDeclaresPasswordHash(t *testing.T) {
	user := &User{ID: 7, Balance: 0.30}
	require.NoError(t, user.SetPassword("old-password"))
	repo := &updateFieldsUserRepoStub{user: user}
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.ChangePassword(context.Background(), 7, ChangePasswordRequest{
		CurrentPassword: "old-password",
		NewPassword:     "new-password",
	})
	require.NoError(t, err)
	require.Equal(t, []UserUpdateFields{{PasswordHash: true}}, repo.updateFields)
}

func TestUpdateStatus_OnlyDeclaresStatus(t *testing.T) {
	repo := &updateFieldsUserRepoStub{user: &User{ID: 7, Balance: 0.30, Status: StatusActive}}
	svc := NewUserService(repo, nil, nil, nil)

	require.NoError(t, svc.UpdateStatus(context.Background(), 7, StatusDisabled))
	require.Equal(t, []UserUpdateFields{{Status: true}}, repo.updateFields)
}
