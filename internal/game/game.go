package game

import (
	"encoding/json"
	"log"
	"sync"

	"tressette-game/internal/protocol"
	"tressette-game/internal/shared"

	"github.com/google/uuid"
)

// GameState represents the current state of the game.
type GameState string

const (
	Waiting   GameState = "Waiting"   // Waiting for players (though Hub manages this mostly)
	Dealing   GameState = "Dealing"   // Cards are being dealt
	Playing   GameState = "Playing"   // Players are playing tricks
	Declaring GameState = "Declaring" // Phase for declaring combinations (optional)
	RoundOver GameState = "RoundOver" // A round (10 tricks) is finished
	GameOver  GameState = "GameOver"  // Target score reached
)

// MessageSender defines the function signature for sending messages back to clients.
// The Hub will provide an implementation of this.
type MessageSender func(clientID string, message []byte)

// Game represents the main game state machine.
type Game struct {
	ID                 		string           	`json:"id"`
	Players            		[4]*shared.Player 	`json:"-"`
	Teams              		[2]*shared.Team   	`json:"-"`
	Deck               		*shared.Deck      	`json:"-"`
	CurrentTrick       		*shared.Trick     	`json:"-"` 
	PlayerTurnIndex    		int              	`json:"player_turn_index"`
	GameState          		GameState        	`json:"game_state"`
	TargetScore        		int              	`json:"-"` 
	CardsOnTable       		[]shared.Card     	`json:"cards_on_table"`
	LedSuit            		shared.Suit       	`json:"led_suit"` 
	LastTrickWinnerIndex 	int              	`json:"last_trick_winner_index"`
	mu                 		sync.Mutex       
	sendMessage        		MessageSender    	`json:"-"`
}

// NewGame initializes a new game instance.
func NewGame(players [4]*shared.Player, targetScore int) *Game {
	teams := [2]*shared.Team{
		shared.NewTeam(1, players[0], players[2]),
		shared.NewTeam(2, players[1], players[3]),
	}

	gameID := uuid.New().String()

	return &Game{
		ID:                 	gameID,
		Players:            	players,
		Teams:              	teams,
		Deck:               	shared.NewDeck(),
		CurrentTrick:       	shared.NewTrick(),
		PlayerTurnIndex:    	0,
		GameState:          	Dealing, // Initial state is Dealing
		TargetScore:        	targetScore,
		CardsOnTable:       	[]shared.Card{},
		LedSuit:            	"",
		LastTrickWinnerIndex: 	-1,
	}
}

// StartGameLoop initializes the game and runs the first round.
// It's called in a goroutine by the Hub.
func (g *Game) StartGameLoop(sender MessageSender) {
	g.mu.Lock()
	g.sendMessage = sender
	log.Printf("Game %s: Starting game loop.", g.ID)

	// 1. Send Game Start message to all players
	playerInfos := make([]protocol.PlayerInfo, 4)
	for i, p := range g.Players {
		playerInfos[i] = protocol.PlayerInfo{ID: p.ID, Name: p.Name}
	}
	teamInfos := make([]protocol.TeamInfo, 2) 
	for i, t := range g.Teams {
		teamInfos[i] = protocol.TeamInfo{
			ID: t.ID,
			Players: []protocol.PlayerInfo{
				{ID: t.Players[0].ID, Name: t.Players[0].Name},
				{ID: t.Players[1].ID, Name: t.Players[1].Name},
			},
			Score: t.Score,
			TeamNumber: t.TeamNumber,
		}
	}

	startPayload := protocol.GameStartPayload{
		GameID:  g.ID,
		Players: playerInfos,
		Teams:   teamInfos,
	}
	startMsg, _ := protocol.NewMessage("game_start", startPayload)
	g.broadcast(startMsg)

	// 2. Start the first round
	g.startRound() // This will deal cards and send initial turn messages
	g.mu.Unlock() // Unlock after initial setup
}

