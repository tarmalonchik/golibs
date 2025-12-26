package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"github.com/tarmalonchik/golibs/test/basetest"
)

type OneConsumeTestSuite struct {
	basetest.Suite
}

func TestOneConsume(t *testing.T) {
	suite.Run(t, new(OneConsumeTestSuite))
}

func (s *OneConsumeTestSuite) TestOnceConsume() {
	basetest.RunWithTimeout(s.Require(), 1*time.Second, func() {
		ctx := context.Background()

		topic := lo.RandomString(10, alphabet)

		msg := []byte("kuku")

		p, err := s.Kafka.NewSyncProducer(ctx, topic, 1, true)
		s.Require().NoError(err)

		err = p.SendMessage(msg, "")
		s.Require().NoError(err)

		c, err := s.Kafka.NewConsumer(ctx, topic, "", 1, true)
		s.Require().NoError(err)
		c.ReadOnlyOne()

		ch := make(chan struct{})

		err = c.Process(
			func(ctx context.Context, readMsg []byte, key string) error {
				s.Require().Equal(msg, readMsg)
				close(ch)
				return nil
			},
			func(err error) {},
		)
		s.Require().NoError(err)
		<-ch
	})
}
