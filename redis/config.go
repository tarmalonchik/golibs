package redis

type Config struct {
	RedisAddress   string `envconfig:"REDIS_ADDRESS" default:"127.0.0.1"`
	RedisPort      string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword  string `envconfig:"REDIS_PASSWORD" default:""`
	RedisKeyPrefix string `envconfig:"REDIS_KEY_PREFIX" required:"true"`
}
