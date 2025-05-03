package shared

// Player represents a player in the Tressette game.
type Player struct {
	ID   string // Unique identifier for the player
	Name string // Player's chosen name
	Hand []Card // Cards currently held by the player
}

// NewPlayer creates a new player with the given ID and name.
func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:   id,
		Name: name, // Initialize name
		Hand: []Card{},
	}
}

// AddCard adds a card to the player's hand.
func (p *Player) AddCard(card Card) {
	p.Hand = append(p.Hand, card)
}

// RemoveCard removes a card from the player's hand.
func (p *Player) RemoveCard(card Card) bool {
	for i, c := range p.Hand {
		if c == card {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) FindCard(suit Suit, rank string) (*Card, bool) {
	for _, card := range p.Hand {
		if card.Suit == suit && card.Rank == rank {
			return &card, true // Card found
		}
	}
	return nil, false // Card not found
}

func (p *Player) HasSuit(suit Suit) bool {
	for _, card := range p.Hand {
		if card.Suit == suit {
			return true // Player has the suit
		}
	}
	return false // Player does not have the suit
}