package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type ConsumerGroup interface {
	Close()
	Process(processorFunc ProcessorFunc, postProcessor PostProcessorFuncCG) error
}

type consumerGroup struct {
	client   sarama.Client
	conGroup sarama.ConsumerGroup
	topic    string
	ctx      context.Context
	cancel   context.CancelFunc
	logger   CustomLogger
	once     func()
}

func (c *client) NewConsumerGroup(ctx context.Context, topic, group string, numPartitions int32, createTopic bool) (ConsumerGroup, error) {
	var out consumerGroup
	var err error

	if topic == "" {
		return nil, errors.New("empty topic")
	}

	if createTopic {
		if err := c.createTopic(c.brokers, topic, numPartitions); err != nil {
			return nil, err
		}
	}

	out.ctx, out.cancel = context.WithCancel(ctx)
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

func (c *consumerGroup) Close() {
	c.cancel()
	c.once()
}

func (c *consumerGroup) Process(processorFunc ProcessorFunc, pp PostProcessorFuncCG) error {
	for {
		select {
		case <-c.ctx.Done():
			c.once()
			if c.logger != nil {
				c.logger.Infof("closing consumer group: %s", c.topic)
			}
			return nil
		default:
			h := handler{
				processor:     processorFunc,
				postProcessor: pp,
				ctx:           c.ctx,
			}
			if err := c.conGroup.Consume(c.ctx, []string{c.topic}, &h); err != nil && c.logger != nil {
				c.logger.Errorf(trace.FuncNameWithError(err), "processing consumer group")
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
	fmt.Println("cha")
	select {
	case <-sess.Context().Done():
		return nil
	default:
		for msg := range claim.Messages() {
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
	return nil
}
