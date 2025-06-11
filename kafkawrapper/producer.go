package kafka

import (
	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

type Producer interface {
	SendMessage(msg []byte, key string, numPartitions int32) error
}

type producer struct {
	pro   sarama.SyncProducer
	topic string
}

func (c *client) NewSyncProducer(topic string) (Producer, error) {
	pro, err := sarama.NewSyncProducer(c.brokers, c.client.Config())
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "creating producer")
	}
	out := producer{
		pro:   pro,
		topic: topic,
	}
	return &out, nil
}

func (p *producer) SendMessage(msg []byte, key string, numPartitions int32) error {
	pNum, err := getPartitionNumberWithKey(p.topic, key, numPartitions)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "getting part number")
	}

	_, _, err = p.pro.SendMessage(&sarama.ProducerMessage{
		Topic:     p.topic,
		Value:     sarama.ByteEncoder(msg),
		Partition: pNum,
	})
	return err
}
