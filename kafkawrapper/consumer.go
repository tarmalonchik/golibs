package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Consumer interface {
	Close()
	Process(processorFunc ProcessorFunc, postProcessor PostProcessorFunc) error
	SetOffset(offset int64)
	SetTimeOffset(time time.Time) error
	ReadOnlyOne()
}

type consumer struct {
	client         sarama.Client
	con            sarama.Consumer
	topic          string
	partition      int32
	offset         int64
	ctx            context.Context
	cancel         context.CancelFunc
	readOnlyOneMsg bool
}

func (c *client) NewConsumer(ctx context.Context, topic string, partition int32) (Consumer, error) {
	var out consumer
	var err error

	out.ctx, out.cancel = context.WithCancel(ctx)
	out.con, err = sarama.NewConsumer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "create consumer")
	}
	out.topic = topic
	out.partition = partition
	out.client = c.client
	go out.trackContext()
	return &out, nil
}

func (c *consumer) trackContext() {
	select {
	case <-c.ctx.Done():
		_ = c.con.Close()
		return
	}
}

func (c *consumer) SetOffset(offset int64) {
	c.offset = offset
}

func (c *consumer) ReadOnlyOne() {
	c.readOnlyOneMsg = true
}

func (c *consumer) SetTimeOffset(time time.Time) error {
	offset, err := c.client.GetOffset(c.topic, c.partition, time.UTC().UnixMilli())
	if err != nil {
		return err
	}
	c.offset = offset
	return nil
}

func (c *consumer) Close() {
	c.cancel()
}

func (c *consumer) Process(processorFunc ProcessorFunc, postProcessor PostProcessorFunc) error {
	partConsumer, err := c.con.ConsumePartition(c.topic, c.partition, c.offset)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "processing consumer")
	}

	for {
		select {
		case <-c.ctx.Done():
			return nil
		case msg := <-partConsumer.Messages():
			err = processorFunc(c.ctx, msg.Value)
			if postProcessor != nil {
				postProcessor(err)
			}
			if err != nil {
				continue
			}
			if c.readOnlyOneMsg {
				c.cancel()
				return nil
			}
		}
	}
}
