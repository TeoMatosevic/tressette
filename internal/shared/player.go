package shared

import (
	"log"
)

type Declaration struct {
	Type string `json:"type"` // Type of declaration (e.g., "napola", "three_or_four_of_kind")
	Suit Suit   `json:"suit"` // Suit of the card involved in the declaration
	Rank string `json:"rank"` // Rank of the card involved in the declaration
}

type DeclarationResult struct {
	Success     bool `json:"success"`      // Indicates if the declaration was successful
	Points      int  `json:"points"`       // Points awarded for the declaration
	WithoutSuit Suit `json:"without_suit"` // Suit of the card involved in the declaration
}

// Player represents a player in the Tressette game.
type Player struct {
	ID           string        // Unique identifier for the player
	Name         string        // Player's chosen name
	Hand         []Card        // Cards currently held by the player
	DesiredTeam  TeamEnum      // Desired team for the player
	Declarations []Declaration // Declarations made by the player
}

// NewPlayer creates a new player with the given ID and name.
func NewPlayer(id string, name string, desired_team TeamEnum) *Player {
	return &Player{
		ID:           id,
		Name:         name, // Initialize name
		Hand:         []Card{},
		DesiredTeam:  desired_team, // Initialize desired team
		Declarations: []Declaration{},
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

func (p *Player) AddDeclaration(declaration Declaration) DeclarationResult {
	switch declaration.Type {
	case "napola":
		num_of_cards := 0
		for _, c := range p.Hand {
			if c.Suit == declaration.Suit && (c.Rank == "1" || c.Rank == "2" || c.Rank == "3") {
				num_of_cards++
			}
		}
		if num_of_cards != 3 {
			log.Printf("Invalid napola declaration: %d cards of suit %s found, expected 3.", num_of_cards, declaration.Suit)
			return DeclarationResult{Success: false, Points: 0}
		}
		for _, d := range p.Declarations {
			if d.Type == "napola" && d.Suit == declaration.Suit {
				log.Printf("Invalid napola declaration: already declared for suit %s.", declaration.Suit)
				return DeclarationResult{Success: false, Points: 0}
			}
		}
		p.Declarations = append(p.Declarations, declaration)
		return DeclarationResult{Success: true, Points: num_of_cards} // Points for napola
	case "three_or_four_of_kind":
		if declaration.Rank != "1" && declaration.Rank != "2" && declaration.Rank != "3" {
			log.Printf("Invalid three_or_four_of_kind declaration: rank %s is not valid.", declaration.Rank)
			return DeclarationResult{Success: false, Points: 0}
		}
		num_of_cards := 0
		suits := map[Suit]bool{
			Denari:  false,
			Spade:   false,
			Bastoni: false,
			Kope:    false,
		}

		for _, c := range p.Hand {
			if c.Rank == declaration.Rank {
				num_of_cards++
				suits[c.Suit] = true
			}
		}
		if num_of_cards != 3 && num_of_cards != 4 {
			log.Printf("Invalid three_or_four_of_kind declaration: %d cards of rank %s found, expected 3 or 4.", num_of_cards, declaration.Rank)
			return DeclarationResult{Success: false, Points: 0}
		}
		for _, d := range p.Declarations {
			if d.Type == "three_or_four_of_kind" && d.Rank == declaration.Rank {
				log.Printf("Invalid three_or_four_of_kind declaration: already declared for rank %s.", declaration.Rank)
				return DeclarationResult{Success: false, Points: 0}
			}
		}
		var without_suit Suit
		// find mising suit
		if num_of_cards == 3 {
			for suit, found := range suits {
				if !found {
					without_suit = suit
					break
				}
			}
		}

		p.Declarations = append(p.Declarations, declaration)
		return DeclarationResult{Success: true, Points: num_of_cards, WithoutSuit: without_suit} // Points for three_or_four_of_kind
	default:
		log.Panicf("Invalid declaration type: %s", declaration.Type)
		return DeclarationResult{Success: false, Points: 0} // Invalid declaration
	}
}
