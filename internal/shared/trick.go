package shared

import "log"

// PlayedCard stores a card along with the index of the player who played it.
type PlayedCard struct {
	Card        Card		`json:"card"`
	PlayerIndex int			`json:"player_index"` // Index of the player who played the card
}

// Trick represents a single trick in the Tressette game.
type Trick struct {
	Cards       []PlayedCard // Cards played in the current trick, with player index
	WinnerIndex int          // Index of the player who won the trick (-1 if not determined)
}

// NewTrick creates a new trick instance.
func NewTrick() *Trick {
	return &Trick{
		Cards:       []PlayedCard{},
		WinnerIndex: -1,
	}
}

// AddCard adds a card and the player's index to the trick.
func (t *Trick) AddCard(card Card, playerIndex int) {
	t.Cards = append(t.Cards, PlayedCard{Card: card, PlayerIndex: playerIndex})
}

// DetermineWinner determines the winner of the trick based on Tressette rules.
// Requires the suit that was led for the trick.
func (t *Trick) DetermineWinner(ledSuit Suit) PlayedCard {
	if len(t.Cards) == 0 {
		log.Panicf("Error: Cannot determine winner of an empty trick.")
	}

	highestOrderInSuit := -1
	var card PlayedCard
	card.PlayerIndex = -1 // Initialize to an invalid player index

	// Find the highest card of the led suit
	for _, playedCard := range t.Cards {
		if playedCard.Card.Suit == ledSuit {
			if playedCard.Card.Order > highestOrderInSuit {
				highestOrderInSuit = playedCard.Card.Order
				card = playedCard // Store the winning card
			}
		}
	}

	// If no card of the led suit was played (shouldn't happen if rules are followed,
	// but handle defensively), the first player who played (the leader of the trick) wins.
	if card.PlayerIndex == -1 {
		log.Panicf("Warning: No card of led suit (%s) found in trick. Assigning win to leader (Player %d).", ledSuit, t.Cards[0].PlayerIndex)
	}

	t.WinnerIndex = card.PlayerIndex
	return card
}