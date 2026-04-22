package kafka

import (
	"fmt"
	"hash/fnv"

	"github.com/IBM/sarama"

	"github.com/tarmalonchik/golibs/trace"
)

func GenerateKey(key string, chunksNumber int32) (int32, error) {
	if chunksNumber <= 0 {
		return 0, fmt.Errorf("invalid chunks number %d", chunksNumber)
	}

	h := fnv.New32a()
	if _, err := h.Write([]byte(key)); err != nil {
		return 0, fmt.Errorf("error writing to hash: %v", err)
	}

	return int32(h.Sum32() % uint32(chunksNumber)), nil
}

func GetPartition(topic string, key string, chunksNumber int32, numPartitions int32) (int32, error) {
	some := sarama.NewHashPartitioner(topic)

	newKey, err := GenerateKey(key, chunksNumber)
	if err != nil {
		return 0, err
	}

	p, err := some.Partition(
		&sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(newKey)},
		numPartitions,
	)
	if err != nil {
		return 0, trace.FuncNameWithErrorMsg(err, "getting partition")
	}

	return p, nil
}
