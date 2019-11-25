package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/dapr/components-contrib/pubsub"
)

type pubSubConsumer struct {
	topic   string
	handler func(msg *pubsub.NewMessage) error
	client  sarama.ConsumerGroup
	ready   chan bool
}

func newPubSubConsumer(topic string, handler func(msg *pubsub.NewMessage) error, client sarama.ConsumerGroup) *pubSubConsumer {
	return &pubSubConsumer{topic: topic, handler: handler, client: client, ready: make(chan bool)}
}

func (consumer *pubSubConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if consumer.handler != nil {
			err := consumer.handler(&pubsub.NewMessage{
				Data:  message.Value,
				Topic: message.Topic,
			})
			if err != nil {
				session.MarkMessage(message, "")
			}
		}
	}
	return nil
}

func (consumer *pubSubConsumer) Setup(sarama.ConsumerGroupSession) error {
	// signal that the consumer is ready
	close(consumer.ready)
	return nil
}

func (consumer *pubSubConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *pubSubConsumer) consume(ctx context.Context) {
	defer consumer.client.Close()

	for {
		if err := consumer.client.Consume(ctx, []string{consumer.topic}, consumer); err != nil {
			log.Errorf("%s failed to consume: %s", logMessagePrefix, err)
		}

		if ctx.Err() != nil {
			return
		}

		consumer.ready = make(chan bool)
	}
}

func (consumer *pubSubConsumer) waitReady(ctx context.Context) {
	select {
	case <-consumer.ready:
		return
	case <-ctx.Done():
		return
	}
}
