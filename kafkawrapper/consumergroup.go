package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/tarmalonchik/golibs/trace"
)

type ConsumerGroup interface {
	Close() error
	Process(ctx context.Context, processorFunc ProcessorFunc, postProcessor PostProcessorFuncCG) error
}

type consumerGroup struct {
	client   sarama.Client
	conGroup sarama.ConsumerGroup
	topic    string
	logger   CustomLogger
	once     func()

	mu     sync.Mutex
	cancel context.CancelFunc
}

type ConsumerGroupConfig struct {
	Topic         string `env:"TOPIC,required"`
	Group         string `env:"GROUP,required"`
	NumPartitions int    `env:"NUM_PARTITIONS" envDefault:"100"`
	CreateTopic   bool   `env:"CREATE_TOPIC" envDefault:"true"`
}

func (c *client) NewConsumerGroup(config ConsumerGroupConfig) (ConsumerGroup, error) {
	if config.Topic == "" {
		return nil, ErrTopicIsEmpty
	}

	config.Topic = c.wrapTopic(config.Topic)
	config.Group = c.wrapTopic(config.Group)

	var out consumerGroup
	var err error

	if config.NumPartitions < 1 {
		return nil, ErrShouldHaveAtLeastOnePartition
	}

	if config.CreateTopic {
		if err := c.createTopic(c.brokers, config.Topic, int32(config.NumPartitions)); err != nil {
			return nil, err
		}
	}

	cfg := *c.client.Config()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	err = retryer(func() error {
		out.conGroup, err = sarama.NewConsumerGroup(c.brokers, config.Group, &cfg)
		if err != nil {
			return trace.FuncNameWithErrorMsg(err, "creating consumer group")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	out.topic = config.Topic
	out.client = c.client
	out.logger = c.logger
	out.once = sync.OnceFunc(func() {
		_ = out.conGroup.Close()
	})

	return &out, nil
}

func (c *consumerGroup) Close() error {
	c.mu.Lock()
	cancel := c.cancel
	c.mu.Unlock()

	if cancel == nil {
		return nil
	}
	cancel()
	c.once()
	return nil
}

func (c *consumerGroup) Process(ctx context.Context, processorFunc ProcessorFunc, pp PostProcessorFuncCG) error {
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			c.once()
			if c.logger != nil {
				c.logger.Info(fmt.Sprintf("closing consumer group: %s", c.topic))
			}
			return nil
		default:
			h := handler{
				processor:     processorFunc,
				postProcessor: pp,
				ctx:           ctx,
			}
			if err := c.conGroup.Consume(ctx, []string{c.topic}, &h); err != nil {

				if c.logger == nil {
					return nil
				}

				c.logger.Error("processing consumer group", zap.Error(err), zap.String("topic", c.topic))
			}
		}
	}
}

type handler struct {
	processor     ProcessorFunc
	postProcessor PostProcessorFuncCG
	ctx           context.Context
}

func (c handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-sess.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			err := c.processor(c.ctx, msg.Value, string(msg.Key))
			if errors.Is(err, ErrInvalidKey) {
				continue
			}

			if c.postProcessor != nil {
				if commit := c.postProcessor(c.ctx, err); commit {
					sess.MarkMessage(msg, "")
				}
				continue
			}

			if err != nil {
				continue
			}

			sess.MarkMessage(msg, "")
		}
	}
}
