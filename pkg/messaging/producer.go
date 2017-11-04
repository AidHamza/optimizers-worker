package messaging

import (
	"github.com/nsqio/go-nsq"
)

type Producer struct {
	nsq *nsq.Producer
}

func NewProducer(host string, port string) (*Producer, error) {
	nqsConfig := nsq.NewConfig()
	producer, err := nsq.NewProducer(host + ":" + port, nqsConfig)
	if err != nil {
		return &Producer{}, err
	}

	return &Producer {
		nsq: producer,
	}, nil
}

func (producer *Producer) PublishMessage(topic string, message []byte) error {
	err := producer.nsq.Publish(topic, message)
	if err != nil {
		return err
	}

	return nil
}