// startRound begins a new round (shuffling, dealing, setting state).
// Assumes lock is held or called appropriately.
func (g *Game) startRound() {
	if g.GameState == GameOver {
		log.Printf("Game %s: Cannot start round, game is over.", g.ID)
		return
	}
	log.Printf("Game %s: Starting round...", g.ID)
	g.GameState = Dealing

	// Reset scores and state for the new round
	for _, team := range g.Teams {
		team.ResetScore()
	}
	g.Deck = shared.NewDeck() 
	g.Deck.Shuffle()
	g.CardsOnTable = []shared.Card{} 
	g.CurrentTrick = shared.NewTrick()
	g.LedSuit = ""

	// Determine who starts based on the last trick winner, or player 0 if first round
	if g.LastTrickWinnerIndex != -1 {
		g.PlayerTurnIndex = g.LastTrickWinnerIndex
	} else {
		g.PlayerTurnIndex = 0 // Default starter
	}

	// Deal 10 cards to each player
	hands := g.Deck.Deal(len(g.Players), 10)
	if hands == nil {
		log.Printf("Error dealing cards in game %s", g.ID)
		g.GameState = GameOver 
		g.broadcastError("Internal server error during dealing.") 
		return
	}
	for i, hand := range hands {
		if g.Players[i] != nil {
			g.Players[i].Hand = hand 
			// Send hand to the specific player
			dealPayload := protocol.DealHandPayload{Hand: hand} 
			dealMsg, _ := protocol.NewMessage("deal_hand", dealPayload) 
			g.sendToPlayer(g.Players[i].ID, dealMsg)
		} else {
			log.Printf("Error: Player %d is nil in game %s during dealing", i, g.ID)
			g.GameState = GameOver
			g.broadcastError("Internal server error: Player setup failed.") // Notify clients
			return
		}
	}

	g.GameState = Playing
	log.Printf("Game %s: Round started. Player %d (%s)'s turn.", g.ID, g.PlayerTurnIndex, g.Players[g.PlayerTurnIndex].Name)

	// Notify the starting player it's their turn and broadcast initial state
	g.broadcastGameState()
	g.notifyCurrentPlayerTurn()
}


// HandlePlayerAction processes incoming actions from a player.
func (g *Game) HandlePlayerAction(clientID string, msg protocol.Message) { 
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if game is already over
	if g.GameState == GameOver {
		log.Printf("Game %s: Action received from %s but game is over.", g.ID, clientID)
		g.sendErrorToPlayer(clientID, "Game is already over.")
		return
	}

	playerIndex := g.GetPlayerIndex(clientID)
	if playerIndex == -1 {
		log.Printf("Game %s: Action from unknown client ID %s", g.ID, clientID)
		return // Don't send error to potentially invalid client
	}

	switch msg.Type {
	case "play_card":
		if g.GameState != Playing {
			log.Printf("Game %s: Received play_card from %s in wrong state %s", g.ID, clientID, g.GameState)
			g.sendErrorToPlayer(clientID, "Cannot play card now.")
			return
		}
		if playerIndex != g.PlayerTurnIndex {
			log.Printf("Game %s: Received play_card from %s out of turn (current: %d)", g.ID, clientID, g.PlayerTurnIndex)
			g.sendErrorToPlayer(clientID, "Not your turn.")
			return
		}

		var payload protocol.PlayCardPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Game %s: Error unmarshalling play_card payload from %s: %v", g.ID, clientID, err)
			g.sendErrorToPlayer(clientID, "Invalid play_card message.")
			return
		}

		// Find the card in the player's hand (using Suit and Rank from payload)
		cardToPlay, found := g.Players[playerIndex].FindCard(shared.Suit(payload.Suit), payload.Rank)
		if !found {
			log.Printf("Game %s: Player %s tried to play card %s %s not in hand.", g.ID, clientID, payload.Rank, payload.Suit)
			g.sendErrorToPlayer(clientID, "Card not in your hand.")
			return
		}

		// Validate and process the play
		if !g.playCard(playerIndex, *cardToPlay) {
			g.sendErrorToPlayer(clientID, "Invalid move.")
		}

	case "declare":
		if g.GameState != Playing {
			log.Printf("Game %s: Received declare from %s in wrong state %s", g.ID, clientID, g.GameState)
			g.sendErrorToPlayer(clientID, "Cannot declare now.")
			return
		}

		var payload protocol.DeclarePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Game %s: Error unmarshalling declare payload from %s: %v", g.ID, clientID, err)
			g.sendErrorToPlayer(clientID, "Invalid declare message.")
			return
		}
		log.Printf("Game %s: Received 'declare' (%s) from %s - Placeholder", g.ID, payload.DeclarationType, clientID)

		g.sendErrorToPlayer(clientID, "Declarations not implemented yet.")

		// TODO: Implement declaration logic here

	default:
		log.Printf("Game %s: Received unhandled action type '%s' from %s", g.ID, msg.Type, clientID)
	}
}

