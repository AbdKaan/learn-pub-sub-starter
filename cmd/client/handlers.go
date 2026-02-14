package main

import (
	"fmt"

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

func handlerMove(gs *gamelogic.GameState) func(move gamelogic.ArmyMove) pubsub.AckType {
	handlerFunc := func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
	return handlerFunc
}
