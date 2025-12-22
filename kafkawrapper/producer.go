package kafka

import (
	"context"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Producer interface {
	SendMessage(msg []byte, key string) error
}

type producer struct {
	pro           sarama.SyncProducer
	topic         string
	numPartitions int32
	logger        CustomLogger
}

func (c *client) NewSyncProducer(ctx context.Context, topic string, numPartitions int32, createTopic bool) (Producer, error) {
	pro, err := sarama.NewSyncProducer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating producer")
	}
	out := producer{
		pro:           pro,
		topic:         topic,
		numPartitions: numPartitions,
		logger:        c.logger,
	}

	if createTopic {
		c.createTopic(ctx, c.brokers, topic, numPartitions)
	}
	return &out, nil
}

func (p *producer) SendMessage(msg []byte, key string) error {
	pNum, err := getPartitionNumberWithKey(p.topic, key, p.numPartitions)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	_, _, err = p.pro.SendMessage(&sarama.ProducerMessage{
		Topic:     p.topic,
		Value:     sarama.ByteEncoder(msg),
		Partition: pNum,
		Key:       sarama.StringEncoder(key),
	})
	return err
}