// playCard handles the logic of playing a card, updating state, and notifying clients.
// Assumes lock is held. Returns true if successful, false otherwise.
func (g *Game) playCard(playerIndex int, card shared.Card) bool {
	player := g.Players[playerIndex]

	// Validate the move
	if !g.isValidPlay(player, card) {
		log.Printf("Game %s: Player %d (%s) attempted invalid play with card %s %s", g.ID, playerIndex, player.Name, card.Rank, card.Suit)
		return false
	}

	// Remove card from hand
	if !player.RemoveCard(card) { 
		// this should not happen if the game state is correct
		log.Printf("Error: Failed to remove card %s %s from player %d's hand unexpectedly.", card.Rank, card.Suit, playerIndex)
		g.broadcastError("Internal server error: Hand inconsistency.")
		g.GameState = GameOver
		log.Panicf("Game %s: Game over due to hand inconsistency.", g.ID)
	}

	// Add card to trick and table
	if len(g.CurrentTrick.Cards) == 0 {
		g.LedSuit = card.Suit
	}
	g.CurrentTrick.AddCard(card, playerIndex)
	g.CardsOnTable = append(g.CardsOnTable, card) // Keep track for state updates
	log.Printf("Game %s: Player %d (%s) played %s %s", g.ID, playerIndex, player.Name, card.Rank, card.Suit)

	g.notifyPlayerPlayedCard(player.ID, card) // Notify player of their action

	// Check if trick is complete
	if len(g.CurrentTrick.Cards) == len(g.Players) {
		g.endTrick() // Handles scoring, next turn/round logic
		} else {
			// Advance turn to the next player
			g.PlayerTurnIndex = (g.PlayerTurnIndex + 1) % len(g.Players)
			log.Printf("Game %s: Turn advanced to player %d (%s)", g.ID, g.PlayerTurnIndex, g.Players[g.PlayerTurnIndex].Name)
			defer g.notifyCurrentPlayerTurn()
		}
		
	// Broadcast game state update
	g.broadcastGameState()

	return true
}

// isValidPlay checks if playing a card is legal. Assumes lock is held.
func (g *Game) isValidPlay(player *shared.Player, card shared.Card) bool {
	if len(g.CurrentTrick.Cards) == 0 {
		return true // Can lead with any card
	}
	if player.HasSuit(g.LedSuit) {
		return card.Suit == g.LedSuit // Must follow suit if possible
	}
	return true // Can play any card if unable to follow suit
}

// endTrick concludes the current trick. Assumes lock is held.
func (g *Game) endTrick() {
	log.Printf("Game %s: Ending trick...", g.ID)
	winnerIndex := g.CurrentTrick.DetermineWinner(g.LedSuit)
	if winnerIndex == -1 {
		log.Panicf("Game %s: Error determining trick winner. No valid winner found.", g.ID)
		return
	}

	g.LastTrickWinnerIndex = winnerIndex
	winningPlayer := g.Players[winnerIndex]
	// Find the winning team based on the player index (0 or 2 -> team 0, 1 or 3 -> team 1)
	// is this good?
	winningTeam := g.Teams[winnerIndex%2]

	trickCardsForScoring := []shared.Card{}
	trickCardInfos := make([]shared.Card, len(g.CurrentTrick.Cards)) // For broadcast
	for i, pc := range g.CurrentTrick.Cards {
		trickCardsForScoring = append(trickCardsForScoring, pc.Card)
		trickCardInfos[i] = pc.Card
	}
	trickPoints := g.calculateTrickPoints(trickCardsForScoring)

	isLastTrick := len(g.Players[0].Hand) == 0
	if isLastTrick {
		trickPoints += 3 // Scaled bonus point for last trick
		log.Printf("Game %s: Last trick bonus point (scaled: 3) awarded.", g.ID)
	}

	winningTeam.AddScore(trickPoints)
	log.Printf("Game %s: Trick won by Player %d (%s). Scaled Points: %d. Team %d",
		g.ID, winnerIndex, winningPlayer.Name, trickPoints, winningTeam.TeamNumber)

	// Broadcast trick end info
	trickEndPayload := protocol.TrickEndPayload{
		WinnerID: winningPlayer.ID,
		Cards:    trickCardInfos,
		Points:   trickPoints,    // Scaled points won
	}
	trickEndMsg, _ := protocol.NewMessage("trick_end", trickEndPayload)
	g.broadcast(trickEndMsg)

	// Reset for next trick or round
	g.CardsOnTable = []shared.Card{}
	g.CurrentTrick = shared.NewTrick()
	g.LedSuit = ""
	g.PlayerTurnIndex = winnerIndex // Winner leads next

	// Check if the round is over
	if isLastTrick {
		g.endRound()
	} else {
		log.Printf("Game %s: Next trick. Player %d (%s)'s turn.", g.ID, g.PlayerTurnIndex, g.Players[g.PlayerTurnIndex].Name)
		// Broadcast state update (shows empty table) and notify next player
		g.broadcastGameState()
		g.notifyCurrentPlayerTurn()
	}
}

