package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/tarmalonchik/golibs/trace"
)

var (
	ErrKeyNotFound  = errors.New("key not found")
	ErrTypeMismatch = errors.New("type mismatch")
)

type Client struct {
	client *redis.Client
	conf   Config
}

func New(conf Config) *Client {
	client := redis.NewClient(&redis.Options{
		MaxRetries:      5,
		MinRetryBackoff: 50 * time.Millisecond,
		MaxRetryBackoff: 5 * time.Second,
		Addr:            fmt.Sprintf("%s:%s", conf.RedisAddress, conf.RedisPort),
		Password:        conf.RedisPassword,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		DialTimeout:     15 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})
	return &Client{
		client: client,
		conf:   conf,
	}
}

func (c *Client) Add(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	key = c.wrapKey(key)

	status := c.client.Set(ctx, key, value, expiration)
	if status.Err() != nil {
		return trace.FuncNameWithError(status.Err())
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	key = c.wrapKey(key)

	status := c.client.Get(ctx, key)
	if status.Err() != nil {
		if errors.Is(status.Err(), redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, trace.FuncNameWithError(status.Err())
	}
	return status.Bytes()
}

func (c *Client) GetInt(ctx context.Context, key string) (int64, error) {
	key = c.wrapKey(key)

	status := c.client.Get(ctx, key)
	if status.Err() != nil {
		if errors.Is(status.Err(), redis.Nil) {
			return 0, ErrKeyNotFound
		}
		return 0, trace.FuncNameWithError(status.Err())
	}
	out, err := status.Int64()
	if err != nil {
		return 0, ErrTypeMismatch
	}

	return out, nil
}

func (c *Client) Del(ctx context.Context, key string) {
	key = c.wrapKey(key)

	_ = c.client.Del(ctx, key)
}

func (c *Client) Dec(ctx context.Context, key string) (int64, error) {
	key = c.wrapKey(key)

	status := c.client.Decr(ctx, key)
	if status.Err() != nil {
		return 0, trace.FuncNameWithError(status.Err())
	}
	return status.Val(), nil
}

func (c *Client) GetValuesByPattern(ctx context.Context, pattern string) (out [][]byte, err error) {
	pattern = c.wrapKey(pattern)

	status := c.client.Keys(ctx, pattern)
	if status.Err() != nil {
		return nil, trace.FuncNameWithError(status.Err())
	}

	if len(status.Val()) == 0 {
		return nil, ErrKeyNotFound
	}

	statusSlice := c.client.MGet(ctx, status.Val()...)
	if statusSlice.Err() != nil {
		return nil, trace.FuncNameWithError(status.Err())
	}

	resp := make([][]byte, 0, len(statusSlice.Val()))
	for _, val := range statusSlice.Val() {
		resp = append(resp, []byte(val.(string)))
	}
	return resp, nil
}

func (c *Client) wrapKey(key string) string {
	if c.conf.RedisKeyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s-%s", c.conf.RedisKeyPrefix, key)
}
