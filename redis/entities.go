package redis

import (
	"errors"
)

var (
	ErrKeyNotFound  = errors.New("key not found")
	ErrTypeMismatch = errors.New("type mismatch")
)

type Config struct {
	RedisAddress   string `env:"REDIS_ADDRESS" envDefault:"127.0.0.1"`
	RedisPort      string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" envDefault:""`
	RedisKeyPrefix string `env:"REDIS_KEY_PREFIX,required"`
	RedisEnableTLS bool   `env:"REDIS_ENABLE_TLS" envDefault:"true"`
}
