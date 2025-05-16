package shared

import (
	"github.com/google/uuid"
)

// TeamEnum represents the two teams in the game.
type TeamEnum int

const (
	TeamRed   TeamEnum  = 1           // Team 1 (players 0 and 2)
	TeamBlue  TeamEnum  = 2           // Team 2 (players 1 and 3)
)

// Team represents a team in the Tressette game.
type Team struct {
	ID         string    	`json:"id"`
	Players    [2]*Player 	`json:"players"`
	Score      int       	`json:"score"`
	TotalScore int       	`json:"total_score"`   
	TeamNumber int       	`json:"-"`
}

// NewTeam creates a new team with the given logical number and players.
// It generates a unique UUID for the team ID.
func NewTeam(teamNumber int, player1, player2 *Player) *Team {
	return &Team{
		ID:         uuid.NewString(), // Generate UUID
		Players:    [2]*Player{player1, player2},
		Score:      0,
		TotalScore: 0,
		TeamNumber: teamNumber, // Store the logical team number
	}
}

// AddScore adds points to the team's total score.
func (t *Team) AddScore(points int) {
	t.Score += points
}

// ResetScore resets the score to 0.
func (t *Team) ResetScore() {
	t.Score = 0
}

// Transfer score to total score.
func (t *Team) TransferScore() {
	// this is needed because t.Score is scaled by 3
	s := (t.Score - t.Score % 3) / 3
	t.TotalScore += s
	t.ResetScore()
}