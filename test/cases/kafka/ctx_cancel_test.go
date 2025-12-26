package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"github.com/tarmalonchik/golibs/test/basetest"
)

type CtxCancelTestSuite struct {
	basetest.Suite
}

func TestCtxCancel(t *testing.T) {
	suite.Run(t, new(CtxCancelTestSuite))
}

func (s *CtxCancelTestSuite) TestProducer() {
	basetest.RunWithTimeout(s.Require(), 1*time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		p, err := s.Kafka.NewSyncProducer(ctx, lo.RandomString(10, alphabet), 1, true)
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
		<-ch
	})
}

func (s *CtxCancelTestSuite) TestConsumer() {
	basetest.RunWithTimeout(s.Require(), 1*time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		p, err := s.Kafka.NewConsumer(ctx, lo.RandomString(10, alphabet), "", 1, true)
		s.Require().NoError(err)

		ch := make(chan struct{})

		go func() {
			for {
				err = p.Process(func(ctx context.Context, msg []byte, key string) error {
					return nil
				}, func(err error) {})
				s.Require().NoError(err)
				break
			}
			close(ch)
		}()

		<-ch
	})
}

func (s *CtxCancelTestSuite) TestConsumerGroup() {
	basetest.RunWithTimeout(s.Require(), 10*time.Second, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		p, err := s.Kafka.NewConsumerGroup(ctx, lo.RandomString(10, alphabet), "test", 1, true)
		s.Require().NoError(err)

		ch := make(chan struct{})

		go func() {
			for {
				err = p.Process(func(ctx context.Context, msg []byte, key string) error {
					return nil
				}, func(err error) bool {
					return true
				})
				s.Require().NoError(err)
				break
			}
			close(ch)
		}()

		<-ch
	})
}
