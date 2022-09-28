package services

import (
	"github.com/Shopify/sarama"
)

type TransactionConsumer struct{}

func (c *TransactionConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *TransactionConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *TransactionConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {

	}
}
