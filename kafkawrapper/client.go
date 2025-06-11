package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Client interface {
	NewConsumer(ctx context.Context, topic string, key string, numPartitions int32, createTopic bool) (Consumer, error)
	NewSyncProducer(topic string, numPartitions int32, createTopic bool) (Producer, error)
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

func getPartitionNumberWithKey(topic string, key string, numPartitions int32) (int32, error) {
	if numPartitions <= 0 {
		return 0, nil
	}

	var saramaKey sarama.ByteEncoder
	if key != "" {
		saramaKey = []byte(key)
	}
	part := sarama.NewHashPartitioner(topic)
	partNum, err := part.Partition(&sarama.ProducerMessage{
		Key: saramaKey,
	}, numPartitions)
	if err != nil {
		return 0, trace.FuncNameWithErrorMsg(err, "defining partition")
	}
	return partNum, nil
}

func (c *client) createTopic(brokers []string, topic string, numPartitions int32) error {
	adm, err := sarama.NewClusterAdmin(brokers, c.client.Config())
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "creating kafka admin")
	}
	err = adm.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: 3,
	}, false)
	if err != nil {
		c.logger.Errorf(trace.FuncNameWithError(err), "deleting topic")
	}
	return nil
}
