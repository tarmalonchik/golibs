package kafka

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	kafka "github.com/tarmalonchik/golibs/kafkawrapper"
	"github.com/tarmalonchik/golibs/test/basetest"
)

var alphabet = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
}

type AutoTopicCreateTestSuite struct {
	basetest.Suite
}

func TestAutoTopicCreate(t *testing.T) {
	suite.Run(t, new(AutoTopicCreateTestSuite))
}

func (s *AutoTopicCreateTestSuite) TestProducer() {
	ctx := context.Background()

	p, err := s.Kafka.NewSyncProducer(ctx, kafka.ProducerConfig{
		Topic:         lo.RandomString(10, alphabet),
		NumPartitions: 1,
		CreateTopic:   true,
	})
	s.Require().NoError(err)

	err = p.SendMessage([]byte("hello world"), "")
	s.Require().NoError(err)

	p.Close()
}

func (s *AutoTopicCreateTestSuite) TestConsumer() {
	c, err := s.Kafka.NewConsumer(kafka.ConsumerConfig{
		Topic:         lo.RandomString(10, alphabet),
		NumPartitions: 1,
		CreateTopic:   true,
	})
	s.Require().NoError(err)

	_ = c.Close()
}

func (s *AutoTopicCreateTestSuite) TestConsumerGroup() {
	c, err := s.Kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		Topic:         lo.RandomString(10, alphabet),
		Group:         "abc",
		NumPartitions: 1,
		CreateTopic:   true,
	})
	s.Require().NoError(err)

	_ = c.Close()
}
