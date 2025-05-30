package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Client interface {
	NewConsumer(ctx context.Context, topic string, partition int32) (Consumer, error)
	NewSyncProducer(topic string) (Producer, error)
	NewConsumerGroup(ctx context.Context, topic, group string) (ConsumerGroup, error)
}

type CustomLogger interface {
	Errorf(err error, format string, args ...any)
}
type client struct {
	brokers []string
	client  sarama.Client
	logger  CustomLogger
}

func NewClient(conf Config, logger CustomLogger) (Client, error) {
	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.Password = conf.KafkaPassword
	config.Net.SASL.User = conf.KafkaUser
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	var out client
	var err error

	for i := range conf.KafkaControllersCount {
		out.brokers = append(out.brokers, fmt.Sprintf(brokerTemplate, fmt.Sprintf(conf.KafkaBrokerURLTemplate, i), conf.KafkaPort))
	}
	out.client, err = sarama.NewClient(out.brokers, config)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating kafka client")
	}
	out.logger = logger
	return &out, nil
}

func (c *client) Close() error {
	return c.client.Close()
}
