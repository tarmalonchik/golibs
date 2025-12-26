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
	KafkaPassword          string `envconfig:"KAFKA_PASSWORD" required:"true"`
	KafkaUser              string `envconfig:"KAFKA_USER" required:"true"`
	KafkaPort              string `envconfig:"KAFKA_PORT" required:"true"`
	KafkaControllersCount  int    `envconfig:"KAFKA_CONTROLLERS_COUNT" required:"true"`
	KafkaBrokerURLTemplate string `envconfig:"KAFKA_BROKER_URL_TEMPLATE" required:"true"`
	KafkaReplicationFactor int16  `envconfig:"KAFKA_REPLICATION_FACTOR" default:"3"`
}
