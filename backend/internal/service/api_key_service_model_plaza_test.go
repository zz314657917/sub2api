package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type modelPlazaUserRepoStub struct {
	UserRepository
	users map[int64]*User
}

func (r *modelPlazaUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	return r.users[id], nil
}

func TestAPIKeyServiceGetUserAllowedGroupIDSet(t *testing.T) {
	service := NewAPIKeyService(nil, &modelPlazaUserRepoStub{users: map[int64]*User{
		7: {ID: 7, AllowedGroups: []int64{10, 20, 10}},
	}}, nil, nil, nil, nil, &config.Config{})

	allowed, err := service.GetUserAllowedGroupIDSet(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, map[int64]struct{}{10: struct{}{}, 20: struct{}{}}, allowed)
}
