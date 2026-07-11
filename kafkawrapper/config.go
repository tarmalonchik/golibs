package kafka

import (
	"context"
	"errors"
)

type ProcessorFunc func(ctx context.Context, msg []byte, key string) error
type PostProcessorFuncCG func(ctx context.Context, err error) (commit bool)
type PostProcessorFunc func(ctx context.Context, err error)

var (
	ErrTopicIsEmpty                  = errors.New("empty topic")
	ErrShouldHaveAtLeastOnePartition = errors.New("should have at least one partition")
)

type Config struct {
	KafkaPassword          string `env:"PASSWORD,required"`
	KafkaUser              string `env:"USER,required"`
	KafkaPort              string `env:"PORT,required"`
	KafkaControllersCount  int    `env:"CONTROLLERS_COUNT,required"`
	KafkaBrokerURLTemplate string `env:"BROKER_URL_TEMPLATE,required"`
	KafkaReplicationFactor int    `env:"REPLICATION_FACTOR" envDefault:"3"`
	KafkaEnableTLS         bool   `env:"ENABLE_TLS" envDefault:"true"`
	KafkaPrefix            string `env:"PREFIX,required"`
}