// calculateTrickPoints calculates scaled points. Assumes lock is held.
func (g *Game) calculateTrickPoints(trickCards []shared.Card) int {
	scaledPoints := 0
	for _, card := range trickCards {
		scaledPoints += card.Value // Values are already scaled
	}
	return scaledPoints
}

// endRound finalizes the round. Assumes lock is held.
func (g *Game) endRound() {
	// This needs to be reworked
	// team.Score is the scaled score of the current round
	// This function should save the score to the database and reset the round
	log.Printf("Game %s: Round ended.", g.ID)
	g.GameState = RoundOver

	// Update total scores
	for _, team := range g.Teams {
		team.AddScore(team.Score)
	}

	// Broadcast round end info
	roundEndPayload := protocol.RoundEndPayload{
		Team1RoundScore: g.Teams[0].Score,
		Team2RoundScore: g.Teams[1].Score,
		Team1TotalScore: g.Teams[0].Score,
		Team2TotalScore: g.Teams[1].Score,
	}
	roundEndMsg, _ := protocol.NewMessage("round_end", roundEndPayload)
	g.broadcast(roundEndMsg)

	// Check for game over
	// This is not needed, multiple rounds is not a feature yet
	gameOver := false
	var winningTeam *shared.Team
	scaledTargetScore := g.TargetScore * 3
	for _, team := range g.Teams {
		if team.Score >= scaledTargetScore {
			g.GameState = GameOver
			gameOver = true
			winningTeam = team
			log.Printf("Game %s: Game Over! Team %d (ID: %s) wins.", g.ID, team.TeamNumber, team.ID)

			// Broadcast game over
			gameOverPayload := protocol.GameOverPayload{
				WinningTeamID: winningTeam.ID,
				FinalScoreT1:  g.Teams[0].Score,
				FinalScoreT2:  g.Teams[1].Score,
			}
			gameOverMsg, _ := protocol.NewMessage("game_over", gameOverPayload)
			g.broadcast(gameOverMsg)

			break
		}
	}

	if !gameOver {
		log.Printf("Game %s: Preparing for next round.", g.ID)
		g.startRound()
	} else {
		log.Printf("Game %s: Final state reached. Winning Team: %d (ID: %s)", g.ID, winningTeam.TeamNumber, winningTeam.ID)
	}
}


// HandlePlayerDisconnect handles a player leaving mid-game.
func (g *Game) HandlePlayerDisconnect(clientID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.GameState == GameOver {
		log.Printf("Game %s: Player %s disconnected, but game already over.", g.ID, clientID)
		return // Game already over
	}

	playerIndex := g.GetPlayerIndex(clientID)
	if playerIndex == -1 {
		log.Printf("Game %s: Disconnect from unknown or already removed client ID %s", g.ID, clientID)
		return
	}

	playerName := g.Players[playerIndex].Name
	log.Printf("Game %s: Player %s (%s) disconnected.", g.ID, clientID, playerName)
	g.GameState = GameOver // Forfeit the game

	// Broadcast player left message
	leftPayload := protocol.PlayerLeftPayload{PlayerID: clientID}
	leftMsg, _ := protocol.NewMessage("player_left", leftPayload)
	g.broadcast(leftMsg) // Notify remaining players

	// Broadcast game over (due to forfeit)
	// Determine winning team (the one that didn't disconnect)
	var winningTeam *shared.Team
	if playerIndex == 0 || playerIndex == 2 { // Player was in Team 1 (index 0)
		winningTeam = g.Teams[1] // Team 2 (index 1) wins
	} else { // Player was in Team 2 (index 1)
		winningTeam = g.Teams[0] // Team 1 (index 0) wins
	}

	gameOverPayload := protocol.GameOverPayload{
		WinningTeamID: winningTeam.ID,
		FinalScoreT1:  g.Teams[0].Score,
		FinalScoreT2:  g.Teams[1].Score,
	}
	gameOverMsg, _ := protocol.NewMessage("game_over", gameOverPayload)
	g.broadcast(gameOverMsg) // Notify remaining players

	// TODO: Signal Hub to clean up this game instance? Or Hub handles based on state?
	// Consider saving winningTeam.TeamNumber (1 or 2) to DB instead of UUID
	log.Printf("Game %s: Game ended due to player %s disconnect. Team %d (ID: %s) wins by forfeit.", g.ID, clientID, winningTeam.TeamNumber, winningTeam.ID)
}


