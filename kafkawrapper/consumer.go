package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"

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

	mu     sync.Mutex
	cancel context.CancelFunc
}

type ConsumerConfig struct {
	Topic         string `env:"TOPIC,required"`
	Key           string `env:"KEY" envDefault:""`
	NumPartitions int    `env:"NUM_PARTITIONS" envDefault:"100"`
	CreateTopic   bool   `env:"CREATE_TOPIC" envDefault:"true"`
}

func (c *client) NewConsumer(config ConsumerConfig) (Consumer, error) {
	if config.Topic == "" {
		return nil, ErrTopicIsEmpty
	}

	config.Topic = c.wrapTopic(config.Topic)

	var out consumer
	var err error

	if config.NumPartitions < 1 {
		return nil, ErrShouldHaveAtLeastOnePartition
	}

	if config.CreateTopic {
		if err := c.createTopic(c.brokers, config.Topic, int32(config.NumPartitions)); err != nil {
			return nil, err
		}
	}

	actualPartitions, err := c.getTopicPartitionCount(config.Topic, int32(config.NumPartitions))
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "getting topic partition count")
	}

	part, err := getPartitionNumberWithKey(config.Topic, config.Key, actualPartitions)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	out = consumer{
		topic:     config.Topic,
		partition: part,
		logger:    c.logger,
		client:    c.client,
	}

	err = retryer(func() error {
		out.con, err = sarama.NewConsumer(c.brokers, c.client.Config())
		if err != nil {
			return trace.FuncNameWithErrorMsg(err, "create consumer")
		}

		return nil
	})
	if err != nil {
		return nil, err
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

func (c *consumer) SetTimeOffset(t time.Time) error {
	if t.After(time.Now().UTC()) {
		return fmt.Errorf("time should not be in the future: %s", t.String())
	}

	offset, err := c.client.GetOffset(c.topic, c.partition, t.UTC().UnixMilli())
	if err != nil {
		return fmt.Errorf("getting offset by time topic: %s, partition: %d, offset: %d %w", c.topic, c.partition, offset, err)
	}
	c.offset = offset

	if c.offset == -1 {
		c.logger.Info(fmt.Sprintf("no offset found for time: %s", t.String()))
		c.offset = sarama.OffsetOldest
	}

	return nil
}

func (c *consumer) Close() error {
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

func (c *consumer) Process(ctx context.Context, processorFunc ProcessorFunc, postProcessor PostProcessorFunc) error {
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()

	partConsumer, err := c.con.ConsumePartition(c.topic, c.partition, c.offset)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, fmt.Sprintf("processing consumer: topic %s, partition: %d", c.topic, c.offset))
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
				postProcessor(ctx, err)
			}

			if err != nil {
				continue
			}

			if c.readOnlyOneMsg {
				cancel()
				return nil
			}
		}
	}
}

func retryer(in func() error) error {
	err := retry.Do(
		in,
		retry.Attempts(5),
		retry.Delay(300*time.Millisecond),
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}
