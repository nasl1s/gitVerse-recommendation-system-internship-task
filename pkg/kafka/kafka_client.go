package kafka

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	Brokers []string
	writer  *kafka.Writer
}

func NewKafkaClient(brokers []string) *KafkaClient {
	return &KafkaClient{
		Brokers: brokers,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaClient) CreateTopic(topic string, numPartitions, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", k.Brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

func (k *KafkaClient) PublishMessage(topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	return k.writer.WriteMessages(context.Background(), msg)
}

func (k *KafkaClient) SubscribeToTopic(ctx context.Context, topic, groupID string) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  k.Brokers,
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 10e6,
	})
	defer r.Close()

	for {
		_, err := r.ReadMessage(ctx)
		if errors.Is(err, context.Canceled) {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (k *KafkaClient) DeleteTopic(topic string) error {
	conn, err := kafka.Dial("tcp", k.Brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	return controllerConn.DeleteTopics(topic)
}

func (k *KafkaClient) Close() error {
	if k.writer != nil {
		return k.writer.Close()
	}
	return nil
}

func (k *KafkaClient) SubscribeToTopicsFallback(ctx context.Context, topics []string, groupID string, handler func(message kafka.Message) error) error {
	for _, topic := range topics {
		go func(topic string) {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:  k.Brokers,
				GroupID:  groupID,
				Topic:    topic,
				MinBytes: 10e3,
				MaxBytes: 10e6,
			})
			defer reader.Close()

			for {
				m, err := reader.ReadMessage(ctx)
				if errors.Is(err, context.Canceled) {
					break
				} else if err != nil {
					continue
				}

				if err := handler(m); err != nil {
					continue
				}
			}
		}(topic)
	}

	<-ctx.Done()
	return nil
}
