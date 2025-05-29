package database

type GameResult struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	Player1     string `json:"player1"`
	Player2     string `json:"player2"`
	Player3     string `json:"player3"`
	Player4     string `json:"player4"`
	Player1Team int    `json:"player1_team"`
	Player2Team int    `json:"player2_team"`
	Player3Team int    `json:"player3_team"`
	Player4Team int    `json:"player4_team"`
	Team1Score  int    `json:"team1_score"`
	Team2Score  int    `json:"team2_score"`
}
