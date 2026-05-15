package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

type openWebUILaunchTokenStore struct {
	rdb *redis.Client
}

func NewOpenWebUILaunchTokenStore(rdb *redis.Client) service.OpenWebUILaunchTokenStore {
	return &openWebUILaunchTokenStore{rdb: rdb}
}

func (s *openWebUILaunchTokenStore) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, payload, ttl).Err()
}

func (s *openWebUILaunchTokenStore) GetDel(ctx context.Context, key string) ([]byte, bool, error) {
	raw, err := s.rdb.GetDel(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}
