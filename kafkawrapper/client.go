package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

var ErrInvalidKey = errors.New("invalid key consumed")

type Client interface {
	NewConsumer(topic string, key string, numPartitions int32, createTopic bool) (Consumer, error)
	NewSyncProducer(ctx context.Context, topic string, numPartitions int32, createTopic bool) (Producer, error)
	NewConsumerGroup(topic, group string, numPartitions int32, createTopic bool) (ConsumerGroup, error)
}

type CustomLogger interface {
	Errorf(err error, format string, args ...any)
	Infof(format string, args ...any)
}
type client struct {
	brokers []string
	client  sarama.Client
	logger  CustomLogger
	conf    Config
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
		if conf.KafkaPort != "" {
			out.brokers = append(out.brokers, strings.Join([]string{fmt.Sprintf(conf.KafkaBrokerURLTemplate, i), conf.KafkaPort}, ":"))
		} else {
			out.brokers = append(out.brokers, fmt.Sprintf(conf.KafkaBrokerURLTemplate, i, i))
		}
	}
	out.client, err = sarama.NewClient(out.brokers, config)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating kafka client")
	}
	out.conf = conf
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
	if err != nil && c.logger != nil {
		c.logger.Errorf(err, "creating cluster admin")
		return nil
	}
	defer func() {
		_ = adm.Close()
	}()

	err = adm.CreateTopic(
		topic,
		&sarama.TopicDetail{
			NumPartitions:     numPartitions,
			ReplicationFactor: c.conf.KafkaReplicationFactor,
		},
		false,
	)
	if err != nil && c.logger != nil {
		if !errors.Is(err, sarama.ErrTopicAlreadyExists) {
			c.logger.Errorf(err, "creating topic name:\"%s\"", topic)
			return err
		}
	}

	return nil
}
