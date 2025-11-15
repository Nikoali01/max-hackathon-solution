package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redis2 "github.com/redis/go-redis/v9"

	"first-max-bot/internal/state"
)

type Repository struct {
	client redis2.Cmdable
	prefix string
	ttl    time.Duration
}

type Option func(*options)

type options struct {
	prefix string
	ttl    time.Duration
}

func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

func New(client redis2.Cmdable, opts ...Option) *Repository {
	cfg := options{
		prefix: "maxbot:user:",
		ttl:    24 * time.Hour,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Repository{
		client: client,
		prefix: cfg.prefix,
		ttl:    cfg.ttl,
	}
}

func (r *Repository) key(userID string) string {
	return fmt.Sprintf("%s%s", r.prefix, userID)
}

func (r *Repository) GetUserState(ctx context.Context, userID string) (*state.UserState, error) {
	raw, err := r.client.Get(ctx, r.key(userID)).Result()
	if err != nil {
		if err == redis2.Nil {
			return nil, nil
		}
		return nil, err
	}

	var st state.UserState
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *Repository) SaveUserState(ctx context.Context, userID string, st state.UserState) error {
	payload, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(userID), payload, r.ttl).Err()
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Repository) Close() error {
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
