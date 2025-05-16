package protocol

import (
	"encoding/json"
	"tressette-game/internal/shared"
)

// Message represents a generic WebSocket message structure.
type Message struct {
	Type    string          `json:"type"`              // Type of the message (e.g., "join_game", "play_card")
	Payload json.RawMessage `json:"payload,omitempty"` // Raw JSON payload, allows flexible structures
}

// --- Client -> Server Payload Structs ---

type CreateGamePayload struct {
	Name 		string 				`json:"name"`
	DesiredTeam shared.TeamEnum 	`json:"desired_team"` // Added desired team
	PointsGoal 	int 				`json:"points_goal"`   // Added points goal
}

type JoinGamePayload struct {
	Name     	string 				`json:"name"`
	GameCode 	string 				`json:"game_code"` // Added game code
	DesiredTeam shared.TeamEnum 	`json:"desired_team"` // Added desired team
}

type PlayCardPayload struct {
	Suit shared.Suit `json:"suit"`
	Rank string      `json:"rank"`
}

type DeclarePayload struct {
	DeclarationType string `json:"declaration_type"`
}

// --- Server -> Client Payload Structs ---

type GameCreatedPayload struct {
	GameCode string `json:"game_code"`
}

type LobbyUpdatePayload struct {
	Players []PlayerInfo `json:"players"`
}

type JoinErrorPayload struct {
	Message string `json:"message"`
}

type GameWaitPayload struct {
	Message string `json:"message"`
}

type PlayerInfo struct {
	ID   		string `json:"id"`
	Name 		string `json:"name"`
	Position 	int    `json:"position"` // Player's position in the game (0-3)
}

type TeamInfo struct {
	ID         string       `json:"id"`
	Players    []PlayerInfo `json:"players"`
	Score      int          `json:"score"`
	TeamNumber int          `json:"team_number"`
}

type GameStartPayload struct {
	GameID  	string       `json:"game_id"`
	Players 	[]PlayerInfo `json:"players"`
	Teams   	[]TeamInfo   `json:"teams"`
	PointsGoal 	int        `json:"points_goal"` // Added points goal
}

type DealHandPayload struct {
	Hand []shared.Card `json:"hand"`
}

type YourTurnPayload struct {
	PlayerID   string        `json:"player_id"`
	ValidMoves []shared.Card `json:"valid_moves,omitempty"`
}

type GameStatePayload struct {
	CurrentPlayerID    	string        `json:"current_player_id"`
	CardsOnTable       	[]shared.Card `json:"cards_on_table"`
	Team1Score         	int           `json:"team1_score"`  
	Team2Score         	int           `json:"team2_score"` 
	LastTrick          	[]shared.Card `json:"last_trick,omitempty"`
	LastTrickWinnerID 	string        `json:"last_winner_id,omitempty"`
	GameState          	string        `json:"game_state"`
}

type TrickEndPayload struct {
	Winner		shared.PlayedCard	`json:"winner"`
	WinnerID	string				`json:"winner_id"`
	Cards    	[]shared.Card 		`json:"cards"`    
	Points   	int           		`json:"points"`   
}

type RoundEndPayload struct {
	Team1RoundScore   int `json:"team1_round_score"` 
	Team2RoundScore   int `json:"team2_round_score"` 
	Team1TotalScore   int `json:"team1_total_score"` 
	Team2TotalScore   int `json:"team2_total_score"` 
}

type DeclarationInfoPayload struct {
	PlayerID        string `json:"player_id"`
	DeclarationType string `json:"declaration_type"`
	Points          int    `json:"points"`
}

type GameOverPayload struct {
	WinningTeamID string `json:"winning_team_id"` 
	FinalScoreT1  int    `json:"final_score_t1"` 
	FinalScoreT2  int    `json:"final_score_t2"` 
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type PlayerLeftPayload struct {
	PlayerID string `json:"player_id"`
}

type PlayerPlayedCardPayload struct {
	PlayerID string        	`json:"player_id"`
	Card     shared.Card 	`json:"card"` 
}

// Helper function to create a JSON message
func NewMessage(msgType string, payload interface{}) ([]byte, error) {
	// Handle nil payload specifically
	if payload == nil {
		msg := Message{
			Type:    msgType,
			Payload: nil, // Explicitly set Payload to nil for clarity
		}
		return json.Marshal(msg)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := Message{
		Type:    msgType,
		Payload: payloadBytes,
	}
	return json.Marshal(msg)
}
