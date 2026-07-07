package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	kafka "github.com/tarmalonchik/golibs/kafkawrapper"
	"github.com/tarmalonchik/golibs/test/basetest"
)

type ValidationTestSuite struct {
	basetest.Suite
}

func TestValidation(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

func (s *ValidationTestSuite) TestValidation() {
	ctx := context.Background()

	_, err := s.Kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		Topic:         "",
		NumPartitions: 1,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewConsumer(kafka.ConsumerConfig{
		Topic:         "",
		NumPartitions: 1,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewSyncProducer(ctx, kafka.ProducerConfig{
		Topic:         "",
		NumPartitions: 1,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		Topic:         "a",
		NumPartitions: 0,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)

	_, err = s.Kafka.NewConsumer(kafka.ConsumerConfig{
		Topic:         "a",
		NumPartitions: 0,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)

	_, err = s.Kafka.NewSyncProducer(ctx, kafka.ProducerConfig{
		Topic:         "a",
		NumPartitions: 0,
		CreateTopic:   false,
	})
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)
}
