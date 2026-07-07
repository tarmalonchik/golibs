package kafka

import (
	"context"
	"errors"
)

type ProcessorFunc func(ctx context.Context, msg []byte, key string) error
type PostProcessorFuncCG func(err error) (commit bool)
type PostProcessorFunc func(err error)

var (
	ErrTopicIsEmpty                  = errors.New("empty topic")
	ErrShouldHaveAtLeastOnePartition = errors.New("should have at least one partition")
)

type Config struct {
	KafkaPassword          string `env:"KAFKA_PASSWORD,required"`
	KafkaUser              string `env:"KAFKA_USER,required"`
	KafkaPort              string `env:"KAFKA_PORT,required"`
	KafkaControllersCount  int    `env:"KAFKA_CONTROLLERS_COUNT,required"`
	KafkaBrokerURLTemplate string `env:"KAFKA_BROKER_URL_TEMPLATE,required"`
	KafkaReplicationFactor int    `env:"KAFKA_REPLICATION_FACTOR" envDefault:"3"`
	KafkaEnableTLS         bool   `env:"KAFKA_ENABLE_TLS" envDefault:"true"`
	KafkaPrefix            string `env:"KAFKA_PREFIX,required"`
}
