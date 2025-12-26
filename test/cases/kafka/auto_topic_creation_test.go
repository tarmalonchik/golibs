package kafka

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
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

	p, err := s.Kafka.NewSyncProducer(ctx, lo.RandomString(10, alphabet), 1, true)
	s.Require().NoError(err)

	err = p.SendMessage([]byte("hello world"), "")
	s.Require().NoError(err)

	p.Close()
}

func (s *AutoTopicCreateTestSuite) TestConsumer() {
	ctx := context.Background()

	c, err := s.Kafka.NewConsumer(ctx, lo.RandomString(10, alphabet), "", 1, true)
	s.Require().NoError(err)

	c.Close()
}

func (s *AutoTopicCreateTestSuite) TestConsumerGroup() {
	ctx := context.Background()

	c, err := s.Kafka.NewConsumerGroup(ctx, lo.RandomString(10, alphabet), "abc", 1, true)
	s.Require().NoError(err)

	c.Close()
}
