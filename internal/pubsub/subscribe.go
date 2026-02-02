package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	// Consumer name is empty string so it will be auto generated
	deliveryChan, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() error {
		for delivery := range deliveryChan {
			var msg T
			err = json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				return err
			}
			handler(msg)
			delivery.Ack(false)
		}
		return nil
	}()
	return nil
}
