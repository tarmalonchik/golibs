package kafka

import (
	"context"
)

const (
	brokerTemplate = "%s:%s"
)

type ProcessorFunc func(ctx context.Context, msg []byte) error
type PostProcessorFuncCG func(err error) (commit bool)
type PostProcessorFunc func(err error)

type Config struct {
	KafkaPassword          string `envconfig:"KAFKA_PASSWORD" required:"true"`
	KafkaUser              string `envconfig:"KAFKA_USER" required:"true"`
	KafkaPort              string `envconfig:"KAFKA_PORT" required:"true"`
	KafkaControllersCount  int    `envconfig:"KAFKA_CONTROLLERS_COUNT" required:"true"`
	KafkaBrokerURLTemplate string `envconfig:"KAFKA_BROKER_URL_TEMPLATE" required:"true"`
}
