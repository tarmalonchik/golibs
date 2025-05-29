package kafka

import (
	"context"

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
}

func (c *client) NewConsumerGroup(ctx context.Context, topic, group string) (ConsumerGroup, error) {
	var out consumerGroup
	var err error

	out.ctx, out.cancel = context.WithCancel(ctx)
	out.conGroup, err = sarama.NewConsumerGroup(c.brokers, group, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating consumer group")
	}
	out.topic = topic
	out.client = c.client
	go out.trackContext()
	return &out, nil
}

func (c *consumerGroup) trackContext() {
	select {
	case <-c.ctx.Done():
		_ = c.conGroup.Close()
		return
	}
}

func (c *consumerGroup) Close() {
	c.cancel()
}

func (c *consumerGroup) Process(processorFunc ProcessorFunc, pp PostProcessorFuncCG) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			h := handler{
				processor:     processorFunc,
				postProcessor: pp,
				ctx:           c.ctx,
			}
			return c.conGroup.Consume(c.ctx, []string{c.topic}, &h)
		}
	}
}

type handler struct {
	processor     ProcessorFunc
	postProcessor PostProcessorFuncCG
	ctx           context.Context
}

func (c handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (c handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := c.processor(c.ctx, msg.Value)
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
	return nil
}
