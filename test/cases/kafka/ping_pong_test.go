package kafka

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"github.com/tarmalonchik/golibs/test/basetest"
)

type PingPongTestSuite struct {
	basetest.Suite
}

func TestPingPong(t *testing.T) {
	suite.Run(t, new(PingPongTestSuite))
}

func (s *PingPongTestSuite) TestPingPong() {
	mx := sync.Mutex{}

	basetest.RunWithTimeout(s.Require(), 1*time.Second, func() {
		ctx := context.Background()

		topic := lo.RandomString(10, alphabet)

		p, err := s.Kafka.NewSyncProducer(ctx, topic, 1, true)
		s.Require().NoError(err)

		c, err := s.Kafka.NewConsumer(ctx, topic, "", 1, true)
		s.Require().NoError(err)

		mp := make(map[int]struct{})

		ch1 := make(chan struct{})
		ch2 := make(chan struct{})

		go func() {
			for i := range 100 {
				mx.Lock()
				mp[i] = struct{}{}
				mx.Unlock()
				err = p.SendMessage([]byte(strconv.Itoa(i)), "")
				s.Require().NoError(err)
			}
			ch1 <- struct{}{}
		}()

		go func() {
			err = c.Process(func(ctx context.Context, msg []byte, key string) error {
				mx.Lock()
				num, err := strconv.Atoi(string(msg))
				s.Require().NoError(err)

				_, ok := mp[num]
				s.Require().True(ok)
				delete(mp, num)

				if len(mp) == 0 {
					close(ch2)
				}
				mx.Unlock()

				return nil
			}, func(err error) {})

			s.Require().NoError(err)
		}()

		<-ch1
		<-ch2
	})
}
