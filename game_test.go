package santase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	hand := NewHand()
	hand.AddCard(NewCard(Nine, Hearts))
	hand.AddCard(NewCard(Jack, Hearts))
	hand.AddCard(NewCard(Queen, Hearts))
	hand.AddCard(NewCard(King, Hearts))
	hand.AddCard(NewCard(Ten, Hearts))
	hand.AddCard(NewCard(Ace, Hearts))

	trumpCard := NewCard(Ace, Spades)

	// check that it compiles and runs without panics
	CreateGame(hand, trumpCard, false)
}

func TestNewGameIncompleteHand(t *testing.T) {
	hand := NewHand()
	trumpCard := NewCard(Ace, Spades)
	assert.PanicsWithValue(
		t, "player's hand is not complete",
		func() { CreateGame(hand, trumpCard, false) },
	)
}

func createSampleHand() Hand {
	hand := NewHand()
	hand.AddCard(NewCard(Nine, Diamonds))
	hand.AddCard(NewCard(King, Spades))
	hand.AddCard(NewCard(Queen, Diamonds))
	hand.AddCard(NewCard(Nine, Spades))
	hand.AddCard(NewCard(Ace, Spades))
	hand.AddCard(NewCard(Ten, Hearts))
	return hand
}

func createSampleGame() Game {
	hand := createSampleHand()
	trumpCard := NewCard(Ten, Clubs)
	return CreateGame(hand, trumpCard, true)
}

func createSampleGameWithTrumpCard(trumpCard Card) Game {
	hand := createSampleHand()
	return CreateGame(hand, trumpCard, true)
}

func TestUpdateOpponentMove(t *testing.T) {
	game := createSampleGame()

	card := NewCard(Ace, Diamonds)
	opponentMove := Move{Card: card}

	assert.Nil(t, game.cardPlayed)
	assert.True(t, game.unseenCards.HasCard(card))

	game.UpdateOpponentMove(opponentMove)

	assert.Equal(t, card, *game.cardPlayed)
	assert.False(t, game.isOpponentMove)
	assert.False(t, game.unseenCards.HasCard(card))
}

func TestUpdateOpponentMoveInferingOpponentCards(t *testing.T) {
	t.Run("when announcing", func(t *testing.T) {
		game := createSampleGame()

		// simulate if one hand has been played already
		game.seenCards.AddCard(NewCard(Queen, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Queen, Diamonds))
		game.seenCards.AddCard(NewCard(Ace, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Ace, Diamonds))
		game.hand.RemoveCard(NewCard(Queen, Diamonds))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))
		game.unseenCards.RemoveCard(NewCard(Jack, Hearts))

		assert.False(t, game.knownOpponentCards.HasCard(NewCard(King, Hearts)))
		assert.True(t, game.unseenCards.HasCard(NewCard(King, Hearts)))
		opponentMove := Move{
			Card:           NewCard(Queen, Hearts),
			IsAnnouncement: true,
		}
		game.UpdateOpponentMove(opponentMove)
		assert.True(t, game.knownOpponentCards.HasCard(NewCard(King, Hearts)))
		assert.False(t, game.unseenCards.HasCard(NewCard(King, Hearts)))
	})

	t.Run("after switching trump card", func(t *testing.T) {
		game := createSampleGame()

		// simulate if one hand has been played already
		game.seenCards.AddCard(NewCard(Queen, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Queen, Diamonds))
		game.seenCards.AddCard(NewCard(Ace, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Ace, Diamonds))
		game.hand.RemoveCard(NewCard(Queen, Diamonds))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))
		game.unseenCards.RemoveCard(NewCard(Jack, Hearts))

		originalTrumpCard := *game.trumpCard
		assert.False(t, game.knownOpponentCards.HasCard(originalTrumpCard))
		opponentMove := Move{
			Card:            NewCard(Queen, Hearts),
			SwitchTrumpCard: true,
		}
		game.UpdateOpponentMove(opponentMove)
		assert.True(t, game.knownOpponentCards.HasCard(originalTrumpCard))
		assert.False(t, game.unseenCards.HasCard(originalTrumpCard))
	})

	t.Run("after switching trump card and announcing", func(t *testing.T) {
		game := createSampleGame()

		// simulate if one hand has been played already
		game.seenCards.AddCard(NewCard(Queen, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Queen, Diamonds))
		game.seenCards.AddCard(NewCard(Ace, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Ace, Diamonds))
		game.hand.RemoveCard(NewCard(Queen, Diamonds))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))
		game.unseenCards.RemoveCard(NewCard(Jack, Hearts))

		originalTrumpCard := *game.trumpCard
		assert.False(t, game.knownOpponentCards.HasCard(originalTrumpCard))
		assert.False(t, game.knownOpponentCards.HasCard(NewCard(King, Hearts)))
		assert.True(t, game.unseenCards.HasCard(NewCard(King, Hearts)))
		opponentMove := Move{
			Card:            NewCard(Queen, Hearts),
			IsAnnouncement:  true,
			SwitchTrumpCard: true,
		}
		game.UpdateOpponentMove(opponentMove)
		assert.True(t, game.knownOpponentCards.HasCard(originalTrumpCard))
		assert.True(t, game.knownOpponentCards.HasCard(NewCard(King, Hearts)))
		assert.False(t, game.unseenCards.HasCard(NewCard(King, Hearts)))
	})

	t.Run("after drawing all cards", func(t *testing.T) {
		// TODO
	})
}

