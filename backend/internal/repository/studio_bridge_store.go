package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

type studioBridgeStore struct {
	rdb *redis.Client
}

func NewStudioBridgeStore(rdb *redis.Client) service.StudioBridgeStore {
	return &studioBridgeStore{rdb: rdb}
}

func (s *studioBridgeStore) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, payload, ttl).Err()
}

func (s *studioBridgeStore) GetDel(ctx context.Context, key string) ([]byte, bool, error) {
	raw, err := s.rdb.GetDel(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}
