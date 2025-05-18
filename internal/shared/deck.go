package shared

import (
	"log"
	"math/rand/v2"
)

// Deck represents a collection of cards.
type Deck struct {
	Cards []Card
}

// NewDeck creates a standard 40-card Tressette deck.
func NewDeck() *Deck {
	suits := []Suit{Denari, Spade, Bastoni, Kope}
	// Ranks in typical display order, but Order field determines strength
	ranks := []string{"13", "12", "11", "1", "7", "6", "5", "4", "3", "2"}

	var cards []Card
	for _, suit := range suits {
		for _, rank := range ranks {
			order, okOrder := cardOrder[rank]
			value, okValue := cardValues[rank] // Use the scaled values
			if !okOrder || !okValue {
				// This should not happen with the current maps
				log.Printf("Error: Invalid rank '%s' encountered during deck creation.", rank)
				continue
			}
			cards = append(cards, Card{
				Suit:  suit,
				Rank:  rank,
				Value: value,
				Order: order,
			})
		}
	}

	return &Deck{Cards: cards}
}

// Shuffle randomizes the order of cards in the deck.
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
	log.Println("Deck shuffled.")
}

// Deal distributes cards to players. Returns nil if not enough cards.
func (d *Deck) Deal(numPlayers, cardsPerPlayer int) [][]Card {
	totalCardsNeeded := numPlayers * cardsPerPlayer
	if len(d.Cards) < totalCardsNeeded {
		log.Printf("Error: Not enough cards in deck (%d) to deal %d cards to %d players.", len(d.Cards), cardsPerPlayer, numPlayers)
		return nil
	}

	dealt := make([][]Card, numPlayers)
	start := 0
	for i := 0; i < numPlayers; i++ {
		end := start + cardsPerPlayer
		// Create a copy for the hand to avoid slice pointing issues if deck is modified later
		hand := make([]Card, cardsPerPlayer)
		copy(hand, d.Cards[start:end])
		dealt[i] = hand
		start = end
	}

	d.Cards = []Card{}
	log.Printf("Dealt %d cards to %d players.", cardsPerPlayer, numPlayers)
	return dealt
}

