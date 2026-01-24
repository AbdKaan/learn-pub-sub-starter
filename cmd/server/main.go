package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	const connectionString string = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf(
			"Error occured while dialing with connection: %v with error: %v",
			connectionString,
			err,
		)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	// Create new channel
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf(
			"Error occured while creating a channel: %v",
			err,
		)
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state...")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf(
					"error occured while publishing the message: %v",
					err,
				)
			}
		case "resume":
			fmt.Println("Publishing resumes game state...")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf(
					"error occured while publishing the message: %v",
					err,
				)
			}
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("This command doesn't exist.")
		}
	}
}
