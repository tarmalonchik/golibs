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
	cancel   context.CancelFunc
}

func (c *client) NewConsumerGroup(topic, group string, numPartitions int32, createTopic bool) (ConsumerGroup, error) {
	var out consumerGroup
	var err error

	if topic == "" {
		return nil, ErrTopicIsEmpty
	}

	if numPartitions < 1 {
		return nil, ErrShouldHaveAtLeastOnePartition
	}

	if createTopic {
		if err := c.createTopic(c.brokers, topic, numPartitions); err != nil {
			return nil, err
		}
	}

	out.conGroup, err = sarama.NewConsumerGroup(c.brokers, group, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating consumer group")
	}
	out.topic = topic
	out.client = c.client
	out.logger = c.logger
	out.once = sync.OnceFunc(func() {
		_ = out.conGroup.Close()
	})

	return &out, nil
}

func (c *consumerGroup) Close() error {
	if c.cancel == nil {
		return nil
	}
	c.cancel()
	c.once()
	return nil
}

func (c *consumerGroup) Process(ctx context.Context, processorFunc ProcessorFunc, pp PostProcessorFuncCG) error {
	ctx, c.cancel = context.WithCancel(ctx)

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

				c.logger.Error("processing consumer group", zap.Error(err))
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
				if commit := c.postProcessor(err); commit {
					sess.MarkMessage(msg, "")
				}
				continue
			}

			if err != nil {
				sess.MarkMessage(msg, "")
			}
		}
	}
}
