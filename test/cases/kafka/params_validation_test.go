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

	_, err := s.Kafka.NewConsumerGroup("", "", 1, false)
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewConsumer("", "", 1, false)
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewSyncProducer(ctx, "", 1, false)
	s.Require().ErrorIs(err, kafka.ErrTopicIsEmpty)

	_, err = s.Kafka.NewConsumerGroup("a", "", 0, false)
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)

	_, err = s.Kafka.NewConsumer("a", "", 0, false)
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)

	_, err = s.Kafka.NewSyncProducer(ctx, "a", 0, false)
	s.Require().ErrorIs(err, kafka.ErrShouldHaveAtLeastOnePartition)
}
