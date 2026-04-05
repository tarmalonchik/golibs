package redis

type Config struct {
	RedisAddress  string `envconfig:"REDIS_HOST" required:"true"`
	RedisPort     string `envconfig:"REDIS_HOST_PORT" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
}
