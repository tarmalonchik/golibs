package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type ConsumerGroup interface {
	Close()
	Process(processorFunc ProcessorFunc, writeErr WriteError) error
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
		return nil, err
	}
	out.topic = topic
	out.client = c.client
	go out.trackContext()
	return &out, nil
}

func (c *consumerGroup) trackContext() {
	select {
	case <-c.ctx.Done():
		fmt.Println("close success")
		_ = c.conGroup.Close()
		return
	}
}

func (c *consumerGroup) Close() {
	c.cancel()
}

func (c *consumerGroup) Process(processorFunc ProcessorFunc, writeErr WriteError) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			h := handler{
				processor: processorFunc,
				writeErr:  writeErr,
			}
			return c.conGroup.Consume(c.ctx, []string{c.topic}, &h)
		}
	}
}

type handler struct {
	processor ProcessorFunc
	writeErr  WriteError
	ctx       context.Context
}

func (c handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (c handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Println("here")
	for msg := range claim.Messages() {
		err := c.processor(c.ctx, msg.Value)
		if err != nil {
			c.writeErr(err)
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
