package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/dapr/components-contrib/pubsub"
)

const (
	errorLogPrefix   = "kafka pub/sub error:"
	logMessagePrefix = "kafka pub/sub"
	brokersKey       = "brokers"
	consumerIDKey    = "consumerID"
	saslUsernameKey  = "saslUsername"
	saslPasswordKey  = "saslPassword"
)

type kafka struct {
	producer             sarama.SyncProducer
	authRequired         bool
	saslUsername         string
	saslPassword         string
	consumerID           string
	brokers              []string
	consumerGroupFactory func(addrs []string, groupID string, config *sarama.Config) (sarama.ConsumerGroup, error)
	producerFactory      func(addrs []string, config *sarama.Config) (sarama.SyncProducer, error)
}

// NewKafka returns a new Kafka pub-sub implementation
func NewKafka() pubsub.PubSub {
	return newKafka(sarama.NewConsumerGroup, sarama.NewSyncProducer)
}

func newKafka(consumerGroupFactory func(addrs []string, groupID string, config *sarama.Config) (sarama.ConsumerGroup, error),
	producerFactory func(addrs []string, config *sarama.Config) (sarama.SyncProducer, error)) pubsub.PubSub {
	return &kafka{consumerGroupFactory: consumerGroupFactory, producerFactory: producerFactory}
}

func (k *kafka) Init(metadata pubsub.Metadata) error {
	if val, ok := metadata.Properties[brokersKey]; ok && val != "" {
		k.brokers = strings.Split(val, ",")
		if len(k.brokers) == 0 {
			return fmt.Errorf("%s broker is empty", errorLogPrefix)
		}
	} else {
		return fmt.Errorf("%s broker missing", errorLogPrefix)
	}

	if val, ok := metadata.Properties[consumerIDKey]; ok && val != "" {
		k.consumerID = val
	} else {
		return fmt.Errorf("%s consumerID is missing", errorLogPrefix)
	}

	if val, ok := metadata.Properties[saslUsernameKey]; ok && val != "" {
		k.saslUsername = val
	}

	if val, ok := metadata.Properties[saslPasswordKey]; ok && val != "" {
		k.saslPassword = val
	}

	if len(k.saslUsername) > 0 && len(k.saslPassword) > 0 {
		k.authRequired = true
	}

	producer, err := k.createProducer()
	if err != nil {
		return err
	}

	k.producer = producer

	return nil
}

func (k *kafka) updateConfigForAuthentication(config *sarama.Config) {
	config.Net.SASL.Enable = true
	config.Net.SASL.User = k.saslUsername
	config.Net.SASL.Password = k.saslUsername
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext

	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		//InsecureSkipVerify: true,
		ClientAuth: 0,
	}
}

func (k *kafka) createProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Version = sarama.V1_0_0_0

	if k.authRequired {
		k.updateConfigForAuthentication(config)
	}

	producer, err := k.producerFactory(k.brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func (k *kafka) Publish(req *pubsub.PublishRequest) error {
	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: req.Topic,
		Value: sarama.ByteEncoder(req.Data)})

	return err
}

func (k *kafka) Subscribe(req pubsub.SubscribeRequest, handler func(msg *pubsub.NewMessage) error) error {
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0

	if k.authRequired {
		k.updateConfigForAuthentication(config)
	}

	client, err := k.consumerGroupFactory(k.brokers, k.consumerID, config)
	if err != nil {
		return err
	}

	consumer := newPubSubConsumer(req.Topic, handler, client)

	ctx := context.Background()
	go consumer.consume(ctx)

	consumer.waitReady(ctx)

	return nil
}
