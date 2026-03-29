package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	handlerFunc := func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
	return handlerFunc
}

func handlerMove(
	gs *gamelogic.GameState,
	ch *amqp.Channel,
) func(move gamelogic.ArmyMove) pubsub.AckType {
	handlerFunc := func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+move.Player.Username,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("Couldn't publish message: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
	return handlerFunc
}

func handlerWar(
	gs *gamelogic.GameState,
	ch *amqp.Channel,
) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	handlerFunc := func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			if err := publishGameLog(
				ch,
				gs.GetUsername(),
				fmt.Sprintf("%s won a war against %s\n", winner, loser),
			); err != nil {
				fmt.Printf("Error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			if err := publishGameLog(
				ch,
				gs.GetUsername(),
				fmt.Sprintf("%s won a war against %s\n", winner, loser),
			); err != nil {
				fmt.Printf("Error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			if err := publishGameLog(
				ch,
				gs.GetUsername(),
				fmt.Sprintf("A war between %s and %s resulted in a draw\n", winner, loser),
			); err != nil {
				fmt.Printf("Error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("error: invalid war outcome")
			return pubsub.NackDiscard
		}
	}
	return handlerFunc
}
