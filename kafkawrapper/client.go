package kafka

import (
	"context"
	"crypto/sha512"
	"crypto/tls"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
	"go.uber.org/zap"

	"github.com/tarmalonchik/golibs/trace"
)

var ErrInvalidKey = errors.New("invalid key consumed")

type Client interface {
	NewConsumer(topic string, key string, numPartitions int32, createTopic bool) (Consumer, error)
	NewSyncProducer(ctx context.Context, topic string, numPartitions int32, createTopic bool) (Producer, error)
	NewConsumerGroup(topic, group string, numPartitions int32, createTopic bool) (ConsumerGroup, error)
}

type CustomLogger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}
type client struct {
	brokers []string
	client  sarama.Client
	logger  CustomLogger
	conf    Config
	prefix  string
}

func (c *client) wrapTopic(key string) string {
	if c.conf.KafkaPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s-%s", c.conf.KafkaUser, key)
}

func NewClient(conf Config, logger CustomLogger) (Client, error) {
	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	config.Net.SASL.Password = conf.KafkaPassword
	config.Net.SASL.User = conf.KafkaUser
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Metadata.Timeout = 10 * time.Second
	config.Producer.Timeout = 5 * time.Second
	if conf.KafkaEnableTLS {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{}
	}
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: func() hash.Hash { return sha512.New() }}
	}

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
	out.prefix = conf.KafkaPrefix
	return &out, nil
}

func (c *client) Close() error {
	return c.client.Close()
}

func (c *client) getTopicPartitionCount(topic string, fallback int32) (int32, error) {
	if err := c.client.RefreshMetadata(topic); err != nil {
		return 0, trace.FuncNameWithErrorMsg(err, "refreshing topic metadata")
	}

	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return 0, trace.FuncNameWithErrorMsg(err, "getting topic partitions")
	}

	actual := int32(len(partitions))
	if actual < 1 {
		return 0, fmt.Errorf("topic %s has no partitions", topic)
	}

	if fallback > 0 && fallback != actual && c.logger != nil {
		c.logger.Info(
			fmt.Sprintf(
				"topic partitions mismatch, use actual count topic:%s expected:%d actual:%d",
				topic,
				fallback,
				actual,
			),
		)
	}

	return actual, nil
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
		if c.logger != nil {
			c.logger.Error("creating cluster admin", zap.Error(err))
		}
		return err
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
			c.logger.Error(fmt.Sprintf("creating topic name:\"%s\"", topic), zap.Error(err))
			return err
		}
	}

	return nil
}

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (s *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	s.Client, err = s.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	s.ClientConversation = s.Client.NewConversation()
	return nil
}

func (s *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	return s.ClientConversation.Step(challenge)
}

func (s *XDGSCRAMClient) Done() bool {
	return s.ClientConversation.Done()
}
