package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/tarmalonchik/golibs/trace"
)

type Consumer interface {
	Close() error
	Process(ctx context.Context, processorFunc ProcessorFunc, postProcessor PostProcessorFunc) error
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
	readOnlyOneMsg bool
	logger         CustomLogger
	once           func()

	cancel context.CancelFunc
}

func (c *client) NewConsumer(topic string, key string, numPartitions int32, createTopic bool) (Consumer, error) {
	var out consumer
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

	part, err := getPartitionNumberWithKey(topic, key, numPartitions)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	out = consumer{
		topic:     topic,
		partition: part,
		logger:    c.logger,
		client:    c.client,
	}

	out.con, err = sarama.NewConsumer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "create consumer")
	}

	out.once = sync.OnceFunc(func() {
		_ = out.con.Close()
	})
	return &out, nil
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

func (c *consumer) Close() error {
	if c.cancel == nil {
		return fmt.Errorf("consumer group has never started")
	}
	c.cancel()
	c.once()
	return nil
}

func (c *consumer) Process(ctx context.Context, processorFunc ProcessorFunc, postProcessor PostProcessorFunc) error {
	ctx, c.cancel = context.WithCancel(ctx)

	partConsumer, err := c.con.ConsumePartition(c.topic, c.partition, c.offset)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "processing consumer")
	}

	for {
		select {
		case <-ctx.Done():
			c.once()
			if c.logger != nil {
				c.logger.Info(fmt.Sprintf("closing consumer: %s", c.topic))
			}
			return nil
		case msg := <-partConsumer.Messages():
			err = processorFunc(ctx, msg.Value, string(msg.Key))
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
