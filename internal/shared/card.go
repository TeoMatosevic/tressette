package shared

// Suit represents the suit of a card (e.g., Denari, Spade, Bastoni, Kope).
type Suit string

const (
	Denari  Suit = "Denari"
	Spade   Suit = "Spade"
	Bastoni Suit = "Bastoni"
	Kope    Suit = "Kope"
)

// Card represents a single card in the Tressette game.
type Card struct {
	Suit  Suit   // The suit of the card
	Rank  string // The rank of the card 
	Value int    // The value of the card for scoring purposes
	Order int    // The rank order within a suit (higher is better)
}

// Define card order for easier comparison
var cardOrder = map[string]int{
	"3":      10,
	"2":      9,
	"1":	  8,
	"4":      1,
	"5":      2,
	"6":      3,
	"7":      4,
	"11":     5, // Fante
	"12":     6, // Cavallo
	"13":     7, // Re
}

// Define card values for scoring (scaled by 3)
var cardValues = map[string]int{
	"1":      3, // Scaled: 1 * 3
	"2":      1, // Scaled: 1/3 * 3
	"3":      1, // Scaled: 1/3 * 3
	"4":      0,
	"5":      0,
	"6":      0,
	"7":      0,
	"11":	  1, // Scaled: 1/3 * 3
	"12": 	  1, // Scaled: 1/3 * 3
	"13":     1, // Scaled: 1/3 * 3
}