func TestUpdateOpponentMoveInvalidSituations(t *testing.T) {
	card := NewCard(Ace, Diamonds)
	opponentMove := Move{Card: card}

	t.Run("not opponents move", func(t *testing.T) {
		game := createSampleGame()
		game.isOpponentMove = false

		assert.PanicsWithValue(
			t, "not opponent's turn",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("card has been played already", func(t *testing.T) {
		game := createSampleGame()

		// simulate if one hand has been played already
		game.seenCards.AddCard(NewCard(Queen, Diamonds))
		game.unseenCards.RemoveCard(NewCard(Queen, Diamonds))
		game.seenCards.AddCard(card)
		game.unseenCards.RemoveCard(card)
		game.hand.RemoveCard(NewCard(Queen, Diamonds))
		game.opponentScore = 14

		assert.PanicsWithValue(
			t, "card has already been played",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("card is in our hand", func(t *testing.T) {
		game := createSampleGame()
		opponentMove := Move{Card: NewCard(Nine, Diamonds)}

		assert.PanicsWithValue(
			t, "card is in ai's hand",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("card is the one on the table", func(t *testing.T) {
		game := createSampleGame()

		// simulate if ai has played first move
		card := NewCard(Ace, Spades)
		game.cardPlayed = &card
		game.hand.RemoveCard(card)
		opponentMove := Move{Card: card}

		assert.PanicsWithValue(
			t, "card is the same as the one on the table",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("playing before drawing", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14

		assert.PanicsWithValue(
			t, "should not play before drawing cards",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("playing trump card", func(t *testing.T) {
		game := createSampleGame()
		opponentMove := Move{Card: *game.trumpCard}

		assert.PanicsWithValue(
			t, "played card is the trump card",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("switching trump card when responding", func(t *testing.T) {
		game := createSampleGame()

		// simulate if ai has played first move
		card := NewCard(Ace, Spades)
		game.cardPlayed = &card
		game.hand.RemoveCard(card)
		opponentMove := Move{
			Card:            opponentMove.Card,
			SwitchTrumpCard: true,
		}

		assert.PanicsWithValue(
			t, "cannot switch trump card when you're not first to play",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("switching trump card on first move", func(t *testing.T) {
		game := createSampleGame()
		opponentMove := Move{
			Card:            opponentMove.Card,
			SwitchTrumpCard: true,
		}

		assert.PanicsWithValue(
			t, "cannot switch trump card on first move",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("switching trump card with only two cards left", func(t *testing.T) {
		// TODO
	})

	t.Run("switching trump card after all cards are taken", func(t *testing.T) {
		// TODO
	})

	t.Run("switching trump card with rank nine", func(t *testing.T) {
		game := createSampleGameWithTrumpCard(NewCard(Nine, Clubs))
		opponentMove := Move{
			Card:            opponentMove.Card,
			SwitchTrumpCard: true,
		}

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))

		assert.PanicsWithValue(
			t, "cannot switch trump card - trump card is a nine",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("announcing in the middle of a play", func(t *testing.T) {
		game := createSampleGame()

		// simulate if ai has played first move
		card := NewCard(Ace, Spades)
		game.cardPlayed = &card
		game.hand.RemoveCard(card)

		opponentMove := Move{
			Card:           NewCard(Queen, Clubs),
			IsAnnouncement: true,
		}

		assert.PanicsWithValue(
			t, "cannot announce when you're not first to play",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("announcing when the other card has already been played", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))

		opponentMove := Move{
			Card:           NewCard(Queen, Spades),
			IsAnnouncement: true,
		}
		assert.PanicsWithValue(
			t, "cannot be an announcement because other card has already been played",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("announcing on first move", func(t *testing.T) {
		game := createSampleGame()

		opponentMove := Move{
			Card:           NewCard(Queen, Hearts),
			IsAnnouncement: true,
		}
		assert.PanicsWithValue(
			t, "cannot announce on first move",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("announcing when the other card is in ai's hand", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))

		opponentMove := Move{
			Card:           NewCard(King, Diamonds),
			IsAnnouncement: true,
		}
		assert.PanicsWithValue(
			t, "cannot be an announcement because other card is in ai's hand",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("announcing when the other card is the trump card", func(t *testing.T) {
		game := createSampleGameWithTrumpCard(NewCard(Queen, Clubs))

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14
		game.hand.AddCard(NewCard(Jack, Hearts))

		opponentMove := Move{
			Card:           NewCard(King, Clubs),
			IsAnnouncement: true,
		}
		assert.PanicsWithValue(
			t, "cannot be an announcement because other card is the trump card",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("closing game on first move", func(t *testing.T) {
		game := createSampleGame()
		opponentMove := Move{
			Card:      card,
			CloseGame: true,
		}

		assert.PanicsWithValue(
			t, "cannot close game on first move",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("closing game when second to play", func(t *testing.T) {
		game := createSampleGame()

		// simulate if ai has played first move
		c := NewCard(Ace, Spades)
		game.cardPlayed = &c
		game.hand.RemoveCard(c)

		opponentMove := Move{
			Card:      card,
			CloseGame: true,
		}

		assert.PanicsWithValue(
			t, "cannot close game when second to move",
			func() { game.UpdateOpponentMove(opponentMove) },
		)
	})

	t.Run("closing game with two cards left in stack", func(t *testing.T) {
		// TODO
	})

	t.Run("closing game after all cards have been drawn", func(t *testing.T) {
		// TODO
	})

	t.Run("closing game after the game has been closed", func(t *testing.T) {
		// TODO
	})

	t.Run("switching trump card after game has been closed", func(t *testing.T) {
		// TODO
	})

	t.Run("drawing cards when the game is closed", func(t *testing.T) {
		// TODO
	})
}

func TestUpdateOpponentMoveEdgeCaseSituations(t *testing.T) {
	// TODO:
	// playing trump card after switching it
	// playing trump card after drawing it
	// playing trump card after switching it + announcing
	// playing trump card after drawing it
}

func TestUpdateDrawnCardInvalidSituations(t *testing.T) {
	t.Run("in the middle of a play", func(t *testing.T) {
		game := createSampleGame()

		// simulate if ai has played first move
		card := NewCard(Ace, Spades)
		game.cardPlayed = &card
		game.hand.RemoveCard(card)

		assert.PanicsWithValue(
			t, "cannot draw cards in the middle of a play",
			func() { game.UpdateDrawnCard(NewCard(Jack, Hearts)) },
		)
	})

	t.Run("before first move", func(t *testing.T) {
		game := createSampleGame()

		assert.PanicsWithValue(
			t, "should not draw cards before the first play",
			func() { game.UpdateDrawnCard(NewCard(Jack, Hearts)) },
		)
	})

	t.Run("drawing twice in a row", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14
		game.UpdateDrawnCard(NewCard(Jack, Hearts))

		assert.PanicsWithValue(
			t, "should not draw cards twice before playing",
			func() { game.UpdateDrawnCard(NewCard(Nine, Hearts)) },
		)
	})

	t.Run("drawing seen card", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14

		assert.PanicsWithValue(
			t, "drawn card has been played before",
			func() { game.UpdateDrawnCard(NewCard(Ten, Spades)) },
		)
	})

	t.Run("drawing card in ai's hand", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14

		assert.PanicsWithValue(
			t, "cannot draw card that is in the hand already",
			func() { game.UpdateDrawnCard(NewCard(Ace, Spades)) },
		)
	})

	t.Run("drawing card in opponent's hand", func(t *testing.T) {
		// TODO
	})

	t.Run("drawing card after every card is drawn", func(t *testing.T) {
		// TODO
	})

	t.Run("drawing trump card", func(t *testing.T) {
		game := createSampleGame()

		// simulating playing one hand
		game.seenCards.AddCard(NewCard(King, Spades))
		game.unseenCards.RemoveCard(NewCard(King, Spades))
		game.seenCards.AddCard(NewCard(Ten, Spades))
		game.unseenCards.RemoveCard(NewCard(Ten, Spades))
		game.hand.RemoveCard(NewCard(King, Spades))
		game.opponentScore = 14

		assert.PanicsWithValue(
			t, "cannot draw trump card yet",
			func() { game.UpdateDrawnCard(*game.trumpCard) },
		)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("StrongerCard", func(t *testing.T) {
		a := NewCard(Ace, Spades)
		b := NewCard(Ten, Spades)
		c := NewCard(King, Hearts)
		d := NewCard(King, Diamonds)

		assert.Equal(t, &a, StrongerCard(&a, &b, Spades))
		assert.Equal(t, &a, StrongerCard(&b, &a, Spades))

		assert.Equal(t, &a, StrongerCard(&a, &b, Hearts))
		assert.Equal(t, &a, StrongerCard(&b, &a, Hearts))

		assert.Equal(t, &c, StrongerCard(&a, &c, Hearts))
		assert.Equal(t, &c, StrongerCard(&c, &a, Hearts))

		assert.Equal(t, &a, StrongerCard(&a, &d, Hearts))
		assert.Equal(t, &d, StrongerCard(&d, &a, Hearts))
	})

	t.Run("Points", func(t *testing.T) {
		nine := NewCard(Nine, Hearts)
		jack := NewCard(Jack, Hearts)
		queen := NewCard(Queen, Hearts)
		king := NewCard(King, Hearts)
		ten := NewCard(Ten, Hearts)
		ace := NewCard(Ace, Hearts)

		assert.Equal(t, 0, Points(&nine))
		assert.Equal(t, 2, Points(&jack))
		assert.Equal(t, 3, Points(&queen))
		assert.Equal(t, 4, Points(&king))
		assert.Equal(t, 10, Points(&ten))
		assert.Equal(t, 11, Points(&ace))
	})

	t.Run("getHiddenCards", func(t *testing.T) {
		hand := NewHand()
		hand.AddCard(NewCard(Nine, Diamonds))
		hand.AddCard(NewCard(King, Spades))
		hand.AddCard(NewCard(Queen, Diamonds))
		hand.AddCard(NewCard(Nine, Spades))
		hand.AddCard(NewCard(Ace, Spades))
		hand.AddCard(NewCard(Ten, Hearts))

		trumpCard := NewCard(Nine, Clubs)

		hidden := getHiddenCards(hand, trumpCard)

		assert.Equal(t, 17, len(hidden))
		assert.True(t, hidden.HasCard(NewCard(Jack, Spades)))
		assert.True(t, hidden.HasCard(NewCard(Queen, Clubs)))
		assert.True(t, hidden.HasCard(NewCard(King, Diamonds)))
		assert.True(t, hidden.HasCard(NewCard(Ten, Diamonds)))
		assert.True(t, hidden.HasCard(NewCard(Jack, Clubs)))
		assert.True(t, hidden.HasCard(NewCard(Ace, Hearts)))
		assert.True(t, hidden.HasCard(NewCard(Nine, Hearts)))
		assert.True(t, hidden.HasCard(NewCard(Ten, Clubs)))
		assert.True(t, hidden.HasCard(NewCard(Jack, Diamonds)))
		assert.True(t, hidden.HasCard(NewCard(King, Clubs)))
		assert.True(t, hidden.HasCard(NewCard(Queen, Spades)))
		assert.True(t, hidden.HasCard(NewCard(Jack, Hearts)))
		assert.True(t, hidden.HasCard(NewCard(Ace, Clubs)))
		assert.True(t, hidden.HasCard(NewCard(Ace, Diamonds)))
		assert.True(t, hidden.HasCard(NewCard(Queen, Hearts)))
		assert.True(t, hidden.HasCard(NewCard(Ten, Spades)))
		assert.True(t, hidden.HasCard(NewCard(King, Hearts)))
	})
}
