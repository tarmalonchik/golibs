package kafka

import (
	"context"
	"errors"
	"fmt"
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
	SetLastExistingMessageOffset() error
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
	logger         CustomLogger
}

func (c *client) NewConsumer(ctx context.Context, topic string, key string, numPartitions int32, createTopic bool) (Consumer, error) {
	var out consumer
	var err error

	if topic == "" {
		return nil, errors.New("empty topic")
	}

	if createTopic {
		c.createTopic(ctx, c.brokers, topic, numPartitions)
	}

	part, err := getPartitionNumberWithKey(topic, key, numPartitions)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	out.ctx, out.cancel = context.WithCancel(ctx)
	out.con, err = sarama.NewConsumer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "create consumer")
	}
	out.topic = topic
	out.partition = part
	out.logger = c.logger
	out.client = c.client
	go out.trackContext()
	return &out, nil
}

func (c *consumer) trackContext() {
	<-c.ctx.Done()
	_ = c.con.Close()
}

func (c *consumer) SetOffset(offset int64) {
	c.offset = offset
}

func (c *consumer) ReadOnlyOne() {
	c.readOnlyOneMsg = true
}

func (c *consumer) SetLastExistingMessageOffset() error {
	if err := c.client.RefreshMetadata(c.topic); err != nil {
		return err
	}

	offset, err := c.client.GetOffset(c.topic, c.partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("getting last existing offset topic: %s, partition: %d, offset: %d %w", c.topic, c.partition, offset, err)
	}

	if offset > 0 {
		c.offset = offset - 1
	}

	return nil
}

func (c *consumer) SetTimeOffset(time time.Time) error {
	offset, err := c.client.GetOffset(c.topic, c.partition, time.UTC().UnixMilli())
	if err != nil {
		return fmt.Errorf("getting offset by time topic: %s, partition: %d, offset: %d %w", c.topic, c.partition, offset, err)
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
			if c.logger != nil {
				c.logger.Infof("closing consumer: %s", c.topic)
			}
			return nil
		case msg := <-partConsumer.Messages():
			err = processorFunc(c.ctx, msg.Value, string(msg.Key))
			if errors.Is(err, ErrInvalidKey) {
				continue
			}

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
