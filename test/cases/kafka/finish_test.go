package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	kafka "github.com/tarmalonchik/golibs/kafkawrapper"
	"github.com/tarmalonchik/golibs/test/basetest"
)

type FinishTestSuite struct {
	basetest.Suite
}

func TestFinish(t *testing.T) {
	suite.Run(t, new(FinishTestSuite))
}

func (s *FinishTestSuite) TestProducerFinish() {
	basetest.RunWithTimeout(s.Require(), time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		p, err := s.Kafka.NewSyncProducer(ctx, kafka.ProducerConfig{
			Topic:         lo.RandomString(10, alphabet),
			NumPartitions: 1,
			CreateTopic:   true,
		})
		s.Require().NoError(err)

		ch := make(chan struct{})

		go func() {
			for {
				err = p.SendMessage([]byte(lo.RandomString(10, alphabet)), "")
				if errors.Is(err, context.Canceled) {
					break
				}
				s.Require().NoError(err)
				time.Sleep(100 * time.Millisecond)
			}
			close(ch)
		}()

		time.Sleep(300 * time.Millisecond)
		p.Close()
		<-ch
	})
}

func (s *FinishTestSuite) TestConsumerFinish() {
	basetest.RunWithTimeout(s.Require(), time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		p, err := s.Kafka.NewConsumer(kafka.ConsumerConfig{
			Topic:         lo.RandomString(10, alphabet),
			NumPartitions: 1,
			CreateTopic:   true,
		})
		s.Require().NoError(err)

		ch := make(chan struct{})

		go func() {
			for {
				err = p.Process(ctx, func(ctx context.Context, msg []byte, key string) error {
					return nil
				}, func(ctx context.Context, err error) {})
				s.Require().NoError(err)
				break
			}
			close(ch)
		}()

		time.Sleep(300 * time.Millisecond)
		_ = p.Close()
		<-ch
	})

}

func (s *FinishTestSuite) TestConsumerGroupFinish() {
	basetest.RunWithTimeout(s.Require(), 10*time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		p, err := s.Kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
			Topic:         lo.RandomString(10, alphabet),
			Group:         "test",
			NumPartitions: 1,
			CreateTopic:   true,
		})
		s.Require().NoError(err)

		ch := make(chan struct{})

		go func() {
			for {
				err = p.Process(ctx, func(ctx context.Context, msg []byte, key string) error {
					return nil
				}, func(ctx context.Context, err error) bool {
					return true
				})
				s.Require().NoError(err)
				break
			}
			close(ch)
		}()

		time.Sleep(300 * time.Millisecond)
		p.Close()
		<-ch
	})
}
