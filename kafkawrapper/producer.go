package kafka

import (
	"context"
	"errors"
	"sync"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Producer interface {
	SendMessage(msg []byte, key string) error
	Close()
}

type producer struct {
	pro           sarama.SyncProducer
	topic         string
	numPartitions int32
	logger        CustomLogger

	ctx    context.Context
	cancel context.CancelFunc
	once   func()
}

func (c *client) NewSyncProducer(ctx context.Context, topic string, numPartitions int32, createTopic bool) (Producer, error) {
	if topic == "" {
		return nil, errors.New("empty topic")
	}

	pro, err := sarama.NewSyncProducer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating producer")
	}

	out := producer{
		pro:           pro,
		topic:         topic,
		numPartitions: numPartitions,
		logger:        c.logger,
		once: sync.OnceFunc(func() {
			_ = pro.Close()
		}),
	}

	out.ctx, out.cancel = context.WithCancel(ctx)

	if createTopic {
		if err := c.createTopic(c.brokers, topic, numPartitions); err != nil {
			return nil, err
		}
	}
	return &out, nil
}

func (p *producer) SendMessage(msg []byte, key string) error {
	pNum, err := getPartitionNumberWithKey(p.topic, key, p.numPartitions)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	select {
	case <-p.ctx.Done():
		p.once()
		return context.Canceled
	default:
		_, _, err = p.pro.SendMessage(&sarama.ProducerMessage{
			Topic:     p.topic,
			Value:     sarama.ByteEncoder(msg),
			Partition: pNum,
			Key:       sarama.StringEncoder(key),
		})
	}
	return err
}

func (p *producer) Close() {
	p.cancel()
	p.once()
}
