package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionStore interface {
	CreateSession(ctx context.Context, token string, ttl time.Duration) error
	GetSession(ctx context.Context, token string) (bool, error)
	DeleteSession(ctx context.Context, token string) error
}

// RedisSessionStore stores sessions as keys with TTL: session:<token> -> "1"
type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(addr string) (*RedisSessionStore, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		// try as simple addr
		opt = &redis.Options{Addr: addr}
	}
	client := redis.NewClient(opt)
	// quick ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return &RedisSessionStore{client: client}, nil
}

func (r *RedisSessionStore) key(token string) string {
	return "session:" + token
}

func (r *RedisSessionStore) CreateSession(ctx context.Context, token string, ttl time.Duration) error {
	k := r.key(token)
	return r.client.Set(ctx, k, "1", ttl).Err()
}

func (r *RedisSessionStore) GetSession(ctx context.Context, token string) (bool, error) {
	k := r.key(token)
	exists, err := r.client.Exists(ctx, k).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *RedisSessionStore) DeleteSession(ctx context.Context, token string) error {
	k := r.key(token)
	del, err := r.client.Del(ctx, k).Result()
	if err != nil {
		return err
	}
	if del == 0 {
		return ErrSessionNotFound
	}
	return nil
}
