package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

type AckType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	// Consumer name is empty string so it will be auto generated
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			ack := handler(target)
			switch ack {
			case Ack:
				fmt.Println("Acking...")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("Nacking and requeueing...")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("Nacking and discarding...")
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := false
	autoDelete := false
	exclusive := false

	switch queueType {
	case SimpleQueueDurable:
		durable = true
	case SimpleQueueTransient:
		autoDelete = true
		exclusive = true
	default:
		log.Fatal("Queue needs to be durable or transient")
	}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}
