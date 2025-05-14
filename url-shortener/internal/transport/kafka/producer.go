package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"urlshortener/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const flushTimeout = 7000

var errUnknownType = errors.New("unknown event type")

type Producer struct {
	cfg      *config.Config
	producer *kafka.Producer
}

func NewProducer(cfg *config.Config) (*Producer, error) {
	const op = "kafka.NewProducer"

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.MsgBroker.Addr, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Producer{
		cfg:      cfg,
		producer: p,
	}, nil
}

func (p *Producer) Details() string {
	return p.producer.String()
}

func (p *Producer) Produce(payload any, topic, key string) error {
	const op = "kafka.Produce"

	msg, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: msg,
		Key:   []byte(key),
	}

	kafkaChan := make(chan kafka.Event)
	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// go func() {
	// 	for e := range p.producer.Events() {
	// 		switch ev := e.(type) {
	// 		case *kafka.Message:
	// 			if ev.TopicPartition.Error != nil {
	// 				slog.Error("Delivery failed", slog.Any("error", ev.TopicPartition))
	// 			} else {
	// 				slog.Info("Delivered message", slog.String("to", ev.TopicPartition.String()))
	// 			}
	// 		}
	// 	}
	// }()
	// return nil

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case *kafka.Error:
		return ev
	default:
		return errUnknownType
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}