// --- Messaging Helpers (Assume lock is held or called safely) ---

// broadcast sends a message to all players in the game.
func (g *Game) broadcast(message []byte) {
	if g.sendMessage == nil {
		log.Printf("Game %s: Error - sendMessage callback is nil during broadcast.", g.ID)
		return
	}
	for _, player := range g.Players {
		if player != nil {
			g.sendMessage(player.ID, message)
		}
	}
}

// sendToPlayer sends a message to a specific player by ID.
func (g *Game) sendToPlayer(playerID string, message []byte) {
	if g.sendMessage == nil {
		log.Printf("Game %s: Error - sendMessage callback is nil when sending to %s.", g.ID, playerID)
		return
	}
	g.sendMessage(playerID, message)
}

// sendErrorToPlayer sends an error message to a specific player.
func (g *Game) sendErrorToPlayer(playerID string, errorMsg string) {
	payload := protocol.ErrorPayload{Message: errorMsg}
	msgBytes, err := protocol.NewMessage("error", payload)
	if err != nil {
		log.Printf("Game %s: Error creating error message for %s: %v", g.ID, playerID, err)
		return
	}
	g.sendToPlayer(playerID, msgBytes)
}

// broadcastError sends an error message to all players.
func (g *Game) broadcastError(errorMsg string) {
	payload := protocol.ErrorPayload{Message: errorMsg}
	msgBytes, err := protocol.NewMessage("error", payload)
	if err != nil {
		log.Printf("Game %s: Error creating broadcast error message: %v", g.ID, err)
		return
	}
	g.broadcast(msgBytes)
}


// broadcastGameState sends the current game state to all players.
func (g *Game) broadcastGameState() {
	// Create payload (ensure sensitive info like full hands isn't sent)
	var currentPlayerID string
	if g.PlayerTurnIndex >= 0 && g.PlayerTurnIndex < len(g.Players) && g.Players[g.PlayerTurnIndex] != nil {
		currentPlayerID = g.Players[g.PlayerTurnIndex].ID
	} else {
		log.Printf("Game %s: Warning - Invalid PlayerTurnIndex %d during broadcastGameState", g.ID, g.PlayerTurnIndex)
		// Handle appropriately, maybe set to empty or log error
	}

	var team1Score, team2Score int
	if len(g.Teams) > 0 && g.Teams[0] != nil {
		team1Score = g.Teams[0].Score
	}
	if len(g.Teams) > 1 && g.Teams[1] != nil {
		team2Score = g.Teams[1].Score
	}


	payload := protocol.GameStatePayload{
		CurrentPlayerID:    currentPlayerID,
		CardsOnTable:       g.CardsOnTable,
		Team1Score:         team1Score,
		Team2Score:         team2Score,
		GameState:          string(g.GameState),
	}
	msgBytes, _ := protocol.NewMessage("game_state_update", payload)
	g.broadcast(msgBytes)
}

// notifyCurrentPlayerTurn sends the 'your_turn' message.
func (g *Game) notifyCurrentPlayerTurn() {
	currentPlayer := g.Players[g.PlayerTurnIndex]

	payload := protocol.YourTurnPayload{
		PlayerID: currentPlayer.ID,
	}
	msgBytes, _ := protocol.NewMessage("your_turn", payload)
	g.sendToPlayer(currentPlayer.ID, msgBytes)
}

// notify the player that just played a card
func (g *Game) notifyPlayerPlayedCard(playerID string, card shared.Card) {
	payload := protocol.PlayerPlayedCardPayload{
		PlayerID: playerID,
		Card:     card,
	}
	msgBytes, _ := protocol.NewMessage("you_played", payload)

	g.sendToPlayer(playerID, msgBytes)
}

// --- Utility Helpers ---

// GetPlayerByID finds a player struct by their ID.
func (g *Game) GetPlayerByID(playerID string) *shared.Player {
	for _, p := range g.Players {
		if p != nil && p.ID == playerID {
			return p
		}
	}
	return nil
}

// GetPlayerIndex finds the index (0-3) of a player by their ID. Returns -1 if not found.
func (g *Game) GetPlayerIndex(playerID string) int {
	for i, p := range g.Players {
		// Add nil check for player
		if p != nil && p.ID == playerID {
			return i
		}
	}
	return -1 // Not found
}